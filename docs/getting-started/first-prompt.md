---
title: First Prompt
description: Run your first prompt with clinvoker.
---

# First Prompt

Let's run your first prompt with clinvoker. This guide covers the basics of executing AI prompts.

## Basic Usage

The simplest way to use clinvoker is to provide a prompt as an argument:

```bash
clinvk "Explain the SOLID principles"
```

By default, this uses your configured default backend (usually Claude Code).

## Specifying a Backend

Use the `-b` or `--backend` flag to choose a specific backend:

```bash
# Use Claude Code
clinvk -b claude "Design a database schema for an e-commerce app"

# Use Codex CLI
clinvk -b codex "Implement a quicksort algorithm in Python"

# Use Gemini CLI
clinvk -b gemini "Research the latest trends in AI"
```

## Output Formats

clinvoker supports multiple output formats:

```bash
# Text output (default, human-readable)
clinvk -o text "Explain REST APIs"

# JSON output (structured, includes metadata)
clinvk -o json "Explain REST APIs"

# Streaming JSON (for real-time updates)
clinvk -o stream-json "Explain REST APIs"
```

## Working Directory

Set the working directory for the AI backend:

```bash
# Current directory
clinvk -w . "Review the code in this project"

# Specific directory
clinvk -w /path/to/project "Analyze the architecture"
```

## Continuing Sessions

clinvoker automatically maintains sessions. To continue a previous session:

```bash
# Continue the last session
clinvk --continue "What about caching?"

# Or just resume without a new prompt
clinvk resume
```

## Dry Run

See what command would be executed without running it:

```bash
clinvk --dry-run -b codex "Implement auth middleware"
```

## Ephemeral Mode

Run without persisting the session (useful for CI/CD):

```bash
clinvk --ephemeral "Quick question about Go syntax"
```

## Common Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--backend` | `-b` | Choose backend | `-b claude` |
| `--model` | `-m` | Specify model | `-m claude-sonnet-4` |
| `--workdir` | `-w` | Set working directory | `-w ./project` |
| `--output-format` | `-o` | Output format | `-o json` |
| `--continue` | `-c` | Continue last session | `-c` |
| `--ephemeral` | | Don't persist session | `--ephemeral` |
| `--dry-run` | | Show command only | `--dry-run` |

## Examples

### Code Review

```bash
# Review a specific file
clinvk -b claude "Review auth.go for security issues"

# Review with context
clinvk -w . -b claude "Review the authentication system"
```

### Code Generation

```bash
# Generate code
clinvk -b codex "Generate a Python function to parse JSON"

# Generate with specific requirements
clinvk -b codex "Create a Go HTTP handler with error handling"
```

### Architecture Questions

```bash
# Design discussions
clinvk -b claude "How should I structure a microservices app?"

# Technology choices
clinvk -b claude "Compare PostgreSQL vs MongoDB for my use case"
```

## Troubleshooting

### "Backend not available"

```bash
# Check available backends
clinvk --help | grep backend

# Verify backend is installed
which claude
which codex
which gemini
```

### "No sessions found"

This is normal for your first run. Sessions are created after your first prompt.

### Session conflicts

```bash
# List all sessions
clinvk sessions list

# Clean up old sessions
clinvk sessions cleanup
```

## Next Steps

- Learn about [Session Management](../how-to/session-management.md)
- Try [Parallel Execution](../how-to/parallel-execution.md)
- Explore [Chain Execution](../how-to/chain-execution.md)
