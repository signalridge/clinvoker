package app

import (
	"os"
	"path/filepath"
	"testing"

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

// Note: Tests for cleanup functions (cleanupBackendSession, cleanupCodexSession, etc.)
// are now in internal/util/cleanup_test.go since those functions were moved to the util package.

func TestUpdateSessionAfterExecutionWithBackendID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := session.NewStoreWithDir(tmpDir)

	t.Run("updates backend session ID", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")

		updateSessionAfterExecutionWithBackendID(store, sess, 0, "", "backend-123", true)

		retrieved, _ := store.Get(sess.ID)
		if retrieved.BackendSessionID != "backend-123" {
			t.Errorf("backend session ID = %q, want %q", retrieved.BackendSessionID, "backend-123")
		}
	})

	t.Run("handles empty backend session ID", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")

		updateSessionAfterExecutionWithBackendID(store, sess, 0, "", "", true)

		retrieved, _ := store.Get(sess.ID)
		if retrieved.BackendSessionID != "" {
			t.Errorf("backend session ID = %q, want empty", retrieved.BackendSessionID)
		}
	})

	t.Run("handles nil session", func(t *testing.T) {
		// Should not panic
		updateSessionAfterExecutionWithBackendID(store, nil, 0, "", "test-id", true)
	})
}
