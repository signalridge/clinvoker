package util

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandRunner is a function type for running commands, allowing mocking in tests.
type CommandRunner func(name string, args ...string) error

// DefaultCommandRunner executes a command using exec.Command.
var DefaultCommandRunner CommandRunner = func(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// ContextCommandRunner creates a command runner that uses context for cancellation.
func ContextCommandRunner(ctx context.Context) CommandRunner {
	return func(name string, args ...string) error {
		return exec.CommandContext(CleanupContext(ctx), name, args...).Run()
	}
}

// CleanupBackendSession cleans up the backend's session after execution in ephemeral mode.
// This is needed for backends that don't support a native --no-session-persistence flag.
func CleanupBackendSession(backendName, sessionID string) {
	CleanupBackendSessionWithRunner(backendName, sessionID, DefaultCommandRunner)
}

// CleanupBackendSessionWithContext cleans up using the provided context for cancellation.
func CleanupBackendSessionWithContext(ctx context.Context, backendName, sessionID string) {
	CleanupBackendSessionWithRunner(backendName, sessionID, ContextCommandRunner(ctx))
}

// CleanupBackendSessionWithRunner is the testable version that accepts a command runner.
func CleanupBackendSessionWithRunner(backendName, sessionID string, runner CommandRunner) {
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
		CleanupCodexSession(sessionID)

		// Claude uses --no-session-persistence flag, no cleanup needed
	}
}

// CleanupCodexSession removes a Codex session file by thread ID.
// Codex stores sessions as files named: rollout-YYYY-MM-DDTHH-MM-SS-UUID.jsonl
func CleanupCodexSession(threadID string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	CleanupCodexSessionInDir(threadID, home, time.Now())
}

// CleanupCodexSessionInDir is the testable version that accepts base directory and time.
func CleanupCodexSessionInDir(threadID, baseDir string, now time.Time) {
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
		if ContainsThreadID(entry.Name(), threadID) {
			sessionPath := filepath.Join(sessionDir, entry.Name())
			_ = os.Remove(sessionPath)
			return
		}
	}
}

// ContainsThreadID checks if a filename contains the thread ID.
func ContainsThreadID(filename, threadID string) bool {
	// Thread ID is a UUID like "019c0523-e080-7fb1-a8ea-0530361cbf0f"
	// File name is like "rollout-2026-01-29T00-06-03-019c0523-e080-7fb1-a8ea-0530361cbf0f.jsonl"
	return threadID != "" && strings.Contains(filename, threadID)
}
