package app

import (
	"fmt"
	"io"
	"os"

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
	stat, _ := os.Stdin.Stat()
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
