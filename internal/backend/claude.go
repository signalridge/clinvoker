package backend

import (
	"os/exec"
)

// Claude implements the Backend interface for Claude Code CLI.
type Claude struct{}

// Name returns the backend identifier.
func (c *Claude) Name() string {
	return "claude"
}

// IsAvailable checks if Claude Code CLI is installed.
func (c *Claude) IsAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// BuildCommand creates an exec.Cmd for running a prompt with Claude Code.
func (c *Claude) BuildCommand(prompt string, opts *Options) *exec.Cmd {
	args := []string{"--print"}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		if opts.AllowedTools != "" {
			args = append(args, "--allowedTools", opts.AllowedTools)
		}
		for _, dir := range opts.AllowedDirs {
			args = append(args, "--add-dir", dir)
		}
		args = append(args, opts.ExtraFlags...)
	}

	args = append(args, prompt)

	cmd := exec.Command("claude", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// ResumeCommand creates an exec.Cmd for resuming a Claude Code session.
func (c *Claude) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
	args := []string{"--resume", sessionID, "--print"}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	if prompt != "" {
		args = append(args, prompt)
	}

	cmd := exec.Command("claude", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// BuildCommandUnified creates an exec.Cmd using unified options.
func (c *Claude) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
	return c.BuildCommand(prompt, MapFromUnified(c.Name(), opts))
}

// ResumeCommandUnified creates a resume exec.Cmd using unified options.
func (c *Claude) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
	return c.ResumeCommand(sessionID, prompt, MapFromUnified(c.Name(), opts))
}
