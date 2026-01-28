package core

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/executor"
)

// Request is the execution input for the core executor.
type Request struct {
	Backend         backend.Backend
	Prompt          string
	Options         *backend.UnifiedOptions
	RequestedFormat backend.OutputFormat
}

// Result is the execution output from the core executor.
type Result struct {
	Output           string
	Usage            *backend.TokenUsage
	BackendSessionID string
	ExitCode         int
	Error            string
}

// Execute runs a prompt with the given backend and options.
// It always uses JSON internally to normalize output and extract session/usage.
func Execute(ctx context.Context, req *Request) (*Result, error) {
	if req == nil || req.Backend == nil {
		return nil, fmt.Errorf("invalid execution request")
	}

	opts := req.Options
	if opts == nil {
		opts = &backend.UnifiedOptions{}
	}

	// Always use JSON internally for parsing/normalization
	opts.OutputFormat = backend.OutputJSON

	// Build command with context
	execCmd := req.Backend.BuildCommandUnified(req.Prompt, opts)
	execCmd = commandWithContext(ctx, execCmd)

	if opts.DryRun {
		return &Result{
			Output:   fmt.Sprintf("Would execute: %s %v", execCmd.Path, execCmd.Args[1:]),
			ExitCode: 0,
		}, nil
	}

	// Execute and capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	runner := executor.New()
	runner.Stdin = nil
	runner.Stdout = &stdoutBuf

	if req.Backend.SeparateStderr() {
		runner.Stderr = &stderrBuf
	} else {
		runner.Stderr = &stdoutBuf
	}

	exitCode, execErr := runner.RunSimple(execCmd)
	errMsg := ""
	if execErr != nil {
		errMsg = execErr.Error()
	}

	rawOutput := selectOutput(stdoutBuf.String(), stderrBuf.String(), exitCode)

	result := &Result{
		ExitCode: exitCode,
		Error:    errMsg,
	}

	resp, parseErr := req.Backend.ParseJSONResponse(rawOutput)
	if parseErr == nil && resp != nil {
		result.Output = resp.Content
		result.BackendSessionID = resp.SessionID
		result.Usage = resp.Usage
		if resp.Error != "" {
			result.Error = resp.Error
			if result.ExitCode == 0 {
				result.ExitCode = 1
			}
		}
	} else {
		result.Output = req.Backend.ParseOutput(rawOutput)
	}

	return result, nil
}

// selectOutput chooses stdout or stderr based on exit code and content.
func selectOutput(stdout, stderr string, exitCode int) string {
	if exitCode != 0 && stderr != "" {
		return stderr
	}
	if stdout == "" {
		return stderr
	}
	return stdout
}

func commandWithContext(ctx context.Context, cmd *exec.Cmd) *exec.Cmd {
	if ctx == nil || cmd == nil {
		return cmd
	}

	if len(cmd.Args) == 0 {
		return exec.CommandContext(ctx, cmd.Path)
	}

	newCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	newCmd.Dir = cmd.Dir
	newCmd.Env = cmd.Env
	newCmd.SysProcAttr = cmd.SysProcAttr
	newCmd.ExtraFiles = cmd.ExtraFiles
	return newCmd
}
