package handlers

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/server/service"
)

func TestNewCustomHandlers(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)

	if handlers == nil {
		t.Error("NewCustomHandlers returned nil")
	}
	if handlers.executor == nil {
		t.Error("executor not set")
	}
}

func TestNewOpenAIHandlers(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewOpenAIHandlers(executor)

	if handlers == nil {
		t.Error("NewOpenAIHandlers returned nil")
	}
	if handlers.runner == nil {
		t.Error("runner not set")
	}
}

func TestNewAnthropicHandlers(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewAnthropicHandlers(executor)

	if handlers == nil {
		t.Error("NewAnthropicHandlers returned nil")
	}
	if handlers.runner == nil {
		t.Error("runner not set")
	}
}

func TestOpenAPIStreamingResponses(t *testing.T) {
	api := humachi.New(chi.NewRouter(), huma.DefaultConfig("test", "1.0"))

	openaiHandlers := NewOpenAIHandlers(service.NewStatelessRunner(nil))
	openaiHandlers.Register(api)

	anthropicHandlers := NewAnthropicHandlers(service.NewStatelessRunner(nil))
	anthropicHandlers.Register(api)

	openaiOp := api.OpenAPI().Paths["/openai/v1/chat/completions"].Post
	if openaiOp == nil || openaiOp.Responses == nil || openaiOp.Responses["200"] == nil {
		t.Fatal("openai chat completions response schema missing")
	}
	if _, ok := openaiOp.Responses["200"].Content["text/event-stream"]; !ok {
		t.Fatal("openai chat completions missing text/event-stream response")
	}
	if _, ok := openaiOp.Responses["200"].Content["application/json"]; !ok {
		t.Fatal("openai chat completions missing application/json response")
	}

	anthropicOp := api.OpenAPI().Paths["/anthropic/v1/messages"].Post
	if anthropicOp == nil || anthropicOp.Responses == nil || anthropicOp.Responses["200"] == nil {
		t.Fatal("anthropic messages response schema missing")
	}
	if _, ok := anthropicOp.Responses["200"].Content["text/event-stream"]; !ok {
		t.Fatal("anthropic messages missing text/event-stream response")
	}
	if _, ok := anthropicOp.Responses["200"].Content["application/json"]; !ok {
		t.Fatal("anthropic messages missing application/json response")
	}
}

func TestMapModelToBackend(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{"direct claude", "claude", backend.BackendClaude},
		{"direct codex", "codex", backend.BackendCodex},
		{"direct gemini", "gemini", backend.BackendGemini},
		{"claude model name", "claude-opus-4-5", backend.BackendClaude},
		{"gpt model name", "gpt-4-turbo", backend.BackendCodex},
		{"gemini model name", "gemini-pro", backend.BackendGemini},
		{"unknown defaults to claude", "some-unknown-model", backend.BackendClaude},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapModelToBackend(tt.model)
			if result != tt.expected {
				t.Errorf("mapModelToBackend(%q) = %q, want %q", tt.model, result, tt.expected)
			}
		})
	}
}

func TestMapAnthropicModelToBackend(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{"direct claude", "claude", backend.BackendClaude},
		{"direct codex", "codex", backend.BackendCodex},
		{"direct gemini", "gemini", backend.BackendGemini},
		{"claude model name", "claude-3-opus", backend.BackendClaude},
		{"unknown defaults to claude", "some-model", backend.BackendClaude},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapAnthropicModelToBackend(tt.model)
			if result != tt.expected {
				t.Errorf("mapAnthropicModelToBackend(%q) = %q, want %q", tt.model, result, tt.expected)
			}
		})
	}
}

func TestPromptRequestToServiceRequest(t *testing.T) {
	req := &PromptRequest{
		Backend:      "claude",
		Prompt:       "test prompt",
		Model:        "claude-opus",
		WorkDir:      "/tmp",
		ApprovalMode: "auto",
		SandboxMode:  "workspace",
		OutputFormat: "json",
		MaxTokens:    1000,
		MaxTurns:     5,
		SystemPrompt: "You are helpful",
		Verbose:      true,
		DryRun:       true,
		Extra:        []string{"--flag"},
		Metadata:     map[string]string{"key": "value"},
	}

	svcReq := req.ToServiceRequest()

	if svcReq.Backend != req.Backend {
		t.Errorf("Backend mismatch: got %q, want %q", svcReq.Backend, req.Backend)
	}
	if svcReq.Prompt != req.Prompt {
		t.Errorf("Prompt mismatch: got %q, want %q", svcReq.Prompt, req.Prompt)
	}
	if svcReq.Model != req.Model {
		t.Errorf("Model mismatch: got %q, want %q", svcReq.Model, req.Model)
	}
	if svcReq.WorkDir != req.WorkDir {
		t.Errorf("WorkDir mismatch: got %q, want %q", svcReq.WorkDir, req.WorkDir)
	}
	if svcReq.ApprovalMode != req.ApprovalMode {
		t.Errorf("ApprovalMode mismatch: got %q, want %q", svcReq.ApprovalMode, req.ApprovalMode)
	}
	if svcReq.MaxTokens != req.MaxTokens {
		t.Errorf("MaxTokens mismatch: got %d, want %d", svcReq.MaxTokens, req.MaxTokens)
	}
	if svcReq.DryRun != req.DryRun {
		t.Errorf("DryRun mismatch: got %v, want %v", svcReq.DryRun, req.DryRun)
	}
}

func TestFromServiceResult(t *testing.T) {
	svcResult := &service.PromptResult{
		SessionID:  "test-session-123",
		Backend:    "claude",
		ExitCode:   0,
		DurationMS: 1500,
		Output:     "test output",
		Error:      "",
	}

	resp := FromServiceResult(svcResult)

	if resp.SessionID != svcResult.SessionID {
		t.Errorf("SessionID mismatch: got %q, want %q", resp.SessionID, svcResult.SessionID)
	}
	if resp.Backend != svcResult.Backend {
		t.Errorf("Backend mismatch: got %q, want %q", resp.Backend, svcResult.Backend)
	}
	if resp.ExitCode != svcResult.ExitCode {
		t.Errorf("ExitCode mismatch: got %d, want %d", resp.ExitCode, svcResult.ExitCode)
	}
	if resp.DurationMS != svcResult.DurationMS {
		t.Errorf("DurationMS mismatch: got %d, want %d", resp.DurationMS, svcResult.DurationMS)
	}
	if resp.Output != svcResult.Output {
		t.Errorf("Output mismatch: got %q, want %q", resp.Output, svcResult.Output)
	}
}

func TestHandleBackends(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)

	// Test that HandleBackends returns backends
	resp, err := handlers.HandleBackends(context.Background(), &BackendsInput{})
	if err != nil {
		t.Fatalf("HandleBackends failed: %v", err)
	}

	if len(resp.Body.Backends) == 0 {
		t.Error("expected at least one backend")
	}

	// Verify backend structure
	for _, b := range resp.Body.Backends {
		if b.Name == "" {
			t.Error("backend name should not be empty")
		}
	}
}

func TestHandleHealth(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)

	resp, err := handlers.HandleHealth(context.Background(), &HealthInput{})
	if err != nil {
		t.Fatalf("HandleHealth failed: %v", err)
	}

	if resp.Body.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Body.Status)
	}
}

func TestHandleSessions(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)

	// Test that HandleSessions returns a response with pagination metadata
	resp, err := handlers.HandleSessions(context.Background(), &SessionsInput{})
	if err != nil {
		t.Fatalf("HandleSessions failed: %v", err)
	}

	// Response should be valid (Total >= 0, Sessions slice not nil)
	if resp.Body.Total < 0 {
		t.Errorf("total should be non-negative, got %d", resp.Body.Total)
	}
	if resp.Body.Sessions == nil {
		t.Error("sessions should not be nil")
	}
}

func TestHandleSessionsWithPagination(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)

	// Test with pagination parameters
	resp, err := handlers.HandleSessions(context.Background(), &SessionsInput{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("HandleSessions with pagination failed: %v", err)
	}

	if resp.Body.Limit != 10 {
		t.Errorf("expected limit 10, got %d", resp.Body.Limit)
	}
	if resp.Body.Offset != 0 {
		t.Errorf("expected offset 0, got %d", resp.Body.Offset)
	}
}

func TestHandleSessionsWithFilters(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)

	// Test with filter parameters
	resp, err := handlers.HandleSessions(context.Background(), &SessionsInput{
		Backend: "claude",
		Status:  "active",
		Limit:   5,
		Offset:  10,
	})
	if err != nil {
		t.Fatalf("HandleSessions with filters failed: %v", err)
	}

	// Verify pagination metadata is returned
	if resp.Body.Limit != 5 {
		t.Errorf("expected limit 5, got %d", resp.Body.Limit)
	}
	if resp.Body.Offset != 10 {
		t.Errorf("expected offset 10, got %d", resp.Body.Offset)
	}
}
