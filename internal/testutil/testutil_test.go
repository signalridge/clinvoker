package testutil

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
)

func TestNewMockBackend(t *testing.T) {
	m := NewMockBackend("test")

	if m.Name() != "test" {
		t.Errorf("Name() = %q, want 'test'", m.Name())
	}
	if !m.IsAvailable() {
		t.Error("IsAvailable() should be true by default")
	}
}

func TestMockBackend_WithOptions(t *testing.T) {
	testErr := errors.New("test error")
	testResp := &backend.UnifiedResponse{Content: "test content"}

	m := NewMockBackend("test",
		WithAvailable(false),
		WithParseOutput("parsed"),
		WithJSONResponse(testResp),
		WithJSONError(testErr),
		WithSeparateStderr(true),
	)

	if m.IsAvailable() {
		t.Error("IsAvailable() should be false")
	}
	if m.ParseOutput("raw") != "parsed" {
		t.Errorf("ParseOutput() = %q, want 'parsed'", m.ParseOutput("raw"))
	}
	if m.SeparateStderr() != true {
		t.Error("SeparateStderr() should be true")
	}

	// JSONError takes precedence
	_, err := m.ParseJSONResponse("raw")
	if err != testErr {
		t.Errorf("ParseJSONResponse error = %v, want %v", err, testErr)
	}
}

func TestMockBackend_ParseJSONResponse(t *testing.T) {
	t.Run("default behavior", func(t *testing.T) {
		m := NewMockBackend("test")
		resp, err := m.ParseJSONResponse("raw output")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Content != "raw output" {
			t.Errorf("Content = %q, want 'raw output'", resp.Content)
		}
	})

	t.Run("custom response", func(t *testing.T) {
		customResp := &backend.UnifiedResponse{Content: "custom"}
		m := NewMockBackend("test", WithJSONResponse(customResp))
		resp, err := m.ParseJSONResponse("ignored")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp != customResp {
			t.Error("should return custom response")
		}
	})
}

func TestMockBackend_BuildCommand(t *testing.T) {
	m := NewMockBackend("test")

	cmd := m.BuildCommand("test prompt", nil)
	if cmd == nil {
		t.Fatal("BuildCommand returned nil")
	}

	cmd = m.BuildCommandUnified("test prompt", nil)
	if cmd == nil {
		t.Fatal("BuildCommandUnified returned nil")
	}

	cmd = m.ResumeCommand("session", "prompt", nil)
	if cmd == nil {
		t.Fatal("ResumeCommand returned nil")
	}

	cmd = m.ResumeCommandUnified("session", "prompt", nil)
	if cmd == nil {
		t.Fatal("ResumeCommandUnified returned nil")
	}
}

func TestMockBackend_WithCommandFunc(t *testing.T) {
	called := false
	m := NewMockBackend("test", WithCommandFunc(func(prompt string, opts *backend.UnifiedOptions) *exec.Cmd {
		called = true
		return exec.Command("custom", prompt)
	}))

	cmd := m.BuildCommandUnified("test", nil)
	if !called {
		t.Error("custom command func was not called")
	}
	if cmd.Args[0] != "custom" {
		t.Errorf("command = %q, want 'custom'", cmd.Args[0])
	}
}

func TestTempDir(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()

	if dir == "" {
		t.Fatal("TempDir returned empty path")
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("failed to stat temp dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("TempDir should return a directory")
	}

	// After cleanup, dir should not exist
	cleanup()
	_, err = os.Stat(dir)
	if !os.IsNotExist(err) {
		t.Error("TempDir cleanup should remove directory")
	}
}

func TestTempFile(t *testing.T) {
	path, cleanup := TempFile(t, "test.txt", "content")
	defer cleanup()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("file content = %q, want 'content'", string(data))
	}

	// Verify file is in temp dir
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("temp dir should exist: %v", err)
	}
}

func TestNewTestServer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	server := NewTestServer(t, handler)
	if server == nil {
		t.Fatal("NewTestServer returned nil")
	}
	if server.URL == "" {
		t.Error("server URL should not be empty")
	}

	// Make a request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestWaitForCondition(t *testing.T) {
	t.Run("condition met", func(t *testing.T) {
		count := 0
		result := WaitForCondition(time.Second, func() bool {
			count++
			return count >= 3
		})
		if !result {
			t.Error("expected condition to be met")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		result := WaitForCondition(50*time.Millisecond, func() bool {
			return false
		})
		if result {
			t.Error("expected timeout")
		}
	})
}

func TestAssertNoError(t *testing.T) {
	// This would fail the test if err != nil
	// Since we can't test failure, just verify it doesn't panic
	AssertNoError(t, nil)
}

func TestAssertError(t *testing.T) {
	// This would fail the test if err == nil
	// Since we can't test failure, just verify it doesn't panic
	AssertError(t, errors.New("test"))
}

func TestAssertEqual(t *testing.T) {
	AssertEqual(t, 1, 1)
	AssertEqual(t, "hello", "hello")
	AssertEqual(t, true, true)
}

func TestAssertContains(t *testing.T) {
	AssertContains(t, "hello world", "world")
	AssertContains(t, "hello world", "")
	AssertContains(t, "hello world", "hello")
}
