---
title: How-To Guides
description: Practical guides for specific tasks and configurations.
---

# How-To Guides

Practical, task-oriented guides for getting things done with clinvoker.

## Core Usage

<div class="grid cards" markdown>

-   :material-console:{ .lg .middle } __Basic Usage__

    ---

    Learn the essential CLI commands and options for daily use.

    [:octicons-arrow-right-24: Read guide](basic-usage.md)

-   :material-folder-account:{ .lg .middle } __Session Management__

    ---

    Manage persistent sessions, resume conversations, and organize your work.

    [:octicons-arrow-right-24: Read guide](session-management.md)

-   :material-cog:{ .lg .middle } __Configuration__

    ---

    Customize clinvoker with configuration files and environment variables.

    [:octicons-arrow-right-24: Read guide](configuration.md)

</div>

## Advanced Workflows

<div class="grid cards" markdown>

-   :material-play-speed:{ .lg .middle } __Parallel Execution__

    ---

    Run multiple AI tasks concurrently for faster results.

    [:octicons-arrow-right-24: Read guide](parallel-execution.md)

-   :material-vector-polyline:{ .lg .middle } __Chain Execution__

    ---

    Build sequential workflows where each step feeds into the next.

    [:octicons-arrow-right-24: Read guide](chain-execution.md)

-   :material-compare:{ .lg .middle } __Backend Comparison__

    ---

    Compare responses from multiple backends side-by-side.

    [:octicons-arrow-right-24: Read guide](backend-comparison.md)

</div>

## Backend-Specific Guides

<div class="grid cards" markdown>

-   :material-robot:{ .lg .middle } __Claude Code__

    ---

    Best practices and tips for using Claude Code backend.

    [:octicons-arrow-right-24: Read guide](backends/claude.md)

-   :material-code-tags:{ .lg .middle } __Codex CLI__

    ---

    Best practices and tips for using Codex CLI backend.

    [:octicons-arrow-right-24: Read guide](backends/codex.md)

-   :material-google:{ .lg .middle } __Gemini CLI__

    ---

    Best practices and tips for using Gemini CLI backend.

    [:octicons-arrow-right-24: Read guide](backends/gemini.md)

</div>

## Guide Categories

| Category | Guides | Description |
|----------|--------|-------------|
| **Core** | Basic, Sessions, Config | Essential daily usage |
| **Workflows** | Parallel, Chain, Compare | Advanced orchestration |
| **Backends** | Claude, Codex, Gemini | Backend-specific tips |

## Quick Reference

### Common Tasks

```bash
# Run a prompt
clinvk "Your prompt"

# Use specific backend
clinvk -b codex "Generate code"

# Parallel execution
clinvk parallel -f tasks.json

# Chain workflow
clinvk chain -f pipeline.json

# Compare backends
clinvk compare --all-backends "Question"

# List sessions
clinvk sessions list

# Resume session
clinvk resume --last
```

### Configuration Locations

| File | Purpose |
|------|---------|
| `~/.clinvk/config.yaml` | Main configuration |
| `~/.clinvk/sessions/` | Session storage |
| `.clinvk/config.yaml` | Project-specific config |

## Getting Help

- See the [FAQ](../development/faq.md) for common questions
- Check [Troubleshooting](../development/troubleshooting.md) for issues
- Review [Exit Codes](../reference/exit-codes.md) for error details
