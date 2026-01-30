# Adding New Backends

Guide to implementing new AI backends for clinvk.

## Overview

clinvk's backend system is designed to be extensible. Each backend is a Go implementation that wraps an external CLI tool.

## Backend Interface

All backends must implement the `Backend` interface:

```go
type Backend interface {
    // Name returns the backend identifier
    Name() string

    // IsAvailable checks if the backend CLI is installed
    IsAvailable() bool

    // BuildCommand creates the command for a new prompt
    BuildCommand(prompt string, opts *Options) *exec.Cmd

    // ResumeCommand creates the command to resume a session
    ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd

    // BuildCommandUnified creates command with unified options
    BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd

    // ResumeCommandUnified resumes with unified options
    ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd

    // ParseOutput extracts the response from raw output
    ParseOutput(rawOutput string) string

    // ParseJSONResponse parses JSON-formatted output
    ParseJSONResponse(rawOutput string) (*UnifiedResponse, error)

    // SeparateStderr indicates if stderr should be captured separately
    SeparateStderr() bool
}
```text

## Step-by-Step Guide

### 1. Create the Backend File

Create `internal/backend/mybackend.go`:

```go
package backend

import (
    "os/exec"
)

// MyBackend implements the Backend interface for MyAI CLI.
type MyBackend struct{}

// NewMyBackend creates a new MyBackend instance.
func NewMyBackend() *MyBackend {
    return &MyBackend{}
}

// Name returns the backend identifier.
func (b *MyBackend) Name() string {
    return "mybackend"
}

// IsAvailable checks if 'myai' is in PATH.
func (b *MyBackend) IsAvailable() bool {
    _, err := exec.LookPath("myai")
    return err == nil
}
```

### 2. Implement Command Building

```go
func (b *MyBackend) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{}

    if opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    if opts.Workdir != "" {
        args = append(args, "--workdir", opts.Workdir)
    }

    args = append(args, prompt)

    cmd := exec.Command("myai", args...)
    return cmd
}

func (b *MyBackend) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
    args := []string{"--session", sessionID}

    if opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    if prompt != "" {
        args = append(args, prompt)
    }

    cmd := exec.Command("myai", args...)
    return cmd
}
```text

### 3. Implement Unified Options

```go
func (b *MyBackend) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
    args := []string{}

    if opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    // Map unified approval mode to backend-specific flag
    switch opts.ApprovalMode {
    case "auto":
        args = append(args, "--auto-approve")
    case "none":
        args = append(args, "--no-approve")
    }

    // Map unified sandbox mode
    switch opts.SandboxMode {
    case "read-only":
        args = append(args, "--read-only")
    case "workspace":
        args = append(args, "--workspace-only")
    }

    if opts.MaxTokens > 0 {
        args = append(args, "--max-tokens", strconv.Itoa(opts.MaxTokens))
    }

    args = append(args, prompt)
    return exec.Command("myai", args...)
}

func (b *MyBackend) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
    // Similar implementation with --session flag
}
```

### 4. Implement Output Parsing

```go
func (b *MyBackend) ParseOutput(rawOutput string) string {
    // Extract meaningful response from raw output
    // This might involve stripping metadata, formatting, etc.
    return strings.TrimSpace(rawOutput)
}

func (b *MyBackend) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
    // Parse JSON output format if supported
    var response UnifiedResponse
    if err := json.Unmarshal([]byte(rawOutput), &response); err != nil {
        return nil, err
    }
    return &response, nil
}

func (b *MyBackend) SeparateStderr() bool {
    // Return true if stderr should be captured separately
    return false
}
```text

### 5. Register the Backend

Edit `internal/backend/registry.go`:

```go
func init() {
    Register("mybackend", NewMyBackend())
}
```

### 6. Add Configuration

Edit `internal/config/config.go`:

```go
type BackendsConfig struct {
    Claude    BackendConfig `mapstructure:"claude"`
    Codex     BackendConfig `mapstructure:"codex"`
    Gemini    BackendConfig `mapstructure:"gemini"`
    MyBackend BackendConfig `mapstructure:"mybackend"` // Add this
}
```yaml

Add environment variable binding:

```go
viper.BindEnv("backends.mybackend.model", "CLINVK_MYBACKEND_MODEL")
```

### 7. Write Tests

Create `internal/backend/mybackend_test.go`:

```go
package backend

import (
    "testing"
)

func TestMyBackend_Name(t *testing.T) {
    b := NewMyBackend()
    if b.Name() != "mybackend" {
        t.Errorf("expected 'mybackend', got %s", b.Name())
    }
}

func TestMyBackend_BuildCommand(t *testing.T) {
    b := NewMyBackend()
    opts := &Options{Model: "test-model"}

    cmd := b.BuildCommand("test prompt", opts)

    // Verify command arguments
    args := cmd.Args
    if args[0] != "myai" {
        t.Errorf("expected 'myai', got %s", args[0])
    }
}
```text

### 8. Update Documentation

Add documentation:

- `docs/user-guide/backends/mybackend.md`
- Update `docs/user-guide/backends/index.md`
- Update `README.md` if needed
- Add to `config.example.yaml`

## Best Practices

### Error Handling

```go
func (b *MyBackend) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    if prompt == "" {
        // Return nil or handle error appropriately
        return nil
    }
    // ...
}
```

### Flag Mapping

Create clear mappings between unified options and backend-specific flags:

| Unified Option | MyBackend Flag |
|----------------|----------------|
| `approval_mode: auto` | `--auto-approve` |
| `sandbox_mode: read-only` | `--read-only` |
| `max_tokens: N` | `--max-tokens N` |

### Testing with Mocks

Use the mock package for testing:

```go
import "github.com/signalridge/clinvoker/internal/mock"

func TestWithMockBackend(t *testing.T) {
    mock := mock.NewMockBackend("test",
        mock.WithParseOutput("mocked output"),
        mock.WithAvailable(true),
    )

    // Use mock in tests
}
```

## Checklist

- [ ] Implement all `Backend` interface methods
- [ ] Register backend in `registry.go`
- [ ] Add configuration support
- [ ] Add environment variable binding
- [ ] Write unit tests
- [ ] Add user documentation
- [ ] Update `config.example.yaml`
- [ ] Test with real backend CLI
