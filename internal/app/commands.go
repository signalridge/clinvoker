package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/executor"
	"github.com/signalridge/clinvoker/internal/session"
	"github.com/spf13/cobra"
)

// versionCmd prints the version information.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("clinvoker %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

// resumeCmd resumes a previous session.
var resumeCmd = &cobra.Command{
	Use:   "resume [session-id] [prompt]",
	Short: "Resume a previous session",
	Long: `Resume a previous AI CLI session.

Examples:
  clinvoker resume abc123 "continue working"
  clinvoker resume --last "follow up"
  clinvoker resume --last
  clinvoker resume --backend claude
  clinvoker resume (interactive picker)`,
	Args: cobra.MaximumNArgs(2),
	RunE: runResume,
}

var (
	resumeLast        bool
	resumeBackend     string
	resumeWorkDir     bool
	resumeInteractive bool
)

func init() {
	resumeCmd.Flags().BoolVar(&resumeLast, "last", false, "resume the most recent session")
	resumeCmd.Flags().StringVarP(&resumeBackend, "backend", "b", "", "filter sessions by backend")
	resumeCmd.Flags().BoolVar(&resumeWorkDir, "here", false, "filter sessions by current working directory")
	resumeCmd.Flags().BoolVarP(&resumeInteractive, "interactive", "i", false, "show interactive session picker")
}

func runResume(cmd *cobra.Command, args []string) error {
	store := session.NewStore()

	var sess *session.Session
	var prompt string
	var err error

	// Build filter based on flags
	filter := &session.ListFilter{}
	if resumeBackend != "" {
		filter.Backend = resumeBackend
	}
	if resumeWorkDir {
		wd, err := os.Getwd()
		if err == nil {
			filter.WorkDir = wd
		}
	}

	if resumeLast {
		// Get the most recent session matching the filter
		sessions, err := store.ListWithFilter(filter)
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
		if len(sessions) == 0 {
			return fmt.Errorf("no sessions found matching criteria")
		}
		sess = sessions[0]
		if len(args) > 0 {
			prompt = args[0]
		}
	} else if len(args) > 0 {
		// Try to find session by ID or prefix
		sess, err = store.GetByPrefix(args[0])
		if err != nil {
			// Fall back to exact match
			sess, err = store.Get(args[0])
			if err != nil {
				return err
			}
		}
		if len(args) > 1 {
			prompt = args[1]
		}
	} else if resumeInteractive || (len(args) == 0 && !resumeLast) {
		// Interactive picker
		sess, err = interactiveSessionPicker(store, filter)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("session ID required (or use --last, --interactive)")
	}

	// Get backend
	b, err := backend.Get(sess.Backend)
	if err != nil {
		return fmt.Errorf("backend error: %w", err)
	}

	if !b.IsAvailable() {
		return fmt.Errorf("backend %q is not available", sess.Backend)
	}

	// Build options
	cfg := config.Get()
	opts := &backend.Options{
		WorkDir: sess.WorkingDir,
		Model:   modelName,
	}

	if bcfg, ok := cfg.Backends[sess.Backend]; ok {
		if opts.Model == "" {
			opts.Model = bcfg.Model
		}
	}

	// Build resume command
	sessionID := sess.BackendSessionID
	if sessionID == "" {
		sessionID = sess.ID
	}
	execCmd := b.ResumeCommand(sessionID, prompt, opts)

	if dryRun {
		fmt.Printf("Would execute: %s %v\n", execCmd.Path, execCmd.Args[1:])
		return nil
	}

	// Update session
	sess.MarkUsed()
	if err := store.Save(sess); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
	}

	// Execute
	exec := executor.New()
	exitCode, err := exec.Run(execCmd)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// interactiveSessionPicker displays an interactive session picker.
func interactiveSessionPicker(store *session.Store, filter *session.ListFilter) (*session.Session, error) {
	sessions, err := store.ListWithFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	// Display sessions with numbers
	fmt.Println("Available sessions:")
	fmt.Println()
	fmt.Printf("  %-3s %-8s %-8s %-20s %s\n", "#", "ID", "BACKEND", "LAST USED", "TITLE/PROMPT")
	fmt.Println("  " + strings.Repeat("-", 70))

	for i, s := range sessions {
		// Limit display to 20 sessions
		if i >= 20 {
			fmt.Printf("  ... and %d more sessions\n", len(sessions)-20)
			break
		}

		title := s.DisplayName()
		if len(title) > 35 {
			title = title[:32] + "..."
		}

		fmt.Printf("  %-3d %-8s %-8s %-20s %s\n",
			i+1,
			s.ID[:8],
			s.Backend,
			formatTimeAgo(s.LastUsed),
			title,
		)
	}

	fmt.Println()
	fmt.Print("Enter session number (or q to quit): ")

	var input string
	fmt.Scanln(&input)

	if input == "q" || input == "" {
		return nil, fmt.Errorf("cancelled")
	}

	var idx int
	_, err = fmt.Sscanf(input, "%d", &idx)
	if err != nil || idx < 1 || idx > len(sessions) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	return sessions[idx-1], nil
}

// formatTimeAgo returns a human-readable time ago string.
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("2006-01-02")
	}
}

// sessionsCmd manages sessions.
var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage sessions",
	Long:  "List, show, delete, or clean up sessions.",
}

var (
	listBackendFilter string
	listStatusFilter  string
	listLimit         int
)

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := session.NewStore()

		filter := &session.ListFilter{
			Backend: listBackendFilter,
			Limit:   listLimit,
		}
		if listStatusFilter != "" {
			filter.Status = session.SessionStatus(listStatusFilter)
		}

		sessions, err := store.ListWithFilter(filter)
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			return nil
		}

		fmt.Printf("%-8s %-8s %-10s %-15s %-12s %s\n", "ID", "BACKEND", "STATUS", "LAST USED", "TOKENS", "TITLE/PROMPT")
		fmt.Println(strings.Repeat("-", 90))
		for _, s := range sessions {
			status := string(s.Status)
			if status == "" {
				status = "unknown"
			}

			tokens := "-"
			if s.TokenUsage != nil && s.TokenUsage.Total() > 0 {
				tokens = fmt.Sprintf("%d", s.TokenUsage.Total())
			}

			title := s.DisplayName()
			if len(title) > 30 {
				title = title[:27] + "..."
			}

			fmt.Printf("%-8s %-8s %-10s %-15s %-12s %s\n",
				s.ID[:8],
				s.Backend,
				status,
				formatTimeAgo(s.LastUsed),
				tokens,
				title,
			)
		}

		return nil
	},
}

func init() {
	sessionsListCmd.Flags().StringVarP(&listBackendFilter, "backend", "b", "", "filter by backend")
	sessionsListCmd.Flags().StringVar(&listStatusFilter, "status", "", "filter by status (active, completed, error)")
	sessionsListCmd.Flags().IntVarP(&listLimit, "limit", "n", 0, "limit number of sessions shown")
}

var sessionsShowCmd = &cobra.Command{
	Use:   "show <session-id>",
	Short: "Show session details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := session.NewStore()
		// Try prefix match first
		sess, err := store.GetByPrefix(args[0])
		if err != nil {
			// Fall back to exact match
			sess, err = store.Get(args[0])
			if err != nil {
				return err
			}
		}

		fmt.Printf("ID:                %s\n", sess.ID)
		fmt.Printf("Backend:           %s\n", sess.Backend)
		if sess.Model != "" {
			fmt.Printf("Model:             %s\n", sess.Model)
		}
		fmt.Printf("Status:            %s\n", sess.Status)
		fmt.Printf("Created:           %s\n", sess.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Last Used:         %s (%s)\n", sess.LastUsed.Format(time.RFC3339), formatTimeAgo(sess.LastUsed))
		fmt.Printf("Working Directory: %s\n", sess.WorkingDir)
		if sess.BackendSessionID != "" {
			fmt.Printf("Backend Session:   %s\n", sess.BackendSessionID)
		}
		if sess.Title != "" {
			fmt.Printf("Title:             %s\n", sess.Title)
		}
		if sess.InitialPrompt != "" {
			prompt := sess.InitialPrompt
			if len(prompt) > 80 {
				prompt = prompt[:77] + "..."
			}
			fmt.Printf("Initial Prompt:    %s\n", prompt)
		}
		fmt.Printf("Turns:             %d\n", sess.TurnCount)
		if sess.TokenUsage != nil {
			fmt.Printf("Token Usage:\n")
			fmt.Printf("  Input:           %d\n", sess.TokenUsage.InputTokens)
			fmt.Printf("  Output:          %d\n", sess.TokenUsage.OutputTokens)
			if sess.TokenUsage.CachedTokens > 0 {
				fmt.Printf("  Cached:          %d\n", sess.TokenUsage.CachedTokens)
			}
			if sess.TokenUsage.ReasoningTokens > 0 {
				fmt.Printf("  Reasoning:       %d\n", sess.TokenUsage.ReasoningTokens)
			}
			fmt.Printf("  Total:           %d\n", sess.TokenUsage.Total())
		}
		if len(sess.Tags) > 0 {
			fmt.Printf("Tags:              %s\n", strings.Join(sess.Tags, ", "))
		}
		if sess.ParentID != "" {
			fmt.Printf("Parent Session:    %s\n", sess.ParentID)
		}
		if sess.ErrorMessage != "" {
			fmt.Printf("Error:             %s\n", sess.ErrorMessage)
		}
		if len(sess.Metadata) > 0 {
			fmt.Println("Metadata:")
			for k, v := range sess.Metadata {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		return nil
	},
}

var sessionsDeleteCmd = &cobra.Command{
	Use:   "delete <session-id>",
	Short: "Delete a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := session.NewStore()
		if err := store.Delete(args[0]); err != nil {
			return err
		}
		fmt.Printf("Session %s deleted.\n", args[0])
		return nil
	},
}

var cleanOlderThan string

var sessionsCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up old sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		var days int
		if cleanOlderThan != "" {
			if strings.HasSuffix(cleanOlderThan, "d") {
				fmt.Sscanf(cleanOlderThan, "%dd", &days)
			} else {
				fmt.Sscanf(cleanOlderThan, "%d", &days)
			}
		}
		if days == 0 {
			days = config.Get().Session.RetentionDays
		}

		store := session.NewStore()
		deleted, err := store.CleanByDays(days)
		if err != nil {
			return err
		}

		fmt.Printf("Deleted %d session(s) older than %d days.\n", deleted, days)
		return nil
	},
}

func init() {
	sessionsCleanCmd.Flags().StringVar(&cleanOlderThan, "older-than", "", "delete sessions older than (e.g., 30d)")
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsShowCmd)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
	sessionsCmd.AddCommand(sessionsCleanCmd)
}

// configCmd manages configuration.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		fmt.Printf("Default Backend: %s\n", cfg.DefaultBackend)
		fmt.Printf("\nBackends:\n")
		for name, bcfg := range cfg.Backends {
			fmt.Printf("  %s:\n", name)
			if bcfg.Model != "" {
				fmt.Printf("    model: %s\n", bcfg.Model)
			}
			if bcfg.AllowedTools != "" {
				fmt.Printf("    allowed_tools: %s\n", bcfg.AllowedTools)
			}
		}
		fmt.Printf("\nSession:\n")
		fmt.Printf("  auto_resume: %v\n", cfg.Session.AutoResume)
		fmt.Printf("  retention_days: %d\n", cfg.Session.RetentionDays)

		fmt.Printf("\nAvailable backends:\n")
		for _, name := range backend.List() {
			b, _ := backend.Get(name)
			status := "not installed"
			if b.IsAvailable() {
				status = "available"
			}
			fmt.Printf("  %s: %s\n", name, status)
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := config.Set(key, value); err != nil {
			return fmt.Errorf("failed to set config: %w", err)
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

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
	parallelCmd.Flags().IntVar(&maxParallel, "max-parallel", 3, "maximum number of parallel tasks")
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
		maxP = 3 // fallback default
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

	// Cancel context for fail-fast
	cancelled := false
	var cancelMu sync.Mutex

	for i, task := range tasks.Tasks {
		wg.Add(1)
		go func(idx int, t ParallelTask) {
			defer wg.Done()

			// Check if cancelled (fail-fast)
			cancelMu.Lock()
			if cancelled {
				cancelMu.Unlock()
				mu.Lock()
				aggregated.Results[idx] = TaskResult{
					Index:    idx,
					TaskID:   t.ID,
					TaskName: t.Name,
					Backend:  t.Backend,
					ExitCode: -1,
					Error:    "cancelled (fail-fast)",
				}
				mu.Unlock()
				return
			}
			cancelMu.Unlock()

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
					cancelMu.Lock()
					cancelled = true
					cancelMu.Unlock()
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
					cancelMu.Lock()
					cancelled = true
					cancelMu.Unlock()
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
				cancelMu.Lock()
				cancelled = true
				cancelMu.Unlock()
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
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("%-4s %-12s %-8s %-10s %-10s %s\n", "#", "BACKEND", "STATUS", "DURATION", "SESSION", "TASK")
		fmt.Println(strings.Repeat("-", 80))

		for _, r := range aggregated.Results {
			status := "OK"
			if r.ExitCode != 0 || r.Error != "" {
				status = "FAILED"
			}
			if r.ExitCode == -1 {
				status = "CANCELLED"
			}

			taskName := r.TaskName
			if taskName == "" {
				taskName = tasks.Tasks[r.Index].Prompt
			}
			if len(taskName) > 35 {
				taskName = taskName[:32] + "..."
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
			if r.Error != "" && r.Error != "cancelled (fail-fast)" {
				fmt.Printf("     Error: %s\n", r.Error)
			}
		}

		fmt.Println(strings.Repeat("-", 80))
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

// compareCmd runs the same prompt on multiple backends for comparison.
var compareCmd = &cobra.Command{
	Use:   "compare <prompt>",
	Short: "Run same prompt on multiple backends and compare outputs",
	Long: `Compare AI outputs by running the same prompt across multiple backends.

Examples:
  clinvoker compare "explain quicksort" --backends claude,gemini,codex
  clinvoker compare "review this code" --backends claude,gemini --model opus
  clinvoker compare "fix the bug" --all-backends`,
	Args: cobra.ExactArgs(1),
	RunE: runCompare,
}

var (
	compareBackends    string
	compareAllBackends bool
	compareJSON        bool
	compareSequential  bool
)

func init() {
	compareCmd.Flags().StringVar(&compareBackends, "backends", "", "comma-separated list of backends to compare")
	compareCmd.Flags().BoolVar(&compareAllBackends, "all-backends", false, "run on all available backends")
	compareCmd.Flags().BoolVar(&compareJSON, "json", false, "output results as JSON")
	compareCmd.Flags().BoolVar(&compareSequential, "sequential", false, "run backends sequentially instead of parallel")
}

// CompareResult represents the result from one backend.
type CompareResult struct {
	Backend   string    `json:"backend"`
	Model     string    `json:"model,omitempty"`
	ExitCode  int       `json:"exit_code"`
	Error     string    `json:"error,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  float64   `json:"duration_seconds"`
	SessionID string    `json:"session_id,omitempty"`
}

// CompareResults represents aggregated comparison results.
type CompareResults struct {
	Prompt        string          `json:"prompt"`
	Backends      []string        `json:"backends"`
	Results       []CompareResult `json:"results"`
	TotalDuration float64         `json:"total_duration_seconds"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
}

func runCompare(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	// Determine which backends to use
	var backends []string
	if compareAllBackends {
		backends = backend.List()
	} else if compareBackends != "" {
		backends = strings.Split(compareBackends, ",")
		for i := range backends {
			backends[i] = strings.TrimSpace(backends[i])
		}
	} else {
		return fmt.Errorf("specify backends with --backends or use --all-backends")
	}

	// Validate backends
	var availableBackends []string
	for _, name := range backends {
		b, err := backend.Get(name)
		if err != nil {
			return fmt.Errorf("unknown backend %q: %w", name, err)
		}
		if b.IsAvailable() {
			availableBackends = append(availableBackends, name)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: backend %q is not available, skipping\n", name)
		}
	}

	if len(availableBackends) == 0 {
		return fmt.Errorf("no available backends to compare")
	}

	if !compareJSON {
		fmt.Printf("Comparing %d backends: %s\n", len(availableBackends), strings.Join(availableBackends, ", "))
		fmt.Printf("Prompt: %s\n", truncatePrompt(prompt, 60))
		fmt.Println(strings.Repeat("=", 80))
	}

	// Create session store
	store := session.NewStore()
	cfg := config.Get()

	results := &CompareResults{
		Prompt:    prompt,
		Backends:  availableBackends,
		Results:   make([]CompareResult, len(availableBackends)),
		StartTime: time.Now(),
	}

	if compareSequential {
		// Run sequentially
		for i, name := range availableBackends {
			result := runCompareTask(name, prompt, cfg, store)
			results.Results[i] = result

			if !compareJSON {
				fmt.Printf("\n[%s] %s (%.2fs)\n", name, statusText(result.ExitCode, result.Error), result.Duration)
				fmt.Println(strings.Repeat("-", 80))
			}
		}
	} else {
		// Run in parallel
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i, name := range availableBackends {
			wg.Add(1)
			go func(idx int, backendName string) {
				defer wg.Done()
				result := runCompareTask(backendName, prompt, cfg, store)
				mu.Lock()
				results.Results[idx] = result
				mu.Unlock()
			}(i, name)
		}

		wg.Wait()
	}

	results.EndTime = time.Now()
	results.TotalDuration = results.EndTime.Sub(results.StartTime).Seconds()

	// Output results
	if compareJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		// Print summary
		fmt.Println()
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("COMPARISON SUMMARY")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("%-12s %-10s %-12s %-10s %s\n", "BACKEND", "STATUS", "DURATION", "SESSION", "MODEL")
		fmt.Println(strings.Repeat("-", 80))

		for _, r := range results.Results {
			status := statusText(r.ExitCode, r.Error)
			sessionID := "-"
			if r.SessionID != "" {
				sessionID = r.SessionID[:8]
			}
			modelName := r.Model
			if modelName == "" {
				modelName = "(default)"
			}

			fmt.Printf("%-12s %-10s %-12.2fs %-10s %s\n",
				r.Backend,
				status,
				r.Duration,
				sessionID,
				modelName,
			)
			if r.Error != "" {
				fmt.Printf("             Error: %s\n", r.Error)
			}
		}

		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("Total time: %.2fs\n", results.TotalDuration)
	}

	// Check for failures
	hasError := false
	for _, r := range results.Results {
		if r.ExitCode != 0 || r.Error != "" {
			hasError = true
			break
		}
	}
	if hasError {
		return fmt.Errorf("some backends failed")
	}

	return nil
}

func runCompareTask(backendName, prompt string, cfg *config.Config, store *session.Store) CompareResult {
	startTime := time.Now()
	result := CompareResult{
		Backend:   backendName,
		StartTime: startTime,
	}

	b, err := backend.Get(backendName)
	if err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime).Seconds()
		return result
	}

	// Get model from config or CLI flag
	model := modelName
	if model == "" {
		if bcfg, ok := cfg.Backends[backendName]; ok {
			model = bcfg.Model
		}
	}
	result.Model = model

	// Build unified options
	opts := &backend.UnifiedOptions{
		Model:        model,
		ApprovalMode: backend.ApprovalMode(cfg.UnifiedFlags.ApprovalMode),
		SandboxMode:  backend.SandboxMode(cfg.UnifiedFlags.SandboxMode),
		MaxTurns:     cfg.UnifiedFlags.MaxTurns,
		MaxTokens:    cfg.UnifiedFlags.MaxTokens,
		Verbose:      cfg.UnifiedFlags.Verbose,
		DryRun:       dryRun,
	}

	// Create session
	sess, sessErr := session.NewSession(backendName, "")
	if sessErr == nil {
		sess.SetModel(model)
		sess.InitialPrompt = prompt
		sess.SetStatus(session.StatusActive)
		sess.AddTag("compare")
		if err := store.Save(sess); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
		}
		result.SessionID = sess.ID
	}

	// Build command
	execCmd := b.BuildCommandUnified(prompt, opts)

	if dryRun {
		fmt.Printf("[%s] Would execute: %s %v\n", backendName, execCmd.Path, execCmd.Args[1:])
		result.ExitCode = 0
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime).Seconds()
		return result
	}

	// Execute
	exec := executor.New()
	exec.Stdin = nil
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
		if err := store.Save(sess); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", err)
		}
	}

	return result
}

func statusText(exitCode int, err string) string {
	if exitCode == 0 && err == "" {
		return "OK"
	}
	return "FAILED"
}

func truncatePrompt(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// chainCmd runs backends in sequence, passing context between them.
var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Chain multiple backends in sequence with context passing",
	Long: `Run multiple AI backends in sequence, optionally passing output between steps.

Read chain definition from stdin or a file:
  clinvoker chain --file chain.json
  cat chain.json | clinvoker chain

Chain format (JSON):
  {
    "steps": [
      {"backend": "claude", "prompt": "analyze this code and list issues"},
      {"backend": "codex", "prompt": "fix the issues: {{previous}}"},
      {"backend": "gemini", "prompt": "write tests for: {{previous}}"}
    ]
  }

The {{previous}} placeholder is replaced with the previous step's session ID,
allowing the next backend to access context via session resume.`,
	RunE: runChain,
}

var (
	chainFile string
	chainJSON bool
)

func init() {
	chainCmd.Flags().StringVarP(&chainFile, "file", "f", "", "file containing chain definition")
	chainCmd.Flags().BoolVar(&chainJSON, "json", false, "output results as JSON")
}

// ChainDefinition represents a chain of backend steps.
type ChainDefinition struct {
	Steps          []ChainStep `json:"steps"`
	StopOnFailure  bool        `json:"stop_on_failure,omitempty"`
	PassSessionID  bool        `json:"pass_session_id,omitempty"`
	PassWorkingDir bool        `json:"pass_working_dir,omitempty"`
}

// ChainStep represents a single step in the chain.
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

// ChainStepResult represents the result of a chain step.
type ChainStepResult struct {
	Step      int       `json:"step"`
	Name      string    `json:"name,omitempty"`
	Backend   string    `json:"backend"`
	ExitCode  int       `json:"exit_code"`
	Error     string    `json:"error,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	Duration  float64   `json:"duration_seconds"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ChainResults represents the aggregated chain execution results.
type ChainResults struct {
	TotalSteps     int               `json:"total_steps"`
	CompletedSteps int               `json:"completed_steps"`
	FailedStep     int               `json:"failed_step,omitempty"`
	Results        []ChainStepResult `json:"results"`
	TotalDuration  float64           `json:"total_duration_seconds"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        time.Time         `json:"end_time"`
}

func runChain(cmd *cobra.Command, args []string) error {
	var input []byte
	var err error

	if chainFile != "" {
		input, err = os.ReadFile(chainFile)
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

	var chain ChainDefinition
	if err := json.Unmarshal(input, &chain); err != nil {
		return fmt.Errorf("failed to parse chain definition: %w", err)
	}

	if len(chain.Steps) == 0 {
		return fmt.Errorf("no steps defined in chain")
	}

	// Default to stop on failure
	if !chain.StopOnFailure {
		chain.StopOnFailure = true
	}

	if !chainJSON {
		fmt.Printf("Executing chain with %d steps\n", len(chain.Steps))
		fmt.Println(strings.Repeat("=", 80))
	}

	store := session.NewStore()
	cfg := config.Get()

	results := &ChainResults{
		TotalSteps: len(chain.Steps),
		Results:    make([]ChainStepResult, 0, len(chain.Steps)),
		StartTime:  time.Now(),
	}

	var previousSessionID string
	var previousWorkDir string

	for i, step := range chain.Steps {
		startTime := time.Now()
		stepResult := ChainStepResult{
			Step:      i + 1,
			Name:      step.Name,
			Backend:   step.Backend,
			StartTime: startTime,
		}

		if !chainJSON {
			stepName := step.Name
			if stepName == "" {
				stepName = fmt.Sprintf("Step %d", i+1)
			}
			fmt.Printf("\n[%d/%d] %s (%s)\n", i+1, len(chain.Steps), stepName, step.Backend)
			fmt.Println(strings.Repeat("-", 80))
		}

		// Get backend
		b, err := backend.Get(step.Backend)
		if err != nil {
			stepResult.Error = err.Error()
			stepResult.ExitCode = 1
			stepResult.EndTime = time.Now()
			stepResult.Duration = stepResult.EndTime.Sub(startTime).Seconds()
			results.Results = append(results.Results, stepResult)
			results.FailedStep = i + 1

			if chain.StopOnFailure {
				break
			}
			continue
		}

		if !b.IsAvailable() {
			stepResult.Error = fmt.Sprintf("backend %q not available", step.Backend)
			stepResult.ExitCode = 1
			stepResult.EndTime = time.Now()
			stepResult.Duration = stepResult.EndTime.Sub(startTime).Seconds()
			results.Results = append(results.Results, stepResult)
			results.FailedStep = i + 1

			if chain.StopOnFailure {
				break
			}
			continue
		}

		// Process prompt with placeholders
		prompt := step.Prompt
		if previousSessionID != "" {
			prompt = strings.ReplaceAll(prompt, "{{previous}}", previousSessionID)
			prompt = strings.ReplaceAll(prompt, "{{session}}", previousSessionID)
		}

		// Determine working directory
		stepWorkDir := step.WorkDir
		if stepWorkDir == "" && chain.PassWorkingDir && previousWorkDir != "" {
			stepWorkDir = previousWorkDir
		}

		// Get model
		model := step.Model
		if model == "" && modelName != "" {
			model = modelName
		}
		if model == "" {
			if bcfg, ok := cfg.Backends[step.Backend]; ok {
				model = bcfg.Model
			}
		}

		// Build unified options
		opts := &backend.UnifiedOptions{
			WorkDir:      stepWorkDir,
			Model:        model,
			ApprovalMode: backend.ApprovalMode(step.ApprovalMode),
			SandboxMode:  backend.SandboxMode(step.SandboxMode),
			MaxTurns:     step.MaxTurns,
			DryRun:       dryRun,
		}

		// Apply config defaults if not set
		if opts.ApprovalMode == "" {
			opts.ApprovalMode = backend.ApprovalMode(cfg.UnifiedFlags.ApprovalMode)
		}
		if opts.SandboxMode == "" {
			opts.SandboxMode = backend.SandboxMode(cfg.UnifiedFlags.SandboxMode)
		}

		// Create session for this step
		sess, sessErr := session.NewSession(step.Backend, stepWorkDir)
		if sessErr != nil {
			if !chainJSON {
				fmt.Fprintf(os.Stderr, "Warning: failed to create session: %v\n", sessErr)
			}
		} else {
			sess.SetModel(model)
			sess.InitialPrompt = prompt
			sess.SetStatus(session.StatusActive)
			sess.AddTag("chain")
			sess.AddTag(fmt.Sprintf("chain-step-%d", i+1))
			if step.Name != "" {
				sess.SetTitle(step.Name)
			}
			if previousSessionID != "" && chain.PassSessionID {
				sess.ParentID = previousSessionID
			}
			if err := store.Save(sess); err != nil && !chainJSON {
				fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
			}
			stepResult.SessionID = sess.ID
		}

		// Build command
		execCmd := b.BuildCommandUnified(prompt, opts)

		if dryRun {
			fmt.Printf("Would execute: %s %v\n", execCmd.Path, execCmd.Args[1:])
			stepResult.ExitCode = 0
			stepResult.EndTime = time.Now()
			stepResult.Duration = stepResult.EndTime.Sub(startTime).Seconds()
			results.Results = append(results.Results, stepResult)
			results.CompletedSteps++

			if sess != nil {
				previousSessionID = sess.ID
			}
			previousWorkDir = stepWorkDir
			continue
		}

		// Execute
		exec := executor.New()
		exitCode, execErr := exec.Run(execCmd)
		if execErr != nil {
			stepResult.Error = execErr.Error()
		}
		stepResult.ExitCode = exitCode
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(startTime).Seconds()

		// Update session (if created successfully)
		if sess != nil {
			sess.IncrementTurn()
			if exitCode == 0 {
				sess.Complete()
			} else {
				sess.SetError(stepResult.Error)
			}
			if err := store.Save(sess); err != nil && !chainJSON {
				fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", err)
			}
		}

		if exitCode == 0 {
			results.CompletedSteps++
		} else {
			results.FailedStep = i + 1
		}

		results.Results = append(results.Results, stepResult)

		// Store for next iteration
		if sess != nil {
			previousSessionID = sess.ID
		}
		previousWorkDir = stepWorkDir

		if exitCode != 0 && chain.StopOnFailure {
			break
		}
	}

	results.EndTime = time.Now()
	results.TotalDuration = results.EndTime.Sub(results.StartTime).Seconds()

	// Output results
	if chainJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		fmt.Println()
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("CHAIN EXECUTION SUMMARY")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("%-6s %-12s %-8s %-10s %-10s %s\n", "STEP", "BACKEND", "STATUS", "DURATION", "SESSION", "NAME")
		fmt.Println(strings.Repeat("-", 80))

		for _, r := range results.Results {
			status := "OK"
			if r.ExitCode != 0 || r.Error != "" {
				status = "FAILED"
			}

			sessionID := "-"
			if r.SessionID != "" {
				sessionID = r.SessionID[:8]
			}

			name := r.Name
			if name == "" {
				name = "-"
			}

			fmt.Printf("%-6d %-12s %-8s %-10.2fs %-10s %s\n",
				r.Step,
				r.Backend,
				status,
				r.Duration,
				sessionID,
				name,
			)
			if r.Error != "" {
				fmt.Printf("       Error: %s\n", r.Error)
			}
		}

		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("Total: %d/%d steps completed (%.2fs)\n",
			results.CompletedSteps,
			results.TotalSteps,
			results.TotalDuration,
		)
	}

	if results.FailedStep > 0 {
		return fmt.Errorf("chain failed at step %d", results.FailedStep)
	}

	return nil
}
