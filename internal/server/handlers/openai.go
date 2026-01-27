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

// OpenAIHandlers provides handlers for OpenAI-compatible API.
type OpenAIHandlers struct {
	executor *service.Executor
}

// NewOpenAIHandlers creates a new OpenAI handlers instance.
func NewOpenAIHandlers(executor *service.Executor) *OpenAIHandlers {
	return &OpenAIHandlers{executor: executor}
}

// Register registers all OpenAI-compatible API routes.
// Endpoints follow OpenAI API spec: https://platform.openai.com/docs/api-reference
func (h *OpenAIHandlers) Register(api huma.API) {
	// Models endpoint - GET /openai/v1/models
	huma.Register(api, huma.Operation{
		OperationID: "openaiListModels",
		Method:      http.MethodGet,
		Path:        "/openai/v1/models",
		Summary:     "List models",
		Description: "Lists the currently available models. Compatible with OpenAI GET /v1/models.",
		Tags:        []string{"OpenAI Compatible"},
	}, h.HandleModels)

	// Chat completions endpoint - POST /openai/v1/chat/completions
	huma.Register(api, huma.Operation{
		OperationID: "openaiChatCompletions",
		Method:      http.MethodPost,
		Path:        "/openai/v1/chat/completions",
		Summary:     "Create chat completion",
		Description: "Creates a model response for the given chat conversation. Compatible with OpenAI POST /v1/chat/completions.",
		Tags:        []string{"OpenAI Compatible"},
	}, h.HandleChatCompletions)
}

// OpenAIModel represents an OpenAI model object.
type OpenAIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// OpenAIModelsResponse is the response for listing models.
type OpenAIModelsResponse struct {
	Body OpenAIModelsResponseBody
}

// OpenAIModelsResponseBody is the body of the models response.
type OpenAIModelsResponseBody struct {
	Object string        `json:"object"`
	Data   []OpenAIModel `json:"data"`
}

// OpenAIModelsInput is the input for the models handler.
type OpenAIModelsInput struct{}

// HandleModels handles the GET /v1/models endpoint.
func (h *OpenAIHandlers) HandleModels(ctx context.Context, _ *OpenAIModelsInput) (*OpenAIModelsResponse, error) {
	backends := backend.List()
	created := time.Now().Unix()

	models := make([]OpenAIModel, len(backends))
	for i, name := range backends {
		models[i] = OpenAIModel{
			ID:      name,
			Object:  "model",
			Created: created,
			OwnedBy: "clinvoker",
		}
	}

	return &OpenAIModelsResponse{
		Body: OpenAIModelsResponseBody{
			Object: "list",
			Data:   models,
		},
	}, nil
}

// OpenAIMessage represents a chat message.
type OpenAIMessage struct {
	Role    string `json:"role" doc:"Message role (system, user, assistant)"`
	Content string `json:"content" doc:"Message content"`
}

// OpenAIChatCompletionRequest is the request for chat completions.
type OpenAIChatCompletionRequest struct {
	Model            string          `json:"model" doc:"Model/backend to use"`
	Messages         []OpenAIMessage `json:"messages" doc:"Chat messages"`
	MaxTokens        int             `json:"max_tokens,omitempty" doc:"Maximum tokens to generate"`
	Temperature      float64         `json:"temperature,omitempty" doc:"Sampling temperature"`
	TopP             float64         `json:"top_p,omitempty" doc:"Nucleus sampling parameter"`
	N                int             `json:"n,omitempty" doc:"Number of completions"`
	Stream           bool            `json:"stream,omitempty" doc:"Stream responses (not supported)"`
	Stop             []string        `json:"stop,omitempty" doc:"Stop sequences"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty" doc:"Presence penalty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty" doc:"Frequency penalty"`
	User             string          `json:"user,omitempty" doc:"User identifier"`
}

// OpenAIChatCompletionChoice represents a completion choice.
type OpenAIChatCompletionChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage represents token usage.
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIChatCompletionResponse is the response for chat completions.
type OpenAIChatCompletionResponse struct {
	Body OpenAIChatCompletionResponseBody
}

// OpenAIChatCompletionResponseBody is the body of the chat completion response.
type OpenAIChatCompletionResponseBody struct {
	ID                string                       `json:"id"`
	Object            string                       `json:"object"`
	Created           int64                        `json:"created"`
	Model             string                       `json:"model"`
	Choices           []OpenAIChatCompletionChoice `json:"choices"`
	Usage             OpenAIUsage                  `json:"usage"`
	SystemFingerprint string                       `json:"system_fingerprint,omitempty"`
}

// OpenAIChatCompletionInput is the input for the chat completions handler.
type OpenAIChatCompletionInput struct {
	Body OpenAIChatCompletionRequest
}

// HandleChatCompletions handles the POST /v1/chat/completions endpoint.
func (h *OpenAIHandlers) HandleChatCompletions(ctx context.Context, input *OpenAIChatCompletionInput) (*OpenAIChatCompletionResponse, error) {
	if input.Body.Model == "" {
		return nil, huma.Error400BadRequest("model is required")
	}
	if len(input.Body.Messages) == 0 {
		return nil, huma.Error400BadRequest("messages are required")
	}

	// Streaming is not supported
	if input.Body.Stream {
		return nil, huma.Error400BadRequest("streaming is not supported")
	}

	// Extract prompt from messages
	// Combine all user messages as the prompt
	var prompt string
	var systemPrompt string
	for _, msg := range input.Body.Messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			if prompt != "" {
				prompt += "\n"
			}
			prompt += msg.Content
		case "assistant":
			// Include assistant context in prompt for continuations
			if prompt != "" {
				prompt += "\n[Previous response: " + msg.Content + "]\n"
			}
		}
	}

	if prompt == "" {
		return nil, huma.Error400BadRequest("no user messages found")
	}

	// Map model to backend
	backendName := mapModelToBackend(input.Body.Model)

	// Execute prompt
	req := &service.PromptRequest{
		Backend:      backendName,
		Prompt:       prompt,
		Model:        input.Body.Model,
		MaxTokens:    input.Body.MaxTokens,
		SystemPrompt: systemPrompt,
	}

	result, err := h.executor.ExecutePrompt(ctx, req)
	if err != nil {
		return nil, huma.Error500InternalServerError("execution failed", err)
	}

	// Build response
	now := time.Now().Unix()
	responseID := fmt.Sprintf("chatcmpl-%s", result.SessionID)
	if result.SessionID == "" {
		responseID = fmt.Sprintf("chatcmpl-%d", now)
	}

	finishReason := "stop"
	if result.ExitCode != 0 {
		finishReason = "error"
	}

	// Estimate token counts (rough approximation)
	promptTokens := len(prompt) / 4
	completionTokens := len(result.Output) / 4

	return &OpenAIChatCompletionResponse{
		Body: OpenAIChatCompletionResponseBody{
			ID:      responseID,
			Object:  "chat.completion",
			Created: now,
			Model:   input.Body.Model,
			Choices: []OpenAIChatCompletionChoice{
				{
					Index: 0,
					Message: OpenAIMessage{
						Role:    "assistant",
						Content: result.Output,
					},
					FinishReason: finishReason,
				},
			},
			Usage: OpenAIUsage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      promptTokens + completionTokens,
			},
		},
	}, nil
}

// mapModelToBackend maps model names to backend names.
func mapModelToBackend(model string) string {
	// If the model is already a backend name, use it
	switch model {
	case backend.BackendClaude, backend.BackendCodex, backend.BackendGemini:
		return model
	}

	// Map OpenAI-style model names to backends
	switch {
	case strings.Contains(model, "claude"):
		return backend.BackendClaude
	case strings.Contains(model, "gpt"):
		return backend.BackendCodex
	case strings.Contains(model, "gemini"):
		return backend.BackendGemini
	default:
		// Default to claude
		return backend.BackendClaude
	}
}
