package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// StdioTransport implements MCP over stdin/stdout using line-delimited JSON-RPC.
type StdioTransport struct {
	dispatcher *Dispatcher
	logger     *slog.Logger
	reader     io.Reader
	writer     io.Writer
	writeMu    sync.Mutex
}

// NewStdioTransport creates a new stdio transport.
func NewStdioTransport(dispatcher *Dispatcher, logger *slog.Logger, reader io.Reader, writer io.Writer) *StdioTransport {
	return &StdioTransport{
		dispatcher: dispatcher,
		logger:     logger,
		reader:     reader,
		writer:     writer,
	}
}

// SendNotification implements NotificationSender for stdio transport.
func (t *StdioTransport) SendNotification(notification *Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	data = append(data, '\n')
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("write notification: %w", err)
	}
	return nil
}

// Run reads JSON-RPC requests from the reader and writes responses to the writer.
// It blocks until the context is canceled or the reader reaches EOF.
func (t *StdioTransport) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(t.reader)
	// MCP messages can be large; allow up to 10MB per line.
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	// Inject this transport as the notification sender for streaming support.
	ctx = WithNotificationSender(ctx, t)

	done := make(chan struct{})
	defer close(done)

	if rc, ok := t.reader.(io.ReadCloser); ok {
		go func() {
			select {
			case <-ctx.Done():
				if err := rc.Close(); err != nil {
					t.logger.Debug("failed to close stdio reader", "error", err)
				}
			case <-done:
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			t.logger.Info("stdio transport shutting down (context canceled)")
			return ctx.Err()
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				if ctx.Err() != nil {
					t.logger.Info("stdio transport shutting down (context canceled)", "error", err)
					return ctx.Err()
				}
				t.logger.Error("stdin read error", "error", err)
				return fmt.Errorf("stdin read error: %w", err)
			}
			if ctx.Err() != nil {
				t.logger.Info("stdio transport shutting down (context canceled)")
				return ctx.Err()
			}
			// EOF — clean shutdown.
			t.logger.Info("stdin EOF, shutting down")
			return nil
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			t.logger.Warn("failed to parse JSON-RPC request", "error", err)
			resp := NewErrorResponse(nil, CodeParseError, "parse error", nil)
			if err := t.writeResponse(resp); err != nil {
				return err
			}
			continue
		}

		if req.JSONRPC != jsonRPCVersion {
			resp := NewErrorResponse(req.ID, CodeInvalidRequest, "invalid JSON-RPC version", nil)
			if err := t.writeResponse(resp); err != nil {
				return err
			}
			continue
		}

		resp := t.dispatcher.Dispatch(ctx, &req)
		if resp != nil {
			if err := t.writeResponse(resp); err != nil {
				return err
			}
		}
	}
}

// writeResponse serializes and writes a JSON-RPC response as a single line.
func (t *StdioTransport) writeResponse(resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		t.logger.Error("failed to marshal response", "error", err)
		return err
	}

	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	// Write as a single line followed by newline.
	data = append(data, '\n')
	if _, err := t.writer.Write(data); err != nil {
		t.logger.Error("failed to write response", "error", err)
		return err
	}
	return nil
}
