package util

import (
	"context"
	"os/exec"
	"testing"
)

func TestCommandWithContext(t *testing.T) {
	t.Run("nil context returns original cmd", func(t *testing.T) {
		cmd := exec.Command("echo", "test")
		result := CommandWithContext(nil, cmd)
		if result != cmd {
			t.Error("expected original cmd when ctx is nil")
		}
	})

	t.Run("nil cmd returns nil", func(t *testing.T) {
		ctx := context.Background()
		result := CommandWithContext(ctx, nil)
		if result != nil {
			t.Error("expected nil when cmd is nil")
		}
	})

	t.Run("wraps cmd with context", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.Command("echo", "hello", "world")
		cmd.Dir = "/tmp"

		result := CommandWithContext(ctx, cmd)

		if result == cmd {
			t.Error("expected new cmd, got same instance")
		}
		if result.Path != cmd.Path {
			t.Errorf("Path = %q, want %q", result.Path, cmd.Path)
		}
		if len(result.Args) != len(cmd.Args) {
			t.Errorf("Args length = %d, want %d", len(result.Args), len(cmd.Args))
		}
		if result.Dir != cmd.Dir {
			t.Errorf("Dir = %q, want %q", result.Dir, cmd.Dir)
		}
	})

	t.Run("handles cmd with no args", func(t *testing.T) {
		ctx := context.Background()
		cmd := &exec.Cmd{Path: "/bin/true", Args: []string{}}

		result := CommandWithContext(ctx, cmd)

		if result.Path != cmd.Path {
			t.Errorf("Path = %q, want %q", result.Path, cmd.Path)
		}
	})
}

func TestCleanupContext(t *testing.T) {
	t.Run("nil returns background context", func(t *testing.T) {
		result := CleanupContext(nil)
		if result == nil {
			t.Error("expected non-nil context")
		}
	})

	t.Run("returns same context if not nil", func(t *testing.T) {
		ctx := context.Background()
		result := CleanupContext(ctx)
		if result != ctx {
			t.Error("expected same context")
		}
	})

	t.Run("returns context with cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		result := CleanupContext(ctx)
		if result != ctx {
			t.Error("expected same context")
		}
	})
}
