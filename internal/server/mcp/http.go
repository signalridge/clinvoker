package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

// HTTPTransport implements MCP over HTTP POST with optional SSE notifications.
type HTTPTransport struct {
	dispatcher *Dispatcher
	logger     *slog.Logger
}

// NewHTTPTransport creates a new HTTP transport.
func NewHTTPTransport(dispatcher *Dispatcher, logger *slog.Logger) *HTTPTransport {
	return &HTTPTransport{
		dispatcher: dispatcher,
		logger:     logger,
	}
}

// Handler returns an http.Handler that processes MCP JSON-RPC requests.
func (t *HTTPTransport) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.logger.Error("failed to read request body", "error", err)
			writeJSONError(w, CodeParseError, "failed to read request body")
			return
		}
		defer func() {
			if closeErr := r.Body.Close(); closeErr != nil {
				t.logger.Warn("failed to close request body", "error", closeErr)
			}
		}()

		var req Request
		if err := json.Unmarshal(body, &req); err != nil {
			t.logger.Warn("failed to parse JSON-RPC request", "error", err)
			writeJSONError(w, CodeParseError, "parse error")
			return
		}

		if req.JSONRPC != jsonRPCVersion {
			writeJSONError(w, CodeInvalidRequest, "invalid JSON-RPC version")
			return
		}

		accept := r.Header.Get("Accept")
		useSSE := acceptsEventStream(accept)
		streamRequested := isStreamRequest(&req)

		ctx := r.Context()

		if streamRequested && !useSSE {
			writeJSONRPCError(w, req.ID, CodeInvalidParams, "streaming requires Accept: text/event-stream")
			return
		}

		if useSSE && streamRequested {
			t.handleSSE(ctx, w, &req)
		} else {
			t.handleJSON(ctx, w, &req)
		}
	})
}

// handleJSON processes a request and returns a single JSON response.
func (t *HTTPTransport) handleJSON(ctx context.Context, w http.ResponseWriter, req *Request) {
	resp := t.dispatcher.Dispatch(ctx, req)
	if resp == nil {
		// Notification — no response needed.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.logger.Error("failed to write response", "error", err)
	}
}

// handleSSE processes a request with SSE notifications during execution.
func (t *HTTPTransport) handleSSE(ctx context.Context, w http.ResponseWriter, req *Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		t.logger.Warn("response writer does not support flushing, falling back to JSON")
		t.handleJSON(ctx, w, req)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Create an SSE notification sender.
	sender := &sseNotificationSender{writer: w, flusher: flusher, logger: t.logger}
	ctx = WithNotificationSender(ctx, sender)

	resp := t.dispatcher.Dispatch(ctx, req)
	if resp == nil {
		return
	}

	// Send the final result as an SSE event.
	data, err := json.Marshal(resp)
	if err != nil {
		t.logger.Error("failed to marshal final response", "error", err)
		return
	}

	sender.mu.Lock()
	defer sender.mu.Unlock()

	if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", data); err != nil {
		t.logger.Error("failed to write SSE completion", "error", err)
		return
	}
	flusher.Flush()
}

// sseNotificationSender sends MCP notifications as SSE events.
type sseNotificationSender struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	logger  *slog.Logger
	mu      sync.Mutex
}

func (s *sseNotificationSender) SendNotification(notification *Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = fmt.Fprintf(s.writer, "event: notification\ndata: %s\n\n", data)
	if err != nil {
		return fmt.Errorf("write SSE notification: %w", err)
	}
	s.flusher.Flush()
	return nil
}

// writeJSONError writes a JSON-RPC error response with HTTP 400.
func writeJSONError(w http.ResponseWriter, code int, message string) {
	resp := NewErrorResponse(nil, code, message, nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}

func writeJSONRPCError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	resp := NewErrorResponse(id, code, message, nil)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}

func acceptsEventStream(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		media := strings.TrimSpace(part)
		if media == "" {
			continue
		}
		if strings.EqualFold(strings.SplitN(media, ";", 2)[0], "text/event-stream") {
			return true
		}
	}
	return false
}

func isStreamRequest(req *Request) bool {
	if req == nil || req.Method != MethodToolsCall || len(req.Params) == 0 {
		return false
	}

	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return false
	}
	if params.Name != "clinvk_prompt" || len(params.Arguments) == 0 {
		return false
	}

	var args struct {
		OutputFormat string `json:"output_format"`
	}
	if err := json.Unmarshal(params.Arguments, &args); err != nil {
		return false
	}

	return strings.EqualFold(strings.TrimSpace(args.OutputFormat), "stream-json")
}
