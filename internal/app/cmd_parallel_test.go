package app

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple alphanumeric",
			input: "hello123",
			want:  "hello123",
		},
		{
			name:  "with hyphen and underscore",
			input: "task-1_test",
			want:  "task-1_test",
		},
		{
			name:  "uppercase preserved",
			input: "HelloWorld",
			want:  "HelloWorld",
		},
		{
			name:  "spaces become underscores",
			input: "hello world",
			want:  "hello_world",
		},
		{
			name:  "special chars become underscores",
			input: "task@#$%^&*()",
			want:  "task",
		},
		{
			name:  "unicode chars become underscores",
			input: "task-unicode-123",
			want:  "task-unicode-123",
		},
		{
			name:  "emoji becomes underscores then trimmed",
			input: "task-test",
			want:  "task-test",
		},
		{
			name:  "leading spaces trimmed",
			input: "  hello",
			want:  "hello",
		},
		{
			name:  "trailing spaces trimmed",
			input: "hello  ",
			want:  "hello",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only spaces",
			input: "   ",
			want:  "",
		},
		{
			name:  "only special chars",
			input: "@#$%",
			want:  "",
		},
		{
			name:  "path separators become underscores",
			input: "path/to/file",
			want:  "path_to_file",
		},
		{
			name:  "long string truncated to 64 chars",
			input: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnop",
			want:  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789ab",
		},
		{
			name:  "exactly 64 chars unchanged",
			input: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789ab",
			want:  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789ab",
		},
		{
			name:  "leading underscores stripped",
			input: "___hello",
			want:  "hello",
		},
		{
			name:  "trailing underscores stripped",
			input: "hello___",
			want:  "hello",
		},
		{
			name:  "middle underscores preserved",
			input: "hello___world",
			want:  "hello___world",
		},
		{
			name:  "dots become underscores",
			input: "file.name.txt",
			want:  "file_name_txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParallelOutputFilename(t *testing.T) {
	tests := []struct {
		name  string
		task  *ParallelTask
		index int
		want  string
	}{
		{
			name: "uses task ID when present",
			task: &ParallelTask{
				ID:   "my-task-id",
				Name: "My Task Name",
			},
			index: 0,
			want:  "001_my-task-id.json",
		},
		{
			name: "uses task name when ID is empty",
			task: &ParallelTask{
				ID:   "",
				Name: "My Task Name",
			},
			index: 1,
			want:  "002_My_Task_Name.json",
		},
		{
			name: "uses index-based name when ID and name are empty",
			task: &ParallelTask{
				ID:   "",
				Name: "",
			},
			index: 2,
			want:  "003_task-3.json",
		},
		{
			name: "sanitizes ID with special chars",
			task: &ParallelTask{
				ID:   "task@#$%special",
				Name: "Name",
			},
			index: 0,
			want:  "001_task____special.json",
		},
		{
			name: "falls back to index when sanitized ID is empty",
			task: &ParallelTask{
				ID:   "@#$%",
				Name: "",
			},
			index: 4,
			want:  "005_task-5.json",
		},
		{
			name: "index zero-padded to 3 digits",
			task: &ParallelTask{
				ID: "task",
			},
			index: 99,
			want:  "100_task.json",
		},
		{
			name: "large index works correctly",
			task: &ParallelTask{
				ID: "task",
			},
			index: 999,
			want:  "1000_task.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parallelOutputFilename(tt.task, tt.index)
			if got != tt.want {
				t.Errorf("parallelOutputFilename(%+v, %d) = %q, want %q", tt.task, tt.index, got, tt.want)
			}
		})
	}
}

func TestCloneStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "nil slice returns nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty slice returns nil",
			input: []string{},
			want:  nil,
		},
		{
			name:  "single element",
			input: []string{"one"},
			want:  []string{"one"},
		},
		{
			name:  "multiple elements",
			input: []string{"one", "two", "three"},
			want:  []string{"one", "two", "three"},
		},
		{
			name:  "elements with empty strings",
			input: []string{"", "a", ""},
			want:  []string{"", "a", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cloneStringSlice(tt.input)

			// Check length and nil
			if tt.want == nil {
				if got != nil {
					t.Errorf("cloneStringSlice(%v) = %v, want nil", tt.input, got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("cloneStringSlice(%v) length = %d, want %d", tt.input, len(got), len(tt.want))
				return
			}

			// Check values
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("cloneStringSlice(%v)[%d] = %q, want %q", tt.input, i, got[i], v)
				}
			}

			// Verify it's a true copy (not sharing underlying array)
			if len(tt.input) > 0 && len(got) > 0 {
				got[0] = "modified"
				if tt.input[0] == "modified" {
					t.Error("cloneStringSlice returned slice sharing underlying array")
				}
			}
		})
	}
}

func TestCloneStringMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
		want  map[string]string
	}{
		{
			name:  "nil map returns nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty map returns nil",
			input: map[string]string{},
			want:  nil,
		},
		{
			name:  "single entry",
			input: map[string]string{"key": "value"},
			want:  map[string]string{"key": "value"},
		},
		{
			name:  "multiple entries",
			input: map[string]string{"a": "1", "b": "2", "c": "3"},
			want:  map[string]string{"a": "1", "b": "2", "c": "3"},
		},
		{
			name:  "entries with empty values",
			input: map[string]string{"empty": "", "normal": "value"},
			want:  map[string]string{"empty": "", "normal": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cloneStringMap(tt.input)

			// Check nil
			if tt.want == nil {
				if got != nil {
					t.Errorf("cloneStringMap(%v) = %v, want nil", tt.input, got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("cloneStringMap(%v) length = %d, want %d", tt.input, len(got), len(tt.want))
				return
			}

			// Check values
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("cloneStringMap(%v)[%q] = %q, want %q", tt.input, k, got[k], v)
				}
			}

			// Verify it's a true copy (not sharing underlying map)
			if len(tt.input) > 0 && len(got) > 0 {
				for k := range got {
					got[k] = "modified"
					if tt.input[k] == "modified" {
						t.Error("cloneStringMap returned map sharing underlying data")
					}
					break
				}
			}
		})
	}
}

func TestResolveTaskStatus(t *testing.T) {
	tests := []struct {
		name   string
		result *TaskResult
		want   string
	}{
		{
			name: "exit code -1 is CANCELED",
			result: &TaskResult{
				ExitCode: -1,
				Error:    "",
			},
			want: "CANCELED",
		},
		{
			name: "exit code 0 with no error is OK",
			result: &TaskResult{
				ExitCode: 0,
				Error:    "",
			},
			want: "OK",
		},
		{
			name: "exit code 1 is FAILED",
			result: &TaskResult{
				ExitCode: 1,
				Error:    "",
			},
			want: "FAILED",
		},
		{
			name: "exit code 0 with error is FAILED",
			result: &TaskResult{
				ExitCode: 0,
				Error:    "some error message",
			},
			want: "FAILED",
		},
		{
			name: "exit code non-zero with error is FAILED",
			result: &TaskResult{
				ExitCode: 127,
				Error:    "command not found",
			},
			want: "FAILED",
		},
		{
			name: "canceled via fail-fast",
			result: &TaskResult{
				ExitCode: -1,
				Error:    "canceled (fail-fast)",
			},
			want: "CANCELED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveTaskStatus(tt.result)
			if got != tt.want {
				t.Errorf("resolveTaskStatus(%+v) = %q, want %q", tt.result, got, tt.want)
			}
		})
	}
}

func TestParallelTasksJSONSerialization(t *testing.T) {
	t.Run("minimal tasks", func(t *testing.T) {
		input := `{
			"tasks": [
				{"backend": "claude", "prompt": "test prompt"}
			]
		}`

		var tasks ParallelTasks
		if err := json.Unmarshal([]byte(input), &tasks); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if len(tasks.Tasks) != 1 {
			t.Errorf("expected 1 task, got %d", len(tasks.Tasks))
		}
		if tasks.Tasks[0].Backend != "claude" {
			t.Errorf("backend = %q, want %q", tasks.Tasks[0].Backend, "claude")
		}
		if tasks.Tasks[0].Prompt != "test prompt" {
			t.Errorf("prompt = %q, want %q", tasks.Tasks[0].Prompt, "test prompt")
		}
	})

	t.Run("full tasks with all fields", func(t *testing.T) {
		input := `{
			"tasks": [
				{
					"backend": "claude",
					"prompt": "review auth module",
					"id": "task-1",
					"name": "Auth Review",
					"workdir": "/project",
					"model": "claude-opus-4-5-20251101",
					"extra": ["--flag1", "--flag2"],
					"approval_mode": "auto",
					"sandbox_mode": "workspace",
					"output_format": "json",
					"max_tokens": 4096,
					"max_turns": 10,
					"system_prompt": "Be helpful",
					"verbose": true,
					"dry_run": false,
					"tags": ["auth", "review"],
					"meta": {"priority": "high", "team": "security"}
				}
			],
			"max_parallel": 5,
			"fail_fast": true,
			"output_dir": "/output"
		}`

		var tasks ParallelTasks
		if err := json.Unmarshal([]byte(input), &tasks); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if tasks.MaxParallel != 5 {
			t.Errorf("max_parallel = %d, want 5", tasks.MaxParallel)
		}
		if !tasks.FailFast {
			t.Error("fail_fast should be true")
		}
		if tasks.OutputDir != "/output" {
			t.Errorf("output_dir = %q, want %q", tasks.OutputDir, "/output")
		}

		task := tasks.Tasks[0]
		if task.ID != "task-1" {
			t.Errorf("id = %q, want %q", task.ID, "task-1")
		}
		if task.Name != "Auth Review" {
			t.Errorf("name = %q, want %q", task.Name, "Auth Review")
		}
		if task.WorkDir != "/project" {
			t.Errorf("workdir = %q, want %q", task.WorkDir, "/project")
		}
		if task.Model != "claude-opus-4-5-20251101" {
			t.Errorf("model = %q, want %q", task.Model, "claude-opus-4-5-20251101")
		}
		if len(task.Extra) != 2 || task.Extra[0] != "--flag1" {
			t.Errorf("extra = %v, want [--flag1, --flag2]", task.Extra)
		}
		if task.ApprovalMode != "auto" {
			t.Errorf("approval_mode = %q, want %q", task.ApprovalMode, "auto")
		}
		if task.SandboxMode != "workspace" {
			t.Errorf("sandbox_mode = %q, want %q", task.SandboxMode, "workspace")
		}
		if task.MaxTokens != 4096 {
			t.Errorf("max_tokens = %d, want 4096", task.MaxTokens)
		}
		if task.MaxTurns != 10 {
			t.Errorf("max_turns = %d, want 10", task.MaxTurns)
		}
		if task.SystemPrompt != "Be helpful" {
			t.Errorf("system_prompt = %q, want %q", task.SystemPrompt, "Be helpful")
		}
		if !task.Verbose {
			t.Error("verbose should be true")
		}
		if len(task.Tags) != 2 {
			t.Errorf("tags length = %d, want 2", len(task.Tags))
		}
		if task.Meta["priority"] != "high" {
			t.Errorf("meta[priority] = %q, want %q", task.Meta["priority"], "high")
		}
	})

	t.Run("marshal and unmarshal round trip", func(t *testing.T) {
		original := ParallelTasks{
			Tasks: []ParallelTask{
				{
					Backend: "claude",
					Prompt:  "test",
					ID:      "task-1",
					Tags:    []string{"tag1"},
					Meta:    map[string]string{"key": "value"},
				},
			},
			MaxParallel: 3,
			FailFast:    true,
			OutputDir:   "/tmp/output",
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded ParallelTasks
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.MaxParallel != original.MaxParallel {
			t.Errorf("max_parallel = %d, want %d", decoded.MaxParallel, original.MaxParallel)
		}
		if decoded.FailFast != original.FailFast {
			t.Errorf("fail_fast = %v, want %v", decoded.FailFast, original.FailFast)
		}
		if len(decoded.Tasks) != len(original.Tasks) {
			t.Errorf("tasks length = %d, want %d", len(decoded.Tasks), len(original.Tasks))
		}
	})
}

func TestParallelTaskJSONSerialization(t *testing.T) {
	t.Run("omitempty fields excluded when empty", func(t *testing.T) {
		task := ParallelTask{
			Backend: "claude",
			Prompt:  "test",
		}

		data, err := json.Marshal(task)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		jsonStr := string(data)
		if jsonStr == "" {
			t.Fatal("marshaled JSON should not be empty")
		}

		// These fields should not appear in output due to omitempty
		omittedFields := []string{
			"workdir", "model", "extra", "approval_mode", "sandbox_mode",
			"output_format", "max_tokens", "max_turns", "system_prompt",
			"verbose", "dry_run", "id", "name", "tags", "meta",
		}

		for _, field := range omittedFields {
			if containsJSONKey(jsonStr, field) {
				t.Errorf("field %q should be omitted when empty, got JSON: %s", field, jsonStr)
			}
		}
	})
}

func TestTaskResultJSONSerialization(t *testing.T) {
	t.Run("full result", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 15, 10, 1, 30, 0, time.UTC)

		result := TaskResult{
			Index:     0,
			TaskID:    "task-1",
			TaskName:  "Test Task",
			Tags:      []string{"tag1", "tag2"},
			Meta:      map[string]string{"env": "test"},
			Backend:   "claude",
			ExitCode:  0,
			Error:     "",
			Output:    "Task completed successfully",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  90.0,
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded TaskResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.Index != result.Index {
			t.Errorf("index = %d, want %d", decoded.Index, result.Index)
		}
		if decoded.TaskID != result.TaskID {
			t.Errorf("task_id = %q, want %q", decoded.TaskID, result.TaskID)
		}
		if decoded.Backend != result.Backend {
			t.Errorf("backend = %q, want %q", decoded.Backend, result.Backend)
		}
		if decoded.ExitCode != result.ExitCode {
			t.Errorf("exit_code = %d, want %d", decoded.ExitCode, result.ExitCode)
		}
		if decoded.Duration != result.Duration {
			t.Errorf("duration = %f, want %f", decoded.Duration, result.Duration)
		}
	})

	t.Run("failed result", func(t *testing.T) {
		result := TaskResult{
			Index:    1,
			Backend:  "codex",
			ExitCode: 1,
			Error:    "command failed: timeout",
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded TaskResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.ExitCode != 1 {
			t.Errorf("exit_code = %d, want 1", decoded.ExitCode)
		}
		if decoded.Error != "command failed: timeout" {
			t.Errorf("error = %q, want %q", decoded.Error, "command failed: timeout")
		}
	})
}

func TestParallelCmdStructure(t *testing.T) {
	t.Run("Use field", func(t *testing.T) {
		if parallelCmd.Use != "parallel" {
			t.Errorf("Use = %q, want %q", parallelCmd.Use, "parallel")
		}
	})

	t.Run("Short field", func(t *testing.T) {
		want := "Run multiple AI tasks in parallel"
		if parallelCmd.Short != want {
			t.Errorf("Short = %q, want %q", parallelCmd.Short, want)
		}
	})

	t.Run("Long field is not empty", func(t *testing.T) {
		if parallelCmd.Long == "" {
			t.Error("Long description should not be empty")
		}
	})

	t.Run("RunE is set", func(t *testing.T) {
		if parallelCmd.RunE == nil {
			t.Error("RunE should be set")
		}
	})

	t.Run("flags are registered", func(t *testing.T) {
		flags := parallelCmd.Flags()

		fileFlag := flags.Lookup("file")
		if fileFlag == nil {
			t.Error("--file flag should be registered")
		} else if fileFlag.Shorthand != "f" {
			t.Errorf("--file shorthand = %q, want %q", fileFlag.Shorthand, "f")
		}

		maxParallelFlag := flags.Lookup("max-parallel")
		if maxParallelFlag == nil {
			t.Error("--max-parallel flag should be registered")
		}

		failFastFlag := flags.Lookup("fail-fast")
		if failFastFlag == nil {
			t.Error("--fail-fast flag should be registered")
		}

		jsonFlag := flags.Lookup("json")
		if jsonFlag == nil {
			t.Error("--json flag should be registered")
		}

		quietFlag := flags.Lookup("quiet")
		if quietFlag == nil {
			t.Error("--quiet flag should be registered")
		} else if quietFlag.Shorthand != "q" {
			t.Errorf("--quiet shorthand = %q, want %q", quietFlag.Shorthand, "q")
		}
	})
}

// containsJSONKey checks if a JSON string contains a specific key.
func containsJSONKey(jsonStr, key string) bool {
	// Simple check for "key": pattern
	pattern := `"` + key + `":`
	return jsonStr != "" && jsonContainsPattern(jsonStr, pattern)
}

func jsonContainsPattern(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		if s[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
}
