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

// ============================================================================
// Registry Type Tests
// ============================================================================

func TestNewRegistry(t *testing.T) {
	t.Run("creates empty registry", func(t *testing.T) {
		r := NewRegistry()
		if r == nil {
			t.Fatal("expected non-nil registry")
		}

		names := r.List()
		if len(names) != 0 {
			t.Errorf("expected empty registry, got %d backends", len(names))
		}
	})

	t.Run("registry operations work", func(t *testing.T) {
		r := NewRegistry()

		// Register a backend
		r.Register(&Claude{})
		if len(r.List()) != 1 {
			t.Errorf("expected 1 backend, got %d", len(r.List()))
		}

		// Get the backend
		b, err := r.Get("claude")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.Name() != "claude" {
			t.Errorf("expected 'claude', got %q", b.Name())
		}

		// Unregister
		r.Unregister("claude")
		if len(r.List()) != 0 {
			t.Errorf("expected empty registry after unregister, got %d", len(r.List()))
		}
	})
}

func TestNewRegistryWithDefaults(t *testing.T) {
	r := NewRegistryWithDefaults()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}

	names := r.List()
	if len(names) != 3 {
		t.Errorf("expected 3 default backends, got %d", len(names))
	}

	// Verify all default backends are present
	expected := map[string]bool{"claude": false, "codex": false, "gemini": false}
	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected default backend %q not found", name)
		}
	}
}

func TestRegistryIsolation(t *testing.T) {
	// Create two separate registries
	r1 := NewRegistry()
	r2 := NewRegistry()

	// Register different backends
	r1.Register(&Claude{})
	r2.Register(&Codex{})

	// Verify isolation
	if len(r1.List()) != 1 {
		t.Errorf("r1: expected 1 backend, got %d", len(r1.List()))
	}
	if len(r2.List()) != 1 {
		t.Errorf("r2: expected 1 backend, got %d", len(r2.List()))
	}

	_, err1 := r1.Get("codex")
	if err1 == nil {
		t.Error("r1 should not have codex backend")
	}

	_, err2 := r2.Get("claude")
	if err2 == nil {
		t.Error("r2 should not have claude backend")
	}
}

func TestDefaultRegistry(t *testing.T) {
	r := DefaultRegistry()
	if r == nil {
		t.Fatal("expected non-nil default registry")
	}

	// Verify it's the same as the global functions
	names := r.List()
	globalNames := List()

	if len(names) != len(globalNames) {
		t.Errorf("default registry has %d backends, global has %d", len(names), len(globalNames))
	}
}

func TestRegistryUnregisterAll(t *testing.T) {
	r := NewRegistryWithDefaults()

	// Verify we have backends
	if len(r.List()) == 0 {
		t.Fatal("expected backends before UnregisterAll")
	}

	r.UnregisterAll()

	if len(r.List()) != 0 {
		t.Errorf("expected 0 backends after UnregisterAll, got %d", len(r.List()))
	}
}

func TestRegistryAvailabilityCache(t *testing.T) {
	r := NewRegistry()
	r.Register(&Claude{})

	// First call should cache the result
	_ = r.IsAvailableCached("claude")

	// Invalidate cache
	r.InvalidateAvailabilityCache("claude")

	// Should still work after invalidation
	_ = r.IsAvailableCached("claude")

	// Invalidate all
	r.InvalidateAllAvailabilityCache()

	// Should still work
	_ = r.IsAvailableCached("claude")
}

func TestRegistryGetUnknownBackend(t *testing.T) {
	r := NewRegistry()
	r.Register(&Claude{})

	_, err := r.Get("unknown")
	if err == nil {
		t.Error("expected error for unknown backend")
	}

	// Error message should list available backends
	if !strings.Contains(err.Error(), "claude") {
		t.Errorf("error should list available backends: %v", err)
	}
}
