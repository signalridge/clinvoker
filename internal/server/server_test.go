package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/signalridge/clinvoker/internal/server/handlers"
)

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

func TestHealthEndpoint(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 8080,
	}
	logger := slog.Default()

	srv := New(cfg, logger)
	customHandlers := handlers.NewCustomHandlers(srv.Executor())
	customHandlers.Register(srv.API())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/backends", nil)
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
	openaiHandlers := handlers.NewOpenAIHandlers(srv.Executor())
	openaiHandlers.Register(srv.API())

	req := httptest.NewRequest(http.MethodGet, "/openai/v1/models", nil)
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

	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("expected data array in response")
	}

	if len(data) == 0 {
		t.Error("expected at least one model")
	}
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
	openaiHandlers := handlers.NewOpenAIHandlers(srv.Executor())
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
		{
			name:       "streaming not supported",
			body:       map[string]interface{}{"model": "claude", "messages": []map[string]string{{"role": "user", "content": "test"}}, "stream": true},
			wantStatus: http.StatusBadRequest, // explicit 400 from handler
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
	anthropicHandlers := handlers.NewAnthropicHandlers(srv.Executor())
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
		{
			name:       "streaming not supported",
			body:       map[string]interface{}{"model": "claude", "max_tokens": 100, "messages": []map[string]string{{"role": "user", "content": "test"}}, "stream": true},
			wantStatus: http.StatusBadRequest, // explicit 400 from handler
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
