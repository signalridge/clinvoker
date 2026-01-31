package app

import (
	"encoding/json"
	"testing"
	"time"
)

// TestStatusText_Extended adds additional edge cases beyond the basic tests in commands_test.go
func TestStatusText_Extended(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		err      string
		want     string
	}{
		{
			name:     "failure with negative exit code",
			exitCode: -1,
			err:      "",
			want:     "FAILED",
		},
		{
			name:     "failure with high exit code 255",
			exitCode: 255,
			err:      "",
			want:     "FAILED",
		},
		{
			name:     "failure with whitespace-only error",
			exitCode: 0,
			err:      "   ",
			want:     "FAILED",
		},
		{
			name:     "failure with newline in error",
			exitCode: 0,
			err:      "error\nwith newline",
			want:     "FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusText(tt.exitCode, tt.err)
			if got != tt.want {
				t.Errorf("statusText(%d, %q) = %q, want %q", tt.exitCode, tt.err, got, tt.want)
			}
		})
	}
}

// TestCompareResult_JSONFieldNames verifies the exact JSON field names used for serialization
func TestCompareResult_JSONFieldNames(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 10, 30, 5, 0, time.UTC)

	result := CompareResult{
		Backend:   "claude",
		Model:     "opus",
		ExitCode:  1,
		Error:     "test error",
		Output:    "test output",
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  5.0,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal CompareResult: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify expected JSON field names
	expectedFields := map[string]bool{
		"backend":          true,
		"model":            true,
		"exit_code":        true,
		"error":            true,
		"output":           true,
		"start_time":       true,
		"end_time":         true,
		"duration_seconds": true,
	}

	for field := range expectedFields {
		if _, exists := data[field]; !exists {
			t.Errorf("expected JSON field %q not found", field)
		}
	}

	// Verify no unexpected fields
	for field := range data {
		if !expectedFields[field] {
			t.Errorf("unexpected JSON field %q found", field)
		}
	}
}

// TestCompareResult_OmitEmpty verifies omitempty behavior for optional fields
func TestCompareResult_OmitEmpty(t *testing.T) {
	tests := []struct {
		name          string
		result        CompareResult
		shouldOmit    []string
		shouldInclude []string
	}{
		{
			name: "empty model is omitted",
			result: CompareResult{
				Backend:  "claude",
				Model:    "",
				ExitCode: 0,
			},
			shouldOmit:    []string{"model"},
			shouldInclude: []string{"backend", "exit_code"},
		},
		{
			name: "empty error is omitted",
			result: CompareResult{
				Backend:  "claude",
				ExitCode: 0,
				Error:    "",
			},
			shouldOmit:    []string{"error"},
			shouldInclude: []string{"backend", "exit_code"},
		},
		{
			name: "empty output is omitted",
			result: CompareResult{
				Backend:  "claude",
				ExitCode: 0,
				Output:   "",
			},
			shouldOmit:    []string{"output"},
			shouldInclude: []string{"backend", "exit_code"},
		},
		{
			name: "non-empty fields are included",
			result: CompareResult{
				Backend:  "claude",
				Model:    "opus",
				ExitCode: 1,
				Error:    "failed",
				Output:   "some output",
			},
			shouldOmit:    []string{},
			shouldInclude: []string{"backend", "model", "exit_code", "error", "output"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var data map[string]interface{}
			if err := json.Unmarshal(jsonData, &data); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			for _, field := range tt.shouldOmit {
				if val, exists := data[field]; exists && val != "" {
					t.Errorf("field %q should be omitted when empty, got %v", field, val)
				}
			}

			for _, field := range tt.shouldInclude {
				if _, exists := data[field]; !exists {
					t.Errorf("field %q should be included", field)
				}
			}
		})
	}
}

// TestCompareResults_JSONFieldNames verifies the exact JSON field names used for CompareResults
func TestCompareResults_JSONFieldNames(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 10, 30, 10, 0, time.UTC)

	results := CompareResults{
		Prompt:        "test prompt",
		Backends:      []string{"claude"},
		Results:       []CompareResult{{Backend: "claude", ExitCode: 0}},
		TotalDuration: 10.0,
		StartTime:     startTime,
		EndTime:       endTime,
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("failed to marshal CompareResults: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify expected JSON field names
	expectedFields := map[string]bool{
		"prompt":                 true,
		"backends":               true,
		"results":                true,
		"total_duration_seconds": true,
		"start_time":             true,
		"end_time":               true,
	}

	for field := range expectedFields {
		if _, exists := data[field]; !exists {
			t.Errorf("expected JSON field %q not found", field)
		}
	}

	// Verify no unexpected fields
	for field := range data {
		if !expectedFields[field] {
			t.Errorf("unexpected JSON field %q found", field)
		}
	}
}

// TestCompareResults_EmptyArraysSerialization verifies empty arrays serialize correctly
func TestCompareResults_EmptyArraysSerialization(t *testing.T) {
	results := CompareResults{
		Prompt:   "test",
		Backends: []string{},
		Results:  []CompareResult{},
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CompareResults
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Backends == nil {
		t.Error("Backends should not be nil after round-trip")
	}
	if len(decoded.Backends) != 0 {
		t.Errorf("Backends length = %d, want 0", len(decoded.Backends))
	}

	if decoded.Results == nil {
		t.Error("Results should not be nil after round-trip")
	}
	if len(decoded.Results) != 0 {
		t.Errorf("Results length = %d, want 0", len(decoded.Results))
	}
}

// TestCompareResults_RoundTrip verifies complete round-trip serialization
func TestCompareResults_RoundTrip(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 10, 30, 10, 0, time.UTC)

	original := CompareResults{
		Prompt:   "explain quicksort algorithm",
		Backends: []string{"claude", "gemini", "codex"},
		Results: []CompareResult{
			{
				Backend:   "claude",
				Model:     "opus",
				ExitCode:  0,
				Output:    "Quicksort is a divide-and-conquer algorithm...",
				StartTime: startTime,
				EndTime:   endTime,
				Duration:  3.5,
			},
			{
				Backend:   "gemini",
				Model:     "pro",
				ExitCode:  0,
				Output:    "Quicksort works by selecting a pivot...",
				StartTime: startTime,
				EndTime:   endTime,
				Duration:  2.8,
			},
			{
				Backend:   "codex",
				ExitCode:  1,
				Error:     "API rate limit exceeded",
				StartTime: startTime,
				EndTime:   endTime,
				Duration:  0.5,
			},
		},
		TotalDuration: 10.0,
		StartTime:     startTime,
		EndTime:       endTime,
	}

	// Serialize
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Deserialize
	var decoded CompareResults
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify
	if decoded.Prompt != original.Prompt {
		t.Errorf("Prompt = %q, want %q", decoded.Prompt, original.Prompt)
	}
	if len(decoded.Backends) != len(original.Backends) {
		t.Errorf("Backends length = %d, want %d", len(decoded.Backends), len(original.Backends))
	}
	if len(decoded.Results) != len(original.Results) {
		t.Errorf("Results length = %d, want %d", len(decoded.Results), len(original.Results))
	}
	if decoded.TotalDuration != original.TotalDuration {
		t.Errorf("TotalDuration = %f, want %f", decoded.TotalDuration, original.TotalDuration)
	}

	// Check individual results
	for i, r := range decoded.Results {
		if r.Backend != original.Results[i].Backend {
			t.Errorf("Results[%d].Backend = %q, want %q", i, r.Backend, original.Results[i].Backend)
		}
		if r.ExitCode != original.Results[i].ExitCode {
			t.Errorf("Results[%d].ExitCode = %d, want %d", i, r.ExitCode, original.Results[i].ExitCode)
		}
	}
}

// TestCompareCmdStructure tests the cobra command structure
func TestCompareCmdStructure(t *testing.T) {
	t.Run("Use field", func(t *testing.T) {
		want := "compare <prompt>"
		if compareCmd.Use != want {
			t.Errorf("compareCmd.Use = %q, want %q", compareCmd.Use, want)
		}
	})

	t.Run("Short description", func(t *testing.T) {
		want := "Run same prompt on multiple backends and compare outputs"
		if compareCmd.Short != want {
			t.Errorf("compareCmd.Short = %q, want %q", compareCmd.Short, want)
		}
	})

	t.Run("Args requires exactly one argument", func(t *testing.T) {
		tests := []struct {
			name      string
			args      []string
			wantError bool
		}{
			{
				name:      "no arguments",
				args:      []string{},
				wantError: true,
			},
			{
				name:      "exactly one argument",
				args:      []string{"test prompt"},
				wantError: false,
			},
			{
				name:      "two arguments",
				args:      []string{"arg1", "arg2"},
				wantError: true,
			},
			{
				name:      "three arguments",
				args:      []string{"arg1", "arg2", "arg3"},
				wantError: true,
			},
			{
				name:      "one argument with spaces",
				args:      []string{"this is a long prompt with spaces"},
				wantError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := compareCmd.Args(compareCmd, tt.args)
				if (err != nil) != tt.wantError {
					t.Errorf("Args(%v) error = %v, wantError = %v", tt.args, err, tt.wantError)
				}
			})
		}
	})

	t.Run("Long description is set", func(t *testing.T) {
		if compareCmd.Long == "" {
			t.Error("compareCmd.Long should not be empty")
		}
		if len(compareCmd.Long) <= len(compareCmd.Short) {
			t.Error("Long description should be longer than Short description")
		}
	})

	t.Run("RunE is set", func(t *testing.T) {
		if compareCmd.RunE == nil {
			t.Error("compareCmd.RunE should not be nil")
		}
	})
}

// TestCompareCmdFlags tests that all expected flags are registered
func TestCompareCmdFlags(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		defaultValue string
	}{
		{
			name:         "backends flag",
			flagName:     "backends",
			defaultValue: "",
		},
		{
			name:         "all-backends flag",
			flagName:     "all-backends",
			defaultValue: "false",
		},
		{
			name:         "json flag",
			flagName:     "json",
			defaultValue: "false",
		},
		{
			name:         "sequential flag",
			flagName:     "sequential",
			defaultValue: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := compareCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %q default value = %q, want %q", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}
