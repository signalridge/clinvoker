package handlers

import (
	"context"
	"testing"
	"time"

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
	handlers := NewOpenAIHandlers(executor, nil)

	if handlers == nil {
		t.Error("NewOpenAIHandlers returned nil")
	}
	if handlers.runner == nil {
		t.Error("runner not set")
	}
	if handlers.logger == nil {
		t.Error("logger not set (should default to slog.Default)")
	}
}

func TestNewAnthropicHandlers(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewAnthropicHandlers(executor, nil)

	if handlers == nil {
		t.Error("NewAnthropicHandlers returned nil")
	}
	if handlers.runner == nil {
		t.Error("runner not set")
	}
	if handlers.logger == nil {
		t.Error("logger not set (should default to slog.Default)")
	}
}

func TestOpenAPIStreamingResponses(t *testing.T) {
	api := humachi.New(chi.NewRouter(), huma.DefaultConfig("test", "1.0"))

	openaiHandlers := NewOpenAIHandlers(service.NewStatelessRunner(nil), nil)
	openaiHandlers.Register(api)

	anthropicHandlers := NewAnthropicHandlers(service.NewStatelessRunner(nil), nil)
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

	// Status can be "ok" or "degraded" depending on backend availability
	if resp.Body.Status != "ok" && resp.Body.Status != "degraded" {
		t.Errorf("expected status 'ok' or 'degraded', got %q", resp.Body.Status)
	}

	// Verify backends field is populated
	if resp.Body.Backends == nil {
		t.Error("expected non-nil backends field in response")
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

func TestCustomHandlersRegister(t *testing.T) {
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("test", "1.0"))

	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)
	handlers.Register(api)

	// Verify endpoints are registered
	paths := api.OpenAPI().Paths
	expectedPaths := []string{
		"/api/v1/prompt",
		"/api/v1/parallel",
		"/api/v1/chain",
		"/api/v1/compare",
		"/api/v1/backends",
		"/api/v1/sessions",
		"/api/v1/sessions/{id}",
		"/health",
	}

	for _, path := range expectedPaths {
		if _, ok := paths[path]; !ok {
			t.Errorf("expected path %q to be registered", path)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "0s"},
		{"seconds only", 45 * time.Second, "45s"},
		{"minutes and seconds", 3*time.Minute + 25*time.Second, "3m25s"},
		{"hours minutes seconds", 2*time.Hour + 15*time.Minute + 30*time.Second, "2h15m30s"},
		{"one hour", time.Hour, "1h0m0s"},
		{"subsecond rounds to zero", 500 * time.Millisecond, "1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestHandleGetSession_NotFound(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)

	_, err := handlers.HandleGetSession(context.Background(), &GetSessionInput{
		ID: "nonexistent-session-id",
	})

	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestHandleDeleteSession_NotFound(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewCustomHandlers(executor)

	_, err := handlers.HandleDeleteSession(context.Background(), &DeleteSessionInput{
		ID: "nonexistent-session-id",
	})

	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestNewCustomHandlersWithHealthInfo(t *testing.T) {
	executor := service.NewExecutor()
	healthInfo := HealthInfo{
		Version:   "2.0.0",
		StartTime: time.Now().Add(-time.Hour),
	}

	handlers := NewCustomHandlersWithHealthInfo(executor, healthInfo)

	if handlers == nil {
		t.Fatal("NewCustomHandlersWithHealthInfo returned nil")
	}
	if handlers.healthInfo.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", handlers.healthInfo.Version, "2.0.0")
	}
}

func TestHealthResponseIncludesAllFields(t *testing.T) {
	executor := service.NewExecutor()
	healthInfo := HealthInfo{
		Version:   "1.2.3",
		StartTime: time.Now().Add(-30 * time.Minute),
	}
	handlers := NewCustomHandlersWithHealthInfo(executor, healthInfo)

	resp, err := handlers.HandleHealth(context.Background(), &HealthInput{})
	if err != nil {
		t.Fatalf("HandleHealth failed: %v", err)
	}

	// Check version is included
	if resp.Body.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", resp.Body.Version, "1.2.3")
	}

	// Check uptime is calculated
	if resp.Body.UptimeMillis <= 0 {
		t.Error("UptimeMillis should be positive")
	}

	// Check uptime string is formatted
	if resp.Body.Uptime == "" {
		t.Error("Uptime string should not be empty")
	}

	// Check session store status is included
	if !resp.Body.SessionStore.Available {
		t.Log("Session store reported as unavailable (may be expected in test env)")
	}
}
