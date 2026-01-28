package util

import (
	"context"
	"os/exec"
)

// CommandWithContext wraps an existing exec.Cmd with context support for cancellation.
// If ctx or cmd is nil, returns the original cmd unchanged.
// This creates a new CommandContext with the same path, args, dir, env, and other settings.
func CommandWithContext(ctx context.Context, cmd *exec.Cmd) *exec.Cmd {
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

// CleanupContext returns a valid context for cleanup operations.
// If ctx is nil, returns context.Background().
func CleanupContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
