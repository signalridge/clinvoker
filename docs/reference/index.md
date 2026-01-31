# Reference Documentation

Complete technical reference for clinvk commands, configuration, and APIs.

## Overview

This reference section provides detailed documentation for all clinvk features, options, and behaviors. Use this section when you need precise information about specific commands, configuration options, or API endpoints.

## How to Use This Reference

- **CLI Commands**: Look up specific commands for syntax, flags, and examples
- **Configuration**: Find all available configuration options and their defaults
- **Environment Variables**: Discover environment-based configuration options
- **Exit Codes**: Understand program exit codes for scripting
- **API Reference**: Integrate clinvk with your applications

## Quick Navigation

| Section | Description | Use When... |
|---------|-------------|-------------|
| [CLI Commands](cli/index.md) | Command-line interface reference | You need command syntax or flags |
| [Configuration](configuration.md) | Configuration file options | Setting up or modifying config |
| [Environment Variables](environment.md) | Environment-based settings | Configuring via environment |
| [Exit Codes](exit-codes.md) | Program exit codes | Writing scripts with clinvk |
| [API Reference](api/index.md) | HTTP API documentation | Integrating with applications |

## CLI Commands Overview

| Command | Purpose | Common Use Case |
|---------|---------|-----------------|
| [`clinvk [prompt]`](cli/prompt.md) | Execute a prompt | Daily AI assistance |
| [`clinvk resume`](cli/resume.md) | Resume a session | Continue conversations |
| [`clinvk sessions`](cli/sessions.md) | Manage sessions | Clean up or inspect sessions |
| [`clinvk config`](cli/config.md) | Manage configuration | View or change settings |
| [`clinvk parallel`](cli/parallel.md) | Parallel execution | Run multiple tasks |
| [`clinvk compare`](cli/compare.md) | Compare backends | Evaluate different AIs |
| [`clinvk chain`](cli/chain.md) | Chain execution | Multi-step workflows |
| [`clinvk serve`](cli/serve.md) | HTTP API server | Application integration |

## Configuration Priority

Configuration values are resolved in this order (highest to lowest priority):

1. **CLI Flags** - Command-line arguments override everything
2. **Environment Variables** - `CLINVK_*` variables
3. **Config File** - `~/.clinvk/config.yaml`
4. **Defaults** - Built-in default values

```bash
# Example: CLI flag wins over environment variable
export CLINVK_BACKEND=codex
clinvk -b claude "prompt"  # Uses claude, not codex
```text

## Common Flags Reference

These flags work across most commands:

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--backend` | `-b` | Backend to use | `-b codex` |
| `--model` | `-m` | Model override | `-m o3-mini` |
| `--workdir` | `-w` | Working directory | `-w ./project` |
| `--output-format` | `-o` | Output format | `-o json` |
| `--config` | | Custom config path | `--config /path/to/config.yaml` |
| `--dry-run` | | Show command only | `--dry-run` |
| `--help` | `-h` | Show help | `-h` |

## Backends Reference

| Backend | Binary | Default Model | Best For |
|---------|--------|---------------|----------|
| Claude | `claude` | Backend default | Complex reasoning, code review |
| Codex | `codex` | Backend default | Quick coding tasks |
| Gemini | `gemini` | Backend default | General assistance |

## Output Formats

| Format | Description | Best For |
|--------|-------------|----------|
| `text` | Plain text output | Human reading |
| `json` | Structured JSON | Scripting, parsing |
| `stream-json` | Streaming JSON events | Real-time processing |

## API Compatibility

clinvk provides three API styles for integration:

| API Style | Endpoint Prefix | Best For |
|-----------|-----------------|----------|
| Native REST | `/api/v1/` | Full clinvk features |
| OpenAI Compatible | `/openai/v1/` | OpenAI SDK users |
| Anthropic Compatible | `/anthropic/v1/` | Anthropic SDK users |

## Getting Help

- Use `--help` with any command for quick reference
- Check [Troubleshooting](../concepts/troubleshooting.md) for common issues
- See [FAQ](../concepts/faq.md) for frequently asked questions
