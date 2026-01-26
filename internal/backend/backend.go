// Package backend provides the interface and implementations for AI CLI backends.
package backend

import (
	"os/exec"
)

// Backend defines the interface for AI CLI backends.
type Backend interface {
	// Name returns the backend identifier.
	Name() string

	// IsAvailable checks if the backend CLI is installed and accessible.
	IsAvailable() bool

	// BuildCommand creates an exec.Cmd for running a prompt.
	BuildCommand(prompt string, opts *Options) *exec.Cmd

	// ResumeCommand creates an exec.Cmd for resuming a session.
	ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd

	// BuildCommandUnified creates an exec.Cmd using unified options.
	BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd

	// ResumeCommandUnified creates a resume exec.Cmd using unified options.
	ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd
}

// Options contains configuration for backend commands.
type Options struct {
	// WorkDir is the working directory for the command.
	WorkDir string

	// Model specifies the model to use.
	Model string

	// AllowedTools specifies allowed tools (backend-specific).
	AllowedTools string

	// AllowedDirs specifies allowed directories (backend-specific).
	AllowedDirs []string

	// ExtraFlags contains additional flags to pass to the backend.
	ExtraFlags []string
}
