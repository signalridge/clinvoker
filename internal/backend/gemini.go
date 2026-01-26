package backend

import (
	"os/exec"
)

// Gemini implements the Backend interface for Gemini CLI.
type Gemini struct{}

// Name returns the backend identifier.
func (g *Gemini) Name() string {
	return "gemini"
}

// IsAvailable checks if Gemini CLI is installed.
func (g *Gemini) IsAvailable() bool {
	_, err := exec.LookPath("gemini")
	return err == nil
}

// BuildCommand creates an exec.Cmd for running a prompt with Gemini CLI.
func (g *Gemini) BuildCommand(prompt string, opts *Options) *exec.Cmd {
	args := []string{}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	args = append(args, prompt)

	cmd := exec.Command("gemini", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// ResumeCommand creates an exec.Cmd for resuming a Gemini session.
func (g *Gemini) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
	args := []string{"--resume", sessionID}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	if prompt != "" {
		args = append(args, prompt)
	}

	cmd := exec.Command("gemini", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// BuildCommandUnified creates an exec.Cmd using unified options.
func (g *Gemini) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
	return g.BuildCommand(prompt, MapFromUnified(g.Name(), opts))
}

// ResumeCommandUnified creates a resume exec.Cmd using unified options.
func (g *Gemini) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
	return g.ResumeCommand(sessionID, prompt, MapFromUnified(g.Name(), opts))
}
