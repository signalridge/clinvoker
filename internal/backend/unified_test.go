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
