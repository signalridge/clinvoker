package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/server/handlers"
	"github.com/signalridge/clinvoker/internal/server/service"
)

type countingSender struct {
	count int
}

func (c *countingSender) SendNotification(notification *Notification) error {
	c.count++
	return nil
}

func TestPromptTool_NonStreaming_InvalidBackendReturnsRPCError(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := promptTool(executor)

	args, err := json.Marshal(handlers.PromptRequest{
		Backend: "invalid-backend",
		Prompt:  "hello",
	})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	result, err := tool.Handler(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for invalid backend")
	}
	if result != nil {
		t.Fatal("expected nil result on error")
	}
	var rpcErr *RPCErrorDetail
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected RPCErrorDetail, got %T", err)
	}
	if rpcErr.Code != CodeInvalidParams {
		t.Fatalf("code = %d, want %d", rpcErr.Code, CodeInvalidParams)
	}
}

func TestPromptTool_StreamRequiresNotificationSender(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := promptTool(executor)

	args, err := json.Marshal(handlers.PromptRequest{
		Backend:      "invalid-backend",
		Prompt:       "hello",
		OutputFormat: "stream-json",
	})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	_, err = tool.Handler(context.Background(), args)
	if err == nil {
		t.Fatal("expected error when streaming without notification sender")
	}

	var rpcErr *RPCErrorDetail
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected RPCErrorDetail, got %T", err)
	}
	if rpcErr.Code != CodeInvalidParams {
		t.Fatalf("code = %d, want %d", rpcErr.Code, CodeInvalidParams)
	}
}

func TestPromptTool_StreamSendsErrorNotification(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := promptTool(executor)

	args, err := json.Marshal(handlers.PromptRequest{
		Backend:      "invalid-backend",
		Prompt:       "hello",
		OutputFormat: "stream-json",
	})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	sender := &countingSender{}
	ctx := WithNotificationSender(context.Background(), sender)

	_, err = tool.Handler(ctx, args)
	if err == nil {
		t.Fatal("expected error for streaming failure")
	}
	if sender.count == 0 {
		t.Fatal("expected error notification to be sent")
	}
}

func TestPromptTool_NonStreaming_DoesNotSendNotifications(t *testing.T) {
	setTempHome(t)
	b, _ := backend.Get("claude")
	if b == nil || !b.IsAvailable() {
		t.Skip("claude backend not available")
	}

	executor := service.NewExecutor()
	tool := promptTool(executor)

	args, err := json.Marshal(handlers.PromptRequest{
		Backend: "claude",
		Prompt:  "hello",
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	sender := &countingSender{}
	ctx := WithNotificationSender(context.Background(), sender)

	_, err = tool.Handler(ctx, args)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if sender.count != 0 {
		t.Fatalf("expected no notifications, got %d", sender.count)
	}
}
