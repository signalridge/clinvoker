# clinvk

<p align="center">
  <strong>Unified AI CLI Orchestrator</strong><br>
  One CLI to rule Claude, Codex, and Gemini
</p>

<p align="center">
  <a href="guide/installation.md">Installation</a> •
  <a href="guide/quick-start.md">Quick Start</a> •
  <a href="guide/use-cases.md">Use Cases</a> •
  <a href="reference/rest-api.md">API Reference</a>
</p>

---

## What is clinvk?

**clinvk** is a unified wrapper that orchestrates multiple AI CLI backends—**Claude Code**, **Codex CLI**, and **Gemini CLI**—into a single, consistent interface.

Think of it as a **universal remote** for AI coding assistants. Instead of learning three different tools, you learn one. Instead of choosing one backend, use them all together.

```
# Use any backend with the same command
clinvk "explain this code"                    # Default backend
clinvk -b codex "optimize this function"      # Switch to Codex
clinvk -b gemini "review for security"        # Switch to Gemini
```

---

## Core Capabilities

### 1. Unified Interface

Same commands, flags, and output format across all backends.

| Feature | clinvk | Claude | Codex | Gemini |
|---------|--------|--------|-------|--------|
| Basic prompt | ✓ | `claude` | `codex exec` | `gemini` |
| Session resume | ✓ | ✓ | ✗ | ✓ |
| JSON output | ✓ | ✓ | ✓ | ✓ |
| Approval modes | ✓ | ✓ | ✓ | ✗ |
| Sandbox control | ✓ | ✓ | ✓ | ✗ |

### 2. Session Management

Persistent conversations with cross-process locking.

```
# Start a session
clinvk "design a database schema for e-commerce"

# Continue later (even from another terminal)
clinvk -c "add user authentication to that schema"

# List and manage sessions
clinvk sessions list
clinvk sessions export <id> -o schema.json
```

### 3. Parallel Execution

Run multiple backends simultaneously for comprehensive analysis.

```
# Three perspectives, one command
clinvk parallel -f security-review.json
```

```
{
  "tasks": [
    {"backend": "claude", "prompt": "Review architecture and design patterns"},
    {"backend": "codex", "prompt": "Check for performance bottlenecks"},
    {"backend": "gemini", "prompt": "Identify security vulnerabilities"}
  ]
}
```

### 4. Chain Execution

Pipeline outputs between backends for complex workflows.

```
# Analyze → Fix → Verify → Document
clinvk chain -f bugfix-pipeline.json
```

```
{
  "steps": [
    {"backend": "claude", "prompt": "Find the root cause of the bug"},
    {"backend": "codex", "prompt": "Fix the issue: {{previous}}"},
    {"backend": "gemini", "prompt": "Write tests for the fix: {{previous}}"}
  ]
}
```

### 5. Backend Comparison

Compare responses side-by-side before making decisions.

```
# Get all opinions on a risky change
clinvk compare --all-backends "Is this database migration safe?"
```

### 6. HTTP API Server

Drop-in replacement for OpenAI/Anthropic APIs.

```
# Start the server
clinvk serve --port 8080
```

```python
# Use with existing OpenAI SDK
from openai import OpenAI
client = OpenAI(base_url="http://localhost:8080/openai/v1", api_key="any")
response = client.chat.completions.create(
    model="claude-opus-4-5-20251101",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

---

## Quick Start

### Install

```
# macOS / Linux
brew install signalridge/tap/clinvk

# Windows
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket
scoop install clinvk

# Go
go install github.com/signalridge/clinvoker/cmd/clinvk@latest
```

### First Steps

```
# 1. Verify installation
clinvk version

# 2. Run your first prompt
clinvk "explain what this codebase does"

# 3. Try a different backend
clinvk -b codex "generate unit tests for auth.go"

# 4. Start the API server
clinvk serve --port 8080
```

---

## Architecture

```
flowchart TB
    subgraph Client["Client Layer"]
        CLI["CLI Commands<br/>prompt, parallel, chain, compare"]
        API["HTTP API<br/>OpenAI/Anthropic Compatible"]
    end

    subgraph Core["Core Layer"]
        SESSION["Session Manager<br/>JSON Store + File Lock"]
        CONFIG["Config Manager<br/>Viper-based"]
        EXEC["Command Executor<br/>PTY Support"]
    end

    subgraph Backends["Backend Layer"]
        CLAUDE["Claude Code"]
        CODEX["Codex CLI"]
        GEMINI["Gemini CLI"]
    end

    CLI --> SESSION
    CLI --> CONFIG
    CLI --> EXEC
    API --> EXEC
    EXEC --> CLAUDE
    EXEC --> CODEX
    EXEC --> GEMINI
```

---

## When to Use clinvk

| Scenario | clinvk Solution |
|----------|----------------|
| **Multi-model code review** | `parallel` with architecture + security + performance tasks |
| **Refactoring pipeline** | `chain` with analyze → fix → verify → document steps |
| **High-risk decisions** | `compare` across all backends before acting |
| **CI/CD automation** | HTTP API with OpenAI-compatible endpoints |
| **A/B testing models** | Same prompt, different backends, consistent output |
| **Vendor flexibility** | Switch backends without changing your workflow |

---

## Documentation Map

### Getting Started
- [Installation](guide/installation.md) - Platform-specific setup
- [Quick Start](guide/quick-start.md) - 5-minute tutorial
- [Basic Usage](guide/basic-usage.md) - Core CLI workflows

### User Guides
- [Configuration](guide/configuration.md) - Hierarchical config system
- [Session Management](guide/session-management.md) - Persistence and lifecycle
- [Parallel Execution](guide/parallel-execution.md) - Multi-backend concurrency
- [Chain Execution](guide/chain-execution.md) - Context-passing pipelines
- [Backend Comparison](guide/backend-comparison.md) - Side-by-side analysis
- [HTTP Server](guide/http-server.md) - API deployment and security

### Use Cases
- [Real-World Scenarios](guide/use-cases.md) - 20+ practical workflows

### Backend-Specific
- [Claude Code](guide/backends/claude.md)
- [Codex CLI](guide/backends/codex.md)
- [Gemini CLI](guide/backends/gemini.md)

### Reference
- [CLI Commands](reference/commands/)
- [REST API](reference/rest-api.md)
- [OpenAI Compatible](reference/openai-compatible.md)
- [Anthropic Compatible](reference/anthropic-compatible.md)
- [Configuration Reference](reference/configuration.md)
- [Environment Variables](reference/environment.md)

### Integration
- [CI/CD Integration](integration/ci-cd.md)
- [LangChain/LangGraph](integration/langchain-langgraph.md)
- [Claude Code Skills](integration/claude-code-skills.md)

### Development
- [Architecture](about/architecture.md)
- [Contributing](development/contributing.md)
- [Design Decisions](about/design-decisions.md)

---

## Philosophy

> **Wrapper, Not Replacement**
>
> clinvk doesn't replace your AI backends—it unifies them. You get all the power of Claude Code, Codex CLI, and Gemini CLI, plus orchestration capabilities they don't provide individually.

---

## License

MIT License - see [LICENSE](https://github.com/signalridge/clinvoker/blob/main/LICENSE) for details.

---

<p align="center">
  Built with ❤️ by <a href="https://github.com/signalridge">signalridge</a>
</p>
