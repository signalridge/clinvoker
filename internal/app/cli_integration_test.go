package app

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
)

func TestCLI_VersionCommand(t *testing.T) {
	// Isolate config initialization from any user config.
	cfgFile = filepath.Join(t.TempDir(), "config.yaml")
	config.Reset()

	rootCmd.SetArgs([]string{"version"})
	defer rootCmd.SetArgs(nil)

	output := captureStdout(t, func() {
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("rootCmd.Execute failed: %v", err)
		}
	})

	if !strings.Contains(output, "clinvk") {
		t.Fatalf("expected version output to contain 'clinvk', got %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = origStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}
	_ = r.Close()

	return string(out)
}
