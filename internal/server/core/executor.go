// Package core provides core execution logic for server handlers.
package core

import (
	"bytes"
	"context"
	"fmt"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/executor"
	"github.com/signalridge/clinvoker/internal/util"
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
// It uses JSON internally (unless stream-json is requested) to normalize output and extract session/usage.
func Execute(ctx context.Context, req *Request) (*Result, error) {
	if req == nil || req.Backend == nil {
		return nil, fmt.Errorf("invalid execution request")
	}

	opts := req.Options
	if opts == nil {
		opts = &backend.UnifiedOptions{}
	}

	// Use InternalOutputFormat to determine actual format
	// This respects stream-json while converting text/default to JSON for parsing
	opts.OutputFormat = util.InternalOutputFormat(req.RequestedFormat)

	// Build command with context using shared util
	execCmd := req.Backend.BuildCommandUnified(req.Prompt, opts)
	execCmd = util.CommandWithContext(ctx, execCmd)

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

	// Use shared util for output selection
	rawOutput := util.SelectOutput(stdoutBuf.String(), stderrBuf.String(), exitCode)

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
