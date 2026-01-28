package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/server/core"
	"github.com/signalridge/clinvoker/internal/session"
)

// PromptRunner executes a single prompt request.
type PromptRunner interface {
	ExecutePrompt(ctx context.Context, req *PromptRequest) (*PromptResult, error)
}

// StatefulRunner executes prompts with session persistence.
type StatefulRunner struct {
	store  *session.Store
	logger *slog.Logger
}

// NewStatefulRunner creates a new stateful runner.
func NewStatefulRunner(store *session.Store, logger *slog.Logger) *StatefulRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &StatefulRunner{store: store, logger: logger}
}

// ExecutePrompt executes a prompt with session persistence.
func (r *StatefulRunner) ExecutePrompt(ctx context.Context, req *PromptRequest) (*PromptResult, error) {
	return executePrompt(ctx, req, r.store, r.logger, false)
}

// StatelessRunner executes prompts without session persistence.
type StatelessRunner struct {
	logger *slog.Logger
}

// NewStatelessRunner creates a new stateless runner.
func NewStatelessRunner(logger *slog.Logger) *StatelessRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &StatelessRunner{logger: logger}
}

// ExecutePrompt executes a prompt in stateless mode.
func (r *StatelessRunner) ExecutePrompt(ctx context.Context, req *PromptRequest) (*PromptResult, error) {
	result, err := executePrompt(ctx, req, nil, r.logger, true)
	if result != nil {
		result.SessionID = ""
	}
	return result, err
}

func executePrompt(ctx context.Context, req *PromptRequest, store *session.Store, logger *slog.Logger, forceStateless bool) (*PromptResult, error) {
	start := time.Now()
	result := &PromptResult{
		Backend: req.Backend,
	}

	prep, err := preparePrompt(req, forceStateless)
	if err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
		result.DurationMS = time.Since(start).Milliseconds()
		return result, nil
	}

	b := prep.backend
	model := prep.model
	opts := prep.opts

	// Create session (skip if ephemeral or no store)
	var sess *session.Session
	if store != nil && !opts.Ephemeral {
		var sessErr error
		sess, sessErr = session.NewSession(req.Backend, req.WorkDir)
		if sessErr == nil {
			sess.SetModel(model)
			sess.InitialPrompt = req.Prompt
			sess.SetStatus(session.StatusActive)
			sess.AddTag("api")
			for k, v := range req.Metadata {
				sess.SetMetadata(k, v)
			}
			if err := store.Save(sess); err == nil {
				result.SessionID = sess.ID
			}
		}
	}

	coreRes, execErr := core.Execute(ctx, &core.Request{
		Backend:         b,
		Prompt:          req.Prompt,
		Options:         opts,
		RequestedFormat: backend.OutputFormat(req.OutputFormat),
	})
	if execErr != nil {
		result.Error = execErr.Error()
		result.ExitCode = 1
		result.DurationMS = time.Since(start).Milliseconds()
		return result, nil
	}

	result.ExitCode = coreRes.ExitCode
	result.Error = coreRes.Error
	result.Output = coreRes.Output
	result.DurationMS = time.Since(start).Milliseconds()
	result.TokenUsage = tokenUsageFromBackend(coreRes.Usage)

	// Update session if needed
	if sess != nil {
		if coreRes.BackendSessionID != "" {
			sess.BackendSessionID = coreRes.BackendSessionID
		}
		updateSessionFromExecResult(sess, result.ExitCode, result.Error, coreRes)
		if err := store.Save(sess); err != nil && logger != nil {
			logger.Warn("failed to save session", "session_id", sess.ID, "error", err)
		}
	}

	// Cleanup backend session for ephemeral requests
	if opts.Ephemeral {
		cleanupBackendSession(ctx, req.Backend, coreRes.BackendSessionID)
	}

	if forceStateless {
		result.SessionID = ""
	}

	return result, nil
}

func applyUnifiedDefaults(opts *backend.UnifiedOptions, cfg *config.Config) {
	if opts == nil || cfg == nil {
		return
	}
	if opts.ApprovalMode == "" && cfg.UnifiedFlags.ApprovalMode != "" {
		opts.ApprovalMode = backend.ApprovalMode(cfg.UnifiedFlags.ApprovalMode)
	}
	if opts.SandboxMode == "" && cfg.UnifiedFlags.SandboxMode != "" {
		opts.SandboxMode = backend.SandboxMode(cfg.UnifiedFlags.SandboxMode)
	}
	if opts.MaxTurns == 0 && cfg.UnifiedFlags.MaxTurns > 0 {
		opts.MaxTurns = cfg.UnifiedFlags.MaxTurns
	}
	if opts.MaxTokens == 0 && cfg.UnifiedFlags.MaxTokens > 0 {
		opts.MaxTokens = cfg.UnifiedFlags.MaxTokens
	}
	if !opts.Verbose && cfg.UnifiedFlags.Verbose {
		opts.Verbose = true
	}
	if cfg.UnifiedFlags.DryRun {
		opts.DryRun = true
	}
}

func tokenUsageFromBackend(usage *backend.TokenUsage) *session.TokenUsage {
	if usage == nil {
		return nil
	}
	return &session.TokenUsage{
		InputTokens:  int64(usage.InputTokens),
		OutputTokens: int64(usage.OutputTokens),
	}
}

func updateSessionFromExecResult(sess *session.Session, exitCode int, errMsg string, res *core.Result) {
	if sess == nil {
		return
	}

	sess.IncrementTurn()

	if res != nil && res.Usage != nil {
		sess.AddTokens(int64(res.Usage.InputTokens), int64(res.Usage.OutputTokens))
	}

	if errMsg != "" {
		sess.SetError(errMsg)
		return
	}

	if exitCode == 0 {
		sess.Complete()
		return
	}

	sess.SetError("backend execution failed")
}

// cleanupBackendSession cleans up backend session for ephemeral requests.
func cleanupBackendSession(ctx context.Context, backendName, sessionID string) {
	switch backendName {
	case "gemini":
		if sessionID == "" {
			return
		}
		_ = execCommandContext(ctx, "gemini", "--delete-session", sessionID).Run()
	case "codex":
		if sessionID == "" {
			return
		}
		cleanupCodexSession(sessionID)
	}
}

func cleanupCodexSession(threadID string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	cleanupCodexSessionInDir(threadID, home, time.Now())
}

func cleanupCodexSessionInDir(threadID, baseDir string, now time.Time) {
	sessionDir := filepath.Join(baseDir, ".codex", "sessions",
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		fmt.Sprintf("%02d", now.Day()),
	)

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if containsThreadID(entry.Name(), threadID) {
			sessionPath := filepath.Join(sessionDir, entry.Name())
			_ = os.Remove(sessionPath)
			return
		}
	}
}

func containsThreadID(filename, threadID string) bool {
	return threadID != "" && strings.Contains(filename, threadID)
}

func execCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	if ctx == nil {
		ctx = context.Background()
	}
	return exec.CommandContext(ctx, name, args...)
}
