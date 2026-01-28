// Package handlers provides HTTP handlers for the API.
package handlers

import (
	"time"

	"github.com/signalridge/clinvoker/internal/server/service"
	"github.com/signalridge/clinvoker/internal/session"
)

// PromptRequest is the API request for prompt execution.
type PromptRequest struct {
	Backend      string            `json:"backend" doc:"Backend to use (claude, codex, gemini)"`
	Prompt       string            `json:"prompt" doc:"The prompt to execute"`
	Model        string            `json:"model,omitempty" doc:"Model to use"`
	WorkDir      string            `json:"workdir,omitempty" doc:"Working directory"`
	ApprovalMode string            `json:"approval_mode,omitempty" doc:"Approval mode (default, auto, none, always)"`
	SandboxMode  string            `json:"sandbox_mode,omitempty" doc:"Sandbox mode (default, read-only, workspace, full)"`
	OutputFormat string            `json:"output_format,omitempty" doc:"Output format (default, text, json, stream-json)"`
	MaxTokens    int               `json:"max_tokens,omitempty" doc:"Maximum tokens for response"`
	MaxTurns     int               `json:"max_turns,omitempty" doc:"Maximum agentic turns"`
	SystemPrompt string            `json:"system_prompt,omitempty" doc:"Custom system prompt"`
	Verbose      bool              `json:"verbose,omitempty" doc:"Enable verbose output"`
	DryRun       bool              `json:"dry_run,omitempty" doc:"Simulate execution"`
	Ephemeral    bool              `json:"ephemeral,omitempty" doc:"Stateless mode: don't persist session (like standard LLM APIs)"`
	Extra        []string          `json:"extra,omitempty" doc:"Extra backend-specific flags"`
	Metadata     map[string]string `json:"metadata,omitempty" doc:"Custom metadata"`
}

// PromptResponse is the API response for prompt execution.
type PromptResponse struct {
	Body PromptResponseBody
}

// PromptResponseBody is the body of a prompt response.
type PromptResponseBody struct {
	SessionID  string              `json:"session_id,omitempty" doc:"Session ID"`
	Backend    string              `json:"backend" doc:"Backend used"`
	ExitCode   int                 `json:"exit_code" doc:"Exit code (0 = success)"`
	DurationMS int64               `json:"duration_ms" doc:"Execution duration in milliseconds"`
	Output     string              `json:"output,omitempty" doc:"Command output"`
	Error      string              `json:"error,omitempty" doc:"Error message if failed"`
	TokenUsage *session.TokenUsage `json:"token_usage,omitempty" doc:"Token usage statistics"`
}

// ParallelTask is a single task in parallel execution.
type ParallelTask struct {
	Backend      string   `json:"backend" doc:"Backend to use"`
	Prompt       string   `json:"prompt" doc:"The prompt to execute"`
	Model        string   `json:"model,omitempty" doc:"Model to use"`
	WorkDir      string   `json:"workdir,omitempty" doc:"Working directory"`
	ApprovalMode string   `json:"approval_mode,omitempty" doc:"Approval mode"`
	SandboxMode  string   `json:"sandbox_mode,omitempty" doc:"Sandbox mode"`
	MaxTurns     int      `json:"max_turns,omitempty" doc:"Maximum turns"`
	Extra        []string `json:"extra,omitempty" doc:"Extra flags"`
}

// ParallelRequest is the API request for parallel execution.
type ParallelRequest struct {
	Tasks       []ParallelTask `json:"tasks" doc:"Tasks to execute in parallel"`
	MaxParallel int            `json:"max_parallel,omitempty" doc:"Maximum concurrent tasks"`
	FailFast    bool           `json:"fail_fast,omitempty" doc:"Stop on first failure"`
	DryRun      bool           `json:"dry_run,omitempty" doc:"Simulate execution without running commands"`
}

// ParallelResponse is the API response for parallel execution.
type ParallelResponse struct {
	Body ParallelResponseBody
}

// ParallelResponseBody is the body of a parallel response.
type ParallelResponseBody struct {
	TotalTasks    int                  `json:"total_tasks" doc:"Total number of tasks"`
	Completed     int                  `json:"completed" doc:"Number of completed tasks"`
	Failed        int                  `json:"failed" doc:"Number of failed tasks"`
	TotalDuration int64                `json:"total_duration_ms" doc:"Total duration in milliseconds"`
	Results       []PromptResponseBody `json:"results" doc:"Results for each task"`
}

// ChainStep is a step in chain execution.
type ChainStep struct {
	Backend      string `json:"backend" doc:"Backend to use"`
	Prompt       string `json:"prompt" doc:"The prompt (supports {{previous}} placeholder)"`
	Model        string `json:"model,omitempty" doc:"Model to use"`
	WorkDir      string `json:"workdir,omitempty" doc:"Working directory"`
	ApprovalMode string `json:"approval_mode,omitempty" doc:"Approval mode"`
	SandboxMode  string `json:"sandbox_mode,omitempty" doc:"Sandbox mode"`
	MaxTurns     int    `json:"max_turns,omitempty" doc:"Maximum turns"`
	Name         string `json:"name,omitempty" doc:"Step name for display"`
}

// ChainRequest is the API request for chain execution.
type ChainRequest struct {
	Steps          []ChainStep `json:"steps" doc:"Steps to execute in sequence"`
	StopOnFailure  bool        `json:"stop_on_failure,omitempty" doc:"Stop chain on first failure"`
	PassSessionID  bool        `json:"pass_session_id,omitempty" doc:"Pass session ID between steps"`
	PassWorkingDir bool        `json:"pass_working_dir,omitempty" doc:"Pass working directory between steps"`
	DryRun         bool        `json:"dry_run,omitempty" doc:"Simulate execution without running commands"`
}

// ChainStepResult is the result of a single chain step.
type ChainStepResult struct {
	Step       int    `json:"step" doc:"Step number (1-indexed)"`
	Name       string `json:"name,omitempty" doc:"Step name"`
	Backend    string `json:"backend" doc:"Backend used"`
	ExitCode   int    `json:"exit_code" doc:"Exit code"`
	Error      string `json:"error,omitempty" doc:"Error message"`
	SessionID  string `json:"session_id,omitempty" doc:"Session ID"`
	DurationMS int64  `json:"duration_ms" doc:"Duration in milliseconds"`
	Output     string `json:"output,omitempty" doc:"Command output"`
}

// ChainResponse is the API response for chain execution.
type ChainResponse struct {
	Body ChainResponseBody
}

// ChainResponseBody is the body of a chain response.
type ChainResponseBody struct {
	TotalSteps     int               `json:"total_steps" doc:"Total number of steps"`
	CompletedSteps int               `json:"completed_steps" doc:"Number of completed steps"`
	FailedStep     int               `json:"failed_step,omitempty" doc:"Step number that failed"`
	TotalDuration  int64             `json:"total_duration_ms" doc:"Total duration in milliseconds"`
	Results        []ChainStepResult `json:"results" doc:"Results for each step"`
}

// CompareRequest is the API request for compare execution.
type CompareRequest struct {
	Backends   []string `json:"backends" doc:"Backends to compare"`
	Prompt     string   `json:"prompt" doc:"The prompt to run on all backends"`
	Model      string   `json:"model,omitempty" doc:"Model to use (if applicable)"`
	WorkDir    string   `json:"workdir,omitempty" doc:"Working directory"`
	Sequential bool     `json:"sequential,omitempty" doc:"Run sequentially instead of parallel"`
	DryRun     bool     `json:"dry_run,omitempty" doc:"Simulate execution without running commands"`
}

// CompareBackendResult is the result from one backend.
type CompareBackendResult struct {
	Backend    string `json:"backend" doc:"Backend name"`
	Model      string `json:"model,omitempty" doc:"Model used"`
	ExitCode   int    `json:"exit_code" doc:"Exit code"`
	Error      string `json:"error,omitempty" doc:"Error message"`
	DurationMS int64  `json:"duration_ms" doc:"Duration in milliseconds"`
	SessionID  string `json:"session_id,omitempty" doc:"Session ID"`
	Output     string `json:"output,omitempty" doc:"Command output"`
}

// CompareResponse is the API response for compare execution.
type CompareResponse struct {
	Body CompareResponseBody
}

// CompareResponseBody is the body of a compare response.
type CompareResponseBody struct {
	Prompt        string                 `json:"prompt" doc:"The prompt that was executed"`
	Backends      []string               `json:"backends" doc:"Backends that were compared"`
	Results       []CompareBackendResult `json:"results" doc:"Results from each backend"`
	TotalDuration int64                  `json:"total_duration_ms" doc:"Total duration in milliseconds"`
}

// BackendInfo represents information about a backend.
type BackendInfo struct {
	Name      string `json:"name" doc:"Backend name"`
	Available bool   `json:"available" doc:"Whether the backend is available"`
}

// BackendsResponse is the API response for listing backends.
type BackendsResponse struct {
	Body BackendsResponseBody
}

// BackendsResponseBody is the body of a backends response.
type BackendsResponseBody struct {
	Backends []BackendInfo `json:"backends" doc:"Available backends"`
}

// SessionInfo represents session information.
type SessionInfo struct {
	ID            string              `json:"id" doc:"Session ID"`
	Backend       string              `json:"backend" doc:"Backend used"`
	CreatedAt     time.Time           `json:"created_at" doc:"Creation timestamp"`
	LastUsed      time.Time           `json:"last_used" doc:"Last used timestamp"`
	WorkingDir    string              `json:"working_dir,omitempty" doc:"Working directory"`
	Model         string              `json:"model,omitempty" doc:"Model used"`
	InitialPrompt string              `json:"initial_prompt,omitempty" doc:"Initial prompt"`
	Status        string              `json:"status,omitempty" doc:"Session status"`
	TurnCount     int                 `json:"turn_count,omitempty" doc:"Number of turns"`
	TokenUsage    *session.TokenUsage `json:"token_usage,omitempty" doc:"Token usage"`
	Tags          []string            `json:"tags,omitempty" doc:"Session tags"`
	Title         string              `json:"title,omitempty" doc:"Session title"`
}

// SessionsResponse is the API response for listing sessions.
type SessionsResponse struct {
	Body SessionsResponseBody
}

// SessionsResponseBody is the body of a sessions response.
type SessionsResponseBody struct {
	Sessions []SessionInfo `json:"sessions" doc:"List of sessions"`
	Total    int           `json:"total" doc:"Total number of sessions matching the filter"`
	Limit    int           `json:"limit" doc:"Maximum number of sessions returned"`
	Offset   int           `json:"offset" doc:"Number of sessions skipped"`
}

// SessionResponse is the API response for getting a single session.
type SessionResponse struct {
	Body SessionInfo
}

// DeleteSessionResponse is the API response for deleting a session.
type DeleteSessionResponse struct {
	Body DeleteSessionResponseBody
}

// DeleteSessionResponseBody is the body of a delete session response.
type DeleteSessionResponseBody struct {
	Deleted bool   `json:"deleted" doc:"Whether the session was deleted"`
	ID      string `json:"id" doc:"Session ID that was deleted"`
}

// HealthResponse is the API response for health check.
type HealthResponse struct {
	Body HealthResponseBody
}

// HealthResponseBody is the body of a health response.
type HealthResponseBody struct {
	Status string `json:"status" doc:"Health status"`
}

// ToServiceRequest converts API request to service request.
func (r *PromptRequest) ToServiceRequest() *service.PromptRequest {
	return &service.PromptRequest{
		Backend:      r.Backend,
		Prompt:       r.Prompt,
		Model:        r.Model,
		WorkDir:      r.WorkDir,
		ApprovalMode: r.ApprovalMode,
		SandboxMode:  r.SandboxMode,
		OutputFormat: r.OutputFormat,
		MaxTokens:    r.MaxTokens,
		MaxTurns:     r.MaxTurns,
		SystemPrompt: r.SystemPrompt,
		Verbose:      r.Verbose,
		DryRun:       r.DryRun,
		Ephemeral:    r.Ephemeral,
		Extra:        r.Extra,
		Metadata:     r.Metadata,
	}
}

// FromServiceResult converts service result to API response body.
func FromServiceResult(r *service.PromptResult) PromptResponseBody {
	return PromptResponseBody{
		SessionID:  r.SessionID,
		Backend:    r.Backend,
		ExitCode:   r.ExitCode,
		DurationMS: r.DurationMS,
		Output:     r.Output,
		Error:      r.Error,
		TokenUsage: r.TokenUsage,
	}
}
