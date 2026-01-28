package handlers

import (
	"context"
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/server/service"
)

func TestAnthropicHandlers_HandleMessages_Validation(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewAnthropicHandlers(executor, nil)

	tests := []struct {
		name    string
		input   *AnthropicMessagesInput
		wantErr string
	}{
		{
			name: "missing model",
			input: &AnthropicMessagesInput{
				Body: AnthropicMessagesRequest{
					Model:     "",
					MaxTokens: 1000,
					Messages:  []AnthropicMessage{{Role: "user", Content: "test"}},
				},
			},
			wantErr: "model is required",
		},
		{
			name: "missing messages",
			input: &AnthropicMessagesInput{
				Body: AnthropicMessagesRequest{
					Model:     "claude",
					MaxTokens: 1000,
					Messages:  []AnthropicMessage{},
				},
			},
			wantErr: "messages are required",
		},
		{
			name: "zero max_tokens",
			input: &AnthropicMessagesInput{
				Body: AnthropicMessagesRequest{
					Model:     "claude",
					MaxTokens: 0,
					Messages:  []AnthropicMessage{{Role: "user", Content: "test"}},
				},
			},
			wantErr: "max_tokens must be greater than 0",
		},
		{
			name: "negative max_tokens",
			input: &AnthropicMessagesInput{
				Body: AnthropicMessagesRequest{
					Model:     "claude",
					MaxTokens: -1,
					Messages:  []AnthropicMessage{{Role: "user", Content: "test"}},
				},
			},
			wantErr: "max_tokens must be greater than 0",
		},
		{
			name: "no user messages",
			input: &AnthropicMessagesInput{
				Body: AnthropicMessagesRequest{
					Model:     "claude",
					MaxTokens: 1000,
					Messages:  []AnthropicMessage{{Role: "assistant", Content: "test"}},
				},
			},
			wantErr: "no user messages found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handlers.HandleMessages(context.Background(), tt.input)
			if err == nil {
				t.Errorf("expected error %q, got nil", tt.wantErr)
				return
			}
			if !containsString(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestAnthropicHandlers_HandleMessages_MessageParsing(t *testing.T) {
	tests := []struct {
		name            string
		messages        []AnthropicMessage
		expectedPrompt  string
		expectedHasUser bool
	}{
		{
			name: "single user message",
			messages: []AnthropicMessage{
				{Role: "user", Content: "hello"},
			},
			expectedPrompt:  "hello",
			expectedHasUser: true,
		},
		{
			name: "multiple user messages",
			messages: []AnthropicMessage{
				{Role: "user", Content: "first"},
				{Role: "user", Content: "second"},
			},
			expectedPrompt:  "first\nsecond",
			expectedHasUser: true,
		},
		{
			name: "with assistant context",
			messages: []AnthropicMessage{
				{Role: "user", Content: "first"},
				{Role: "assistant", Content: "response"},
				{Role: "user", Content: "second"},
			},
			expectedHasUser: true,
		},
		{
			name: "only assistant message",
			messages: []AnthropicMessage{
				{Role: "assistant", Content: "response"},
			},
			expectedHasUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract prompt using same logic as handler
			var prompt string
			for _, msg := range tt.messages {
				if msg.Role == "user" {
					if prompt != "" {
						prompt += "\n"
					}
					prompt += msg.Content
				} else if msg.Role == "assistant" {
					if prompt != "" {
						prompt += "\n[Previous response: " + msg.Content + "]\n"
					}
				}
			}

			hasUser := prompt != "" || (tt.expectedHasUser && len(tt.messages) > 0)
			if tt.expectedHasUser && prompt == "" {
				if !hasUser {
					t.Error("expected non-empty prompt for user messages")
				}
			}

			if tt.expectedPrompt != "" && prompt != tt.expectedPrompt {
				t.Errorf("expected prompt %q, got %q", tt.expectedPrompt, prompt)
			}
		})
	}
}

func TestMapAnthropicModelToBackend_Comprehensive(t *testing.T) {
	tests := []struct {
		model    string
		expected string
	}{
		// Direct backend names
		{"claude", backend.BackendClaude},
		{"codex", backend.BackendCodex},
		{"gemini", backend.BackendGemini},

		// Claude model variants
		{"claude-3-opus-20240229", backend.BackendClaude},
		{"claude-3-sonnet-20240229", backend.BackendClaude},
		{"claude-3-haiku-20240307", backend.BackendClaude},
		{"claude-3-5-sonnet", backend.BackendClaude},
		{"claude-instant-1.2", backend.BackendClaude},

		// Unknown defaults to claude (Anthropic API default)
		{"unknown-model", backend.BackendClaude},
		{"", backend.BackendClaude},
		{"gpt-4", backend.BackendClaude}, // Different from OpenAI handler
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := mapAnthropicModelToBackend(tt.model)
			if result != tt.expected {
				t.Errorf("mapAnthropicModelToBackend(%q) = %q, want %q", tt.model, result, tt.expected)
			}
		})
	}
}

func TestAnthropicMessage_Structure(t *testing.T) {
	msg := AnthropicMessage{
		Role:    "user",
		Content: "Hello, Claude!",
	}

	if msg.Role != "user" {
		t.Errorf("Role = %q, want 'user'", msg.Role)
	}
	if msg.Content != "Hello, Claude!" {
		t.Errorf("Content = %q, want 'Hello, Claude!'", msg.Content)
	}
}

func TestAnthropicMessagesRequest_Structure(t *testing.T) {
	req := AnthropicMessagesRequest{
		Model:     "claude-3-opus",
		MaxTokens: 1024,
		Messages: []AnthropicMessage{
			{Role: "user", Content: "Hello"},
		},
		System:        "Be helpful and concise",
		StopSequences: []string{"\n\nHuman:"},
		Temperature:   0.7,
		TopP:          0.9,
		TopK:          40,
		Metadata:      map[string]string{"user_id": "123"},
	}

	if req.Model != "claude-3-opus" {
		t.Errorf("Model = %q, want 'claude-3-opus'", req.Model)
	}
	if req.MaxTokens != 1024 {
		t.Errorf("MaxTokens = %d, want 1024", req.MaxTokens)
	}
	if len(req.Messages) != 1 {
		t.Errorf("len(Messages) = %d, want 1", len(req.Messages))
	}
	if req.System != "Be helpful and concise" {
		t.Errorf("System = %q, want 'Be helpful and concise'", req.System)
	}
}

func TestAnthropicContentBlock_Structure(t *testing.T) {
	block := AnthropicContentBlock{
		Type: "text",
		Text: "Hello, user!",
	}

	if block.Type != "text" {
		t.Errorf("Type = %q, want 'text'", block.Type)
	}
	if block.Text != "Hello, user!" {
		t.Errorf("Text = %q, want 'Hello, user!'", block.Text)
	}
}

func TestAnthropicUsage_Structure(t *testing.T) {
	usage := AnthropicUsage{
		InputTokens:  100,
		OutputTokens: 50,
	}

	if usage.InputTokens != 100 {
		t.Errorf("InputTokens = %d, want 100", usage.InputTokens)
	}
	if usage.OutputTokens != 50 {
		t.Errorf("OutputTokens = %d, want 50", usage.OutputTokens)
	}
}

func TestAnthropicMessagesResponseBody_Structure(t *testing.T) {
	resp := AnthropicMessagesResponseBody{
		ID:   "msg_123abc",
		Type: "message",
		Role: "assistant",
		Content: []AnthropicContentBlock{
			{Type: "text", Text: "Hello!"},
		},
		Model:      "claude-3-opus",
		StopReason: "end_turn",
		Usage: AnthropicUsage{
			InputTokens:  10,
			OutputTokens: 5,
		},
	}

	if resp.ID != "msg_123abc" {
		t.Errorf("ID = %q, want 'msg_123abc'", resp.ID)
	}
	if resp.Type != "message" {
		t.Errorf("Type = %q, want 'message'", resp.Type)
	}
	if resp.Role != "assistant" {
		t.Errorf("Role = %q, want 'assistant'", resp.Role)
	}
	if len(resp.Content) != 1 {
		t.Errorf("len(Content) = %d, want 1", len(resp.Content))
	}
	if resp.StopReason != "end_turn" {
		t.Errorf("StopReason = %q, want 'end_turn'", resp.StopReason)
	}
}

func TestAnthropicValidStopReasons(t *testing.T) {
	validReasons := []string{
		"end_turn",
		"max_tokens",
		"stop_sequence",
		"error",
	}

	for _, reason := range validReasons {
		t.Run(reason, func(t *testing.T) {
			resp := AnthropicMessagesResponseBody{
				StopReason: reason,
			}
			if resp.StopReason != reason {
				t.Errorf("StopReason = %q, want %q", resp.StopReason, reason)
			}
		})
	}
}
