# Design Decisions

This document explains key design decisions in clinvk and the reasoning behind them.

## Why Wrap CLI Tools?

### Decision

clinvk wraps existing AI CLI tools (Claude Code, Codex CLI, Gemini CLI) rather than calling APIs directly.

### Reasoning

1. **Leverage Existing Tooling**: Each AI provider has invested heavily in their CLI tools with features like:
   - Agentic capabilities (file editing, command execution)
   - Context management
   - Safety features and guardrails

2. **Automatic Updates**: When providers update their CLIs with new features or models, clinvk benefits automatically without code changes.

3. **No API Key Management**: The underlying CLIs handle authentication, so clinvk doesn't need to manage API keys.

4. **Full Feature Access**: All CLI features remain accessible, including tool use, file operations, and interactive modes.

### Trade-offs

- Requires CLI tools to be installed
- Dependent on CLI output format stability
- Less control over low-level API parameters

## Why a Unified Interface?

### Decision

Provide the same commands and options regardless of which backend is used.

### Reasoning

1. **Lower Learning Curve**: Learn once, use everywhere
2. **Easy Switching**: Change backends with a single flag
3. **Workflow Portability**: Same scripts work with different backends
4. **Fair Comparison**: Compare backends using identical prompts

### Implementation

```go
type Backend interface {
    Name() string
    Execute(ctx context.Context, opts ExecuteOptions) (*Result, error)
    Available() bool
}
```

All backends implement this interface, enabling polymorphic usage.

## Why Ephemeral Chain Mode?

### Decision

Chain execution always runs in ephemeral mode—no sessions are persisted between steps.

### Reasoning

1. **Predictability**: Each chain run produces the same result given the same inputs
2. **Isolation**: Steps don't accidentally inherit context from previous runs
3. **Simplicity**: No need to track or clean up sessions
4. **Composability**: Chains can be nested or combined without side effects

### Trade-off

Users who want session persistence in chains must manage it explicitly outside clinvk.

## Why OpenAI/Anthropic Compatible APIs?

### Decision

The HTTP server provides endpoints compatible with OpenAI and Anthropic SDKs.

### Reasoning

1. **Ecosystem Compatibility**: Use existing SDKs, tools, and frameworks
2. **Easy Integration**: Drop-in replacement for existing code
3. **LangChain/LangGraph**: Works with popular AI frameworks out of the box
4. **Lower Migration Cost**: Switch from direct API to CLI-backed without rewriting

### Implementation

```
POST /openai/v1/chat/completions
POST /anthropic/v1/messages
```

These endpoints translate SDK requests into CLI executions.

## Why YAML Configuration?

### Decision

Use YAML for configuration files instead of JSON, TOML, or environment-only.

### Reasoning

1. **Human Readable**: Easy to read and edit manually
2. **Comments Supported**: Can document configuration inline
3. **Familiar**: Common in DevOps and cloud-native tools
4. **Hierarchical**: Natural fit for nested configuration

### Example

```yaml
# Default backend for all commands
default_backend: claude

backends:
  claude:
    model: claude-opus-4-5-20251101
    # Include docs in context
    extra_flags:
      - "--add-dir"
      - "./docs"
```

## Why Parallel and Chain as Primitives?

### Decision

Provide `parallel` and `chain` as first-class commands rather than scripting solutions.

### Reasoning

1. **Common Patterns**: These are the two most common multi-backend workflows
2. **Optimized Implementation**: Built-in concurrency control, error handling
3. **Consistent Output**: Structured results regardless of task count
4. **Composability**: Can be combined with shell scripts or used via API

### Parallel vs Chain

| Aspect | Parallel | Chain |
|--------|----------|-------|
| Execution | Concurrent | Sequential |
| Data Flow | Independent | Previous output available |
| Use Case | Multi-perspective | Multi-stage processing |
| Speed | Fast (max task time) | Slow (sum of task times) |

## Why No Built-in Authentication?

### Decision

The HTTP server has no built-in authentication mechanism.

### Reasoning

1. **Local-First**: Primary use case is local development
2. **Diverse Requirements**: Auth needs vary widely (API keys, OAuth, mTLS)
3. **Better Solutions Exist**: Reverse proxies handle auth better
4. **Simplicity**: Keeps the codebase focused on core functionality

### Recommendation

For production use, place clinvk behind a reverse proxy (nginx, Caddy, Traefik) that handles:

- TLS termination
- Authentication
- Rate limiting
- Request logging

## Why Go?

### Decision

Implement clinvk in Go.

### Reasoning

1. **Single Binary**: No runtime dependencies, easy distribution
2. **Cross-Platform**: Compile for Linux, macOS, Windows from one codebase
3. **Performance**: Fast startup, low memory usage
4. **Concurrency**: goroutines perfect for parallel execution
5. **CLI Ecosystem**: Excellent libraries (Cobra, Viper)

## Future Considerations

### Potential Additions

- **Streaming Support**: Real-time output from backends
- **Plugin System**: Custom backends without forking
- **Workflow DSL**: Complex multi-step workflows in declarative format
- **Metrics/Tracing**: Observability for production use

### Guiding Principles

1. **Keep It Simple**: Avoid feature creep
2. **Wrapper First**: Don't duplicate what CLIs already do
3. **Composability**: Build complex from simple
4. **Stability**: Maintain backward compatibility
