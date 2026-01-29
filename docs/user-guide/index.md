# User Guide

This guide covers all features of clinvk in detail.

## Overview

clinvk provides a unified interface to multiple AI coding assistants, with powerful orchestration features for complex workflows.

## Core Features

- **[Basic Usage](basic-usage.md)** - Learn the fundamentals of running prompts and using backends
- **[Session Management](session-management.md)** - Track and resume conversations across sessions
- **[Parallel Execution](parallel-execution.md)** - Run multiple tasks concurrently for faster workflows
- **[Chain Execution](chain-execution.md)** - Pipeline prompts through multiple backends sequentially
- **[Backend Comparison](backend-comparison.md)** - Compare responses from different AI backends

## Backend Guides

Learn about each supported backend:

- **[Claude Code](backends/claude.md)** - Anthropic's AI coding assistant
- **[Codex CLI](backends/codex.md)** - OpenAI's code-focused CLI tool
- **[Gemini CLI](backends/gemini.md)** - Google's Gemini AI assistant

## Workflow Examples

### Solo Development

```bash
# Start working on a feature
clinvk "implement user authentication"

# Continue the conversation
clinvk -c "add password hashing"

# Get a different perspective
clinvk -b gemini "review the implementation"
```

### Code Review

```bash
# Get reviews from multiple backends
clinvk compare --all-backends "review this PR for issues"
```

### Complex Tasks

```bash
# Run multiple independent tasks
clinvk parallel --file tasks.json

# Chain multiple perspectives
clinvk chain --file review-pipeline.json
```

## Tips

!!! tip "Use Session Continuity"
    Always use `--continue` or `clinvk resume` to maintain context in longer conversations.

!!! tip "Backend Selection"
    Different backends excel at different tasks. Claude is great for complex reasoning, Codex for code generation, and Gemini for broad knowledge.

!!! tip "Dry Run First"
    Use `--dry-run` to see what command would be executed without actually running it.
