// Package app provides the CLI application using cobra.
package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
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
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "text", "output format: text, json, stream-json")
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

func runPrompt(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && !continueLastSession {
		return cmd.Help()
	}

	var prompt string
	if len(args) > 0 {
		prompt = args[0]
	}

	// Normalize and validate output format
	outputFormat = strings.ToLower(outputFormat)
	switch backend.OutputFormat(outputFormat) {
	case backend.OutputDefault, backend.OutputText, backend.OutputJSON, backend.OutputStreamJSON, "":
		// Valid formats
	default:
		return fmt.Errorf("invalid output format %q: must be one of: text, json, stream-json", outputFormat)
	}

	// If --continue flag is set, resume the last session
	if continueLastSession {
		return runContinueLastSession(prompt)
	}

	// Determine backend
	cfg := config.Get()
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
		return fmt.Errorf("backend error: %w", err)
	}

	// Skip availability check in dry-run mode
	if !dryRun && !b.IsAvailable() {
		return fmt.Errorf("backend %q is not available (CLI not found in PATH)", bn)
	}

	// Determine internal output format
	// For text output, we use JSON internally to capture session ID, then extract content
	userOutputFormat := backend.OutputFormat(outputFormat)
	internalOutputFormat := userOutputFormat
	if userOutputFormat == backend.OutputText || userOutputFormat == backend.OutputDefault || userOutputFormat == "" {
		internalOutputFormat = backend.OutputJSON
	}

	// Build unified options with internal output format
	opts := &backend.UnifiedOptions{
		WorkDir:      workDir,
		Model:        modelName,
		OutputFormat: internalOutputFormat,
		Ephemeral:    ephemeralMode,
	}

	// Get backend-specific config
	if bcfg, ok := cfg.Backends[bn]; ok {
		if opts.Model == "" {
			opts.Model = bcfg.Model
		}
	}

	// Build command
	execCmd := b.BuildCommandUnified(prompt, opts)

	if dryRun {
		fmt.Printf("Would execute: %s %v\n", execCmd.Path, execCmd.Args[1:])
		return nil
	}

	// Create session (skip if ephemeral mode)
	var store *session.Store
	var sess *session.Session
	if !ephemeralMode {
		store = session.NewStore()
		sess, err = store.Create(bn, workDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create session: %v\n", err)
		}
	}

	// Execute based on output format and capture backend session ID
	var exitCode int
	var backendSessionID string
	switch userOutputFormat {
	case backend.OutputJSON:
		exitCode, backendSessionID, err = executeWithJSONOutputAndCapture(b, execCmd, sess)
	case backend.OutputStreamJSON:
		exitCode, err = executeWithStreamOutput(b, execCmd)
	default:
		// Text output: use JSON internally, extract content for display
		exitCode, backendSessionID, err = executeTextViaJSON(b, execCmd)
	}

	// Update session with backend session ID (skip if ephemeral mode)
	if sess != nil && store != nil {
		sess.MarkUsed()
		if backendSessionID != "" {
			sess.BackendSessionID = backendSessionID
		}
		if err := store.Save(sess); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
		}
	}

	// Clean up backend session if ephemeral mode (for backends that don't support --no-session-persistence)
	if ephemeralMode {
		cleanupBackendSession(bn, backendSessionID)
	}

	if err != nil {
		return err
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// runContinueLastSession continues the most recent session.
func runContinueLastSession(prompt string) error {
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

	sess := sessions[0]

	// Get backend
	b, err := backend.Get(sess.Backend)
	if err != nil {
		return fmt.Errorf("backend error: %w", err)
	}

	if !dryRun && !b.IsAvailable() {
		return fmt.Errorf("backend %q is not available", sess.Backend)
	}

	// Determine internal output format (use JSON internally for text to capture session ID)
	userOutputFormat := backend.OutputFormat(outputFormat)
	internalOutputFormat := userOutputFormat
	if userOutputFormat == backend.OutputText || userOutputFormat == backend.OutputDefault || userOutputFormat == "" {
		internalOutputFormat = backend.OutputJSON
	}

	// Build unified options
	opts := &backend.UnifiedOptions{
		WorkDir:      sess.WorkingDir,
		Model:        modelName,
		OutputFormat: internalOutputFormat,
	}

	// Get backend-specific config
	cfg := config.Get()
	if bcfg, ok := cfg.Backends[sess.Backend]; ok {
		if opts.Model == "" {
			opts.Model = bcfg.Model
		}
	}

	// Get session ID for resume
	backendSessionID := sess.BackendSessionID
	if backendSessionID == "" {
		backendSessionID = sess.ID
	}

	// Build resume command
	execCmd := b.ResumeCommandUnified(backendSessionID, prompt, opts)

	if dryRun {
		fmt.Printf("Would continue session %s (%s)\n", shortSessionID(sess.ID), sess.Backend)
		fmt.Printf("Command: %s %v\n", execCmd.Path, execCmd.Args[1:])
		return nil
	}

	// Update session
	sess.MarkUsed()
	if err := store.Save(sess); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
	}

	// Execute based on output format
	var exitCode int
	switch userOutputFormat {
	case backend.OutputJSON:
		exitCode, _, err = executeWithJSONOutputAndCapture(b, execCmd, sess)
	case backend.OutputStreamJSON:
		exitCode, err = executeWithStreamOutput(b, execCmd)
	default:
		// Text output: use JSON internally, extract content for display
		exitCode, _, err = executeTextViaJSON(b, execCmd)
	}

	if err != nil {
		return err
	}

	if exitCode != 0 {
		os.Exit(exitCode)
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

// executeTextViaJSON executes a command with JSON output internally,
// extracts the content for text display, and captures the session ID.
func executeTextViaJSON(b backend.Backend, cmd *exec.Cmd) (exitCode int, sessionID string, err error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdoutBuf
	// Always capture stderr for JSON parsing (some backends output errors to stderr)
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return 1, "", err
	}

	waitErr := cmd.Wait()
	exitCode = 0
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return 1, "", waitErr
		}
	}

	rawOutput := selectOutput(stdoutBuf.String(), stderrBuf.String(), exitCode)

	// Parse JSON response to get content and session ID
	resp, parseErr := b.ParseJSONResponse(rawOutput)
	if parseErr == nil && resp != nil {
		sessionID = resp.SessionID

		// Check for errors first
		if resp.Error != "" {
			fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", b.Name(), resp.Error)
			if exitCode == 0 {
				exitCode = 1
			}
			return exitCode, sessionID, nil
		}

		// Print content as plain text
		if resp.Content != "" {
			fmt.Print(resp.Content)
			if resp.Content[len(resp.Content)-1] != '\n' {
				fmt.Println()
			}
		}
	} else {
		// Fallback: print raw output if JSON parsing fails
		output := b.ParseOutput(rawOutput)
		if output != "" {
			fmt.Print(output)
			if output[len(output)-1] != '\n' {
				fmt.Println()
			}
		}
	}

	return exitCode, sessionID, nil
}

// ExecuteAndCapture executes a command and returns the parsed output.
// This is used by commands like compare, parallel, and chain that need to capture output.
// Note: This function expects text output format, not JSON.
func ExecuteAndCapture(b backend.Backend, cmd *exec.Cmd) (output string, exitCode int, err error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdin = nil
	cmd.Stdout = &stdoutBuf
	// Always capture stderr (some backends output errors to stderr)
	cmd.Stderr = &stderrBuf

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

// executeWithJSONOutputAndCapture executes a command, outputs unified JSON, and returns backend session ID.
func executeWithJSONOutputAndCapture(b backend.Backend, cmd *exec.Cmd, sess *session.Session) (exitCode int, sessionID string, err error) {
	startTime := time.Now()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdoutBuf
	// Always capture stderr for JSON parsing (some backends output errors to stderr)
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return 1, "", err
	}

	waitErr := cmd.Wait()
	exitCode = 0
	var errMsg string
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			errMsg = exitErr.Error()
		} else {
			return 1, "", waitErr
		}
	}

	duration := time.Since(startTime).Seconds()
	rawOutput := selectOutput(stdoutBuf.String(), stderrBuf.String(), exitCode)

	// Parse the backend's JSON response into unified format
	resp, parseErr := b.ParseJSONResponse(rawOutput)

	result := PromptResult{
		Backend:  b.Name(),
		Duration: duration,
		ExitCode: exitCode,
		Error:    errMsg,
	}

	if parseErr == nil && resp != nil {
		result.Content = resp.Content
		result.SessionID = resp.SessionID
		result.Model = resp.Model
		result.Usage = resp.Usage
		result.Raw = resp.Raw
		sessionID = resp.SessionID
		// Include backend error in result (prioritize over exec error)
		if resp.Error != "" {
			result.Error = resp.Error
		}
	} else {
		// Fallback to text parsing if JSON parsing fails
		result.Content = b.ParseOutput(rawOutput)
		if sess != nil {
			result.SessionID = sess.ID
		}
	}

	// Output unified JSON
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return 1, "", err
	}

	return exitCode, sessionID, nil
}

// executeWithStreamOutput executes a command and streams output directly.
// For stream-json, we pass through the backend's native streaming format.
// Note: stderr is always shown to the user (not discarded) so errors are visible.
func executeWithStreamOutput(_ backend.Backend, cmd *exec.Cmd) (int, error) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout // Direct pass-through for streaming
	cmd.Stderr = os.Stderr // Always show stderr so errors are visible

	if err := cmd.Start(); err != nil {
		return 1, err
	}

	err := cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return 1, err
		}
	}

	return exitCode, nil
}
