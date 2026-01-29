package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/util"
)

// chainCmd runs backends in sequence, passing context between them.
var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Chain multiple backends in sequence with context passing",
	Long: `Run multiple AI backends in sequence, passing output between steps.

Read chain definition from stdin or a file:
  clinvk chain --file chain.json
  cat chain.json | clinvk chain

Chain format (JSON):
  {
    "steps": [
      {"backend": "claude", "prompt": "analyze this code and list issues"},
      {"backend": "codex", "prompt": "fix the issues: {{previous}}"},
      {"backend": "gemini", "prompt": "write tests for: {{previous}}"}
    ]
  }

Placeholders:
  {{previous}} - replaced with the previous step's output text

Note: chain is always ephemeral (no sessions are persisted).`,
	RunE: runChain,
}

var (
	chainFile     string
	chainInputFile string
	chainJSONFlag bool
)

func init() {
	chainCmd.Flags().StringVarP(&chainFile, "file", "f", "", "file containing chain definition")
	chainCmd.Flags().StringVar(&chainInputFile, "input", "", "file containing chain definition (deprecated, use --file)")
	chainCmd.Flags().BoolVar(&chainJSONFlag, "json", false, "output results as JSON")
}

// ChainDefinition represents a chain of backend steps.
type ChainDefinition struct {
	Steps          []ChainStep `json:"steps"`
	StopOnFailure  bool        `json:"stop_on_failure,omitempty"`
	PassWorkingDir bool        `json:"pass_working_dir,omitempty"`
	// Deprecated/unsupported fields (chain is always ephemeral).
	PassSessionID  bool `json:"pass_session_id,omitempty"`
	PersistSessions bool `json:"persist_sessions,omitempty"`
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
	Output    string    `json:"output,omitempty"`
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

// chainContext holds state that's passed between chain steps.
type chainContext struct {
	previousWorkDir   string
	previousOutput    string
	hasPreviousOutput bool
	cfg               *config.Config
}

func runChain(cmd *cobra.Command, args []string) error {
	chain, err := parseChainDefinition()
	if err != nil {
		return err
	}

	if !chainJSONFlag {
		fmt.Printf("Executing chain with %d steps\n", len(chain.Steps))
		fmt.Println(strings.Repeat("=", tableSeparatorWidth))
	}

	results := executeChain(chain)
	outputChainResults(results, chain)

	if results.FailedStep > 0 {
		return fmt.Errorf("chain failed at step %d", results.FailedStep)
	}
	return nil
}

// parseChainDefinition reads and parses the chain definition from file or stdin.
func parseChainDefinition() (*ChainDefinition, error) {
	file := chainFile
	if file == "" && chainInputFile != "" {
		file = chainInputFile
	}
	input, err := readInputFromFileOrStdin(file)
	if err != nil {
		return nil, err
	}

	var chain ChainDefinition
	if err := json.Unmarshal(input, &chain); err != nil {
		return nil, fmt.Errorf("failed to parse chain definition: %w", err)
	}

	if len(chain.Steps) == 0 {
		return nil, fmt.Errorf("no steps defined in chain")
	}
	if chain.PassSessionID || chain.PersistSessions {
		return nil, fmt.Errorf("chain is always ephemeral; pass_session_id and persist_sessions are not supported")
	}
	for i, step := range chain.Steps {
		if strings.Contains(step.Prompt, "{{session}}") {
			return nil, fmt.Errorf("chain step %d uses {{session}} but sessions are not persisted", i+1)
		}
	}

	// Default to stop on failure
	if !chain.StopOnFailure {
		chain.StopOnFailure = true
	}

	return &chain, nil
}

// executeChain runs all steps in the chain and returns the results.
func executeChain(chain *ChainDefinition) *ChainResults {
	results := &ChainResults{
		TotalSteps: len(chain.Steps),
		Results:    make([]ChainStepResult, 0, len(chain.Steps)),
		StartTime:  time.Now(),
	}

	ctx := &chainContext{
		cfg: config.Get(),
	}

	for i := range chain.Steps {
		stepResult := executeChainStep(i, &chain.Steps[i], chain, ctx)
		results.Results = append(results.Results, stepResult)

		if stepResult.ExitCode == 0 && stepResult.Error == "" {
			results.CompletedSteps++
		} else {
			results.FailedStep = i + 1
			if chain.StopOnFailure {
				break
			}
		}
	}

	results.EndTime = time.Now()
	results.TotalDuration = results.EndTime.Sub(results.StartTime).Seconds()
	return results
}

// executeChainStep executes a single step in the chain.
func executeChainStep(index int, step *ChainStep, chain *ChainDefinition, ctx *chainContext) ChainStepResult {
	startTime := time.Now()
	result := ChainStepResult{
		Step:      index + 1,
		Name:      step.Name,
		Backend:   step.Backend,
		StartTime: startTime,
	}

	if !chainJSONFlag {
		printStepHeader(index, len(chain.Steps), step)
	}

	// Get and validate backend
	b, err := getBackendOrError(step.Backend)
	if err != nil {
		failStepResult(&result, startTime, err.Error())
		return result
	}

	// Prepare execution context
	prompt := substitutePromptPlaceholders(step.Prompt, ctx.previousOutput, ctx.hasPreviousOutput)
	stepWorkDir := resolveStepWorkDir(step.WorkDir, chain.PassWorkingDir, ctx.previousWorkDir)
	model := resolveModel(step.Model, step.Backend, modelName)

	// Build unified options (always ephemeral)
	opts := buildChainStepOptions(step, stepWorkDir, model, ctx.cfg, true)

	// Build and execute command
	execCmd := b.BuildCommandUnified(prompt, opts)

	if dryRun {
		fmt.Printf("Would execute: %s %v\n", execCmd.Path, execCmd.Args[1:])
		result.ExitCode = 0
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime).Seconds()
		updateChainContext(ctx, stepWorkDir, "", false)
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

	// Print output if not in JSON mode
	if !chainJSONFlag && output != "" {
		fmt.Println(output)
	}

	updateChainContext(ctx, stepWorkDir, result.Output, true)

	return result
}

// failStepResult creates a failed step result.
func failStepResult(result *ChainStepResult, startTime time.Time, errMsg string) {
	result.Error = errMsg
	result.ExitCode = 1
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime).Seconds()
}

// printStepHeader prints the header for a chain step.
func printStepHeader(index, total int, step *ChainStep) {
	stepName := step.Name
	if stepName == "" {
		stepName = fmt.Sprintf("Step %d", index+1)
	}
	fmt.Printf("\n[%d/%d] %s (%s)\n", index+1, total, stepName, step.Backend)
	fmt.Println(strings.Repeat("-", tableSeparatorWidth))
}

// substitutePromptPlaceholders replaces placeholders in the prompt.
func substitutePromptPlaceholders(prompt, previousOutput string, hasPreviousOutput bool) string {
	if hasPreviousOutput {
		prompt = strings.ReplaceAll(prompt, "{{previous}}", previousOutput)
	}
	return prompt
}

// resolveStepWorkDir determines the working directory for a step.
func resolveStepWorkDir(explicit string, passWorkDir bool, previousWorkDir string) string {
	if explicit != "" {
		return explicit
	}
	if passWorkDir && previousWorkDir != "" {
		return previousWorkDir
	}
	return ""
}

// buildChainStepOptions builds unified options for a chain step.
func buildChainStepOptions(step *ChainStep, workDir, model string, cfg *config.Config, ephemeral bool) *backend.UnifiedOptions {
	opts := &backend.UnifiedOptions{
		WorkDir:      workDir,
		Model:        model,
		ApprovalMode: backend.ApprovalMode(step.ApprovalMode),
		SandboxMode:  backend.SandboxMode(step.SandboxMode),
		MaxTurns:     step.MaxTurns,
		DryRun:       dryRun,
		Ephemeral:    ephemeral,
	}

	// Apply config defaults if not set
	if opts.ApprovalMode == "" {
		opts.ApprovalMode = backend.ApprovalMode(cfg.UnifiedFlags.ApprovalMode)
	}
	if opts.SandboxMode == "" {
		opts.SandboxMode = backend.SandboxMode(cfg.UnifiedFlags.SandboxMode)
	}

	// Apply output format default from config
	opts.OutputFormat = backend.OutputFormat(util.ApplyOutputFormatDefault("", cfg))

	return opts
}

// updateChainContext updates the context for the next step.
func updateChainContext(ctx *chainContext, workDir, output string, hasOutput bool) {
	ctx.previousWorkDir = workDir
	if hasOutput {
		ctx.previousOutput = output
		ctx.hasPreviousOutput = true
	}
}

// outputChainResults outputs the chain results in the appropriate format.
func outputChainResults(results *ChainResults, _ *ChainDefinition) {
	if chainJSONFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
		return
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", tableSeparatorWidth))
	fmt.Println("CHAIN EXECUTION SUMMARY")
	fmt.Println(strings.Repeat("=", tableSeparatorWidth))

	fmt.Printf("%-6s %-12s %-8s %-10s %s\n", "STEP", "BACKEND", "STATUS", "DURATION", "NAME")
	fmt.Println(strings.Repeat("-", tableSeparatorWidth))

	for _, r := range results.Results {
		status := "OK"
		if r.ExitCode != 0 || r.Error != "" {
			status = "FAILED"
		}

		name := r.Name
		if name == "" {
			name = "-"
		}

		fmt.Printf("%-6d %-12s %-8s %-10.2fs %s\n",
			r.Step, r.Backend, status, r.Duration, name)

		if r.Error != "" {
			fmt.Printf("       Error: %s\n", r.Error)
		}
	}

	fmt.Println(strings.Repeat("-", tableSeparatorWidth))
	fmt.Printf("Total: %d/%d steps completed (%.2fs)\n",
		results.CompletedSteps, results.TotalSteps, results.TotalDuration)
}
