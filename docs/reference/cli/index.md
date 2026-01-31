# Commands Reference

Complete reference for all clinvk CLI commands.

## Synopsis

```bash
clinvk [global-flags] [prompt]
clinvk [command] [subcommand] [flags]
```

## Command Overview

| Command | Description | Common Use |
|---------|-------------|------------|
| [`[prompt]`](prompt.md) | Execute a prompt (default command) | Daily AI tasks |
| [`resume`](resume.md) | Resume a previous session | Continue conversations |
| [`sessions`](sessions.md) | Manage sessions | List, show, delete sessions |
| [`config`](config.md) | Manage configuration | View or change settings |
| [`parallel`](parallel.md) | Execute tasks in parallel | Run multiple tasks |
| [`compare`](compare.md) | Compare backend responses | Evaluate different AIs |
| [`chain`](chain.md) | Execute prompt chain | Multi-step workflows |
| [`serve`](serve.md) | Start HTTP API server | Application integration |
| `version` | Show version information | Check installed version |
| `help` | Show help | Get command help |

## Global Flags

These flags work with all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | AI backend to use |
| `--model` | `-m` | string | | Model to use |
| `--workdir` | `-w` | string | | Working directory |
| `--output-format` | `-o` | string | `json` | Output format |
| `--config` | | string | | Config file path |
| `--dry-run` | | bool | `false` | Show command only |
| `--ephemeral` | | bool | `false` | Stateless mode |
| `--help` | `-h` | | | Show help |

### Prompt-Specific Flags

These flags work only with the default prompt command:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--continue` | `-c` | bool | `false` | Continue the most recent session |

## Flag Details

### --backend, -b

Select the AI backend to use:

```bash
clinvk --backend claude "prompt"
clinvk -b codex "prompt"
clinvk -b gemini "prompt"
```

Available backends: `claude`, `codex`, `gemini`

### --model, -m

Override the default model for the selected backend:

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "prompt"
clinvk -b codex -m o3-mini "prompt"
```

### --workdir, -w

Set the working directory for the AI backend:

```bash
clinvk --workdir /path/to/project "analyze this codebase"
```

### --output-format, -o

Control output format:

| Value | Description |
|-------|-------------|
| `text` | Plain text |
| `json` | Structured JSON (default) |
| `stream-json` | Streaming JSON events |

```bash
clinvk --output-format json "prompt"
clinvk -o stream-json "prompt"
```

### --config

Use a custom configuration file:

```bash
clinvk --config /path/to/config.yaml "prompt"
```

### --dry-run

Show the command that would be executed without running it:

```bash
clinvk --dry-run "implement feature X"
# Output: Would execute: claude --model claude-opus-4-5-20251101 "implement feature X"
```

### --ephemeral

Run in stateless mode without creating a session:

```bash
clinvk --ephemeral "quick question"
```

## Command Categories

### Core Commands

Commands for everyday use:

- `[prompt]` - Execute prompts
- `resume` - Continue sessions
- `sessions` - Manage sessions

### Configuration Commands

Commands for managing settings:

- `config` - View and modify configuration

### Execution Commands

Commands for advanced execution patterns:

- `parallel` - Run multiple tasks concurrently
- `chain` - Execute sequential pipelines
- `compare` - Compare multiple backends

### Server Commands

Commands for running the HTTP API:

- `serve` - Start the API server

## Usage Examples

### Basic Usage

```bash
# Execute a prompt
clinvk "fix the bug in auth.go"

# Specify backend
clinvk -b codex "implement feature"

# Specify model
clinvk -b claude -m claude-sonnet-4-20250514 "quick review"
```

### Session Management

```bash
# List sessions
clinvk sessions list

# Show session details
clinvk sessions show abc123

# Resume last session
clinvk resume --last

# Delete old sessions
clinvk sessions clean --older-than 30d
```

### Configuration

```bash
# Show current config
clinvk config show

# Set a value
clinvk config set default_backend codex
```

### Advanced Execution

```bash
# Run tasks in parallel
clinvk parallel --file tasks.json

# Compare backends
clinvk compare --all-backends "explain this code"

# Execute a chain
clinvk chain --file pipeline.json
```

### Server

```bash
# Start server
clinvk serve --port 8080
```

## Exit Codes

All commands return exit codes:

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Backend not available |
| 3 | Invalid configuration |
| 4 | Session error |

See [Exit Codes](../exit-codes.md) for complete reference.

## Getting Help

Get help for any command:

```bash
# General help
clinvk --help

# Command help
clinvk [command] --help

# Example
clinvk parallel --help
```

## See Also

- [Configuration Reference](../configuration.md) - Configuration options
- [Environment Variables](../environment.md) - Environment-based settings
- [Exit Codes](../exit-codes.md) - Exit code reference
