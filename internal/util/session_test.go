package util

import (
	"os"
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/session"
)

func TestUpdateSessionFromResponse(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "util-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := session.NewStoreWithDir(tmpDir)

	t.Run("nil session does nothing", func(t *testing.T) {
		// Should not panic
		UpdateSessionFromResponse(nil, 0, "", nil)
	})

	t.Run("success sets completed status", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")
		UpdateSessionFromResponse(sess, 0, "", nil)

		if sess.Status != session.StatusCompleted {
			t.Errorf("status = %q, want %q", sess.Status, session.StatusCompleted)
		}
		if sess.TurnCount != 1 {
			t.Errorf("turn count = %d, want 1", sess.TurnCount)
		}
	})

	t.Run("error in response sets error status", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")
		resp := &backend.UnifiedResponse{
			Error: "API error",
		}
		UpdateSessionFromResponse(sess, 0, "", resp)

		if sess.Status != session.StatusError {
			t.Errorf("status = %q, want %q", sess.Status, session.StatusError)
		}
		if sess.ErrorMessage != "API error" {
			t.Errorf("error = %q, want %q", sess.ErrorMessage, "API error")
		}
	})

	t.Run("non-zero exit code with errMsg sets error", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")
		UpdateSessionFromResponse(sess, 1, "command failed", nil)

		if sess.Status != session.StatusError {
			t.Errorf("status = %q, want %q", sess.Status, session.StatusError)
		}
		if sess.ErrorMessage != "command failed" {
			t.Errorf("error = %q, want %q", sess.ErrorMessage, "command failed")
		}
	})

	t.Run("non-zero exit code without errMsg sets generic error", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")
		UpdateSessionFromResponse(sess, 1, "", nil)

		if sess.Status != session.StatusError {
			t.Errorf("status = %q, want %q", sess.Status, session.StatusError)
		}
		if sess.ErrorMessage != "backend execution failed" {
			t.Errorf("error = %q, want %q", sess.ErrorMessage, "backend execution failed")
		}
	})

	t.Run("records token usage", func(t *testing.T) {
		sess, _ := store.Create("claude", "/tmp")
		resp := &backend.UnifiedResponse{
			Usage: &backend.TokenUsage{
				InputTokens:  100,
				OutputTokens: 200,
			},
		}
		UpdateSessionFromResponse(sess, 0, "", resp)

		if sess.TokenUsage == nil {
			t.Fatal("expected TokenUsage to be set")
		}
		if sess.TokenUsage.InputTokens != 100 {
			t.Errorf("InputTokens = %d, want 100", sess.TokenUsage.InputTokens)
		}
		if sess.TokenUsage.OutputTokens != 200 {
			t.Errorf("OutputTokens = %d, want 200", sess.TokenUsage.OutputTokens)
		}
	})
}

func TestTokenUsageFromBackend(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		result := TokenUsageFromBackend(nil)
		if result != nil {
			t.Error("expected nil")
		}
	})

	t.Run("converts usage correctly", func(t *testing.T) {
		usage := &backend.TokenUsage{
			InputTokens:  100,
			OutputTokens: 200,
			TotalTokens:  300,
		}
		result := TokenUsageFromBackend(usage)

		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.InputTokens != 100 {
			t.Errorf("InputTokens = %d, want 100", result.InputTokens)
		}
		if result.OutputTokens != 200 {
			t.Errorf("OutputTokens = %d, want 200", result.OutputTokens)
		}
	})
}
