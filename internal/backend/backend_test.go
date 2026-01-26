package backend

import (
	"strings"
	"testing"
)

func TestRegistry(t *testing.T) {
	t.Run("List returns all backends", func(t *testing.T) {
		names := List()
		if len(names) != 3 {
			t.Errorf("expected 3 backends, got %d", len(names))
		}

		expected := map[string]bool{"claude": false, "codex": false, "gemini": false}
		for _, name := range names {
			if _, ok := expected[name]; ok {
				expected[name] = true
			}
		}

		for name, found := range expected {
			if !found {
				t.Errorf("expected backend %q not found", name)
			}
		}
	})

	t.Run("Get returns backend by name", func(t *testing.T) {
		tests := []struct {
			name    string
			wantErr bool
		}{
			{"claude", false},
			{"codex", false},
			{"gemini", false},
			{"unknown", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				b, err := Get(tt.name)
				if tt.wantErr {
					if err == nil {
						t.Errorf("expected error for backend %q", tt.name)
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if b.Name() != tt.name {
						t.Errorf("expected name %q, got %q", tt.name, b.Name())
					}
				}
			})
		}
	})
}

func TestClaudeBackend(t *testing.T) {
	b := &Claude{}

	t.Run("Name returns claude", func(t *testing.T) {
		if b.Name() != "claude" {
			t.Errorf("expected 'claude', got %q", b.Name())
		}
	})

	t.Run("BuildCommand creates correct command", func(t *testing.T) {
		opts := &Options{
			WorkDir:      "/test/dir",
			Model:        "test-model",
			AllowedTools: "all",
		}

		cmd := b.BuildCommand("test prompt", opts)

		if cmd.Dir != "/test/dir" {
			t.Errorf("expected workdir '/test/dir', got %q", cmd.Dir)
		}

		args := strings.Join(cmd.Args, " ")
		if !strings.Contains(args, "--model test-model") {
			t.Errorf("expected --model flag, got: %s", args)
		}
		if !strings.Contains(args, "--allowedTools all") {
			t.Errorf("expected --allowedTools flag, got: %s", args)
		}
		if !strings.Contains(args, "test prompt") {
			t.Errorf("expected prompt in args, got: %s", args)
		}
	})

	t.Run("ResumeCommand includes session ID", func(t *testing.T) {
		cmd := b.ResumeCommand("session-123", "follow up", nil)

		args := strings.Join(cmd.Args, " ")
		if !strings.Contains(args, "--resume session-123") {
			t.Errorf("expected --resume flag, got: %s", args)
		}
	})
}

func TestCodexBackend(t *testing.T) {
	b := &Codex{}

	t.Run("Name returns codex", func(t *testing.T) {
		if b.Name() != "codex" {
			t.Errorf("expected 'codex', got %q", b.Name())
		}
	})

	t.Run("BuildCommand creates correct command", func(t *testing.T) {
		opts := &Options{
			Model: "o3",
		}

		cmd := b.BuildCommand("test prompt", opts)

		args := strings.Join(cmd.Args, " ")
		if !strings.Contains(args, "--model o3") {
			t.Errorf("expected --model flag, got: %s", args)
		}
	})

	t.Run("ResumeCommand includes session flag", func(t *testing.T) {
		cmd := b.ResumeCommand("session-456", "", nil)

		args := strings.Join(cmd.Args, " ")
		// Codex CLI uses "resume <session-id>" subcommand
		if !strings.Contains(args, "resume session-456") {
			t.Errorf("expected resume subcommand with session, got: %s", args)
		}
	})
}

func TestGeminiBackend(t *testing.T) {
	b := &Gemini{}

	t.Run("Name returns gemini", func(t *testing.T) {
		if b.Name() != "gemini" {
			t.Errorf("expected 'gemini', got %q", b.Name())
		}
	})

	t.Run("ResumeCommand uses --resume flag", func(t *testing.T) {
		cmd := b.ResumeCommand("session-789", "continue", nil)

		args := strings.Join(cmd.Args, " ")
		// Gemini CLI uses "--resume <session-id>"
		if !strings.Contains(args, "--resume session-789") {
			t.Errorf("expected --resume flag, got: %s", args)
		}
	})
}

func TestOptions(t *testing.T) {
	t.Run("Nil options handled gracefully", func(t *testing.T) {
		b := &Claude{}
		cmd := b.BuildCommand("test", nil)

		if cmd == nil {
			t.Error("expected non-nil command")
		}
	})

	t.Run("ExtraFlags are included", func(t *testing.T) {
		b := &Claude{}
		opts := &Options{
			ExtraFlags: []string{"--verbose", "--debug"},
		}

		cmd := b.BuildCommand("test", opts)

		args := strings.Join(cmd.Args, " ")
		if !strings.Contains(args, "--verbose") {
			t.Errorf("expected --verbose flag, got: %s", args)
		}
		if !strings.Contains(args, "--debug") {
			t.Errorf("expected --debug flag, got: %s", args)
		}
	})
}
