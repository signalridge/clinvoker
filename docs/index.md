# clinvk

Unified AI CLI wrapper for orchestrating multiple AI CLI backends with session persistence, parallel task execution, HTTP API server, and unified output formatting.

## Features

<div class="grid cards" markdown>

-   :material-robot-outline:{ .lg .middle } **Multi-Backend Support**

    ---

    Seamlessly switch between Claude Code, Codex CLI, and Gemini CLI

-   :material-cog-outline:{ .lg .middle } **Unified Options**

    ---

    Consistent configuration options work across all backends

-   :material-history:{ .lg .middle } **Session Persistence**

    ---

    Automatic session tracking with resume capability

-   :material-layers-triple:{ .lg .middle } **Parallel Execution**

    ---

    Run multiple AI tasks concurrently with fail-fast support

-   :material-compare:{ .lg .middle } **Backend Comparison**

    ---

    Compare responses from multiple backends side-by-side

-   :material-link-variant:{ .lg .middle } **Chain Execution**

    ---

    Pipeline prompts through multiple backends sequentially

-   :material-api:{ .lg .middle } **HTTP API Server**

    ---

    RESTful API with OpenAI and Anthropic compatible endpoints

-   :material-tune-vertical:{ .lg .middle } **Configuration Cascade**

    ---

    CLI flags → Environment variables → Config file → Defaults

</div>

## Quick Start

```bash
# Run with default backend (Claude Code)
clinvk "fix the bug in auth.go"

# Specify a backend
clinvk --backend codex "implement user registration"

# Resume a session
clinvk resume --last "continue working"

# Compare backends
clinvk compare --all-backends "explain this code"

# Start HTTP API server
clinvk serve --port 8080
```

## Supported Backends

| Backend | CLI Tool | Description |
|---------|----------|-------------|
| Claude Code | `claude` | Anthropic's AI coding assistant |
| Codex CLI | `codex` | OpenAI's code-focused CLI |
| Gemini CLI | `gemini` | Google's Gemini AI CLI |

## Next Steps

- [Installation](getting-started/installation.md) - Install clinvk on your system
- [Quick Start](getting-started/quick-start.md) - Get up and running in minutes
- [User Guide](user-guide/index.md) - Learn about all features
- [HTTP API](server/index.md) - Use the REST API server
