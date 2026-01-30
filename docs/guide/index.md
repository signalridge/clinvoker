# User Guide

Complete guides for using clinvk effectively.

---

## Learning Paths

### Quick Start Path (15 minutes)

New to clinvk? Follow this path:

1. [Installation](installation.md) - Get clinvk on your system
2. [Quick Start](quick-start.md) - Run your first commands
3. [Basic Usage](basic-usage.md) - Understand core workflows

### Power User Path (1 hour)

Ready to unlock clinvk's full potential?

1. [Configuration](configuration.md) - Master the config system
2. [Session Management](session-management.md) - Persistent conversations
3. [Parallel Execution](parallel-execution.md) - Multi-backend workflows
4. [Chain Execution](chain-execution.md) - Context-passing pipelines
5. [Use Cases](use-cases.md) - Real-world scenarios

### Integration Path (30 minutes)

Integrating clinvk into your toolchain?

1. [HTTP Server](http-server.md) - API deployment
2. [Backend Comparison](backend-comparison.md) - Choose the right backend
3. [CI/CD Integration](../integration/ci-cd.md) - Automated workflows

---

## Guide Categories

### Getting Started

| Guide | Description | Time |
|-------|-------------|------|
| [Installation](installation.md) | Platform-specific installation | 5 min |
| [Quick Start](quick-start.md) | First commands and concepts | 10 min |
| [Basic Usage](basic-usage.md) | Core CLI patterns | 15 min |

### Core Concepts

| Guide | Description | Time |
|-------|-------------|------|
| [Configuration](configuration.md) | Hierarchical config system | 15 min |
| [Session Management](session-management.md) | Persistence and lifecycle | 10 min |
| [Backend Comparison](backend-comparison.md) | Comparing responses | 10 min |

### Advanced Workflows

| Guide | Description | Time |
|-------|-------------|------|
| [Parallel Execution](parallel-execution.md) | Concurrent multi-backend tasks | 15 min |
| [Chain Execution](chain-execution.md) | Sequential pipelines | 15 min |
| [HTTP Server](http-server.md) | OpenAI-compatible API | 20 min |

### Backend-Specific

| Guide | Description |
|-------|-------------|
| [Claude Code](backends/claude.md) | Claude-specific features |
| [Codex CLI](backends/codex.md) | Codex-specific features |
| [Gemini CLI](backends/gemini.md) | Gemini-specific features |

---

## Common Tasks

### Running Prompts

```
# Basic prompt
clinvk "explain this code"

# With specific backend
clinvk -b codex "optimize this function"

# With output format
clinvk -o json "generate test cases"
```

### Managing Sessions

```
# Continue last session
clinvk -c "add error handling"

# List sessions
clinvk sessions list

# Resume specific session
clinvk resume abc123
```

### Multi-Backend Workflows

```
# Parallel execution
clinvk parallel -f tasks.json

# Chain execution
clinvk chain -f pipeline.json

# Compare backends
clinvk compare --all-backends "review this code"
```

---

## Tips

- Use `--dry-run` to preview commands without executing
- Set `default_backend` in config to avoid typing `-b` every time
- Use `--ephemeral` for one-off tasks that don't need sessions
- Chain commands with `&&` for sequential workflows

---

## Getting Help

- [FAQ](../development/faq.md) - Frequently asked questions
- [Troubleshooting](../development/troubleshooting.md) - Common issues and solutions
- [Reference](../reference/) - Complete command and API reference
