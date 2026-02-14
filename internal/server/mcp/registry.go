package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/output"
	"github.com/signalridge/clinvoker/internal/server/handlers"
	"github.com/signalridge/clinvoker/internal/server/service"
)

// ToolHandler is the function signature for handling a tool call.
// It receives the raw JSON arguments and returns a result or error.
type ToolHandler func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error)

// ToolDefinition bundles a Tool (name + schema) with its handler.
type ToolDefinition struct {
	Tool    Tool
	Handler ToolHandler
}

// Registry holds all registered MCP tools.
type Registry struct {
	tools map[string]*ToolDefinition
	order []string // preserves registration order
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*ToolDefinition),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(def *ToolDefinition) {
	name := def.Tool.Name
	if _, exists := r.tools[name]; !exists {
		r.order = append(r.order, name)
	}
	r.tools[name] = def
}

// Get returns a tool definition by name.
func (r *Registry) Get(name string) (*ToolDefinition, bool) {
	def, ok := r.tools[name]
	return def, ok
}

// ListTools returns all registered tools in registration order.
func (r *Registry) ListTools() []Tool {
	tools := make([]Tool, 0, len(r.order))
	for _, name := range r.order {
		tools = append(tools, r.tools[name].Tool)
	}
	return tools
}

// RegisterAllTools registers all clinvoker tools into the registry.
func RegisterAllTools(registry *Registry, executor *service.Executor) {
	registry.Register(promptTool(executor))
	registry.Register(parallelTool(executor))
	registry.Register(chainTool(executor))
	registry.Register(compareTool(executor))
	registry.Register(backendsTool(executor))
	registry.Register(sessionsTool(executor))
	registry.Register(sessionGetTool(executor))
	registry.Register(sessionDeleteTool(executor))
	registry.Register(healthTool(executor))
}

func promptTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_prompt",
			Description:  "Execute a prompt on an AI backend (claude, codex, gemini). Returns the model's response.",
			InputSchema:  mustSchemaJSON(handlers.PromptRequest{}),
			OutputSchema: mustSchemaJSON(handlers.PromptResponseBody{}),
		},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			var req handlers.PromptRequest
			if err := json.Unmarshal(args, &req); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err), nil)
			}
			if err := handlers.ValidatePromptRequest(req); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, err.Error(), nil)
			}

			serviceReq := req.ToServiceRequest()
			requestedFormat := strings.ToLower(strings.TrimSpace(req.OutputFormat))
			if requestedFormat == string(backend.OutputStreamJSON) {
				sender := GetNotificationSender(ctx)
				if sender == nil {
					return nil, NewRPCErrorDetail(CodeInvalidParams, "streaming requested but transport does not support notifications", nil)
				}
				payload, err := executeStreamingPrompt(ctx, executor, serviceReq, sender)
				if err != nil {
					return nil, err
				}
				return marshalToToolCallResult(payload)
			}

			result, err := executor.ExecutePrompt(ctx, serviceReq)
			if err != nil {
				code, msg := MapExecutorError(err)
				return nil, NewRPCErrorDetail(code, msg, nil)
			}
			if result.Error != "" {
				code, msg := MapExecutorError(errors.New(result.Error))
				return nil, NewRPCErrorDetail(code, msg, nil)
			}
			if result.ExitCode != 0 {
				return nil, NewRPCErrorDetail(CodeToolExecutionError,
					fmt.Sprintf("execution failed with exit code %d", result.ExitCode), nil)
			}
			payload := handlers.FromServiceResult(result)
			return marshalToToolCallResult(payload)
		},
	}
}

func parallelTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_parallel",
			Description:  "Execute multiple prompts in parallel across AI backends.",
			InputSchema:  mustSchemaJSON(handlers.ParallelRequest{}),
			OutputSchema: mustSchemaJSON(handlers.ParallelResponseBody{}),
		},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			var req handlers.ParallelRequest
			if err := json.Unmarshal(args, &req); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err), nil)
			}
			if err := handlers.ValidateParallelRequest(req); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, err.Error(), nil)
			}
			serviceReq := &service.ParallelRequest{
				MaxParallel: req.MaxParallel,
				FailFast:    req.FailFast,
				DryRun:      req.DryRun,
				Tasks:       make([]service.PromptRequest, len(req.Tasks)),
			}
			for i, t := range req.Tasks {
				serviceReq.Tasks[i] = service.PromptRequest{
					Backend:      t.Backend,
					Prompt:       t.Prompt,
					Model:        t.Model,
					WorkDir:      t.WorkDir,
					ApprovalMode: t.ApprovalMode,
					SandboxMode:  t.SandboxMode,
					OutputFormat: t.OutputFormat,
					MaxTokens:    t.MaxTokens,
					MaxTurns:     t.MaxTurns,
					SystemPrompt: t.SystemPrompt,
					Verbose:      t.Verbose,
					Ephemeral:    t.Ephemeral,
					Extra:        t.Extra,
					Metadata:     t.Metadata,
				}
			}
			result, err := executor.ExecuteParallel(ctx, serviceReq)
			if err != nil {
				code, msg := MapExecutorError(err)
				return nil, NewRPCErrorDetail(code, msg, nil)
			}
			return marshalToToolCallResult(result)
		},
	}
}

func chainTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_chain",
			Description:  "Execute prompts in sequence (chain), where each step can reference the previous output via {{previous}}.",
			InputSchema:  mustSchemaJSON(handlers.ChainRequest{}),
			OutputSchema: mustSchemaJSON(handlers.ChainResponseBody{}),
		},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			var req handlers.ChainRequest
			if err := json.Unmarshal(args, &req); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err), nil)
			}
			if err := handlers.ValidateChainRequest(req); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, err.Error(), nil)
			}
			serviceReq := &service.ChainRequest{
				StopOnFailure:  req.StopOnFailure,
				PassWorkingDir: req.PassWorkingDir,
				DryRun:         req.DryRun,
				Steps:          make([]service.ChainStep, len(req.Steps)),
			}
			for i, s := range req.Steps {
				serviceReq.Steps[i] = service.ChainStep{
					Backend:      s.Backend,
					Prompt:       s.Prompt,
					Model:        s.Model,
					WorkDir:      s.WorkDir,
					ApprovalMode: s.ApprovalMode,
					SandboxMode:  s.SandboxMode,
					MaxTokens:    s.MaxTokens,
					MaxTurns:     s.MaxTurns,
					SystemPrompt: s.SystemPrompt,
					Verbose:      s.Verbose,
					Extra:        s.Extra,
					Name:         s.Name,
				}
			}
			result, err := executor.ExecuteChain(ctx, serviceReq)
			if err != nil {
				code, msg := MapExecutorError(err)
				return nil, NewRPCErrorDetail(code, msg, nil)
			}
			return marshalToToolCallResult(result)
		},
	}
}

func compareTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_compare",
			Description:  "Run the same prompt on multiple backends and compare their responses.",
			InputSchema:  mustSchemaJSON(handlers.CompareRequest{}),
			OutputSchema: mustSchemaJSON(handlers.CompareResponseBody{}),
		},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			var req handlers.CompareRequest
			if err := json.Unmarshal(args, &req); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err), nil)
			}
			if err := handlers.ValidateCompareRequest(req); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, err.Error(), nil)
			}
			serviceReq := &service.CompareRequest{
				Backends:   req.Backends,
				Prompt:     req.Prompt,
				Model:      req.Model,
				WorkDir:    req.WorkDir,
				Sequential: req.Sequential,
				DryRun:     req.DryRun,
			}
			result, err := executor.ExecuteCompare(ctx, serviceReq)
			if err != nil {
				code, msg := MapExecutorError(err)
				return nil, NewRPCErrorDetail(code, msg, nil)
			}
			return marshalToToolCallResult(result)
		},
	}
}

func backendsTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_backends",
			Description:  "List available AI backends and their availability status.",
			InputSchema:  mustSchemaJSON(handlers.BackendsInput{}),
			OutputSchema: mustSchemaJSON(handlers.BackendsResponseBody{}),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (*ToolCallResult, error) {
			backends := executor.ListBackends(ctx)
			return marshalToToolCallResult(map[string]any{"backends": backends})
		},
	}
}

func sessionsTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_sessions",
			Description:  "List all sessions with optional filtering.",
			InputSchema:  mustSchemaJSON(handlers.SessionsInput{}),
			OutputSchema: mustSchemaJSON(handlers.SessionsResponseBody{}),
		},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			var opts service.SessionListOptions
			if len(args) > 0 {
				if err := json.Unmarshal(args, &opts); err != nil {
					return nil, NewRPCErrorDetail(CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err), nil)
				}
			}
			result, err := executor.ListSessionsPaginated(ctx, &opts)
			if err != nil {
				code, msg := MapExecutorError(err)
				return nil, NewRPCErrorDetail(code, msg, nil)
			}
			return marshalToToolCallResult(result)
		},
	}
}

func sessionGetTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_session_get",
			Description:  "Get a session by ID or ID prefix.",
			InputSchema:  mustSchemaJSON(handlers.GetSessionInput{}),
			OutputSchema: mustSchemaJSON(handlers.SessionInfo{}),
		},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			var params struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err), nil)
			}
			if strings.TrimSpace(params.ID) == "" {
				return nil, NewRPCErrorDetail(CodeInvalidParams, "session id is required", nil)
			}
			result, err := executor.GetSession(ctx, params.ID)
			if err != nil {
				code, msg := MapExecutorError(err)
				return nil, NewRPCErrorDetail(code, msg, nil)
			}
			return marshalToToolCallResult(result)
		},
	}
}

func sessionDeleteTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_session_delete",
			Description:  "Delete a session by ID or ID prefix.",
			InputSchema:  mustSchemaJSON(handlers.DeleteSessionInput{}),
			OutputSchema: mustSchemaJSON(handlers.DeleteSessionResponseBody{}),
		},
		Handler: func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error) {
			var params struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return nil, NewRPCErrorDetail(CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err), nil)
			}
			if strings.TrimSpace(params.ID) == "" {
				return nil, NewRPCErrorDetail(CodeInvalidParams, "session id is required", nil)
			}
			if err := executor.DeleteSession(ctx, params.ID); err != nil {
				code, msg := MapExecutorError(err)
				return nil, NewRPCErrorDetail(code, msg, nil)
			}
			return marshalToToolCallResult(map[string]any{"deleted": true, "id": params.ID})
		},
	}
}

func healthTool(executor *service.Executor) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:         "clinvk_health",
			Description:  "Check the health of clinvoker and its backends.",
			InputSchema:  mustSchemaJSON(handlers.HealthInput{}),
			OutputSchema: mustSchemaJSON(handlers.HealthResponseBody{}),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (*ToolCallResult, error) {
			backends := executor.ListBackends(ctx)
			storeHealth := executor.GetSessionStoreHealth(ctx)
			backendStatus := make([]handlers.BackendHealthStatus, len(backends))
			allBackendsAvailable := true
			for i, b := range backends {
				backendStatus[i] = handlers.BackendHealthStatus{
					Name:      b.Name,
					Available: b.Available,
				}
				if !b.Available {
					allBackendsAvailable = false
				}
			}

			sessionStoreStatus := handlers.SessionStoreStatus{
				Available:    storeHealth.Available,
				SessionCount: storeHealth.SessionCount,
				Error:        storeHealth.Error,
			}

			status := "ok"
			if !allBackendsAvailable {
				status = "degraded"
			}
			if !storeHealth.Available {
				status = "unhealthy"
			}

			uptime := time.Since(mcpStartTime)
			result := handlers.HealthResponseBody{
				Status:       status,
				Version:      ServerVersion,
				Uptime:       formatDuration(uptime),
				UptimeMillis: uptime.Milliseconds(),
				Backends:     backendStatus,
				SessionStore: sessionStoreStatus,
			}

			return marshalToToolCallResult(result)
		},
	}
}

var mcpStartTime = time.Now()

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// marshalToToolCallResult marshals any value to JSON and wraps it in a ToolCallResult.
func marshalToToolCallResult(v any) (*ToolCallResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return &ToolCallResult{
		Content: []ContentBlock{TextContent(string(data))},
	}, nil
}

// mustJSON marshals a value to json.RawMessage, panicking on error.
func mustJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("mustJSON: %v", err))
	}
	return data
}

// executeStreamingPrompt runs a prompt via StreamPrompt, sending MCP notifications
// for each event and returning the final accumulated result.
func executeStreamingPrompt(ctx context.Context, executor *service.Executor, req *service.PromptRequest, sender NotificationSender) (*handlers.PromptResponseBody, error) {
	start := time.Now()
	progressLabel := "clinvk_prompt"
	var accumulated strings.Builder
	var lastError string
	var sawErrorEvent bool

	onEvent := func(event *output.UnifiedEvent) error {
		// Accumulate message text for the final result.
		if event.Type == output.EventMessage {
			if content, err := event.GetMessageContent(); err == nil {
				accumulated.WriteString(content.Text)
			}
		}

		// Track errors.
		if event.Type == output.EventError {
			if content, err := event.GetErrorContent(); err == nil {
				lastError = content.Message
				sawErrorEvent = true
			}
		}

		notification := TranslateEvent(event, progressLabel)
		if notification == nil {
			return nil
		}

		return sender.SendNotification(notification)
	}

	result, err := executor.StreamPrompt(ctx, req, onEvent)
	if err != nil || result == nil || result.Error != "" || lastError != "" {
		errMsg := "streaming error"
		if err != nil {
			errMsg = err.Error()
		} else if result != nil && result.Error != "" {
			errMsg = result.Error
		} else if lastError != "" {
			errMsg = lastError
		}

		if !sawErrorEvent {
			errEvent := output.NewUnifiedEvent(output.EventError, req.Backend, "")
			if setErr := errEvent.SetContent(&output.ErrorContent{Message: errMsg}); setErr == nil {
				if notification := TranslateEvent(errEvent, progressLabel); notification != nil {
					_ = sender.SendNotification(notification)
				}
			}
		}

		code, msg := MapExecutorError(errors.New(errMsg))
		return nil, NewRPCErrorDetail(code, msg, nil)
	}

	if result.ExitCode != 0 {
		return nil, NewRPCErrorDetail(
			CodeToolExecutionError,
			fmt.Sprintf("execution failed with exit code %d", result.ExitCode),
			nil,
		)
	}

	durationMS := time.Since(start).Milliseconds()
	payload := &handlers.PromptResponseBody{
		Backend:    req.Backend,
		ExitCode:   result.ExitCode,
		DurationMS: durationMS,
		Output:     accumulated.String(),
		TokenUsage: result.TokenUsage,
	}
	return payload, nil
}
