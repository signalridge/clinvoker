package service

import (
	"context"
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/output"
	"github.com/signalridge/clinvoker/internal/session"
)

func TestStreamPrompt_InvalidBackend(t *testing.T) {
	req := &PromptRequest{
		Backend: "invalid-backend",
		Prompt:  "test",
	}

	result, err := StreamPrompt(context.Background(), req, nil, true, nil)
	if err == nil {
		t.Fatal("expected error for invalid backend")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestStreamPrompt_DryRun(t *testing.T) {
	// Skip if no backend available
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	req := &PromptRequest{
		Backend: "claude",
		Prompt:  "test prompt",
		DryRun:  true,
	}

	var eventCount int
	result, err := StreamPrompt(context.Background(), req, nil, true, func(event *output.UnifiedEvent) error {
		eventCount++
		return nil
	})

	// Dry run returns early with special result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Dry run might not produce stream events, that's ok
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestStreamPrompt_NilCallback(t *testing.T) {
	// Skip if no backend available
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	req := &PromptRequest{
		Backend: "claude",
		Prompt:  "test prompt",
		DryRun:  true,
	}

	// nil callback should be handled gracefully
	result, err := StreamPrompt(context.Background(), req, nil, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestStreamPrompt_ContextCancellation(t *testing.T) {
	// Skip if no backend available
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &PromptRequest{
		Backend: "claude",
		Prompt:  "test prompt",
	}

	// Canceled context should not cause panic
	_, _ = StreamPrompt(ctx, req, nil, true, nil)
	// We don't care about the result, just that it doesn't panic
}

func TestStreamPrompt_ForceStateless(t *testing.T) {
	// Skip if no backend available
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	req := &PromptRequest{
		Backend: "claude",
		Prompt:  "test prompt",
		DryRun:  true,
	}

	// forceStateless = true should work
	result, err := StreamPrompt(context.Background(), req, nil, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestStreamResult_Structure(t *testing.T) {
	result := StreamResult{
		ExitCode:         0,
		Error:            "",
		TokenUsage:       &session.TokenUsage{InputTokens: 100, OutputTokens: 50},
		BackendSessionID: "session-123",
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Error != "" {
		t.Errorf("Error = %q, want empty", result.Error)
	}
	if result.TokenUsage.InputTokens != 100 {
		t.Errorf("TokenUsage.InputTokens = %d, want 100", result.TokenUsage.InputTokens)
	}
	if result.BackendSessionID != "session-123" {
		t.Errorf("BackendSessionID = %q, want 'session-123'", result.BackendSessionID)
	}
}

func TestStreamResult_WithError(t *testing.T) {
	result := StreamResult{
		ExitCode:   1,
		Error:      "command failed",
		TokenUsage: nil,
	}

	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
	if result.Error != "command failed" {
		t.Errorf("Error = %q, want 'command failed'", result.Error)
	}
	if result.TokenUsage != nil {
		t.Error("TokenUsage should be nil")
	}
}
