package service

import (
	"context"
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
)

func TestNewStatefulRunner(t *testing.T) {
	runner := NewStatefulRunner(nil, nil)
	if runner == nil {
		t.Fatal("NewStatefulRunner returned nil")
	}
	if runner.logger == nil {
		t.Error("logger should default to slog.Default")
	}
}

func TestNewStatelessRunner(t *testing.T) {
	runner := NewStatelessRunner(nil)
	if runner == nil {
		t.Fatal("NewStatelessRunner returned nil")
	}
	if runner.logger == nil {
		t.Error("logger should default to slog.Default")
	}
}

func TestStatelessRunner_ExecutePrompt_InvalidBackend(t *testing.T) {
	runner := NewStatelessRunner(nil)

	req := &PromptRequest{
		Backend: "invalid-backend",
		Prompt:  "test",
	}

	result, err := runner.ExecutePrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return result with error, not fail
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
	if result.Error == "" {
		t.Error("expected error message for invalid backend")
	}
	// Stateless should not return session ID
	if result.SessionID != "" {
		t.Errorf("stateless runner should not return session ID, got %q", result.SessionID)
	}
}

func TestStatelessRunner_ExecutePrompt_DryRun(t *testing.T) {
	// Skip if no backend available
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	runner := NewStatelessRunner(nil)

	req := &PromptRequest{
		Backend: "claude",
		Prompt:  "test prompt",
		DryRun:  true,
	}

	result, err := runner.ExecutePrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0 for dry run, got %d", result.ExitCode)
	}
	if result.Output == "" {
		t.Error("expected output for dry run")
	}
	// Stateless should not return session ID
	if result.SessionID != "" {
		t.Errorf("stateless runner should not return session ID, got %q", result.SessionID)
	}
}

func TestStatefulRunner_ExecutePrompt_InvalidBackend(t *testing.T) {
	runner := NewStatefulRunner(nil, nil)

	req := &PromptRequest{
		Backend: "invalid-backend",
		Prompt:  "test",
	}

	result, err := runner.ExecutePrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return result with error
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
	if result.Error == "" {
		t.Error("expected error message for invalid backend")
	}
}

func TestStatefulRunner_ExecutePrompt_DryRun(t *testing.T) {
	// Skip if no backend available
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	runner := NewStatefulRunner(nil, nil)

	req := &PromptRequest{
		Backend: "claude",
		Prompt:  "test prompt",
		DryRun:  true,
	}

	result, err := runner.ExecutePrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0 for dry run, got %d", result.ExitCode)
	}
	if result.Output == "" {
		t.Error("expected output for dry run")
	}
}

func TestStatelessRunner_ExecutePrompt_Ephemeral(t *testing.T) {
	// Skip if no backend available
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	runner := NewStatelessRunner(nil)

	req := &PromptRequest{
		Backend:   "claude",
		Prompt:    "test ephemeral",
		DryRun:    true,
		Ephemeral: true,
	}

	result, err := runner.ExecutePrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stateless + ephemeral should not return a session ID
	if result.SessionID != "" {
		t.Errorf("expected no session ID for stateless ephemeral, got %q", result.SessionID)
	}
}

func TestStatefulRunner_ExecutePrompt_WithModel(t *testing.T) {
	// Skip if no backend available
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	runner := NewStatefulRunner(nil, nil)

	req := &PromptRequest{
		Backend: "claude",
		Prompt:  "test prompt",
		Model:   "claude-opus",
		DryRun:  true,
	}

	result, err := runner.ExecutePrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestPromptRunner_Interface(t *testing.T) {
	// Verify both runners implement PromptRunner interface
	var _ PromptRunner = (*StatefulRunner)(nil)
	var _ PromptRunner = (*StatelessRunner)(nil)
}
