package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/signalridge/clinvoker/internal/server/service"
)

func TestNewDispatcher(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	dispatcher := NewDispatcher(registry, logger)
	if dispatcher == nil {
		t.Fatal("NewDispatcher returned nil")
	}
	if dispatcher.registry != registry {
		t.Error("registry not set correctly")
	}
	if dispatcher.logger == nil {
		t.Error("logger not set correctly")
	}
}

func TestDispatcher_Dispatch_Initialize(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05"}`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("expected result")
	}

	// Verify result structure
	var result struct {
		ProtocolVersion string `json:"protocolVersion"`
		Capabilities    struct {
			Tools struct{} `json:"tools"`
		} `json:"capabilities"`
		ServerInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if result.ProtocolVersion != "2024-11-05" {
		t.Errorf("protocolVersion = %q, want %q", result.ProtocolVersion, "2024-11-05")
	}
	if result.ServerInfo.Name != "clinvoker" {
		t.Errorf("serverInfo.name = %q, want %q", result.ServerInfo.Name, "clinvoker")
	}
}

func TestDispatcher_Dispatch_ToolsList(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/list",
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("expected result")
	}

	// Verify tools are returned
	// The registry is empty by default, so tools list should be empty
	var result struct {
		Tools []Tool `json:"tools"`
	}
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Registry starts empty, so expect 0 tools unless RegisterAllTools was called
	// Just verify the result structure is valid
	if result.Tools == nil {
		t.Error("expected tools array to be non-nil")
	}
}

func TestDispatcher_Dispatch_UnknownMethod(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`3`),
		Method:  "unknown/method",
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeMethodNotFound)
	}
}

func TestDispatcher_Dispatch_Notification(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	// Notifications have no ID
	req := &Request{
		JSONRPC: "2.0",
		ID:      nil,
		Method:  "notifications/initialized",
		Params:  json.RawMessage(`{}`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	// Notifications should not return a response
	if resp != nil {
		t.Error("expected nil response for notification")
	}
}

func TestDispatcher_Dispatch_ToolsCall_UnknownTool(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`4`),
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "unknown_tool",
			"arguments": {}
		}`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	// The dispatcher returns CodeInvalidParams for unknown tools
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
}

func TestDispatcher_Dispatch_ToolsCall_InvalidParams(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`5`),
		Method:  "tools/call",
		Params:  json.RawMessage(`invalid json`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
}

func TestDispatcher_Dispatch_ToolsCall_InvalidArguments(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	RegisterAllTools(registry, service.NewExecutor())
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`6`),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"clinvk_prompt","arguments":"not-an-object"}`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
}

func TestDispatcher_Dispatch_ToolsCall_PromptExecutionError(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	RegisterAllTools(registry, service.NewExecutor())
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`7`),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello"}}`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
}

func TestDispatcher_Dispatch_ToolsCall_MissingName(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`6`),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"arguments":{}}`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeInvalidParams)
	}
}

func TestDispatcher_Dispatch_ToolsCall_Backends(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	RegisterAllTools(registry, nil) // No executor needed for backends tool

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`7`),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"clinvk_backends","arguments":{}}`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("expected result")
	}

	var result ToolCallResult
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if result.Content == nil {
		t.Error("expected content in result")
	}
}

func TestDispatcher_Dispatch_ToolsCall_NotificationExecutes(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

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

	req := &Request{
		JSONRPC: "2.0",
		ID:      nil,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"test_tool","arguments":{}}`),
	}

	resp := dispatcher.Dispatch(context.Background(), req)
	if resp != nil {
		t.Fatal("expected nil response for notification")
	}
	if !called {
		t.Fatal("expected notification to execute handler")
	}
}

func TestDispatcher_Dispatch_Ping(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dispatcher := NewDispatcher(registry, logger)

	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`8`),
		Method:  "ping",
	}

	resp := dispatcher.Dispatch(context.Background(), req)

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("expected result")
	}

	// Ping returns an empty object (just verifies connectivity)
	var result map[string]interface{}
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Just verify result is valid JSON (empty object is valid)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestToolCallParams(t *testing.T) {
	// Test the ToolCallParams structure used in handleToolsCall
	tests := []struct {
		name   string
		params ToolCallParams
	}{
		{
			name: "valid params",
			params: ToolCallParams{
				Name:      "test_tool",
				Arguments: json.RawMessage(`{"key":"value"}`),
			},
		},
		{
			name: "no arguments",
			params: ToolCallParams{
				Name:      "test_tool",
				Arguments: nil,
			},
		},
		{
			name: "empty name",
			params: ToolCallParams{
				Name:      "",
				Arguments: json.RawMessage(`{}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the params can be marshaled and unmarshaled
			data, err := json.Marshal(tt.params)
			if err != nil {
				t.Errorf("failed to marshal params: %v", err)
				return
			}

			var parsed ToolCallParams
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Errorf("failed to unmarshal params: %v", err)
				return
			}

			if parsed.Name != tt.params.Name {
				t.Errorf("name = %q, want %q", parsed.Name, tt.params.Name)
			}
		})
	}
}
