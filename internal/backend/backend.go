// Package backend provides the interface and implementations for AI CLI backends.
package backend

import (
	"os/exec"
)

// Backend name constants for consistent references across the codebase.
const (
	BackendClaude = "claude"
	BackendCodex  = "codex"
	BackendGemini = "gemini"
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

	// ParseOutput extracts the clean response text from raw CLI output.
	// This normalizes output across different backends.
	ParseOutput(rawOutput string) string

	// ParseJSONResponse extracts a unified response from JSON output.
	// Returns the response content, session ID, and any error.
	ParseJSONResponse(rawOutput string) (*UnifiedResponse, error)

	// SeparateStderr returns true if this backend's stderr should be
	// captured separately (to filter out noise like credential messages).
	SeparateStderr() bool
}

// UnifiedResponse represents a normalized response from any backend.
type UnifiedResponse struct {
	// Content is the main response text.
	Content string `json:"content"`

	// SessionID is the session identifier.
	SessionID string `json:"session_id,omitempty"`

	// Model is the model used for the response.
	Model string `json:"model,omitempty"`

	// DurationMs is the execution duration in milliseconds.
	DurationMs int64 `json:"duration_ms,omitempty"`

	// Usage contains token usage information.
	Usage *TokenUsage `json:"usage,omitempty"`

	// Raw contains the original backend response (for debugging).
	Raw map[string]any `json:"raw,omitempty"`
}

// TokenUsage represents token usage information.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
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
