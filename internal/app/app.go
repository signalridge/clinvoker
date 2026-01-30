// Package app provides the CLI application using cobra.
package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
	"github.com/signalridge/clinvoker/internal/util"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	cfgFile             string
	backendName         string
	modelName           string
	workDir             string
	dryRun              bool
	outputFormat        string // text, json, stream-json
	continueLastSession bool   // continue last session
	ephemeralMode       bool   // stateless mode, no session persisted
)

var rootCmd = &cobra.Command{
	Use:   "clinvk [prompt]",
	Short: "Unified AI CLI wrapper for multiple backends",
	Long: `clinvk is a unified CLI wrapper that orchestrates multiple AI CLI backends
(Claude Code, Codex CLI, Gemini CLI) with session persistence and parallel task execution.

Examples:
  clinvk "fix the bug in auth.go"
  clinvk --backend codex "implement user registration"
  clinvk -b gemini "generate unit tests"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrompt,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.clinvk/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&backendName, "backend", "b", "", "AI backend to use (claude, codex, gemini)")
	rootCmd.PersistentFlags().StringVarP(&modelName, "model", "m", "", "model to use for the backend")
	rootCmd.PersistentFlags().StringVarP(&workDir, "workdir", "w", "", "working directory for the AI backend")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print command without executing")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "json", "output format: text, json, stream-json")
	rootCmd.PersistentFlags().BoolVar(&ephemeralMode, "ephemeral", false, "stateless mode: don't persist session (like standard LLM APIs)")
	rootCmd.Flags().BoolVarP(&continueLastSession, "continue", "c", false, "continue the last session")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(parallelCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(chainCmd)
}

func initConfig() {
	if err := config.Init(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the version info for the CLI.
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

// promptContext holds the context for prompt execution.
type promptContext struct {
	cfg         *config.Config
	backend     backend.Backend
	backendName string
	opts        *backend.UnifiedOptions
	store       *session.Store
	sess        *session.Session
	dryRun      bool
	userFormat  backend.OutputFormat
	ephemeral   bool
}

// normalizedFlags holds normalized flag values after applying config defaults.
type normalizedFlags struct {
	outputFormat string
	dryRun       bool
}

// normalizeFlags normalizes output format and dry-run flags using config defaults.
// This is extracted so it can be reused by --continue path without backend resolution.
func normalizeFlags(cmd *cobra.Command) *normalizedFlags {
	cfg := config.Get()

	// Apply config default output format if flag not explicitly set
	effectiveOutputFormat := outputFormat
	if !cmd.Flags().Changed("output-format") {
		// Ignore flag default to allow config to override
		effectiveOutputFormat = ""
	}
	effectiveOutputFormat = util.ApplyOutputFormatDefault(effectiveOutputFormat, cfg)
	effectiveOutputFormat = strings.ToLower(effectiveOutputFormat)

	effectiveDryRun := dryRun
	if !cmd.Flags().Changed("dry-run") && cfg.UnifiedFlags.DryRun {
		effectiveDryRun = true
	}

	return &normalizedFlags{
		outputFormat: effectiveOutputFormat,
		dryRun:       effectiveDryRun,
	}
}

// preparePromptContext prepares the context for prompt execution.
func preparePromptContext(cmd *cobra.Command, _ string) (*promptContext, error) {
	cfg := config.Get()
	flags := normalizeFlags(cmd)

	// Validate output format
	switch backend.OutputFormat(flags.outputFormat) {
	case backend.OutputDefault, backend.OutputText, backend.OutputJSON, backend.OutputStreamJSON, "":
		// Valid formats
	default:
		return nil, fmt.Errorf("invalid output format %q: must be one of: text, json, stream-json", flags.outputFormat)
	}

	// Determine backend
	bn := backendName
	if bn == "" {
		bn = cfg.DefaultBackend
	}
	if bn == "" {
		bn = "claude"
	}

	// Get backend
	b, err := backend.Get(bn)
	if err != nil {
		return nil, fmt.Errorf("backend error: %w", err)
	}

	// Skip availability check in dry-run mode
	if !flags.dryRun && !b.IsAvailable() {
		return nil, fmt.Errorf("backend %q is not available (CLI not found in PATH)", bn)
	}

	userFormat := backend.OutputFormat(flags.outputFormat)
	internalFormat := DetermineInternalFormat(userFormat)

	// Build unified options
	opts := &backend.UnifiedOptions{
		WorkDir:      workDir,
		Model:        modelName,
		OutputFormat: internalFormat,
		Ephemeral:    ephemeralMode,
	}
	applyUnifiedDefaults(opts, cfg, flags.dryRun)
	applyBackendDefaults(opts, bn, cfg)

	// Get backend-specific model if not already set
	if opts.Model == "" {
		if bcfg, ok := cfg.Backends[bn]; ok {
			opts.Model = bcfg.Model
		}
	}

	return &promptContext{
		cfg:         cfg,
		backend:     b,
		backendName: bn,
		opts:        opts,
		dryRun:      flags.dryRun,
		userFormat:  userFormat,
		ephemeral:   ephemeralMode,
	}, nil
}

func runPrompt(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && !continueLastSession {
		return cmd.Help()
	}

	var prompt string
	if len(args) > 0 {
		prompt = args[0]
	}

	// If --continue flag is set or auto-resume config is true, resume the last session
	// Use normalizeFlags directly to avoid checking default backend availability
	// (the session's backend is what matters, not the default backend)
	cfg := config.Get()
	if continueLastSession || (cfg.Session.AutoResume && !ephemeralMode) {
		flags := normalizeFlags(cmd)
		// Check if there is any session to resume
		store := session.NewStore()
		filter := &session.ListFilter{}
		if backendName != "" {
			filter.Backend = backendName
		}
		sessions, err := store.ListWithFilter(filter)
		if err == nil && len(sessions) > 0 {
			if continueLastSession {
				return runContinueLastSession(cmd, prompt, flags)
			}
			if len(filterResumableSessions(sessions)) > 0 {
				return runContinueLastSession(cmd, prompt, flags)
			}
		}
		// If no sessions found, fall back to creating new session
	}

	ctx, err := preparePromptContext(cmd, prompt)
	if err != nil {
		return err
	}

	// Build command
	execCmd := ctx.backend.BuildCommandUnified(prompt, ctx.opts)

	if ctx.dryRun {
		fmt.Printf("Would execute: %s %v\n", execCmd.Path, execCmd.Args[1:])
		return nil
	}

	// Create session (skip if ephemeral mode)
	if !ctx.ephemeral {
		ctx.store = session.NewStore()
		sessOpts := &session.SessionOptions{
			Model:         ctx.opts.Model,
			InitialPrompt: prompt,
			Tags:          append([]string{}, cfg.Session.DefaultTags...),
		}
		ctx.sess, err = ctx.store.CreateWithOptions(ctx.backendName, workDir, sessOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create session: %v\n", err)
		}
	}

	// Execute using unified execution
	execCfg := &ExecutionConfig{
		Backend:    ctx.backend,
		Session:    ctx.sess,
		OutputMode: DetermineOutputMode(ctx.userFormat),
		Stdin:      true,
		Timeout:    GetCommandTimeout(),
	}
	result, err := ExecuteCommand(execCfg, execCmd)

	// Update session with backend session ID (skip if ephemeral mode)
	if ctx.sess != nil && ctx.store != nil {
		ctx.sess.MarkUsed()
		if result != nil && result.SessionID != "" {
			ctx.sess.BackendSessionID = result.SessionID
		}
		if saveErr := ctx.store.Save(ctx.sess); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", saveErr)
		}
	}

	// Clean up backend session if ephemeral mode
	if ctx.ephemeral && result != nil {
		cleanupBackendSession(ctx.backendName, result.SessionID)
	}

	if err != nil {
		return err
	}

	if result != nil && result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	return nil
}

// runContinueLastSession continues the most recent session.
// It uses normalizedFlags directly to avoid checking default backend availability.
func runContinueLastSession(_ *cobra.Command, prompt string, flags *normalizedFlags) error {
	// Validate output format
	switch backend.OutputFormat(flags.outputFormat) {
	case backend.OutputDefault, backend.OutputText, backend.OutputJSON, backend.OutputStreamJSON, "":
		// Valid formats
	default:
		return fmt.Errorf("invalid output format %q: must be one of: text, json, stream-json", flags.outputFormat)
	}

	store := session.NewStore()

	// Build filter based on flags
	filter := &session.ListFilter{}
	if backendName != "" {
		filter.Backend = backendName
	}

	// Get the most recent session
	sessions, err := store.ListWithFilter(filter)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}
	if len(sessions) == 0 {
		return fmt.Errorf("no sessions found to continue")
	}

	resumable := filterResumableSessions(sessions)
	if len(resumable) == 0 {
		return fmt.Errorf("no resumable sessions found (missing backend session id)")
	}

	sess := resumable[0]

	// Get backend from session (not default backend)
	b, err := backend.Get(sess.Backend)
	if err != nil {
		return fmt.Errorf("backend error: %w", err)
	}

	cfg := config.Get()

	// Check session's backend availability (not default backend)
	if !flags.dryRun && !b.IsAvailable() {
		return fmt.Errorf("backend %q is not available", sess.Backend)
	}

	// Determine output formats (use normalized format)
	userFormat := backend.OutputFormat(flags.outputFormat)
	internalFormat := DetermineInternalFormat(userFormat)

	// Build unified options
	opts := &backend.UnifiedOptions{
		WorkDir:      sess.WorkingDir,
		Model:        modelName,
		OutputFormat: internalFormat,
	}
	applyUnifiedDefaults(opts, cfg, flags.dryRun)
	applyBackendDefaults(opts, sess.Backend, cfg)

	// Get backend-specific model if not already set
	if opts.Model == "" {
		if bcfg, ok := cfg.Backends[sess.Backend]; ok {
			opts.Model = bcfg.Model
		}
	}

	// Get session ID for resume
	bSessionID := sess.BackendSessionID
	if bSessionID == "" {
		return fmt.Errorf("session %s has no backend session id; cannot resume", shortSessionID(sess.ID))
	}

	// Build resume command
	execCmd := b.ResumeCommandUnified(bSessionID, prompt, opts)

	if flags.dryRun {
		fmt.Printf("Would continue session %s (%s)\n", shortSessionID(sess.ID), sess.Backend)
		fmt.Printf("Command: %s %v\n", execCmd.Path, execCmd.Args[1:])
		return nil
	}

	// Update session
	sess.MarkUsed()
	if err := store.Save(sess); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
	}

	// Execute using unified execution
	execCfg := &ExecutionConfig{
		Backend:    b,
		Session:    sess,
		OutputMode: DetermineOutputMode(userFormat),
		Stdin:      true,
		Timeout:    GetCommandTimeout(),
	}
	result, err := ExecuteCommand(execCfg, execCmd)

	// Persist session updates (including backend session ID) after execution.
	if result != nil && sess != nil {
		sess.MarkUsed()
		if result.SessionID != "" {
			sess.BackendSessionID = result.SessionID
		}
		if saveErr := store.Save(sess); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", saveErr)
		}
	}

	if err != nil {
		return err
	}

	if result != nil && result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	return nil
}

// selectOutput determines which output stream to use based on exit code.
// If exit code is non-zero and stderr has content, prefer stderr (likely error message).
// Otherwise use stdout, falling back to stderr if stdout is empty.
func selectOutput(stdout, stderr string, exitCode int) string {
	if exitCode != 0 && stderr != "" {
		return stderr
	}
	if stdout == "" {
		return stderr
	}
	return stdout
}

// ExecuteAndCapture executes a command and returns the parsed output.
// This is used by commands like compare, parallel, and chain that need to capture output.
// Note: This function expects text output format, not JSON.
// Deprecated: Use ExecuteAndCaptureWithJSON for proper session ID capture.
func ExecuteAndCapture(b backend.Backend, cmd *exec.Cmd) (output string, exitCode int, err error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdin = nil
	cmd.Stdout = &stdoutBuf
	if b.SeparateStderr() {
		cmd.Stderr = &stderrBuf
	} else {
		cmd.Stderr = &stdoutBuf
	}

	if err := cmd.Start(); err != nil {
		return "", 1, err
	}

	waitErr := cmd.Wait()
	exitCode = 0
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", 1, waitErr
		}
	}

	rawOutput := selectOutput(stdoutBuf.String(), stderrBuf.String(), exitCode)

	// Use text parsing (this function is for text output mode)
	output = b.ParseOutput(rawOutput)
	return output, exitCode, nil
}

// CaptureResult contains the result of command execution with JSON parsing.
type CaptureResult struct {
	Content          string // Parsed text content
	BackendSessionID string // Session ID from backend's JSON response
	ExitCode         int
	Error            string
	Response         *backend.UnifiedResponse // Full parsed response (may be nil)
}

// ExecuteAndCaptureWithJSON executes a command that uses JSON output format internally
// and returns parsed content along with backend session ID.
// This properly captures backend session IDs for resume functionality.
// The cmd should already be built with JSON output format.
func ExecuteAndCaptureWithJSON(b backend.Backend, cmd *exec.Cmd) (*CaptureResult, error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdin = nil
	cmd.Stdout = &stdoutBuf
	if b.SeparateStderr() {
		cmd.Stderr = &stderrBuf
	} else {
		cmd.Stderr = &stdoutBuf
	}

	if err := cmd.Start(); err != nil {
		return &CaptureResult{ExitCode: 1, Error: err.Error()}, err
	}

	waitErr := cmd.Wait()
	result := &CaptureResult{}

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = exitErr.Error()
		} else {
			return &CaptureResult{ExitCode: 1, Error: waitErr.Error()}, waitErr
		}
	}

	rawOutput := selectOutput(stdoutBuf.String(), stderrBuf.String(), result.ExitCode)

	// Try to parse as JSON response
	resp, parseErr := b.ParseJSONResponse(rawOutput)
	if parseErr == nil && resp != nil {
		result.Response = resp
		result.Content = resp.Content
		result.BackendSessionID = resp.SessionID
		if resp.Error != "" {
			result.Error = resp.Error
			// If backend reports error but process exited 0, treat as failure
			if result.ExitCode == 0 {
				result.ExitCode = 1
			}
		}
	} else {
		// Fallback to text parsing if JSON parsing fails
		result.Content = b.ParseOutput(rawOutput)
	}

	return result, nil
}

// PromptResult represents a unified result for JSON output.
type PromptResult struct {
	Backend   string              `json:"backend"`
	Content   string              `json:"content"`
	SessionID string              `json:"session_id,omitempty"`
	Model     string              `json:"model,omitempty"`
	Duration  float64             `json:"duration_seconds"`
	ExitCode  int                 `json:"exit_code"`
	Error     string              `json:"error,omitempty"`
	Usage     *backend.TokenUsage `json:"usage,omitempty"`
	Raw       map[string]any      `json:"raw,omitempty"`
}
