// Package service provides the execution layer for API handlers.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/metrics"
	"github.com/signalridge/clinvoker/internal/output"
	"github.com/signalridge/clinvoker/internal/session"
)

// Default values for executor configuration.
const (
	// DefaultMaxParallelWorkers is the default number of parallel workers when not configured.
	DefaultMaxParallelWorkers = 3
)

// defaultBlockedWorkDirPrefixes contains paths that should never be used as working directories.
// These are applied even if AllowedWorkDirPrefixes is configured.
var defaultBlockedWorkDirPrefixes = []string{
	"/etc",
	"/var/run",
	"/var/lock",
	"/root",
	"/sys",
	"/proc",
	"/dev",
	"/usr/bin",
	"/usr/sbin",
	"/bin",
	"/sbin",
	"/boot",
	"/lib",
	"/lib64",
	"/usr/lib",
	"/usr/lib64",
}

// hasPathPrefix checks if path starts with prefix using proper path boundary checks.
// This prevents "/etcfoo" from matching prefix "/etc".
//
// Returns true if:
//   - path equals prefix exactly (e.g., "/etc" matches "/etc")
//   - path is a child of prefix (e.g., "/etc/passwd" matches "/etc")
//
// Returns false if:
//   - path merely starts with prefix string (e.g., "/etcfoo" does NOT match "/etc")
//   - prefix has trailing slash but path equals the prefix without slash
//     (e.g., "/etc" does NOT match "/etc/" because "/etc" is not under "/etc/")
//
// Note: When configuring allowed/blocked prefixes, use paths WITHOUT trailing slashes
// to match both the directory itself and its children (e.g., use "/etc" not "/etc/").
func hasPathPrefix(path, prefix string) bool {
	if path == prefix {
		return true
	}
	// Ensure prefix ends with separator for proper boundary check
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix = prefix + string(filepath.Separator)
	}
	return strings.HasPrefix(path, prefix)
}

// validateWorkDir validates that the working directory is safe and exists.
func validateWorkDir(workDir string) error {
	return validateWorkDirWithConfig(workDir, nil, nil)
}

// validateWorkDirWithConfig validates the working directory with configurable restrictions.
// allowedPrefixes: if non-empty, workDir must start with one of these prefixes.
// blockedPrefixes: workDir must not start with any of these prefixes (defaults applied if nil).
func validateWorkDirWithConfig(workDir string, allowedPrefixes, blockedPrefixes []string) error {
	if workDir == "" {
		return nil // Empty workDir is allowed, will use current directory
	}

	// Must be absolute path
	if !filepath.IsAbs(workDir) {
		return fmt.Errorf("invalid work directory: must be an absolute path")
	}

	// Clean the path to resolve any ".." or "." components
	cleanPath := filepath.Clean(workDir)

	// Verify the cleaned path is still absolute (sanity check)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("invalid work directory: path traversal not allowed")
	}

	// Resolve symbolic links to get the real path
	realPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("work directory does not exist: %s", workDir)
		}
		return fmt.Errorf("cannot resolve work directory: %w", err)
	}

	// Verify the resolved path matches expectations (no symlink escape)
	// The real path must also be absolute
	if !filepath.IsAbs(realPath) {
		return fmt.Errorf("invalid work directory: resolved path is not absolute")
	}

	// Apply blocked prefixes (use defaults if not specified)
	blocked := blockedPrefixes
	if len(blocked) == 0 {
		blocked = defaultBlockedWorkDirPrefixes
	}
	for _, prefix := range blocked {
		if hasPathPrefix(realPath, prefix) {
			return fmt.Errorf("work directory not allowed: %s (blocked path)", workDir)
		}
	}

	// Apply allowed prefixes if specified
	if len(allowedPrefixes) > 0 {
		allowed := false
		for _, prefix := range allowedPrefixes {
			if hasPathPrefix(realPath, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("work directory not allowed: %s (not in allowed paths)", workDir)
		}
	}

	// Check if directory exists and is actually a directory
	info, err := os.Stat(realPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("work directory does not exist: %s", workDir)
		}
		return fmt.Errorf("cannot access work directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("work directory is not a directory: %s", workDir)
	}

	return nil
}

// ValidateWorkDirFromConfig validates workDir using configuration from config.Get().
func ValidateWorkDirFromConfig(workDir string) error {
	cfg := config.Get()
	return validateWorkDirWithConfig(
		workDir,
		cfg.Server.AllowedWorkDirPrefixes,
		cfg.Server.BlockedWorkDirPrefixes,
	)
}

// Executor handles the execution of AI backend commands.
type Executor struct {
	store  *session.Store
	logger *slog.Logger
}

// NewExecutor creates a new executor.
func NewExecutor() *Executor {
	store := session.NewStore()
	e := &Executor{
		store:  store,
		logger: slog.Default(),
	}
	// Initialize active sessions metric if enabled
	e.updateActiveSessionsMetric()
	return e
}

// NewExecutorWithLogger creates a new executor with a custom logger.
func NewExecutorWithLogger(logger *slog.Logger) *Executor {
	if logger == nil {
		logger = slog.Default()
	}
	store := session.NewStore()
	e := &Executor{
		store:  store,
		logger: logger,
	}
	// Initialize active sessions metric if enabled
	e.updateActiveSessionsMetric()
	return e
}

// updateActiveSessionsMetric updates the active sessions gauge metric.
func (e *Executor) updateActiveSessionsMetric() {
	if !config.Get().Server.MetricsEnabled {
		return
	}
	// Count active sessions (exclude completed/error status)
	sessions, err := e.store.List()
	if err != nil {
		e.logger.Warn("failed to count sessions for metrics", "error", err)
		return
	}
	var activeCount float64
	for _, s := range sessions {
		if s.Status == session.StatusActive || s.Status == session.StatusPaused {
			activeCount++
		}
	}
	metrics.SetActiveSessions(activeCount)
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
	Ephemeral    bool              `json:"ephemeral,omitempty"`
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
	runner := NewStatefulRunner(e.store, e.logger)
	return runner.ExecutePrompt(ctx, req)
}

// StreamPrompt executes a single prompt in streaming mode.
func (e *Executor) StreamPrompt(ctx context.Context, req *PromptRequest, onEvent func(*output.UnifiedEvent) error) (*StreamResult, error) {
	return StreamPrompt(ctx, req, e.store, e.logger, false, onEvent)
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
			maxP = DefaultMaxParallelWorkers
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

			// Acquire semaphore with context cancellation support
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
			case sem <- struct{}{}:
				// Acquired semaphore
			}
			defer func() { <-sem }()

			// Apply request-level DryRun to each task
			if req.DryRun {
				t.DryRun = true
			}
			// Parallel execution is always ephemeral (clean mode).
			t.Ephemeral = true

			res, err := e.ExecutePrompt(ctx, &t)
			if err != nil {
				e.logger.Warn("prompt execution returned error", "task_index", idx, "backend", t.Backend, "error", err)
			}

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
	Backend      string   `json:"backend"`
	Prompt       string   `json:"prompt"`
	Model        string   `json:"model,omitempty"`
	WorkDir      string   `json:"workdir,omitempty"`
	ApprovalMode string   `json:"approval_mode,omitempty"`
	SandboxMode  string   `json:"sandbox_mode,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
	MaxTurns     int      `json:"max_turns,omitempty"`
	SystemPrompt string   `json:"system_prompt,omitempty"`
	Verbose      bool     `json:"verbose,omitempty"`
	Extra        []string `json:"extra,omitempty"`
	Name         string   `json:"name,omitempty"`
}

// ChainRequest represents a chain execution request.
type ChainRequest struct {
	Steps          []ChainStep `json:"steps"`
	StopOnFailure  bool        `json:"stop_on_failure,omitempty"`
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

	var previousWorkDir string
	var previousOutput string
	var hasPreviousOutput bool

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
		if hasPreviousOutput {
			prompt = replacePlaceholder(prompt, "{{previous}}", previousOutput)
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
			MaxTokens:    step.MaxTokens,
			MaxTurns:     step.MaxTurns,
			SystemPrompt: step.SystemPrompt,
			Verbose:      step.Verbose,
			DryRun:       req.DryRun,
			Ephemeral:    true,
			Extra:        step.Extra,
		}

		res, err := e.ExecutePrompt(ctx, promptReq)
		if err != nil {
			e.logger.Warn("chain step execution returned error", "step", i+1, "name", step.Name, "backend", step.Backend, "error", err)
		}

		stepResult.ExitCode = res.ExitCode
		stepResult.Error = res.Error
		stepResult.SessionID = ""
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

		previousWorkDir = workDir
		previousOutput = res.Output
		hasPreviousOutput = true
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
		// Compare is always ephemeral (clean mode).
		Ephemeral: true,
	}

	res, err := e.ExecutePrompt(ctx, promptReq)
	if err != nil {
		e.logger.Warn("compare backend execution returned error", "backend", backendName, "error", err)
	}

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

// SessionListOptions contains options for listing sessions.
type SessionListOptions struct {
	Backend string
	Status  string
	Limit   int
	Offset  int
}

// SessionListResult contains paginated session results.
type SessionListResult struct {
	Sessions []SessionInfo `json:"sessions"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
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

// ListSessionsPaginated returns sessions with pagination support.
func (e *Executor) ListSessionsPaginated(ctx context.Context, opts *SessionListOptions) (*SessionListResult, error) {
	if opts == nil {
		opts = &SessionListOptions{}
	}

	filter := &session.ListFilter{
		Backend: opts.Backend,
		Status:  session.SessionStatus(opts.Status),
		Limit:   opts.Limit,
		Offset:  opts.Offset,
	}

	listResult, err := e.store.ListPaginated(filter)
	if err != nil {
		return nil, err
	}

	sessions := make([]SessionInfo, len(listResult.Sessions))
	for i, s := range listResult.Sessions {
		sessions[i] = sessionToInfo(s)
	}

	return &SessionListResult{
		Sessions: sessions,
		Total:    listResult.Total,
		Limit:    listResult.Limit,
		Offset:   listResult.Offset,
	}, nil
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
// Uses cached availability checks for better performance during frequent health checks.
func (e *Executor) ListBackends(ctx context.Context) []BackendInfo {
	names := backend.List()
	result := make([]BackendInfo, len(names))

	for i, name := range names {
		result[i] = BackendInfo{
			Name:      name,
			Available: backend.IsAvailableCached(name),
		}
	}

	return result
}

// SessionStoreHealth represents the health status of the session store.
type SessionStoreHealth struct {
	Available    bool
	SessionCount int
	Error        string
}

// GetSessionStoreHealth returns the health status of the session store.
func (e *Executor) GetSessionStoreHealth(ctx context.Context) SessionStoreHealth {
	health := SessionStoreHealth{Available: true}

	sessions, err := e.store.List()
	if err != nil {
		health.Available = false
		health.Error = err.Error()
		return health
	}

	health.SessionCount = len(sessions)
	return health
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
