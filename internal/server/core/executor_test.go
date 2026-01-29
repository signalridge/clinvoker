package core

import (
	"context"
	"testing"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/mock"
)

func TestExecute_NilRequest(t *testing.T) {
	_, err := Execute(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
	if err.Error() != "invalid execution request" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExecute_NilBackend(t *testing.T) {
	req := &Request{
		Backend: nil,
		Prompt:  "test",
	}
	_, err := Execute(context.Background(), req)
	if err == nil {
		t.Error("expected error for nil backend")
	}
	if err.Error() != "invalid execution request" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExecute_DryRun(t *testing.T) {
	mockBackend := mock.NewMockBackend("mock",
		mock.WithAvailable(true),
		mock.WithParseOutput("mocked response"),
	)

	req := &Request{
		Backend: mockBackend,
		Prompt:  "test prompt",
		Options: &backend.UnifiedOptions{
			DryRun: true,
		},
	}

	result, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Output == "" {
		t.Error("expected non-empty output for dry run")
	}
}

func TestExecute_NilOptions(t *testing.T) {
	mockBackend := mock.NewMockBackend("mock",
		mock.WithAvailable(true),
		mock.WithParseOutput("mocked response"),
	)

	req := &Request{
		Backend: mockBackend,
		Prompt:  "test prompt",
		Options: nil, // nil options should be handled gracefully
	}

	// This will actually execute the mock, but we're testing nil options handling
	result, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Result should be populated
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestExecute_WithRequestedFormat(t *testing.T) {
	tests := []struct {
		name            string
		requestedFormat backend.OutputFormat
	}{
		{"default format", backend.OutputDefault},
		{"json format", backend.OutputJSON},
		{"text format", backend.OutputText},
		{"stream-json format", backend.OutputStreamJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBackend := mock.NewMockBackend("mock",
				mock.WithAvailable(true),
				mock.WithParseOutput("mocked response"),
			)

			req := &Request{
				Backend:         mockBackend,
				Prompt:          "test prompt",
				Options:         &backend.UnifiedOptions{DryRun: true},
				RequestedFormat: tt.requestedFormat,
			}

			result, err := Execute(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ExitCode != 0 {
				t.Errorf("expected exit code 0, got %d", result.ExitCode)
			}
		})
	}
}

func TestExecute_OptionsNotMutated(t *testing.T) {
	mockBackend := mock.NewMockBackend("mock",
		mock.WithAvailable(true),
		mock.WithParseOutput("mocked response"),
	)

	originalFormat := backend.OutputText
	opts := &backend.UnifiedOptions{
		OutputFormat: originalFormat,
		DryRun:       true,
	}

	req := &Request{
		Backend:         mockBackend,
		Prompt:          "test prompt",
		Options:         opts,
		RequestedFormat: backend.OutputJSON,
	}

	_, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original options were not mutated
	if opts.OutputFormat != originalFormat {
		t.Errorf("options were mutated: OutputFormat changed from %v to %v", originalFormat, opts.OutputFormat)
	}
}

func TestExecute_FallbackToParseOutputWhenJSONContentEmpty(t *testing.T) {
	mockBackend := mock.NewMockBackend("mock",
		mock.WithAvailable(true),
		mock.WithParseOutput("parsed output"),
		mock.WithJSONResponse(&backend.UnifiedResponse{}),
	)

	req := &Request{
		Backend: mockBackend,
		Prompt:  "test prompt",
		Options: &backend.UnifiedOptions{},
	}

	result, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "parsed output" {
		t.Errorf("Output = %q, want %q", result.Output, "parsed output")
	}
}

func TestRequest_Structure(t *testing.T) {
	mockBackend := mock.NewMockBackend("mock", mock.WithAvailable(true))

	req := Request{
		Backend:         mockBackend,
		Prompt:          "test",
		Options:         &backend.UnifiedOptions{},
		RequestedFormat: backend.OutputJSON,
	}

	if req.Backend == nil {
		t.Error("Backend should not be nil")
	}
	if req.Prompt != "test" {
		t.Errorf("Prompt = %q, want 'test'", req.Prompt)
	}
}

func TestResult_Structure(t *testing.T) {
	result := Result{
		Output:           "test output",
		Usage:            &backend.TokenUsage{InputTokens: 10, OutputTokens: 5},
		BackendSessionID: "session-123",
		ExitCode:         0,
		Error:            "",
	}

	if result.Output != "test output" {
		t.Errorf("Output = %q, want 'test output'", result.Output)
	}
	if result.Usage.InputTokens != 10 {
		t.Errorf("Usage.InputTokens = %d, want 10", result.Usage.InputTokens)
	}
	if result.BackendSessionID != "session-123" {
		t.Errorf("BackendSessionID = %q, want 'session-123'", result.BackendSessionID)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}
