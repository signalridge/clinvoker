package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/auth"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/server/handlers"
	"github.com/signalridge/clinvoker/internal/server/mcp"
	"github.com/signalridge/clinvoker/internal/server/service"
)

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "clinvk-server-tests-")
	if err != nil {
		os.Exit(1)
	}
	_ = os.Setenv("HOME", tempDir)
	_ = os.Unsetenv(auth.EnvAPIKeys)
	_ = os.Unsetenv(auth.EnvAPIKeysGopassPath)
	auth.ResetCache()
	config.Reset()

	code := m.Run()

	_ = os.RemoveAll(tempDir)
	os.Exit(code)
}

func TestServerCreation(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 8080,
	}
	logger := slog.Default()

	srv := New(cfg, logger)
	if srv == nil {
		t.Fatal("expected server to be created")
	}

	if srv.API() == nil {
		t.Error("expected API to be set")
	}

	if srv.Router() == nil {
		t.Error("expected router to be set")
	}

	if srv.Executor() == nil {
		t.Error("expected executor to be set")
	}
}

func TestNewRouter_EnforcesAPIKeyOnMCP(t *testing.T) {
	setTempHome(t)
	t.Setenv(auth.EnvAPIKeys, "test-key")
	auth.ResetCache()
	t.Cleanup(auth.ResetCache)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, limiter := NewRouter(logger)
	if limiter != nil {
		defer limiter.Stop()
	}

	registry := mcp.NewRegistry()
	mcp.RegisterAllTools(registry, service.NewExecutor())
	dispatcher := mcp.NewDispatcher(registry, logger)
	transport := mcp.NewHTTPTransport(dispatcher, logger)
	router.Handle("/mcp", transport.Handler())

	body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestNewRouter_MCP_SSE_RequiresAPIKey(t *testing.T) {
	setTempHome(t)
	t.Setenv(auth.EnvAPIKeys, "test-key")
	auth.ResetCache()
	t.Cleanup(auth.ResetCache)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, limiter := NewRouter(logger)
	if limiter != nil {
		defer limiter.Stop()
	}

	registry := mcp.NewRegistry()
	mcp.RegisterAllTools(registry, service.NewExecutor())
	dispatcher := mcp.NewDispatcher(registry, logger)
	transport := mcp.NewHTTPTransport(dispatcher, logger)
	router.Handle("/mcp", transport.Handler())

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestNewRouter_MCP_SSE_WithAPIKey_Succeeds(t *testing.T) {
	setTempHome(t)
	t.Setenv(auth.EnvAPIKeys, "test-key")
	auth.ResetCache()
	t.Cleanup(auth.ResetCache)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, limiter := NewRouter(logger)
	if limiter != nil {
		defer limiter.Stop()
	}

	registry := mcp.NewRegistry()
	mcp.RegisterAllTools(registry, service.NewExecutor())
	dispatcher := mcp.NewDispatcher(registry, logger)
	transport := mcp.NewHTTPTransport(dispatcher, logger)
	router.Handle("/mcp", transport.Handler())

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("content-type = %q, want %q", got, "text/event-stream")
	}
}

func TestNewRouter_MCP_NonStream_AcceptSSE_StillJSON(t *testing.T) {
	setTempHome(t)
	t.Setenv(auth.EnvAPIKeys, "test-key")
	auth.ResetCache()
	t.Cleanup(auth.ResetCache)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, limiter := NewRouter(logger)
	if limiter != nil {
		defer limiter.Stop()
	}

	registry := mcp.NewRegistry()
	mcp.RegisterAllTools(registry, service.NewExecutor())
	dispatcher := mcp.NewDispatcher(registry, logger)
	transport := mcp.NewHTTPTransport(dispatcher, logger)
	router.Handle("/mcp", transport.Handler())

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("X-Api-Key", "test-key")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type = %q, want %q", got, "application/json")
	}
}

func setTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	config.Reset()
}

func TestHealthEndpoint(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 8080,
	}
	logger := slog.Default()

	srv := New(cfg, logger)
	customHandlers := handlers.NewCustomHandlers(srv.Executor())
	customHandlers.Register(srv.API())

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Status can be "ok" or "degraded" depending on backend availability
	status, ok := resp["status"].(string)
	if !ok || (status != "ok" && status != "degraded") {
		t.Errorf("expected status 'ok' or 'degraded', got %v", resp["status"])
	}

	// Verify backends field exists
	if _, ok := resp["backends"]; !ok {
		t.Error("expected 'backends' field in response")
	}
}

func TestBackendsEndpoint(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 8080,
	}
	logger := slog.Default()

	srv := New(cfg, logger)
	customHandlers := handlers.NewCustomHandlers(srv.Executor())
	customHandlers.Register(srv.API())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/backends", http.NoBody)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	backends, ok := resp["backends"].([]interface{})
	if !ok {
		t.Fatal("expected backends array in response")
	}

	// Should have at least one backend registered
	if len(backends) == 0 {
		t.Error("expected at least one backend")
	}
}

func TestOpenAIModelsEndpoint(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 8080,
	}
	logger := slog.Default()

	srv := New(cfg, logger)
	openaiHandlers := handlers.NewOpenAIHandlers(srv.Executor(), logger)
	openaiHandlers.Register(srv.API())

	req := httptest.NewRequest(http.MethodGet, "/openai/v1/models", http.NoBody)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["object"] != "list" {
		t.Errorf("expected object 'list', got %v", resp["object"])
	}

	if _, ok := resp["data"].([]interface{}); !ok {
		t.Fatal("expected data array in response")
	}

	// List may be empty if no enabled backends are available in this environment.
}

func TestPromptEndpointValidation(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 8080,
	}
	logger := slog.Default()

	srv := New(cfg, logger)
	customHandlers := handlers.NewCustomHandlers(srv.Executor())
	customHandlers.Register(srv.API())

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name:       "missing backend",
			body:       map[string]interface{}{"prompt": "test"},
			wantStatus: http.StatusUnprocessableEntity, // huma returns 422 for schema validation
		},
		{
			name:       "missing prompt",
			body:       map[string]interface{}{"backend": "claude"},
			wantStatus: http.StatusUnprocessableEntity, // huma returns 422 for schema validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/prompt", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.Router().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestChatCompletionsValidation(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 8080,
	}
	logger := slog.Default()

	srv := New(cfg, logger)
	openaiHandlers := handlers.NewOpenAIHandlers(srv.Executor(), logger)
	openaiHandlers.Register(srv.API())

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name:       "missing model",
			body:       map[string]interface{}{"messages": []map[string]string{{"role": "user", "content": "test"}}},
			wantStatus: http.StatusUnprocessableEntity, // huma returns 422 for schema validation
		},
		{
			name:       "missing messages",
			body:       map[string]interface{}{"model": "claude"},
			wantStatus: http.StatusUnprocessableEntity, // huma returns 422 for schema validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/openai/v1/chat/completions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.Router().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestAnthropicMessagesValidation(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 8080,
	}
	logger := slog.Default()

	srv := New(cfg, logger)
	anthropicHandlers := handlers.NewAnthropicHandlers(srv.Executor(), logger)
	anthropicHandlers.Register(srv.API())

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name:       "missing model",
			body:       map[string]interface{}{"max_tokens": 100, "messages": []map[string]string{{"role": "user", "content": "test"}}},
			wantStatus: http.StatusUnprocessableEntity, // huma returns 422 for schema validation
		},
		{
			name:       "missing max_tokens",
			body:       map[string]interface{}{"model": "claude", "messages": []map[string]string{{"role": "user", "content": "test"}}},
			wantStatus: http.StatusUnprocessableEntity, // huma returns 422 for schema validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/anthropic/v1/messages", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.Router().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestGracefulShutdown(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 0, // Use random port
	}
	logger := slog.Default()

	srv := New(cfg, logger)

	// Shutdown should work even if server wasn't started
	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("expected no error on shutdown before start, got %v", err)
	}
}
