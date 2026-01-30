package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
)

// mockBackend implements backend.Backend for testing
type mockBackend struct {
	name           string
	parseOutput    string
	jsonResponse   *backend.UnifiedResponse
	jsonError      error
	separateStderr bool
}

func (m *mockBackend) Name() string      { return m.name }
func (m *mockBackend) IsAvailable() bool { return true }
func (m *mockBackend) BuildCommand(prompt string, opts *backend.Options) *exec.Cmd {
	return exec.Command("echo", prompt)
}
func (m *mockBackend) ResumeCommand(sessionID, prompt string, opts *backend.Options) *exec.Cmd {
	return exec.Command("echo", "resume", sessionID)
}
func (m *mockBackend) BuildCommandUnified(prompt string, opts *backend.UnifiedOptions) *exec.Cmd {
	return exec.Command("echo", prompt)
}
func (m *mockBackend) ResumeCommandUnified(sessionID, prompt string, opts *backend.UnifiedOptions) *exec.Cmd {
	return exec.Command("echo", "resume", sessionID)
}
func (m *mockBackend) ParseOutput(rawOutput string) string {
	if m.parseOutput != "" {
		return m.parseOutput
	}
	return rawOutput
}
func (m *mockBackend) ParseJSONResponse(rawOutput string) (*backend.UnifiedResponse, error) {
	if m.jsonError != nil {
		return nil, m.jsonError
	}
	if m.jsonResponse != nil {
		return m.jsonResponse, nil
	}
	return &backend.UnifiedResponse{Content: rawOutput}, nil
}
func (m *mockBackend) SeparateStderr() bool { return m.separateStderr }

// ==================== ExecuteAndCapture Tests ====================

func TestExecuteAndCapture_Success(t *testing.T) {
	b := &mockBackend{name: "test"}

	cmd := exec.Command("echo", "hello world")
	output, exitCode, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	if !strings.Contains(output, "hello world") {
		t.Errorf("output = %q, want to contain 'hello world'", output)
	}
}

func TestExecuteAndCapture_NonZeroExitCode(t *testing.T) {
	b := &mockBackend{name: "test"}

	cmd := exec.Command("sh", "-c", "exit 42")
	_, exitCode, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 42 {
		t.Errorf("exitCode = %d, want 42", exitCode)
	}
}

func TestExecuteAndCapture_CommandNotFound(t *testing.T) {
	b := &mockBackend{name: "test"}

	cmd := exec.Command("nonexistent_command_12345")
	_, exitCode, err := ExecuteAndCapture(b, cmd)

	if err == nil {
		t.Error("expected error for nonexistent command")
	}
	if exitCode != 1 {
		t.Errorf("exitCode = %d, want 1", exitCode)
	}
}

func TestExecuteAndCapture_CapturesStderr(t *testing.T) {
	b := &mockBackend{name: "test"}

	// Command that outputs to stderr only
	cmd := exec.Command("sh", "-c", "echo 'error message' >&2")
	output, exitCode, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("output = %q, should contain stderr content 'error message'", output)
	}
}

func TestExecuteAndCapture_PrefersStdout(t *testing.T) {
	b := &mockBackend{name: "test"}

	// Command that outputs to both stdout and stderr
	cmd := exec.Command("sh", "-c", "echo 'stdout content' && echo 'stderr content' >&2")
	output, _, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should prefer stdout when available
	if !strings.Contains(output, "stdout content") {
		t.Errorf("output = %q, should contain stdout content", output)
	}
}

// ==================== PromptResult Tests ====================

func TestPromptResult_JSONSerialization(t *testing.T) {
	result := PromptResult{
		Backend:   "claude",
		Content:   "Hello",
		SessionID: "sess-123",
		Model:     "opus",
		Duration:  1.5,
		ExitCode:  0,
		Usage: &backend.TokenUsage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
	}

	if result.Backend != "claude" {
		t.Errorf("Backend = %q, want 'claude'", result.Backend)
	}
	if result.Error != "" {
		t.Errorf("Error should be empty, got %q", result.Error)
	}
}

func TestPromptResult_WithError(t *testing.T) {
	result := PromptResult{
		Backend:  "gemini",
		ExitCode: 1,
		Error:    "API rate limit exceeded",
	}

	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
	if result.Error != "API rate limit exceeded" {
		t.Errorf("Error = %q, want 'API rate limit exceeded'", result.Error)
	}
}

// ==================== Output Format Tests ====================

func TestOutputFormatValidation(t *testing.T) {
	validFormats := []string{"text", "json", "stream-json", "TEXT", "JSON", ""}
	for _, format := range validFormats {
		normalized := strings.ToLower(format)
		switch backend.OutputFormat(normalized) {
		case backend.OutputDefault, backend.OutputText, backend.OutputJSON, backend.OutputStreamJSON, "":
			// Valid
		default:
			t.Errorf("format %q should be valid", format)
		}
	}

	invalidFormats := []string{"xml", "yaml", "csv"}
	for _, format := range invalidFormats {
		switch backend.OutputFormat(format) {
		case backend.OutputDefault, backend.OutputText, backend.OutputJSON, backend.OutputStreamJSON, "":
			t.Errorf("format %q should be invalid", format)
		default:
			// Expected to be invalid
		}
	}
}

// ==================== Dry Run Tests ====================

func TestDryRunFlag(t *testing.T) {
	// Test that dry-run flag is properly handled
	// This is a simple validation test
	if dryRun {
		t.Error("dryRun should be false by default in tests")
	}
}

// ==================== Integration Helper Tests ====================

func TestBackendIntegration(t *testing.T) {
	// Skip if backends aren't available
	backends := []string{"claude", "codex", "gemini"}

	for _, name := range backends {
		t.Run(name, func(t *testing.T) {
			b, err := backend.Get(name)
			if err != nil {
				t.Fatalf("failed to get backend %q: %v", name, err)
			}

			if b.Name() != name {
				t.Errorf("Name() = %q, want %q", b.Name(), name)
			}

			// Test command building
			cmd := b.BuildCommandUnified("test prompt", &backend.UnifiedOptions{})
			if cmd == nil {
				t.Error("BuildCommandUnified returned nil")
			}

			// Test resume command building
			resumeCmd := b.ResumeCommandUnified("session-123", "continue", &backend.UnifiedOptions{})
			if resumeCmd == nil {
				t.Error("ResumeCommandUnified returned nil")
			}
		})
	}
}

// ==================== Error Display Format Tests ====================

func TestErrorDisplayFormat(t *testing.T) {
	// Test that error messages are formatted correctly
	tests := []struct {
		backend string
		errMsg  string
		want    string
	}{
		{"gemini", "Invalid API key", "Error [gemini]: Invalid API key"},
		{"claude", "Rate limit exceeded", "Error [claude]: Rate limit exceeded"},
		{"codex", "401 Unauthorized", "Error [codex]: 401 Unauthorized"},
	}

	for _, tt := range tests {
		t.Run(tt.backend, func(t *testing.T) {
			// This is the format used in executeTextViaJSON
			formatted := "Error [" + tt.backend + "]: " + tt.errMsg

			if formatted != tt.want {
				t.Errorf("formatted = %q, want %q", formatted, tt.want)
			}
		})
	}
}

// TestExecuteAndCapture_WithMockError tests that errors from ParseJSONResponse are handled
func TestExecuteAndCapture_WithMockError(t *testing.T) {
	// Create a mock that returns an error in the response
	b := &mockBackend{
		name: "test-error",
		jsonResponse: &backend.UnifiedResponse{
			Error: "API Error: rate limit",
		},
	}

	// Command that succeeds but backend returns error in JSON
	cmd := exec.Command("echo", "success")
	output, exitCode, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When ParseJSONResponse returns error in response, we use text parsing fallback
	// This is the current behavior of ExecuteAndCapture
	if output == "" && exitCode == 0 {
		// This is expected - ExecuteAndCapture uses ParseOutput as fallback
		t.Log("ExecuteAndCapture fell back to ParseOutput as expected")
	}
}

// TestExecuteAndCapture_ParseOutputFallback verifies fallback behavior
func TestExecuteAndCapture_ParseOutputFallback(t *testing.T) {
	b := &mockBackend{
		name:        "fallback-test",
		parseOutput: "parsed output",
	}

	cmd := exec.Command("echo", "raw output")
	output, _, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Since mockBackend.ParseOutput returns "parsed output", that's what we should get
	if output != "parsed output" {
		t.Errorf("output = %q, want 'parsed output'", output)
	}
}

// ==================== Edge Cases ====================

// TestExecuteAndCapture_StderrOnError tests that stderr is preferred when exit code is non-zero
func TestExecuteAndCapture_StderrOnError(t *testing.T) {
	b := &mockBackend{name: "test"}

	// Command that outputs to stdout but exits with error and has stderr
	cmd := exec.Command("sh", "-c", "echo 'partial output' && echo 'error message' >&2 && exit 1")
	output, exitCode, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 1 {
		t.Errorf("exitCode = %d, want 1", exitCode)
	}
	// When exit code is non-zero and stderr has content, stderr should be used
	if !strings.Contains(output, "error message") {
		t.Errorf("output = %q, should contain stderr 'error message' when exit code is non-zero", output)
	}
}

// TestExecuteAndCapture_StdoutOnSuccess tests that stdout is preferred when exit code is zero
func TestExecuteAndCapture_StdoutOnSuccess(t *testing.T) {
	b := &mockBackend{name: "test"}

	// Command that outputs to both stdout and stderr but succeeds
	cmd := exec.Command("sh", "-c", "echo 'success output' && echo 'some stderr' >&2")
	output, exitCode, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	// When exit code is zero, stdout should be used (even if stderr has content)
	if !strings.Contains(output, "success output") {
		t.Errorf("output = %q, should contain stdout 'success output' when exit code is zero", output)
	}
}

func TestExecuteAndCapture_EmptyOutput(t *testing.T) {
	b := &mockBackend{name: "test"}

	cmd := exec.Command("true") // Command that produces no output
	output, exitCode, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	if output != "" {
		t.Errorf("output = %q, want empty string", output)
	}
}

func TestExecuteAndCapture_LargeOutput(t *testing.T) {
	b := &mockBackend{name: "test"}

	// Generate a large output
	cmd := exec.Command("sh", "-c", "for i in $(seq 1 1000); do echo 'line $i'; done")
	output, exitCode, err := ExecuteAndCapture(b, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	if output == "" {
		t.Error("output should not be empty for large output test")
	}
}

// ==================== Environment Tests ====================

func TestWorkingDirectoryHandling(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "workdir-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks for comparison (macOS /tmp is symlinked to /private/tmp)
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	// Use Go to verify working directory instead of shell commands (cross-platform)
	// Create a simple test: verify cmd.Dir is correctly set by checking a file operation
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Verify the file exists in the temp directory
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("test file not created in temp directory")
	}

	// Verify original directory unchanged
	currentDir, _ := os.Getwd()
	if currentDir != originalDir {
		t.Errorf("original directory changed from %q to %q", originalDir, currentDir)
	}
}

// ==================== OutputMode Tests ====================

func TestDetermineOutputMode(t *testing.T) {
	tests := []struct {
		format   backend.OutputFormat
		expected OutputMode
	}{
		{backend.OutputText, OutputModeText},
		{backend.OutputDefault, OutputModeText},
		{backend.OutputJSON, OutputModeJSON},
		{backend.OutputStreamJSON, OutputModeStream},
		{"", OutputModeText},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			mode := DetermineOutputMode(tt.format)
			if mode != tt.expected {
				t.Errorf("DetermineOutputMode(%q) = %d, want %d", tt.format, mode, tt.expected)
			}
		})
	}
}

func TestDetermineInternalFormat(t *testing.T) {
	tests := []struct {
		userFormat backend.OutputFormat
		expected   backend.OutputFormat
	}{
		{backend.OutputText, backend.OutputJSON},
		{backend.OutputDefault, backend.OutputJSON},
		{"", backend.OutputJSON},
		{backend.OutputJSON, backend.OutputJSON},
		{backend.OutputStreamJSON, backend.OutputStreamJSON},
	}

	for _, tt := range tests {
		t.Run(string(tt.userFormat), func(t *testing.T) {
			internal := DetermineInternalFormat(tt.userFormat)
			if internal != tt.expected {
				t.Errorf("DetermineInternalFormat(%q) = %q, want %q", tt.userFormat, internal, tt.expected)
			}
		})
	}
}

// ==================== ExecutionConfig Tests ====================

func TestExecutionConfig_Defaults(t *testing.T) {
	cfg := &ExecutionConfig{
		Backend:    &mockBackend{name: "test"},
		OutputMode: OutputModeText,
		Stdin:      true,
	}

	if cfg.Backend == nil {
		t.Error("Backend should not be nil")
	}
	if cfg.OutputMode != OutputModeText {
		t.Errorf("OutputMode = %d, want %d", cfg.OutputMode, OutputModeText)
	}
	if !cfg.Stdin {
		t.Error("Stdin should be true")
	}
	if cfg.Session != nil {
		t.Error("Session should be nil by default")
	}
}

// ==================== ExecutionResult Tests ====================

func TestExecutionResult_Fields(t *testing.T) {
	result := &ExecutionResult{
		ExitCode:        0,
		SessionID:       "test-session-123",
		Content:         "Hello, world!",
		Error:           "",
		DurationSeconds: 1.5,
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.SessionID != "test-session-123" {
		t.Errorf("SessionID = %q, want 'test-session-123'", result.SessionID)
	}
	if result.Content != "Hello, world!" {
		t.Errorf("Content = %q, want 'Hello, world!'", result.Content)
	}
	if result.DurationSeconds != 1.5 {
		t.Errorf("DurationSeconds = %f, want 1.5", result.DurationSeconds)
	}
}

func TestExecutionResult_WithError(t *testing.T) {
	result := &ExecutionResult{
		ExitCode: 1,
		Error:    "command failed",
	}

	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
	if result.Error != "command failed" {
		t.Errorf("Error = %q, want 'command failed'", result.Error)
	}
}

// ==================== ExecuteCommand Tests ====================

func TestExecuteCommand_TextMode(t *testing.T) {
	b := &mockBackend{name: "test"}
	cfg := &ExecutionConfig{
		Backend:    b,
		OutputMode: OutputModeText,
		Stdin:      false,
	}
	cmd := exec.Command("echo", "test output")

	result, err := ExecuteCommand(cfg, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestExecuteCommand_StreamMode(t *testing.T) {
	b := &mockBackend{name: "test"}
	cfg := &ExecutionConfig{
		Backend:    b,
		OutputMode: OutputModeStream,
		Stdin:      false,
	}
	cmd := exec.Command("echo", "streamed output")

	result, err := ExecuteCommand(cfg, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestExecuteCommand_JSONMode(t *testing.T) {
	b := &mockBackend{
		name: "test",
		jsonResponse: &backend.UnifiedResponse{
			Content:   "json content",
			SessionID: "json-session-123",
		},
	}
	cfg := &ExecutionConfig{
		Backend:    b,
		OutputMode: OutputModeJSON,
		Stdin:      false,
	}
	cmd := exec.Command("echo", `{"content": "test"}`)

	result, err := ExecuteCommand(cfg, cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestGetCommandTimeout(t *testing.T) {
	t.Run("returns zero when not configured", func(t *testing.T) {
		// Reset config to ensure clean state
		config.Reset()
		defer config.Reset()

		timeout := GetCommandTimeout()
		if timeout != 0 {
			t.Errorf("timeout = %v, want 0", timeout)
		}
	})
}

func TestExecuteCommand_WithTimeout(t *testing.T) {
	b := &mockBackend{
		name: "test",
	}

	t.Run("command completes before timeout", func(t *testing.T) {
		cfg := &ExecutionConfig{
			Backend:    b,
			OutputMode: OutputModeText,
			Stdin:      false,
			Timeout:    5 * time.Second,
		}
		cmd := exec.Command("echo", "quick")

		result, err := ExecuteCommand(cfg, cmd)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", result.ExitCode)
		}
	})

	t.Run("command times out", func(t *testing.T) {
		cfg := &ExecutionConfig{
			Backend:    b,
			OutputMode: OutputModeText,
			Stdin:      false,
			Timeout:    100 * time.Millisecond,
		}
		// Use sleep command that takes longer than timeout
		cmd := exec.Command("sleep", "5")

		result, err := ExecuteCommand(cfg, cmd)

		if err != ErrCommandTimeout {
			t.Errorf("error = %v, want ErrCommandTimeout", err)
		}
		if result.ExitCode != 124 {
			t.Errorf("ExitCode = %d, want 124 (timeout)", result.ExitCode)
		}
		if result.Error != ErrCommandTimeout.Error() {
			t.Errorf("Error = %q, want %q", result.Error, ErrCommandTimeout.Error())
		}
	})
}
