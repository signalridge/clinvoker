package app

import (
	"os"
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/session"
)

func TestIsStaleSessionError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "empty string",
			errMsg:   "",
			expected: false,
		},
		{
			name:     "no conversation found",
			errMsg:   "No conversation found with session ID: abc123",
			expected: true,
		},
		{
			name:     "session not found",
			errMsg:   "session not found",
			expected: true,
		},
		{
			name:     "conversation not found",
			errMsg:   "conversation not found",
			expected: true,
		},
		{
			name:     "invalid session",
			errMsg:   "invalid session",
			expected: true,
		},
		{
			name:     "session has expired",
			errMsg:   "session has expired",
			expected: true,
		},
		{
			name:     "case insensitive match",
			errMsg:   "NO CONVERSATION FOUND WITH SESSION ID",
			expected: true,
		},
		{
			name:     "connection timeout - not stale",
			errMsg:   "connection timeout",
			expected: false,
		},
		{
			name:     "rate limit exceeded - not stale",
			errMsg:   "rate limit exceeded",
			expected: false,
		},
		{
			name:     "general error - not stale",
			errMsg:   "internal server error",
			expected: false,
		},
		{
			name:     "partial match in longer message",
			errMsg:   "Error: No conversation found with session ID abc123. Please start a new session.",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStaleSessionError(tt.errMsg)
			if got != tt.expected {
				t.Errorf("isStaleSessionError(%q) = %v, want %v", tt.errMsg, got, tt.expected)
			}
		})
	}
}

func TestGetResumableSessionsExcluding(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stale-session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := session.NewStoreWithDir(tmpDir)

	// Create some sessions
	sess1, _ := store.Create("claude", "/tmp")
	sess1.BackendSessionID = "backend-1"
	store.Save(sess1)

	sess2, _ := store.Create("claude", "/tmp")
	sess2.BackendSessionID = "backend-2"
	store.Save(sess2)

	sess3, _ := store.Create("codex", "/tmp")
	sess3.BackendSessionID = "backend-3"
	store.Save(sess3)

	// Create an expired session
	sess4, _ := store.Create("claude", "/tmp")
	sess4.BackendSessionID = "backend-4"
	sess4.MarkExpired()
	store.Save(sess4)

	t.Run("excludes specified session ID", func(t *testing.T) {
		result := getResumableSessionsExcluding(store, sess1.ID)

		// Should have sess2 and sess3, but not sess1 or sess4 (expired)
		if len(result) != 2 {
			t.Errorf("expected 2 sessions, got %d", len(result))
		}

		for _, s := range result {
			if s.ID == sess1.ID {
				t.Error("should not include excluded session ID")
			}
			if s.ID == sess4.ID {
				t.Error("should not include expired session")
			}
		}
	})

	t.Run("excludes expired sessions", func(t *testing.T) {
		result := getResumableSessionsExcluding(store, "nonexistent")

		for _, s := range result {
			if s.Status == session.StatusExpired {
				t.Error("should not include expired sessions")
			}
		}
	})
}

func TestStaleSessionChoice_Constants(t *testing.T) {
	// Verify constants are distinct
	if StaleSessionNew == StaleSessionSelect {
		t.Error("StaleSessionNew should not equal StaleSessionSelect")
	}
	if StaleSessionSelect == StaleSessionCancel {
		t.Error("StaleSessionSelect should not equal StaleSessionCancel")
	}
	if StaleSessionNew == StaleSessionCancel {
		t.Error("StaleSessionNew should not equal StaleSessionCancel")
	}
}

func TestStaleSessionPatterns(t *testing.T) {
	// Verify all patterns are lowercase for case-insensitive matching
	for _, pattern := range staleSessionPatterns {
		if pattern != strings.ToLower(pattern) {
			t.Errorf("pattern %q should be lowercase for case-insensitive matching", pattern)
		}
	}

	// Verify we have reasonable coverage of common error messages
	expectedPatterns := []string{
		"session not found",
		"conversation not found",
		"invalid session",
	}
	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range staleSessionPatterns {
			if strings.Contains(pattern, expected) || strings.Contains(expected, pattern) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected pattern containing %q not found in staleSessionPatterns", expected)
		}
	}
}

func TestGetResumableSessionsExcluding_EmptyStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stale-session-empty-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := session.NewStoreWithDir(tmpDir)

	result := getResumableSessionsExcluding(store, "any-id")
	if len(result) != 0 {
		t.Errorf("expected 0 sessions from empty store, got %d", len(result))
	}
}

func TestGetResumableSessionsExcluding_AllExpired(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stale-session-expired-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := session.NewStoreWithDir(tmpDir)

	// Create only expired sessions
	for i := 0; i < 3; i++ {
		sess, _ := store.Create("claude", "/tmp")
		sess.BackendSessionID = "backend-" + string(rune('a'+i))
		sess.MarkExpired()
		store.Save(sess)
	}

	result := getResumableSessionsExcluding(store, "nonexistent")
	if len(result) != 0 {
		t.Errorf("expected 0 resumable sessions when all are expired, got %d", len(result))
	}
}

func TestStaleSessionContext_Fields(t *testing.T) {
	// Verify staleSessionContext can be created with all required fields
	tmpDir, err := os.MkdirTemp("", "stale-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := session.NewStoreWithDir(tmpDir)
	sess, _ := store.Create("claude", "/tmp")

	ctx := &staleSessionContext{
		store:   store,
		sess:    sess,
		prompt:  "test prompt",
		flags:   &normalizedFlags{outputFormat: "json", dryRun: false},
		backend: nil, // Would be set in real usage
		opts:    nil, // Would be set in real usage
	}

	if ctx.store == nil {
		t.Error("store should not be nil")
	}
	if ctx.sess == nil {
		t.Error("sess should not be nil")
	}
	if ctx.prompt != "test prompt" {
		t.Errorf("prompt = %q, want %q", ctx.prompt, "test prompt")
	}
	if ctx.flags.outputFormat != "json" {
		t.Errorf("flags.outputFormat = %q, want %q", ctx.flags.outputFormat, "json")
	}
}
