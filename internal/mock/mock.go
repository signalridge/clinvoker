// Package mock provides mock implementations and testing utilities for clinvoker tests.
package mock

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
)

// MockBackend implements backend.Backend for testing.
type MockBackend struct {
	name           string
	available      bool
	parseOutput    string
	jsonResponse   *backend.UnifiedResponse
	jsonError      error
	separateStderr bool
	commandFunc    func(prompt string, opts *backend.UnifiedOptions) *exec.Cmd
}

// MockBackendOption configures a MockBackend.
type MockBackendOption func(*MockBackend)

// NewMockBackend creates a new mock backend with the given options.
func NewMockBackend(name string, opts ...MockBackendOption) *MockBackend {
	m := &MockBackend{
		name:      name,
		available: true,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WithAvailable sets the availability of the mock backend.
func WithAvailable(available bool) MockBackendOption {
	return func(m *MockBackend) {
		m.available = available
	}
}

// WithParseOutput sets the output returned by ParseOutput.
func WithParseOutput(output string) MockBackendOption {
	return func(m *MockBackend) {
		m.parseOutput = output
	}
}

// WithJSONResponse sets the response returned by ParseJSONResponse.
func WithJSONResponse(resp *backend.UnifiedResponse) MockBackendOption {
	return func(m *MockBackend) {
		m.jsonResponse = resp
	}
}

// WithJSONError sets the error returned by ParseJSONResponse.
func WithJSONError(err error) MockBackendOption {
	return func(m *MockBackend) {
		m.jsonError = err
	}
}

// WithSeparateStderr sets whether stderr should be separated.
func WithSeparateStderr(separate bool) MockBackendOption {
	return func(m *MockBackend) {
		m.separateStderr = separate
	}
}

// WithCommandFunc sets a custom function for building commands.
func WithCommandFunc(f func(prompt string, opts *backend.UnifiedOptions) *exec.Cmd) MockBackendOption {
	return func(m *MockBackend) {
		m.commandFunc = f
	}
}

// Name returns the backend name.
func (m *MockBackend) Name() string { return m.name }

// IsAvailable returns whether the backend is available.
func (m *MockBackend) IsAvailable() bool { return m.available }

// BuildCommand builds a command for the given prompt.
func (m *MockBackend) BuildCommand(prompt string, _ *backend.Options) *exec.Cmd {
	return exec.Command("echo", prompt)
}

// ResumeCommand builds a command to resume a session.
func (m *MockBackend) ResumeCommand(sessionID, prompt string, _ *backend.Options) *exec.Cmd {
	return exec.Command("echo", "resume", sessionID, prompt)
}

// BuildCommandUnified builds a command using unified options.
func (m *MockBackend) BuildCommandUnified(prompt string, opts *backend.UnifiedOptions) *exec.Cmd {
	if m.commandFunc != nil {
		return m.commandFunc(prompt, opts)
	}
	return exec.Command("echo", prompt)
}

// ResumeCommandUnified builds a resume command using unified options.
func (m *MockBackend) ResumeCommandUnified(sessionID, prompt string, _ *backend.UnifiedOptions) *exec.Cmd {
	return exec.Command("echo", "resume", sessionID, prompt)
}

// ParseOutput parses the raw output from the backend.
func (m *MockBackend) ParseOutput(rawOutput string) string {
	if m.parseOutput != "" {
		return m.parseOutput
	}
	return rawOutput
}

// ParseJSONResponse parses JSON output from the backend.
func (m *MockBackend) ParseJSONResponse(rawOutput string) (*backend.UnifiedResponse, error) {
	if m.jsonError != nil {
		return nil, m.jsonError
	}
	if m.jsonResponse != nil {
		return m.jsonResponse, nil
	}
	return &backend.UnifiedResponse{Content: rawOutput}, nil
}

// SeparateStderr returns whether stderr should be captured separately.
func (m *MockBackend) SeparateStderr() bool { return m.separateStderr }

// TempDir creates a temporary directory for testing.
// It returns the directory path and a cleanup function.
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "clinvoker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Resolve symlinks (macOS /tmp is symlinked to /private/tmp)
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		_ = os.RemoveAll(dir)
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return dir, cleanup
}

// TempFile creates a temporary file with the given content.
// It returns the file path and a cleanup function.
func TempFile(t *testing.T, name, content string) (string, func()) {
	t.Helper()

	dir, dirCleanup := TempDir(t)
	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		dirCleanup()
		t.Fatalf("failed to write temp file: %v", err)
	}

	return path, dirCleanup
}

// NewTestServer creates a test HTTP server with the given handler.
// It returns the server and automatically handles cleanup when the test ends.
func NewTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})

	return server
}

// WaitForCondition waits for a condition to become true within the timeout.
// It returns true if the condition was met, false if timeout occurred.
func WaitForCondition(timeout time.Duration, condition func() bool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if condition() {
				return true
			}
		}
	}
}

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// AssertEqual fails the test if got != want.
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertContains fails the test if s does not contain substr.
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func contains(s, substr string) bool {
	return substr == "" || (len(s) >= len(substr) && searchSubstring(s, substr))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// WithMockBackend registers a mock backend and returns a cleanup function
// that unregisters all backends and re-registers the standard ones.
// Usage:
//
//	cleanup := mock.WithMockBackend(t, mockBackend)
//	t.Cleanup(cleanup)
func WithMockBackend(t *testing.T, m *MockBackend) func() {
	t.Helper()
	backend.Register(m)
	return func() {
		backend.UnregisterAll()
		backend.Register(&backend.Claude{})
		backend.Register(&backend.Codex{})
		backend.Register(&backend.Gemini{})
	}
}
