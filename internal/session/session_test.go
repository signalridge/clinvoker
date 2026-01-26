package session

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	t.Run("creates session with generated ID", func(t *testing.T) {
		sess, err := NewSession("claude", "/test/dir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if sess.ID == "" {
			t.Error("expected non-empty ID")
		}
		if len(sess.ID) != 16 { // 8 bytes = 16 hex chars
			t.Errorf("expected ID length 16, got %d", len(sess.ID))
		}
		if sess.Backend != "claude" {
			t.Errorf("expected backend 'claude', got %q", sess.Backend)
		}
		if sess.WorkingDir != "/test/dir" {
			t.Errorf("expected workdir '/test/dir', got %q", sess.WorkingDir)
		}
	})

	t.Run("sets timestamps on creation", func(t *testing.T) {
		before := time.Now()
		sess, _ := NewSession("codex", "/tmp")
		after := time.Now()

		if sess.CreatedAt.Before(before) || sess.CreatedAt.After(after) {
			t.Error("CreatedAt not set correctly")
		}
		if sess.LastUsed.Before(before) || sess.LastUsed.After(after) {
			t.Error("LastUsed not set correctly")
		}
	})

	t.Run("initializes empty metadata", func(t *testing.T) {
		sess, _ := NewSession("gemini", "/tmp")

		if sess.Metadata == nil {
			t.Error("expected non-nil metadata map")
		}
	})
}

func TestSession_MarkUsed(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")
	originalTime := sess.LastUsed

	time.Sleep(10 * time.Millisecond)
	sess.MarkUsed()

	if !sess.LastUsed.After(originalTime) {
		t.Error("LastUsed should be updated after MarkUsed")
	}
}

func TestSession_Age(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")
	time.Sleep(10 * time.Millisecond)

	age := sess.Age()
	if age < 10*time.Millisecond {
		t.Errorf("expected age >= 10ms, got %v", age)
	}
}

func TestSession_IdleDuration(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")
	time.Sleep(10 * time.Millisecond)

	idle := sess.IdleDuration()
	if idle < 10*time.Millisecond {
		t.Errorf("expected idle >= 10ms, got %v", idle)
	}

	sess.MarkUsed()
	idle = sess.IdleDuration()
	if idle > 5*time.Millisecond {
		t.Errorf("expected idle < 5ms after MarkUsed, got %v", idle)
	}
}

func TestSession_SetBackendSessionID(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")
	sess.SetBackendSessionID("backend-123")

	if sess.BackendSessionID != "backend-123" {
		t.Errorf("expected 'backend-123', got %q", sess.BackendSessionID)
	}
}

func TestSession_SetMetadata(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.SetMetadata("key1", "value1")
	sess.SetMetadata("key2", "value2")

	if sess.Metadata["key1"] != "value1" {
		t.Errorf("expected 'value1', got %q", sess.Metadata["key1"])
	}
	if sess.Metadata["key2"] != "value2" {
		t.Errorf("expected 'value2', got %q", sess.Metadata["key2"])
	}
}

func TestGenerateID(t *testing.T) {
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id, err := generateID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ids[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		ids[id] = true

		if len(id) != 16 {
			t.Errorf("expected ID length 16, got %d", len(id))
		}
	}
}

// ==================== TokenUsage Tests ====================

func TestTokenUsage_Total(t *testing.T) {
	tu := &TokenUsage{
		InputTokens:  100,
		OutputTokens: 200,
	}

	if tu.Total() != 300 {
		t.Errorf("expected Total() = 300, got %d", tu.Total())
	}
}

func TestSession_AddTokens(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.AddTokens(100, 200)
	if sess.TokenUsage.InputTokens != 100 {
		t.Errorf("expected InputTokens = 100, got %d", sess.TokenUsage.InputTokens)
	}
	if sess.TokenUsage.OutputTokens != 200 {
		t.Errorf("expected OutputTokens = 200, got %d", sess.TokenUsage.OutputTokens)
	}

	// Test accumulation
	sess.AddTokens(50, 75)
	if sess.TokenUsage.InputTokens != 150 {
		t.Errorf("expected InputTokens = 150, got %d", sess.TokenUsage.InputTokens)
	}
	if sess.TokenUsage.OutputTokens != 275 {
		t.Errorf("expected OutputTokens = 275, got %d", sess.TokenUsage.OutputTokens)
	}
}

func TestSession_AddTokens_NilUsage(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")
	sess.TokenUsage = nil // Force nil

	sess.AddTokens(100, 200)
	if sess.TokenUsage == nil {
		t.Error("expected TokenUsage to be initialized")
	}
	if sess.TokenUsage.InputTokens != 100 {
		t.Errorf("expected InputTokens = 100, got %d", sess.TokenUsage.InputTokens)
	}
}

func TestSession_AddCachedTokens(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.AddCachedTokens(500)
	if sess.TokenUsage.CachedTokens != 500 {
		t.Errorf("expected CachedTokens = 500, got %d", sess.TokenUsage.CachedTokens)
	}

	sess.AddCachedTokens(200)
	if sess.TokenUsage.CachedTokens != 700 {
		t.Errorf("expected CachedTokens = 700, got %d", sess.TokenUsage.CachedTokens)
	}
}

func TestSession_AddCachedTokens_NilUsage(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")
	sess.TokenUsage = nil

	sess.AddCachedTokens(100)
	if sess.TokenUsage == nil {
		t.Error("expected TokenUsage to be initialized")
	}
}

func TestSession_AddReasoningTokens(t *testing.T) {
	sess, _ := NewSession("codex", "/tmp")

	sess.AddReasoningTokens(1000)
	if sess.TokenUsage.ReasoningTokens != 1000 {
		t.Errorf("expected ReasoningTokens = 1000, got %d", sess.TokenUsage.ReasoningTokens)
	}

	sess.AddReasoningTokens(500)
	if sess.TokenUsage.ReasoningTokens != 1500 {
		t.Errorf("expected ReasoningTokens = 1500, got %d", sess.TokenUsage.ReasoningTokens)
	}
}

func TestSession_AddReasoningTokens_NilUsage(t *testing.T) {
	sess, _ := NewSession("codex", "/tmp")
	sess.TokenUsage = nil

	sess.AddReasoningTokens(100)
	if sess.TokenUsage == nil {
		t.Error("expected TokenUsage to be initialized")
	}
}

// ==================== Status Tests ====================

func TestSession_SetStatus(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	if sess.Status != StatusActive {
		t.Errorf("expected initial status %q, got %q", StatusActive, sess.Status)
	}

	sess.SetStatus(StatusPaused)
	if sess.Status != StatusPaused {
		t.Errorf("expected status %q, got %q", StatusPaused, sess.Status)
	}

	sess.SetStatus(StatusCompleted)
	if sess.Status != StatusCompleted {
		t.Errorf("expected status %q, got %q", StatusCompleted, sess.Status)
	}
}

func TestSession_SetError(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.SetError("something went wrong")

	if sess.Status != StatusError {
		t.Errorf("expected status %q, got %q", StatusError, sess.Status)
	}
	if sess.ErrorMessage != "something went wrong" {
		t.Errorf("expected error message %q, got %q", "something went wrong", sess.ErrorMessage)
	}
}

func TestSession_Complete(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.Complete()

	if sess.Status != StatusCompleted {
		t.Errorf("expected status %q, got %q", StatusCompleted, sess.Status)
	}
}

// ==================== Turn Count Tests ====================

func TestSession_IncrementTurn(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	if sess.TurnCount != 0 {
		t.Errorf("expected initial TurnCount = 0, got %d", sess.TurnCount)
	}

	sess.IncrementTurn()
	if sess.TurnCount != 1 {
		t.Errorf("expected TurnCount = 1, got %d", sess.TurnCount)
	}

	sess.IncrementTurn()
	sess.IncrementTurn()
	if sess.TurnCount != 3 {
		t.Errorf("expected TurnCount = 3, got %d", sess.TurnCount)
	}
}

// ==================== Tag Tests ====================

func TestSession_AddTag(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.AddTag("important")
	if len(sess.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(sess.Tags))
	}
	if sess.Tags[0] != "important" {
		t.Errorf("expected tag 'important', got %q", sess.Tags[0])
	}

	// Add another tag
	sess.AddTag("debug")
	if len(sess.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(sess.Tags))
	}

	// Try to add duplicate
	sess.AddTag("important")
	if len(sess.Tags) != 2 {
		t.Errorf("expected 2 tags (no duplicate), got %d", len(sess.Tags))
	}
}

func TestSession_RemoveTag(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.AddTag("tag1")
	sess.AddTag("tag2")
	sess.AddTag("tag3")

	sess.RemoveTag("tag2")
	if len(sess.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(sess.Tags))
	}
	if sess.HasTag("tag2") {
		t.Error("tag2 should have been removed")
	}

	// Remove non-existent tag (should not panic)
	sess.RemoveTag("nonexistent")
	if len(sess.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(sess.Tags))
	}
}

func TestSession_HasTag(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.AddTag("test-tag")

	if !sess.HasTag("test-tag") {
		t.Error("expected HasTag('test-tag') = true")
	}
	if sess.HasTag("other-tag") {
		t.Error("expected HasTag('other-tag') = false")
	}
}

// ==================== Title and Model Tests ====================

func TestSession_SetTitle(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.SetTitle("My Important Session")
	if sess.Title != "My Important Session" {
		t.Errorf("expected title 'My Important Session', got %q", sess.Title)
	}
}

func TestSession_SetModel(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")

	sess.SetModel("opus")
	if sess.Model != "opus" {
		t.Errorf("expected model 'opus', got %q", sess.Model)
	}
}

// ==================== DisplayName Tests ====================

func TestSession_DisplayName(t *testing.T) {
	t.Run("returns title if set", func(t *testing.T) {
		sess, _ := NewSession("claude", "/tmp")
		sess.SetTitle("My Session")

		if sess.DisplayName() != "My Session" {
			t.Errorf("expected 'My Session', got %q", sess.DisplayName())
		}
	})

	t.Run("returns prompt if no title", func(t *testing.T) {
		sess, _ := NewSession("claude", "/tmp")
		sess.InitialPrompt = "Fix the bug in auth.go"

		if sess.DisplayName() != "Fix the bug in auth.go" {
			t.Errorf("expected 'Fix the bug in auth.go', got %q", sess.DisplayName())
		}
	})

	t.Run("truncates long prompt", func(t *testing.T) {
		sess, _ := NewSession("claude", "/tmp")
		sess.InitialPrompt = "This is a very long prompt that should be truncated because it exceeds fifty characters"

		display := sess.DisplayName()
		if len(display) > 50 {
			t.Errorf("expected truncated display name, got length %d", len(display))
		}
		if display[len(display)-3:] != "..." {
			t.Errorf("expected '...' at end, got %q", display)
		}
	})

	t.Run("returns short ID if no title or prompt", func(t *testing.T) {
		sess, _ := NewSession("claude", "/tmp")

		display := sess.DisplayName()
		if len(display) != 8 {
			t.Errorf("expected 8-char short ID, got length %d: %q", len(display), display)
		}
		if display != sess.ID[:8] {
			t.Errorf("expected short ID %q, got %q", sess.ID[:8], display)
		}
	})
}

// ==================== Fork Tests ====================

func TestSession_Fork(t *testing.T) {
	sess, _ := NewSession("claude", "/tmp")
	sess.SetModel("opus")
	sess.AddTag("original")
	sess.SetMetadata("key", "value")

	forked, err := sess.Fork()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify forked session
	if forked.ID == sess.ID {
		t.Error("forked session should have different ID")
	}
	if forked.Backend != sess.Backend {
		t.Errorf("expected backend %q, got %q", sess.Backend, forked.Backend)
	}
	if forked.WorkingDir != sess.WorkingDir {
		t.Errorf("expected workdir %q, got %q", sess.WorkingDir, forked.WorkingDir)
	}
	if forked.Model != sess.Model {
		t.Errorf("expected model %q, got %q", sess.Model, forked.Model)
	}
	if forked.ParentID != sess.ID {
		t.Errorf("expected parent ID %q, got %q", sess.ID, forked.ParentID)
	}
	if !forked.HasTag("original") {
		t.Error("forked session should have copied tags")
	}
	if forked.Metadata["key"] != "value" {
		t.Error("forked session should have copied metadata")
	}

	// Verify tags are independent
	sess.AddTag("new-tag")
	if forked.HasTag("new-tag") {
		t.Error("forked tags should be independent")
	}
}

// ==================== NewSessionWithOptions Tests ====================

func TestNewSessionWithOptions(t *testing.T) {
	opts := &SessionOptions{
		Model:         "opus",
		InitialPrompt: "Test prompt",
		Title:         "Test Title",
		Tags:          []string{"tag1", "tag2"},
		ParentID:      "parent-123",
	}

	sess, err := NewSessionWithOptions("claude", "/project", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sess.Model != "opus" {
		t.Errorf("expected model 'opus', got %q", sess.Model)
	}
	if sess.InitialPrompt != "Test prompt" {
		t.Errorf("expected prompt 'Test prompt', got %q", sess.InitialPrompt)
	}
	if sess.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", sess.Title)
	}
	if len(sess.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(sess.Tags))
	}
	if sess.ParentID != "parent-123" {
		t.Errorf("expected parent ID 'parent-123', got %q", sess.ParentID)
	}
	if sess.Status != StatusActive {
		t.Errorf("expected status %q, got %q", StatusActive, sess.Status)
	}
	if sess.TokenUsage == nil {
		t.Error("expected TokenUsage to be initialized")
	}
}

func TestNewSessionWithOptions_NilOpts(t *testing.T) {
	sess, err := NewSessionWithOptions("claude", "/project", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sess.Model != "" {
		t.Errorf("expected empty model, got %q", sess.Model)
	}
	if sess.InitialPrompt != "" {
		t.Errorf("expected empty prompt, got %q", sess.InitialPrompt)
	}
}
