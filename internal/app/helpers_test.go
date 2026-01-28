package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/session"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length unchanged",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "long string truncated",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "very short maxLen",
			input:  "hello",
			maxLen: 3,
			want:   "hel",
		},
		{
			name:   "maxLen of 4",
			input:  "hello world",
			maxLen: 4,
			want:   "h...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "zero maxLen",
			input:  "hello",
			maxLen: 0,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestResolveModel(t *testing.T) {
	tests := []struct {
		name        string
		explicit    string
		backendName string
		globalModel string
		want        string
	}{
		{
			name:        "explicit takes priority",
			explicit:    "explicit-model",
			backendName: "claude",
			globalModel: "global-model",
			want:        "explicit-model",
		},
		{
			name:        "global model when no explicit",
			explicit:    "",
			backendName: "claude",
			globalModel: "global-model",
			want:        "global-model",
		},
		{
			name:        "empty when nothing set",
			explicit:    "",
			backendName: "claude",
			globalModel: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveModel(tt.explicit, tt.backendName, tt.globalModel)
			if got != tt.want {
				t.Errorf("resolveModel(%q, %q, %q) = %q, want %q",
					tt.explicit, tt.backendName, tt.globalModel, got, tt.want)
			}
		})
	}
}

func TestReadInputFromFileOrStdin(t *testing.T) {
	// Create a temp file for testing
	tmpDir, err := os.MkdirTemp("", "helpers-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.json")
	testContent := `{"test": "data"}`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	t.Run("read from file", func(t *testing.T) {
		data, err := readInputFromFileOrStdin(testFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != testContent {
			t.Errorf("got %q, want %q", string(data), testContent)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := readInputFromFileOrStdin("/nonexistent/file.json")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	// Note: Testing stdin reading is complex and would require redirecting stdin,
	// which is better suited for integration tests.
}

func TestCreateAndSaveSession(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := session.NewStoreWithDir(tmpDir)

	t.Run("creates session with all fields", func(t *testing.T) {
		sess := createAndSaveSession(
			store,
			"claude",
			"/test/workdir",
			"gpt-4",
			"test prompt",
			[]string{"tag1", "tag2"},
			"Test Title",
			true,
		)

		if sess == nil {
			t.Fatal("expected session to be created")
		}

		if sess.Backend != "claude" {
			t.Errorf("backend = %q, want %q", sess.Backend, "claude")
		}
		if sess.Model != "gpt-4" {
			t.Errorf("model = %q, want %q", sess.Model, "gpt-4")
		}
		if sess.InitialPrompt != "test prompt" {
			t.Errorf("prompt = %q, want %q", sess.InitialPrompt, "test prompt")
		}
		if sess.Title != "Test Title" {
			t.Errorf("title = %q, want %q", sess.Title, "Test Title")
		}
		if !sess.HasTag("tag1") || !sess.HasTag("tag2") {
			t.Error("expected tags to be set")
		}

		// Verify it was saved
		retrieved, err := store.Get(sess.ID)
		if err != nil {
			t.Fatalf("failed to retrieve saved session: %v", err)
		}
		if retrieved.ID != sess.ID {
			t.Errorf("retrieved ID = %q, want %q", retrieved.ID, sess.ID)
		}
	})

	t.Run("creates session without optional fields", func(t *testing.T) {
		sess := createAndSaveSession(
			store,
			"codex",
			"",
			"",
			"simple prompt",
			nil,
			"",
			true,
		)

		if sess == nil {
			t.Fatal("expected session to be created")
		}

		if sess.Backend != "codex" {
			t.Errorf("backend = %q, want %q", sess.Backend, "codex")
		}
	})
}

func TestShortSessionID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "long ID truncated to 8 chars",
			id:   "019c0523-e080-7fb1-a8ea-0530361cbf0f",
			want: "019c0523",
		},
		{
			name: "exactly 8 chars unchanged",
			id:   "12345678",
			want: "12345678",
		},
		{
			name: "short ID unchanged",
			id:   "abc",
			want: "abc",
		},
		{
			name: "empty ID unchanged",
			id:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortSessionID(tt.id)
			if got != tt.want {
				t.Errorf("shortSessionID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestCleanupBackendSessionWithRunner(t *testing.T) {
	t.Run("gemini calls runner with correct args", func(t *testing.T) {
		var calledName string
		var calledArgs []string
		mockRunner := func(name string, args ...string) error {
			calledName = name
			calledArgs = args
			return nil
		}

		cleanupBackendSessionWithRunner("gemini", "test-uuid-1234", mockRunner)

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

		cleanupBackendSessionWithRunner("gemini", "", mockRunner)

		if called {
			t.Error("runner should not be called when sessionID is empty")
		}
	})

	t.Run("codex skips when sessionID is empty", func(t *testing.T) {
		// This test verifies the early return for codex
		cleanupBackendSessionWithRunner("codex", "", nil)
		// Should not panic
	})

	t.Run("claude does nothing (uses native flag)", func(t *testing.T) {
		called := false
		mockRunner := func(name string, args ...string) error {
			called = true
			return nil
		}

		cleanupBackendSessionWithRunner("claude", "some-session-id", mockRunner)

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

		cleanupBackendSessionWithRunner("unknown", "some-id", mockRunner)

		if called {
			t.Error("runner should not be called for unknown backend")
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
		cleanupCodexSessionInDir(threadID, tmpDir, testTime)

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
		cleanupCodexSessionInDir("non-existent-uuid", tmpDir, testTime)

		// Verify file still exists
		if _, err := os.Stat(otherFile); os.IsNotExist(err) {
			t.Error("non-matching file should not be deleted")
		}
	})

	t.Run("handles missing directory gracefully", func(t *testing.T) {
		// Use a different time that has no directory
		differentTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

		// Should not panic
		cleanupCodexSessionInDir("some-uuid", tmpDir, differentTime)
	})

	t.Run("skips directories", func(t *testing.T) {
		// Create a subdirectory with matching name (edge case)
		subDir := filepath.Join(sessionDir, "rollout-2026-01-29T00-06-03-subdir-uuid")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}

		// Should not delete the directory
		cleanupCodexSessionInDir("subdir-uuid", tmpDir, testTime)

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
			got := containsThreadID(tt.filename, tt.threadID)
			if got != tt.want {
				t.Errorf("containsThreadID(%q, %q) = %v, want %v",
					tt.filename, tt.threadID, got, tt.want)
			}
		})
	}
}

func TestUpdateSessionAfterExecution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := session.NewStoreWithDir(tmpDir)

	t.Run("updates session on success", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")

		updateSessionAfterExecution(store, sess, 0, "", true)

		retrieved, _ := store.Get(sess.ID)
		if retrieved.Status != session.StatusCompleted {
			t.Errorf("status = %q, want %q", retrieved.Status, session.StatusCompleted)
		}
		if retrieved.TurnCount != 1 {
			t.Errorf("turn count = %d, want 1", retrieved.TurnCount)
		}
	})

	t.Run("updates session on failure", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")

		updateSessionAfterExecution(store, sess, 1, "command failed", true)

		retrieved, _ := store.Get(sess.ID)
		if retrieved.Status != session.StatusError {
			t.Errorf("status = %q, want %q", retrieved.Status, session.StatusError)
		}
		if retrieved.ErrorMessage != "command failed" {
			t.Errorf("error = %q, want %q", retrieved.ErrorMessage, "command failed")
		}
	})

	t.Run("handles nil session", func(t *testing.T) {
		// Should not panic
		updateSessionAfterExecution(store, nil, 0, "", true)
	})
}
