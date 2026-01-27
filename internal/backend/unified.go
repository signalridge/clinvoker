package backend

import (
	"fmt"
	"strings"
)

// dangerousFlags contains flags that could bypass security controls.
var dangerousFlags = map[string]bool{
	"--dangerously-skip-permissions": true,
	"--no-verify":                    true,
	"--skip-hooks":                   true,
	"--force":                        true,
	"-f":                             true,
}

// ValidateExtraFlags validates that extra flags don't contain dangerous options.
func ValidateExtraFlags(flags []string) error {
	for _, flag := range flags {
		// Normalize flag (handle --flag=value format)
		normalizedFlag := flag
		if idx := strings.Index(flag, "="); idx != -1 {
			normalizedFlag = flag[:idx]
		}

		// Check for dangerous flags
		if dangerousFlags[normalizedFlag] {
			return fmt.Errorf("dangerous flag not allowed: %s", flag)
		}

		// Flags must start with - or --
		if !strings.HasPrefix(flag, "-") {
			return fmt.Errorf("invalid flag format: %s (must start with - or --)", flag)
		}
	}
	return nil
}

// UnifiedOptions provides a backend-agnostic way to configure AI CLI commands.
// These options are automatically mapped to backend-specific flags.
type UnifiedOptions struct {
	// WorkDir is the working directory for the command.
	WorkDir string

	// Model specifies the model to use (will be mapped to backend-specific model names).
	Model string

	// ApprovalMode controls how the backend asks for user approval.
	ApprovalMode ApprovalMode

	// SandboxMode controls file/network access restrictions.
	SandboxMode SandboxMode

	// OutputFormat controls the output format.
	OutputFormat OutputFormat

	// AllowedDirs specifies directories the backend can access.
	AllowedDirs []string

	// Interactive enables interactive mode (TUI).
	Interactive bool

	// Verbose enables verbose output.
	Verbose bool

	// DryRun simulates execution without making changes.
	DryRun bool

	// MaxTokens limits the maximum tokens for the response.
	MaxTokens int

	// MaxTurns limits the number of agentic turns.
	MaxTurns int

	// SystemPrompt provides a custom system prompt.
	SystemPrompt string

	// ExtraFlags contains additional backend-specific flags.
	ExtraFlags []string
}

// ApprovalMode controls how the backend asks for user approval.
type ApprovalMode string

const (
	// ApprovalDefault uses the backend's default approval behavior.
	ApprovalDefault ApprovalMode = "default"

	// ApprovalAuto automatically approves safe operations.
	ApprovalAuto ApprovalMode = "auto"

	// ApprovalNone disables all approval prompts (dangerous).
	ApprovalNone ApprovalMode = "none"

	// ApprovalAlways always asks for approval.
	ApprovalAlways ApprovalMode = "always"
)

// SandboxMode controls file/network access restrictions.
type SandboxMode string

const (
	// SandboxDefault uses the backend's default sandbox behavior.
	SandboxDefault SandboxMode = "default"

	// SandboxReadOnly allows only read access.
	SandboxReadOnly SandboxMode = "read-only"

	// SandboxWorkspace allows read/write in workspace only.
	SandboxWorkspace SandboxMode = "workspace"

	// SandboxFull allows full system access (dangerous).
	SandboxFull SandboxMode = "full"
)

// OutputFormat controls the output format.
type OutputFormat string

const (
	// OutputDefault uses the backend's default output format.
	OutputDefault OutputFormat = "default"

	// OutputText outputs plain text.
	OutputText OutputFormat = "text"

	// OutputJSON outputs JSON.
	OutputJSON OutputFormat = "json"

	// OutputStreamJSON outputs streaming JSON (NDJSON/JSONL).
	OutputStreamJSON OutputFormat = "stream-json"
)

// flagMapper maps unified options to backend-specific flags.
type flagMapper struct {
	backend string
}

// newFlagMapper creates a new flag mapper for a backend.
func newFlagMapper(backend string) *flagMapper {
	return &flagMapper{backend: backend}
}

// MapToOptions converts UnifiedOptions to backend-specific Options.
func (m *flagMapper) MapToOptions(unified *UnifiedOptions) *Options {
	if unified == nil {
		return nil
	}

	opts := &Options{
		WorkDir:     unified.WorkDir,
		Model:       m.mapModel(unified.Model),
		AllowedDirs: unified.AllowedDirs,
		ExtraFlags:  make([]string, 0),
	}

	// Add approval mode flags
	opts.ExtraFlags = append(opts.ExtraFlags, m.mapApprovalMode(unified.ApprovalMode)...)

	// Add sandbox mode flags
	opts.ExtraFlags = append(opts.ExtraFlags, m.mapSandboxMode(unified.SandboxMode)...)

	// Add output format flags
	opts.ExtraFlags = append(opts.ExtraFlags, m.mapOutputFormat(unified.OutputFormat)...)

	// Add other flags
	if unified.Verbose {
		opts.ExtraFlags = append(opts.ExtraFlags, m.mapVerbose()...)
	}

	if unified.DryRun {
		opts.ExtraFlags = append(opts.ExtraFlags, m.mapDryRun()...)
	}

	if unified.MaxTokens > 0 {
		opts.ExtraFlags = append(opts.ExtraFlags, m.mapMaxTokens(unified.MaxTokens)...)
	}

	if unified.MaxTurns > 0 {
		opts.ExtraFlags = append(opts.ExtraFlags, m.mapMaxTurns(unified.MaxTurns)...)
	}

	if unified.SystemPrompt != "" {
		opts.ExtraFlags = append(opts.ExtraFlags, m.mapSystemPrompt(unified.SystemPrompt)...)
	}

	// Add any extra flags from user
	opts.ExtraFlags = append(opts.ExtraFlags, unified.ExtraFlags...)

	return opts
}

// mapModel maps unified model names to backend-specific names.
func (m *flagMapper) mapModel(model string) string {
	if model == "" {
		return ""
	}

	// Map unified model aliases to backend-specific names
	switch m.backend {
	case "claude":
		return m.mapClaudeModel(model)
	case "gemini":
		return m.mapGeminiModel(model)
	case "codex":
		return m.mapCodexModel(model)
	default:
		return model
	}
}

func (m *flagMapper) mapClaudeModel(model string) string {
	switch model {
	case "fast", "quick":
		return "haiku"
	case "balanced", "default":
		return "sonnet"
	case "best", "powerful":
		return "opus"
	default:
		return model
	}
}

func (m *flagMapper) mapGeminiModel(model string) string {
	switch model {
	case "fast", "quick":
		return "gemini-2.5-flash"
	case "balanced", "default", "best", "powerful":
		return "gemini-2.5-pro"
	default:
		return model
	}
}

func (m *flagMapper) mapCodexModel(model string) string {
	switch model {
	case "fast", "quick":
		return "gpt-4.1-mini"
	case "balanced", "default":
		return "gpt-5.2"
	case "best", "powerful":
		return "gpt-5-codex"
	default:
		return model
	}
}

// mapApprovalMode maps approval mode to backend-specific flags.
func (m *flagMapper) mapApprovalMode(mode ApprovalMode) []string {
	if mode == "" || mode == ApprovalDefault {
		return nil
	}

	switch m.backend {
	case "claude":
		switch mode {
		case ApprovalAuto:
			return []string{"--permission-mode", "acceptEdits"}
		case ApprovalNone:
			return []string{"--permission-mode", "dontAsk"}
		case ApprovalAlways:
			return []string{"--permission-mode", "default"}
		}

	case "gemini":
		switch mode {
		case ApprovalAuto:
			return []string{"--approval-mode", "auto_edit"}
		case ApprovalNone:
			return []string{"--yolo"}
		case ApprovalAlways:
			return []string{"--approval-mode", "default"}
		}

	case "codex":
		switch mode {
		case ApprovalAuto:
			return []string{"--ask-for-approval", "on-request"}
		case ApprovalNone:
			return []string{"--ask-for-approval", "never"}
		case ApprovalAlways:
			return []string{"--ask-for-approval", "untrusted"}
		}
	}

	return nil
}

// mapSandboxMode maps sandbox mode to backend-specific flags.
func (m *flagMapper) mapSandboxMode(mode SandboxMode) []string {
	if mode == "" || mode == SandboxDefault {
		return nil
	}

	switch m.backend {
	case "claude":
		// Claude doesn't have a direct sandbox flag
		return nil

	case "gemini":
		switch mode {
		case SandboxReadOnly, SandboxWorkspace:
			return []string{"--sandbox"}
		case SandboxFull:
			return nil // No sandbox
		}

	case "codex":
		switch mode {
		case SandboxReadOnly:
			return []string{"--sandbox", "read-only"}
		case SandboxWorkspace:
			return []string{"--sandbox", "workspace-write"}
		case SandboxFull:
			return []string{"--sandbox", "danger-full-access"}
		}
	}

	return nil
}

// mapOutputFormat maps output format to backend-specific flags.
func (m *flagMapper) mapOutputFormat(format OutputFormat) []string {
	if format == "" || format == OutputDefault {
		return nil
	}

	switch m.backend {
	case "claude":
		switch format {
		case OutputText:
			return []string{"--output-format", "text"}
		case OutputJSON:
			return []string{"--output-format", "json"}
		case OutputStreamJSON:
			// Claude requires --verbose for stream-json output
			return []string{"--output-format", "stream-json", "--verbose"}
		}

	case "gemini":
		switch format {
		case OutputText:
			return []string{"--output-format", "text"}
		case OutputJSON:
			return []string{"--output-format", "json"}
		case OutputStreamJSON:
			return []string{"--output-format", "stream-json"}
		}

	case "codex":
		switch format {
		case OutputJSON, OutputStreamJSON:
			return []string{"--json"}
		}
	}

	return nil
}

// mapVerbose returns backend-specific verbose flags.
func (m *flagMapper) mapVerbose() []string {
	switch m.backend {
	case "claude":
		return []string{"--verbose"}
	case "gemini":
		return []string{"--debug"}
	case "codex":
		return []string{} // Codex doesn't have a direct verbose flag
	}
	return nil
}

// mapDryRun returns backend-specific dry-run flags.
func (m *flagMapper) mapDryRun() []string {
	switch m.backend {
	case "claude":
		return []string{} // Claude uses permission modes instead
	case "gemini":
		return []string{} // Gemini uses sandbox instead
	case "codex":
		return []string{"--sandbox", "read-only"}
	}
	return nil
}

// mapMaxTokens returns backend-specific max tokens flags.
func (m *flagMapper) mapMaxTokens(_ int) []string {
	// Most backends don't expose this directly via CLI
	return nil
}

// mapMaxTurns returns backend-specific max turns flags.
func (m *flagMapper) mapMaxTurns(turns int) []string {
	switch m.backend {
	case "claude":
		return []string{"--max-turns", fmt.Sprintf("%d", turns)}
	}
	return nil
}

// mapSystemPrompt returns backend-specific system prompt flags.
func (m *flagMapper) mapSystemPrompt(prompt string) []string {
	switch m.backend {
	case "claude":
		return []string{"--system-prompt", prompt}
	}
	return nil
}

// MapFromUnified is a convenience function to map unified options to backend options.
func MapFromUnified(backend string, unified *UnifiedOptions) *Options {
	return newFlagMapper(backend).MapToOptions(unified)
}
