# Architecture Overview

This document describes the architecture of clinvk.

## System Architecture

```mermaid
flowchart TD
    CLI["CLI (cmd/clinvk)"] --> App["App layer (internal/app)"]
    App --> Executor["Executor (internal/executor)"]
    App --> Server["HTTP server (internal/server)"]
    App --> Session["Session store (internal/session)"]

    Executor --> Claude["Claude backend"]
    Executor --> Codex["Codex backend"]
    Executor --> Gemini["Gemini backend"]

    Server --> Handlers["HTTP handlers"]
    Handlers --> Service["Service layer"]
    Service --> Executor

    Claude --> ExtClaude["claude binary"]
    Codex --> ExtCodex["codex binary"]
    Gemini --> ExtGemini["gemini binary"]

    style CLI fill:#e3f2fd,stroke:#1976d2
    style App fill:#fff3e0,stroke:#f57c00
    style Executor fill:#ffecb3,stroke:#ffa000
    style Server fill:#e8f5e9,stroke:#388e3c
    style Session fill:#f3e5f5,stroke:#7b1fa2
    style Claude fill:#f3e5f5,stroke:#7b1fa2
    style Codex fill:#e8f5e9,stroke:#388e3c
    style Gemini fill:#ffebee,stroke:#c62828
```

## Layer Overview

### Entry Point (`cmd/clinvk/`)

The main entry point initializes the CLI application and delegates to the app layer.

### Application Layer (`internal/app/`)

Implements CLI commands and orchestrates other modules:

| File | Purpose |
|------|---------|
| `app.go` | Root command, global flags, prompt execution |
| `cmd_parallel.go` | Concurrent multi-task execution |
| `cmd_chain.go` | Sequential pipeline execution |
| `cmd_compare.go` | Multi-backend comparison |
| `cmd_serve.go` | HTTP server startup |
| `cmd_sessions.go` | Session management commands |
| `cmd_config.go` | Configuration commands |
| `cmd_version.go` | Version command |

### Backend Layer (`internal/backend/`)

Provides a unified interface for different AI CLI tools:

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

Implementations:

- `claude.go` - Claude Code backend
- `codex.go` - Codex CLI backend
- `gemini.go` - Gemini CLI backend

### Executor Layer (`internal/executor/`)

Handles actual execution of backend commands:

| File | Purpose |
|------|---------|
| `executor.go` | Command execution with PTY support |
| `signal.go` | Signal forwarding |
| `signal_unix.go` | Unix signal handling |
| `signal_windows.go` | Windows signal handling |

### Server Layer (`internal/server/`)

HTTP API server with multiple API styles:

```
/api/v1/          - Custom RESTful API
/openai/v1/       - OpenAI-compatible API
/anthropic/v1/    - Anthropic-compatible API
```

Components:

- `server.go` - Server setup and routing
- `handlers/` - Request handlers
- `service/` - Business logic & orchestration
- `core/` - Backend execution core (stateless)

### Session Layer (`internal/session/`)

Manages persistent sessions:

```go
type Session struct {
    ID        string
    Backend   string
    Model     string
    Workdir   string
    CreatedAt time.Time
    UpdatedAt time.Time
    Metadata  map[string]any
}
```

Storage: JSON files in `~/.clinvk/sessions/`

### Configuration (`internal/config/`)

Handles configuration loading with cascade:

```
CLI Flags > Environment Variables > Config File > Defaults
```

## Data Flow

### Single Prompt Execution

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant CLI
    participant App
    participant Exec as Executor
    participant Backend
    participant Store as Session store

    User->>CLI: clinvk "prompt"
    CLI->>App: parse args
    App->>Backend: build command
    App->>Exec: execute
    Exec->>Backend: run external binary
    Backend-->>Exec: output
    Exec->>App: parse output
    App->>Store: persist session (unless ephemeral)
    App-->>User: display result
```

### Parallel Execution

```mermaid
flowchart LR
    Tasks["Task list"] --> Pool["Worker pool"]
    Pool --> B1["Backend 1"]
    Pool --> B2["Backend 2"]
    Pool --> B3["Backend 3"]
    B1 --> Agg["Aggregate results"]
    B2 --> Agg
    B3 --> Agg
    Agg --> Output["Output"]

    style Tasks fill:#e3f2fd,stroke:#1976d2
    style Pool fill:#fff3e0,stroke:#f57c00
    style B1 fill:#f3e5f5,stroke:#7b1fa2
    style B2 fill:#e8f5e9,stroke:#388e3c
    style B3 fill:#ffebee,stroke:#c62828
    style Agg fill:#ffecb3,stroke:#ffa000
    style Output fill:#c8e6c9,stroke:#2e7d32
```

### Chain Execution

```mermaid
flowchart LR
    S1["Step 1"] --> B1["Backend A"]
    B1 --> O1["Output 1"]
    O1 --> S2["Step 2"]
    S2 --> B2["Backend B"]
    B2 --> O2["Output 2"]
    O2 --> S3["Step 3"]
    S3 --> B3["Backend C"]
    B3 --> Final["Final output"]

    style S1 fill:#e3f2fd,stroke:#1976d2
    style S2 fill:#e3f2fd,stroke:#1976d2
    style S3 fill:#e3f2fd,stroke:#1976d2
    style B1 fill:#f3e5f5,stroke:#7b1fa2
    style B2 fill:#e8f5e9,stroke:#388e3c
    style B3 fill:#ffebee,stroke:#c62828
    style O1 fill:#fff3e0,stroke:#f57c00
    style O2 fill:#fff3e0,stroke:#f57c00
    style Final fill:#c8e6c9,stroke:#2e7d32
```

## Key Design Decisions

### 1. Backend Abstraction

All backends implement a common interface, enabling:

- Easy addition of new backends
- Consistent behavior across backends
- Backend-agnostic orchestration

### 2. Configuration Cascade

Priority: CLI flags > env vars > config file > defaults

This allows:

- Quick overrides via CLI
- Environment-specific configuration
- Persistent preferences in config file

### 3. Session Persistence

Sessions stored as JSON files for:

- Resumability across invocations
- Easy debugging and inspection
- No database dependency

### 4. HTTP API Compatibility

Multiple API styles for integration:

- Custom API for full functionality
- OpenAI-compatible for existing tooling
- Anthropic-compatible for Claude clients

### 5. Streaming Output

Real-time output streaming via:

- Subprocess stdout/stderr pipes
- Chunk-based parsing utilities

## Security Considerations

### Subprocess Execution

- Commands are built programmatically, not shell-interpreted
- Working directory is validated
- Timeouts prevent runaway processes

### Configuration

- Config file uses restrictive permissions
- No sensitive data stored in sessions
- API keys handled by underlying CLI tools

### HTTP Server

- Bind to localhost by default
- No authentication (intended for local use)
- Request validation via huma/v2

## Performance Considerations

### Parallel Execution

- Configurable worker pool size
- Fail-fast option for early termination
- Memory-efficient result aggregation

### Session Store

- Indexed lookups for common queries
- Pagination for large session lists
- Lazy loading of session content

### Output Parsing

- Streaming parse for large outputs
- Efficient JSON unmarshaling
- Compiled regex patterns
