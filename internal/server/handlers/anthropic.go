package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/output"
	"github.com/signalridge/clinvoker/internal/server/service"
)

// Message role constants.
const (
	roleUser      = "user"
	roleAssistant = "assistant"
)

// AnthropicHandlers provides handlers for Anthropic-compatible API.
type AnthropicHandlers struct {
	runner service.PromptRunner
}

// NewAnthropicHandlers creates a new Anthropic handlers instance.
func NewAnthropicHandlers(runner service.PromptRunner) *AnthropicHandlers {
	return &AnthropicHandlers{runner: runner}
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
		Responses: map[string]*huma.Response{
			"200": {
				Description: "OK",
				Content: map[string]*huma.MediaType{
					"application/json": {},
					"text/event-stream": {
						Schema: &huma.Schema{
							Type:        huma.TypeString,
							Description: "Server-sent events stream of message deltas (when stream=true).",
						},
					},
				},
			},
		},
		Tags: []string{"Anthropic Compatible"},
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
	Stream        bool               `json:"stream,omitempty" doc:"Stream responses"`
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

type anthropicStreamMessageStart struct {
	Type    string                 `json:"type"`
	Message anthropicStreamMessage `json:"message"`
}

type anthropicStreamMessage struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []AnthropicContentBlock `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason,omitempty"`
	StopSequence string                  `json:"stop_sequence,omitempty"`
	Usage        AnthropicUsage          `json:"usage"`
}

type anthropicStreamContentBlockStart struct {
	Type         string                     `json:"type"`
	Index        int                        `json:"index"`
	ContentBlock anthropicStreamContentText `json:"content_block"`
}

type anthropicStreamContentText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicStreamContentBlockDelta struct {
	Type  string                   `json:"type"`
	Index int                      `json:"index"`
	Delta anthropicStreamTextDelta `json:"delta"`
}

type anthropicStreamTextDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicStreamContentBlockStop struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

type anthropicStreamMessageDelta struct {
	Type  string                          `json:"type"`
	Delta anthropicStreamMessageDeltaData `json:"delta"`
	Usage AnthropicUsage                  `json:"usage,omitempty"`
}

type anthropicStreamMessageDeltaData struct {
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

type anthropicStreamMessageStop struct {
	Type string `json:"type"`
}

type anthropicStreamError struct {
	Type  string                      `json:"type"`
	Error anthropicStreamErrorDetails `json:"error"`
}

type anthropicStreamErrorDetails struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// AnthropicMessagesInput is the input for the messages handler.
type AnthropicMessagesInput struct {
	Body AnthropicMessagesRequest
}

// HandleMessages handles the POST /v1/messages endpoint.
func (h *AnthropicHandlers) HandleMessages(ctx context.Context, input *AnthropicMessagesInput) (*huma.StreamResponse, error) {
	if input.Body.Model == "" {
		return nil, huma.Error400BadRequest("model is required")
	}
	if len(input.Body.Messages) == 0 {
		return nil, huma.Error400BadRequest("messages are required")
	}
	if input.Body.MaxTokens <= 0 {
		return nil, huma.Error400BadRequest("max_tokens must be greater than 0")
	}

	// Extract prompt from messages
	var prompt string
	for _, msg := range input.Body.Messages {
		if msg.Role == roleUser {
			if prompt != "" {
				prompt += "\n"
			}
			prompt += msg.Content
		} else if msg.Role == roleAssistant {
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

	// Execute prompt (non-streaming)
	req := &service.PromptRequest{
		Backend:      backendName,
		Prompt:       prompt,
		Model:        input.Body.Model,
		MaxTokens:    input.Body.MaxTokens,
		SystemPrompt: input.Body.System,
		Metadata:     input.Body.Metadata,
	}

	if !input.Body.Stream {
		result, err := h.runner.ExecutePrompt(ctx, req)
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

		// Token counts (use backend usage if available, fallback to rough estimate)
		inputTokens := len(prompt) / 4
		outputTokens := len(result.Output) / 4
		if result.TokenUsage != nil {
			inputTokens = int(result.TokenUsage.InputTokens)
			outputTokens = int(result.TokenUsage.OutputTokens)
		}

		body := AnthropicMessagesResponseBody{
			ID:   responseID,
			Type: "message",
			Role: roleAssistant,
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
		}

		return &huma.StreamResponse{
			Body: func(hctx huma.Context) {
				hctx.SetStatus(http.StatusOK)
				hctx.SetHeader("Content-Type", "application/json")
				_ = json.NewEncoder(hctx.BodyWriter()).Encode(body)
			},
		}, nil
	}

	created := time.Now().Unix()
	responseID := fmt.Sprintf("msg_%d", created)

	return &huma.StreamResponse{
		Body: func(hctx huma.Context) {
			setEventStreamHeaders(hctx)

			_ = writeSSEEvent(hctx, "message_start", anthropicStreamMessageStart{
				Type: "message_start",
				Message: anthropicStreamMessage{
					ID:      responseID,
					Type:    "message",
					Role:    roleAssistant,
					Content: []AnthropicContentBlock{},
					Model:   input.Body.Model,
					Usage: AnthropicUsage{
						InputTokens:  0,
						OutputTokens: 0,
					},
				},
			})

			_ = writeSSEEvent(hctx, "content_block_start", anthropicStreamContentBlockStart{
				Type:  "content_block_start",
				Index: 0,
				ContentBlock: anthropicStreamContentText{
					Type: "text",
					Text: "",
				},
			})

			streamReq := *req
			streamCtx := hctx.Context()

			streamResult, streamErr := service.StreamPrompt(streamCtx, &streamReq, nil, true, func(event *output.UnifiedEvent) error {
				if event.Type != output.EventMessage {
					return nil
				}
				content, err := event.GetMessageContent()
				if err != nil {
					return err
				}
				if content.Text == "" {
					return nil
				}
				return writeSSEEvent(hctx, "content_block_delta", anthropicStreamContentBlockDelta{
					Type:  "content_block_delta",
					Index: 0,
					Delta: anthropicStreamTextDelta{
						Type: "text_delta",
						Text: content.Text,
					},
				})
			})

			_ = writeSSEEvent(hctx, "content_block_stop", anthropicStreamContentBlockStop{
				Type:  "content_block_stop",
				Index: 0,
			})

			stopReason := "end_turn"
			if streamErr != nil || streamResult == nil || streamResult.ExitCode != 0 || streamResult.Error != "" {
				stopReason = "error"
			}

			usage := AnthropicUsage{}
			if streamResult != nil && streamResult.TokenUsage != nil {
				usage.InputTokens = int(streamResult.TokenUsage.InputTokens)
				usage.OutputTokens = int(streamResult.TokenUsage.OutputTokens)
			}

			_ = writeSSEEvent(hctx, "message_delta", anthropicStreamMessageDelta{
				Type: "message_delta",
				Delta: anthropicStreamMessageDeltaData{
					StopReason: stopReason,
				},
				Usage: usage,
			})

			if streamErr != nil {
				_ = writeSSEEvent(hctx, "error", anthropicStreamError{
					Type: "error",
					Error: anthropicStreamErrorDetails{
						Type:    "stream_error",
						Message: streamErr.Error(),
					},
				})
			}

			_ = writeSSEEvent(hctx, "message_stop", anthropicStreamMessageStop{
				Type: "message_stop",
			})
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
