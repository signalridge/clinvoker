---
title: Design Decisions
description: Rationale behind key architectural choices in clinvoker.
---

# Design Decisions

This document explains the key architectural decisions made during clinvoker development and the rationale behind them. Understanding these decisions helps developers contribute effectively and users understand the system's behavior.

## Why Go Was Chosen

### The Decision

clinvoker is implemented in Go (Golang).

### Rationale

1. **Single Binary Deployment**: Go compiles to a single static binary with no runtime dependencies, making distribution trivial.

2. **Excellent Concurrency**: Go's goroutines and channels provide lightweight concurrency primitives perfect for handling multiple backends and requests.

3. **Standard Library**: Rich standard library includes HTTP servers, JSON handling, and subprocess management without external dependencies.

4. **Cross-Platform**: Native support for Windows, macOS, and Linux with minimal platform-specific code.

5. **Fast Compilation**: Quick build times improve developer productivity.

### Alternatives Considered

| Language | Pros | Cons |
|----------|------|------|
| Python | Ecosystem, AI libraries | Deployment complexity, GIL limitations |
| Rust | Performance, Safety | Steeper learning curve, longer compile times |
| Node.js | JavaScript familiarity | Runtime dependency, callback complexity |
| Java | Mature ecosystem | JVM requirement, verbose |

## Why Cobra for CLI

### The Decision

Cobra is used as the CLI framework.

### Rationale

1. **Industry Standard**: Used by Kubernetes, Hugo, and many other major Go projects.

2. **Rich Features**: Built-in help generation, shell completion, and flag parsing.

3. **Command Hierarchy**: Natural support for subcommands with persistent and local flags.

4. **Documentation**: Auto-generated documentation from code.

5. **Validation**: Built-in flag validation and error handling.

### Implementation Pattern

```go
var rootCmd = &cobra.Command{
    Use:   "clinvk",
    Short: "Unified AI CLI wrapper",
    Long:  `A unified interface for Claude Code, Codex CLI, and Gemini CLI.`,
    RunE:  runRoot,
}

func init() {
    rootCmd.PersistentFlags().String("backend", "", "AI backend to use")
    rootCmd.PersistentFlags().String("model", "", "Model to use")
    // ...
}
```text

## Why Chi for HTTP Router

### The Decision

Chi is used as the HTTP router and middleware framework.

### Rationale

1. **Lightweight**: Minimal overhead, idiomatic Go design.

2. **Middleware Chain**: Elegant middleware composition with `Use()` pattern.

3. **Context-Aware**: Built on `context.Context` for request-scoped values.

4. **URL Parameters**: Clean URL parameter extraction.

5. **Compatibility**: Works seamlessly with standard `http.Handler`.

### Middleware Stack

```go
router := chi.NewRouter()
router.Use(middleware.RequestID)
router.Use(middleware.RealIP)
router.Use(middleware.Recoverer)
router.Use(middleware.Logger)
router.Use(middleware.Timeout(60 * time.Second))
```text

## Why Huma for OpenAPI

### The Decision

Huma is used for OpenAPI generation and request/response validation.

### Rationale

1. **Code-First**: Generate OpenAPI spec from Go code, not vice versa.

2. **Type Safety**: Request/response types validated at compile time.

3. **Automatic Documentation**: Interactive docs generated from code.

4. **Validation**: Automatic request validation based on struct tags.

5. **Multiple Adapters**: Works with Chi, Gin, and other routers.

### Example Usage

```go
huma.Register(api, huma.Operation{
    OperationID: "create-chat-completion",
    Method:      http.MethodPost,
    Path:        "/openai/v1/chat/completions",
}, func(ctx context.Context, input *ChatRequest) (*ChatResponse, error) {
    // Handler implementation
})
```text

## Why Subprocess Execution Instead of SDK

### The Decision

clinvoker executes AI CLI tools as subprocesses rather than using their SDKs directly.

### Rationale

1. **Zero Configuration**: CLI tools handle authentication, API keys, and configuration automatically.

2. **Always Up-to-Date**: No need to update clinvoker when SDK APIs change.

3. **Feature Parity**: CLI tools often have features not available in SDKs.

4. **Session Management**: Leverage built-in session handling of CLI tools.

5. **Simplicity**: One abstraction layer instead of multiple SDK integrations.

### Trade-offs

| Aspect | Subprocess Approach | SDK Approach |
|--------|---------------------|--------------|
| Startup Time | Slightly slower | Faster |
| Dependencies | Fewer | More libraries |
| Maintenance | Lower | Higher |
| Feature Access | Full CLI features | SDK-limited |
| Authentication | CLI handles it | Code required |

## SDK Compatibility Approach

### The Decision

Provide OpenAI and Anthropic-compatible API endpoints alongside native REST API.

### Rationale

1. **Ecosystem Compatibility**: Existing tools using OpenAI SDK work without modification.

2. **Migration Path**: Easy transition from cloud APIs to local CLI tools.

3. **Framework Support**: LangChain, LangGraph, and similar frameworks work out of the box.

4. **Familiar Interface**: Developers already know these APIs.

### Implementation Strategy

```mermaid
flowchart TB
    subgraph Input["Client Request"]
        OPENAI[OpenAI Format]
        ANTH[Anthropic Format]
        NATIVE[Native Format]
    end

    subgraph Transform["Transformation Layer"]
        MAP[Unified Options Mapper]
    end

    subgraph Internal["Internal Processing"]
        EXEC[Executor]
        BACKEND[Backend]
    end

    OPENAI --> MAP
    ANTH --> MAP
    NATIVE --> MAP
    MAP --> EXEC
    EXEC --> BACKEND
```typescript

## Session Persistence Trade-offs

### The Decision

Sessions are persisted to local filesystem with JSON format.

### Rationale

1. **Simplicity**: No external database required.

2. **Portability**: Easy to backup, migrate, and inspect.

3. **Human-Readable**: JSON format allows manual inspection and debugging.

4. **Version Control**: Sessions can be version controlled if desired.

### File-based vs Database Storage

| Aspect | File-based | Database |
|--------|------------|----------|
| Setup | None required | Installation required |
| Complexity | Low | Higher |
| Querying | Limited | Rich |
| Concurrency | File locking | ACID transactions |
| Scalability | Single machine | Distributed |
| Backup | File copy | Database backup |

### Why Not SQLite?

SQLite was considered but rejected because:
- JSON files are easier to inspect and debug
- No schema migration complexity
- Simpler backup and restore
- Cross-process access is straightforward with file locking

## Backend Abstraction Design Choices

### The Decision

Use a common `Backend` interface with unified options mapping.

### Rationale

1. **Polymorphism**: Treat all backends uniformly in core code.

2. **Extensibility**: Easy to add new backends without modifying core.

3. **Testability**: Mock backends for testing.

4. **Consistency**: Same API regardless of backend.

### Interface Design

```go
type Backend interface {
    Name() string
    IsAvailable() bool
    BuildCommand(prompt string, opts *Options) *exec.Cmd
    ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd
    ParseOutput(rawOutput string) string
    ParseJSONResponse(rawOutput string) (*UnifiedResponse, error)
}
```text

## Concurrency Model Selection

### The Decision

Use `sync.RWMutex` for in-process concurrency and file locking for cross-process synchronization.

### Rationale

1. **Read-Heavy Workload**: Most operations are reads (listing, getting sessions).

2. **Go Idiomatic**: Standard Go pattern for concurrent access.

3. **Cross-Process Safety**: File locks enable CLI and server coexistence.

4. **Simplicity**: Easier to reason about than channel-based approaches.

### Concurrency Patterns

```go
// Read operation
func (s *Store) Get(id string) (*Session, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.getLocked(id)
}

// Write operation
func (s *Store) Save(sess *Session) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.saveLocked(sess)
}
```text

## Configuration Cascade Design

### The Decision

Configuration follows a cascade: CLI flags -> Environment -> Config file -> Defaults.

### Rationale

1. **Predictable Override**: Higher priority sources always win.

2. **Environment Friendly**: Works well in containers and CI/CD.

3. **User-Controllable**: Easy to override without changing files.

4. **Secure Defaults**: Safe configuration when nothing specified.

### Resolution Example

```yaml
# Config file: ~/.clinvk/config.yaml
backend: claude
timeout: 60

# Environment
CLINVK_TIMEOUT=120

# CLI
clinvk --backend codex "prompt"

# Result: backend=codex (CLI), timeout=120 (env)
```bash

## HTTP Server Design

### The Decision

Single binary serves all endpoints with graceful shutdown.

### Key Features

1. **Standard HTTP/1.1**: Maximum compatibility.

2. **SSE for Streaming**: Server-Sent Events for real-time output.

3. **CORS Configurable**: For browser-based clients.

4. **Health Endpoint**: `/health` for load balancers.

### Why Not gRPC?

- HTTP is universally supported
- Browser compatibility important
- Simpler debugging with curl
- Most AI SDKs use HTTP/REST

## Error Handling Philosophy

### The Decision

Propagate errors with context, fail gracefully.

### Principles

1. **Preserve CLI Exit Codes**: Backend errors propagated accurately.

2. **Structured Errors**: JSON format with error details.

3. **Graceful Degradation**: Partial results in parallel mode.

4. **Detailed Logging**: Debug information when needed.

### Error Response Format

```json
{
  "error": {
    "code": "backend_error",
    "message": "Claude CLI exited with code 1",
    "backend": "claude",
    "details": "rate limit exceeded"
  }
}
```text

## Summary Table

| Decision | Choice | Key Reason |
|----------|--------|------------|
| Language | Go | Single binary, excellent concurrency |
| CLI Framework | Cobra | Industry standard, rich features |
| HTTP Router | Chi | Lightweight, idiomatic Go |
| OpenAPI | Huma | Code-first, type-safe |
| Execution | Subprocess | Zero configuration, always up-to-date |
| API Format | Multiple | Framework compatibility |
| Sessions | File-based JSON | Simplicity, portability |
| Concurrency | RWMutex + FileLock | Read-heavy, cross-process safe |
| Config | Cascade | Predictable, environment-friendly |
| Server | HTTP/SSE | Universal compatibility |

## Future Considerations

### MCP Server Support

We're evaluating adding Model Context Protocol (MCP) server support to enable:

- Direct integration with Claude Desktop
- Standardized tool calling interface
- Ecosystem compatibility

### Additional Backends

The backend abstraction allows adding new AI CLIs as they become available. Requirements for new backends:

- CLI supports non-interactive mode
- Structured output (JSON preferred)
- Session management (optional but preferred)

## Related Documentation

- [Architecture Overview](architecture.md) - System architecture
- [Backend System](backend-system.md) - Backend abstraction details
- [Session System](session-system.md) - Session persistence design
