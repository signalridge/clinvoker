package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Writer writes unified events to an output stream.
type Writer struct {
	w       io.Writer
	mu      sync.Mutex
	format  OutputFormat
	pretty  bool
	onEvent func(*UnifiedEvent)
}

// OutputFormat specifies the output format.
type OutputFormat string

const (
	// FormatJSON outputs each event as a single JSON line (JSONL/NDJSON).
	FormatJSON OutputFormat = "json"

	// FormatText outputs human-readable text.
	FormatText OutputFormat = "text"

	// FormatQuiet outputs only final results.
	FormatQuiet OutputFormat = "quiet"
)

// WriterOption configures a Writer.
type WriterOption func(*Writer)

// WithFormat sets the output format.
func WithFormat(format OutputFormat) WriterOption {
	return func(w *Writer) {
		w.format = format
	}
}

// WithPrettyPrint enables pretty printing for JSON output.
func WithPrettyPrint(pretty bool) WriterOption {
	return func(w *Writer) {
		w.pretty = pretty
	}
}

// WithEventCallback sets a callback for each event.
func WithEventCallback(fn func(*UnifiedEvent)) WriterOption {
	return func(w *Writer) {
		w.onEvent = fn
	}
}

// NewWriter creates a new event writer.
func NewWriter(w io.Writer, opts ...WriterOption) *Writer {
	writer := &Writer{
		w:      w,
		format: FormatJSON,
	}
	for _, opt := range opts {
		opt(writer)
	}
	return writer
}

// WriteEvent writes a single event to the output.
func (w *Writer) WriteEvent(event *UnifiedEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.onEvent != nil {
		w.onEvent(event)
	}

	switch w.format {
	case FormatJSON:
		return w.writeJSON(event)
	case FormatText:
		return w.writeText(event)
	case FormatQuiet:
		return w.writeQuiet(event)
	default:
		return w.writeJSON(event)
	}
}

// writeJSON writes the event as JSON.
func (w *Writer) writeJSON(event *UnifiedEvent) error {
	var data []byte
	var err error

	if w.pretty {
		data, err = json.MarshalIndent(event, "", "  ")
	} else {
		data, err = json.Marshal(event)
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w.w, string(data))
	return err
}

// writeText writes the event as human-readable text.
func (w *Writer) writeText(event *UnifiedEvent) error {
	switch event.Type {
	case EventMessage:
		content, err := event.GetMessageContent()
		if err != nil {
			return err
		}
		if content.IsPartial {
			_, err = fmt.Fprint(w.w, content.Text)
		} else {
			_, err = fmt.Fprintln(w.w, content.Text)
		}
		return err

	case EventThinking:
		var content ThinkingContent
		if err := json.Unmarshal(event.Content, &content); err != nil {
			return err
		}
		if content.IsPartial {
			_, err := fmt.Fprint(w.w, content.Text)
			return err
		}
		_, err := fmt.Fprintf(w.w, "[thinking] %s\n", content.Text)
		return err

	case EventToolUse:
		content, err := event.GetToolUseContent()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w.w, "[tool: %s]\n", content.ToolName)
		return err

	case EventToolResult:
		content, err := event.GetToolResultContent()
		if err != nil {
			return err
		}
		if content.IsError {
			_, err = fmt.Fprintf(w.w, "[error] %s: %s\n", content.ToolName, content.ErrorMsg)
		}
		return err

	case EventError:
		content, err := event.GetErrorContent()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w.w, "Error: %s\n", content.Message)
		return err

	case EventDone:
		content, err := event.GetDoneContent()
		if err != nil {
			return err
		}
		if content.TokenUsage != nil {
			_, err = fmt.Fprintf(w.w, "\n[tokens: %d in, %d out]\n",
				content.TokenUsage.InputTokens,
				content.TokenUsage.OutputTokens)
		}
		return err

	default:
		// Skip other event types in text mode
		return nil
	}
}

// writeQuiet only outputs final message content.
func (w *Writer) writeQuiet(event *UnifiedEvent) error {
	if event.Type != EventMessage {
		return nil
	}

	content, err := event.GetMessageContent()
	if err != nil {
		return err
	}

	if content.IsPartial {
		return nil
	}

	_, err = fmt.Fprintln(w.w, content.Text)
	return err
}

// Collector collects events for later processing.
type Collector struct {
	events []*UnifiedEvent
	mu     sync.Mutex
}

// NewCollector creates a new event collector.
func NewCollector() *Collector {
	return &Collector{
		events: make([]*UnifiedEvent, 0),
	}
}

// Collect adds an event to the collection.
func (c *Collector) Collect(event *UnifiedEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event)
}

// Events returns all collected events.
func (c *Collector) Events() []*UnifiedEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]*UnifiedEvent, len(c.events))
	copy(result, c.events)
	return result
}

// FilterByType returns events of a specific type.
func (c *Collector) FilterByType(eventType EventType) []*UnifiedEvent {
	c.mu.Lock()
	defer c.mu.Unlock()

	var filtered []*UnifiedEvent
	for _, e := range c.events {
		if e.Type == eventType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// Messages returns all message contents concatenated.
func (c *Collector) Messages() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var builder strings.Builder
	for _, e := range c.events {
		if e.Type == EventMessage {
			content, err := e.GetMessageContent()
			if err == nil {
				builder.WriteString(content.Text)
			}
		}
	}
	return builder.String()
}

// LastError returns the last error event, if any.
func (c *Collector) LastError() *ErrorContent {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := len(c.events) - 1; i >= 0; i-- {
		if c.events[i].Type == EventError {
			content, err := c.events[i].GetErrorContent()
			if err == nil {
				return content
			}
		}
	}
	return nil
}

// TotalTokens returns the total tokens from DoneContent events.
func (c *Collector) TotalTokens() (input, output int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, e := range c.events {
		if e.Type == EventDone {
			content, err := e.GetDoneContent()
			if err == nil && content.TokenUsage != nil {
				input += content.TokenUsage.InputTokens
				output += content.TokenUsage.OutputTokens
			}
		}
		if e.Type == EventTokenUsage {
			var content TokenUsageContent
			if err := json.Unmarshal(e.Content, &content); err == nil {
				input += content.InputTokens
				output += content.OutputTokens
			}
		}
	}
	return
}

// Clear removes all collected events.
func (c *Collector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = c.events[:0]
}
