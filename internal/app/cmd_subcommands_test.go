package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/mock"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestParseParallelTasks_Errors(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	tmpDir := t.TempDir()

	t.Run("invalid json", func(t *testing.T) {
		parallelFile = writeTempFile(t, tmpDir, "invalid.json", "not-json")
		_, err := parseParallelTasks()
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("no tasks", func(t *testing.T) {
		parallelFile = writeTempFile(t, tmpDir, "empty.json", `{"tasks":[]}`)
		_, err := parseParallelTasks()
		if err == nil {
			t.Fatal("expected error for empty tasks")
		}
		if !strings.Contains(err.Error(), "no tasks provided") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "no tasks provided")
		}
	})
}

func TestRunParallel_DryRun(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	mockBackend := mock.NewMockBackend("mock", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	dryRun = true
	parallelJSON = true

	tmpDir := t.TempDir()
	parallelFile = writeTempFile(t, tmpDir, "tasks.json", `{"tasks":[{"backend":"mock","prompt":"hello"}]}`)

	if err := runParallel(parallelCmd, nil); err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}
}

func TestParseChainDefinition_Errors(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	tmpDir := t.TempDir()

	t.Run("no steps", func(t *testing.T) {
		chainFile = writeTempFile(t, tmpDir, "chain-empty.json", `{"steps":[]}`)
		_, err := parseChainDefinition()
		if err == nil {
			t.Fatal("expected error for empty chain")
		}
		if !strings.Contains(err.Error(), "no steps defined") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "no steps defined")
		}
	})

	t.Run("pass_session_id not supported", func(t *testing.T) {
		chainFile = writeTempFile(t, tmpDir, "chain-pass-session.json", `{"steps":[{"backend":"mock","prompt":"hi"}],"pass_session_id":true}`)
		_, err := parseChainDefinition()
		if err == nil {
			t.Fatal("expected error for pass_session_id")
		}
		if !strings.Contains(err.Error(), "always ephemeral") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "always ephemeral")
		}
	})

	t.Run("session placeholder not supported", func(t *testing.T) {
		chainFile = writeTempFile(t, tmpDir, "chain-session-placeholder.json", `{"steps":[{"backend":"mock","prompt":"{{session}}"}]}`)
		_, err := parseChainDefinition()
		if err == nil {
			t.Fatal("expected error for {{session}} placeholder")
		}
		if !strings.Contains(err.Error(), "sessions are not persisted") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "sessions are not persisted")
		}
	})
}

func TestRunChain_DryRun(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	mockBackend := mock.NewMockBackend("mock", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	dryRun = true
	chainJSONFlag = true

	tmpDir := t.TempDir()
	chainFile = writeTempFile(t, tmpDir, "chain.json", `{"steps":[{"backend":"mock","prompt":"one"},{"backend":"mock","prompt":"two"}]}`)

	if err := runChain(chainCmd, nil); err != nil {
		t.Fatalf("runChain failed: %v", err)
	}
}

func TestRunChain_DryRun_JSONOutputValid(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	mockBackend := mock.NewMockBackend("mock", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	dryRun = true
	chainJSONFlag = true

	tmpDir := t.TempDir()
	chainFile = writeTempFile(t, tmpDir, "chain.json", `{"steps":[{"backend":"mock","prompt":"one"},{"backend":"mock","prompt":"two"}]}`)

	output := captureStdout(t, func() {
		if err := runChain(chainCmd, nil); err != nil {
			t.Fatalf("runChain failed: %v", err)
		}
	})

	var parsed ChainResults
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("runChain JSON output should be valid JSON, err=%v output=%s", err, output)
	}
	if parsed.TotalSteps != 2 {
		t.Fatalf("total_steps = %d, want 2", parsed.TotalSteps)
	}
	if len(parsed.Results) != 2 {
		t.Fatalf("results length = %d, want 2", len(parsed.Results))
	}
	if parsed.Results[0].Backend != "mock" || parsed.Results[1].Backend != "mock" {
		t.Fatalf("unexpected backends in chain results: %+v", parsed.Results)
	}
}

func TestRunCompare_DryRun(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	mockBackend := mock.NewMockBackend("mock", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	dryRun = true
	compareJSON = true
	compareBackends = "mock"

	if err := runCompare(compareCmd, []string{"hello"}); err != nil {
		t.Fatalf("runCompare failed: %v", err)
	}
}

func TestRunCompare_DryRun_JSONOutputValid(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	mockBackend := mock.NewMockBackend("mock", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	dryRun = true
	compareJSON = true
	compareBackends = "mock"

	output := captureStdout(t, func() {
		if err := runCompare(compareCmd, []string{"hello"}); err != nil {
			t.Fatalf("runCompare failed: %v", err)
		}
	})

	var parsed CompareResults
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("runCompare JSON output should be valid JSON, err=%v output=%s", err, output)
	}
	if len(parsed.Results) != 1 {
		t.Fatalf("results length = %d, want 1", len(parsed.Results))
	}
	if parsed.Results[0].Backend != "mock" {
		t.Fatalf("backend = %q, want %q", parsed.Results[0].Backend, "mock")
	}
}

func TestRunMCP_InvalidTransport(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	mcpTransport = "nope"
	if err := runMCP(mcpCmd, nil); err == nil {
		t.Fatal("expected error for unsupported transport")
	}
}
