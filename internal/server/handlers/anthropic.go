package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/server/service"
)

// AnthropicHandlers provides handlers for Anthropic-compatible API.
type AnthropicHandlers struct {
	executor *service.Executor
}

// NewAnthropicHandlers creates a new Anthropic handlers instance.
func NewAnthropicHandlers(executor *service.Executor) *AnthropicHandlers {
	return &AnthropicHandlers{executor: executor}
}

// Register registers all Anthropic-compatible API routes.
// Endpoints follow Anthropic API spec: https://docs.anthropic.com/en/api/messages
func (h *AnthropicHandlers) Register(api huma.API) {
	// Messages endpoint - POST /anthropic/v1/messages
	huma.Register(api, huma.Operation{
		OperationID: "anthropicCreateMessage",
		Method:      http.MethodPost,
		Path:        "/anthropic/v1/messages",
		Summary:     "Create a Message",
		Description: "Send a structured list of input messages with text and/or image content, and the model will generate the next message in the conversation. Compatible with Anthropic POST /v1/messages.",
		Tags:        []string{"Anthropic Compatible"},
	}, h.HandleMessages)
}

// AnthropicMessage represents an Anthropic message.
type AnthropicMessage struct {
	Role    string `json:"role" doc:"Message role (user, assistant)"`
	Content string `json:"content" doc:"Message content"`
}

// AnthropicMessagesRequest is the request for creating messages.
type AnthropicMessagesRequest struct {
	Model         string             `json:"model" doc:"Model to use"`
	MaxTokens     int                `json:"max_tokens" doc:"Maximum tokens to generate"`
	Messages      []AnthropicMessage `json:"messages" doc:"Conversation messages"`
	System        string             `json:"system,omitempty" doc:"System prompt"`
	StopSequences []string           `json:"stop_sequences,omitempty" doc:"Stop sequences"`
	Stream        bool               `json:"stream,omitempty" doc:"Stream responses (not supported)"`
	Temperature   float64            `json:"temperature,omitempty" doc:"Sampling temperature"`
	TopP          float64            `json:"top_p,omitempty" doc:"Nucleus sampling parameter"`
	TopK          int                `json:"top_k,omitempty" doc:"Top-k sampling parameter"`
	Metadata      map[string]string  `json:"metadata,omitempty" doc:"Request metadata"`
}

// AnthropicContentBlock represents a content block in the response.
type AnthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// AnthropicUsage represents token usage.
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicMessagesResponse is the response for creating messages.
type AnthropicMessagesResponse struct {
	Body AnthropicMessagesResponseBody
}

// AnthropicMessagesResponseBody is the body of the messages response.
type AnthropicMessagesResponseBody struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []AnthropicContentBlock `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason"`
	StopSequence string                  `json:"stop_sequence,omitempty"`
	Usage        AnthropicUsage          `json:"usage"`
}

// AnthropicMessagesInput is the input for the messages handler.
type AnthropicMessagesInput struct {
	Body AnthropicMessagesRequest
}

// HandleMessages handles the POST /v1/messages endpoint.
func (h *AnthropicHandlers) HandleMessages(ctx context.Context, input *AnthropicMessagesInput) (*AnthropicMessagesResponse, error) {
	if input.Body.Model == "" {
		return nil, huma.Error400BadRequest("model is required")
	}
	if len(input.Body.Messages) == 0 {
		return nil, huma.Error400BadRequest("messages are required")
	}
	if input.Body.MaxTokens <= 0 {
		return nil, huma.Error400BadRequest("max_tokens must be greater than 0")
	}

	// Streaming is not supported
	if input.Body.Stream {
		return nil, huma.Error400BadRequest("streaming is not supported")
	}

	// Extract prompt from messages
	var prompt string
	for _, msg := range input.Body.Messages {
		if msg.Role == "user" {
			if prompt != "" {
				prompt += "\n"
			}
			prompt += msg.Content
		} else if msg.Role == "assistant" {
			// Include assistant context for continuations
			if prompt != "" {
				prompt += "\n[Previous response: " + msg.Content + "]\n"
			}
		}
	}

	if prompt == "" {
		return nil, huma.Error400BadRequest("no user messages found")
	}

	// Map model to backend
	backendName := mapAnthropicModelToBackend(input.Body.Model)

	// Execute prompt
	req := &service.PromptRequest{
		Backend:      backendName,
		Prompt:       prompt,
		Model:        input.Body.Model,
		MaxTokens:    input.Body.MaxTokens,
		SystemPrompt: input.Body.System,
		Metadata:     input.Body.Metadata,
	}

	result, err := h.executor.ExecutePrompt(ctx, req)
	if err != nil {
		return nil, huma.Error500InternalServerError("execution failed", err)
	}

	// Build response
	now := time.Now().Unix()
	responseID := fmt.Sprintf("msg_%s", result.SessionID)
	if result.SessionID == "" {
		responseID = fmt.Sprintf("msg_%d", now)
	}

	stopReason := "end_turn"
	if result.ExitCode != 0 {
		stopReason = "error"
	}

	// Estimate token counts
	inputTokens := len(prompt) / 4
	outputTokens := len(result.Output) / 4

	return &AnthropicMessagesResponse{
		Body: AnthropicMessagesResponseBody{
			ID:   responseID,
			Type: "message",
			Role: "assistant",
			Content: []AnthropicContentBlock{
				{
					Type: "text",
					Text: result.Output,
				},
			},
			Model:      input.Body.Model,
			StopReason: stopReason,
			Usage: AnthropicUsage{
				InputTokens:  inputTokens,
				OutputTokens: outputTokens,
			},
		},
	}, nil
}

// mapAnthropicModelToBackend maps Anthropic model names to backend names.
func mapAnthropicModelToBackend(model string) string {
	// If the model is already a backend name, use it
	switch model {
	case backend.BackendClaude, backend.BackendCodex, backend.BackendGemini:
		return model
	}

	// Map Anthropic-style model names to backends
	if strings.Contains(model, "claude") {
		return backend.BackendClaude
	}

	// Default to claude for Anthropic API
	return backend.BackendClaude
}
