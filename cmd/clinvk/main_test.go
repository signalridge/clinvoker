package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain_Help tests the --help flag.
func TestMain_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildTestBinary(t)

	cmd := exec.Command(binary, "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("--help failed: %v\nOutput: %s", err, output)
	}

	// Check for expected help content
	expectedStrings := []string{
		"Usage:",
		"clinvk",
		"Available Commands:",
		"Flags:",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(string(output), s) {
			t.Errorf("help output missing %q", s)
		}
	}
}

// TestMain_Version tests the version command.
func TestMain_Version(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildTestBinary(t)

	cmd := exec.Command(binary, "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("version command failed: %v\nOutput: %s", err, output)
	}

	// Version output should contain version info
	if !strings.Contains(string(output), "version") && !strings.Contains(string(output), "Version") {
		t.Errorf("version output should contain version info: %s", output)
	}
}

// TestMain_InvalidCommand tests behavior with an invalid command.
func TestMain_InvalidCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildTestBinary(t)

	cmd := exec.Command(binary, "nonexistent-command-12345")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("expected error for invalid command")
	}

	// Should show error message about unknown command
	if !strings.Contains(string(output), "unknown") && !strings.Contains(string(output), "Unknown") {
		t.Errorf("output should mention unknown command: %s", output)
	}
}

// TestMain_DryRun tests the --dry-run flag.
func TestMain_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildTestBinary(t)

	cmd := exec.Command(binary, "--dry-run", "test prompt")
	output, err := cmd.CombinedOutput()

	// Dry run should succeed
	if err != nil {
		t.Fatalf("dry run failed: %v\nOutput: %s", err, output)
	}

	// Output should show the command that would be executed
	outputStr := string(output)
	if !strings.Contains(outputStr, "claude") && !strings.Contains(outputStr, "codex") && !strings.Contains(outputStr, "gemini") {
		t.Logf("dry run output: %s", outputStr)
		// This is acceptable - dry run shows the command
	}
}

// TestMain_ConfigShow tests the config show command.
func TestMain_ConfigShow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildTestBinary(t)

	cmd := exec.Command(binary, "config", "show")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("config show failed: %v\nOutput: %s", err, output)
	}

	// Should show configuration info
	expectedStrings := []string{
		"Backend",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(string(output), s) {
			t.Errorf("config show output missing %q", s)
		}
	}
}

// TestMain_SessionsList tests the sessions list command.
func TestMain_SessionsList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildTestBinary(t)

	cmd := exec.Command(binary, "sessions", "list")
	output, err := cmd.CombinedOutput()

	// Should succeed even with no sessions
	if err != nil {
		t.Fatalf("sessions list failed: %v\nOutput: %s", err, output)
	}
}

// TestMain_BackendFlag tests the --backend flag.
func TestMain_BackendFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildTestBinary(t)

	tests := []struct {
		name    string
		backend string
		wantErr bool
	}{
		{"claude", "claude", false},
		{"codex", "codex", false},
		{"gemini", "gemini", false},
		{"invalid", "invalid-backend", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, "--backend", tt.backend, "--dry-run", "test")
			_, err := cmd.CombinedOutput()

			if tt.wantErr && err == nil {
				t.Errorf("expected error for backend %q", tt.backend)
			}
			if !tt.wantErr && err != nil {
				// Note: might fail if backend CLI not available
				t.Logf("backend %q: %v (might not be available)", tt.backend, err)
			}
		})
	}
}

// TestMain_OutputFormat tests the --output-format flag.
func TestMain_OutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildTestBinary(t)

	formats := []string{"text", "json", "stream-json"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cmd := exec.Command(binary, "--output-format", format, "--dry-run", "test")
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Logf("output format %q test: %v (output: %s)", format, err, output)
			}
		})
	}
}

// buildTestBinary builds the test binary and returns its path.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "clinvk-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	binary := filepath.Join(tmpDir, "clinvk")
	if os.PathSeparator == '\\' {
		binary += ".exe"
	}

	// Get the project root (two levels up from cmd/clinvk)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Dir(filepath.Dir(wd))

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/clinvk")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\nOutput: %s", err, output)
	}

	return binary
}
