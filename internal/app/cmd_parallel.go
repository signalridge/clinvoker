package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/executor"
	"github.com/signalridge/clinvoker/internal/session"
)

// parallelCmd runs multiple tasks in parallel.
var parallelCmd = &cobra.Command{
	Use:   "parallel",
	Short: "Run multiple AI tasks in parallel",
	Long: `Run multiple AI tasks in parallel across different backends.

Read tasks from stdin or a file:
  cat tasks.json | clinvk parallel
  clinvk parallel --file tasks.json

Basic task format (JSON):
  {
    "tasks": [
      {"backend": "claude", "prompt": "review auth module"},
      {"backend": "codex", "prompt": "add logging to api"},
      {"backend": "gemini", "prompt": "generate tests for utils"}
    ],
    "max_parallel": 3
  }

Extended task format with per-task options:
  {
    "tasks": [
      {
        "backend": "claude",
        "prompt": "review auth module",
        "id": "task-1",
        "name": "Auth Review",
        "model": "claude-opus-4-5-20251101",
        "approval_mode": "auto",
        "sandbox_mode": "workspace",
        "max_turns": 10
      }
    ],
    "max_parallel": 3,
    "fail_fast": true
  }`,
	RunE: runParallel,
}

var (
	parallelFile     string
	maxParallel      int
	parallelFailFast bool
	parallelJSON     bool
	parallelQuiet    bool
)

func init() {
	parallelCmd.Flags().StringVarP(&parallelFile, "file", "f", "", "file containing task definitions")
	parallelCmd.Flags().IntVar(&maxParallel, "max-parallel", defaultMaxParallel, "maximum number of parallel tasks")
	parallelCmd.Flags().BoolVar(&parallelFailFast, "fail-fast", false, "stop all tasks on first failure")
	parallelCmd.Flags().BoolVar(&parallelJSON, "json", false, "output results as JSON")
	parallelCmd.Flags().BoolVarP(&parallelQuiet, "quiet", "q", false, "suppress task output (show only results)")
}

// ParallelTasks represents the input format for parallel execution.
type ParallelTasks struct {
	Tasks       []ParallelTask `json:"tasks"`
	MaxParallel int            `json:"max_parallel,omitempty"`
	FailFast    bool           `json:"fail_fast,omitempty"`
	OutputDir   string         `json:"output_dir,omitempty"`
}

// ParallelTask represents a single task in parallel execution.
type ParallelTask struct {
	// Required fields
	Backend string `json:"backend"`
	Prompt  string `json:"prompt"`

	// Basic options
	WorkDir string   `json:"workdir,omitempty"`
	Model   string   `json:"model,omitempty"`
	Extra   []string `json:"extra,omitempty"`

	// Unified options
	ApprovalMode string `json:"approval_mode,omitempty"` // default, auto, none, always
	SandboxMode  string `json:"sandbox_mode,omitempty"`  // default, read-only, workspace, full
	OutputFormat string `json:"output_format,omitempty"` // default, text, json, stream-json
	MaxTokens    int    `json:"max_tokens,omitempty"`
	MaxTurns     int    `json:"max_turns,omitempty"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	Verbose      bool   `json:"verbose,omitempty"`
	DryRun       bool   `json:"dry_run,omitempty"`

	// Task metadata
	ID   string            `json:"id,omitempty"`
	Name string            `json:"name,omitempty"`
	Tags []string          `json:"tags,omitempty"`
	Meta map[string]string `json:"meta,omitempty"`
}

// TaskResult represents the result of a parallel task.
type TaskResult struct {
	Index     int       `json:"index"`
	TaskID    string    `json:"task_id,omitempty"`
	TaskName  string    `json:"task_name,omitempty"`
	Backend   string    `json:"backend"`
	ExitCode  int       `json:"exit_code"`
	Error     string    `json:"error,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  float64   `json:"duration_seconds"`
	SessionID string    `json:"session_id,omitempty"`
}

// ParallelResults represents the aggregated results of parallel execution.
type ParallelResults struct {
	TotalTasks    int          `json:"total_tasks"`
	Completed     int          `json:"completed"`
	Failed        int          `json:"failed"`
	TotalDuration float64      `json:"total_duration_seconds"`
	Results       []TaskResult `json:"results"`
	StartTime     time.Time    `json:"start_time"`
	EndTime       time.Time    `json:"end_time"`
}

// parallelContext holds shared state for parallel execution.
type parallelContext struct {
	store    *session.Store
	cfg      *config.Config
	failFast bool
	quiet    bool
}

func runParallel(cmd *cobra.Command, args []string) error {
	tasks, err := parseParallelTasks()
	if err != nil {
		return err
	}

	maxP, failFast := resolveParallelConfig(tasks)

	if !parallelQuiet && !parallelJSON {
		printParallelHeader(len(tasks.Tasks), maxP, failFast)
	}

	results := executeParallelTasks(tasks, maxP, failFast)
	outputParallelResults(results, tasks)

	if results.Failed > 0 {
		return fmt.Errorf("%d task(s) failed", results.Failed)
	}
	return nil
}

// parseParallelTasks reads and parses the parallel tasks definition.
func parseParallelTasks() (*ParallelTasks, error) {
	input, err := readInputFromFileOrStdin(parallelFile)
	if err != nil {
		return nil, err
	}

	var tasks ParallelTasks
	if err := json.Unmarshal(input, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	if len(tasks.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks provided")
	}

	return &tasks, nil
}

// resolveParallelConfig determines max parallel and fail-fast settings.
func resolveParallelConfig(tasks *ParallelTasks) (int, bool) {
	cfg := config.Get()

	// Determine max parallel
	maxP := maxParallel
	if tasks.MaxParallel > 0 {
		maxP = tasks.MaxParallel
	}
	if maxP == 0 && cfg.Parallel.MaxWorkers > 0 {
		maxP = cfg.Parallel.MaxWorkers
	}
	if maxP == 0 {
		maxP = defaultMaxParallel
	}

	// Determine fail-fast
	failFast := parallelFailFast || tasks.FailFast || cfg.Parallel.FailFast

	return maxP, failFast
}

// printParallelHeader prints the parallel execution header.
func printParallelHeader(taskCount, maxP int, failFast bool) {
	fmt.Printf("Running %d tasks (max %d parallel", taskCount, maxP)
	if failFast {
		fmt.Print(", fail-fast")
	}
	fmt.Println(")...")
	fmt.Println()
}

// executeParallelTasks executes all tasks in parallel.
func executeParallelTasks(tasks *ParallelTasks, maxP int, failFast bool) *ParallelResults {
	results := &ParallelResults{
		TotalTasks: len(tasks.Tasks),
		Results:    make([]TaskResult, len(tasks.Tasks)),
		StartTime:  time.Now(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pCtx := &parallelContext{
		store:    session.NewStore(),
		cfg:      config.Get(),
		failFast: failFast,
		quiet:    parallelQuiet,
	}

	sem := make(chan struct{}, maxP)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range tasks.Tasks {
		wg.Add(1)
		go func(idx int, t *ParallelTask) {
			defer wg.Done()

			// Acquire semaphore first, then check context
			// This ensures we don't start work if canceled
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				mu.Lock()
				results.Results[idx] = createCanceledResult(idx, t)
				mu.Unlock()
				return
			}

			// Check again after acquiring semaphore
			select {
			case <-ctx.Done():
				mu.Lock()
				results.Results[idx] = createCanceledResult(idx, t)
				mu.Unlock()
				return
			default:
			}

			result := executeParallelTask(idx, t, pCtx)

			mu.Lock()
			results.Results[idx] = result
			if result.ExitCode == 0 && result.Error == "" {
				results.Completed++
			} else {
				results.Failed++
			}
			mu.Unlock()

			if failFast && result.ExitCode != 0 {
				cancel()
			}
		}(i, &tasks.Tasks[i])
	}

	wg.Wait()
	results.EndTime = time.Now()
	results.TotalDuration = results.EndTime.Sub(results.StartTime).Seconds()

	return results
}

// createCanceledResult creates a result for a canceled task.
func createCanceledResult(idx int, t *ParallelTask) TaskResult {
	return TaskResult{
		Index:    idx,
		TaskID:   t.ID,
		TaskName: t.Name,
		Backend:  t.Backend,
		ExitCode: -1,
		Error:    "canceled (fail-fast)",
	}
}

// executeParallelTask executes a single parallel task.
func executeParallelTask(idx int, t *ParallelTask, pCtx *parallelContext) TaskResult {
	startTime := time.Now()
	result := TaskResult{
		Index:     idx,
		TaskID:    t.ID,
		TaskName:  t.Name,
		Backend:   t.Backend,
		StartTime: startTime,
	}

	// Get and validate backend
	b, err := getBackendOrError(t.Backend)
	if err != nil {
		failTaskResult(&result, startTime, err.Error())
		return result
	}

	// Build unified options
	opts := buildParallelTaskOptions(t)

	// Create and save session
	tags := append([]string{"parallel"}, t.Tags...)
	sess := createAndSaveSession(pCtx.store, t.Backend, t.WorkDir, t.Model, t.Prompt, tags, t.Name, pCtx.quiet)
	if sess != nil {
		result.SessionID = sess.ID
	}

	// Build command
	execCmd := b.BuildCommandUnified(t.Prompt, opts)

	if t.DryRun || dryRun {
		if !pCtx.quiet {
			fmt.Printf("[%d] Would execute: %s %v\n", idx+1, execCmd.Path, execCmd.Args[1:])
		}
		result.ExitCode = 0
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime).Seconds()
		return result
	}

	// Execute
	exec := executor.New()
	exec.Stdin = nil // No stdin for parallel tasks
	if pCtx.quiet {
		exec.Stdout = io.Discard
		exec.Stderr = io.Discard
	}

	exitCode, execErr := exec.RunSimple(execCmd)
	if execErr != nil {
		result.Error = execErr.Error()
	}
	result.ExitCode = exitCode
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime).Seconds()

	// Update session
	updateSessionAfterExecution(pCtx.store, sess, exitCode, result.Error, pCtx.quiet)

	return result
}

// failTaskResult populates a failed task result.
func failTaskResult(result *TaskResult, startTime time.Time, errMsg string) {
	result.Error = errMsg
	result.ExitCode = 1
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime).Seconds()
}

// buildParallelTaskOptions builds unified options for a parallel task.
func buildParallelTaskOptions(t *ParallelTask) *backend.UnifiedOptions {
	return &backend.UnifiedOptions{
		WorkDir:      t.WorkDir,
		Model:        t.Model,
		ApprovalMode: backend.ApprovalMode(t.ApprovalMode),
		SandboxMode:  backend.SandboxMode(t.SandboxMode),
		OutputFormat: backend.OutputFormat(t.OutputFormat),
		MaxTokens:    t.MaxTokens,
		MaxTurns:     t.MaxTurns,
		SystemPrompt: t.SystemPrompt,
		Verbose:      t.Verbose,
		DryRun:       t.DryRun || dryRun,
		ExtraFlags:   t.Extra,
	}
}

// outputParallelResults outputs the parallel execution results.
func outputParallelResults(results *ParallelResults, tasks *ParallelTasks) {
	if parallelJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
		return
	}

	fmt.Println("\nResults:")
	fmt.Println(strings.Repeat("-", tableSeparatorWidth))
	fmt.Printf("%-4s %-12s %-8s %-10s %-10s %s\n", "#", "BACKEND", "STATUS", "DURATION", "SESSION", "TASK")
	fmt.Println(strings.Repeat("-", tableSeparatorWidth))

	for i := range results.Results {
		r := &results.Results[i]
		status := resolveTaskStatus(r)
		taskName := resolveTaskDisplayName(r, tasks)
		sessionID := "-"
		if r.SessionID != "" {
			sessionID = r.SessionID[:8]
		}

		fmt.Printf("%-4d %-12s %-8s %-10.2fs %-10s %s\n",
			r.Index+1, r.Backend, status, r.Duration, sessionID, taskName)

		if r.Error != "" && r.Error != "canceled (fail-fast)" {
			fmt.Printf("     Error: %s\n", r.Error)
		}
	}

	fmt.Println(strings.Repeat("-", tableSeparatorWidth))
	fmt.Printf("Total: %d tasks, %d completed, %d failed (%.2fs)\n",
		results.TotalTasks, results.Completed, results.Failed, results.TotalDuration)
}

// resolveTaskStatus determines the display status for a task result.
func resolveTaskStatus(r *TaskResult) string {
	if r.ExitCode == -1 {
		return "CANCELED"
	}
	if r.ExitCode != 0 || r.Error != "" {
		return "FAILED"
	}
	return "OK"
}

// resolveTaskDisplayName determines the display name for a task.
func resolveTaskDisplayName(r *TaskResult, tasks *ParallelTasks) string {
	name := r.TaskName
	if name == "" {
		name = tasks.Tasks[r.Index].Prompt
	}
	return truncateString(name, maxTaskNameLen)
}
