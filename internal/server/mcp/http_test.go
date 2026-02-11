package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/server/handlers"
	"github.com/signalridge/clinvoker/internal/server/service"
)

func TestNewHTTPTransport(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)

	transport := NewHTTPTransport(dispatcher, logger)
	if transport == nil {
		t.Fatal("NewHTTPTransport returned nil")
	}
	if transport.dispatcher != dispatcher {
		t.Error("dispatcher not set correctly")
	}
	if transport.logger == nil {
		t.Error("logger not set correctly")
	}
}

func TestHTTPTransport_Handler_MethodNotAllowed(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	// Test GET request (should be rejected)
	req := httptest.NewRequest(http.MethodGet, "/mcp", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHTTPTransport_Handler_InvalidJSON(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error in response")
	}
	if resp.Error.Code != CodeParseError {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeParseError)
	}
}

func TestHTTPTransport_Handler_InvalidJSONRPCVersion(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"1.0","id":1,"method":"ping"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error in response")
	}
	if resp.Error.Code != CodeInvalidRequest {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidRequest)
	}
}

func TestHTTPTransport_Handler_Initialize(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("content-type = %q, want %q", contentType, "application/json")
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Error("expected result")
	}
}

func TestHTTPTransport_Handler_Initialize_UnsupportedProtocolVersion(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"9999-01-01"}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
	if !strings.Contains(resp.Error.Message, "unsupported protocol version") {
		t.Errorf("error message = %q, want contains %q", resp.Error.Message, "unsupported protocol version")
	}
}

func TestHTTPTransport_Handler_ToolsList(t *testing.T) {
	registry := NewRegistry()
	RegisterAllTools(registry, nil)
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	var result struct {
		Tools []Tool `json:"tools"`
	}
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(result.Tools) == 0 {
		t.Error("expected tools in response")
	}
	for _, tool := range result.Tools {
		if len(tool.InputSchema) == 0 {
			t.Errorf("tool %q missing input schema", tool.Name)
		}
		if len(tool.OutputSchema) == 0 {
			t.Errorf("tool %q missing output schema", tool.Name)
		}
	}
}

func TestHTTPTransport_Handler_Ping(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":3,"method":"ping"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	// Ping returns empty object (just verifies connectivity)
	var result map[string]interface{}
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestHTTPTransport_Handler_ToolsCall_NonStreaming(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	RegisterAllTools(registry, service.NewExecutor())
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"clinvk_backends","arguments":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result ToolCallResult
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("failed to unmarshal tool result: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var bodyResp handlers.BackendsResponseBody
	if err := json.Unmarshal([]byte(result.Content[0].Text), &bodyResp); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if bodyResp.Backends == nil {
		t.Fatal("expected backends to be non-nil")
	}
}

func TestHTTPTransport_Handler_ToolsCall_UnknownTool(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != CodeToolNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeToolNotFound)
	}
}

func TestHTTPTransport_Handler_ToolsCallNotification_IgnoredNoExecution(t *testing.T) {
	registry := NewRegistry()
	called := false
	registry.Register(&ToolDefinition{
		Tool: Tool{
			Name:         "test_tool",
			Description:  "test tool",
			InputSchema:  json.RawMessage(`{}`),
			OutputSchema: json.RawMessage(`{}`),
		},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			called = true
			return &ToolCallResult{Content: []ContentBlock{TextContent("ok")}}, nil
		},
	})
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"test_tool","arguments":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Error("notification response should have empty body")
	}
	if called {
		t.Fatal("expected tools/call notification to be ignored without executing handler")
	}
}

func TestHTTPTransport_Handler_Notification(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Notifications should return 204 No Content
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Error("notification response should have empty body")
	}
}

func TestHTTPTransport_Handler_AcceptSSEWithoutStreaming(t *testing.T) {
	registry := NewRegistry()
	RegisterAllTools(registry, nil)
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":4,"method":"tools/list"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("content-type = %q, want %q", contentType, "application/json")
	}
}

func TestHTTPTransport_Handler_StreamRequestedWithoutSSE(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	RegisterAllTools(registry, service.NewExecutor())
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error response")
	}
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
}

func TestHTTPTransport_Handler_StreamRequestedWithSSE(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	RegisterAllTools(registry, service.NewExecutor())
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream, application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("content-type = %q, want %q", contentType, "text/event-stream")
	}
	if rec.Header().Get("Cache-Control") != "no-cache" {
		t.Error("expected Cache-Control: no-cache")
	}
}

func TestHTTPTransport_Handler_StreamSSEEmitsNotifications(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	RegisterAllTools(registry, service.NewExecutor())
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	bodyStr := rec.Body.String()
	if !strings.Contains(bodyStr, "event: notification") {
		t.Fatal("expected SSE notifications in response")
	}
	if !strings.Contains(bodyStr, "event: message") {
		t.Fatal("expected SSE completion message in response")
	}
}

func TestSSENotificationSender(t *testing.T) {
	logger := slog.Default()
	rec := httptest.NewRecorder()

	// Create a mock flusher
	flusher := &mockFlusher{ResponseWriter: rec}

	sender := &sseNotificationSender{
		writer:  rec,
		flusher: flusher,
		logger:  logger,
	}

	notification := &Notification{
		JSONRPC: "2.0",
		Method:  "notifications/message",
		Params:  json.RawMessage(`{"level":"info"}`),
	}

	err := sender.SendNotification(notification)
	if err != nil {
		t.Fatalf("SendNotification error: %v", err)
	}

	// Verify SSE format
	output := rec.Body.String()
	if !strings.HasPrefix(output, "event: notification") {
		t.Errorf("expected SSE event prefix, got: %s", output)
	}
	if !strings.Contains(output, "data:") {
		t.Error("expected data: field in SSE output")
	}
}

func TestWriteJSONError(t *testing.T) {
	rec := httptest.NewRecorder()

	writeJSONError(rec, CodeInvalidParams, "Invalid parameters")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("content-type = %q, want %q", contentType, "application/json")
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
	if resp.Error.Message != "Invalid parameters" {
		t.Errorf("error message = %q, want %q", resp.Error.Message, "Invalid parameters")
	}
}

func TestHTTPTransport_Handler_UnknownMethod(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":5,"method":"unknown/method"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should still return 200 OK with JSON-RPC error
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeMethodNotFound)
	}
}

// mockFlusher is a mock http.Flusher for testing
type mockFlusher struct {
	http.ResponseWriter
	flushed bool
}

func (m *mockFlusher) Flush() {
	m.flushed = true
}

func TestHTTPTransport_Handler_ContextCancellation(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	body := `{"jsonrpc":"2.0","id":6,"method":"ping"}`
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Response should still be written
	if rec.Code != http.StatusOK {
		t.Logf("status = %d (may vary with context cancellation)", rec.Code)
	}
}

func TestHTTPTransport_Handler_LargeRequest(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	// Create a large request body
	largeBody := bytes.Repeat([]byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}`), 1000)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Large body should be parsed without error
	if rec.Code != http.StatusBadRequest {
		// Multiple JSON objects are invalid, so we expect an error
		t.Logf("status = %d", rec.Code)
	}
}

func TestHTTPTransport_Handler_EmptyBody(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHTTPTransport_Handler_ReadBodyError(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)
	handler := transport.Handler()

	// Create a reader that returns an error
	badReader := &errorReader{err: errors.New("read error")}
	req := httptest.NewRequest(http.MethodPost, "/mcp", badReader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// errorReader is an io.Reader that always returns an error
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func TestHTTPTransport_ImplementsHandler(t *testing.T) {
	// Verify HTTPTransport.Handler returns an http.Handler
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)
	transport := NewHTTPTransport(dispatcher, logger)

	var _ = transport.Handler()
}
