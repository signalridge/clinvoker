package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
)

func TestConfigLintCmd_ValidConfig(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	configLintFile = ""
	configLintJSON = false

	output := captureStdout(t, func() {
		if err := configLintCmd.RunE(configLintCmd, nil); err != nil {
			t.Fatalf("config lint should succeed on valid default config: %v", err)
		}
	})

	if !strings.Contains(output, "Configuration is valid.") {
		t.Fatalf("expected valid config message, got: %s", output)
	}
}

func TestConfigLintCmd_InvalidConfigFile(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.yaml")
	content := "default_backend: unknown\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	configLintFile = path
	configLintJSON = true

	output := captureStdout(t, func() {
		err := configLintCmd.RunE(configLintCmd, nil)
		if err == nil {
			t.Fatal("expected config lint to fail")
		}
		if !errors.Is(err, ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed, got %v", err)
		}
	})

	var report configLintReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Fatalf("parse lint json output: %v\noutput=%s", err, output)
	}
	if report.Valid {
		t.Fatalf("report.valid = true, want false")
	}
	if report.ErrorCount == 0 {
		t.Fatalf("expected at least one lint error")
	}
}
