package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os/exec"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/output"
	"github.com/signalridge/clinvoker/internal/session"
	"github.com/signalridge/clinvoker/internal/util"
)

// StreamResult represents the result of a streaming prompt execution.
type StreamResult struct {
	ExitCode         int
	Error            string
	TokenUsage       *session.TokenUsage
	BackendSessionID string
}

// StreamPrompt executes a prompt and emits unified events as they stream.
func StreamPrompt(ctx context.Context, req *PromptRequest, logger *slog.Logger, forceStateless bool, onEvent func(*output.UnifiedEvent) error) (*StreamResult, error) {
	if logger == nil {
		logger = slog.Default()
	}

	prep, err := preparePrompt(req, forceStateless)
	if err != nil {
		return nil, err
	}

	// Copy options to avoid mutating caller's struct
	opts := *prep.opts
	opts.OutputFormat = backend.OutputStreamJSON

	cmd := prep.backend.BuildCommandUnified(req.Prompt, &opts)
	cmd = util.CommandWithContext(ctx, cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Capture stderr instead of discarding - errors may appear there
	var stderrBuf bytes.Buffer
	if prep.backend.SeparateStderr() {
		cmd.Stderr = &stderrBuf
	} else {
		cmd.Stderr = cmd.Stdout
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	parser := output.NewParser(prep.backend.Name(), "")
	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	const maxStreamLine = 10 * 1024 * 1024
	scanner.Buffer(buf, maxStreamLine)

	var backendSessionID string
	var tokenUsage *session.TokenUsage
	var handlerErr error
	var streamErr error

	for scanner.Scan() {
		line := scanner.Text()
		event, parseErr := parser.ParseLine(line)
		if parseErr != nil {
			logger.Warn("failed to parse stream line", "backend", prep.backend.Name(), "error", parseErr)
			continue
		}
		if event == nil {
			continue
		}

		switch event.Type {
		case output.EventInit:
			if content, err := event.GetInitContent(); err == nil && content.BackendSessionID != "" {
				backendSessionID = content.BackendSessionID
			}
		case output.EventDone:
			if content, err := event.GetDoneContent(); err == nil && content.TokenUsage != nil {
				tokenUsage = &session.TokenUsage{
					InputTokens:  content.TokenUsage.InputTokens,
					OutputTokens: content.TokenUsage.OutputTokens,
				}
			}
		case output.EventError:
			if content, err := event.GetErrorContent(); err == nil && content.Message != "" {
				streamErr = errors.New(content.Message)
			}
		}

		if onEvent != nil {
			if err := onEvent(event); err != nil {
				handlerErr = err
				break
			}
		}
	}

	if scanErr := scanner.Err(); scanErr != nil && handlerErr == nil {
		// Scanner stops on overly long tokens; treat as a stream error.
		streamErr = scanErr
	}

	if handlerErr != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}

	waitErr := cmd.Wait()
	exitCode := 0
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
			if streamErr == nil {
				streamErr = waitErr
			}
		}
	}

	if opts.Ephemeral {
		util.CleanupBackendSessionWithContext(ctx, req.Backend, backendSessionID)
	}

	result := &StreamResult{
		ExitCode:         exitCode,
		TokenUsage:       tokenUsage,
		BackendSessionID: backendSessionID,
	}

	if handlerErr != nil {
		result.Error = handlerErr.Error()
		return result, handlerErr
	}

	if streamErr != nil {
		result.Error = streamErr.Error()
	} else if exitCode != 0 && stderrBuf.Len() > 0 {
		// Use stderr as fallback error message when exit code is non-zero
		result.Error = stderrBuf.String()
	}

	return result, nil
}
