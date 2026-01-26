package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// ==================== Event Tests ====================

func TestNewUnifiedEvent(t *testing.T) {
	t.Run("creates event with correct fields", func(t *testing.T) {
		before := time.Now()
		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		after := time.Now()

		if event.Type != EventMessage {
			t.Errorf("expected type %q, got %q", EventMessage, event.Type)
		}
		if event.Backend != "claude" {
			t.Errorf("expected backend %q, got %q", "claude", event.Backend)
		}
		if event.SessionID != "session-123" {
			t.Errorf("expected session ID %q, got %q", "session-123", event.SessionID)
		}
		if event.Timestamp.Before(before) || event.Timestamp.After(after) {
			t.Errorf("timestamp not in expected range")
		}
	})

	t.Run("creates event with all event types", func(t *testing.T) {
		types := []EventType{
			EventInit, EventMessage, EventToolUse, EventToolResult,
			EventThinking, EventError, EventDone, EventProgress, EventTokenUsage,
		}
		for _, et := range types {
			event := NewUnifiedEvent(et, "test", "test-session")
			if event.Type != et {
				t.Errorf("expected type %q, got %q", et, event.Type)
			}
		}
	})
}

func TestUnifiedEvent_SetContent(t *testing.T) {
	t.Run("sets content successfully", func(t *testing.T) {
		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		content := &MessageContent{Text: "Hello", Role: "assistant"}

		err := event.SetContent(content)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if event.Content == nil {
			t.Error("content should not be nil")
		}

		// Verify content can be unmarshaled back
		var parsed MessageContent
		if err := json.Unmarshal(event.Content, &parsed); err != nil {
			t.Errorf("failed to unmarshal content: %v", err)
		}
		if parsed.Text != "Hello" {
			t.Errorf("expected text %q, got %q", "Hello", parsed.Text)
		}
	})

	t.Run("handles nil content", func(t *testing.T) {
		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		err := event.SetContent(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestUnifiedEvent_GetInitContent(t *testing.T) {
	t.Run("returns content for init event", func(t *testing.T) {
		event := NewUnifiedEvent(EventInit, "claude", "session-123")
		event.SetContent(&InitContent{
			Model:            "opus",
			BackendSessionID: "backend-123",
			WorkingDir:       "/tmp",
		})

		content, err := event.GetInitContent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if content.Model != "opus" {
			t.Errorf("expected model %q, got %q", "opus", content.Model)
		}
		if content.BackendSessionID != "backend-123" {
			t.Errorf("expected backend session ID %q, got %q", "backend-123", content.BackendSessionID)
		}
	})

	t.Run("returns error for wrong event type", func(t *testing.T) {
		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		_, err := event.GetInitContent()
		if err != ErrInvalidEventType {
			t.Errorf("expected ErrInvalidEventType, got %v", err)
		}
	})
}

func TestUnifiedEvent_GetMessageContent(t *testing.T) {
	t.Run("returns content for message event", func(t *testing.T) {
		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		event.SetContent(&MessageContent{
			Text:      "Hello world",
			Role:      "assistant",
			IsPartial: true,
		})

		content, err := event.GetMessageContent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if content.Text != "Hello world" {
			t.Errorf("expected text %q, got %q", "Hello world", content.Text)
		}
		if content.Role != "assistant" {
			t.Errorf("expected role %q, got %q", "assistant", content.Role)
		}
		if !content.IsPartial {
			t.Error("expected IsPartial to be true")
		}
	})

	t.Run("returns error for wrong event type", func(t *testing.T) {
		event := NewUnifiedEvent(EventInit, "claude", "session-123")
		_, err := event.GetMessageContent()
		if err != ErrInvalidEventType {
			t.Errorf("expected ErrInvalidEventType, got %v", err)
		}
	})
}

func TestUnifiedEvent_GetToolUseContent(t *testing.T) {
	t.Run("returns content for tool use event", func(t *testing.T) {
		event := NewUnifiedEvent(EventToolUse, "claude", "session-123")
		event.SetContent(&ToolUseContent{
			ToolID:   "tool-1",
			ToolName: "read_file",
			Input:    json.RawMessage(`{"path": "/tmp/test.txt"}`),
		})

		content, err := event.GetToolUseContent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if content.ToolID != "tool-1" {
			t.Errorf("expected tool ID %q, got %q", "tool-1", content.ToolID)
		}
		if content.ToolName != "read_file" {
			t.Errorf("expected tool name %q, got %q", "read_file", content.ToolName)
		}
	})

	t.Run("returns error for wrong event type", func(t *testing.T) {
		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		_, err := event.GetToolUseContent()
		if err != ErrInvalidEventType {
			t.Errorf("expected ErrInvalidEventType, got %v", err)
		}
	})
}

func TestUnifiedEvent_GetToolResultContent(t *testing.T) {
	t.Run("returns content for tool result event", func(t *testing.T) {
		event := NewUnifiedEvent(EventToolResult, "claude", "session-123")
		event.SetContent(&ToolResultContent{
			ToolID:   "tool-1",
			ToolName: "read_file",
			Output:   "file content",
			IsError:  false,
		})

		content, err := event.GetToolResultContent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if content.Output != "file content" {
			t.Errorf("expected output %q, got %q", "file content", content.Output)
		}
		if content.IsError {
			t.Error("expected IsError to be false")
		}
	})

	t.Run("handles error result", func(t *testing.T) {
		event := NewUnifiedEvent(EventToolResult, "claude", "session-123")
		event.SetContent(&ToolResultContent{
			ToolID:   "tool-1",
			ToolName: "read_file",
			IsError:  true,
			ErrorMsg: "file not found",
		})

		content, err := event.GetToolResultContent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !content.IsError {
			t.Error("expected IsError to be true")
		}
		if content.ErrorMsg != "file not found" {
			t.Errorf("expected error message %q, got %q", "file not found", content.ErrorMsg)
		}
	})
}

func TestUnifiedEvent_GetErrorContent(t *testing.T) {
	t.Run("returns content for error event", func(t *testing.T) {
		event := NewUnifiedEvent(EventError, "claude", "session-123")
		event.SetContent(&ErrorContent{
			Code:    "RATE_LIMIT",
			Message: "Too many requests",
			Details: "Retry after 60 seconds",
		})

		content, err := event.GetErrorContent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if content.Code != "RATE_LIMIT" {
			t.Errorf("expected code %q, got %q", "RATE_LIMIT", content.Code)
		}
		if content.Message != "Too many requests" {
			t.Errorf("expected message %q, got %q", "Too many requests", content.Message)
		}
	})
}

func TestUnifiedEvent_GetDoneContent(t *testing.T) {
	t.Run("returns content for done event", func(t *testing.T) {
		event := NewUnifiedEvent(EventDone, "claude", "session-123")
		event.SetContent(&DoneContent{
			TokenUsage: &TokenUsageContent{
				InputTokens:  100,
				OutputTokens: 200,
			},
			TurnCount: 5,
		})

		content, err := event.GetDoneContent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if content.TokenUsage.InputTokens != 100 {
			t.Errorf("expected input tokens %d, got %d", 100, content.TokenUsage.InputTokens)
		}
		if content.TokenUsage.OutputTokens != 200 {
			t.Errorf("expected output tokens %d, got %d", 200, content.TokenUsage.OutputTokens)
		}
		if content.TurnCount != 5 {
			t.Errorf("expected turn count %d, got %d", 5, content.TurnCount)
		}
	})
}

func TestUnifiedEvent_JSON(t *testing.T) {
	t.Run("returns valid JSON", func(t *testing.T) {
		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		event.SetContent(&MessageContent{Text: "Hello"})
		event.Sequence = 1

		jsonStr, err := event.JSON()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify it's valid JSON
		var parsed map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			t.Errorf("invalid JSON: %v", err)
		}

		if parsed["type"] != "message" {
			t.Errorf("expected type %q, got %v", "message", parsed["type"])
		}
		if parsed["backend"] != "claude" {
			t.Errorf("expected backend %q, got %v", "claude", parsed["backend"])
		}
	})
}

// ==================== Parser Tests ====================

func TestNewParser(t *testing.T) {
	t.Run("creates parser with correct fields", func(t *testing.T) {
		p := NewParser("claude", "session-123")
		if p.backend != "claude" {
			t.Errorf("expected backend %q, got %q", "claude", p.backend)
		}
		if p.sessionID != "session-123" {
			t.Errorf("expected session ID %q, got %q", "session-123", p.sessionID)
		}
		if p.sequence != 0 {
			t.Errorf("expected sequence 0, got %d", p.sequence)
		}
	})
}

func TestParser_ParseLine(t *testing.T) {
	t.Run("empty line returns nil", func(t *testing.T) {
		p := NewParser("claude", "session-123")
		event, err := p.ParseLine("")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event != nil {
			t.Error("expected nil event for empty line")
		}
	})

	t.Run("whitespace-only line returns nil", func(t *testing.T) {
		p := NewParser("claude", "session-123")
		event, err := p.ParseLine("   \t  ")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event != nil {
			t.Error("expected nil event for whitespace line")
		}
	})

	t.Run("unknown backend returns error", func(t *testing.T) {
		p := NewParser("unknown", "session-123")
		_, err := p.ParseLine("test")
		if err == nil {
			t.Error("expected error for unknown backend")
		}
		if !strings.Contains(err.Error(), "unknown backend") {
			t.Errorf("expected 'unknown backend' error, got: %v", err)
		}
	})

	t.Run("increments sequence", func(t *testing.T) {
		p := NewParser("claude", "session-123")
		p.ParseLine("test1")
		p.ParseLine("test2")
		event, _ := p.ParseLine("test3")

		if event.Sequence != 3 {
			t.Errorf("expected sequence 3, got %d", event.Sequence)
		}
	})
}

func TestParser_ParseClaudeLine(t *testing.T) {
	p := NewParser("claude", "session-123")

	t.Run("parses system event", func(t *testing.T) {
		line := `{"type": "system", "session_id": "backend-session-456"}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventInit {
			t.Errorf("expected type %q, got %q", EventInit, event.Type)
		}

		content, _ := event.GetInitContent()
		if content.BackendSessionID != "backend-session-456" {
			t.Errorf("expected backend session ID %q, got %q", "backend-session-456", content.BackendSessionID)
		}
	})

	t.Run("parses assistant message with text", func(t *testing.T) {
		line := `{"type": "assistant", "message": {"content": [{"type": "text", "text": "Hello!"}]}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventMessage {
			t.Errorf("expected type %q, got %q", EventMessage, event.Type)
		}

		content, _ := event.GetMessageContent()
		if content.Text != "Hello!" {
			t.Errorf("expected text %q, got %q", "Hello!", content.Text)
		}
	})

	t.Run("parses assistant tool use", func(t *testing.T) {
		line := `{"type": "assistant", "message": {"content": [{"type": "tool_use", "id": "tool-1", "name": "read_file", "input": {"path": "/tmp"}}]}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventToolUse {
			t.Errorf("expected type %q, got %q", EventToolUse, event.Type)
		}

		content, _ := event.GetToolUseContent()
		if content.ToolID != "tool-1" {
			t.Errorf("expected tool ID %q, got %q", "tool-1", content.ToolID)
		}
		if content.ToolName != "read_file" {
			t.Errorf("expected tool name %q, got %q", "read_file", content.ToolName)
		}
	})

	t.Run("parses content_block_delta text", func(t *testing.T) {
		line := `{"type": "content_block_delta", "delta": {"type": "text_delta", "text": "Hello"}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventMessage {
			t.Errorf("expected type %q, got %q", EventMessage, event.Type)
		}

		content, _ := event.GetMessageContent()
		if content.Text != "Hello" {
			t.Errorf("expected text %q, got %q", "Hello", content.Text)
		}
		if !content.IsPartial {
			t.Error("expected IsPartial to be true for delta")
		}
	})

	t.Run("parses content_block_delta thinking", func(t *testing.T) {
		line := `{"type": "content_block_delta", "delta": {"type": "thinking_delta", "thinking": "Let me think..."}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventThinking {
			t.Errorf("expected type %q, got %q", EventThinking, event.Type)
		}
	})

	t.Run("parses tool_result", func(t *testing.T) {
		line := `{"type": "tool_result", "tool_use_id": "tool-1", "content": "file content here"}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventToolResult {
			t.Errorf("expected type %q, got %q", EventToolResult, event.Type)
		}

		content, _ := event.GetToolResultContent()
		if content.ToolID != "tool-1" {
			t.Errorf("expected tool ID %q, got %q", "tool-1", content.ToolID)
		}
		if content.Output != "file content here" {
			t.Errorf("expected output %q, got %q", "file content here", content.Output)
		}
	})

	t.Run("parses error event", func(t *testing.T) {
		line := `{"type": "error", "error": {"message": "Rate limit exceeded"}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventError {
			t.Errorf("expected type %q, got %q", EventError, event.Type)
		}

		content, _ := event.GetErrorContent()
		if content.Message != "Rate limit exceeded" {
			t.Errorf("expected message %q, got %q", "Rate limit exceeded", content.Message)
		}
	})

	t.Run("parses message_stop event", func(t *testing.T) {
		line := `{"type": "message_stop", "usage": {"input_tokens": 100, "output_tokens": 200}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventDone {
			t.Errorf("expected type %q, got %q", EventDone, event.Type)
		}

		content, _ := event.GetDoneContent()
		if content.TokenUsage.InputTokens != 100 {
			t.Errorf("expected input tokens %d, got %d", 100, content.TokenUsage.InputTokens)
		}
	})

	t.Run("parses plain text as message", func(t *testing.T) {
		line := "This is plain text output"
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventMessage {
			t.Errorf("expected type %q, got %q", EventMessage, event.Type)
		}

		content, _ := event.GetMessageContent()
		if content.Text != "This is plain text output" {
			t.Errorf("expected text %q, got %q", "This is plain text output", content.Text)
		}
	})
}

func TestParser_ParseGeminiLine(t *testing.T) {
	p := NewParser("gemini", "session-123")

	t.Run("parses init event", func(t *testing.T) {
		line := `{"type": "init", "sessionId": "gemini-session-456"}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventInit {
			t.Errorf("expected type %q, got %q", EventInit, event.Type)
		}

		content, _ := event.GetInitContent()
		if content.BackendSessionID != "gemini-session-456" {
			t.Errorf("expected backend session ID %q, got %q", "gemini-session-456", content.BackendSessionID)
		}
	})

	t.Run("parses message event", func(t *testing.T) {
		line := `{"type": "message", "content": "Hello from Gemini!"}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventMessage {
			t.Errorf("expected type %q, got %q", EventMessage, event.Type)
		}

		content, _ := event.GetMessageContent()
		if content.Text != "Hello from Gemini!" {
			t.Errorf("expected text %q, got %q", "Hello from Gemini!", content.Text)
		}
	})

	t.Run("parses tool_use event", func(t *testing.T) {
		line := `{"type": "tool_use", "toolCallId": "call-1", "toolName": "execute_code", "parameters": {"code": "print(1)"}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventToolUse {
			t.Errorf("expected type %q, got %q", EventToolUse, event.Type)
		}

		content, _ := event.GetToolUseContent()
		if content.ToolID != "call-1" {
			t.Errorf("expected tool ID %q, got %q", "call-1", content.ToolID)
		}
		if content.ToolName != "execute_code" {
			t.Errorf("expected tool name %q, got %q", "execute_code", content.ToolName)
		}
	})

	t.Run("parses tool_result event", func(t *testing.T) {
		line := `{"type": "tool_result", "toolCallId": "call-1", "toolName": "execute_code", "result": "1"}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventToolResult {
			t.Errorf("expected type %q, got %q", EventToolResult, event.Type)
		}

		content, _ := event.GetToolResultContent()
		if content.Output != "1" {
			t.Errorf("expected output %q, got %q", "1", content.Output)
		}
	})

	t.Run("parses error event", func(t *testing.T) {
		line := `{"type": "error", "code": "INTERNAL_ERROR", "message": "Something went wrong", "details": "Stack trace..."}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventError {
			t.Errorf("expected type %q, got %q", EventError, event.Type)
		}

		content, _ := event.GetErrorContent()
		if content.Code != "INTERNAL_ERROR" {
			t.Errorf("expected code %q, got %q", "INTERNAL_ERROR", content.Code)
		}
		if content.Message != "Something went wrong" {
			t.Errorf("expected message %q, got %q", "Something went wrong", content.Message)
		}
	})

	t.Run("parses result event", func(t *testing.T) {
		line := `{"type": "result", "stats": {"tokenUsage": {"inputTokens": 50, "outputTokens": 100}}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventDone {
			t.Errorf("expected type %q, got %q", EventDone, event.Type)
		}

		content, _ := event.GetDoneContent()
		if content.TokenUsage.InputTokens != 50 {
			t.Errorf("expected input tokens %d, got %d", 50, content.TokenUsage.InputTokens)
		}
	})
}

func TestParser_ParseCodexLine(t *testing.T) {
	p := NewParser("codex", "session-123")

	t.Run("parses thread.started event", func(t *testing.T) {
		line := `{"type": "thread.started", "thread_id": "thread-123"}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventInit {
			t.Errorf("expected type %q, got %q", EventInit, event.Type)
		}

		content, _ := event.GetInitContent()
		if content.BackendSessionID != "thread-123" {
			t.Errorf("expected backend session ID %q, got %q", "thread-123", content.BackendSessionID)
		}
	})

	t.Run("parses turn.started event", func(t *testing.T) {
		line := `{"type": "turn.started"}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventProgress {
			t.Errorf("expected type %q, got %q", EventProgress, event.Type)
		}
	})

	t.Run("parses item.completed message", func(t *testing.T) {
		line := `{"type": "item.completed", "item": {"type": "message", "content": [{"type": "output_text", "text": "Hello from Codex"}]}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventMessage {
			t.Errorf("expected type %q, got %q", EventMessage, event.Type)
		}

		content, _ := event.GetMessageContent()
		if content.Text != "Hello from Codex" {
			t.Errorf("expected text %q, got %q", "Hello from Codex", content.Text)
		}
	})

	t.Run("parses item.completed function_call", func(t *testing.T) {
		line := `{"type": "item.completed", "item": {"type": "function_call", "call_id": "call-1", "name": "shell", "arguments": {"command": "ls"}}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventToolUse {
			t.Errorf("expected type %q, got %q", EventToolUse, event.Type)
		}

		content, _ := event.GetToolUseContent()
		if content.ToolID != "call-1" {
			t.Errorf("expected tool ID %q, got %q", "call-1", content.ToolID)
		}
		if content.ToolName != "shell" {
			t.Errorf("expected tool name %q, got %q", "shell", content.ToolName)
		}
	})

	t.Run("parses item.completed function_call_output", func(t *testing.T) {
		line := `{"type": "item.completed", "item": {"type": "function_call_output", "call_id": "call-1", "output": "file1.txt\nfile2.txt"}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventToolResult {
			t.Errorf("expected type %q, got %q", EventToolResult, event.Type)
		}

		content, _ := event.GetToolResultContent()
		if content.Output != "file1.txt\nfile2.txt" {
			t.Errorf("expected output %q, got %q", "file1.txt\nfile2.txt", content.Output)
		}
	})

	t.Run("parses turn.completed event", func(t *testing.T) {
		line := `{"type": "turn.completed", "usage": {"input_tokens": 150, "output_tokens": 300}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventDone {
			t.Errorf("expected type %q, got %q", EventDone, event.Type)
		}

		content, _ := event.GetDoneContent()
		if content.TokenUsage.InputTokens != 150 {
			t.Errorf("expected input tokens %d, got %d", 150, content.TokenUsage.InputTokens)
		}
	})

	t.Run("parses error event", func(t *testing.T) {
		line := `{"type": "error", "code": "context_length_exceeded", "message": "Context too long"}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventError {
			t.Errorf("expected type %q, got %q", EventError, event.Type)
		}

		content, _ := event.GetErrorContent()
		if content.Code != "context_length_exceeded" {
			t.Errorf("expected code %q, got %q", "context_length_exceeded", content.Code)
		}
	})

	t.Run("parses reasoning event", func(t *testing.T) {
		line := `{"type": "reasoning", "text": "Let me analyze this..."}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventThinking {
			t.Errorf("expected type %q, got %q", EventThinking, event.Type)
		}
	})

	t.Run("parses event_msg token_count", func(t *testing.T) {
		line := `{"type": "event_msg", "payload": {"type": "token_count", "input_tokens": 50, "output_tokens": 75}}`
		event, err := p.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if event.Type != EventTokenUsage {
			t.Errorf("expected type %q, got %q", EventTokenUsage, event.Type)
		}
	})
}

func TestParser_ParseStream(t *testing.T) {
	t.Run("parses multiple lines", func(t *testing.T) {
		input := `{"type": "system", "session_id": "s1"}
{"type": "assistant", "message": {"content": [{"type": "text", "text": "Hello"}]}}
{"type": "message_stop", "usage": {"input_tokens": 10, "output_tokens": 20}}`

		p := NewParser("claude", "session-123")
		eventCh := make(chan *UnifiedEvent, 10)
		errCh := make(chan error, 10)

		go p.ParseStream(strings.NewReader(input), eventCh, errCh)

		var events []*UnifiedEvent
		for event := range eventCh {
			events = append(events, event)
		}

		// Drain error channel
		for range errCh {
		}

		if len(events) != 3 {
			t.Errorf("expected 3 events, got %d", len(events))
		}

		if events[0].Type != EventInit {
			t.Errorf("expected first event to be %q, got %q", EventInit, events[0].Type)
		}
		if events[1].Type != EventMessage {
			t.Errorf("expected second event to be %q, got %q", EventMessage, events[1].Type)
		}
		if events[2].Type != EventDone {
			t.Errorf("expected third event to be %q, got %q", EventDone, events[2].Type)
		}
	})
}

// ==================== Writer Tests ====================

func TestNewWriter(t *testing.T) {
	t.Run("creates writer with defaults", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf)

		if w.format != FormatJSON {
			t.Errorf("expected default format %q, got %q", FormatJSON, w.format)
		}
		if w.pretty {
			t.Error("expected pretty to be false by default")
		}
	})

	t.Run("applies options", func(t *testing.T) {
		var buf bytes.Buffer
		var callbackCalled bool

		w := NewWriter(&buf,
			WithFormat(FormatText),
			WithPrettyPrint(true),
			WithEventCallback(func(*UnifiedEvent) { callbackCalled = true }),
		)

		if w.format != FormatText {
			t.Errorf("expected format %q, got %q", FormatText, w.format)
		}
		if !w.pretty {
			t.Error("expected pretty to be true")
		}

		// Test callback
		event := NewUnifiedEvent(EventMessage, "claude", "test")
		event.SetContent(&MessageContent{Text: "test"})
		w.WriteEvent(event)

		if !callbackCalled {
			t.Error("expected callback to be called")
		}
	})
}

func TestWriter_WriteEvent_JSON(t *testing.T) {
	t.Run("writes JSON format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatJSON))

		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		event.SetContent(&MessageContent{Text: "Hello"})

		err := w.WriteEvent(event)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"type":"message"`) {
			t.Errorf("expected JSON output, got: %s", output)
		}
	})

	t.Run("writes pretty JSON", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatJSON), WithPrettyPrint(true))

		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		event.SetContent(&MessageContent{Text: "Hello"})

		w.WriteEvent(event)

		output := buf.String()
		if !strings.Contains(output, "  ") {
			t.Errorf("expected indented JSON, got: %s", output)
		}
	})
}

func TestWriter_WriteEvent_Text(t *testing.T) {
	t.Run("writes message event", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatText))

		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		event.SetContent(&MessageContent{Text: "Hello world", Role: "assistant"})

		w.WriteEvent(event)

		output := buf.String()
		if !strings.Contains(output, "Hello world") {
			t.Errorf("expected text output, got: %s", output)
		}
	})

	t.Run("writes partial message without newline", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatText))

		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		event.SetContent(&MessageContent{Text: "Hello", IsPartial: true})

		w.WriteEvent(event)

		output := buf.String()
		if strings.HasSuffix(output, "\n") {
			t.Errorf("partial message should not have newline, got: %q", output)
		}
	})

	t.Run("writes tool use event", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatText))

		event := NewUnifiedEvent(EventToolUse, "claude", "session-123")
		event.SetContent(&ToolUseContent{ToolName: "read_file"})

		w.WriteEvent(event)

		output := buf.String()
		if !strings.Contains(output, "[tool: read_file]") {
			t.Errorf("expected tool output, got: %s", output)
		}
	})

	t.Run("writes error event", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatText))

		event := NewUnifiedEvent(EventError, "claude", "session-123")
		event.SetContent(&ErrorContent{Message: "Something went wrong"})

		w.WriteEvent(event)

		output := buf.String()
		if !strings.Contains(output, "Error: Something went wrong") {
			t.Errorf("expected error output, got: %s", output)
		}
	})

	t.Run("writes done event with tokens", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatText))

		event := NewUnifiedEvent(EventDone, "claude", "session-123")
		event.SetContent(&DoneContent{
			TokenUsage: &TokenUsageContent{InputTokens: 100, OutputTokens: 200},
		})

		w.WriteEvent(event)

		output := buf.String()
		if !strings.Contains(output, "100 in") || !strings.Contains(output, "200 out") {
			t.Errorf("expected token output, got: %s", output)
		}
	})
}

func TestWriter_WriteEvent_Quiet(t *testing.T) {
	t.Run("only writes complete messages", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatQuiet))

		// Partial message - should not be written
		event1 := NewUnifiedEvent(EventMessage, "claude", "session-123")
		event1.SetContent(&MessageContent{Text: "partial", IsPartial: true})
		w.WriteEvent(event1)

		if buf.Len() != 0 {
			t.Error("partial message should not be written in quiet mode")
		}

		// Complete message - should be written
		event2 := NewUnifiedEvent(EventMessage, "claude", "session-123")
		event2.SetContent(&MessageContent{Text: "complete", IsPartial: false})
		w.WriteEvent(event2)

		if !strings.Contains(buf.String(), "complete") {
			t.Errorf("complete message should be written, got: %s", buf.String())
		}
	})

	t.Run("ignores non-message events", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, WithFormat(FormatQuiet))

		event := NewUnifiedEvent(EventToolUse, "claude", "session-123")
		event.SetContent(&ToolUseContent{ToolName: "read_file"})
		w.WriteEvent(event)

		if buf.Len() != 0 {
			t.Error("non-message events should not be written in quiet mode")
		}
	})
}

// ==================== Collector Tests ====================

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Error("expected non-nil collector")
	}
	if len(c.Events()) != 0 {
		t.Error("expected empty events")
	}
}

func TestCollector_Collect(t *testing.T) {
	c := NewCollector()

	event1 := NewUnifiedEvent(EventMessage, "claude", "session-123")
	event2 := NewUnifiedEvent(EventToolUse, "claude", "session-123")

	c.Collect(event1)
	c.Collect(event2)

	events := c.Events()
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestCollector_FilterByType(t *testing.T) {
	c := NewCollector()

	event1 := NewUnifiedEvent(EventMessage, "claude", "session-123")
	event2 := NewUnifiedEvent(EventToolUse, "claude", "session-123")
	event3 := NewUnifiedEvent(EventMessage, "claude", "session-123")

	c.Collect(event1)
	c.Collect(event2)
	c.Collect(event3)

	messages := c.FilterByType(EventMessage)
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	toolUses := c.FilterByType(EventToolUse)
	if len(toolUses) != 1 {
		t.Errorf("expected 1 tool use, got %d", len(toolUses))
	}
}

func TestCollector_Messages(t *testing.T) {
	c := NewCollector()

	event1 := NewUnifiedEvent(EventMessage, "claude", "session-123")
	event1.SetContent(&MessageContent{Text: "Hello "})

	event2 := NewUnifiedEvent(EventToolUse, "claude", "session-123")
	event2.SetContent(&ToolUseContent{ToolName: "test"})

	event3 := NewUnifiedEvent(EventMessage, "claude", "session-123")
	event3.SetContent(&MessageContent{Text: "World"})

	c.Collect(event1)
	c.Collect(event2)
	c.Collect(event3)

	messages := c.Messages()
	if messages != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", messages)
	}
}

func TestCollector_LastError(t *testing.T) {
	t.Run("returns nil when no errors", func(t *testing.T) {
		c := NewCollector()

		event := NewUnifiedEvent(EventMessage, "claude", "session-123")
		c.Collect(event)

		if c.LastError() != nil {
			t.Error("expected nil when no errors")
		}
	})

	t.Run("returns last error", func(t *testing.T) {
		c := NewCollector()

		event1 := NewUnifiedEvent(EventError, "claude", "session-123")
		event1.SetContent(&ErrorContent{Message: "First error"})

		event2 := NewUnifiedEvent(EventMessage, "claude", "session-123")

		event3 := NewUnifiedEvent(EventError, "claude", "session-123")
		event3.SetContent(&ErrorContent{Message: "Last error"})

		c.Collect(event1)
		c.Collect(event2)
		c.Collect(event3)

		lastErr := c.LastError()
		if lastErr == nil {
			t.Error("expected non-nil error")
		}
		if lastErr.Message != "Last error" {
			t.Errorf("expected 'Last error', got %q", lastErr.Message)
		}
	})
}

func TestCollector_TotalTokens(t *testing.T) {
	c := NewCollector()

	// Add done event with tokens
	event1 := NewUnifiedEvent(EventDone, "claude", "session-123")
	event1.SetContent(&DoneContent{
		TokenUsage: &TokenUsageContent{InputTokens: 100, OutputTokens: 200},
	})

	// Add token usage event
	event2 := NewUnifiedEvent(EventTokenUsage, "claude", "session-123")
	event2.SetContent(&TokenUsageContent{InputTokens: 50, OutputTokens: 75})

	c.Collect(event1)
	c.Collect(event2)

	input, output := c.TotalTokens()
	if input != 150 {
		t.Errorf("expected input tokens 150, got %d", input)
	}
	if output != 275 {
		t.Errorf("expected output tokens 275, got %d", output)
	}
}

func TestCollector_Clear(t *testing.T) {
	c := NewCollector()

	event := NewUnifiedEvent(EventMessage, "claude", "session-123")
	c.Collect(event)

	if len(c.Events()) != 1 {
		t.Error("expected 1 event before clear")
	}

	c.Clear()

	if len(c.Events()) != 0 {
		t.Error("expected 0 events after clear")
	}
}

// ==================== Helper Function Tests ====================

func TestGetString(t *testing.T) {
	m := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": nil,
	}

	if getString(m, "key1") != "value1" {
		t.Error("expected 'value1'")
	}
	if getString(m, "key2") != "" {
		t.Error("expected empty string for non-string")
	}
	if getString(m, "key3") != "" {
		t.Error("expected empty string for nil")
	}
	if getString(m, "nonexistent") != "" {
		t.Error("expected empty string for nonexistent key")
	}
}

func TestGetInt64(t *testing.T) {
	m := map[string]any{
		"float":  float64(100),
		"int64":  int64(200),
		"int":    int(300),
		"string": "400",
	}

	if getInt64(m, "float") != 100 {
		t.Error("expected 100 for float64")
	}
	if getInt64(m, "int64") != 200 {
		t.Error("expected 200 for int64")
	}
	if getInt64(m, "int") != 300 {
		t.Error("expected 300 for int")
	}
	if getInt64(m, "string") != 0 {
		t.Error("expected 0 for string")
	}
	if getInt64(m, "nonexistent") != 0 {
		t.Error("expected 0 for nonexistent key")
	}
	if getInt64(nil, "any") != 0 {
		t.Error("expected 0 for nil map")
	}
}
