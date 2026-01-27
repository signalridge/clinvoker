package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
)

// compareCmd runs the same prompt on multiple backends for comparison.
var compareCmd = &cobra.Command{
	Use:   "compare <prompt>",
	Short: "Run same prompt on multiple backends and compare outputs",
	Long: `Compare AI outputs by running the same prompt across multiple backends.

Examples:
  clinvk compare "explain quicksort" --backends claude,gemini,codex
  clinvk compare "review this code" --backends claude,gemini --model opus
  clinvk compare "fix the bug" --all-backends`,
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
	Output    string    `json:"output,omitempty"`
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
		fmt.Println(strings.Repeat("=", tableSeparatorWidth))
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
				fmt.Printf("[%s] %s\n", name, result.Output)
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

		// Print outputs for parallel mode
		if !compareJSON {
			for _, r := range results.Results {
				fmt.Printf("[%s] %s\n", r.Backend, r.Output)
			}
		}
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
		fmt.Println(strings.Repeat("=", tableSeparatorWidth))
		fmt.Println("COMPARISON SUMMARY")
		fmt.Println(strings.Repeat("=", tableSeparatorWidth))
		fmt.Printf("%-12s %-10s %-12s %-10s %s\n", "BACKEND", "STATUS", "DURATION", "SESSION", "MODEL")
		fmt.Println(strings.Repeat("-", tableSeparatorWidth))

		for _, r := range results.Results {
			status := statusText(r.ExitCode, r.Error)
			sessionID := "-"
			if r.SessionID != "" {
				sessionID = r.SessionID[:8]
			}
			model := r.Model
			if model == "" {
				model = "(default)"
			}

			fmt.Printf("%-12s %-10s %-12.2fs %-10s %s\n",
				r.Backend,
				status,
				r.Duration,
				sessionID,
				model,
			)
			if r.Error != "" {
				fmt.Printf("             Error: %s\n", r.Error)
			}
		}

		fmt.Println(strings.Repeat("-", tableSeparatorWidth))
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

	// Execute with output capture and parsing
	output, exitCode, execErr := ExecuteAndCapture(b, execCmd)
	if execErr != nil {
		result.Error = execErr.Error()
	}
	result.ExitCode = exitCode
	result.Output = output
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
