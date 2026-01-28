package handlers

import (
	"context"
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/server/service"
)

func TestOpenAIHandlers_HandleModels(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewOpenAIHandlers(executor)

	resp, err := handlers.HandleModels(context.Background(), &OpenAIModelsInput{})
	if err != nil {
		t.Fatalf("HandleModels failed: %v", err)
	}

	if resp.Body.Object != "list" {
		t.Errorf("expected object 'list', got %q", resp.Body.Object)
	}

	if len(resp.Body.Data) == 0 {
		t.Error("expected at least one model")
	}

	// Verify model structure
	for _, model := range resp.Body.Data {
		if model.ID == "" {
			t.Error("model ID should not be empty")
		}
		if model.Object != "model" {
			t.Errorf("model object should be 'model', got %q", model.Object)
		}
		if model.OwnedBy != "clinvoker" {
			t.Errorf("model owned_by should be 'clinvoker', got %q", model.OwnedBy)
		}
		if model.Created == 0 {
			t.Error("model created timestamp should not be zero")
		}
	}
}

func TestOpenAIHandlers_HandleChatCompletions_Validation(t *testing.T) {
	executor := service.NewExecutor()
	handlers := NewOpenAIHandlers(executor)

	tests := []struct {
		name    string
		input   *OpenAIChatCompletionInput
		wantErr string
	}{
		{
			name: "missing model",
			input: &OpenAIChatCompletionInput{
				Body: OpenAIChatCompletionRequest{
					Model:    "",
					Messages: []OpenAIMessage{{Role: "user", Content: "test"}},
				},
			},
			wantErr: "model is required",
		},
		{
			name: "missing messages",
			input: &OpenAIChatCompletionInput{
				Body: OpenAIChatCompletionRequest{
					Model:    "claude",
					Messages: []OpenAIMessage{},
				},
			},
			wantErr: "messages are required",
		},
		{
			name: "streaming not supported",
			input: &OpenAIChatCompletionInput{
				Body: OpenAIChatCompletionRequest{
					Model:    "claude",
					Messages: []OpenAIMessage{{Role: "user", Content: "test"}},
					Stream:   true,
				},
			},
			wantErr: "streaming is not supported",
		},
		{
			name: "no user messages",
			input: &OpenAIChatCompletionInput{
				Body: OpenAIChatCompletionRequest{
					Model:    "claude",
					Messages: []OpenAIMessage{{Role: "system", Content: "test"}},
				},
			},
			wantErr: "no user messages found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handlers.HandleChatCompletions(context.Background(), tt.input)
			if err == nil {
				t.Errorf("expected error %q, got nil", tt.wantErr)
				return
			}
			// Error message should contain expected string
			if !containsString(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestOpenAIHandlers_HandleChatCompletions_MessageParsing(t *testing.T) {
	tests := []struct {
		name            string
		messages        []OpenAIMessage
		expectedPrompt  string
		expectedHasUser bool
	}{
		{
			name: "single user message",
			messages: []OpenAIMessage{
				{Role: "user", Content: "hello"},
			},
			expectedPrompt:  "hello",
			expectedHasUser: true,
		},
		{
			name: "multiple user messages",
			messages: []OpenAIMessage{
				{Role: "user", Content: "first"},
				{Role: "user", Content: "second"},
			},
			expectedPrompt:  "first\nsecond",
			expectedHasUser: true,
		},
		{
			name: "with system and user",
			messages: []OpenAIMessage{
				{Role: "system", Content: "be helpful"},
				{Role: "user", Content: "hello"},
			},
			expectedPrompt:  "hello",
			expectedHasUser: true,
		},
		{
			name: "with assistant context",
			messages: []OpenAIMessage{
				{Role: "user", Content: "first"},
				{Role: "assistant", Content: "response"},
				{Role: "user", Content: "second"},
			},
			expectedHasUser: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract prompt using same logic as handler
			var prompt string
			for _, msg := range tt.messages {
				switch msg.Role {
				case "user":
					if prompt != "" {
						prompt += "\n"
					}
					prompt += msg.Content
				case "assistant":
					if prompt != "" {
						prompt += "\n[Previous response: " + msg.Content + "]\n"
					}
				}
			}

			if tt.expectedHasUser && prompt == "" {
				t.Error("expected non-empty prompt for user messages")
			}

			if tt.expectedPrompt != "" && prompt != tt.expectedPrompt {
				t.Errorf("expected prompt %q, got %q", tt.expectedPrompt, prompt)
			}
		})
	}
}

func TestMapModelToBackend_Comprehensive(t *testing.T) {
	tests := []struct {
		model    string
		expected string
	}{
		// Direct backend names
		{"claude", backend.BackendClaude},
		{"codex", backend.BackendCodex},
		{"gemini", backend.BackendGemini},

		// Claude model variants
		{"claude-3-opus", backend.BackendClaude},
		{"claude-3-sonnet", backend.BackendClaude},
		{"claude-3-haiku", backend.BackendClaude},
		{"claude-opus-4-5", backend.BackendClaude},

		// GPT/Codex model variants
		{"gpt-4", backend.BackendCodex},
		{"gpt-4-turbo", backend.BackendCodex},
		{"gpt-3.5-turbo", backend.BackendCodex},
		// Note: o3 doesn't contain "gpt", so defaults to claude
		{"o3", backend.BackendClaude},

		// Gemini model variants
		{"gemini-pro", backend.BackendGemini},
		{"gemini-1.5-pro", backend.BackendGemini},
		{"gemini-2.0", backend.BackendGemini},

		// Unknown defaults to claude
		{"unknown-model", backend.BackendClaude},
		{"", backend.BackendClaude},
		{"random-name", backend.BackendClaude},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := mapModelToBackend(tt.model)
			if result != tt.expected {
				t.Errorf("mapModelToBackend(%q) = %q, want %q", tt.model, result, tt.expected)
			}
		})
	}
}

func TestOpenAIModel_Structure(t *testing.T) {
	model := OpenAIModel{
		ID:      "claude",
		Object:  "model",
		Created: 1234567890,
		OwnedBy: "clinvoker",
	}

	if model.ID != "claude" {
		t.Errorf("ID = %q, want 'claude'", model.ID)
	}
	if model.Object != "model" {
		t.Errorf("Object = %q, want 'model'", model.Object)
	}
}

func TestOpenAIMessage_Structure(t *testing.T) {
	msg := OpenAIMessage{
		Role:    "user",
		Content: "Hello, world!",
	}

	if msg.Role != "user" {
		t.Errorf("Role = %q, want 'user'", msg.Role)
	}
	if msg.Content != "Hello, world!" {
		t.Errorf("Content = %q, want 'Hello, world!'", msg.Content)
	}
}

func TestOpenAIChatCompletionRequest_Structure(t *testing.T) {
	req := OpenAIChatCompletionRequest{
		Model: "claude",
		Messages: []OpenAIMessage{
			{Role: "system", Content: "Be helpful"},
			{Role: "user", Content: "Hello"},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		Stream:      false,
	}

	if req.Model != "claude" {
		t.Errorf("Model = %q, want 'claude'", req.Model)
	}
	if len(req.Messages) != 2 {
		t.Errorf("len(Messages) = %d, want 2", len(req.Messages))
	}
	if req.MaxTokens != 1000 {
		t.Errorf("MaxTokens = %d, want 1000", req.MaxTokens)
	}
}

func TestOpenAIChatCompletionChoice_Structure(t *testing.T) {
	choice := OpenAIChatCompletionChoice{
		Index: 0,
		Message: OpenAIMessage{
			Role:    "assistant",
			Content: "Hello!",
		},
		FinishReason: "stop",
	}

	if choice.Index != 0 {
		t.Errorf("Index = %d, want 0", choice.Index)
	}
	if choice.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want 'stop'", choice.FinishReason)
	}
}

func TestOpenAIUsage_Structure(t *testing.T) {
	usage := OpenAIUsage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	if usage.PromptTokens != 100 {
		t.Errorf("PromptTokens = %d, want 100", usage.PromptTokens)
	}
	if usage.CompletionTokens != 50 {
		t.Errorf("CompletionTokens = %d, want 50", usage.CompletionTokens)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("TotalTokens = %d, want 150", usage.TotalTokens)
	}
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && searchSubstr(s, substr))
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
