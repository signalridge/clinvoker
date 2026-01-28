# Architecture Overview

This document describes the architecture of clinvk, a unified AI CLI wrapper for orchestrating multiple AI backends.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLI Interface                               │
│                           (cmd/clinvk/main.go)                          │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Application Layer                              │
│                            (internal/app/)                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │  prompt  │ │ parallel │ │  chain   │ │ compare  │ │  serve   │      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘      │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                    ┌────────────────┼────────────────┐
                    ▼                ▼                ▼
┌──────────────────────┐ ┌──────────────────┐ ┌──────────────────────────┐
│   Executor Layer     │ │   HTTP Server    │ │    Session Manager       │
│ (internal/executor/) │ │ (internal/server)│ │   (internal/session/)    │
└──────────────────────┘ └──────────────────┘ └──────────────────────────┘
                    │                │
                    ▼                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            Backend Layer                                 │
│                          (internal/backend/)                             │
│  ┌────────────────┐ ┌────────────────┐ ┌────────────────┐              │
│  │  Claude Code   │ │   Codex CLI    │ │   Gemini CLI   │              │
│  └────────────────┘ └────────────────┘ └────────────────┘              │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           External Binaries                              │
│                    (claude, codex, gemini - in PATH)                    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Module Overview

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
| `helpers.go` | Shared helpers |

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
- `unified.go` - Unified interface utilities

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

Config location: `~/.clinvk/config.yaml`

### Output Parsing (`internal/output/`)

Parses and normalizes output from different backends:

- Event-based streaming parsing
- JSON output normalization
- Token usage extraction

## Data Flow

### Single Prompt Execution

```
User Input → CLI Parser → App Layer → Executor → Backend → External Binary
                                          ↓
                                    Parse Output
                                          ↓
                                    Session Store
                                          ↓
                                    Output to User
```

### Parallel Execution

```
Task List → Worker Pool → [Backend 1] ──┐
                        → [Backend 2] ──┼─→ Aggregate Results → Output
                        → [Backend 3] ──┘
```

### Chain Execution

```
Step 1 → Backend A → Output 1 → {{previous}} substitution
                                      ↓
Step 2 → Backend B → Output 2 → {{previous}} substitution
                                      ↓
Step 3 → Backend C → Final Output
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
- Anthropic-compatible for Claude-specific clients

### 5. Streaming Output

Real-time output streaming via:
- Subprocess stdout/stderr pipes
- Chunk-based parsing utilities in `internal/output/`

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

## Testing Strategy

### Unit Tests

- Co-located with source files
- Mock external dependencies
- Table-driven tests

### Integration Tests

- CLI command execution
- HTTP API endpoints
- Cross-platform compatibility

### Test Helpers

Located in `internal/testutil/`:
- Mock backends
- Temporary directories
- Test server utilities

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

## Future Directions

### Planned Features

- Plugin system for custom backends
- WebSocket streaming support
- Session export/import

### Technical Improvements

- Structured logging with slog
- Metrics and tracing
- Configuration validation

## Directory Structure

```
clinvoker/
├── cmd/
│   └── clinvk/           # Entry point
├── internal/
│   ├── app/              # CLI commands
│   ├── backend/          # Backend implementations
│   ├── config/           # Configuration
│   ├── executor/         # Execution logic
│   ├── output/           # Output parsing
│   ├── server/           # HTTP server
│   │   ├── handlers/     # API handlers
│   │   └── service/      # Business logic
│   ├── session/          # Session management
│   └── testutil/         # Test utilities
├── docs/                 # Documentation
└── testdata/             # Test fixtures
```

## Related Documentation

- [CLI Reference](CLI.md)
- [Configuration Guide](CONFIGURATION.md)
- [Contributing Guide](../CONTRIBUTING.md)
