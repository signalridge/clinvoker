package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/signalridge/clinvoker/internal/server/service"
)

// CustomHandlers provides handlers for the custom RESTful API.
type CustomHandlers struct {
	executor *service.Executor
}

// NewCustomHandlers creates a new custom handlers instance.
func NewCustomHandlers(executor *service.Executor) *CustomHandlers {
	return &CustomHandlers{executor: executor}
}

// Register registers all custom API routes.
func (h *CustomHandlers) Register(api huma.API) {
	// Prompt endpoint
	huma.Register(api, huma.Operation{
		OperationID: "executePrompt",
		Method:      http.MethodPost,
		Path:        "/api/v1/prompt",
		Summary:     "Execute a prompt",
		Description: "Execute a prompt on a specified backend",
		Tags:        []string{"Custom API"},
	}, h.HandlePrompt)

	// Parallel endpoint
	huma.Register(api, huma.Operation{
		OperationID: "executeParallel",
		Method:      http.MethodPost,
		Path:        "/api/v1/parallel",
		Summary:     "Execute parallel tasks",
		Description: "Execute multiple prompts in parallel across backends",
		Tags:        []string{"Custom API"},
	}, h.HandleParallel)

	// Chain endpoint
	huma.Register(api, huma.Operation{
		OperationID: "executeChain",
		Method:      http.MethodPost,
		Path:        "/api/v1/chain",
		Summary:     "Execute chain of tasks",
		Description: "Execute a chain of prompts in sequence with context passing",
		Tags:        []string{"Custom API"},
	}, h.HandleChain)

	// Compare endpoint
	huma.Register(api, huma.Operation{
		OperationID: "executeCompare",
		Method:      http.MethodPost,
		Path:        "/api/v1/compare",
		Summary:     "Compare backends",
		Description: "Run the same prompt on multiple backends for comparison",
		Tags:        []string{"Custom API"},
	}, h.HandleCompare)

	// Backends endpoint
	huma.Register(api, huma.Operation{
		OperationID: "listBackends",
		Method:      http.MethodGet,
		Path:        "/api/v1/backends",
		Summary:     "List backends",
		Description: "List all available AI backends",
		Tags:        []string{"Custom API"},
	}, h.HandleBackends)

	// Sessions endpoints
	huma.Register(api, huma.Operation{
		OperationID: "listSessions",
		Method:      http.MethodGet,
		Path:        "/api/v1/sessions",
		Summary:     "List sessions",
		Description: "List all sessions",
		Tags:        []string{"Custom API"},
	}, h.HandleSessions)

	huma.Register(api, huma.Operation{
		OperationID: "getSession",
		Method:      http.MethodGet,
		Path:        "/api/v1/sessions/{id}",
		Summary:     "Get session",
		Description: "Get details of a specific session",
		Tags:        []string{"Custom API"},
	}, h.HandleGetSession)

	huma.Register(api, huma.Operation{
		OperationID: "deleteSession",
		Method:      http.MethodDelete,
		Path:        "/api/v1/sessions/{id}",
		Summary:     "Delete session",
		Description: "Delete a session",
		Tags:        []string{"Custom API"},
	}, h.HandleDeleteSession)

	// Health endpoint
	huma.Register(api, huma.Operation{
		OperationID: "healthCheck",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Description: "Check if the server is healthy",
		Tags:        []string{"Health"},
	}, h.HandleHealth)
}

// PromptInput is the input for the prompt handler.
type PromptInput struct {
	Body PromptRequest
}

// HandlePrompt handles prompt execution requests.
func (h *CustomHandlers) HandlePrompt(ctx context.Context, input *PromptInput) (*PromptResponse, error) {
	if input.Body.Backend == "" {
		return nil, huma.Error400BadRequest("backend is required")
	}
	if input.Body.Prompt == "" {
		return nil, huma.Error400BadRequest("prompt is required")
	}

	result, err := h.executor.ExecutePrompt(ctx, input.Body.ToServiceRequest())
	if err != nil {
		return nil, huma.Error500InternalServerError("execution failed", err)
	}

	return &PromptResponse{
		Body: FromServiceResult(result),
	}, nil
}

// ParallelInput is the input for the parallel handler.
type ParallelInput struct {
	Body ParallelRequest
}

// HandleParallel handles parallel execution requests.
func (h *CustomHandlers) HandleParallel(ctx context.Context, input *ParallelInput) (*ParallelResponse, error) {
	if len(input.Body.Tasks) == 0 {
		return nil, huma.Error400BadRequest("tasks are required")
	}

	// Convert to service request
	serviceReq := &service.ParallelRequest{
		MaxParallel: input.Body.MaxParallel,
		FailFast:    input.Body.FailFast,
		DryRun:      input.Body.DryRun,
		Tasks:       make([]service.PromptRequest, len(input.Body.Tasks)),
	}

	for i, t := range input.Body.Tasks {
		serviceReq.Tasks[i] = service.PromptRequest{
			Backend:      t.Backend,
			Prompt:       t.Prompt,
			Model:        t.Model,
			WorkDir:      t.WorkDir,
			ApprovalMode: t.ApprovalMode,
			SandboxMode:  t.SandboxMode,
			MaxTurns:     t.MaxTurns,
			Extra:        t.Extra,
		}
	}

	result, err := h.executor.ExecuteParallel(ctx, serviceReq)
	if err != nil {
		return nil, huma.Error500InternalServerError("parallel execution failed", err)
	}

	// Convert results
	results := make([]PromptResponseBody, len(result.Results))
	for i, r := range result.Results {
		results[i] = PromptResponseBody{
			SessionID:  r.SessionID,
			Backend:    r.Backend,
			ExitCode:   r.ExitCode,
			DurationMS: r.DurationMS,
			Output:     r.Output,
			Error:      r.Error,
			TokenUsage: r.TokenUsage,
		}
	}

	return &ParallelResponse{
		Body: ParallelResponseBody{
			TotalTasks:    result.TotalTasks,
			Completed:     result.Completed,
			Failed:        result.Failed,
			TotalDuration: result.TotalDuration,
			Results:       results,
		},
	}, nil
}

// ChainInput is the input for the chain handler.
type ChainInput struct {
	Body ChainRequest
}

// HandleChain handles chain execution requests.
func (h *CustomHandlers) HandleChain(ctx context.Context, input *ChainInput) (*ChainResponse, error) {
	if len(input.Body.Steps) == 0 {
		return nil, huma.Error400BadRequest("steps are required")
	}
	if input.Body.PassSessionID || input.Body.PersistSessions {
		return nil, huma.Error400BadRequest("chain is always ephemeral; pass_session_id and persist_sessions are not supported")
	}
	for i, step := range input.Body.Steps {
		if strings.Contains(step.Prompt, "{{session}}") {
			return nil, huma.Error400BadRequest(fmt.Sprintf("chain step %d uses {{session}} but sessions are not persisted", i+1))
		}
	}

	// Convert to service request
	serviceReq := &service.ChainRequest{
		StopOnFailure:  input.Body.StopOnFailure,
		PassWorkingDir: input.Body.PassWorkingDir,
		DryRun:         input.Body.DryRun,
		Steps:          make([]service.ChainStep, len(input.Body.Steps)),
	}

	for i, s := range input.Body.Steps {
		serviceReq.Steps[i] = service.ChainStep{
			Backend:      s.Backend,
			Prompt:       s.Prompt,
			Model:        s.Model,
			WorkDir:      s.WorkDir,
			ApprovalMode: s.ApprovalMode,
			SandboxMode:  s.SandboxMode,
			MaxTurns:     s.MaxTurns,
			Name:         s.Name,
		}
	}

	result, err := h.executor.ExecuteChain(ctx, serviceReq)
	if err != nil {
		return nil, huma.Error500InternalServerError("chain execution failed", err)
	}

	// Convert results
	results := make([]ChainStepResult, len(result.Results))
	for i, r := range result.Results {
		results[i] = ChainStepResult{
			Step:       r.Step,
			Name:       r.Name,
			Backend:    r.Backend,
			ExitCode:   r.ExitCode,
			Error:      r.Error,
			SessionID:  r.SessionID,
			DurationMS: r.DurationMS,
			Output:     r.Output,
		}
	}

	return &ChainResponse{
		Body: ChainResponseBody{
			TotalSteps:     result.TotalSteps,
			CompletedSteps: result.CompletedSteps,
			FailedStep:     result.FailedStep,
			TotalDuration:  result.TotalDuration,
			Results:        results,
		},
	}, nil
}

// CompareInput is the input for the compare handler.
type CompareInput struct {
	Body CompareRequest
}

// HandleCompare handles compare execution requests.
func (h *CustomHandlers) HandleCompare(ctx context.Context, input *CompareInput) (*CompareResponse, error) {
	if len(input.Body.Backends) == 0 {
		return nil, huma.Error400BadRequest("backends are required")
	}
	if input.Body.Prompt == "" {
		return nil, huma.Error400BadRequest("prompt is required")
	}

	serviceReq := &service.CompareRequest{
		Backends:   input.Body.Backends,
		Prompt:     input.Body.Prompt,
		Model:      input.Body.Model,
		WorkDir:    input.Body.WorkDir,
		Sequential: input.Body.Sequential,
		DryRun:     input.Body.DryRun,
	}

	result, err := h.executor.ExecuteCompare(ctx, serviceReq)
	if err != nil {
		return nil, huma.Error500InternalServerError("compare execution failed", err)
	}

	// Convert results
	results := make([]CompareBackendResult, len(result.Results))
	for i, r := range result.Results {
		results[i] = CompareBackendResult{
			Backend:    r.Backend,
			Model:      r.Model,
			ExitCode:   r.ExitCode,
			Error:      r.Error,
			DurationMS: r.DurationMS,
			SessionID:  r.SessionID,
			Output:     r.Output,
		}
	}

	return &CompareResponse{
		Body: CompareResponseBody{
			Prompt:        result.Prompt,
			Backends:      result.Backends,
			Results:       results,
			TotalDuration: result.TotalDuration,
		},
	}, nil
}

// BackendsInput is the input for the backends handler.
type BackendsInput struct{}

// HandleBackends handles backend listing requests.
func (h *CustomHandlers) HandleBackends(ctx context.Context, _ *BackendsInput) (*BackendsResponse, error) {
	backends := h.executor.ListBackends(ctx)

	infos := make([]BackendInfo, len(backends))
	for i, b := range backends {
		infos[i] = BackendInfo{
			Name:      b.Name,
			Available: b.Available,
		}
	}

	return &BackendsResponse{
		Body: BackendsResponseBody{
			Backends: infos,
		},
	}, nil
}

// SessionsInput is the input for the sessions handler.
type SessionsInput struct {
	Backend string `query:"backend" doc:"Filter by backend name"`
	Status  string `query:"status" doc:"Filter by status (active, completed, error)"`
	Limit   int    `query:"limit" doc:"Maximum number of sessions to return (default: 100)"`
	Offset  int    `query:"offset" doc:"Number of sessions to skip for pagination"`
}

// HandleSessions handles session listing requests.
func (h *CustomHandlers) HandleSessions(ctx context.Context, input *SessionsInput) (*SessionsResponse, error) {
	opts := &service.SessionListOptions{
		Backend: input.Backend,
		Status:  input.Status,
		Limit:   input.Limit,
		Offset:  input.Offset,
	}

	result, err := h.executor.ListSessionsPaginated(ctx, opts)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list sessions", err)
	}

	infos := make([]SessionInfo, len(result.Sessions))
	for i, s := range result.Sessions {
		infos[i] = SessionInfo{
			ID:            s.ID,
			Backend:       s.Backend,
			CreatedAt:     s.CreatedAt,
			LastUsed:      s.LastUsed,
			WorkingDir:    s.WorkingDir,
			Model:         s.Model,
			InitialPrompt: s.InitialPrompt,
			Status:        s.Status,
			TurnCount:     s.TurnCount,
			TokenUsage:    s.TokenUsage,
			Tags:          s.Tags,
			Title:         s.Title,
		}
	}

	return &SessionsResponse{
		Body: SessionsResponseBody{
			Sessions: infos,
			Total:    result.Total,
			Limit:    result.Limit,
			Offset:   result.Offset,
		},
	}, nil
}

// GetSessionInput is the input for getting a single session.
type GetSessionInput struct {
	ID string `path:"id" doc:"Session ID or prefix"`
}

// HandleGetSession handles get session requests.
func (h *CustomHandlers) HandleGetSession(ctx context.Context, input *GetSessionInput) (*SessionResponse, error) {
	sess, err := h.executor.GetSession(ctx, input.ID)
	if err != nil {
		return nil, huma.Error404NotFound("session not found", err)
	}

	return &SessionResponse{
		Body: SessionInfo{
			ID:            sess.ID,
			Backend:       sess.Backend,
			CreatedAt:     sess.CreatedAt,
			LastUsed:      sess.LastUsed,
			WorkingDir:    sess.WorkingDir,
			Model:         sess.Model,
			InitialPrompt: sess.InitialPrompt,
			Status:        sess.Status,
			TurnCount:     sess.TurnCount,
			TokenUsage:    sess.TokenUsage,
			Tags:          sess.Tags,
			Title:         sess.Title,
		},
	}, nil
}

// DeleteSessionInput is the input for deleting a session.
type DeleteSessionInput struct {
	ID string `path:"id" doc:"Session ID or prefix"`
}

// HandleDeleteSession handles delete session requests.
func (h *CustomHandlers) HandleDeleteSession(ctx context.Context, input *DeleteSessionInput) (*DeleteSessionResponse, error) {
	err := h.executor.DeleteSession(ctx, input.ID)
	if err != nil {
		return nil, huma.Error404NotFound("session not found", err)
	}

	return &DeleteSessionResponse{
		Body: DeleteSessionResponseBody{
			Deleted: true,
			ID:      input.ID,
		},
	}, nil
}

// HealthInput is the input for the health handler.
type HealthInput struct{}

// HandleHealth handles health check requests.
// Returns overall status and individual backend availability.
func (h *CustomHandlers) HandleHealth(ctx context.Context, _ *HealthInput) (*HealthResponse, error) {
	// Get backend status
	backends := h.executor.ListBackends(ctx)

	backendStatus := make([]BackendHealthStatus, len(backends))
	allAvailable := true
	for i, b := range backends {
		backendStatus[i] = BackendHealthStatus{
			Name:      b.Name,
			Available: b.Available,
		}
		if !b.Available {
			allAvailable = false
		}
	}

	// Determine overall status
	status := "ok"
	if !allAvailable {
		status = "degraded"
	}

	return &HealthResponse{
		Body: HealthResponseBody{
			Status:   status,
			Backends: backendStatus,
		},
	}, nil
}
