package service

import (
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/mock"
)

func TestPreparePrompt_AppliesOutputFormatDefault(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	if err := config.Init(""); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	cfg := config.Get()
	cfg.Output.Format = "stream-json"

	mockBackend := mock.NewMockBackend("mock-format-default", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	prep, err := preparePrompt(&PromptRequest{
		Backend: "mock-format-default",
		Prompt:  "test",
	}, false)
	if err != nil {
		t.Fatalf("preparePrompt failed: %v", err)
	}

	if prep.requestedFormat != backend.OutputStreamJSON {
		t.Errorf("requestedFormat = %q, want %q", prep.requestedFormat, backend.OutputStreamJSON)
	}
}

func TestPreparePrompt_ExplicitOutputFormatWins(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	if err := config.Init(""); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	cfg := config.Get()
	cfg.Output.Format = "json"

	mockBackend := mock.NewMockBackend("mock-format-explicit", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	prep, err := preparePrompt(&PromptRequest{
		Backend:      "mock-format-explicit",
		Prompt:       "test",
		OutputFormat: "text",
	}, false)
	if err != nil {
		t.Fatalf("preparePrompt failed: %v", err)
	}

	if prep.requestedFormat != backend.OutputText {
		t.Errorf("requestedFormat = %q, want %q", prep.requestedFormat, backend.OutputText)
	}
}

func TestPreparePrompt_InvalidOutputFormat(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	if err := config.Init(""); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	mockBackend := mock.NewMockBackend("mock-invalid-format", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	_, err := preparePrompt(&PromptRequest{
		Backend:      "mock-invalid-format",
		Prompt:       "test",
		OutputFormat: "xml",
	}, false)
	if err == nil {
		t.Fatal("expected error for invalid output format")
	}
	if !strings.Contains(err.Error(), "invalid output_format") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "invalid output_format")
	}
}

func TestPreparePrompt_DisabledBackend(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	if err := config.Init(""); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	disabled := false
	cfg := config.Get()
	cfg.Backends["mock-disabled"] = config.BackendConfig{Enabled: &disabled}

	mockBackend := mock.NewMockBackend("mock-disabled", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	_, err := preparePrompt(&PromptRequest{
		Backend: "mock-disabled",
		Prompt:  "test",
	}, false)
	if err == nil {
		t.Fatal("expected error for disabled backend")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "disabled")
	}
}

func TestPreparePrompt_InvalidWorkDir(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	if err := config.Init(""); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	mockBackend := mock.NewMockBackend("mock-workdir", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	_, err := preparePrompt(&PromptRequest{
		Backend: "mock-workdir",
		Prompt:  "test",
		WorkDir: "relative/path",
	}, false)
	if err == nil {
		t.Fatal("expected error for invalid workdir")
	}
	if !strings.Contains(err.Error(), "work directory") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "work directory")
	}
}

func TestPreparePrompt_NilRequest(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	if err := config.Init(""); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	_, err := preparePrompt(nil, false)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}
