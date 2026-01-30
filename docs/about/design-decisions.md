# Design Decisions

This document explains the key design decisions made during clinvk development and the rationale behind them.

## Why Subprocess Execution Instead of SDK?

### The Decision

clinvk executes AI CLI tools as subprocesses rather than using their SDKs directly.

### Rationale

1. **Zero Configuration**: CLI tools handle authentication, API keys, and configuration automatically
2. **Always Up-to-Date**: No need to update clinvk when SDK APIs change
3. **Feature Parity**: CLI tools often have features not available in SDKs
4. **Session Management**: Leverage built-in session handling of CLI tools
5. **Simplicity**: One abstraction layer instead of multiple SDK integrations

### Trade-offs

| Aspect | Subprocess Approach | SDK Approach |
|--------|---------------------|--------------|
| Startup Time | Slightly slower | Faster |
| Dependencies | Fewer | More libraries |
| Maintenance | Lower | Higher |
| Feature Access | Full CLI features | SDK-limited |
| Authentication | CLI handles it | Code required |

## Why Multiple API Formats?

### The Decision

clinvk provides three API endpoint families: OpenAI-compatible, Anthropic-compatible, and custom REST.

### Rationale

1. **OpenAI-Compatible** (`/openai/v1/*`)
   - Most AI frameworks use OpenAI SDK format
   - LangChain, LangGraph work out of the box
   - Easy migration from OpenAI to other backends

2. **Anthropic-Compatible** (`/anthropic/v1/*`)
   - Native support for Anthropic SDK users
   - Full Claude-specific features
   - Messages API format

3. **Custom REST** (`/api/v1/*`)
   - Simpler, more direct interface
   - clinvk-specific features (parallel, chain)
   - Optimal for Claude Code Skills integration

## Session Management Design

### The Decision

Sessions are tracked per-backend with configurable persistence modes.

### Options Considered

1. **Global Sessions** - Single session across all backends
2. **Per-Backend Sessions** - Each backend has independent sessions
3. **Hybrid** - Shared context with backend-specific state

### Why Per-Backend?

- Different backends have incompatible session formats
- Avoids context pollution between backends
- Simpler mental model for users
- Matches CLI tools' native behavior

### Session Modes

| Mode | Persistence | Use Case |
|------|-------------|----------|
| `ephemeral` | None | Stateless API calls |
| `auto` | Auto-named | Default interactive use |
| `named` | User-specified | Long-running projects |

## Parallel vs Chain Execution

### The Decision

Provide both parallel (concurrent) and chain (sequential) execution modes.

### Design Principles

**Parallel Execution:**

```text
Task A ─┬─→ Backend 1 ──┬─→ Result A
        ├─→ Backend 2 ──┤   Result B
        └─→ Backend 3 ──┘   Result C
```yaml

- Independent tasks run concurrently
- Fail-fast option for efficiency
- Results aggregated at completion

**Chain Execution:**

```text
Input → Backend 1 → {{previous}} → Backend 2 → {{previous}} → Backend 3 → Output
```

- Sequential pipeline
- Each step accesses previous output via `{{previous}}`
- Enables multi-stage workflows

## Configuration Cascade

### The Decision

Configuration follows a cascade: CLI flags → Environment → Config file → Defaults.

### Rationale

1. **Predictable Override** - Higher priority sources always win
2. **Environment Friendly** - Works well in containers and CI/CD
3. **User-Controllable** - Easy to override without changing files
4. **Secure Defaults** - Safe configuration when nothing specified

### Example Resolution

```yaml
# Config file: ~/.clinvk/config.yaml
backend: claude
timeout: 60

# Environment
CLINVK_TIMEOUT=120

# CLI
clinvk --backend codex "prompt"

# Result: backend=codex (CLI), timeout=120 (env)
```text

## HTTP Server Design

### The Decision

Single binary serves all endpoints with graceful shutdown.

### Key Features

1. **Standard HTTP/1.1** - Maximum compatibility
2. **SSE for Streaming** - Server-Sent Events for real-time output
3. **CORS Configurable** - For browser-based clients
4. **Health Endpoint** - `/health` for load balancers

### Why Not gRPC?

- HTTP is universally supported
- Browser compatibility important
- Simpler debugging with curl
- Most AI SDKs use HTTP/REST

## Error Handling Philosophy

### The Decision

Propagate errors with context, fail gracefully.

### Principles

1. **Preserve CLI Exit Codes** - Backend errors propagated accurately
2. **Structured Errors** - JSON format with error details
3. **Graceful Degradation** - Partial results in parallel mode
4. **Detailed Logging** - Debug information when needed

### Error Response Format

```json
{
  "error": {
    "type": "backend_error",
    "message": "Claude CLI exited with code 1",
    "backend": "claude",
    "details": "rate limit exceeded"
  }
}
```

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

## Summary

| Decision | Choice | Key Reason |
|----------|--------|------------|
| Execution | Subprocess | Zero configuration, always up-to-date |
| API Format | Multiple | Framework compatibility |
| Sessions | Per-backend | Isolation and simplicity |
| Orchestration | Parallel + Chain | Different workflow needs |
| Config | Cascade | Predictable, environment-friendly |
| Server | HTTP/SSE | Universal compatibility |
