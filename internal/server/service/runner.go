package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/server/core"
	"github.com/signalridge/clinvoker/internal/session"
	"github.com/signalridge/clinvoker/internal/util"
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
		RequestedFormat: prep.requestedFormat,
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
	result.TokenUsage = util.TokenUsageFromBackend(coreRes.Usage)

	// Update session if needed
	if sess != nil {
		if coreRes.BackendSessionID != "" {
			sess.BackendSessionID = coreRes.BackendSessionID
		}
		// Convert core.Result to backend.UnifiedResponse for util function
		var resp *backend.UnifiedResponse
		if coreRes != nil {
			resp = &backend.UnifiedResponse{
				Usage: coreRes.Usage,
				Error: coreRes.Error,
			}
		}
		util.UpdateSessionFromResponse(sess, result.ExitCode, result.Error, resp)
		if err := store.Save(sess); err != nil && logger != nil {
			logger.Warn("failed to save session", "session_id", sess.ID, "error", err)
		}
	}

	// Cleanup backend session for ephemeral requests
	if opts.Ephemeral {
		util.CleanupBackendSessionWithContext(ctx, req.Backend, coreRes.BackendSessionID)
	}

	if forceStateless {
		result.SessionID = ""
	}

	return result, nil
}
