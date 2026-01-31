---
title: Backend System
description: Deep dive into clinvoker's backend abstraction layer, registry pattern, and unified interface.
---

# Backend System

This document provides a comprehensive deep dive into clinvoker's backend abstraction layer, explaining how different AI CLI tools are unified under a common interface, the registry pattern implementation, thread-safe design, and how to extend the system with new backends.

## Backend Interface Design

The `Backend` interface (`internal/backend/backend.go:16-46`) is the core abstraction that enables clinvoker to work with multiple AI CLI tools seamlessly:

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
```typescript

### Interface Design Rationale

The interface is designed around the lifecycle of an AI interaction:

1. **Discovery**: `Name()` and `IsAvailable()` for backend identification and detection
2. **Command Building**: `BuildCommand*` methods create executable commands
3. **Session Resumption**: `ResumeCommand*` methods continue existing conversations
4. **Output Processing**: `ParseOutput()` and `ParseJSONResponse()` normalize responses
5. **Error Handling**: `SeparateStderr()` determines stderr handling strategy

## Registry Pattern

The backend registry (`internal/backend/registry.go`) manages backend registration and lookup using a thread-safe registry pattern.

### Registry Structure

```mermaid
flowchart TB
    subgraph Registry["Registry (internal/backend/registry.go:11-16)"]
        RWMU[sync.RWMutex]
        BACKENDS[map[string]Backend]
        CACHE[availabilityCache]
        TTL[30s TTL]
    end

    subgraph Operations["Registry Operations"]
        REGISTER[Register]
        UNREGISTER[Unregister]
        GET[Get]
        LIST[List]
        AVAILABLE[Available]
    end

    RWMU --> BACKENDS
    BACKENDS --> CACHE
    CACHE --> TTL

    REGISTER --> RWMU
    UNREGISTER --> RWMU
    GET --> RWMU
    LIST --> RWMU
    AVAILABLE --> CACHE
```text

### Thread-Safe Design

The registry uses `sync.RWMutex` for concurrent access:

```go
// Read operations use RLock for concurrent reads
func (r *Registry) Get(name string) (Backend, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // ... lookup logic
}

// Write operations use Lock for exclusive access
func (r *Registry) Register(b Backend) {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ... registration logic
    delete(r.availabilityCache, b.Name()) // Invalidate cache
}
```text

This design allows:
- Multiple concurrent readers (e.g., health checks, listing)
- Exclusive writers (e.g., registration, unregistration)
- Safe concurrent access from multiple goroutines

### Availability Caching

The registry implements a 30-second TTL cache for availability checks:

```go
type cachedAvailability struct {
    available bool
    checkedAt time.Time
}

func (r *Registry) isAvailableCachedLocked(b Backend) bool {
    name := b.Name()
    if cached, ok := r.availabilityCache[name]; ok &&
       time.Since(cached.checkedAt) < r.availabilityCacheTTL {
        return cached.available
    }

    available := b.IsAvailable()
    r.availabilityCache[name] = &cachedAvailability{
        available: available,
        checkedAt: time.Now(),
    }
    return available
}
```text

**Rationale for 30s TTL**:
- **Performance**: Avoids frequent `exec.LookPath()` calls
- **Freshness**: 30 seconds is short enough to detect installation changes
- **Balance**: Trade-off between accuracy and performance

## Backend Implementations

### Claude Backend

```go
type Claude struct{}

func (c *Claude) Name() string { return "claude" }

func (c *Claude) IsAvailable() bool {
    _, err := exec.LookPath("claude")
    return err == nil
}

func (c *Claude) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--print"}

    if opts != nil {
        if opts.Model != "" {
            args = append(args, "--model", opts.Model)
        }
        // ... additional options
    }

    args = append(args, prompt)
    cmd := exec.Command("claude", args...)

    if opts != nil && opts.WorkDir != "" {
        cmd.Dir = opts.WorkDir
    }
    return cmd
}
```text

### Codex Backend

```go
type Codex struct{}

func (c *Codex) Name() string { return "codex" }

func (c *Codex) IsAvailable() bool {
    _, err := exec.LookPath("codex")
    return err == nil
}

func (c *Codex) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    return exec.Command("codex", args...)
}
```text

### Gemini Backend

```go
type Gemini struct{}

func (g *Gemini) Name() string { return "gemini" }

func (g *Gemini) IsAvailable() bool {
    _, err := exec.LookPath("gemini")
    return err == nil
}

func (g *Gemini) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--output-format", "json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    return exec.Command("gemini", args...)
}
```bash

## Unified Options Handling

The `UnifiedOptions` struct (`internal/backend/unified.go:174-219`) provides a backend-agnostic way to configure AI CLI commands:

```go
type UnifiedOptions struct {
    WorkDir       string
    Model         string
    ApprovalMode  ApprovalMode
    SandboxMode   SandboxMode
    OutputFormat  OutputFormat
    AllowedTools  string
    AllowedDirs   []string
    Interactive   bool
    Verbose       bool
    DryRun        bool
    MaxTokens     int
    MaxTurns      int
    SystemPrompt  string
    ExtraFlags    []string
    Ephemeral     bool
}
```text

### Flag Mapping Architecture

```mermaid
flowchart TB
    subgraph Unified["UnifiedOptions"]
        MODEL[Model]
        APPROVAL[ApprovalMode]
        SANDBOX[SandboxMode]
        OUTPUT[OutputFormat]
    end

    subgraph Mapper["Flag Mapper (internal/backend/unified.go:273-568)"]
        MAP_MODEL[mapModel()]
        MAP_APPROVAL[mapApprovalMode()]
        MAP_SANDBOX[mapSandboxMode()]
        MAP_OUTPUT[mapOutputFormat()]
    end

    subgraph Backends["Backend-Specific Flags"]
        CLAUDE[Claude Flags]
        CODEX[Codex Flags]
        GEMINI[Gemini Flags]
    end

    MODEL --> MAP_MODEL
    APPROVAL --> MAP_APPROVAL
    SANDBOX --> MAP_SANDBOX
    OUTPUT --> MAP_OUTPUT

    MAP_MODEL --> CLAUDE
    MAP_MODEL --> CODEX
    MAP_MODEL --> GEMINI
    MAP_APPROVAL --> CLAUDE
    MAP_APPROVAL --> CODEX
    MAP_APPROVAL --> GEMINI
    MAP_SANDBOX --> CLAUDE
    MAP_SANDBOX --> CODEX
    MAP_SANDBOX --> GEMINI
    MAP_OUTPUT --> CLAUDE
    MAP_OUTPUT --> CODEX
    MAP_OUTPUT --> GEMINI
```text

### Model Name Mapping

Unified model aliases are mapped to backend-specific names:

| Unified Alias | Claude | Codex | Gemini |
|--------------|--------|-------|--------|
| `fast` | `haiku` | `gpt-4.1-mini` | `gemini-2.5-flash` |
| `balanced` | `sonnet` | `gpt-5.2` | `gemini-2.5-pro` |
| `best` | `opus` | `gpt-5-codex` | `gemini-2.5-pro` |

### Approval Mode Mapping

Approval modes control how the backend asks for user confirmation:

```go
func (m *flagMapper) mapApprovalMode(mode ApprovalMode) []string {
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
    case "codex":
        switch mode {
        case ApprovalAuto:
            return []string{"--ask-for-approval", "on-request"}
        case ApprovalNone:
            return []string{"--ask-for-approval", "never"}
        case ApprovalAlways:
            return []string{"--ask-for-approval", "untrusted"}
        }
    // ...
    }
    return nil
}
```text

## Output Parsing and Normalization

Each backend parses its native output into a unified format:

### JSON Response Parsing

```go
func (c *Claude) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
    // First try to parse as error response
    var errResp claudeErrorResponse
    if err := json.Unmarshal([]byte(rawOutput), &errResp); err == nil {
        if errResp.Error != "" {
            return &UnifiedResponse{
                SessionID: errResp.SessionID,
                Error:     errResp.Error,
            }, nil
        }
    }

    var resp claudeJSONResponse
    if err := json.Unmarshal([]byte(rawOutput), &resp); err != nil {
        return nil, err
    }

    return &UnifiedResponse{
        Content:    resp.Result,
        SessionID:  resp.SessionID,
        DurationMs: resp.DurationMs,
        Usage: &TokenUsage{
            InputTokens:  resp.Usage.InputTokens,
            OutputTokens: resp.Usage.OutputTokens,
        },
    }, nil
}
```text

### Unified Response Structure

```go
type UnifiedResponse struct {
    Content    string
    SessionID  string
    Model      string
    DurationMs int64
    Usage      *TokenUsage
    Error      string
    Raw        map[string]any
}
```bash

## Adding New Backends

To add a new AI CLI backend to clinvoker:

### Step 1: Create Implementation File

Create `internal/backend/newbackend.go`:

```go
package backend

import "os/exec"

type NewBackend struct{}

func (n *NewBackend) Name() string {
    return "newbackend"
}

func (n *NewBackend) IsAvailable() bool {
    _, err := exec.LookPath("newbackend-cli")
    return err == nil
}

func (n *NewBackend) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--output", "json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    cmd := exec.Command("newbackend-cli", args...)

    if opts != nil && opts.WorkDir != "" {
        cmd.Dir = opts.WorkDir
    }
    return cmd
}

func (n *NewBackend) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
    args := []string{"--resume", sessionID, "--output", "json"}

    if prompt != "" {
        args = append(args, prompt)
    }

    return exec.Command("newbackend-cli", args...)
}

func (n *NewBackend) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
    return n.BuildCommand(prompt, MapFromUnified(n.Name(), opts))
}

func (n *NewBackend) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
    return n.ResumeCommand(sessionID, prompt, MapFromUnified(n.Name(), opts))
}

func (n *NewBackend) ParseOutput(rawOutput string) string {
    return rawOutput
}

func (n *NewBackend) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
    var resp struct {
        Content   string `json:"content"`
        SessionID string `json:"session_id"`
        Usage     struct {
            Input  int `json:"input_tokens"`
            Output int `json:"output_tokens"`
        } `json:"usage"`
    }

    if err := json.Unmarshal([]byte(rawOutput), &resp); err != nil {
        return nil, err
    }

    return &UnifiedResponse{
        Content:   resp.Content,
        SessionID: resp.SessionID,
        Usage: &TokenUsage{
            InputTokens:  resp.Usage.Input,
            OutputTokens: resp.Usage.Output,
        },
    }, nil
}

func (n *NewBackend) SeparateStderr() bool {
    return false
}
```bash

### Step 2: Register in Registry

Add to `internal/backend/registry.go`:

```go
func init() {
    globalRegistry.Register(&Claude{})
    globalRegistry.Register(&Codex{})
    globalRegistry.Register(&Gemini{})
    globalRegistry.Register(&NewBackend{}) // Add this line
}
```bash

### Step 3: Add Unified Options Mapping

Update `internal/backend/unified.go` to add flag mappings:

```go
func (m *flagMapper) mapModel(model string) string {
    switch m.backend {
    case "newbackend":
        switch model {
        case "fast":
            return "newbackend-fast"
        case "balanced":
            return "newbackend-balanced"
        case "best":
            return "newbackend-pro"
        default:
            return model
        }
    // ...
    }
}
```bash

### Step 4: Add Allowed Flags

Update the allowlist in `internal/backend/unified.go:10-27`:

```go
var allowedFlagPatterns = map[string][]string{
    "newbackend": {
        "--model", "--output", "--verbose",
        "--resume", "--sandbox",
    },
    // ...
}
```text

## Best Practices

### Command Building

- Always validate paths before execution
- Escape arguments properly (Go's `exec.Command` handles this)
- Support both interactive and batch modes
- Use `--print` or equivalent for non-interactive output

### Output Parsing

- Handle partial/invalid JSON gracefully
- Strip control characters and ANSI codes
- Preserve error messages for debugging
- Return structured errors when possible

### Error Handling

- Distinguish between backend errors and system errors
- Provide clear, actionable error messages
- Include troubleshooting hints in error output

## Related Documentation

- [Architecture Overview](architecture.md) - High-level system architecture
- [Session System](session-system.md) - Session persistence mechanisms
- [API Design](api-design.md) - REST API architecture
