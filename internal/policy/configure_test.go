package policy

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
)

func TestConfigureFromConfig_InvalidRulesFileFails(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Policy: config.PolicyConfig{
				Enabled:         true,
				Mode:            "shadow",
				FailureMode:     "fail-open",
				DefaultDecision: "allow",
				QuotaStore:      "memory",
				RulesFile:       filepath.Join(t.TempDir(), "missing.yaml"),
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := ConfigureFromConfig(cfg, logger)
	if err == nil {
		t.Fatal("ConfigureFromConfig should fail when rules file is missing")
	}
	if Current() != nil {
		t.Fatal("Current engine should remain nil after failed configure")
	}
}

func TestConfigureFromConfig_LoadsEngine(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	rulesPath := filepath.Join(t.TempDir(), "rules.yaml")
	rulesYAML := `version: v1
rules:
  - id: allow-all
    enabled: true
    priority: 1
    action:
      type: allow
`
	if err := os.WriteFile(rulesPath, []byte(rulesYAML), 0o600); err != nil {
		t.Fatalf("write rules file error = %v", err)
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Policy: config.PolicyConfig{
				Enabled:         true,
				Mode:            "shadow",
				FailureMode:     "fail-open",
				DefaultDecision: "allow",
				QuotaStore:      "memory",
				RulesFile:       rulesPath,
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := ConfigureFromConfig(cfg, logger); err != nil {
		t.Fatalf("ConfigureFromConfig() error = %v", err)
	}

	engine := Current()
	if engine == nil {
		t.Fatal("Current engine should be non-nil after successful configure")
	}
	if engine.runtime.Mode != ModeShadow {
		t.Fatalf("engine mode = %q, want %q", engine.runtime.Mode, ModeShadow)
	}
}
