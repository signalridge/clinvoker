package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/metrics"
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
// If store is provided and the request is not ephemeral, it persists a session.
func StreamPrompt(ctx context.Context, req *PromptRequest, store *session.Store, logger *slog.Logger, forceStateless bool, onEvent func(*output.UnifiedEvent) error) (*StreamResult, error) {
	if logger == nil {
		logger = slog.Default()
	}

	start := time.Now()

	prep, err := preparePrompt(req, forceStateless)
	if err != nil {
		return nil, err
	}

	// Copy options to avoid mutating caller's struct
	opts := *prep.opts
	opts.OutputFormat = backend.OutputStreamJSON

	var sess *session.Session
	sessionID := ""

	if store != nil && !opts.Ephemeral {
		cfg := config.Get()
		tags := append([]string{}, cfg.Session.DefaultTags...)
		tags = append(tags, "api")

		var sessErr error
		sess, sessErr = store.CreateWithOptions(req.Backend, req.WorkDir, &session.SessionOptions{
			Model:         prep.model,
			InitialPrompt: req.Prompt,
			Tags:          tags,
		})
		if sessErr != nil {
			logger.Warn("failed to create session", "backend", req.Backend, "error", sessErr)
		} else {
			sessionID = sess.ID
			for k, v := range req.Metadata {
				sess.SetMetadata(k, v)
			}
			if err := store.Save(sess); err != nil {
				logger.Warn("failed to save session metadata", "session_id", sess.ID, "error", err)
			} else if cfg.Server.MetricsEnabled {
				metrics.IncrementSessionsCreated()
			}
		}
	}

	cmd := prep.backend.BuildCommandUnified(req.Prompt, &opts)
	cmd = util.CommandWithContext(ctx, cmd)

	if opts.DryRun {
		msg := fmt.Sprintf("Would execute: %s %v", cmd.Path, cmd.Args[1:])
		if onEvent != nil {
			event := output.NewUnifiedEvent(output.EventMessage, prep.backend.Name(), sessionID)
			if err := event.SetContent(&output.MessageContent{Text: msg, Role: "assistant"}); err == nil {
				if err := onEvent(event); err != nil {
					return &StreamResult{ExitCode: 1, Error: err.Error()}, err
				}
			}

			done := output.NewUnifiedEvent(output.EventDone, prep.backend.Name(), sessionID)
			if err := done.SetContent(&output.DoneContent{}); err == nil {
				if err := onEvent(done); err != nil {
					return &StreamResult{ExitCode: 1, Error: err.Error()}, err
				}
			}
		}

		result := &StreamResult{ExitCode: 0}

		// Record backend execution metrics if enabled
		execDuration := time.Since(start).Seconds()
		if config.Get().Server.MetricsEnabled {
			metrics.RecordBackendExecution(req.Backend, "success")
			metrics.RecordBackendExecutionDuration(req.Backend, execDuration)
		}

		if sess != nil {
			util.UpdateSessionFromResponse(sess, result.ExitCode, "", nil)
			if err := store.Save(sess); err != nil && logger != nil {
				logger.Warn("failed to save session", "session_id", sess.ID, "error", err)
			}
		}

		return result, nil
	}

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

	parser := output.NewParser(prep.backend.Name(), sessionID)
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
		// Scanner stops on overly long tokens; emit an error event and treat as a stream error.
		errMsg := scanErr.Error()
		if scanErr == bufio.ErrTooLong {
			errMsg = fmt.Sprintf("stream event exceeded maximum size limit of %d bytes; consider reducing output size", maxStreamLine)
		}

		// Emit error event to client if callback is available
		if onEvent != nil {
			errEvent := output.NewUnifiedEvent(output.EventError, prep.backend.Name(), sessionID)
			if err := errEvent.SetContent(&output.ErrorContent{
				Code:    "stream_line_too_long",
				Message: errMsg,
			}); err == nil {
				_ = onEvent(errEvent) // Best effort, ignore error
			}
		}

		streamErr = errors.New(errMsg)
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

	// Record backend execution metrics if enabled
	execDuration := time.Since(start).Seconds()
	if config.Get().Server.MetricsEnabled {
		status := "success"
		if streamErr != nil || exitCode != 0 {
			status = "error"
		}
		metrics.RecordBackendExecution(req.Backend, status)
		metrics.RecordBackendExecutionDuration(req.Backend, execDuration)
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
		if result.ExitCode == 0 {
			result.ExitCode = 1
		}
		return result, handlerErr
	}

	if streamErr != nil {
		result.Error = streamErr.Error()
		if result.ExitCode == 0 {
			result.ExitCode = 1
		}
	} else if exitCode != 0 && stderrBuf.Len() > 0 {
		// Use stderr as fallback error message when exit code is non-zero
		result.Error = stderrBuf.String()
	}

	if sess != nil {
		if backendSessionID != "" {
			sess.BackendSessionID = backendSessionID
		}
		var respUsage *backend.TokenUsage
		if tokenUsage != nil {
			respUsage = &backend.TokenUsage{
				InputTokens:  int(tokenUsage.InputTokens),
				OutputTokens: int(tokenUsage.OutputTokens),
			}
		}
		resp := &backend.UnifiedResponse{
			Usage: respUsage,
			Error: result.Error,
		}
		util.UpdateSessionFromResponse(sess, result.ExitCode, result.Error, resp)
		if err := store.Save(sess); err != nil && logger != nil {
			logger.Warn("failed to save session", "session_id", sess.ID, "error", err)
		}
	}

	return result, nil
}
