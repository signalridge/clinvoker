package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/server/service"
)

func TestValidatePromptRequest_Errors(t *testing.T) {
	if err := ValidatePromptRequest(PromptRequest{}); err == nil {
		t.Fatal("expected error for missing backend and prompt")
	}
	if err := ValidatePromptRequest(PromptRequest{Backend: "claude"}); err == nil {
		t.Fatal("expected error for missing prompt")
	}
}

func TestValidateParallelRequest_Errors(t *testing.T) {
	if err := ValidateParallelRequest(ParallelRequest{}); err == nil {
		t.Fatal("expected error for missing tasks")
	}
}

func TestValidateChainRequest_Errors(t *testing.T) {
	if err := ValidateChainRequest(ChainRequest{}); err == nil {
		t.Fatal("expected error for missing steps")
	}

	bad := ChainRequest{
		Steps:         []ChainStep{{Backend: "claude", Prompt: "hello"}},
		PassSessionID: true,
	}
	if err := ValidateChainRequest(bad); err == nil {
		t.Fatal("expected error for pass_session_id")
	}

	bad = ChainRequest{
		Steps: []ChainStep{{Backend: "claude", Prompt: "{{session}}"}},
	}
	if err := ValidateChainRequest(bad); err == nil {
		t.Fatal("expected error for {{session}} placeholder")
	}
}

func TestValidateCompareRequest_Errors(t *testing.T) {
	if err := ValidateCompareRequest(CompareRequest{}); err == nil {
		t.Fatal("expected error for missing backends and prompt")
	}
	if err := ValidateCompareRequest(CompareRequest{Backends: []string{"claude"}}); err == nil {
		t.Fatal("expected error for missing prompt")
	}
}

func TestHandlePrompt_ValidationError(t *testing.T) {
	h := NewCustomHandlers(service.NewExecutor())

	_, err := h.HandlePrompt(context.Background(), &PromptInput{Body: PromptRequest{}})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "backend is required") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "backend is required")
	}
}

func TestFromServiceResult_ErrorFields(t *testing.T) {
	resp := FromServiceResult(&service.PromptResult{
		Backend:  "claude",
		ExitCode: 1,
		Error:    "boom",
	})
	if resp.Error != "boom" {
		t.Fatalf("Error = %q, want %q", resp.Error, "boom")
	}
	if resp.ExitCode != 1 {
		t.Fatalf("ExitCode = %d, want %d", resp.ExitCode, 1)
	}
}
