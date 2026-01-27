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
  cat tasks.json | clinvoker parallel
  clinvoker parallel --file tasks.json

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

func runParallel(cmd *cobra.Command, args []string) error {
	var input []byte
	var err error

	if parallelFile != "" {
		input, err = os.ReadFile(parallelFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("no input provided (use --file or pipe JSON to stdin)")
		}
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
	}

	var tasks ParallelTasks
	if err := json.Unmarshal(input, &tasks); err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
	}

	if len(tasks.Tasks) == 0 {
		return fmt.Errorf("no tasks provided")
	}

	// Determine max parallel from CLI flag or JSON config
	maxP := maxParallel
	if tasks.MaxParallel > 0 {
		maxP = tasks.MaxParallel
	}
	// Use config default if not specified
	cfg := config.Get()
	if maxP == 0 && cfg.Parallel.MaxWorkers > 0 {
		maxP = cfg.Parallel.MaxWorkers
	}
	if maxP == 0 {
		maxP = defaultMaxParallel
	}

	// Determine fail-fast from CLI flag or JSON config
	failFast := parallelFailFast || tasks.FailFast || cfg.Parallel.FailFast

	if !parallelQuiet && !parallelJSON {
		fmt.Printf("Running %d tasks (max %d parallel", len(tasks.Tasks), maxP)
		if failFast {
			fmt.Print(", fail-fast")
		}
		fmt.Println(")...")
		fmt.Println()
	}

	// Create session store for tracking
	store := session.NewStore()

	// Aggregate results
	aggregated := &ParallelResults{
		TotalTasks: len(tasks.Tasks),
		Results:    make([]TaskResult, len(tasks.Tasks)),
		StartTime:  time.Now(),
	}

	// Semaphore for limiting parallelism
	sem := make(chan struct{}, maxP)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Context for fail-fast cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, task := range tasks.Tasks {
		wg.Add(1)
		go func(idx int, t ParallelTask) {
			defer wg.Done()

			// Check if canceled (fail-fast)
			select {
			case <-ctx.Done():
				mu.Lock()
				aggregated.Results[idx] = TaskResult{
					Index:    idx,
					TaskID:   t.ID,
					TaskName: t.Name,
					Backend:  t.Backend,
					ExitCode: -1,
					Error:    "canceled (fail-fast)",
				}
				mu.Unlock()
				return
			default:
			}

			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			startTime := time.Now()
			result := TaskResult{
				Index:     idx,
				TaskID:    t.ID,
				TaskName:  t.Name,
				Backend:   t.Backend,
				StartTime: startTime,
			}

			// Get backend
			b, err := backend.Get(t.Backend)
			if err != nil {
				result.Error = err.Error()
				result.ExitCode = 1
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(startTime).Seconds()
				mu.Lock()
				aggregated.Results[idx] = result
				aggregated.Failed++
				mu.Unlock()
				if failFast {
					cancel()
				}
				return
			}

			if !b.IsAvailable() {
				result.Error = fmt.Sprintf("backend %q not available", t.Backend)
				result.ExitCode = 1
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(startTime).Seconds()
				mu.Lock()
				aggregated.Results[idx] = result
				aggregated.Failed++
				mu.Unlock()
				if failFast {
					cancel()
				}
				return
			}

			// Build unified options from task config
			unifiedOpts := &backend.UnifiedOptions{
				WorkDir:      t.WorkDir,
				Model:        t.Model,
				ApprovalMode: backend.ApprovalMode(t.ApprovalMode),
				SandboxMode:  backend.SandboxMode(t.SandboxMode),
				OutputFormat: backend.OutputFormat(t.OutputFormat),
				MaxTokens:    t.MaxTokens,
				MaxTurns:     t.MaxTurns,
				SystemPrompt: t.SystemPrompt,
				Verbose:      t.Verbose,
				DryRun:       t.DryRun || dryRun, // inherit global dry-run
				ExtraFlags:   t.Extra,
			}

			// Create session for this task
			sess, sessErr := session.NewSession(t.Backend, t.WorkDir)
			if sessErr != nil {
				if !parallelQuiet {
					fmt.Fprintf(os.Stderr, "Warning: failed to create session: %v\n", sessErr)
				}
			} else {
				sess.SetModel(t.Model)
				sess.InitialPrompt = t.Prompt
				sess.SetStatus(session.StatusActive)
				if len(t.Tags) > 0 {
					for _, tag := range t.Tags {
						sess.AddTag(tag)
					}
				}
				sess.AddTag("parallel")
				if t.Name != "" {
					sess.SetTitle(t.Name)
				}
				if err := store.Save(sess); err != nil && !parallelQuiet {
					fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
				}
				result.SessionID = sess.ID
			}

			// Build command using unified options
			execCmd := b.BuildCommandUnified(t.Prompt, unifiedOpts)

			if t.DryRun || dryRun {
				if !parallelQuiet {
					fmt.Printf("[%d] Would execute: %s %v\n", idx+1, execCmd.Path, execCmd.Args[1:])
				}
				result.ExitCode = 0
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(startTime).Seconds()
				mu.Lock()
				aggregated.Results[idx] = result
				aggregated.Completed++
				mu.Unlock()
				return
			}

			// Execute
			exec := executor.New()
			exec.Stdin = nil // No stdin for parallel tasks
			if parallelQuiet {
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

			// Update session (if created successfully)
			if sess != nil {
				sess.IncrementTurn()
				if exitCode == 0 {
					sess.Complete()
				} else {
					sess.SetError(result.Error)
				}
				if err := store.Save(sess); err != nil && !parallelQuiet {
					fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", err)
				}
			}

			mu.Lock()
			aggregated.Results[idx] = result
			if exitCode == 0 {
				aggregated.Completed++
			} else {
				aggregated.Failed++
			}
			mu.Unlock()

			if failFast && exitCode != 0 {
				cancel()
			}
		}(i, task)
	}

	wg.Wait()

	aggregated.EndTime = time.Now()
	aggregated.TotalDuration = aggregated.EndTime.Sub(aggregated.StartTime).Seconds()

	// Output results
	if parallelJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(aggregated); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		// Print results
		fmt.Println("\nResults:")
		fmt.Println(strings.Repeat("-", tableSeparatorWidth))
		fmt.Printf("%-4s %-12s %-8s %-10s %-10s %s\n", "#", "BACKEND", "STATUS", "DURATION", "SESSION", "TASK")
		fmt.Println(strings.Repeat("-", tableSeparatorWidth))

		for _, r := range aggregated.Results {
			status := "OK"
			if r.ExitCode != 0 || r.Error != "" {
				status = "FAILED"
			}
			if r.ExitCode == -1 {
				status = "CANCELED"
			}

			taskName := r.TaskName
			if taskName == "" {
				taskName = tasks.Tasks[r.Index].Prompt
			}
			if len(taskName) > maxTaskNameLen {
				taskName = taskName[:maxTaskNameLen-3] + "..."
			}

			sessionID := "-"
			if r.SessionID != "" {
				sessionID = r.SessionID[:8]
			}

			fmt.Printf("%-4d %-12s %-8s %-10.2fs %-10s %s\n",
				r.Index+1,
				r.Backend,
				status,
				r.Duration,
				sessionID,
				taskName,
			)
			if r.Error != "" && r.Error != "canceled (fail-fast)" {
				fmt.Printf("     Error: %s\n", r.Error)
			}
		}

		fmt.Println(strings.Repeat("-", tableSeparatorWidth))
		fmt.Printf("Total: %d tasks, %d completed, %d failed (%.2fs)\n",
			aggregated.TotalTasks,
			aggregated.Completed,
			aggregated.Failed,
			aggregated.TotalDuration,
		)
	}

	if aggregated.Failed > 0 {
		return fmt.Errorf("%d task(s) failed", aggregated.Failed)
	}

	return nil
}
