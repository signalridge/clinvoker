package app

import (
	"encoding/json"
	"testing"
	"time"
)

// TestSubstitutePromptPlaceholders_TableDriven provides comprehensive table-driven tests
// for the substitutePromptPlaceholders function with various edge cases.
func TestSubstitutePromptPlaceholders_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		prompt            string
		previousOutput    string
		hasPreviousOutput bool
		want              string
	}{
		{
			name:              "replaces single {{previous}} placeholder",
			prompt:            "fix the issues: {{previous}}",
			previousOutput:    "error in line 10",
			hasPreviousOutput: true,
			want:              "fix the issues: error in line 10",
		},
		{
			name:              "replaces multiple {{previous}} placeholders",
			prompt:            "{{previous}} and also {{previous}}",
			previousOutput:    "test output",
			hasPreviousOutput: true,
			want:              "test output and also test output",
		},
		{
			name:              "no replacement when hasPreviousOutput is false",
			prompt:            "fix the issues: {{previous}}",
			previousOutput:    "error in line 10",
			hasPreviousOutput: false,
			want:              "fix the issues: {{previous}}",
		},
		{
			name:              "no placeholder in prompt",
			prompt:            "analyze this code",
			previousOutput:    "some output",
			hasPreviousOutput: true,
			want:              "analyze this code",
		},
		{
			name:              "empty previous output",
			prompt:            "process: {{previous}}",
			previousOutput:    "",
			hasPreviousOutput: true,
			want:              "process: ",
		},
		{
			name:              "empty prompt",
			prompt:            "",
			previousOutput:    "output",
			hasPreviousOutput: true,
			want:              "",
		},
		{
			name:              "multiline previous output",
			prompt:            "review:\n{{previous}}",
			previousOutput:    "line 1\nline 2\nline 3",
			hasPreviousOutput: true,
			want:              "review:\nline 1\nline 2\nline 3",
		},
		{
			name:              "placeholder with special characters in output",
			prompt:            "fix: {{previous}}",
			previousOutput:    "error: $var && {brace}",
			hasPreviousOutput: true,
			want:              "fix: error: $var && {brace}",
		},
		{
			name:              "placeholder at start of prompt",
			prompt:            "{{previous}} - analyze this",
			previousOutput:    "code review",
			hasPreviousOutput: true,
			want:              "code review - analyze this",
		},
		{
			name:              "placeholder at end of prompt",
			prompt:            "continue with: {{previous}}",
			previousOutput:    "step output",
			hasPreviousOutput: true,
			want:              "continue with: step output",
		},
		{
			name:              "placeholder only",
			prompt:            "{{previous}}",
			previousOutput:    "full replacement",
			hasPreviousOutput: true,
			want:              "full replacement",
		},
		{
			name:              "unicode in output",
			prompt:            "result: {{previous}}",
			previousOutput:    "success! \u2714",
			hasPreviousOutput: true,
			want:              "result: success! \u2714",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substitutePromptPlaceholders(tt.prompt, tt.previousOutput, tt.hasPreviousOutput)
			if got != tt.want {
				t.Errorf("substitutePromptPlaceholders(%q, %q, %v) = %q, want %q",
					tt.prompt, tt.previousOutput, tt.hasPreviousOutput, got, tt.want)
			}
		})
	}
}

func TestResolveStepWorkDir(t *testing.T) {
	tests := []struct {
		name            string
		explicit        string
		passWorkDir     bool
		previousWorkDir string
		want            string
	}{
		{
			name:            "explicit takes priority",
			explicit:        "/explicit/path",
			passWorkDir:     true,
			previousWorkDir: "/previous/path",
			want:            "/explicit/path",
		},
		{
			name:            "uses previous when passWorkDir is true",
			explicit:        "",
			passWorkDir:     true,
			previousWorkDir: "/previous/path",
			want:            "/previous/path",
		},
		{
			name:            "empty when passWorkDir is false and no explicit",
			explicit:        "",
			passWorkDir:     false,
			previousWorkDir: "/previous/path",
			want:            "",
		},
		{
			name:            "empty when passWorkDir is true but no previous",
			explicit:        "",
			passWorkDir:     true,
			previousWorkDir: "",
			want:            "",
		},
		{
			name:            "all empty",
			explicit:        "",
			passWorkDir:     false,
			previousWorkDir: "",
			want:            "",
		},
		{
			name:            "explicit path with spaces",
			explicit:        "/path/with spaces/here",
			passWorkDir:     false,
			previousWorkDir: "",
			want:            "/path/with spaces/here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveStepWorkDir(tt.explicit, tt.passWorkDir, tt.previousWorkDir)
			if got != tt.want {
				t.Errorf("resolveStepWorkDir(%q, %v, %q) = %q, want %q",
					tt.explicit, tt.passWorkDir, tt.previousWorkDir, got, tt.want)
			}
		})
	}
}

// TestChainDefinitionJSONParsing_Comprehensive provides comprehensive JSON parsing tests
// for ChainDefinition struct.
func TestChainDefinitionJSONParsing_Comprehensive(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    ChainDefinition
		wantErr bool
	}{
		{
			name: "basic chain with steps",
			json: `{
				"steps": [
					{"backend": "claude", "prompt": "analyze code"},
					{"backend": "codex", "prompt": "fix issues: {{previous}}"}
				]
			}`,
			want: ChainDefinition{
				Steps: []ChainStep{
					{Backend: "claude", Prompt: "analyze code"},
					{Backend: "codex", Prompt: "fix issues: {{previous}}"},
				},
			},
			wantErr: false,
		},
		{
			name: "chain with all options",
			json: `{
				"steps": [
					{
						"backend": "claude",
						"prompt": "test",
						"model": "claude-3",
						"workdir": "/tmp",
						"approval_mode": "full-auto",
						"sandbox_mode": "relaxed",
						"max_turns": 5,
						"name": "step 1"
					}
				],
				"stop_on_failure": true,
				"pass_working_dir": true
			}`,
			want: ChainDefinition{
				Steps: []ChainStep{
					{
						Backend:      "claude",
						Prompt:       "test",
						Model:        "claude-3",
						WorkDir:      "/tmp",
						ApprovalMode: "full-auto",
						SandboxMode:  "relaxed",
						MaxTurns:     5,
						Name:         "step 1",
					},
				},
				StopOnFailure:  true,
				PassWorkingDir: true,
			},
			wantErr: false,
		},
		{
			name: "empty steps array",
			json: `{"steps": []}`,
			want: ChainDefinition{
				Steps: []ChainStep{},
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{"steps": [}`,
			want:    ChainDefinition{},
			wantErr: true,
		},
		{
			name:    "missing steps field",
			json:    `{"stop_on_failure": true}`,
			want:    ChainDefinition{StopOnFailure: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ChainDefinition
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got.Steps) != len(tt.want.Steps) {
				t.Errorf("steps count = %d, want %d", len(got.Steps), len(tt.want.Steps))
				return
			}
			if got.StopOnFailure != tt.want.StopOnFailure {
				t.Errorf("StopOnFailure = %v, want %v", got.StopOnFailure, tt.want.StopOnFailure)
			}
			if got.PassWorkingDir != tt.want.PassWorkingDir {
				t.Errorf("PassWorkingDir = %v, want %v", got.PassWorkingDir, tt.want.PassWorkingDir)
			}
		})
	}
}

func TestChainStepJSONParsing(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    ChainStep
		wantErr bool
	}{
		{
			name: "minimal step",
			json: `{"backend": "claude", "prompt": "test"}`,
			want: ChainStep{
				Backend: "claude",
				Prompt:  "test",
			},
			wantErr: false,
		},
		{
			name: "full step with all fields",
			json: `{
				"backend": "gemini",
				"prompt": "analyze this",
				"model": "gemini-pro",
				"workdir": "/home/user",
				"approval_mode": "suggest",
				"sandbox_mode": "strict",
				"max_turns": 10,
				"name": "Analysis Step"
			}`,
			want: ChainStep{
				Backend:      "gemini",
				Prompt:       "analyze this",
				Model:        "gemini-pro",
				WorkDir:      "/home/user",
				ApprovalMode: "suggest",
				SandboxMode:  "strict",
				MaxTurns:     10,
				Name:         "Analysis Step",
			},
			wantErr: false,
		},
		{
			name: "step with zero max_turns",
			json: `{"backend": "codex", "prompt": "test", "max_turns": 0}`,
			want: ChainStep{
				Backend:  "codex",
				Prompt:   "test",
				MaxTurns: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ChainStep
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Backend != tt.want.Backend {
				t.Errorf("Backend = %q, want %q", got.Backend, tt.want.Backend)
			}
			if got.Prompt != tt.want.Prompt {
				t.Errorf("Prompt = %q, want %q", got.Prompt, tt.want.Prompt)
			}
			if got.Model != tt.want.Model {
				t.Errorf("Model = %q, want %q", got.Model, tt.want.Model)
			}
			if got.WorkDir != tt.want.WorkDir {
				t.Errorf("WorkDir = %q, want %q", got.WorkDir, tt.want.WorkDir)
			}
			if got.ApprovalMode != tt.want.ApprovalMode {
				t.Errorf("ApprovalMode = %q, want %q", got.ApprovalMode, tt.want.ApprovalMode)
			}
			if got.SandboxMode != tt.want.SandboxMode {
				t.Errorf("SandboxMode = %q, want %q", got.SandboxMode, tt.want.SandboxMode)
			}
			if got.MaxTurns != tt.want.MaxTurns {
				t.Errorf("MaxTurns = %d, want %d", got.MaxTurns, tt.want.MaxTurns)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
		})
	}
}

func TestChainStepResultJSONSerialization(t *testing.T) {
	fixedStartTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fixedEndTime := time.Date(2024, 1, 15, 10, 0, 5, 0, time.UTC)

	tests := []struct {
		name   string
		result ChainStepResult
	}{
		{
			name: "successful step result",
			result: ChainStepResult{
				Step:      1,
				Name:      "Analysis",
				Backend:   "claude",
				ExitCode:  0,
				Output:    "analysis complete",
				SessionID: "sess-123",
				Duration:  5.0,
				StartTime: fixedStartTime,
				EndTime:   fixedEndTime,
			},
		},
		{
			name: "failed step result",
			result: ChainStepResult{
				Step:      2,
				Backend:   "codex",
				ExitCode:  1,
				Error:     "command failed",
				Duration:  2.5,
				StartTime: fixedStartTime,
				EndTime:   fixedEndTime,
			},
		},
		{
			name: "minimal step result",
			result: ChainStepResult{
				Step:     1,
				Backend:  "gemini",
				ExitCode: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Unmarshal back
			var got ChainStepResult
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			// Verify fields
			if got.Step != tt.result.Step {
				t.Errorf("Step = %d, want %d", got.Step, tt.result.Step)
			}
			if got.Name != tt.result.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.result.Name)
			}
			if got.Backend != tt.result.Backend {
				t.Errorf("Backend = %q, want %q", got.Backend, tt.result.Backend)
			}
			if got.ExitCode != tt.result.ExitCode {
				t.Errorf("ExitCode = %d, want %d", got.ExitCode, tt.result.ExitCode)
			}
			if got.Error != tt.result.Error {
				t.Errorf("Error = %q, want %q", got.Error, tt.result.Error)
			}
			if got.Output != tt.result.Output {
				t.Errorf("Output = %q, want %q", got.Output, tt.result.Output)
			}
			if got.Duration != tt.result.Duration {
				t.Errorf("Duration = %f, want %f", got.Duration, tt.result.Duration)
			}
		})
	}
}

func TestChainResultsJSONSerialization(t *testing.T) {
	fixedStartTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fixedEndTime := time.Date(2024, 1, 15, 10, 0, 10, 0, time.UTC)

	tests := []struct {
		name    string
		results ChainResults
	}{
		{
			name: "successful chain results",
			results: ChainResults{
				TotalSteps:     3,
				CompletedSteps: 3,
				FailedStep:     0,
				Results: []ChainStepResult{
					{Step: 1, Backend: "claude", ExitCode: 0},
					{Step: 2, Backend: "codex", ExitCode: 0},
					{Step: 3, Backend: "gemini", ExitCode: 0},
				},
				TotalDuration: 10.0,
				StartTime:     fixedStartTime,
				EndTime:       fixedEndTime,
			},
		},
		{
			name: "failed chain results",
			results: ChainResults{
				TotalSteps:     3,
				CompletedSteps: 1,
				FailedStep:     2,
				Results: []ChainStepResult{
					{Step: 1, Backend: "claude", ExitCode: 0},
					{Step: 2, Backend: "codex", ExitCode: 1, Error: "failed"},
				},
				TotalDuration: 5.0,
				StartTime:     fixedStartTime,
				EndTime:       fixedEndTime,
			},
		},
		{
			name: "empty results",
			results: ChainResults{
				TotalSteps:     0,
				CompletedSteps: 0,
				Results:        []ChainStepResult{},
				StartTime:      fixedStartTime,
				EndTime:        fixedEndTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.results)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Unmarshal back
			var got ChainResults
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			// Verify fields
			if got.TotalSteps != tt.results.TotalSteps {
				t.Errorf("TotalSteps = %d, want %d", got.TotalSteps, tt.results.TotalSteps)
			}
			if got.CompletedSteps != tt.results.CompletedSteps {
				t.Errorf("CompletedSteps = %d, want %d", got.CompletedSteps, tt.results.CompletedSteps)
			}
			if got.FailedStep != tt.results.FailedStep {
				t.Errorf("FailedStep = %d, want %d", got.FailedStep, tt.results.FailedStep)
			}
			if len(got.Results) != len(tt.results.Results) {
				t.Errorf("Results count = %d, want %d", len(got.Results), len(tt.results.Results))
			}
			if got.TotalDuration != tt.results.TotalDuration {
				t.Errorf("TotalDuration = %f, want %f", got.TotalDuration, tt.results.TotalDuration)
			}
		})
	}
}

func TestChainResultsJSONOmitEmpty(t *testing.T) {
	// Test that omitempty fields are properly omitted
	results := ChainResults{
		TotalSteps:     1,
		CompletedSteps: 1,
		FailedStep:     0, // Should be omitted due to omitempty
		Results:        []ChainStepResult{},
	}

	data, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Check that failed_step is not in the JSON when 0
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if _, exists := m["failed_step"]; exists {
		t.Error("failed_step should be omitted when 0")
	}
}

func TestChainStepResultJSONOmitEmpty(t *testing.T) {
	// Test that omitempty fields are properly omitted
	result := ChainStepResult{
		Step:     1,
		Backend:  "claude",
		ExitCode: 0,
		// Name, Error, Output, SessionID are empty and should be omitted
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	omitEmptyFields := []string{"name", "error", "output", "session_id"}
	for _, field := range omitEmptyFields {
		if _, exists := m[field]; exists {
			t.Errorf("%s should be omitted when empty", field)
		}
	}
}

func TestChainCmdStructure(t *testing.T) {
	t.Run("command use field", func(t *testing.T) {
		if chainCmd.Use != "chain" {
			t.Errorf("chainCmd.Use = %q, want %q", chainCmd.Use, "chain")
		}
	})

	t.Run("command short description", func(t *testing.T) {
		want := "Chain multiple backends in sequence with context passing"
		if chainCmd.Short != want {
			t.Errorf("chainCmd.Short = %q, want %q", chainCmd.Short, want)
		}
	})

	t.Run("command has long description", func(t *testing.T) {
		if chainCmd.Long == "" {
			t.Error("chainCmd.Long should not be empty")
		}
	})

	t.Run("command has RunE function", func(t *testing.T) {
		if chainCmd.RunE == nil {
			t.Error("chainCmd.RunE should not be nil")
		}
	})
}

func TestChainDefinitionDeprecatedFields(t *testing.T) {
	// Test that deprecated fields are parsed but noted as deprecated
	jsonInput := `{
		"steps": [{"backend": "claude", "prompt": "test"}],
		"pass_session_id": true,
		"persist_sessions": true
	}`

	var chain ChainDefinition
	err := json.Unmarshal([]byte(jsonInput), &chain)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify deprecated fields are parsed
	if !chain.PassSessionID {
		t.Error("PassSessionID should be true")
	}
	if !chain.PersistSessions {
		t.Error("PersistSessions should be true")
	}
}

func TestUpdateChainContext(t *testing.T) {
	tests := []struct {
		name              string
		initialCtx        chainContext
		workDir           string
		output            string
		hasOutput         bool
		wantWorkDir       string
		wantOutput        string
		wantHasPrevOutput bool
	}{
		{
			name:              "updates all fields when hasOutput is true",
			initialCtx:        chainContext{},
			workDir:           "/new/dir",
			output:            "new output",
			hasOutput:         true,
			wantWorkDir:       "/new/dir",
			wantOutput:        "new output",
			wantHasPrevOutput: true,
		},
		{
			name:              "updates only workDir when hasOutput is false",
			initialCtx:        chainContext{previousOutput: "old output", hasPreviousOutput: true},
			workDir:           "/new/dir",
			output:            "ignored",
			hasOutput:         false,
			wantWorkDir:       "/new/dir",
			wantOutput:        "old output",
			wantHasPrevOutput: true,
		},
		{
			name:              "empty output with hasOutput true",
			initialCtx:        chainContext{previousOutput: "old", hasPreviousOutput: false},
			workDir:           "",
			output:            "",
			hasOutput:         true,
			wantWorkDir:       "",
			wantOutput:        "",
			wantHasPrevOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.initialCtx
			updateChainContext(&ctx, tt.workDir, tt.output, tt.hasOutput)

			if ctx.previousWorkDir != tt.wantWorkDir {
				t.Errorf("previousWorkDir = %q, want %q", ctx.previousWorkDir, tt.wantWorkDir)
			}
			if ctx.previousOutput != tt.wantOutput {
				t.Errorf("previousOutput = %q, want %q", ctx.previousOutput, tt.wantOutput)
			}
			if ctx.hasPreviousOutput != tt.wantHasPrevOutput {
				t.Errorf("hasPreviousOutput = %v, want %v", ctx.hasPreviousOutput, tt.wantHasPrevOutput)
			}
		})
	}
}

func TestFailStepResult(t *testing.T) {
	startTime := time.Now()
	result := &ChainStepResult{
		Step:    1,
		Backend: "claude",
	}

	failStepResult(result, startTime, "test error message")

	if result.Error != "test error message" {
		t.Errorf("Error = %q, want %q", result.Error, "test error message")
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want %d", result.ExitCode, 1)
	}
	if result.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}
	if result.Duration < 0 {
		t.Errorf("Duration = %f, want >= 0", result.Duration)
	}
}

func TestChainStepJSONTags(t *testing.T) {
	// Verify JSON tags are correctly applied by marshaling and checking output
	step := ChainStep{
		Backend:      "claude",
		Prompt:       "test",
		Model:        "opus",
		WorkDir:      "/tmp",
		ApprovalMode: "auto",
		SandboxMode:  "strict",
		MaxTurns:     5,
		Name:         "Test Step",
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	expectedKeys := map[string]bool{
		"backend":       true,
		"prompt":        true,
		"model":         true,
		"workdir":       true,
		"approval_mode": true,
		"sandbox_mode":  true,
		"max_turns":     true,
		"name":          true,
	}

	for key := range expectedKeys {
		if _, exists := m[key]; !exists {
			t.Errorf("expected JSON key %q not found", key)
		}
	}
}

func TestChainStepResultJSONTags(t *testing.T) {
	// Verify JSON tags are correctly applied
	result := ChainStepResult{
		Step:      1,
		Name:      "Test",
		Backend:   "claude",
		ExitCode:  0,
		Error:     "test error",
		Output:    "test output",
		SessionID: "sess-123",
		Duration:  1.5,
		StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC),
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	expectedKeys := map[string]bool{
		"step":             true,
		"name":             true,
		"backend":          true,
		"exit_code":        true,
		"error":            true,
		"output":           true,
		"session_id":       true,
		"duration_seconds": true,
		"start_time":       true,
		"end_time":         true,
	}

	for key := range expectedKeys {
		if _, exists := m[key]; !exists {
			t.Errorf("expected JSON key %q not found", key)
		}
	}
}

func TestChainResultsJSONTags(t *testing.T) {
	// Verify JSON tags are correctly applied
	results := ChainResults{
		TotalSteps:     2,
		CompletedSteps: 1,
		FailedStep:     2,
		Results:        []ChainStepResult{{Step: 1}},
		TotalDuration:  5.0,
		StartTime:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:        time.Date(2024, 1, 1, 0, 0, 5, 0, time.UTC),
	}

	data, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	expectedKeys := map[string]bool{
		"total_steps":            true,
		"completed_steps":        true,
		"failed_step":            true,
		"results":                true,
		"total_duration_seconds": true,
		"start_time":             true,
		"end_time":               true,
	}

	for key := range expectedKeys {
		if _, exists := m[key]; !exists {
			t.Errorf("expected JSON key %q not found", key)
		}
	}
}

func TestChainContextInitialization(t *testing.T) {
	ctx := chainContext{}

	if ctx.previousWorkDir != "" {
		t.Error("previousWorkDir should be empty on init")
	}
	if ctx.previousOutput != "" {
		t.Error("previousOutput should be empty on init")
	}
	if ctx.hasPreviousOutput {
		t.Error("hasPreviousOutput should be false on init")
	}
	if ctx.cfg != nil {
		t.Error("cfg should be nil on init")
	}
}
