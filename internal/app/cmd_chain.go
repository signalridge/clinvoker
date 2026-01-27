package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/executor"
	"github.com/signalridge/clinvoker/internal/session"
)

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
	chainFile     string
	chainJSONFlag bool
)

func init() {
	chainCmd.Flags().StringVarP(&chainFile, "file", "f", "", "file containing chain definition")
	chainCmd.Flags().BoolVar(&chainJSONFlag, "json", false, "output results as JSON")
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

	if !chainJSONFlag {
		fmt.Printf("Executing chain with %d steps\n", len(chain.Steps))
		fmt.Println(strings.Repeat("=", tableSeparatorWidth))
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

		if !chainJSONFlag {
			stepName := step.Name
			if stepName == "" {
				stepName = fmt.Sprintf("Step %d", i+1)
			}
			fmt.Printf("\n[%d/%d] %s (%s)\n", i+1, len(chain.Steps), stepName, step.Backend)
			fmt.Println(strings.Repeat("-", tableSeparatorWidth))
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
			if !chainJSONFlag {
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
			if err := store.Save(sess); err != nil && !chainJSONFlag {
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
			if err := store.Save(sess); err != nil && !chainJSONFlag {
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
	if chainJSONFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		fmt.Println()
		fmt.Println(strings.Repeat("=", tableSeparatorWidth))
		fmt.Println("CHAIN EXECUTION SUMMARY")
		fmt.Println(strings.Repeat("=", tableSeparatorWidth))
		fmt.Printf("%-6s %-12s %-8s %-10s %-10s %s\n", "STEP", "BACKEND", "STATUS", "DURATION", "SESSION", "NAME")
		fmt.Println(strings.Repeat("-", tableSeparatorWidth))

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

		fmt.Println(strings.Repeat("-", tableSeparatorWidth))
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
