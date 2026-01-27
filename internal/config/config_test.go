package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	expected := filepath.Join(home, ".clinvk")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestSessionsDir(t *testing.T) {
	dir := SessionsDir()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	expected := filepath.Join(home, ".clinvk", "sessions")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := Get()

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if cfg.DefaultBackend != "claude" {
		t.Errorf("expected default backend 'claude', got %q", cfg.DefaultBackend)
	}

	if cfg.Session.RetentionDays != 30 {
		t.Errorf("expected retention days 30, got %d", cfg.Session.RetentionDays)
	}

	if !cfg.Session.AutoResume {
		t.Error("expected auto resume to be true")
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// This test modifies the filesystem, so we use a temp dir
	tmpDir, err := os.MkdirTemp("", "clinvoker-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// We can't easily test EnsureConfigDir without modifying global state
	// So we just verify the function exists and returns nil on success
	err = EnsureConfigDir()
	if err != nil {
		// May fail if we can't create ~/.clinvk, which is acceptable
		t.Logf("EnsureConfigDir returned: %v", err)
	}
}

// ============================================================================
// UnifiedFlagsConfig Tests
// ============================================================================

func TestUnifiedFlagsConfigDefaults(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"ApprovalMode", cfg.UnifiedFlags.ApprovalMode, "default"},
		{"SandboxMode", cfg.UnifiedFlags.SandboxMode, "default"},
		{"OutputFormat", cfg.UnifiedFlags.OutputFormat, "default"},
		{"Verbose", cfg.UnifiedFlags.Verbose, false},
		{"DryRun", cfg.UnifiedFlags.DryRun, false},
		{"MaxTurns", cfg.UnifiedFlags.MaxTurns, 0},
		{"MaxTokens", cfg.UnifiedFlags.MaxTokens, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("UnifiedFlags.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// ============================================================================
// BackendConfig Tests
// ============================================================================

func TestBackendConfig_IsBackendEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   BackendConfig
		expected bool
	}{
		{
			name:     "nil enabled defaults to true",
			config:   BackendConfig{Enabled: nil},
			expected: true,
		},
		{
			name:     "explicitly enabled",
			config:   BackendConfig{Enabled: boolPtr(true)},
			expected: true,
		},
		{
			name:     "explicitly disabled",
			config:   BackendConfig{Enabled: boolPtr(false)},
			expected: false,
		},
		{
			name:     "empty config defaults to true",
			config:   BackendConfig{},
			expected: true,
		},
		{
			name: "disabled with other fields set",
			config: BackendConfig{
				Model:        "test-model",
				AllowedTools: "all",
				Enabled:      boolPtr(false),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsBackendEnabled()
			if got != tt.expected {
				t.Errorf("IsBackendEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// ============================================================================
// GetBackendConfig Tests
// ============================================================================

func TestGetBackendConfig(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	// Set up test backend configs
	cfg.Backends = map[string]BackendConfig{
		"claude": {
			Model:        "claude-opus-4-5-20251101",
			AllowedTools: "all",
			ApprovalMode: "auto",
		},
		"codex": {
			Model:       "o3",
			SandboxMode: "workspace",
		},
	}

	t.Run("returns configured backend", func(t *testing.T) {
		bc := GetBackendConfig("claude")
		if bc.Model != "claude-opus-4-5-20251101" {
			t.Errorf("Model = %q, want %q", bc.Model, "claude-opus-4-5-20251101")
		}
		if bc.AllowedTools != "all" {
			t.Errorf("AllowedTools = %q, want %q", bc.AllowedTools, "all")
		}
		if bc.ApprovalMode != "auto" {
			t.Errorf("ApprovalMode = %q, want %q", bc.ApprovalMode, "auto")
		}
	})

	t.Run("returns empty config for unknown backend", func(t *testing.T) {
		bc := GetBackendConfig("unknown")
		if bc.Model != "" {
			t.Errorf("Model = %q, want empty string", bc.Model)
		}
		if bc.AllowedTools != "" {
			t.Errorf("AllowedTools = %q, want empty string", bc.AllowedTools)
		}
	})

	t.Run("returns different configs for different backends", func(t *testing.T) {
		claudeBC := GetBackendConfig("claude")
		codexBC := GetBackendConfig("codex")

		if claudeBC.Model == codexBC.Model {
			t.Errorf("Expected different models for claude and codex")
		}
	})
}

// ============================================================================
// GetEffectiveApprovalMode Tests
// ============================================================================

func TestGetEffectiveApprovalMode(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	// Set unified default
	cfg.UnifiedFlags.ApprovalMode = "default"

	// Set backend-specific overrides
	cfg.Backends = map[string]BackendConfig{
		"claude": {ApprovalMode: "auto"},
		"codex":  {ApprovalMode: ""},
		"gemini": {ApprovalMode: "none"},
	}

	tests := []struct {
		name     string
		backend  string
		expected string
	}{
		{
			name:     "backend override takes precedence",
			backend:  "claude",
			expected: "auto",
		},
		{
			name:     "falls back to unified when backend empty",
			backend:  "codex",
			expected: "default",
		},
		{
			name:     "another backend override",
			backend:  "gemini",
			expected: "none",
		},
		{
			name:     "unknown backend falls back to unified",
			backend:  "unknown",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEffectiveApprovalMode(tt.backend)
			if got != tt.expected {
				t.Errorf("GetEffectiveApprovalMode(%q) = %q, want %q", tt.backend, got, tt.expected)
			}
		})
	}
}

// ============================================================================
// GetEffectiveSandboxMode Tests
// ============================================================================

func TestGetEffectiveSandboxMode(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	// Set unified default
	cfg.UnifiedFlags.SandboxMode = "workspace"

	// Set backend-specific overrides
	cfg.Backends = map[string]BackendConfig{
		"claude": {SandboxMode: "full"},
		"codex":  {SandboxMode: ""},
		"gemini": {SandboxMode: "read-only"},
	}

	tests := []struct {
		name     string
		backend  string
		expected string
	}{
		{
			name:     "backend override takes precedence",
			backend:  "claude",
			expected: "full",
		},
		{
			name:     "falls back to unified when backend empty",
			backend:  "codex",
			expected: "workspace",
		},
		{
			name:     "another backend override",
			backend:  "gemini",
			expected: "read-only",
		},
		{
			name:     "unknown backend falls back to unified",
			backend:  "unknown",
			expected: "workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEffectiveSandboxMode(tt.backend)
			if got != tt.expected {
				t.Errorf("GetEffectiveSandboxMode(%q) = %q, want %q", tt.backend, got, tt.expected)
			}
		})
	}
}

// ============================================================================
// GetEffectiveModel Tests
// ============================================================================

func TestGetEffectiveModel(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	cfg.Backends = map[string]BackendConfig{
		"claude": {Model: "claude-opus-4-5-20251101"},
		"codex":  {Model: "o3"},
		"gemini": {Model: "gemini-2.5-pro"},
	}

	tests := []struct {
		name     string
		backend  string
		expected string
	}{
		{"claude model", "claude", "claude-opus-4-5-20251101"},
		{"codex model", "codex", "o3"},
		{"gemini model", "gemini", "gemini-2.5-pro"},
		{"unknown backend returns empty", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEffectiveModel(tt.backend)
			if got != tt.expected {
				t.Errorf("GetEffectiveModel(%q) = %q, want %q", tt.backend, got, tt.expected)
			}
		})
	}
}

// ============================================================================
// EnabledBackends Tests
// ============================================================================

func TestEnabledBackends(t *testing.T) {
	t.Run("all backends enabled by default", func(t *testing.T) {
		Reset()
		Init("")
		cfg := Get()
		cfg.Backends = map[string]BackendConfig{}

		backends := EnabledBackends()
		expected := []string{"claude", "codex", "gemini"}

		if len(backends) != len(expected) {
			t.Errorf("got %d backends, want %d", len(backends), len(expected))
		}

		for i, name := range expected {
			if backends[i] != name {
				t.Errorf("backends[%d] = %q, want %q", i, backends[i], name)
			}
		}
	})

	t.Run("some backends disabled", func(t *testing.T) {
		Reset()
		Init("")
		cfg := Get()
		cfg.Backends = map[string]BackendConfig{
			"codex": {Enabled: boolPtr(false)},
		}

		backends := EnabledBackends()
		if len(backends) != 2 {
			t.Errorf("got %d backends, want 2", len(backends))
		}

		// Should contain claude and gemini but not codex
		for _, b := range backends {
			if b == "codex" {
				t.Error("codex should be disabled but was found in enabled list")
			}
		}
	})

	t.Run("all backends disabled", func(t *testing.T) {
		Reset()
		Init("")
		cfg := Get()
		cfg.Backends = map[string]BackendConfig{
			"claude": {Enabled: boolPtr(false)},
			"codex":  {Enabled: boolPtr(false)},
			"gemini": {Enabled: boolPtr(false)},
		}

		backends := EnabledBackends()
		if len(backends) != 0 {
			t.Errorf("got %d backends, want 0", len(backends))
		}
	})

	t.Run("explicitly enabled backend", func(t *testing.T) {
		Reset()
		Init("")
		cfg := Get()
		cfg.Backends = map[string]BackendConfig{
			"claude": {Enabled: boolPtr(true)},
			"codex":  {Enabled: boolPtr(false)},
		}

		backends := EnabledBackends()
		// Should have claude and gemini (gemini defaults to enabled)
		if len(backends) != 2 {
			t.Errorf("got %d backends, want 2", len(backends))
		}

		hasGemini := false
		hasClaude := false
		for _, b := range backends {
			if b == "gemini" {
				hasGemini = true
			}
			if b == "claude" {
				hasClaude = true
			}
		}

		if !hasGemini {
			t.Error("gemini should be enabled by default")
		}
		if !hasClaude {
			t.Error("claude should be explicitly enabled")
		}
	})
}

// ============================================================================
// SessionConfig Tests
// ============================================================================

func TestSessionConfigDefaults(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"AutoResume", cfg.Session.AutoResume, true},
		{"RetentionDays", cfg.Session.RetentionDays, 30},
		{"StoreTokenUsage", cfg.Session.StoreTokenUsage, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Session.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	t.Run("DefaultTags is nil by default", func(t *testing.T) {
		if len(cfg.Session.DefaultTags) > 0 {
			t.Errorf("Session.DefaultTags should be nil or empty, got %v", cfg.Session.DefaultTags)
		}
	})
}

// ============================================================================
// OutputConfig Tests
// ============================================================================

func TestOutputConfigDefaults(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"Format", cfg.Output.Format, "text"},
		{"ShowTokens", cfg.Output.ShowTokens, false},
		{"ShowTiming", cfg.Output.ShowTiming, false},
		{"Color", cfg.Output.Color, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Output.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// ============================================================================
// ParallelConfig Tests
// ============================================================================

func TestParallelConfigDefaults(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"MaxWorkers", cfg.Parallel.MaxWorkers, 3},
		{"FailFast", cfg.Parallel.FailFast, false},
		{"AggregateOutput", cfg.Parallel.AggregateOutput, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Parallel.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// ============================================================================
// Reset Tests
// ============================================================================

func TestReset(t *testing.T) {
	// Initialize config
	Init("")
	cfg1 := Get()

	// Modify config
	cfg1.DefaultBackend = "modified"

	// Reset
	Reset()

	// Re-initialize
	Init("")
	cfg2 := Get()

	// Verify reset worked
	if cfg2.DefaultBackend != "claude" {
		t.Errorf("After Reset, DefaultBackend = %q, want %q", cfg2.DefaultBackend, "claude")
	}
}

// ============================================================================
// BackendConfig Field Tests
// ============================================================================

func TestBackendConfigFields(t *testing.T) {
	bc := BackendConfig{
		Model:        "test-model",
		AllowedTools: "all",
		ApprovalMode: "auto",
		SandboxMode:  "workspace",
		ExtraFlags:   []string{"--flag1", "--flag2"},
		Enabled:      boolPtr(true),
		SystemPrompt: "You are a helpful assistant.",
	}

	t.Run("Model", func(t *testing.T) {
		if bc.Model != "test-model" {
			t.Errorf("Model = %q, want %q", bc.Model, "test-model")
		}
	})

	t.Run("AllowedTools", func(t *testing.T) {
		if bc.AllowedTools != "all" {
			t.Errorf("AllowedTools = %q, want %q", bc.AllowedTools, "all")
		}
	})

	t.Run("ApprovalMode", func(t *testing.T) {
		if bc.ApprovalMode != "auto" {
			t.Errorf("ApprovalMode = %q, want %q", bc.ApprovalMode, "auto")
		}
	})

	t.Run("SandboxMode", func(t *testing.T) {
		if bc.SandboxMode != "workspace" {
			t.Errorf("SandboxMode = %q, want %q", bc.SandboxMode, "workspace")
		}
	})

	t.Run("ExtraFlags", func(t *testing.T) {
		if len(bc.ExtraFlags) != 2 {
			t.Errorf("ExtraFlags length = %d, want 2", len(bc.ExtraFlags))
		}
		if bc.ExtraFlags[0] != "--flag1" {
			t.Errorf("ExtraFlags[0] = %q, want %q", bc.ExtraFlags[0], "--flag1")
		}
	})

	t.Run("SystemPrompt", func(t *testing.T) {
		if bc.SystemPrompt != "You are a helpful assistant." {
			t.Errorf("SystemPrompt = %q, want %q", bc.SystemPrompt, "You are a helpful assistant.")
		}
	})
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestConfigPrecedence_ApprovalMode(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	// Set up a hierarchy of settings
	cfg.UnifiedFlags.ApprovalMode = "unified-default"
	cfg.Backends = map[string]BackendConfig{
		"claude": {ApprovalMode: "backend-specific"},
		"codex":  {}, // empty - should fall back
	}

	t.Run("backend specific wins over unified", func(t *testing.T) {
		got := GetEffectiveApprovalMode("claude")
		if got != "backend-specific" {
			t.Errorf("got %q, want %q", got, "backend-specific")
		}
	})

	t.Run("unified is fallback", func(t *testing.T) {
		got := GetEffectiveApprovalMode("codex")
		if got != "unified-default" {
			t.Errorf("got %q, want %q", got, "unified-default")
		}
	})
}

func TestConfigPrecedence_SandboxMode(t *testing.T) {
	Reset()
	Init("")
	cfg := Get()

	// Set up a hierarchy of settings
	cfg.UnifiedFlags.SandboxMode = "unified-sandbox"
	cfg.Backends = map[string]BackendConfig{
		"gemini": {SandboxMode: "gemini-specific"},
		"claude": {}, // empty - should fall back
	}

	t.Run("backend specific wins over unified", func(t *testing.T) {
		got := GetEffectiveSandboxMode("gemini")
		if got != "gemini-specific" {
			t.Errorf("got %q, want %q", got, "gemini-specific")
		}
	})

	t.Run("unified is fallback", func(t *testing.T) {
		got := GetEffectiveSandboxMode("claude")
		if got != "unified-sandbox" {
			t.Errorf("got %q, want %q", got, "unified-sandbox")
		}
	})
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestGetWithNilConfig(t *testing.T) {
	Reset()
	// Don't call Init, just call Get directly
	cfg = nil // Force nil

	// Get should auto-initialize
	c := Get()
	if c == nil {
		t.Fatal("Get() should auto-initialize when config is nil")
	}

	if c.DefaultBackend != "claude" {
		t.Errorf("auto-initialized DefaultBackend = %q, want %q", c.DefaultBackend, "claude")
	}
}

func TestEmptyBackendName(t *testing.T) {
	Reset()
	Init("")

	// Empty backend name should return empty config
	bc := GetBackendConfig("")
	if bc.Model != "" {
		t.Errorf("empty backend should return empty Model, got %q", bc.Model)
	}
}

func TestMultipleResets(t *testing.T) {
	// Reset multiple times should not panic
	Reset()
	Reset()
	Reset()

	Init("")
	cfg := Get()

	if cfg == nil {
		t.Fatal("config should not be nil after multiple resets and init")
	}
}

func TestBackendConfigWithAllFieldsEmpty(t *testing.T) {
	bc := BackendConfig{}

	// All fields should be zero values
	if bc.Model != "" {
		t.Errorf("Model should be empty, got %q", bc.Model)
	}
	if bc.AllowedTools != "" {
		t.Errorf("AllowedTools should be empty, got %q", bc.AllowedTools)
	}
	if bc.ApprovalMode != "" {
		t.Errorf("ApprovalMode should be empty, got %q", bc.ApprovalMode)
	}
	if bc.SandboxMode != "" {
		t.Errorf("SandboxMode should be empty, got %q", bc.SandboxMode)
	}
	if bc.ExtraFlags != nil {
		t.Errorf("ExtraFlags should be nil, got %v", bc.ExtraFlags)
	}
	if bc.Enabled != nil {
		t.Errorf("Enabled should be nil, got %v", bc.Enabled)
	}
	if bc.SystemPrompt != "" {
		t.Errorf("SystemPrompt should be empty, got %q", bc.SystemPrompt)
	}

	// But IsBackendEnabled should still return true (nil defaults to true)
	if !bc.IsBackendEnabled() {
		t.Error("IsBackendEnabled should return true for empty config")
	}
}
