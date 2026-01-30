package service

import (
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
