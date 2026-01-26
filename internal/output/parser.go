package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Parser converts backend-specific output to unified events.
type Parser struct {
	backend   string
	sessionID string
	sequence  int
}

// NewParser creates a new parser for the specified backend.
func NewParser(backend, sessionID string) *Parser {
	return &Parser{
		backend:   backend,
		sessionID: sessionID,
		sequence:  0,
	}
}

// ParseLine parses a single line of output from a backend.
func (p *Parser) ParseLine(line string) (*UnifiedEvent, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	p.sequence++

	switch p.backend {
	case "claude":
		return p.parseClaudeLine(line)
	case "gemini":
		return p.parseGeminiLine(line)
	case "codex":
		return p.parseCodexLine(line)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownBackend, p.backend)
	}
}

// ParseStream reads and parses a stream of output.
func (p *Parser) ParseStream(r io.Reader, eventCh chan<- *UnifiedEvent, errCh chan<- error) {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for large outputs
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		event, err := p.ParseLine(line)
		if err != nil {
			errCh <- err
			continue
		}
		if event != nil {
			eventCh <- event
		}
	}

	if err := scanner.Err(); err != nil {
		errCh <- err
	}

	close(eventCh)
	close(errCh)
}

// parseClaudeLine parses a line from Claude Code stream-json output.
// Claude Code format: {"type": "assistant", "message": {...}} or similar
func (p *Parser) parseClaudeLine(line string) (*UnifiedEvent, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		// Not JSON, treat as plain text message
		return p.createMessageEvent(line, false)
	}

	eventType, _ := raw["type"].(string)

	switch eventType {
	case "system":
		return p.createEvent(EventInit, &InitContent{
			BackendSessionID: getString(raw, "session_id"),
		})

	case "assistant":
		msg, _ := raw["message"].(map[string]any)
		if msg != nil {
			content, _ := msg["content"].([]any)
			if len(content) > 0 {
				for _, c := range content {
					cm, ok := c.(map[string]any)
					if !ok {
						continue
					}
					ctype, _ := cm["type"].(string)
					switch ctype {
					case "text":
						text, _ := cm["text"].(string)
						return p.createMessageEvent(text, false)
					case "tool_use":
						return p.createToolUseEvent(
							getString(cm, "id"),
							getString(cm, "name"),
							cm["input"],
						)
					}
				}
			}
		}
		return nil, nil

	case "content_block_delta":
		delta, _ := raw["delta"].(map[string]any)
		if delta != nil {
			deltaType, _ := delta["type"].(string)
			switch deltaType {
			case "text_delta":
				text, _ := delta["text"].(string)
				return p.createMessageEvent(text, true)
			case "thinking_delta":
				text, _ := delta["thinking"].(string)
				return p.createThinkingEvent(text, true)
			}
		}
		return nil, nil

	case "tool_result":
		return p.createToolResultEvent(
			getString(raw, "tool_use_id"),
			"",
			getString(raw, "content"),
			false,
		)

	case "error":
		errData, _ := raw["error"].(map[string]any)
		msg := getString(errData, "message")
		if msg == "" {
			msg = getString(raw, "message")
		}
		return p.createErrorEvent("", msg, "")

	case "message_stop", "result":
		usage, _ := raw["usage"].(map[string]any)
		return p.createDoneEvent(
			getInt64(usage, "input_tokens"),
			getInt64(usage, "output_tokens"),
		)

	default:
		// Unknown type, skip
		return nil, nil
	}
}

// parseGeminiLine parses a line from Gemini CLI stream-json/NDJSON output.
// Gemini format: {"type": "init|message|tool_use|tool_result|error|result", ...}
func (p *Parser) parseGeminiLine(line string) (*UnifiedEvent, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return p.createMessageEvent(line, false)
	}

	eventType, _ := raw["type"].(string)

	switch eventType {
	case "init":
		return p.createEvent(EventInit, &InitContent{
			BackendSessionID: getString(raw, "sessionId"),
		})

	case "message":
		content, _ := raw["content"].(string)
		return p.createMessageEvent(content, true)

	case "tool_use":
		return p.createToolUseEvent(
			getString(raw, "toolCallId"),
			getString(raw, "toolName"),
			raw["parameters"],
		)

	case "tool_result":
		result, _ := raw["result"].(string)
		return p.createToolResultEvent(
			getString(raw, "toolCallId"),
			getString(raw, "toolName"),
			result,
			false,
		)

	case "error":
		return p.createErrorEvent(
			getString(raw, "code"),
			getString(raw, "message"),
			getString(raw, "details"),
		)

	case "result":
		stats, _ := raw["stats"].(map[string]any)
		tokenUsage, _ := stats["tokenUsage"].(map[string]any)
		return p.createDoneEvent(
			getInt64(tokenUsage, "inputTokens"),
			getInt64(tokenUsage, "outputTokens"),
		)

	default:
		return nil, nil
	}
}

// parseCodexLine parses a line from Codex CLI JSON Lines output.
// Codex format: {"type": "thread.started|turn.started|item.completed|...", ...}
func (p *Parser) parseCodexLine(line string) (*UnifiedEvent, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return p.createMessageEvent(line, false)
	}

	eventType, _ := raw["type"].(string)

	switch eventType {
	case "thread.started":
		return p.createEvent(EventInit, &InitContent{
			BackendSessionID: getString(raw, "thread_id"),
		})

	case "turn.started":
		return p.createEvent(EventProgress, &ProgressContent{
			Stage:   "turn_started",
			Message: "Processing...",
		})

	case "item.completed":
		item, _ := raw["item"].(map[string]any)
		if item != nil {
			itemType, _ := item["type"].(string)
			switch itemType {
			case "message":
				content, _ := item["content"].([]any)
				for _, c := range content {
					cm, ok := c.(map[string]any)
					if !ok {
						continue
					}
					ctype, _ := cm["type"].(string)
					if ctype == "output_text" {
						text, _ := cm["text"].(string)
						return p.createMessageEvent(text, false)
					}
				}
			case "function_call":
				return p.createToolUseEvent(
					getString(item, "call_id"),
					getString(item, "name"),
					item["arguments"],
				)
			case "function_call_output":
				return p.createToolResultEvent(
					getString(item, "call_id"),
					"",
					getString(item, "output"),
					false,
				)
			}
		}
		return nil, nil

	case "turn.completed":
		usage, _ := raw["usage"].(map[string]any)
		return p.createDoneEvent(
			getInt64(usage, "input_tokens"),
			getInt64(usage, "output_tokens"),
		)

	case "error":
		return p.createErrorEvent(
			getString(raw, "code"),
			getString(raw, "message"),
			"",
		)

	case "reasoning":
		text, _ := raw["text"].(string)
		return p.createThinkingEvent(text, false)

	case "event_msg":
		payload, _ := raw["payload"].(map[string]any)
		if payload != nil {
			payloadType, _ := payload["type"].(string)
			if payloadType == "token_count" {
				return p.createEvent(EventTokenUsage, &TokenUsageContent{
					InputTokens:  getInt64(payload, "input_tokens"),
					OutputTokens: getInt64(payload, "output_tokens"),
				})
			}
		}
		return nil, nil

	default:
		return nil, nil
	}
}

// Helper methods for creating events

func (p *Parser) createEvent(eventType EventType, content any) (*UnifiedEvent, error) {
	event := &UnifiedEvent{
		Type:      eventType,
		Backend:   p.backend,
		SessionID: p.sessionID,
		Timestamp: time.Now(),
		Sequence:  p.sequence,
	}
	if content != nil {
		if err := event.SetContent(content); err != nil {
			return nil, err
		}
	}
	return event, nil
}

func (p *Parser) createMessageEvent(text string, isPartial bool) (*UnifiedEvent, error) {
	return p.createEvent(EventMessage, &MessageContent{
		Text:      text,
		Role:      "assistant",
		IsPartial: isPartial,
	})
}

func (p *Parser) createThinkingEvent(text string, isPartial bool) (*UnifiedEvent, error) {
	return p.createEvent(EventThinking, &ThinkingContent{
		Text:      text,
		IsPartial: isPartial,
	})
}

func (p *Parser) createToolUseEvent(toolID, toolName string, input any) (*UnifiedEvent, error) {
	inputJSON, _ := json.Marshal(input)
	return p.createEvent(EventToolUse, &ToolUseContent{
		ToolID:   toolID,
		ToolName: toolName,
		Input:    inputJSON,
	})
}

func (p *Parser) createToolResultEvent(toolID, toolName, output string, isError bool) (*UnifiedEvent, error) {
	return p.createEvent(EventToolResult, &ToolResultContent{
		ToolID:   toolID,
		ToolName: toolName,
		Output:   output,
		IsError:  isError,
	})
}

func (p *Parser) createErrorEvent(code, message, details string) (*UnifiedEvent, error) {
	return p.createEvent(EventError, &ErrorContent{
		Code:    code,
		Message: message,
		Details: details,
	})
}

func (p *Parser) createDoneEvent(inputTokens, outputTokens int64) (*UnifiedEvent, error) {
	return p.createEvent(EventDone, &DoneContent{
		TokenUsage: &TokenUsageContent{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
		},
	})
}

// Helper functions for extracting values from maps

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt64(m map[string]any, key string) int64 {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}
