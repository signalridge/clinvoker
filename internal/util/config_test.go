package util

import (
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
)

func TestApplyUnifiedDefaults(t *testing.T) {
	t.Run("nil opts does nothing", func(t *testing.T) {
		cfg := &config.Config{}
		ApplyUnifiedDefaults(nil, cfg, false)
		// Should not panic
	})

	t.Run("nil config does nothing", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		ApplyUnifiedDefaults(opts, nil, false)
		// Should not panic
	})

	t.Run("applies approval mode from config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			UnifiedFlags: config.UnifiedFlagsConfig{
				ApprovalMode: "auto",
			},
		}
		ApplyUnifiedDefaults(opts, cfg, false)

		if opts.ApprovalMode != "auto" {
			t.Errorf("ApprovalMode = %q, want %q", opts.ApprovalMode, "auto")
		}
	})

	t.Run("does not override explicit approval mode", func(t *testing.T) {
		opts := &backend.UnifiedOptions{
			ApprovalMode: "always",
		}
		cfg := &config.Config{
			UnifiedFlags: config.UnifiedFlagsConfig{
				ApprovalMode: "auto",
			},
		}
		ApplyUnifiedDefaults(opts, cfg, false)

		if opts.ApprovalMode != "always" {
			t.Errorf("ApprovalMode = %q, want %q", opts.ApprovalMode, "always")
		}
	})

	t.Run("applies sandbox mode from config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			UnifiedFlags: config.UnifiedFlagsConfig{
				SandboxMode: "read-only",
			},
		}
		ApplyUnifiedDefaults(opts, cfg, false)

		if opts.SandboxMode != "read-only" {
			t.Errorf("SandboxMode = %q, want %q", opts.SandboxMode, "read-only")
		}
	})

	t.Run("applies max turns from config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			UnifiedFlags: config.UnifiedFlagsConfig{
				MaxTurns: 10,
			},
		}
		ApplyUnifiedDefaults(opts, cfg, false)

		if opts.MaxTurns != 10 {
			t.Errorf("MaxTurns = %d, want 10", opts.MaxTurns)
		}
	})

	t.Run("applies max tokens from config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			UnifiedFlags: config.UnifiedFlagsConfig{
				MaxTokens: 4096,
			},
		}
		ApplyUnifiedDefaults(opts, cfg, false)

		if opts.MaxTokens != 4096 {
			t.Errorf("MaxTokens = %d, want 4096", opts.MaxTokens)
		}
	})

	t.Run("applies verbose from config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			UnifiedFlags: config.UnifiedFlagsConfig{
				Verbose: true,
			},
		}
		ApplyUnifiedDefaults(opts, cfg, false)

		if !opts.Verbose {
			t.Error("expected Verbose to be true")
		}
	})

	t.Run("applies effective dry run", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{}
		ApplyUnifiedDefaults(opts, cfg, true)

		if !opts.DryRun {
			t.Error("expected DryRun to be true")
		}
	})

	t.Run("does not override explicit dry run false with effective true", func(t *testing.T) {
		opts := &backend.UnifiedOptions{DryRun: false}
		cfg := &config.Config{}
		ApplyUnifiedDefaults(opts, cfg, true)

		// effectiveDryRun should set DryRun to true
		if !opts.DryRun {
			t.Error("expected DryRun to be true from effectiveDryRun")
		}
	})
}

func TestApplyOutputFormatDefault(t *testing.T) {
	tests := []struct {
		name    string
		current string
		cfg     *config.Config
		want    string
	}{
		{
			name:    "empty with config default",
			current: "",
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "json",
				},
			},
			want: "json",
		},
		{
			name:    "explicit value not overridden",
			current: "text",
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "json",
				},
			},
			want: "text",
		},
		{
			name:    "default value gets replaced",
			current: "default",
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "json",
				},
			},
			want: "json",
		},
		{
			name:    "nil config returns current",
			current: "text",
			cfg:     nil,
			want:    "text",
		},
		{
			name:    "empty config uses builtin default",
			current: "",
			cfg:     &config.Config{},
			want:    "json",
		},
		{
			name:    "nil config uses builtin default",
			current: "",
			cfg:     nil,
			want:    "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyOutputFormatDefault(tt.current, tt.cfg)
			if got != tt.want {
				t.Errorf("ApplyOutputFormatDefault(%q, cfg) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

func TestInternalOutputFormat(t *testing.T) {
	tests := []struct {
		name      string
		requested backend.OutputFormat
		want      backend.OutputFormat
	}{
		{"empty uses JSON", "", backend.OutputJSON},
		{"default uses JSON", backend.OutputDefault, backend.OutputJSON},
		{"text uses JSON", backend.OutputText, backend.OutputJSON},
		{"json stays json", backend.OutputJSON, backend.OutputJSON},
		{"stream-json stays stream-json", backend.OutputStreamJSON, backend.OutputStreamJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InternalOutputFormat(tt.requested)
			if got != tt.want {
				t.Errorf("InternalOutputFormat(%q) = %q, want %q", tt.requested, got, tt.want)
			}
		})
	}
}

func TestApplyBackendDefaults(t *testing.T) {
	t.Run("nil opts does nothing", func(t *testing.T) {
		cfg := &config.Config{}
		ApplyBackendDefaults(nil, "claude", cfg)
		// Should not panic
	})

	t.Run("nil config does nothing", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		ApplyBackendDefaults(opts, "claude", nil)
		// Should not panic
	})

	t.Run("empty backend name does nothing", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {SystemPrompt: "test"},
			},
		}
		ApplyBackendDefaults(opts, "", cfg)
		if opts.SystemPrompt != "" {
			t.Errorf("expected empty SystemPrompt, got %q", opts.SystemPrompt)
		}
	})

	t.Run("unknown backend does nothing", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {SystemPrompt: "test"},
			},
		}
		ApplyBackendDefaults(opts, "unknown", cfg)
		if opts.SystemPrompt != "" {
			t.Errorf("expected empty SystemPrompt, got %q", opts.SystemPrompt)
		}
	})

	t.Run("applies system prompt from backend config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {SystemPrompt: "You are helpful."},
			},
		}
		ApplyBackendDefaults(opts, "claude", cfg)
		if opts.SystemPrompt != "You are helpful." {
			t.Errorf("SystemPrompt = %q, want %q", opts.SystemPrompt, "You are helpful.")
		}
	})

	t.Run("does not override explicit system prompt", func(t *testing.T) {
		opts := &backend.UnifiedOptions{SystemPrompt: "User prompt"}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {SystemPrompt: "Config prompt"},
			},
		}
		ApplyBackendDefaults(opts, "claude", cfg)
		if opts.SystemPrompt != "User prompt" {
			t.Errorf("SystemPrompt = %q, want %q", opts.SystemPrompt, "User prompt")
		}
	})

	t.Run("appends extra flags from backend config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{ExtraFlags: []string{"--existing"}}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {ExtraFlags: []string{"--config-flag", "--another"}},
			},
		}
		ApplyBackendDefaults(opts, "claude", cfg)
		if len(opts.ExtraFlags) != 3 {
			t.Errorf("ExtraFlags len = %d, want 3", len(opts.ExtraFlags))
		}
		if opts.ExtraFlags[0] != "--existing" {
			t.Errorf("ExtraFlags[0] = %q, want %q", opts.ExtraFlags[0], "--existing")
		}
		if opts.ExtraFlags[1] != "--config-flag" {
			t.Errorf("ExtraFlags[1] = %q, want %q", opts.ExtraFlags[1], "--config-flag")
		}
	})

	t.Run("applies approval mode from backend config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {ApprovalMode: "auto"},
			},
		}
		ApplyBackendDefaults(opts, "claude", cfg)
		if opts.ApprovalMode != "auto" {
			t.Errorf("ApprovalMode = %q, want %q", opts.ApprovalMode, "auto")
		}
	})

	t.Run("applies sandbox mode from backend config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {SandboxMode: "workspace"},
			},
		}
		ApplyBackendDefaults(opts, "claude", cfg)
		if opts.SandboxMode != "workspace" {
			t.Errorf("SandboxMode = %q, want %q", opts.SandboxMode, "workspace")
		}
	})

	t.Run("does not override explicit approval mode", func(t *testing.T) {
		opts := &backend.UnifiedOptions{ApprovalMode: "always"}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {ApprovalMode: "auto"},
			},
		}
		ApplyBackendDefaults(opts, "claude", cfg)
		if opts.ApprovalMode != "always" {
			t.Errorf("ApprovalMode = %q, want %q", opts.ApprovalMode, "always")
		}
	})

	t.Run("overrides default approval mode with backend config", func(t *testing.T) {
		opts := &backend.UnifiedOptions{ApprovalMode: backend.ApprovalDefault}
		cfg := &config.Config{
			Backends: map[string]config.BackendConfig{
				"claude": {ApprovalMode: "auto"},
			},
		}
		ApplyBackendDefaults(opts, "claude", cfg)
		if opts.ApprovalMode != "auto" {
			t.Errorf("ApprovalMode = %q, want %q", opts.ApprovalMode, "auto")
		}
	})
}
