package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/server/handlers"
	"github.com/signalridge/clinvoker/internal/server/service"
)

func TestNewStdioTransport(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)

	reader := &bytes.Buffer{}
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)
	if transport == nil {
		t.Fatal("NewStdioTransport returned nil")
	}
	if transport.dispatcher != dispatcher {
		t.Error("dispatcher not set correctly")
	}
}

func TestStdioTransport_SendNotification(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)

	reader := &bytes.Buffer{}
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	notification := &Notification{
		JSONRPC: "2.0",
		Method:  "notifications/message",
		Params:  map[string]string{"level": "info", "message": "test"},
	}

	err := transport.SendNotification(notification)
	if err != nil {
		t.Fatalf("SendNotification error: %v", err)
	}

	// Verify output
	output := writer.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}

	// Should be a single line (JSON-RPC message)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}

	// Verify it's valid JSON
	var notif Notification
	if err := json.Unmarshal([]byte(lines[0]), &notif); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
	if notif.Method != "notifications/message" {
		t.Errorf("method = %q, want %q", notif.Method, "notifications/message")
	}
}

func TestStdioTransport_Run_SingleRequest(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	// Prepare input with a single request
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := transport.Run(ctx)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		// EOF from scanner is expected when input is exhausted
		if err.Error() != "EOF" {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// Verify output
	output := writer.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}

	// Should have a response
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 1 {
		t.Fatal("expected at least one line of output")
	}

	// Verify it's a valid JSON-RPC response
	var resp Response
	if err := json.Unmarshal([]byte(lines[0]), &resp); err != nil {
		t.Fatalf("output is not valid JSON-RPC: %v", err)
	}
	if resp.JSONRPC != "2.0" {
		t.Errorf("jsonrpc = %q, want %q", resp.JSONRPC, "2.0")
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestStdioTransport_Run_InvalidJSON(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	// Prepare input with invalid JSON
	input := `invalid json` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := transport.Run(ctx)
	// EOF or timeout is expected
	if err != nil && err.Error() != "EOF" && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify error response was written
	output := writer.String()
	if output == "" {
		t.Fatal("expected error response, got empty string")
	}

	var resp Response
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &resp); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error in response")
	}
	if resp.Error.Code != CodeParseError {
		t.Errorf("error code = %d, want %d", resp.Error.Code, CodeParseError)
	}
}

func TestStdioTransport_Run_MultipleRequests(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	// Prepare input with multiple requests
	input := `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := transport.Run(ctx)
	if err != nil && err.Error() != "EOF" && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify outputs
	output := writer.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}

	// Should have two responses
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	// Both should be valid JSON-RPC responses
	for i, line := range lines {
		var resp Response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Errorf("line %d is not valid JSON-RPC: %v", i+1, err)
		}
	}
}

func TestStdioTransport_ContextCancellation(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	// Empty input - will block on scanner
	reader := &bytes.Buffer{}
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := transport.Run(ctx)
	if err == nil {
		// Context cancellation may or may not return an error
		t.Log("Run completed without error (context canceled)")
	}
}

func TestStdioTransport_ContextCancelClosesReader(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	reader := newBlockingReader()
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- transport.Run(ctx)
	}()

	select {
	case <-reader.readStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("reader did not start")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestStdioTransport_Run_WriteResponseError(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	input := `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"
	reader := strings.NewReader(input)
	writer := &errorWriter{err: errors.New("write failed")}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := transport.Run(ctx)
	if err == nil {
		t.Fatal("expected write error")
	}
}

func TestStdioTransport_ImplementsNotificationSender(t *testing.T) {
	// Verify StdioTransport implements NotificationSender interface
	var _ NotificationSender = (*StdioTransport)(nil)
}

func TestWithNotificationSender_Stdio(t *testing.T) {
	registry := NewRegistry()
	logger := slog.Default()
	dispatcher := NewDispatcher(registry, logger)

	reader := &bytes.Buffer{}
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	// Verify transport can be used as notification sender
	ctx := WithNotificationSender(context.Background(), transport)
	sender := GetNotificationSender(ctx)

	if sender != transport {
		t.Error("GetNotificationSender did not return the same transport")
	}
}

func TestStdioTransport_Run_ToolsList(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	RegisterAllTools(registry, service.NewExecutor())
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := transport.Run(ctx)
	if err != nil && err.Error() != "EOF" && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(writer.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var resp Response
	if err := json.Unmarshal([]byte(lines[0]), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result ToolsListResult
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("unmarshal tools list: %v", err)
	}
	if len(result.Tools) == 0 {
		t.Fatal("expected tools in response")
	}
	for _, tool := range result.Tools {
		if len(tool.InputSchema) == 0 {
			t.Errorf("tool %q missing input schema", tool.Name)
		}
		if len(tool.OutputSchema) == 0 {
			t.Errorf("tool %q missing output schema", tool.Name)
		}
	}
}

func TestStdioTransport_Run_ToolsCall(t *testing.T) {
	setTempHome(t)
	registry := NewRegistry()
	RegisterAllTools(registry, service.NewExecutor())
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	input := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"clinvk_backends","arguments":{}}}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := transport.Run(ctx)
	if err != nil && err.Error() != "EOF" && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(writer.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var resp Response
	if err := json.Unmarshal([]byte(lines[0]), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result ToolCallResult
	resultJSON, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("unmarshal tool result: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.BackendsResponseBody
	if err := json.Unmarshal([]byte(result.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if body.Backends == nil {
		t.Fatal("expected backends to be non-nil")
	}
}

func TestStdioTransport_Run_NotificationNoResponse(t *testing.T) {
	registry := NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dispatcher := NewDispatcher(registry, logger)

	input := `{"jsonrpc":"2.0","method":"ping"}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	transport := NewStdioTransport(dispatcher, logger, reader, writer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := transport.Run(ctx)
	if err != nil && err.Error() != "EOF" && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %v", err)
	}

	if strings.TrimSpace(writer.String()) != "" {
		t.Fatal("expected no response for notification")
	}
}

type errorWriter struct {
	err error
}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, e.err
}

type blockingReader struct {
	readStarted chan struct{}
	closed      chan struct{}
	once        sync.Once
}

func newBlockingReader() *blockingReader {
	return &blockingReader{
		readStarted: make(chan struct{}),
		closed:      make(chan struct{}),
	}
}

func (b *blockingReader) Read(p []byte) (int, error) {
	b.once.Do(func() {
		close(b.readStarted)
	})
	<-b.closed
	return 0, io.EOF
}

func (b *blockingReader) Close() error {
	close(b.closed)
	return nil
}
