// Package service provides the execution layer for API handlers.
package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/executor"
	"github.com/signalridge/clinvoker/internal/session"
)

// Executor handles the execution of AI backend commands.
type Executor struct {
	store *session.Store
}

// NewExecutor creates a new executor.
func NewExecutor() *Executor {
	return &Executor{
		store: session.NewStore(),
	}
}

// PromptRequest represents a prompt execution request.
type PromptRequest struct {
	Backend      string            `json:"backend"`
	Prompt       string            `json:"prompt"`
	Model        string            `json:"model,omitempty"`
	WorkDir      string            `json:"workdir,omitempty"`
	ApprovalMode string            `json:"approval_mode,omitempty"`
	SandboxMode  string            `json:"sandbox_mode,omitempty"`
	OutputFormat string            `json:"output_format,omitempty"`
	MaxTokens    int               `json:"max_tokens,omitempty"`
	MaxTurns     int               `json:"max_turns,omitempty"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Verbose      bool              `json:"verbose,omitempty"`
	DryRun       bool              `json:"dry_run,omitempty"`
	Extra        []string          `json:"extra,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PromptResult represents the result of a prompt execution.
type PromptResult struct {
	SessionID  string              `json:"session_id,omitempty"`
	Backend    string              `json:"backend"`
	ExitCode   int                 `json:"exit_code"`
	DurationMS int64               `json:"duration_ms"`
	Output     string              `json:"output,omitempty"`
	Error      string              `json:"error,omitempty"`
	TokenUsage *session.TokenUsage `json:"token_usage,omitempty"`
}

// ExecutePrompt executes a single prompt.
func (e *Executor) ExecutePrompt(ctx context.Context, req *PromptRequest) (*PromptResult, error) {
	start := time.Now()
	result := &PromptResult{
		Backend: req.Backend,
	}

	// Get backend
	b, err := backend.Get(req.Backend)
	if err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
		result.DurationMS = time.Since(start).Milliseconds()
		return result, nil
	}

	if !b.IsAvailable() {
		result.Error = fmt.Sprintf("backend %q is not available", req.Backend)
		result.ExitCode = 1
		result.DurationMS = time.Since(start).Milliseconds()
		return result, nil
	}

	// Get model from config if not specified
	model := req.Model
	if model == "" {
		cfg := config.Get()
		if bcfg, ok := cfg.Backends[req.Backend]; ok {
			model = bcfg.Model
		}
	}

	// Build unified options
	opts := &backend.UnifiedOptions{
		WorkDir:      req.WorkDir,
		Model:        model,
		ApprovalMode: backend.ApprovalMode(req.ApprovalMode),
		SandboxMode:  backend.SandboxMode(req.SandboxMode),
		OutputFormat: backend.OutputFormat(req.OutputFormat),
		MaxTokens:    req.MaxTokens,
		MaxTurns:     req.MaxTurns,
		SystemPrompt: req.SystemPrompt,
		Verbose:      req.Verbose,
		DryRun:       req.DryRun,
		ExtraFlags:   req.Extra,
	}

	// Create session
	sess, sessErr := session.NewSession(req.Backend, req.WorkDir)
	if sessErr == nil {
		sess.SetModel(model)
		sess.InitialPrompt = req.Prompt
		sess.SetStatus(session.StatusActive)
		sess.AddTag("api")
		for k, v := range req.Metadata {
			sess.SetMetadata(k, v)
		}
		if err := e.store.Save(sess); err == nil {
			result.SessionID = sess.ID
		}
	}

	// Build command
	execCmd := b.BuildCommandUnified(req.Prompt, opts)

	if req.DryRun {
		result.Output = fmt.Sprintf("Would execute: %s %v", execCmd.Path, execCmd.Args[1:])
		result.ExitCode = 0
		result.DurationMS = time.Since(start).Milliseconds()
		return result, nil
	}

	// Execute with output capture
	var outputBuf bytes.Buffer
	exec := executor.New()
	exec.Stdin = nil
	exec.Stdout = &outputBuf
	exec.Stderr = &outputBuf

	exitCode, execErr := exec.RunSimple(execCmd)
	if execErr != nil {
		result.Error = execErr.Error()
	}

	result.ExitCode = exitCode
	result.Output = outputBuf.String()
	result.DurationMS = time.Since(start).Milliseconds()

	// Update session
	if sess != nil {
		sess.IncrementTurn()
		if exitCode == 0 {
			sess.Complete()
		} else {
			sess.SetError(result.Error)
		}
		_ = e.store.Save(sess)
	}

	return result, nil
}

// ParallelRequest represents a parallel execution request.
type ParallelRequest struct {
	Tasks       []PromptRequest `json:"tasks"`
	MaxParallel int             `json:"max_parallel,omitempty"`
	FailFast    bool            `json:"fail_fast,omitempty"`
	DryRun      bool            `json:"dry_run,omitempty"`
}

// ParallelResult represents the result of parallel execution.
type ParallelResult struct {
	TotalTasks    int            `json:"total_tasks"`
	Completed     int            `json:"completed"`
	Failed        int            `json:"failed"`
	TotalDuration int64          `json:"total_duration_ms"`
	Results       []PromptResult `json:"results"`
}

// ExecuteParallel executes multiple prompts in parallel.
func (e *Executor) ExecuteParallel(ctx context.Context, req *ParallelRequest) (*ParallelResult, error) {
	start := time.Now()

	maxP := req.MaxParallel
	if maxP <= 0 {
		cfg := config.Get()
		if cfg.Parallel.MaxWorkers > 0 {
			maxP = cfg.Parallel.MaxWorkers
		} else {
			maxP = 3
		}
	}

	result := &ParallelResult{
		TotalTasks: len(req.Tasks),
		Results:    make([]PromptResult, len(req.Tasks)),
	}

	sem := make(chan struct{}, maxP)
	var wg sync.WaitGroup
	var mu sync.Mutex

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, task := range req.Tasks {
		wg.Add(1)
		go func(idx int, t PromptRequest) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				mu.Lock()
				result.Results[idx] = PromptResult{
					Backend:  t.Backend,
					ExitCode: -1,
					Error:    "canceled",
				}
				mu.Unlock()
				return
			default:
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			// Apply request-level DryRun to each task
			if req.DryRun {
				t.DryRun = true
			}

			res, _ := e.ExecutePrompt(ctx, &t)

			mu.Lock()
			result.Results[idx] = *res
			if res.ExitCode == 0 && res.Error == "" {
				result.Completed++
			} else {
				result.Failed++
			}
			mu.Unlock()

			if req.FailFast && res.ExitCode != 0 {
				cancel()
			}
		}(i, task)
	}

	wg.Wait()
	result.TotalDuration = time.Since(start).Milliseconds()

	return result, nil
}

// ChainStep represents a step in a chain execution.
type ChainStep struct {
	Backend      string `json:"backend"`
	Prompt       string `json:"prompt"`
	Model        string `json:"model,omitempty"`
	WorkDir      string `json:"workdir,omitempty"`
	ApprovalMode string `json:"approval_mode,omitempty"`
	SandboxMode  string `json:"sandbox_mode,omitempty"`
	MaxTurns     int    `json:"max_turns,omitempty"`
	Name         string `json:"name,omitempty"`
}

// ChainRequest represents a chain execution request.
type ChainRequest struct {
	Steps          []ChainStep `json:"steps"`
	StopOnFailure  bool        `json:"stop_on_failure,omitempty"`
	PassSessionID  bool        `json:"pass_session_id,omitempty"`
	PassWorkingDir bool        `json:"pass_working_dir,omitempty"`
	DryRun         bool        `json:"dry_run,omitempty"`
}

// ChainStepResult represents the result of a chain step.
type ChainStepResult struct {
	Step       int    `json:"step"`
	Name       string `json:"name,omitempty"`
	Backend    string `json:"backend"`
	ExitCode   int    `json:"exit_code"`
	Error      string `json:"error,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	DurationMS int64  `json:"duration_ms"`
	Output     string `json:"output,omitempty"`
}

// ChainResult represents the result of chain execution.
type ChainResult struct {
	TotalSteps     int               `json:"total_steps"`
	CompletedSteps int               `json:"completed_steps"`
	FailedStep     int               `json:"failed_step,omitempty"`
	TotalDuration  int64             `json:"total_duration_ms"`
	Results        []ChainStepResult `json:"results"`
}

// ExecuteChain executes steps in sequence.
func (e *Executor) ExecuteChain(ctx context.Context, req *ChainRequest) (*ChainResult, error) {
	start := time.Now()

	result := &ChainResult{
		TotalSteps: len(req.Steps),
		Results:    make([]ChainStepResult, 0, len(req.Steps)),
	}

	var previousSessionID string
	var previousWorkDir string

	for i, step := range req.Steps {
		select {
		case <-ctx.Done():
			result.TotalDuration = time.Since(start).Milliseconds()
			return result, ctx.Err()
		default:
		}

		stepStart := time.Now()
		stepResult := ChainStepResult{
			Step:    i + 1,
			Name:    step.Name,
			Backend: step.Backend,
		}

		// Process prompt with placeholders
		prompt := step.Prompt
		if previousSessionID != "" {
			prompt = replacePlaceholder(prompt, "{{previous}}", previousSessionID)
			prompt = replacePlaceholder(prompt, "{{session}}", previousSessionID)
		}

		// Determine working directory
		workDir := step.WorkDir
		if workDir == "" && req.PassWorkingDir && previousWorkDir != "" {
			workDir = previousWorkDir
		}

		promptReq := &PromptRequest{
			Backend:      step.Backend,
			Prompt:       prompt,
			Model:        step.Model,
			WorkDir:      workDir,
			ApprovalMode: step.ApprovalMode,
			SandboxMode:  step.SandboxMode,
			MaxTurns:     step.MaxTurns,
			DryRun:       req.DryRun,
		}

		res, _ := e.ExecutePrompt(ctx, promptReq)

		stepResult.ExitCode = res.ExitCode
		stepResult.Error = res.Error
		stepResult.SessionID = res.SessionID
		stepResult.Output = res.Output
		stepResult.DurationMS = time.Since(stepStart).Milliseconds()

		result.Results = append(result.Results, stepResult)

		if res.ExitCode == 0 && res.Error == "" {
			result.CompletedSteps++
		} else {
			result.FailedStep = i + 1
			if req.StopOnFailure {
				break
			}
		}

		previousSessionID = res.SessionID
		previousWorkDir = workDir
	}

	result.TotalDuration = time.Since(start).Milliseconds()
	return result, nil
}

// CompareRequest represents a compare execution request.
type CompareRequest struct {
	Backends   []string `json:"backends"`
	Prompt     string   `json:"prompt"`
	Model      string   `json:"model,omitempty"`
	WorkDir    string   `json:"workdir,omitempty"`
	Sequential bool     `json:"sequential,omitempty"`
	DryRun     bool     `json:"dry_run,omitempty"`
}

// CompareBackendResult represents the result from one backend in comparison.
type CompareBackendResult struct {
	Backend    string `json:"backend"`
	Model      string `json:"model,omitempty"`
	ExitCode   int    `json:"exit_code"`
	Error      string `json:"error,omitempty"`
	DurationMS int64  `json:"duration_ms"`
	SessionID  string `json:"session_id,omitempty"`
	Output     string `json:"output,omitempty"`
}

// CompareResult represents the result of comparison execution.
type CompareResult struct {
	Prompt        string                 `json:"prompt"`
	Backends      []string               `json:"backends"`
	Results       []CompareBackendResult `json:"results"`
	TotalDuration int64                  `json:"total_duration_ms"`
}

// ExecuteCompare runs the same prompt on multiple backends for comparison.
func (e *Executor) ExecuteCompare(ctx context.Context, req *CompareRequest) (*CompareResult, error) {
	start := time.Now()

	result := &CompareResult{
		Prompt:   req.Prompt,
		Backends: req.Backends,
		Results:  make([]CompareBackendResult, len(req.Backends)),
	}

	if req.Sequential {
		for i, backendName := range req.Backends {
			result.Results[i] = e.runCompareBackend(ctx, backendName, req)
		}
	} else {
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i, backendName := range req.Backends {
			wg.Add(1)
			go func(idx int, bn string) {
				defer wg.Done()
				res := e.runCompareBackend(ctx, bn, req)
				mu.Lock()
				result.Results[idx] = res
				mu.Unlock()
			}(i, backendName)
		}

		wg.Wait()
	}

	result.TotalDuration = time.Since(start).Milliseconds()
	return result, nil
}

func (e *Executor) runCompareBackend(ctx context.Context, backendName string, req *CompareRequest) CompareBackendResult {
	start := time.Now()
	result := CompareBackendResult{
		Backend: backendName,
		Model:   req.Model,
	}

	promptReq := &PromptRequest{
		Backend: backendName,
		Prompt:  req.Prompt,
		Model:   req.Model,
		WorkDir: req.WorkDir,
		DryRun:  req.DryRun,
	}

	res, _ := e.ExecutePrompt(ctx, promptReq)

	result.ExitCode = res.ExitCode
	result.Error = res.Error
	result.SessionID = res.SessionID
	result.Output = res.Output
	result.DurationMS = time.Since(start).Milliseconds()

	return result
}

// SessionInfo represents session information for API responses.
type SessionInfo struct {
	ID            string              `json:"id"`
	Backend       string              `json:"backend"`
	CreatedAt     time.Time           `json:"created_at"`
	LastUsed      time.Time           `json:"last_used"`
	WorkingDir    string              `json:"working_dir,omitempty"`
	Model         string              `json:"model,omitempty"`
	InitialPrompt string              `json:"initial_prompt,omitempty"`
	Status        string              `json:"status,omitempty"`
	TurnCount     int                 `json:"turn_count,omitempty"`
	TokenUsage    *session.TokenUsage `json:"token_usage,omitempty"`
	Tags          []string            `json:"tags,omitempty"`
	Title         string              `json:"title,omitempty"`
}

// ListSessions returns all sessions.
func (e *Executor) ListSessions(ctx context.Context) ([]SessionInfo, error) {
	sessions, err := e.store.List()
	if err != nil {
		return nil, err
	}

	result := make([]SessionInfo, len(sessions))
	for i, s := range sessions {
		result[i] = sessionToInfo(s)
	}

	return result, nil
}

// GetSession returns a session by ID.
func (e *Executor) GetSession(ctx context.Context, id string) (*SessionInfo, error) {
	s, err := e.store.GetByPrefix(id)
	if err != nil {
		return nil, err
	}

	info := sessionToInfo(s)
	return &info, nil
}

// DeleteSession deletes a session by ID.
func (e *Executor) DeleteSession(ctx context.Context, id string) error {
	s, err := e.store.GetByPrefix(id)
	if err != nil {
		return err
	}
	return e.store.Delete(s.ID)
}

// BackendInfo represents backend information for API responses.
type BackendInfo struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// ListBackends returns all registered backends.
func (e *Executor) ListBackends(ctx context.Context) []BackendInfo {
	names := backend.List()
	result := make([]BackendInfo, len(names))

	for i, name := range names {
		b, _ := backend.Get(name)
		available := false
		if b != nil {
			available = b.IsAvailable()
		}
		result[i] = BackendInfo{
			Name:      name,
			Available: available,
		}
	}

	return result
}

func sessionToInfo(s *session.Session) SessionInfo {
	return SessionInfo{
		ID:            s.ID,
		Backend:       s.Backend,
		CreatedAt:     s.CreatedAt,
		LastUsed:      s.LastUsed,
		WorkingDir:    s.WorkingDir,
		Model:         s.Model,
		InitialPrompt: s.InitialPrompt,
		Status:        string(s.Status),
		TurnCount:     s.TurnCount,
		TokenUsage:    s.TokenUsage,
		Tags:          s.Tags,
		Title:         s.Title,
	}
}

func replacePlaceholder(s, old, replacement string) string {
	return strings.ReplaceAll(s, old, replacement)
}

// Ensure io.Writer is used to avoid import errors
var _ io.Writer = (*bytes.Buffer)(nil)
