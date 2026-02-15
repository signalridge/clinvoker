package app

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/util"
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
	Backend   string          `json:"backend"`
	Model     string          `json:"model,omitempty"`
	ExitCode  int             `json:"exit_code"`
	Error     string          `json:"error,omitempty"`
	Output    string          `json:"output,omitempty"`
	OutputLen int             `json:"output_length"`
	Retry     *RetryTelemetry `json:"retry,omitempty"`
	StartTime time.Time       `json:"start_time"`
	EndTime   time.Time       `json:"end_time"`
	Duration  float64         `json:"duration_seconds"`
}

// CompareSummaryEntry describes one backend's score and ranking dimensions.
type CompareSummaryEntry struct {
	Backend        string  `json:"backend"`
	Status         string  `json:"status"`
	LatencySeconds float64 `json:"latency_seconds"`
	OutputLength   int     `json:"output_length"`
	Score          float64 `json:"score"`
	Rank           int     `json:"rank"`
}

// CompareSummary contains ranking and scoring metadata for compare results.
type CompareSummary struct {
	ScoreFormulaVersion string                `json:"score_formula_version"`
	Dimensions          []string              `json:"dimensions"`
	Ranking             []CompareSummaryEntry `json:"ranking"`
}

// CompareResults represents aggregated comparison results.
type CompareResults struct {
	Prompt        string          `json:"prompt"`
	Backends      []string        `json:"backends"`
	Results       []CompareResult `json:"results"`
	Summary       CompareSummary  `json:"summary"`
	TotalDuration float64         `json:"total_duration_seconds"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
}

func runCompare(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	// Determine which backends to use
	var backends []string
	if compareAllBackends {
		backends = config.EnabledBackends()
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
		if !config.IsBackendEnabled(name) {
			return fmt.Errorf("backend %q is disabled in config", name)
		}
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
		fmt.Printf("Prompt: %s\n", truncateString(prompt, 60))
		fmt.Println(strings.Repeat("=", tableSeparatorWidth))
	}

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
			result := runCompareTask(name, prompt, cfg)
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
				result := runCompareTask(backendName, prompt, cfg)
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
	results.Summary = buildCompareSummary(results.Results)

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
		fmt.Printf("%-12s %-10s %-12s %-11s %-8s %s\n", "BACKEND", "STATUS", "DURATION", "OUT_LEN", "SCORE", "MODEL")
		fmt.Println(strings.Repeat("-", tableSeparatorWidth))

		resultsByBackend := make(map[string]CompareResult, len(results.Results))
		for _, r := range results.Results {
			resultsByBackend[r.Backend] = r
		}

		for _, entry := range results.Summary.Ranking {
			r := resultsByBackend[entry.Backend]
			model := r.Model
			if model == "" {
				model = "(default)"
			}

			fmt.Printf("%-12s %-10s %-12.2fs %-11d %-8.2f %s\n",
				r.Backend,
				entry.Status,
				r.Duration,
				r.OutputLen,
				entry.Score,
				model,
			)
			if r.Error != "" {
				fmt.Printf("             Error: %s\n", r.Error)
			}
		}

		fmt.Println(strings.Repeat("-", tableSeparatorWidth))
		fmt.Println("Ranking rule: score desc, duration asc, backend asc")
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

func runCompareTask(backendName, prompt string, cfg *config.Config) CompareResult {
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
	// Always use internal JSON format to capture errors properly
	opts := &backend.UnifiedOptions{
		Model:        model,
		DryRun:       dryRun,
		OutputFormat: backend.OutputJSON, // Force JSON for proper error capture
		Ephemeral:    true,               // Compare is always ephemeral
	}
	util.ApplyUnifiedDefaults(opts, cfg, dryRun)
	util.ApplyBackendDefaults(opts, backendName, cfg)

	if dryRun {
		execCmd := b.BuildCommandUnified(prompt, opts)
		if !compareJSON {
			fmt.Printf("[%s] Would execute: %s %v\n", backendName, execCmd.Path, execCmd.Args[1:])
		}
		result.ExitCode = 0
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime).Seconds()
		return result
	}

	buildCmd := func() *exec.Cmd {
		return b.BuildCommandUnified(prompt, opts)
	}
	captureResult, retryTelemetry, execErr := executeWithRetryJSON(
		b,
		buildCmd,
		cfg,
		backendName,
		"compare",
		GetCommandTimeout(),
		true,
	)

	if retryTelemetry.Enabled {
		rt := retryTelemetry
		result.Retry = &rt
	}

	if execErr != nil && captureResult.Error == "" {
		result.Error = execErr.Error()
	} else if captureResult.Error != "" {
		result.Error = captureResult.Error
	}
	result.ExitCode = captureResult.ExitCode
	result.Output = captureResult.Content
	result.OutputLen = utf8.RuneCountInString(result.Output)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime).Seconds()
	recordCompareBackendMetrics(cfg, backendName, result.ExitCode, result.Error, result.Duration)

	// Ensure ephemeral compare runs remain clean on the backend.
	if opts.Ephemeral {
		cleanupBackendSession(backendName, captureResult.BackendSessionID)
	}

	return result
}

func buildCompareSummary(results []CompareResult) CompareSummary {
	summary := CompareSummary{
		ScoreFormulaVersion: "v1",
		Dimensions:          []string{"status", "latency_seconds", "output_length"},
		Ranking:             make([]CompareSummaryEntry, 0, len(results)),
	}

	var maxDuration float64
	for _, r := range results {
		if r.ExitCode == 0 && r.Error == "" && r.Duration > maxDuration {
			maxDuration = r.Duration
		}
	}

	for _, r := range results {
		entry := CompareSummaryEntry{
			Backend:        r.Backend,
			Status:         statusText(r.ExitCode, r.Error),
			LatencySeconds: r.Duration,
			OutputLength:   r.OutputLen,
			Score:          compareScore(&r, maxDuration),
		}
		summary.Ranking = append(summary.Ranking, entry)
	}

	sort.Slice(summary.Ranking, func(i, j int) bool {
		if summary.Ranking[i].Score == summary.Ranking[j].Score {
			if summary.Ranking[i].LatencySeconds == summary.Ranking[j].LatencySeconds {
				return summary.Ranking[i].Backend < summary.Ranking[j].Backend
			}
			return summary.Ranking[i].LatencySeconds < summary.Ranking[j].LatencySeconds
		}
		return summary.Ranking[i].Score > summary.Ranking[j].Score
	})

	for i := range summary.Ranking {
		summary.Ranking[i].Rank = i + 1
	}

	return summary
}

func compareScore(result *CompareResult, maxDuration float64) float64 {
	if result.ExitCode != 0 || result.Error != "" {
		return 0
	}

	statusScore := 60.0

	latencyScore := 30.0
	if maxDuration > 0 {
		latencyScore = 30.0 * (1.0 - (result.Duration / maxDuration))
		if latencyScore < 0 {
			latencyScore = 0
		}
	}

	outputScore := math.Min(10.0, math.Log1p(float64(result.OutputLen)))
	total := statusScore + latencyScore + outputScore

	return math.Round(total*100) / 100
}

func statusText(exitCode int, err string) string {
	if exitCode == 0 && err == "" {
		return "OK"
	}
	return "FAILED"
}
