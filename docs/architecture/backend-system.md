---
title: Backend System
description: Deep dive into clinvoker's backend abstraction layer.
---

# Backend System

Deep dive into clinvoker's backend abstraction layer and how different AI CLIs are unified.

## Backend Interface

All backends implement the `Backend` interface:

```go
type Backend interface {
    Name() string
    IsAvailable() bool
    BuildCommand(prompt string, opts *Options) *exec.Cmd
    ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd
    BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd
    ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd
    ParseOutput(rawOutput string) string
    ParseJSONResponse(rawOutput string) (*UnifiedResponse, error)
    SeparateStderr() bool
}
```

## Backend Implementations

### Claude Backend

Wraps Claude Code CLI:

```go
const (
    claudeBin = "claude"
)

func (b *ClaudeBackend) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
    args := []string{"--output-format", "stream-json"}

    if opts.WorkDir != "" {
        args = append(args, "--working-dir", opts.WorkDir)
    }

    // Map unified options to Claude flags
    switch opts.ApprovalMode {
    case "auto":
        args = append(args, "--auto-accept")
    case "none":
        args = append(args, "--no-approval")
    }

    args = append(args, prompt)

    return exec.Command(claudeBin, args...)
}
```

### Codex Backend

Wraps Codex CLI:

```go
const (
    codexBin = "codex"
)

func (b *CodexBackend) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
    args := []string{"--format", "json"}

    if opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    if opts.WorkDir != "" {
        args = append(args, "--workdir", opts.WorkDir)
    }

    args = append(args, prompt)

    return exec.Command(codexBin, args...)
}
```

### Gemini Backend

Wraps Gemini CLI:

```go
const (
    geminiBin = "gemini"
)

func (b *GeminiBackend) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
    args := []string{"--output", "json"}

    if opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)

    return exec.Command(geminiBin, args...)
}
```

## Unified Options

The `UnifiedOptions` struct normalizes backend-specific flags:

```go
type UnifiedOptions struct {
    WorkDir       string
    Model         string
    ApprovalMode  string  // default, auto, none, always
    SandboxMode   string  // default, read-only, workspace, full
    OutputFormat  OutputFormat
    MaxTokens     int
    MaxTurns      int
    Verbose       bool
    Ephemeral     bool
}
```

## Output Parsing

Each backend parses its native output into a unified format:

### JSON Response Parsing

```go
func (b *ClaudeBackend) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
    // Parse Claude's JSON output
    var claudeResp struct {
        Content   string `json:"content"`
        SessionID string `json:"session_id"`
        Model     string `json:"model"`
        Usage     struct {
            InputTokens  int `json:"input_tokens"`
            OutputTokens int `json:"output_tokens"`
        } `json:"usage"`
    }

    if err := json.Unmarshal([]byte(rawOutput), &claudeResp); err != nil {
        return nil, err
    }

    return &UnifiedResponse{
        Content:   claudeResp.Content,
        SessionID: claudeResp.SessionID,
        Model:     claudeResp.Model,
        Usage: &TokenUsage{
            InputTokens:  claudeResp.Usage.InputTokens,
            OutputTokens: claudeResp.Usage.OutputTokens,
        },
    }, nil
}
```

### Text Output Parsing

```go
func (b *ClaudeBackend) ParseOutput(rawOutput string) string {
    // Extract clean text from various output formats
    // Remove ANSI codes, prompts, etc.
    cleaned := stripANSI(rawOutput)
    cleaned = extractResponse(cleaned)
    return strings.TrimSpace(cleaned)
}
```

## Backend Selection

Backends are selected based on:

1. **Explicit flag**: `--backend claude`
2. **Configuration**: `default_backend: claude`
3. **Availability**: First available backend
4. **Model mapping**: From SDK requests

```go
func Get(name string) (Backend, error) {
    switch name {
    case BackendClaude:
        return &ClaudeBackend{}, nil
    case BackendCodex:
        return &CodexBackend{}, nil
    case BackendGemini:
        return &GeminiBackend{}, nil
    default:
        return nil, fmt.Errorf("unknown backend: %s", name)
    }
}
```

## Backend Availability

Check if a backend CLI is installed:

```go
func (b *ClaudeBackend) IsAvailable() bool {
    _, err := exec.LookPath(claudeBin)
    return err == nil
}
```

## Adding a New Backend

To add a new backend:

1. Create a new file: `internal/backend/newbackend.go`
2. Implement the `Backend` interface
3. Register in `backend.go`
4. Add configuration options

Example:

```go
// internal/backend/newbackend.go
package backend

type NewBackend struct{}

func (b *NewBackend) Name() string { return "newbackend" }
func (b *NewBackend) IsAvailable() bool { ... }
func (b *NewBackend) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd { ... }
// ... implement other methods
```

## Backend Pool (Future)

For high-throughput scenarios, backends can be pooled:

```go
type BackendPool struct {
    backends []Backend
    queue    chan *Task
}

func (p *BackendPool) Execute(ctx context.Context, task *Task) (*Result, error) {
    // Round-robin or least-loaded selection
    backend := p.selectBackend()
    return backend.Execute(ctx, task)
}
```

## Best Practices

### 1. Command Building

- Always validate paths before execution
- Escape arguments properly
- Support both interactive and batch modes

### 2. Output Parsing

- Handle partial/invalid JSON gracefully
- Strip control characters
- Preserve error messages

### 3. Error Handling

- Distinguish between backend errors and system errors
- Provide clear error messages
- Include troubleshooting hints

## Related Documentation

- [Adding Backends](../development/adding-backends.md) - Step-by-step guide
- [Architecture Overview](overview.md) - High-level architecture
