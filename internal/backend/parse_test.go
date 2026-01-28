package backend

import (
	"strings"
	"testing"
)

// ==================== Claude ParseJSONResponse Tests ====================

func TestClaude_ParseJSONResponse_Success(t *testing.T) {
	b := &Claude{}

	input := `{
		"type": "result",
		"result": "Hello! How can I help you?",
		"session_id": "abc-123-def",
		"duration_ms": 1500,
		"usage": {
			"input_tokens": 10,
			"output_tokens": 20
		}
	}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello! How can I help you?" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello! How can I help you?")
	}
	if resp.SessionID != "abc-123-def" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "abc-123-def")
	}
	if resp.DurationMs != 1500 {
		t.Errorf("DurationMs = %d, want %d", resp.DurationMs, 1500)
	}
	if resp.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if resp.Usage.InputTokens != 10 {
		t.Errorf("InputTokens = %d, want %d", resp.Usage.InputTokens, 10)
	}
	if resp.Usage.OutputTokens != 20 {
		t.Errorf("OutputTokens = %d, want %d", resp.Usage.OutputTokens, 20)
	}
	if resp.Usage.TotalTokens != 30 {
		t.Errorf("TotalTokens = %d, want %d", resp.Usage.TotalTokens, 30)
	}
	if resp.Error != "" {
		t.Errorf("Error should be empty, got %q", resp.Error)
	}
}

func TestClaude_ParseJSONResponse_ErrorResponse(t *testing.T) {
	b := &Claude{}

	tests := []struct {
		name        string
		input       string
		wantErr     string
		wantContent string // should be empty for error responses
	}{
		{
			name:        "error field",
			input:       `{"error": "API rate limit exceeded"}`,
			wantErr:     "API rate limit exceeded",
			wantContent: "",
		},
		{
			name:        "error type with message",
			input:       `{"type": "error", "message": "Invalid API key"}`,
			wantErr:     "Invalid API key",
			wantContent: "",
		},
		{
			name:        "error with result field should prioritize error",
			input:       `{"error": "Request failed", "result": "some content"}`,
			wantErr:     "Request failed",
			wantContent: "", // error takes priority
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := b.ParseJSONResponse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Error != tt.wantErr {
				t.Errorf("Error = %q, want %q", resp.Error, tt.wantErr)
			}
			if resp.Content != tt.wantContent {
				t.Errorf("Content = %q, want %q (should be empty for error)", resp.Content, tt.wantContent)
			}
		})
	}
}

// TestClaude_ParseJSONResponse_SuccessNotError verifies success response doesn't trigger error path
func TestClaude_ParseJSONResponse_SuccessNotError(t *testing.T) {
	b := &Claude{}

	input := `{"type": "result", "result": "Hello!", "session_id": "sess-123"}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("Error should be empty for success response, got %q", resp.Error)
	}
	if resp.Content != "Hello!" {
		t.Errorf("Content = %q, want 'Hello!'", resp.Content)
	}
}

func TestClaude_ParseJSONResponse_InvalidJSON(t *testing.T) {
	b := &Claude{}

	_, err := b.ParseJSONResponse("not valid json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestClaude_ParseJSONResponse_EmptyInput(t *testing.T) {
	b := &Claude{}

	_, err := b.ParseJSONResponse("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

// ==================== Claude ParseOutput Tests ====================

func TestClaude_ParseOutput(t *testing.T) {
	b := &Claude{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text",
			input: "Hello world",
			want:  "Hello world",
		},
		{
			name:  "with newlines",
			input: "Line 1\nLine 2\nLine 3",
			want:  "Line 1\nLine 2\nLine 3",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := b.ParseOutput(tt.input)
			if got != tt.want {
				t.Errorf("ParseOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ==================== Codex ParseJSONResponse Tests ====================

func TestCodex_ParseJSONResponse_Success(t *testing.T) {
	b := &Codex{}

	input := `{"type":"thread.started","thread_id":"thread-123"}
{"type":"turn.started"}
{"type":"item.completed","item":{"id":"item_0","type":"reasoning","text":"Thinking..."}}
{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"Hello!"}}
{"type":"turn.completed","usage":{"input_tokens":100,"cached_input_tokens":50,"output_tokens":20}}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello!")
	}
	if resp.SessionID != "thread-123" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "thread-123")
	}
	if resp.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if resp.Usage.InputTokens != 150 { // 100 + 50 cached
		t.Errorf("InputTokens = %d, want %d", resp.Usage.InputTokens, 150)
	}
	if resp.Usage.OutputTokens != 20 {
		t.Errorf("OutputTokens = %d, want %d", resp.Usage.OutputTokens, 20)
	}
	if resp.Error != "" {
		t.Errorf("Error should be empty, got %q", resp.Error)
	}
}

func TestCodex_ParseJSONResponse_MultipleMessages(t *testing.T) {
	b := &Codex{}

	input := `{"type":"thread.started","thread_id":"thread-456"}
{"type":"item.completed","item":{"type":"agent_message","text":"First message"}}
{"type":"item.completed","item":{"type":"agent_message","text":"Second message"}}
{"type":"turn.completed","usage":{"input_tokens":10,"output_tokens":5}}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(resp.Content, "First message") {
		t.Errorf("Content should contain 'First message', got %q", resp.Content)
	}
	if !strings.Contains(resp.Content, "Second message") {
		t.Errorf("Content should contain 'Second message', got %q", resp.Content)
	}
}

func TestCodex_ParseJSONResponse_ErrorEvent(t *testing.T) {
	b := &Codex{}

	input := `{"type":"thread.started","thread_id":"thread-789"}
{"type":"error","message":"401 Unauthorized"}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error != "401 Unauthorized" {
		t.Errorf("Error = %q, want %q", resp.Error, "401 Unauthorized")
	}
	// Verify session ID is still captured even with error
	if resp.SessionID != "thread-789" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "thread-789")
	}
	// Content should be empty when there's an error (no agent_message events)
	if resp.Content != "" {
		t.Errorf("Content should be empty for error, got %q", resp.Content)
	}
}

func TestCodex_ParseJSONResponse_TurnFailed(t *testing.T) {
	b := &Codex{}

	input := `{"type":"thread.started","thread_id":"thread-abc"}
{"type":"turn.failed","error":{"message":"Model overloaded"}}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error != "Model overloaded" {
		t.Errorf("Error = %q, want %q", resp.Error, "Model overloaded")
	}
	if resp.SessionID != "thread-abc" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "thread-abc")
	}
}

// TestCodex_ParseJSONResponse_ErrorAfterContent tests error that occurs after some content
func TestCodex_ParseJSONResponse_ErrorAfterContent(t *testing.T) {
	b := &Codex{}

	// Simulates a scenario where partial content was received before error
	input := `{"type":"thread.started","thread_id":"thread-partial"}
{"type":"item.completed","item":{"type":"agent_message","text":"Starting..."}}
{"type":"error","message":"Connection lost"}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both content and error should be captured
	if resp.Error != "Connection lost" {
		t.Errorf("Error = %q, want %q", resp.Error, "Connection lost")
	}
	if resp.Content != "Starting..." {
		t.Errorf("Content = %q, want 'Starting...'", resp.Content)
	}
}

// TestCodex_ParseJSONResponse_MultipleErrors tests that only the last error is captured
func TestCodex_ParseJSONResponse_MultipleErrors(t *testing.T) {
	b := &Codex{}

	// Simulates multiple error events - only the last one should be captured
	input := `{"type":"thread.started","thread_id":"thread-multi-err"}
{"type":"error","message":"First error"}
{"type":"error","message":"Second error"}
{"type":"turn.failed","error":{"message":"Final error"}}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the last error (turn.failed) should be captured
	if resp.Error != "Final error" {
		t.Errorf("Error = %q, want 'Final error' (last error)", resp.Error)
	}
}

func TestCodex_ParseJSONResponse_EmptyInput(t *testing.T) {
	b := &Codex{}

	resp, err := b.ParseJSONResponse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty input should return empty response, not error
	if resp.Content != "" {
		t.Errorf("Content should be empty, got %q", resp.Content)
	}
}

// ==================== Codex ParseOutput Tests ====================

func TestCodex_ParseOutput(t *testing.T) {
	b := &Codex{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "extracts agent_message",
			input: `{"type":"item.completed","item":{"type":"agent_message","text":"Hello"}}
{"type":"item.completed","item":{"type":"reasoning","text":"Thinking"}}`,
			want: "Hello",
		},
		{
			name: "multiple agent_messages",
			input: `{"type":"item.completed","item":{"type":"agent_message","text":"First"}}
{"type":"item.completed","item":{"type":"agent_message","text":"Second"}}`,
			want: "First\nSecond",
		},
		{
			name:  "ignores non-JSONL input",
			input: "Just plain text",
			want:  "", // Codex ParseOutput only extracts from JSONL
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "ignores invalid JSON lines",
			input: "invalid json\n{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"Valid\"}}",
			want:  "Valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := b.ParseOutput(tt.input)
			if got != tt.want {
				t.Errorf("ParseOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ==================== Gemini ParseJSONResponse Tests ====================

func TestGemini_ParseJSONResponse_Success(t *testing.T) {
	b := &Gemini{}

	input := `{
		"response": "Hello from Gemini!",
		"session_id": "gemini-session-123",
		"stats": {
			"models": {
				"gemini-2.5-flash": {
					"tokens": {
						"input": 100,
						"candidates": 50,
						"total": 150
					}
				}
			}
		}
	}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello from Gemini!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from Gemini!")
	}
	if resp.SessionID != "gemini-session-123" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "gemini-session-123")
	}
	if resp.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if resp.Usage.InputTokens != 100 {
		t.Errorf("InputTokens = %d, want %d", resp.Usage.InputTokens, 100)
	}
	if resp.Usage.OutputTokens != 50 {
		t.Errorf("OutputTokens = %d, want %d", resp.Usage.OutputTokens, 50)
	}
	if resp.Error != "" {
		t.Errorf("Error should be empty, got %q", resp.Error)
	}
}

func TestGemini_ParseJSONResponse_WithCredentialsPrefix(t *testing.T) {
	b := &Gemini{}

	input := `Loaded cached credentials.
{
	"response": "Hello!",
	"session_id": "sess-456"
}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello!")
	}
	if resp.SessionID != "sess-456" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "sess-456")
	}
}

func TestGemini_ParseJSONResponse_ErrorResponse(t *testing.T) {
	b := &Gemini{}

	input := `{
		"session_id": "sess-err",
		"error": {
			"type": "INVALID_ARGUMENT",
			"message": "Invalid model specified",
			"code": 400
		}
	}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error != "Invalid model specified" {
		t.Errorf("Error = %q, want %q", resp.Error, "Invalid model specified")
	}
	if resp.SessionID != "sess-err" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "sess-err")
	}
	// Content should be empty for error response
	if resp.Content != "" {
		t.Errorf("Content should be empty for error, got %q", resp.Content)
	}
}

func TestGemini_ParseJSONResponse_PlainTextError(t *testing.T) {
	b := &Gemini{}

	input := `Loaded cached credentials.
Error resuming session: Invalid session identifier "abc123".`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should strip "Loaded cached credentials." and return the error
	if !strings.Contains(resp.Error, "Error resuming session") {
		t.Errorf("Error should contain 'Error resuming session', got %q", resp.Error)
	}
	if strings.Contains(resp.Error, "Loaded cached credentials") {
		t.Errorf("Error should not contain 'Loaded cached credentials', got %q", resp.Error)
	}
}

func TestGemini_ParseJSONResponse_PlainTextErrorWithoutCredentials(t *testing.T) {
	b := &Gemini{}

	input := `Error: API key invalid`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error != "Error: API key invalid" {
		t.Errorf("Error = %q, want %q", resp.Error, "Error: API key invalid")
	}
}

func TestGemini_ParseJSONResponse_EmptyInput(t *testing.T) {
	b := &Gemini{}

	_, err := b.ParseJSONResponse("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestGemini_ParseJSONResponse_MultipleModels(t *testing.T) {
	b := &Gemini{}

	input := `{
		"response": "Multi-model response",
		"session_id": "sess-multi",
		"stats": {
			"models": {
				"gemini-2.5-flash": {
					"tokens": {"input": 100, "candidates": 20, "total": 120}
				},
				"gemini-2.5-flash-lite": {
					"tokens": {"input": 50, "candidates": 10, "total": 60}
				}
			}
		}
	}`

	resp, err := b.ParseJSONResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should aggregate tokens from all models
	if resp.Usage.InputTokens != 150 { // 100 + 50
		t.Errorf("InputTokens = %d, want %d", resp.Usage.InputTokens, 150)
	}
	if resp.Usage.OutputTokens != 30 { // 20 + 10
		t.Errorf("OutputTokens = %d, want %d", resp.Usage.OutputTokens, 30)
	}
	if resp.Usage.TotalTokens != 180 { // 120 + 60
		t.Errorf("TotalTokens = %d, want %d", resp.Usage.TotalTokens, 180)
	}
}

// ==================== Gemini ParseOutput Tests ====================

func TestGemini_ParseOutput(t *testing.T) {
	b := &Gemini{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text",
			input: "Hello from Gemini",
			want:  "Hello from Gemini",
		},
		{
			name:  "with newlines",
			input: "Line 1\nLine 2",
			want:  "Line 1\nLine 2",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := b.ParseOutput(tt.input)
			if got != tt.want {
				t.Errorf("ParseOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ==================== SeparateStderr Tests ====================

func TestSeparateStderr(t *testing.T) {
	tests := []struct {
		backend Backend
		want    bool
	}{
		{&Claude{}, false},
		{&Codex{}, true},
		{&Gemini{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.backend.Name(), func(t *testing.T) {
			got := tt.backend.SeparateStderr()
			if got != tt.want {
				t.Errorf("SeparateStderr() = %v, want %v", got, tt.want)
			}
		})
	}
}
