package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/signalridge/clinvoker/internal/output"
)

// notifySenderKey is the context key for NotificationSender.
type notifySenderKey struct{}

// NotificationSender is an interface for sending JSON-RPC notifications during tool execution.
type NotificationSender interface {
	SendNotification(notification *Notification) error
}

// WithNotificationSender returns a new context with the given NotificationSender.
func WithNotificationSender(ctx context.Context, sender NotificationSender) context.Context {
	return context.WithValue(ctx, notifySenderKey{}, sender)
}

// GetNotificationSender returns the NotificationSender from the context, or nil.
func GetNotificationSender(ctx context.Context) NotificationSender {
	s, _ := ctx.Value(notifySenderKey{}).(NotificationSender)
	return s
}

// MCP notification method names for streaming.
const (
	NotificationProgress = "notifications/progress"
	NotificationMessage  = "notifications/message"
)

// ProgressNotificationParams represents the params for a progress notification.
type ProgressNotificationParams struct {
	ProgressToken any     `json:"progressToken"`
	Progress      float64 `json:"progress"`
	Total         float64 `json:"total,omitempty"`
	Message       string  `json:"message,omitempty"`
}

// LogMessageParams represents a log/message notification.
type LogMessageParams struct {
	Level  string `json:"level"`
	Logger string `json:"logger,omitempty"`
	Data   any    `json:"data"`
}

// TranslateEvent converts a UnifiedEvent to an MCP notification.
// Returns nil if the event should not be forwarded as a notification.
func TranslateEvent(event *output.UnifiedEvent, progressToken any) *Notification {
	if event == nil {
		return nil
	}

	switch event.Type {
	case output.EventInit:
		content, err := event.GetInitContent()
		if err != nil {
			return nil
		}
		return NewNotification(NotificationMessage, LogMessageParams{
			Level:  "info",
			Logger: "clinvk",
			Data: map[string]any{
				"type":    "init",
				"backend": event.Backend,
				"model":   content.Model,
			},
		})

	case output.EventMessage:
		content, err := event.GetMessageContent()
		if err != nil {
			return nil
		}
		return NewNotification(NotificationMessage, LogMessageParams{
			Level:  "info",
			Logger: "clinvk",
			Data: map[string]any{
				"type": "message",
				"text": content.Text,
				"role": content.Role,
			},
		})

	case output.EventToolUse:
		content, err := event.GetToolUseContent()
		if err != nil {
			return nil
		}
		return NewNotification(NotificationProgress, ProgressNotificationParams{
			ProgressToken: progressToken,
			Message:       fmt.Sprintf("Using tool: %s", content.ToolName),
		})

	case output.EventToolResult:
		content, err := event.GetToolResultContent()
		if err != nil {
			return nil
		}
		msg := fmt.Sprintf("Tool result: %s", content.ToolName)
		if content.IsError {
			msg = fmt.Sprintf("Tool error: %s - %s", content.ToolName, content.ErrorMsg)
		}
		return NewNotification(NotificationProgress, ProgressNotificationParams{
			ProgressToken: progressToken,
			Message:       msg,
		})

	case output.EventThinking:
		// Thinking events are optional; emit as log message.
		content := struct {
			Text string `json:"text"`
		}{}
		if err := json.Unmarshal(event.Content, &content); err != nil {
			return nil
		}
		// Truncate long thinking text for notifications.
		text := content.Text
		if len(text) > 200 {
			text = text[:200] + "..."
		}
		return NewNotification(NotificationMessage, LogMessageParams{
			Level:  "debug",
			Logger: "clinvk",
			Data: map[string]any{
				"type": "thinking",
				"text": text,
			},
		})

	case output.EventError:
		content, err := event.GetErrorContent()
		if err != nil {
			return nil
		}
		return NewNotification(NotificationMessage, LogMessageParams{
			Level:  "error",
			Logger: "clinvk",
			Data: map[string]any{
				"type":    "error",
				"code":    content.Code,
				"message": content.Message,
			},
		})

	case output.EventProgress:
		var content output.ProgressContent
		if err := json.Unmarshal(event.Content, &content); err != nil {
			return nil
		}
		return NewNotification(NotificationProgress, ProgressNotificationParams{
			ProgressToken: progressToken,
			Progress:      content.Percent,
			Total:         100,
			Message:       content.Message,
		})

	case output.EventDone:
		// Done event signals completion; no notification needed (final result follows).
		return nil

	case output.EventTokenUsage:
		// Token usage is included in the final result, not as a notification.
		return nil

	default:
		return nil
	}
}

// StreamingToolResult holds the accumulated output from a streaming tool call.
type StreamingToolResult struct {
	Output strings.Builder
	Error  string
}
