package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if registry.tools == nil {
		t.Error("tools map not initialized")
	}
	// order is lazily initialized on first Register
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	tool := Tool{
		Name:         "test_tool",
		Description:  "A test tool",
		InputSchema:  json.RawMessage(`{"type":"object"}`),
		OutputSchema: json.RawMessage(`{"type":"object"}`),
	}

	handler := func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
		return &ToolCallResult{
			Content: []ContentBlock{TextContent("test result")},
		}, nil
	}

	registry.Register(&ToolDefinition{
		Tool:    tool,
		Handler: handler,
	})

	// Verify tool was registered
	def, ok := registry.Get("test_tool")
	if !ok {
		t.Fatal("tool not found after registration")
	}
	if def.Tool.Name != "test_tool" {
		t.Errorf("tool name = %q, want %q", def.Tool.Name, "test_tool")
	}
	if def.Tool.Description != "A test tool" {
		t.Errorf("tool description = %q, want %q", def.Tool.Description, "A test tool")
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	registry := NewRegistry()

	def, ok := registry.Get("nonexistent")
	if ok {
		t.Error("expected false for non-existent tool")
	}
	if def != nil {
		t.Error("expected nil definition for non-existent tool")
	}
}

func TestRegistry_ListTools(t *testing.T) {
	registry := NewRegistry()

	// Register multiple tools
	tools := []Tool{
		{Name: "tool1", Description: "Tool 1", InputSchema: json.RawMessage(`{}`)},
		{Name: "tool2", Description: "Tool 2", InputSchema: json.RawMessage(`{}`)},
		{Name: "tool3", Description: "Tool 3", InputSchema: json.RawMessage(`{}`)},
	}

	for _, tool := range tools {
		registry.Register(&ToolDefinition{
			Tool:    tool,
			Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) { return nil, nil },
		})
	}

	list := registry.ListTools()
	if len(list) != len(tools) {
		t.Errorf("expected %d tools, got %d", len(tools), len(list))
	}

	// Verify all tools are in the list
	toolNames := make(map[string]bool)
	for _, tool := range list {
		toolNames[tool.Name] = true
	}
	for _, tool := range tools {
		if !toolNames[tool.Name] {
			t.Errorf("tool %q not found in list", tool.Name)
		}
	}
}

func TestRegistry_ListTools_Empty(t *testing.T) {
	registry := NewRegistry()

	list := registry.ListTools()
	if list == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(list) != 0 {
		t.Errorf("expected 0 tools, got %d", len(list))
	}
}

func TestRegistry_Register_Overwrite(t *testing.T) {
	registry := NewRegistry()

	// Register first version
	registry.Register(&ToolDefinition{
		Tool:    Tool{Name: "tool1", Description: "Version 1", InputSchema: json.RawMessage(`{}`)},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) { return nil, nil },
	})

	// Register second version (should overwrite but preserve order)
	registry.Register(&ToolDefinition{
		Tool:    Tool{Name: "tool1", Description: "Version 2", InputSchema: json.RawMessage(`{}`)},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) { return nil, nil },
	})

	def, ok := registry.Get("tool1")
	if !ok {
		t.Fatal("tool not found")
	}
	if def.Tool.Description != "Version 2" {
		t.Errorf("description = %q, want %q", def.Tool.Description, "Version 2")
	}

	// Should still only have one tool in the list
	list := registry.ListTools()
	if len(list) != 1 {
		t.Errorf("expected 1 tool, got %d", len(list))
	}
}

func TestToolDefinition_Handler(t *testing.T) {
	handlerCalled := false
	def := &ToolDefinition{
		Tool: Tool{Name: "test", InputSchema: json.RawMessage(`{}`)},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			handlerCalled = true
			return &ToolCallResult{
				Content: []ContentBlock{TextContent("result")},
			}, nil
		},
	}

	result, err := def.Handler(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
	if result == nil {
		t.Error("expected result, got nil")
	}
}

func TestContentBlock(t *testing.T) {
	tests := []struct {
		name     string
		block    ContentBlock
		wantType string
		wantText string
	}{
		{
			name:     "text content",
			block:    TextContent("Hello, World!"),
			wantType: "text",
			wantText: "Hello, World!",
		},
		{
			name:     "empty content",
			block:    ContentBlock{Type: "text"},
			wantType: "text",
			wantText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.block.Type != tt.wantType {
				t.Errorf("type = %q, want %q", tt.block.Type, tt.wantType)
			}
			if tt.block.Text != tt.wantText {
				t.Errorf("text = %q, want %q", tt.block.Text, tt.wantText)
			}
		})
	}
}

func TestTool_MarshalJSON(t *testing.T) {
	tool := Tool{
		Name:         "test_tool",
		Description:  "A test tool for testing",
		InputSchema:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"ok":{"type":"boolean"}}}`),
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if unmarshaled["name"] != "test_tool" {
		t.Errorf("name = %v, want %q", unmarshaled["name"], "test_tool")
	}
	if unmarshaled["description"] != "A test tool for testing" {
		t.Errorf("description = %v, want %q", unmarshaled["description"], "A test tool for testing")
	}
	if _, ok := unmarshaled["inputSchema"]; !ok {
		t.Error("inputSchema not present in marshaled JSON")
	}
	if _, ok := unmarshaled["outputSchema"]; !ok {
		t.Error("outputSchema not present in marshaled JSON")
	}
}

func TestRegisterAllTools(t *testing.T) {
	registry := NewRegistry()
	RegisterAllTools(registry, nil)

	// Verify all expected tools are registered
	expectedTools := []string{
		"clinvk_prompt",
		"clinvk_parallel",
		"clinvk_chain",
		"clinvk_compare",
		"clinvk_backends",
		"clinvk_sessions",
		"clinvk_session_get",
		"clinvk_session_delete",
		"clinvk_health",
	}

	for _, toolName := range expectedTools {
		def, ok := registry.Get(toolName)
		if !ok {
			t.Errorf("tool %q not found", toolName)
			continue
		}
		if def.Tool.Name != toolName {
			t.Errorf("tool name = %q, want %q", def.Tool.Name, toolName)
		}
		if def.Tool.Description == "" {
			t.Errorf("tool %q has empty description", toolName)
		}
		if def.Tool.InputSchema == nil {
			t.Errorf("tool %q has nil inputSchema", toolName)
		}
		if def.Tool.OutputSchema == nil {
			t.Errorf("tool %q has nil outputSchema", toolName)
		}
	}

	// Verify we have exactly 9 tools
	allTools := registry.ListTools()
	if len(allTools) != 9 {
		t.Errorf("expected 9 tools, got %d", len(allTools))
	}
}

func TestErrorResult(t *testing.T) {
	result := ErrorResult("something went wrong")

	if result == nil {
		t.Fatal("ErrorResult returned nil")
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	if len(result.Content) != 1 {
		t.Errorf("expected 1 content block, got %d", len(result.Content))
	}
	if result.Content[0].Text != "something went wrong" {
		t.Errorf("text = %q, want %q", result.Content[0].Text, "something went wrong")
	}
}

func TestMarshalToToolCallResult(t *testing.T) {
	data := map[string]string{"key": "value"}
	result, err := marshalToToolCallResult(data)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.Content) != 1 {
		t.Errorf("expected 1 content block, got %d", len(result.Content))
	}
	if result.Content[0].Type != "text" {
		t.Errorf("type = %q, want %q", result.Content[0].Type, "text")
	}
}

func TestMustJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	result := mustJSON(data)

	if result == nil {
		t.Error("mustJSON returned nil")
	}

	// Verify it's valid JSON
	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Errorf("mustJSON did not return valid JSON: %v", err)
	}
	if parsed["key"] != "value" {
		t.Errorf("parsed value = %q, want %q", parsed["key"], "value")
	}
}

func TestMustJSON_Panic(t *testing.T) {
	// Test with a value that can't be marshaled (channel)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unmarshalable value")
		}
	}()

	mustJSON(make(chan int))
}
