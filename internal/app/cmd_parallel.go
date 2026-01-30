package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/util"
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
	Index     int               `json:"index"`
	TaskID    string            `json:"task_id,omitempty"`
	TaskName  string            `json:"task_name,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	Backend   string            `json:"backend"`
	ExitCode  int               `json:"exit_code"`
	Error     string            `json:"error,omitempty"`
	Output    string            `json:"output,omitempty"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Duration  float64           `json:"duration_seconds"`
}

// parallelTaskOutput represents a persisted task output payload.
type parallelTaskOutput struct {
	Task   ParallelTask `json:"task"`
	Result TaskResult   `json:"result"`
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
	cfg      *config.Config
	ctx      context.Context
	failFast bool
	quiet    bool
}

func runParallel(cmd *cobra.Command, args []string) error {
	tasks, err := parseParallelTasks()
	if err != nil {
		return err
	}

	maxP, failFast := resolveParallelConfig(tasks)
	effectiveQuiet := parallelQuiet || parallelJSON

	if !effectiveQuiet {
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
		cfg:      config.Get(),
		ctx:      ctx,
		failFast: failFast,
		quiet:    parallelQuiet || parallelJSON,
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
		Tags:     cloneStringSlice(t.Tags),
		Meta:     cloneStringMap(t.Meta),
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
		Tags:      cloneStringSlice(t.Tags),
		Meta:      cloneStringMap(t.Meta),
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
	opts := buildParallelTaskOptions(t, pCtx.cfg)

	// Build command
	execCmd := b.BuildCommandUnified(t.Prompt, opts)
	execCmd = util.CommandWithContext(pCtx.ctx, execCmd)

	if opts.DryRun {
		if !pCtx.quiet {
			fmt.Printf("[%d] Would execute: %s %v\n", idx+1, execCmd.Path, execCmd.Args[1:])
		}
		result.ExitCode = 0
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime).Seconds()
		return result
	}

	// Execute with JSON output capture for proper error extraction
	captureResult, execErr := ExecuteAndCaptureWithJSON(b, execCmd)
	if execErr != nil && captureResult.Error == "" {
		if errors.Is(execErr, context.Canceled) {
			result.Error = "canceled (fail-fast)"
			result.ExitCode = -1
		} else {
			result.Error = execErr.Error()
		}
	} else if captureResult.Error != "" {
		result.Error = captureResult.Error
	}
	if result.ExitCode == 0 {
		result.ExitCode = captureResult.ExitCode
	}
	result.Output = captureResult.Content
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime).Seconds()

	// Print output if not in quiet mode
	if !pCtx.quiet && captureResult.Content != "" {
		fmt.Printf("[%d] %s\n", idx+1, captureResult.Content)
	}

	// Ensure ephemeral parallel runs remain clean on the backend.
	if opts.Ephemeral {
		cleanupBackendSession(t.Backend, captureResult.BackendSessionID)
	}

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
func buildParallelTaskOptions(t *ParallelTask, cfg *config.Config) *backend.UnifiedOptions {
	model := t.Model
	if model == "" && cfg != nil {
		if bcfg, ok := cfg.Backends[t.Backend]; ok {
			model = bcfg.Model
		}
	}

	opts := &backend.UnifiedOptions{
		WorkDir:      t.WorkDir,
		Model:        model,
		ApprovalMode: backend.ApprovalMode(t.ApprovalMode),
		SandboxMode:  backend.SandboxMode(t.SandboxMode),
		OutputFormat: backend.OutputJSON, // Force JSON for proper error capture
		MaxTokens:    t.MaxTokens,
		MaxTurns:     t.MaxTurns,
		SystemPrompt: t.SystemPrompt,
		Verbose:      t.Verbose,
		DryRun:       t.DryRun,
		ExtraFlags:   t.Extra,
		Ephemeral:    true, // Parallel tasks are always ephemeral
	}

	// Apply config defaults using shared util function
	effectiveDryRun := dryRun || (cfg != nil && cfg.UnifiedFlags.DryRun)
	util.ApplyUnifiedDefaults(opts, cfg, effectiveDryRun)
	util.ApplyBackendDefaults(opts, t.Backend, cfg)

	// Note: OutputFormat is intentionally kept as JSON for internal parsing
	// The user's requested format is used for final output presentation

	return opts
}

// outputParallelResults outputs the parallel execution results.
func outputParallelResults(results *ParallelResults, tasks *ParallelTasks) {
	cfg := config.Get()

	if tasks.OutputDir != "" {
		if err := persistParallelOutputs(results, tasks); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write parallel outputs: %v\n", err)
		}
	}

	if parallelJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
		return
	}

	// If aggregate output is disabled, just print results without table
	if !cfg.Parallel.AggregateOutput {
		for _, r := range results.Results {
			if r.Error != "" {
				fmt.Printf("Error [%d]: %s\n", r.Index+1, r.Error)
			} else if r.Output != "" {
				fmt.Printf("[%d] %s\n", r.Index+1, r.Output)
			}
		}
		return
	}

	fmt.Println("\nResults:")
	fmt.Println(strings.Repeat("-", tableSeparatorWidth))
	fmt.Printf("%-4s %-12s %-8s %-10s %s\n", "#", "BACKEND", "STATUS", "DURATION", "TASK")
	fmt.Println(strings.Repeat("-", tableSeparatorWidth))

	for i := range results.Results {
		r := &results.Results[i]
		status := resolveTaskStatus(r)
		taskName := resolveTaskDisplayName(r, tasks)

		fmt.Printf("%-4d %-12s %-8s %-10.2fs %s\n",
			r.Index+1, r.Backend, status, r.Duration, taskName)

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

func persistParallelOutputs(results *ParallelResults, tasks *ParallelTasks) error {
	if tasks.OutputDir == "" {
		return nil
	}

	if err := os.MkdirAll(tasks.OutputDir, 0o755); err != nil {
		return err
	}

	summaryPath := filepath.Join(tasks.OutputDir, "summary.json")
	if err := writeJSONFile(summaryPath, results); err != nil {
		return err
	}

	for i := range tasks.Tasks {
		output := parallelTaskOutput{
			Task:   tasks.Tasks[i],
			Result: results.Results[i],
		}
		filename := parallelOutputFilename(&tasks.Tasks[i], i)
		if err := writeJSONFile(filepath.Join(tasks.OutputDir, filename), output); err != nil {
			return err
		}
	}

	return nil
}

func parallelOutputFilename(task *ParallelTask, index int) string {
	base := task.ID
	if base == "" {
		base = task.Name
	}
	if base == "" {
		base = fmt.Sprintf("task-%d", index+1)
	}
	base = sanitizeFilename(base)
	if base == "" {
		base = fmt.Sprintf("task-%d", index+1)
	}
	return fmt.Sprintf("%03d_%s.json", index+1, base)
}

func writeJSONFile(path string, v any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		switch {
		case r <= 127 && (r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_'):
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}

	sanitized := strings.Trim(b.String(), "_")
	if len(sanitized) > 64 {
		sanitized = sanitized[:64]
	}
	return sanitized
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
