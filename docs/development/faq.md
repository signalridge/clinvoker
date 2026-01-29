# Frequently Asked Questions

## General

### What is clinvk?

clinvk is a unified AI CLI wrapper that lets you work with multiple AI coding assistants (Claude Code, Codex CLI, Gemini CLI) through a single interface. It provides session management, parallel execution, backend comparison, and an HTTP API server.

### Why use clinvk instead of the individual CLIs?

- **Unified interface** - Same commands work across all backends
- **Session management** - Easily resume conversations
- **Parallel execution** - Run multiple tasks concurrently
- **Backend comparison** - Compare responses side-by-side
- **HTTP API** - Integrate AI capabilities into other tools
- **Configuration cascade** - Consistent settings management

### Which backends are supported?

Currently supported:

- **Claude Code** - Anthropic's AI coding assistant
- **Codex CLI** - OpenAI's code-focused CLI
- **Gemini CLI** - Google's Gemini AI CLI

### Is clinvk free?

clinvk itself is free and open source. However, the underlying AI backends may have their own pricing and usage limits.

## Installation

### How do I install clinvk?

Multiple options:

```bash
# Homebrew
brew install signalridge/tap/clinvk

# Go
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# Nix
nix run github:signalridge/clinvoker
```

See [Installation](../guide/installation.md) for all options.

### Do I need all backends installed?

No. clinvk works with any combination of backends. Install only the ones you want to use.

### How do I verify which backends are available?

```bash
clinvk config show
```

Look for `available: true` under each backend.

## Usage

### How do I change the default backend?

```bash
clinvk config set default_backend codex
```

Or set the environment variable:

```bash
export CLINVK_BACKEND=codex
```

### How do I continue a conversation?

Use `--continue` or the resume command:

```bash
# Quick continue
clinvk -c "follow up message"

# Resume command
clinvk resume --last "follow up message"
```

### Can I use different models?

Yes, specify the model with `--model`:

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "quick task"
```

Or set it in configuration:

```bash
clinvk config set backends.claude.model claude-sonnet-4-20250514
```

### How do I run tasks in parallel?

Create a tasks file and use the parallel command:

```bash
clinvk parallel --file tasks.json
```

See [Parallel Execution](../guide/parallel-execution.md).

### How do I compare backend responses?

```bash
clinvk compare --all-backends "your prompt"
```

See [Backend Comparison](../guide/backend-comparison.md).

## Configuration

### Where is the config file?

Default location: `~/.clinvk/config.yaml`

### What's the configuration priority?

1. CLI flags (highest)
2. Environment variables
3. Config file
4. Defaults (lowest)

### How do I see the current configuration?

```bash
clinvk config show
```

## Sessions

### Where are sessions stored?

Sessions are stored as JSON files in `~/.clinvk/sessions/`.

### How do I clean up old sessions?

```bash
clinvk sessions clean --older-than 30d
```

### Can I disable session tracking?

Use ephemeral mode:

```bash
clinvk --ephemeral "prompt"
```

## HTTP Server

### Is the server authenticated?

No. The server has no built-in authentication and is designed for local use only.

### How do I expose the server publicly?

Place it behind a reverse proxy with authentication:

```bash
# Bind to all interfaces (use with caution)
clinvk serve --host 0.0.0.0
```

### Can I use OpenAI client libraries?

Yes, the server provides OpenAI-compatible endpoints at `/openai/v1/`:

```python
from openai import OpenAI
client = OpenAI(base_url="http://localhost:8080/openai/v1", api_key="not-needed")
```

## Troubleshooting

### Why is my backend not detected?

Check if the CLI is in your PATH:

```bash
which claude codex gemini
```

### Why is my configuration not applied?

Check priority: CLI flags override environment variables, which override config file.

### Where can I get help?

- [Troubleshooting Guide](troubleshooting.md)
- [GitHub Issues](https://github.com/signalridge/clinvoker/issues)

## Contributing

### How can I contribute?

See [Contributing Guide](../development/contributing.md).

### How do I add a new backend?

See [Adding Backends](../development/adding-backends.md).

### How do I report a bug?

Open an issue on [GitHub](https://github.com/signalridge/clinvoker/issues) with:

- clinvk version
- OS and version
- Steps to reproduce
- Error message
