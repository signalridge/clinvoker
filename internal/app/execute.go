package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/output"
	"github.com/signalridge/clinvoker/internal/session"
)

// OutputMode controls how execution output is handled.
type OutputMode int

const (
	// OutputModeText outputs plain text content.
	OutputModeText OutputMode = iota
	// OutputModeJSON outputs unified JSON response.
	OutputModeJSON
	// OutputModeStream passes through streaming output directly.
	OutputModeStream
)

// ExecutionResult contains the result of a backend command execution.
type ExecutionResult struct {
	ExitCode        int
	SessionID       string
	Content         string
	Error           string
	Response        *backend.UnifiedResponse
	DurationSeconds float64
}

// ExecutionConfig holds configuration for command execution.
type ExecutionConfig struct {
	Backend    backend.Backend
	Session    *session.Session
	OutputMode OutputMode
	Stdin      bool          // Whether to connect stdin
	Timeout    time.Duration // Command timeout (0 = no timeout)
}

// ErrCommandTimeout is returned when a command exceeds its timeout.
var ErrCommandTimeout = errors.New("command execution timed out")

// ExecuteCommand executes a backend command and returns the result.
// This is the unified execution function that consolidates the execution logic.
func ExecuteCommand(cfg *ExecutionConfig, cmd *exec.Cmd) (*ExecutionResult, error) {
	// Create context with timeout if specified
	ctx := context.Background()
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	switch cfg.OutputMode {
	case OutputModeStream:
		return executeStream(ctx, cfg.Backend, cmd, cfg.Session, cfg.Stdin)
	case OutputModeJSON:
		return executeWithCapture(ctx, cfg.Backend, cmd, cfg.Session, true, cfg.Stdin)
	default: // OutputModeText
		return executeWithCapture(ctx, cfg.Backend, cmd, cfg.Session, false, cfg.Stdin)
	}
}

// executeStream executes a command with direct stream output.
func executeStream(ctx context.Context, b backend.Backend, cmd *exec.Cmd, sess *session.Session, useStdin bool) (*ExecutionResult, error) {
	startTime := time.Now()

	if useStdin {
		cmd.Stdin = os.Stdin
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &ExecutionResult{ExitCode: 1}, err
	}

	var stderrBuf bytes.Buffer
	if b.SeparateStderr() {
		cmd.Stderr = io.MultiWriter(&stderrBuf, os.Stderr)
	} else {
		cmd.Stderr = cmd.Stdout
	}

	if err := cmd.Start(); err != nil {
		return &ExecutionResult{ExitCode: 1}, err
	}

	// Monitor context for timeout
	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	parser := output.NewParser(b.Name(), "")
	if sess != nil {
		parser = output.NewParser(b.Name(), sess.ID)
	}

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	const maxStreamLine = 10 * 1024 * 1024
	scanner.Buffer(buf, maxStreamLine)

	var backendSessionID string
	var tokenUsage *backend.TokenUsage
	var streamErr error
	timedOut := false

scanLoop:
	for scanner.Scan() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			timedOut = true
			_ = cmd.Process.Kill()
			break scanLoop
		default:
		}

		line := scanner.Text()
		fmt.Fprintln(os.Stdout, line)

		event, parseErr := parser.ParseLine(line)
		if parseErr != nil || event == nil {
			continue
		}

		switch event.Type {
		case output.EventInit:
			if content, err := event.GetInitContent(); err == nil && content.BackendSessionID != "" {
				backendSessionID = content.BackendSessionID
			}
		case output.EventDone:
			if content, err := event.GetDoneContent(); err == nil && content.TokenUsage != nil {
				tokenUsage = &backend.TokenUsage{
					InputTokens:  int(content.TokenUsage.InputTokens),
					OutputTokens: int(content.TokenUsage.OutputTokens),
				}
			}
		case output.EventError:
			if content, err := event.GetErrorContent(); err == nil && content.Message != "" {
				streamErr = errors.New(content.Message)
			}
		}
	}

	if scanErr := scanner.Err(); scanErr != nil && streamErr == nil && !timedOut {
		streamErr = scanErr
	}

	// Wait for command to finish or timeout
	var waitErr error
	select {
	case waitErr = <-waitDone:
		// Command finished
	case <-ctx.Done():
		// Context cancelled (timeout)
		_ = cmd.Process.Kill()
		<-waitDone // Wait for process to exit after kill
		timedOut = true
	}

	result := &ExecutionResult{
		DurationSeconds: time.Since(startTime).Seconds(),
		SessionID:       backendSessionID,
	}

	if timedOut {
		result.ExitCode = 124 // Standard timeout exit code
		result.Error = ErrCommandTimeout.Error()
		return result, ErrCommandTimeout
	}

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return &ExecutionResult{ExitCode: 1}, waitErr
		}
	}

	if streamErr != nil {
		result.Error = streamErr.Error()
		if result.ExitCode == 0 {
			result.ExitCode = 1
		}
	} else if result.ExitCode != 0 && stderrBuf.Len() > 0 {
		result.Error = stderrBuf.String()
	}

	var resp *backend.UnifiedResponse
	if backendSessionID != "" || tokenUsage != nil || result.Error != "" {
		resp = &backend.UnifiedResponse{
			SessionID: backendSessionID,
			Usage:     tokenUsage,
			Error:     result.Error,
		}
	}
	result.Response = resp

	if sess != nil {
		updateSessionFromResponse(sess, result.ExitCode, result.Error, resp)
	}

	if result.Error != "" && result.ExitCode != 0 {
		fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", b.Name(), result.Error)
	}

	return result, nil
}

// executeWithCapture executes a command and captures output.
func executeWithCapture(ctx context.Context, b backend.Backend, cmd *exec.Cmd, sess *session.Session, outputJSON bool, useStdin bool) (*ExecutionResult, error) {
	startTime := time.Now()

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	if b.SeparateStderr() {
		cmd.Stderr = &stderrBuf
	} else {
		cmd.Stderr = &stdoutBuf
	}

	if useStdin {
		cmd.Stdin = os.Stdin
	}

	if err := cmd.Start(); err != nil {
		return &ExecutionResult{ExitCode: 1}, err
	}

	// Monitor context for timeout
	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	var waitErr error
	timedOut := false

	select {
	case waitErr = <-waitDone:
		// Command finished normally
	case <-ctx.Done():
		// Context cancelled (timeout)
		_ = cmd.Process.Kill()
		<-waitDone // Wait for process to exit after kill
		timedOut = true
	}

	result := &ExecutionResult{
		DurationSeconds: time.Since(startTime).Seconds(),
	}

	if timedOut {
		result.ExitCode = 124 // Standard timeout exit code
		result.Error = ErrCommandTimeout.Error()
		return result, ErrCommandTimeout
	}

	var errMsg string
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			errMsg = exitErr.Error()
		} else {
			return &ExecutionResult{ExitCode: 1}, waitErr
		}
	}

	rawOutput := selectOutput(stdoutBuf.String(), stderrBuf.String(), result.ExitCode)

	// Parse JSON response
	resp, parseErr := b.ParseJSONResponse(rawOutput)
	result.Response = resp

	if parseErr == nil && resp != nil {
		result.SessionID = resp.SessionID
		result.Content = resp.Content
		if resp.Error != "" {
			result.Error = resp.Error
			// If backend reports an error but process exited 0, treat as failure
			if result.ExitCode == 0 {
				result.ExitCode = 1
			}
		}
	} else {
		// Fallback to text parsing
		result.Content = b.ParseOutput(rawOutput)
	}

	// Capture process-level error if no response error was set
	if result.Error == "" && errMsg != "" {
		result.Error = errMsg
	}

	// Update session if available
	if sess != nil {
		updateSessionFromResponse(sess, result.ExitCode, errMsg, resp)
	}

	// Output based on mode
	if outputJSON {
		if err := outputJSONResult(b, result, sess); err != nil {
			// Set exit code to indicate encoding failure if not already set
			if result.ExitCode == 0 {
				result.ExitCode = 1
			}
		}
	} else {
		outputTextResult(b, result)
	}

	return result, nil
}

// outputTextResult outputs the result as plain text.
func outputTextResult(b backend.Backend, result *ExecutionResult) {
	if result.Response != nil && result.Response.Error != "" {
		fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", b.Name(), result.Response.Error)
	}

	if result.Content != "" {
		fmt.Print(result.Content)
		if len(result.Content) > 0 && result.Content[len(result.Content)-1] != '\n' {
			fmt.Println()
		}
	}

	cfg := config.Get()
	if cfg.Output.ShowTokens && result.Response != nil && result.Response.Usage != nil {
		total := result.Response.Usage.TotalTokens
		if total == 0 {
			total = result.Response.Usage.InputTokens + result.Response.Usage.OutputTokens
		}
		fmt.Printf("\nTokens: %d (input: %d, output: %d)\n",
			total,
			result.Response.Usage.InputTokens,
			result.Response.Usage.OutputTokens)
	}

	if cfg.Output.ShowTiming {
		fmt.Printf("Time: %.2fs\n", result.DurationSeconds)
	}
}

// outputJSONResult outputs the result as unified JSON.
// Returns an error if JSON encoding fails.
func outputJSONResult(b backend.Backend, result *ExecutionResult, sess *session.Session) error {
	pr := PromptResult{
		Backend:  b.Name(),
		Duration: result.DurationSeconds,
		ExitCode: result.ExitCode,
		Content:  result.Content,
		Error:    result.Error,
	}

	if result.Response != nil {
		pr.SessionID = result.Response.SessionID
		pr.Model = result.Response.Model
		pr.Usage = result.Response.Usage
		pr.Raw = result.Response.Raw
	} else if sess != nil {
		pr.SessionID = sess.ID
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(pr); err != nil {
		// Write error to stderr since stdout JSON may be corrupt
		fmt.Fprintf(os.Stderr, "Error encoding JSON output: %v\n", err)
		return err
	}
	return nil
}

// DetermineOutputMode converts OutputFormat to OutputMode.
func DetermineOutputMode(format backend.OutputFormat) OutputMode {
	switch format {
	case backend.OutputStreamJSON:
		return OutputModeStream
	case backend.OutputJSON:
		return OutputModeJSON
	default:
		return OutputModeText
	}
}

// DetermineInternalFormat returns the internal output format to use.
// For text output, we use JSON internally to capture session IDs.
func DetermineInternalFormat(userFormat backend.OutputFormat) backend.OutputFormat {
	if userFormat == backend.OutputText || userFormat == backend.OutputDefault || userFormat == "" {
		return backend.OutputJSON
	}
	return userFormat
}

// GetCommandTimeout returns the command timeout from config.
// Returns 0 if no timeout is configured.
func GetCommandTimeout() time.Duration {
	cfg := config.Get()
	if cfg.UnifiedFlags.CommandTimeoutSecs > 0 {
		return time.Duration(cfg.UnifiedFlags.CommandTimeoutSecs) * time.Second
	}
	return 0
}
