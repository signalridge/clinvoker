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

func TestParallelTool_ReturnsResponseBody(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := parallelTool(executor)

	args, err := json.Marshal(handlers.ParallelRequest{
		Tasks: []handlers.ParallelTask{
			{Backend: "invalid-backend", Prompt: "hello"},
		},
	})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	result, err := tool.Handler(context.Background(), args)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.ParallelResponseBody
	if err := json.Unmarshal([]byte(result.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}

	if body.TotalTasks != 1 {
		t.Errorf("total_tasks = %d, want 1", body.TotalTasks)
	}
	if body.Completed != 0 {
		t.Errorf("completed = %d, want 0", body.Completed)
	}
	if body.Failed != 1 {
		t.Errorf("failed = %d, want 1", body.Failed)
	}
	if len(body.Results) != 1 {
		t.Fatalf("results length = %d, want 1", len(body.Results))
	}
	if body.Results[0].Backend != "invalid-backend" {
		t.Errorf("result backend = %q, want %q", body.Results[0].Backend, "invalid-backend")
	}
	if body.Results[0].ExitCode != 1 {
		t.Errorf("result exit_code = %d, want 1", body.Results[0].ExitCode)
	}
	if body.Results[0].Error == "" {
		t.Error("expected error for invalid backend")
	}
}

func TestParallelTool_EmptyTasksReturnsRPCError(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := parallelTool(executor)

	args, err := json.Marshal(handlers.ParallelRequest{})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	result, err := tool.Handler(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty tasks")
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
	if rpcErr.Message != "tasks are required" {
		t.Fatalf("message = %q, want %q", rpcErr.Message, "tasks are required")
	}
}

func TestChainTool_ReturnsResponseBody(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := chainTool(executor)

	args, err := json.Marshal(handlers.ChainRequest{
		Steps: []handlers.ChainStep{
			{Backend: "invalid-backend", Prompt: "hello"},
		},
	})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	result, err := tool.Handler(context.Background(), args)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.ChainResponseBody
	if err := json.Unmarshal([]byte(result.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}

	if body.TotalSteps != 1 {
		t.Errorf("total_steps = %d, want 1", body.TotalSteps)
	}
	if len(body.Results) != 1 {
		t.Fatalf("results length = %d, want 1", len(body.Results))
	}
	if body.Results[0].Backend != "invalid-backend" {
		t.Errorf("result backend = %q, want %q", body.Results[0].Backend, "invalid-backend")
	}
	if body.Results[0].ExitCode != 1 {
		t.Errorf("result exit_code = %d, want 1", body.Results[0].ExitCode)
	}
	if body.Results[0].Error == "" {
		t.Error("expected error for invalid backend")
	}
}

func TestChainTool_ValidationErrors(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := chainTool(executor)

	cases := []struct {
		name string
		req  handlers.ChainRequest
		want string
	}{
		{
			name: "missing steps",
			req:  handlers.ChainRequest{},
			want: "steps are required",
		},
		{
			name: "pass_session_id not supported",
			req: handlers.ChainRequest{
				PassSessionID: true,
				Steps: []handlers.ChainStep{
					{Backend: "invalid-backend", Prompt: "hello"},
				},
			},
			want: "chain is always ephemeral; pass_session_id and persist_sessions are not supported",
		},
		{
			name: "persist_sessions not supported",
			req: handlers.ChainRequest{
				PersistSessions: true,
				Steps: []handlers.ChainStep{
					{Backend: "invalid-backend", Prompt: "hello"},
				},
			},
			want: "chain is always ephemeral; pass_session_id and persist_sessions are not supported",
		},
		{
			name: "session placeholder not allowed",
			req: handlers.ChainRequest{
				Steps: []handlers.ChainStep{
					{Backend: "invalid-backend", Prompt: "hello {{session}}"},
				},
			},
			want: "chain step 1 uses {{session}} but sessions are not persisted",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatalf("marshal args: %v", err)
			}

			result, err := tool.Handler(context.Background(), args)
			if err == nil {
				t.Fatal("expected error")
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
			if rpcErr.Message != tc.want {
				t.Fatalf("message = %q, want %q", rpcErr.Message, tc.want)
			}
		})
	}
}

func TestCompareTool_ReturnsResponseBody(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := compareTool(executor)

	args, err := json.Marshal(handlers.CompareRequest{
		Backends:   []string{"invalid-backend"},
		Prompt:     "hello",
		Sequential: true,
	})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	result, err := tool.Handler(context.Background(), args)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.CompareResponseBody
	if err := json.Unmarshal([]byte(result.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}

	if body.Prompt != "hello" {
		t.Errorf("prompt = %q, want %q", body.Prompt, "hello")
	}
	if len(body.Results) != 1 {
		t.Fatalf("results length = %d, want 1", len(body.Results))
	}
	if body.Results[0].Backend != "invalid-backend" {
		t.Errorf("result backend = %q, want %q", body.Results[0].Backend, "invalid-backend")
	}
	if body.Results[0].ExitCode != 1 {
		t.Errorf("result exit_code = %d, want 1", body.Results[0].ExitCode)
	}
	if body.Results[0].Error == "" {
		t.Error("expected error for invalid backend")
	}
}

func TestCompareTool_ValidationErrors(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := compareTool(executor)

	cases := []struct {
		name string
		req  handlers.CompareRequest
		want string
	}{
		{
			name: "missing backends",
			req:  handlers.CompareRequest{Prompt: "hello"},
			want: "backends are required",
		},
		{
			name: "missing prompt",
			req:  handlers.CompareRequest{Backends: []string{"invalid-backend"}},
			want: "prompt is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatalf("marshal args: %v", err)
			}

			result, err := tool.Handler(context.Background(), args)
			if err == nil {
				t.Fatal("expected error")
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
			if rpcErr.Message != tc.want {
				t.Fatalf("message = %q, want %q", rpcErr.Message, tc.want)
			}
		})
	}
}

func TestBackendsTool_ReturnsResponseBody(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := backendsTool(executor)

	result, err := tool.Handler(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.BackendsResponseBody
	if err := json.Unmarshal([]byte(result.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if body.Backends == nil {
		t.Fatal("expected backends slice to be non-nil")
	}
}

func TestSessionsTool_ReturnsResponseBody(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := sessionsTool(executor)

	result, err := tool.Handler(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.SessionsResponseBody
	if err := json.Unmarshal([]byte(result.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if body.Sessions == nil {
		t.Fatal("expected sessions slice to be non-nil")
	}
	if body.Total < 0 || body.Limit < 0 || body.Offset < 0 {
		t.Fatalf("invalid pagination values: %+v", body)
	}
}

func TestSessionGetTool_ReturnsResponseBody(t *testing.T) {
	setTempHome(t)
	backendName := availableBackend(t)

	executor := service.NewExecutor()
	result, err := executor.ExecutePrompt(context.Background(), &service.PromptRequest{
		Backend: backendName,
		Prompt:  "hello",
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("execute prompt error: %v", err)
	}
	if result.SessionID == "" {
		t.Fatal("expected session id")
	}

	tool := sessionGetTool(executor)
	args, err := json.Marshal(map[string]string{"id": result.SessionID})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	toolResult, err := tool.Handler(context.Background(), args)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if toolResult == nil || len(toolResult.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.SessionInfo
	if err := json.Unmarshal([]byte(toolResult.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if body.ID != result.SessionID {
		t.Errorf("id = %q, want %q", body.ID, result.SessionID)
	}
	if body.Backend != backendName {
		t.Errorf("backend = %q, want %q", body.Backend, backendName)
	}
}

func TestSessionDeleteTool_ReturnsResponseBody(t *testing.T) {
	setTempHome(t)
	backendName := availableBackend(t)

	executor := service.NewExecutor()
	result, err := executor.ExecutePrompt(context.Background(), &service.PromptRequest{
		Backend: backendName,
		Prompt:  "hello",
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("execute prompt error: %v", err)
	}
	if result.SessionID == "" {
		t.Fatal("expected session id")
	}

	tool := sessionDeleteTool(executor)
	args, err := json.Marshal(map[string]string{"id": result.SessionID})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	toolResult, err := tool.Handler(context.Background(), args)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if toolResult == nil || len(toolResult.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.DeleteSessionResponseBody
	if err := json.Unmarshal([]byte(toolResult.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if !body.Deleted {
		t.Error("expected deleted=true")
	}
	if body.ID != result.SessionID {
		t.Errorf("id = %q, want %q", body.ID, result.SessionID)
	}
}

func TestHealthTool_ReturnsResponseBody(t *testing.T) {
	setTempHome(t)
	executor := service.NewExecutor()
	tool := healthTool(executor)

	result, err := tool.Handler(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	var body handlers.HealthResponseBody
	if err := json.Unmarshal([]byte(result.Content[0].Text), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if body.Status == "" {
		t.Error("expected status")
	}
	if body.Version == "" {
		t.Error("expected version")
	}
	if body.Uptime == "" {
		t.Error("expected uptime")
	}
}

func availableBackend(t *testing.T) string {
	t.Helper()
	for _, name := range backend.List() {
		b, _ := backend.Get(name)
		if b != nil && b.IsAvailable() {
			return name
		}
	}
	t.Skip("no backend available")
	return ""
}
