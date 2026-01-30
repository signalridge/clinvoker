# Commands Reference

Complete reference for all clinvk CLI commands.

## Synopsis

```bash
clinvk [flags] [prompt]
clinvk [command] [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| [`[prompt]`](prompt.md) | Execute a prompt (default command) |
| [`resume`](resume.md) | Resume a previous session |
| [`sessions`](sessions.md) | Manage sessions |
| [`config`](config.md) | Manage configuration |
| [`parallel`](parallel.md) | Execute tasks in parallel |
| [`compare`](compare.md) | Compare backend responses |
| [`chain`](chain.md) | Execute prompt chain |
| [`serve`](serve.md) | Start HTTP API server |
| `version` | Show version information |
| `help` | Show help |

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

## Flag Details

### --backend, -b

Select the AI backend to use:

```bash
clinvk --backend claude "prompt"
clinvk -b codex "prompt"
clinvk -b gemini "prompt"
```yaml

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
```bash

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
```bash

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
```bash

## Examples

```bash
# Basic prompt
clinvk "fix the bug in auth.go"

# With options
clinvk -b codex -m o3 -w ./project "implement feature"

# Show command without running
clinvk --dry-run -b claude "complex task"

# JSON output
clinvk -o json "explain this"

# Stateless query
clinvk --ephemeral "what is 2+2"
```
