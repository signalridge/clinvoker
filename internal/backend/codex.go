package backend

import (
	"os/exec"
)

// Codex implements the Backend interface for Codex CLI.
type Codex struct{}

// Name returns the backend identifier.
func (c *Codex) Name() string {
	return "codex"
}

// IsAvailable checks if Codex CLI is installed.
func (c *Codex) IsAvailable() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

// BuildCommand creates an exec.Cmd for running a prompt with Codex CLI.
func (c *Codex) BuildCommand(prompt string, opts *Options) *exec.Cmd {
	args := []string{}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	args = append(args, prompt)

	cmd := exec.Command("codex", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// ResumeCommand creates an exec.Cmd for resuming a Codex session.
func (c *Codex) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
	args := []string{"resume", sessionID}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	if prompt != "" {
		args = append(args, prompt)
	}

	cmd := exec.Command("codex", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// BuildCommandUnified creates an exec.Cmd using unified options.
func (c *Codex) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
	return c.BuildCommand(prompt, MapFromUnified(c.Name(), opts))
}

// ResumeCommandUnified creates a resume exec.Cmd using unified options.
func (c *Codex) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
	return c.ResumeCommand(sessionID, prompt, MapFromUnified(c.Name(), opts))
}
