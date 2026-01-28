package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
)

// readInputFromFileOrStdin reads input from a file or stdin.
// If filePath is provided, reads from that file.
// Otherwise, reads from stdin if data is available.
func readInputFromFileOrStdin(filePath string) ([]byte, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return data, nil
	}

	// Check if stdin has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat stdin: %w", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("no input provided (use --file or pipe JSON to stdin)")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}
	return data, nil
}

// getBackendOrError gets a backend by name and validates it's available.
// Returns the backend and nil error on success, or nil and an error on failure.
func getBackendOrError(backendName string) (backend.Backend, error) {
	b, err := backend.Get(backendName)
	if err != nil {
		return nil, err
	}
	if !b.IsAvailable() {
		return nil, fmt.Errorf("backend %q not available", backendName)
	}
	return b, nil
}

// resolveModel determines the model to use based on explicit value, config, etc.
func resolveModel(explicit, backendName, globalModel string) string {
	if explicit != "" {
		return explicit
	}
	if globalModel != "" {
		return globalModel
	}
	cfg := config.Get()
	if bcfg, ok := cfg.Backends[backendName]; ok {
		return bcfg.Model
	}
	return ""
}

// applyUnifiedDefaults applies unified flag defaults from config to options.
func applyUnifiedDefaults(opts *backend.UnifiedOptions, cfg *config.Config, effectiveDryRun bool) {
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
	if !opts.DryRun && effectiveDryRun {
		opts.DryRun = true
	}
}

// createAndSaveSession creates a new session and saves it to the store.
// Returns the session (may be nil if creation failed) and logs warnings if quiet is false.
func createAndSaveSession(store *session.Store, backendName, workDir, model, prompt string, tags []string, title string, quiet bool) *session.Session {
	sess, err := session.NewSession(backendName, workDir)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Warning: failed to create session: %v\n", err)
		}
		return nil
	}

	sess.SetModel(model)
	sess.InitialPrompt = prompt
	sess.SetStatus(session.StatusActive)

	for _, tag := range tags {
		sess.AddTag(tag)
	}

	if title != "" {
		sess.SetTitle(title)
	}

	if err := store.Save(sess); err != nil && !quiet {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
	}

	return sess
}

// updateSessionFromResponse updates a session with execution results.
func updateSessionFromResponse(sess *session.Session, exitCode int, errMsg string, resp *backend.UnifiedResponse) {
	if sess == nil {
		return
	}

	sess.IncrementTurn()

	if resp != nil && resp.Usage != nil {
		sess.AddTokens(int64(resp.Usage.InputTokens), int64(resp.Usage.OutputTokens))
	}

	if resp != nil && resp.Error != "" {
		sess.SetError(resp.Error)
		return
	}

	if exitCode == 0 {
		sess.Complete()
		return
	}

	if errMsg != "" {
		sess.SetError(errMsg)
		return
	}

	sess.SetError("backend execution failed")
}

// updateSessionAfterExecution updates session status after command execution.
func updateSessionAfterExecution(store *session.Store, sess *session.Session, exitCode int, errorMsg string, quiet bool) {
	if sess == nil {
		return
	}

	sess.IncrementTurn()
	if exitCode == 0 {
		sess.Complete()
	} else {
		sess.SetError(errorMsg)
	}

	if err := store.Save(sess); err != nil && !quiet {
		fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", err)
	}
}

// truncateString truncates a string to maxLen, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// shortSessionID returns the first 8 characters of a session ID, or the full ID if shorter.
func shortSessionID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// commandRunner is a function type for running commands, allowing mocking in tests.
type commandRunner func(name string, args ...string) error

// defaultCommandRunner executes a command using exec.Command.
var defaultCommandRunner commandRunner = func(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// runCommand is the package-level command runner, can be replaced in tests.
var runCommand = defaultCommandRunner

// cleanupBackendSession cleans up the backend's session after execution in ephemeral mode.
// This is needed for backends that don't support a native --no-session-persistence flag.
func cleanupBackendSession(backendName, sessionID string) {
	cleanupBackendSessionWithRunner(backendName, sessionID, runCommand)
}

// cleanupBackendSessionWithRunner is the testable version that accepts a command runner.
func cleanupBackendSessionWithRunner(backendName, sessionID string, runner commandRunner) {
	switch backendName {
	case "gemini":
		// Gemini: delete the session by UUID
		if sessionID == "" {
			return
		}
		_ = runner("gemini", "--delete-session", sessionID) // Ignore errors silently

	case "codex":
		// Codex: delete the session file from ~/.codex/sessions/
		if sessionID == "" {
			return
		}
		cleanupCodexSession(sessionID)

		// Claude uses --no-session-persistence flag, no cleanup needed
	}
}

// cleanupCodexSession removes a Codex session file by thread ID.
// Codex stores sessions as files named: rollout-YYYY-MM-DDTHH-MM-SS-UUID.jsonl
func cleanupCodexSession(threadID string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	cleanupCodexSessionInDir(threadID, home, time.Now())
}

// cleanupCodexSessionInDir is the testable version that accepts base directory and time.
func cleanupCodexSessionInDir(threadID, baseDir string, now time.Time) {
	// Codex stores sessions in {baseDir}/.codex/sessions/YYYY/MM/DD/
	sessionDir := filepath.Join(baseDir, ".codex", "sessions",
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		fmt.Sprintf("%02d", now.Day()),
	)

	// Find and delete the session file containing the thread ID in its name
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Session files contain UUID in name: rollout-...-UUID.jsonl
		if containsThreadID(entry.Name(), threadID) {
			sessionPath := filepath.Join(sessionDir, entry.Name())
			_ = os.Remove(sessionPath)
			return
		}
	}
}

// containsThreadID checks if a filename contains the thread ID.
func containsThreadID(filename, threadID string) bool {
	// Thread ID is a UUID like "019c0523-e080-7fb1-a8ea-0530361cbf0f"
	// File name is like "rollout-2026-01-29T00-06-03-019c0523-e080-7fb1-a8ea-0530361cbf0f.jsonl"
	return threadID != "" && strings.Contains(filename, threadID)
}
