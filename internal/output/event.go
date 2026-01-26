// Package output provides unified output format handling across backends.
package output

import (
	"encoding/json"
	"time"
)

// EventType represents the type of output event.
type EventType string

const (
	// EventInit is emitted when a session starts.
	EventInit EventType = "init"

	// EventMessage is emitted for text content from the model.
	EventMessage EventType = "message"

	// EventToolUse is emitted when the model requests a tool execution.
	EventToolUse EventType = "tool_use"

	// EventToolResult is emitted after a tool completes execution.
	EventToolResult EventType = "tool_result"

	// EventThinking is emitted for reasoning/thinking content (extended thinking).
	EventThinking EventType = "thinking"

	// EventError is emitted when an error occurs.
	EventError EventType = "error"

	// EventDone is emitted when the response is complete.
	EventDone EventType = "done"

	// EventProgress is emitted for progress updates.
	EventProgress EventType = "progress"

	// EventTokenUsage is emitted for token consumption updates.
	EventTokenUsage EventType = "token_usage"
)

// UnifiedEvent represents a normalized event from any backend.
type UnifiedEvent struct {
	// Type is the event type.
	Type EventType `json:"type"`

	// Backend identifies which backend produced this event.
	Backend string `json:"backend"`

	// SessionID is the clinvoker session ID.
	SessionID string `json:"session_id,omitempty"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Content contains the event-specific data.
	Content json.RawMessage `json:"content,omitempty"`

	// Sequence is the event sequence number within the session.
	Sequence int `json:"sequence,omitempty"`
}

// InitContent represents initialization event data.
type InitContent struct {
	Model            string `json:"model,omitempty"`
	BackendSessionID string `json:"backend_session_id,omitempty"`
	WorkingDir       string `json:"working_dir,omitempty"`
}

// MessageContent represents a message event data.
type MessageContent struct {
	Text      string `json:"text"`
	Role      string `json:"role,omitempty"` // assistant, user, system
	IsPartial bool   `json:"is_partial,omitempty"`
}

// ToolUseContent represents a tool use request event data.
type ToolUseContent struct {
	ToolID   string          `json:"tool_id,omitempty"`
	ToolName string          `json:"tool_name"`
	Input    json.RawMessage `json:"input,omitempty"`
}

// ToolResultContent represents a tool result event data.
type ToolResultContent struct {
	ToolID   string `json:"tool_id,omitempty"`
	ToolName string `json:"tool_name"`
	Output   string `json:"output,omitempty"`
	IsError  bool   `json:"is_error,omitempty"`
	ErrorMsg string `json:"error_message,omitempty"`
}

// ThinkingContent represents thinking/reasoning event data.
type ThinkingContent struct {
	Text      string `json:"text"`
	IsPartial bool   `json:"is_partial,omitempty"`
}

// ErrorContent represents an error event data.
type ErrorContent struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// DoneContent represents completion event data.
type DoneContent struct {
	TokenUsage *TokenUsageContent `json:"token_usage,omitempty"`
	Duration   time.Duration      `json:"duration,omitempty"`
	TurnCount  int                `json:"turn_count,omitempty"`
}

// ProgressContent represents a progress update event data.
type ProgressContent struct {
	Stage   string  `json:"stage"`
	Percent float64 `json:"percent,omitempty"`
	Message string  `json:"message,omitempty"`
}

// TokenUsageContent represents token consumption data.
type TokenUsageContent struct {
	InputTokens     int64 `json:"input_tokens"`
	OutputTokens    int64 `json:"output_tokens"`
	CachedTokens    int64 `json:"cached_tokens,omitempty"`
	ReasoningTokens int64 `json:"reasoning_tokens,omitempty"`
}

// NewUnifiedEvent creates a new unified event.
func NewUnifiedEvent(eventType EventType, backend, sessionID string) *UnifiedEvent {
	return &UnifiedEvent{
		Type:      eventType,
		Backend:   backend,
		SessionID: sessionID,
		Timestamp: time.Now(),
	}
}

// SetContent sets the event content from a typed struct.
func (e *UnifiedEvent) SetContent(content any) error {
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}
	e.Content = data
	return nil
}

// GetInitContent parses the content as InitContent.
func (e *UnifiedEvent) GetInitContent() (*InitContent, error) {
	if e.Type != EventInit {
		return nil, ErrInvalidEventType
	}
	var content InitContent
	if err := json.Unmarshal(e.Content, &content); err != nil {
		return nil, err
	}
	return &content, nil
}

// GetMessageContent parses the content as MessageContent.
func (e *UnifiedEvent) GetMessageContent() (*MessageContent, error) {
	if e.Type != EventMessage {
		return nil, ErrInvalidEventType
	}
	var content MessageContent
	if err := json.Unmarshal(e.Content, &content); err != nil {
		return nil, err
	}
	return &content, nil
}

// GetToolUseContent parses the content as ToolUseContent.
func (e *UnifiedEvent) GetToolUseContent() (*ToolUseContent, error) {
	if e.Type != EventToolUse {
		return nil, ErrInvalidEventType
	}
	var content ToolUseContent
	if err := json.Unmarshal(e.Content, &content); err != nil {
		return nil, err
	}
	return &content, nil
}

// GetToolResultContent parses the content as ToolResultContent.
func (e *UnifiedEvent) GetToolResultContent() (*ToolResultContent, error) {
	if e.Type != EventToolResult {
		return nil, ErrInvalidEventType
	}
	var content ToolResultContent
	if err := json.Unmarshal(e.Content, &content); err != nil {
		return nil, err
	}
	return &content, nil
}

// GetErrorContent parses the content as ErrorContent.
func (e *UnifiedEvent) GetErrorContent() (*ErrorContent, error) {
	if e.Type != EventError {
		return nil, ErrInvalidEventType
	}
	var content ErrorContent
	if err := json.Unmarshal(e.Content, &content); err != nil {
		return nil, err
	}
	return &content, nil
}

// GetDoneContent parses the content as DoneContent.
func (e *UnifiedEvent) GetDoneContent() (*DoneContent, error) {
	if e.Type != EventDone {
		return nil, ErrInvalidEventType
	}
	var content DoneContent
	if err := json.Unmarshal(e.Content, &content); err != nil {
		return nil, err
	}
	return &content, nil
}

// JSON returns the event as a JSON string.
func (e *UnifiedEvent) JSON() (string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
