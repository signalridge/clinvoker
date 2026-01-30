package app

import (
	"fmt"
	"io"
	"os"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
	"github.com/signalridge/clinvoker/internal/util"
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
// This is a thin wrapper around util.ApplyUnifiedDefaults for package convenience.
func applyUnifiedDefaults(opts *backend.UnifiedOptions, cfg *config.Config, effectiveDryRun bool) {
	util.ApplyUnifiedDefaults(opts, cfg, effectiveDryRun)
}

// applyBackendDefaults applies backend-specific config defaults to options.
// This should be called after applyUnifiedDefaults to allow backend config to override.
func applyBackendDefaults(opts *backend.UnifiedOptions, backendName string, cfg *config.Config) {
	util.ApplyBackendDefaults(opts, backendName, cfg)
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
// This is a thin wrapper around util.UpdateSessionFromResponse for package convenience.
func updateSessionFromResponse(sess *session.Session, exitCode int, errMsg string, resp *backend.UnifiedResponse) {
	util.UpdateSessionFromResponse(sess, exitCode, errMsg, resp)
}

// updateSessionAfterExecution updates session status after command execution.
// This variant also saves the session to the store.
func updateSessionAfterExecution(store *session.Store, sess *session.Session, exitCode int, errorMsg string, quiet bool) {
	updateSessionAfterExecutionWithBackendID(store, sess, exitCode, errorMsg, "", quiet)
}

// updateSessionAfterExecutionWithBackendID updates session status after command execution,
// including the backend's session ID for resume functionality.
func updateSessionAfterExecutionWithBackendID(store *session.Store, sess *session.Session, exitCode int, errorMsg string, backendSessionID string, quiet bool) {
	if sess == nil {
		return
	}

	sess.IncrementTurn()

	// Store the backend's session ID if provided
	if backendSessionID != "" {
		sess.BackendSessionID = backendSessionID
	}

	if exitCode == 0 && errorMsg == "" {
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

// filterResumableSessions returns sessions that have a backend session ID.
// Sessions without a backend session ID cannot be resumed.
func filterResumableSessions(sessions []*session.Session) []*session.Session {
	if len(sessions) == 0 {
		return nil
	}
	resumable := make([]*session.Session, 0, len(sessions))
	for _, s := range sessions {
		if s != nil && s.BackendSessionID != "" {
			resumable = append(resumable, s)
		}
	}
	return resumable
}

// cleanupBackendSession cleans up the backend's session after execution in ephemeral mode.
// This is a thin wrapper around util.CleanupBackendSession for package convenience.
func cleanupBackendSession(backendName, sessionID string) {
	util.CleanupBackendSession(backendName, sessionID)
}
