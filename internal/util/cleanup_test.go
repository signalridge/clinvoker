package util

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupBackendSessionWithRunner(t *testing.T) {
	t.Run("gemini calls runner with correct args", func(t *testing.T) {
		var calledName string
		var calledArgs []string
		mockRunner := func(name string, args ...string) error {
			calledName = name
			calledArgs = args
			return nil
		}

		CleanupBackendSessionWithRunner("gemini", "test-uuid-1234", mockRunner)

		if calledName != "gemini" {
			t.Errorf("expected command 'gemini', got %q", calledName)
		}
		if len(calledArgs) != 2 || calledArgs[0] != "--delete-session" || calledArgs[1] != "test-uuid-1234" {
			t.Errorf("expected args ['--delete-session', 'test-uuid-1234'], got %v", calledArgs)
		}
	})

	t.Run("gemini skips when sessionID is empty", func(t *testing.T) {
		called := false
		mockRunner := func(name string, args ...string) error {
			called = true
			return nil
		}

		CleanupBackendSessionWithRunner("gemini", "", mockRunner)

		if called {
			t.Error("runner should not be called when sessionID is empty")
		}
	})

	t.Run("codex skips when sessionID is empty", func(t *testing.T) {
		// This test verifies the early return for codex
		CleanupBackendSessionWithRunner("codex", "", nil)
		// Should not panic
	})

	t.Run("claude does nothing (uses native flag)", func(t *testing.T) {
		called := false
		mockRunner := func(name string, args ...string) error {
			called = true
			return nil
		}

		CleanupBackendSessionWithRunner("claude", "some-session-id", mockRunner)

		if called {
			t.Error("runner should not be called for claude (uses native --no-session-persistence)")
		}
	})

	t.Run("unknown backend does nothing", func(t *testing.T) {
		called := false
		mockRunner := func(name string, args ...string) error {
			called = true
			return nil
		}

		CleanupBackendSessionWithRunner("unknown", "some-id", mockRunner)

		if called {
			t.Error("runner should not be called for unknown backend")
		}
	})
}

func TestContextCommandRunner(t *testing.T) {
	t.Run("creates runner from context", func(t *testing.T) {
		ctx := context.Background()
		runner := ContextCommandRunner(ctx)
		if runner == nil {
			t.Error("expected non-nil runner")
		}
	})

	t.Run("runner works with valid command", func(t *testing.T) {
		ctx := context.Background()
		runner := ContextCommandRunner(ctx)
		err := runner("true") // Unix true command always succeeds
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestCleanupCodexSessionInDir(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "codex-cleanup-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a fixed time for predictable directory structure
	testTime := time.Date(2026, 1, 29, 12, 0, 0, 0, time.UTC)

	// Create the expected directory structure: {tmpDir}/.codex/sessions/2026/01/29/
	sessionDir := filepath.Join(tmpDir, ".codex", "sessions", "2026", "01", "29")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	t.Run("deletes matching session file", func(t *testing.T) {
		threadID := "019c0523-e080-7fb1-a8ea-0530361cbf0f"
		sessionFile := filepath.Join(sessionDir, "rollout-2026-01-29T00-06-03-"+threadID+".jsonl")

		// Create the session file
		if err := os.WriteFile(sessionFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create session file: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
			t.Fatal("session file should exist before cleanup")
		}

		// Run cleanup
		CleanupCodexSessionInDir(threadID, tmpDir, testTime)

		// Verify file is deleted
		if _, err := os.Stat(sessionFile); !os.IsNotExist(err) {
			t.Error("session file should be deleted after cleanup")
		}
	})

	t.Run("does not delete non-matching files", func(t *testing.T) {
		// Create a file with different UUID
		otherFile := filepath.Join(sessionDir, "rollout-2026-01-29T00-06-03-different-uuid.jsonl")
		if err := os.WriteFile(otherFile, []byte("other content"), 0644); err != nil {
			t.Fatalf("failed to create other file: %v", err)
		}

		// Run cleanup with non-matching thread ID
		CleanupCodexSessionInDir("non-existent-uuid", tmpDir, testTime)

		// Verify file still exists
		if _, err := os.Stat(otherFile); os.IsNotExist(err) {
			t.Error("non-matching file should not be deleted")
		}
	})

	t.Run("handles missing directory gracefully", func(t *testing.T) {
		// Use a different time that has no directory
		differentTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

		// Should not panic
		CleanupCodexSessionInDir("some-uuid", tmpDir, differentTime)
	})

	t.Run("skips directories", func(t *testing.T) {
		// Create a subdirectory with matching name (edge case)
		subDir := filepath.Join(sessionDir, "rollout-2026-01-29T00-06-03-subdir-uuid")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}

		// Should not delete the directory
		CleanupCodexSessionInDir("subdir-uuid", tmpDir, testTime)

		// Verify directory still exists
		info, err := os.Stat(subDir)
		if os.IsNotExist(err) {
			t.Error("subdirectory should not be deleted")
		}
		if err == nil && !info.IsDir() {
			t.Error("should still be a directory")
		}
	})
}

func TestContainsThreadID(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		threadID string
		want     bool
	}{
		{
			name:     "matches UUID in codex filename",
			filename: "rollout-2026-01-29T00-06-03-019c0523-e080-7fb1-a8ea-0530361cbf0f.jsonl",
			threadID: "019c0523-e080-7fb1-a8ea-0530361cbf0f",
			want:     true,
		},
		{
			name:     "no match for different UUID",
			filename: "rollout-2026-01-29T00-06-03-019c0523-e080-7fb1-a8ea-0530361cbf0f.jsonl",
			threadID: "different-uuid-1234-5678-abcd",
			want:     false,
		},
		{
			name:     "empty threadID returns false",
			filename: "rollout-2026-01-29T00-06-03-019c0523-e080-7fb1-a8ea-0530361cbf0f.jsonl",
			threadID: "",
			want:     false,
		},
		{
			name:     "empty filename returns false",
			filename: "",
			threadID: "019c0523-e080-7fb1-a8ea-0530361cbf0f",
			want:     false,
		},
		{
			name:     "partial UUID match",
			filename: "rollout-2026-01-29T00-06-03-019c0523-e080-7fb1-a8ea-0530361cbf0f.jsonl",
			threadID: "019c0523",
			want:     true, // substring match is intentional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsThreadID(tt.filename, tt.threadID)
			if got != tt.want {
				t.Errorf("ContainsThreadID(%q, %q) = %v, want %v",
					tt.filename, tt.threadID, got, tt.want)
			}
		})
	}
}
