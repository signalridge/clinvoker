// Package executor provides process execution with PTY support.
package executor

import (
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// Executor handles process execution with terminal support.
type Executor struct {
	// Stdin is the input reader (default: os.Stdin)
	Stdin io.Reader
	// Stdout is the output writer (default: os.Stdout)
	Stdout io.Writer
	// Stderr is the error writer (default: os.Stderr)
	Stderr io.Writer
}

// New creates a new Executor with default I/O.
func New() *Executor {
	return &Executor{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Run executes a command with PTY support for terminal handling.
// Returns the exit code and any error.
func (e *Executor) Run(cmd *exec.Cmd) (int, error) {
	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		// Fallback to non-PTY execution if PTY fails
		return e.runWithoutPTY(cmd)
	}
	defer func() { _ = ptmx.Close() }()

	// Set up signal handling
	sigHandler := NewSignalHandler(cmd.Process, ptmx)
	sigHandler.Start()
	defer sigHandler.Stop()

	// Handle PTY size changes
	if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
		// Non-fatal, continue without resize handling
	}

	// Copy stdin to PTY in a goroutine
	go func() {
		_, err := io.Copy(ptmx, e.Stdin)
		// Ignore EOF and ErrClosedPipe as they're expected when PTY closes
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
			// Non-critical: stdin copy failure doesn't affect command execution
		}
	}()

	// Copy PTY output to stdout
	// Ignore EOF errors as they're expected when the command exits
	if _, err := io.Copy(e.Stdout, ptmx); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
		// Non-critical: we still want to wait for the command
	}

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	return 0, nil
}

// runWithoutPTY executes a command without PTY support (fallback).
func (e *Executor) runWithoutPTY(cmd *exec.Cmd) (int, error) {
	cmd.Stdin = e.Stdin
	cmd.Stdout = e.Stdout
	cmd.Stderr = e.Stderr

	sigHandler := NewSignalHandler(cmd.Process, nil)

	if err := cmd.Start(); err != nil {
		return 1, err
	}

	sigHandler.SetProcess(cmd.Process)
	sigHandler.Start()
	defer sigHandler.Stop()

	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	return 0, nil
}

// RunSimple executes a command without PTY support.
// Useful for non-interactive commands.
func (e *Executor) RunSimple(cmd *exec.Cmd) (int, error) {
	return e.runWithoutPTY(cmd)
}
