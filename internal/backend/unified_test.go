package backend

import (
	"strings"
	"testing"
)

// ==================== UnifiedOptions Tests ====================

func TestMapFromUnified_NilOptions(t *testing.T) {
	opts := MapFromUnified("claude", nil)
	if opts != nil {
		t.Error("expected nil for nil input")
	}
}

func TestMapFromUnified_BasicOptions(t *testing.T) {
	unified := &UnifiedOptions{
		WorkDir:     "/test/dir",
		Model:       "test-model",
		AllowedDirs: []string{"/dir1", "/dir2"},
	}

	opts := MapFromUnified("claude", unified)

	if opts.WorkDir != "/test/dir" {
		t.Errorf("expected WorkDir '/test/dir', got %q", opts.WorkDir)
	}
	if opts.Model != "test-model" {
		t.Errorf("expected Model 'test-model', got %q", opts.Model)
	}
	if len(opts.AllowedDirs) != 2 {
		t.Errorf("expected 2 AllowedDirs, got %d", len(opts.AllowedDirs))
	}
}

func TestMapFromUnified_ExtraFlags(t *testing.T) {
	unified := &UnifiedOptions{
		ExtraFlags: []string{"--custom1", "value1", "--custom2"},
	}

	opts := MapFromUnified("claude", unified)

	found := 0
	for _, f := range opts.ExtraFlags {
		if f == "--custom1" || f == "value1" || f == "--custom2" {
			found++
		}
	}
	if found != 3 {
		t.Errorf("expected all extra flags, found %d", found)
	}
}

// ==================== Model Mapping Tests ====================

func TestMapModel_Claude(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fast", "haiku"},
		{"quick", "haiku"},
		{"balanced", "sonnet"},
		{"default", "sonnet"},
		{"best", "opus"},
		{"powerful", "opus"},
		{"claude-opus-4-5-20251101", "claude-opus-4-5-20251101"}, // passthrough
		{"", ""},
	}

	m := newFlagMapper("claude")
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := m.mapModel(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMapModel_Gemini(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fast", "gemini-2.5-flash"},
		{"quick", "gemini-2.5-flash"},
		{"balanced", "gemini-2.5-pro"},
		{"default", "gemini-2.5-pro"},
		{"best", "gemini-2.5-pro"},
		{"powerful", "gemini-2.5-pro"},
		{"gemini-1.5-pro", "gemini-1.5-pro"}, // passthrough
	}

	m := newFlagMapper("gemini")
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := m.mapModel(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMapModel_Codex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fast", "gpt-4.1-mini"},
		{"quick", "gpt-4.1-mini"},
		{"balanced", "gpt-5.2"},
		{"default", "gpt-5.2"},
		{"best", "gpt-5-codex"},
		{"powerful", "gpt-5-codex"},
		{"o3", "o3"}, // passthrough
	}

	m := newFlagMapper("codex")
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := m.mapModel(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMapModel_UnknownBackend(t *testing.T) {
	m := newFlagMapper("unknown")
	result := m.mapModel("fast")
	if result != "fast" {
		t.Errorf("expected passthrough 'fast', got %q", result)
	}
}

// ==================== ApprovalMode Mapping Tests ====================

func TestMapApprovalMode_Claude(t *testing.T) {
	m := newFlagMapper("claude")

	tests := []struct {
		mode     ApprovalMode
		expected []string
	}{
		{ApprovalDefault, nil},
		{"", nil},
		{ApprovalAuto, []string{"--permission-mode", "acceptEdits"}},
		{ApprovalNone, []string{"--permission-mode", "dontAsk"}},
		{ApprovalAlways, []string{"--permission-mode", "default"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			result := m.mapApprovalMode(tt.mode)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMapApprovalMode_Gemini(t *testing.T) {
	m := newFlagMapper("gemini")

	tests := []struct {
		mode     ApprovalMode
		expected []string
	}{
		{ApprovalDefault, nil},
		{ApprovalAuto, []string{"--approval-mode", "auto_edit"}},
		{ApprovalNone, []string{"--yolo"}},
		{ApprovalAlways, []string{"--approval-mode", "default"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			result := m.mapApprovalMode(tt.mode)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMapApprovalMode_Codex(t *testing.T) {
	m := newFlagMapper("codex")

	tests := []struct {
		mode     ApprovalMode
		expected []string
	}{
		{ApprovalDefault, nil},
		{ApprovalAuto, []string{"--ask-for-approval", "on-request"}},
		{ApprovalNone, []string{"--ask-for-approval", "never"}},
		{ApprovalAlways, []string{"--ask-for-approval", "untrusted"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			result := m.mapApprovalMode(tt.mode)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ==================== SandboxMode Mapping Tests ====================

func TestMapSandboxMode_Claude(t *testing.T) {
	m := newFlagMapper("claude")

	// Claude doesn't have sandbox flags
	modes := []SandboxMode{SandboxDefault, SandboxReadOnly, SandboxWorkspace, SandboxFull, ""}
	for _, mode := range modes {
		result := m.mapSandboxMode(mode)
		if result != nil {
			t.Errorf("Claude should not have sandbox flags, got %v for mode %q", result, mode)
		}
	}
}

func TestMapSandboxMode_Gemini(t *testing.T) {
	m := newFlagMapper("gemini")

	tests := []struct {
		mode     SandboxMode
		expected []string
	}{
		{SandboxDefault, nil},
		{"", nil},
		{SandboxReadOnly, []string{"--sandbox"}},
		{SandboxWorkspace, []string{"--sandbox"}},
		{SandboxFull, nil},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			result := m.mapSandboxMode(tt.mode)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMapSandboxMode_Codex(t *testing.T) {
	m := newFlagMapper("codex")

	tests := []struct {
		mode     SandboxMode
		expected []string
	}{
		{SandboxDefault, nil},
		{SandboxReadOnly, []string{"--sandbox", "read-only"}},
		{SandboxWorkspace, []string{"--sandbox", "workspace-write"}},
		{SandboxFull, []string{"--sandbox", "danger-full-access"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			result := m.mapSandboxMode(tt.mode)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ==================== OutputFormat Mapping Tests ====================

func TestMapOutputFormat_Claude(t *testing.T) {
	m := newFlagMapper("claude")

	tests := []struct {
		format   OutputFormat
		expected []string
	}{
		{OutputDefault, nil},
		{"", nil},
		{OutputText, []string{"--output-format", "text"}},
		{OutputJSON, []string{"--output-format", "json"}},
		{OutputStreamJSON, []string{"--output-format", "stream-json", "--verbose"}}, // Claude requires --verbose for stream-json
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			result := m.mapOutputFormat(tt.format)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMapOutputFormat_Gemini(t *testing.T) {
	m := newFlagMapper("gemini")

	tests := []struct {
		format   OutputFormat
		expected []string
	}{
		{OutputDefault, nil},
		{OutputText, []string{"--output-format", "text"}},
		{OutputJSON, []string{"--output-format", "json"}},
		{OutputStreamJSON, []string{"--output-format", "stream-json"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			result := m.mapOutputFormat(tt.format)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMapOutputFormat_Codex(t *testing.T) {
	m := newFlagMapper("codex")

	tests := []struct {
		format   OutputFormat
		expected []string
	}{
		{OutputDefault, nil},
		{OutputText, nil}, // Codex doesn't have text format flag
		{OutputJSON, []string{"--json"}},
		{OutputStreamJSON, []string{"--json"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			result := m.mapOutputFormat(tt.format)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ==================== Verbose Mapping Tests ====================

func TestMapVerbose(t *testing.T) {
	tests := []struct {
		backend  string
		expected []string
	}{
		{"claude", []string{"--verbose"}},
		{"gemini", []string{"--debug"}},
		{"codex", []string{}},
		{"unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.backend, func(t *testing.T) {
			m := newFlagMapper(tt.backend)
			result := m.mapVerbose()
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ==================== DryRun Mapping Tests ====================

func TestMapDryRun(t *testing.T) {
	tests := []struct {
		backend  string
		expected []string
	}{
		{"claude", []string{}},
		{"gemini", []string{}},
		{"codex", []string{"--sandbox", "read-only"}},
		{"unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.backend, func(t *testing.T) {
			m := newFlagMapper(tt.backend)
			result := m.mapDryRun()
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ==================== MaxTokens Mapping Tests ====================

func TestMapMaxTokens(t *testing.T) {
	// MaxTokens is not mapped for any backend currently
	backends := []string{"claude", "gemini", "codex"}
	for _, backend := range backends {
		m := newFlagMapper(backend)
		result := m.mapMaxTokens(1000)
		if result != nil {
			t.Errorf("expected nil for %s, got %v", backend, result)
		}
	}
}

// ==================== MaxTurns Mapping Tests ====================

func TestMapMaxTurns(t *testing.T) {
	tests := []struct {
		backend  string
		turns    int
		expected []string
	}{
		{"claude", 10, []string{"--max-turns", "10"}},
		{"gemini", 10, nil},
		{"codex", 10, nil},
	}

	for _, tt := range tests {
		t.Run(tt.backend, func(t *testing.T) {
			m := newFlagMapper(tt.backend)
			result := m.mapMaxTurns(tt.turns)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ==================== SystemPrompt Mapping Tests ====================

func TestMapSystemPrompt(t *testing.T) {
	tests := []struct {
		backend  string
		prompt   string
		expected []string
	}{
		{"claude", "You are a helpful assistant", []string{"--system-prompt", "You are a helpful assistant"}},
		{"gemini", "You are a helpful assistant", nil},
		{"codex", "You are a helpful assistant", nil},
	}

	for _, tt := range tests {
		t.Run(tt.backend, func(t *testing.T) {
			m := newFlagMapper(tt.backend)
			result := m.mapSystemPrompt(tt.prompt)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ==================== Ephemeral Mapping Tests ====================

func TestMapEphemeral(t *testing.T) {
	tests := []struct {
		backend  string
		expected []string
	}{
		{"claude", []string{"--no-session-persistence"}},
		{"gemini", nil}, // Gemini doesn't have a native flag, cleanup is done post-execution
		{"codex", nil},  // Codex doesn't have a native flag, cleanup is done post-execution
		{"unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.backend, func(t *testing.T) {
			m := newFlagMapper(tt.backend)
			result := m.mapEphemeral()
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMapToOptions_Ephemeral(t *testing.T) {
	unified := &UnifiedOptions{
		Ephemeral: true,
	}

	// Claude should have --no-session-persistence flag
	claudeOpts := MapFromUnified("claude", unified)
	found := false
	for _, f := range claudeOpts.ExtraFlags {
		if f == "--no-session-persistence" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected --no-session-persistence in claude flags, got: %v", claudeOpts.ExtraFlags)
	}

	// Gemini should NOT have any ephemeral flag (handled post-execution)
	geminiOpts := MapFromUnified("gemini", unified)
	for _, f := range geminiOpts.ExtraFlags {
		if strings.Contains(f, "session") || strings.Contains(f, "ephemeral") {
			t.Errorf("gemini should not have ephemeral flags, got: %v", geminiOpts.ExtraFlags)
		}
	}

	// Codex should NOT have any ephemeral flag (handled post-execution)
	codexOpts := MapFromUnified("codex", unified)
	for _, f := range codexOpts.ExtraFlags {
		if strings.Contains(f, "session") || strings.Contains(f, "ephemeral") {
			t.Errorf("codex should not have ephemeral flags, got: %v", codexOpts.ExtraFlags)
		}
	}
}

// ==================== Full Integration Tests ====================

func TestMapToOptions_FullClaude(t *testing.T) {
	unified := &UnifiedOptions{
		WorkDir:      "/project",
		Model:        "fast",
		ApprovalMode: ApprovalAuto,
		SandboxMode:  SandboxWorkspace,
		OutputFormat: OutputStreamJSON,
		Verbose:      true,
		MaxTurns:     5,
		SystemPrompt: "Be helpful",
		ExtraFlags:   []string{"--custom"},
	}

	opts := MapFromUnified("claude", unified)

	if opts.WorkDir != "/project" {
		t.Errorf("expected WorkDir '/project', got %q", opts.WorkDir)
	}
	if opts.Model != "haiku" {
		t.Errorf("expected Model 'haiku', got %q", opts.Model)
	}

	// Check flags
	flags := strings.Join(opts.ExtraFlags, " ")
	expectedFlags := []string{
		"--permission-mode acceptEdits",
		"--output-format stream-json",
		"--verbose",
		"--max-turns 5",
		"--system-prompt",
		"--custom",
	}

	for _, expected := range expectedFlags {
		if !strings.Contains(flags, expected) {
			t.Errorf("expected flags to contain %q, got: %s", expected, flags)
		}
	}
}

func TestMapToOptions_FullGemini(t *testing.T) {
	unified := &UnifiedOptions{
		Model:        "balanced",
		ApprovalMode: ApprovalNone,
		SandboxMode:  SandboxReadOnly,
		OutputFormat: OutputJSON,
		Verbose:      true,
	}

	opts := MapFromUnified("gemini", unified)

	if opts.Model != "gemini-2.5-pro" {
		t.Errorf("expected Model 'gemini-2.5-pro', got %q", opts.Model)
	}

	flags := strings.Join(opts.ExtraFlags, " ")
	expectedFlags := []string{
		"--yolo",
		"--sandbox",
		"--output-format json",
		"--debug",
	}

	for _, expected := range expectedFlags {
		if !strings.Contains(flags, expected) {
			t.Errorf("expected flags to contain %q, got: %s", expected, flags)
		}
	}
}

func TestMapToOptions_FullCodex(t *testing.T) {
	unified := &UnifiedOptions{
		Model:        "best",
		ApprovalMode: ApprovalAuto,
		SandboxMode:  SandboxFull,
		OutputFormat: OutputJSON,
		DryRun:       true,
	}

	opts := MapFromUnified("codex", unified)

	if opts.Model != "gpt-5-codex" {
		t.Errorf("expected Model 'gpt-5-codex', got %q", opts.Model)
	}

	flags := strings.Join(opts.ExtraFlags, " ")
	expectedFlags := []string{
		"--ask-for-approval on-request",
		"--sandbox danger-full-access",
		"--json",
	}

	for _, expected := range expectedFlags {
		if !strings.Contains(flags, expected) {
			t.Errorf("expected flags to contain %q, got: %s", expected, flags)
		}
	}
}

// ==================== BuildCommandUnified Tests ====================

func TestClaude_BuildCommandUnified(t *testing.T) {
	b := &Claude{}

	unified := &UnifiedOptions{
		Model:        "fast",
		ApprovalMode: ApprovalAuto,
		OutputFormat: OutputStreamJSON,
	}

	cmd := b.BuildCommandUnified("test prompt", unified)

	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--model haiku") {
		t.Errorf("expected --model haiku, got: %s", args)
	}
	if !strings.Contains(args, "--permission-mode acceptEdits") {
		t.Errorf("expected --permission-mode acceptEdits, got: %s", args)
	}
	if !strings.Contains(args, "--output-format stream-json") {
		t.Errorf("expected --output-format stream-json, got: %s", args)
	}
	if !strings.Contains(args, "test prompt") {
		t.Errorf("expected prompt in args, got: %s", args)
	}
}

func TestClaude_ResumeCommandUnified(t *testing.T) {
	b := &Claude{}

	unified := &UnifiedOptions{
		Model: "balanced",
	}

	cmd := b.ResumeCommandUnified("session-123", "continue", unified)

	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--resume session-123") {
		t.Errorf("expected --resume session-123, got: %s", args)
	}
	if !strings.Contains(args, "--model sonnet") {
		t.Errorf("expected --model sonnet, got: %s", args)
	}
}

func TestCodex_BuildCommandUnified(t *testing.T) {
	b := &Codex{}

	unified := &UnifiedOptions{
		Model:        "quick",
		ApprovalMode: ApprovalNone,
		OutputFormat: OutputJSON,
	}

	cmd := b.BuildCommandUnified("test prompt", unified)

	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--model gpt-4.1-mini") {
		t.Errorf("expected --model gpt-4.1-mini, got: %s", args)
	}
	if !strings.Contains(args, "--ask-for-approval never") {
		t.Errorf("expected --ask-for-approval never, got: %s", args)
	}
	if !strings.Contains(args, "--json") {
		t.Errorf("expected --json, got: %s", args)
	}
}

func TestGemini_BuildCommandUnified(t *testing.T) {
	b := &Gemini{}

	unified := &UnifiedOptions{
		Model:       "best",
		SandboxMode: SandboxWorkspace,
		Verbose:     true,
	}

	cmd := b.BuildCommandUnified("test prompt", unified)

	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--model gemini-2.5-pro") {
		t.Errorf("expected --model gemini-2.5-pro, got: %s", args)
	}
	if !strings.Contains(args, "--sandbox") {
		t.Errorf("expected --sandbox, got: %s", args)
	}
	if !strings.Contains(args, "--debug") {
		t.Errorf("expected --debug, got: %s", args)
	}
}

// ==================== Validation Tests ====================

func TestValidateExtraFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		wantErr bool
	}{
		{
			name:    "empty flags",
			flags:   []string{},
			wantErr: false,
		},
		{
			name:    "valid flags with values",
			flags:   []string{"--model=gpt-4", "-v"},
			wantErr: false,
		},
		{
			name:    "valid flag with value",
			flags:   []string{"--output-format=json"},
			wantErr: false,
		},
		{
			name:    "unknown flag --dangerously-skip-permissions (not in allowlist)",
			flags:   []string{"--dangerously-skip-permissions"},
			wantErr: true,
		},
		{
			name:    "unknown flag --no-verify (not in allowlist)",
			flags:   []string{"--no-verify"},
			wantErr: true,
		},
		{
			name:    "unknown flag --force (not in allowlist)",
			flags:   []string{"--force"},
			wantErr: true,
		},
		{
			name:    "unknown flag -f (not in allowlist)",
			flags:   []string{"-f"},
			wantErr: true,
		},
		{
			name:    "invalid flag format - no dash",
			flags:   []string{"model"},
			wantErr: true,
		},
		{
			name:    "unknown flag with value (not in allowlist)",
			flags:   []string{"--force=true"},
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid",
			flags:   []string{"--model=gpt-4", "--force"},
			wantErr: true,
		},
		{
			name:    "case insensitive - uppercase allowed",
			flags:   []string{"--MODEL=gpt-4"},
			wantErr: false,
		},
		{
			name:    "case insensitive - mixed case allowed",
			flags:   []string{"--Output-Format=json"},
			wantErr: false,
		},
		// Tests for flag-value pair format (--flag value instead of --flag=value)
		{
			name:    "valid flag-value pair format",
			flags:   []string{"--add-dir", "./docs"},
			wantErr: false,
		},
		{
			name:    "valid flag-value pair with absolute path",
			flags:   []string{"--add-dir", "/home/user/project"},
			wantErr: false,
		},
		{
			name:    "valid model flag with value token",
			flags:   []string{"--model", "gpt-4"},
			wantErr: false,
		},
		{
			name:    "multiple flag-value pairs",
			flags:   []string{"--model", "gpt-4", "--add-dir", "./docs", "-v"},
			wantErr: false,
		},
		{
			name:    "mixed equals and space formats",
			flags:   []string{"--model=gpt-4", "--add-dir", "./src"},
			wantErr: false,
		},
		{
			name:    "standalone non-flag token without preceding flag is rejected",
			flags:   []string{"./docs"},
			wantErr: true,
		},
		{
			name:    "value tokens cannot be first",
			flags:   []string{"value-first", "--model"},
			wantErr: true,
		},
		// Tests for consecutive flags (no bypass via flag without value)
		{
			name:    "consecutive flags without values - verbose then model",
			flags:   []string{"--verbose", "--model", "gpt-4"},
			wantErr: false,
		},
		{
			name:    "consecutive flags - dangerous flag after verbose should be rejected",
			flags:   []string{"--verbose", "--dangerously-skip-permissions"},
			wantErr: true,
		},
		{
			name:    "flag starting with dash is always validated even after flag without =",
			flags:   []string{"--model", "--force"},
			wantErr: true, // --force is not allowed
		},
		{
			name:    "short flag followed by long flag",
			flags:   []string{"-v", "--model", "gpt-4"},
			wantErr: false,
		},
		{
			name:    "multiple consecutive flags without values",
			flags:   []string{"-v", "-q", "-h"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExtraFlags(tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExtraFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateExtraFlagsForBackend(t *testing.T) {
	tests := []struct {
		name    string
		backend string
		flags   []string
		wantErr bool
	}{
		{
			name:    "claude valid flags",
			backend: "claude",
			flags:   []string{"--model=opus", "--verbose", "--permission-mode=acceptEdits"},
			wantErr: false,
		},
		{
			name:    "claude invalid flag",
			backend: "claude",
			flags:   []string{"--json"}, // json is codex-only
			wantErr: true,
		},
		{
			name:    "codex valid flags",
			backend: "codex",
			flags:   []string{"--model=gpt-5", "--json", "--sandbox=read-only"},
			wantErr: false,
		},
		{
			name:    "codex invalid flag",
			backend: "codex",
			flags:   []string{"--permission-mode=acceptEdits"}, // permission-mode is claude-only
			wantErr: true,
		},
		{
			name:    "gemini valid flags",
			backend: "gemini",
			flags:   []string{"--model=gemini-2.5-pro", "--yolo", "--debug"},
			wantErr: false,
		},
		{
			name:    "common flags allowed for all backends",
			backend: "claude",
			flags:   []string{"-v", "-m", "--help"},
			wantErr: false,
		},
		{
			name:    "dangerous flag blocked via allowlist",
			backend: "claude",
			flags:   []string{"--DANGEROUSLY-skip-PERMISSIONS"},
			wantErr: true,
		},
		{
			name:    "case insensitive matching",
			backend: "claude",
			flags:   []string{"--VERBOSE", "--Model=opus"},
			wantErr: false,
		},
		{
			name:    "unknown backend uses common flags only",
			backend: "unknown",
			flags:   []string{"-v", "-h"},
			wantErr: false,
		},
		{
			name:    "unknown backend rejects backend-specific flags",
			backend: "unknown",
			flags:   []string{"--verbose"}, // verbose is backend-specific
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExtraFlagsForBackend(tt.backend, tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExtraFlagsForBackend(%q, %v) error = %v, wantErr %v",
					tt.backend, tt.flags, err, tt.wantErr)
			}
		})
	}
}

func TestExtractFlagName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"--flag", "--flag"},
		{"--flag=value", "--flag"},
		{"-f", "-f"},
		{"-f=value", "-f"},
		{"--long-flag=some=value", "--long-flag"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractFlagName(tt.input)
			if result != tt.expected {
				t.Errorf("extractFlagName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== Helper ====================

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
