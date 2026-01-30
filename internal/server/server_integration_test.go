package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/signalridge/clinvoker/internal/server/handlers"
)

func setupTestServer() *Server {
	srv := New(Config{Host: "127.0.0.1", Port: 0}, nil)
	custom := handlers.NewCustomHandlers(srv.Executor())
	custom.Register(srv.API())
	return srv
}

func TestPromptEndpoint_Validation(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/prompt", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", rr.Code)
	}
}

func TestParallelEndpoint_Validation(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/parallel", bytes.NewBufferString(`{"tasks":[]}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
