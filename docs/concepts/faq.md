---
title: Frequently Asked Questions
description: Common questions and answers about clinvoker installation, usage, configuration, and integration.
---

# Frequently Asked Questions

This FAQ covers common questions about clinvoker, organized by topic. If you don't find your answer here, check the [Troubleshooting Guide](troubleshooting.md) or [GitHub Discussions](https://github.com/signalridge/clinvoker/discussions).

## General Questions

### What is clinvoker?

clinvoker is a unified AI CLI wrapper that provides a single interface for working with multiple AI coding assistants (Claude Code, Codex CLI, Gemini CLI). It offers session management, parallel execution, backend comparison, and an HTTP API server for integration with other tools.

### Why use clinvoker instead of individual CLIs?

- **Unified interface**: Same commands work across all backends
- **Session management**: Easily resume conversations across sessions
- **Parallel execution**: Run multiple tasks concurrently across backends
- **Backend comparison**: Compare responses from different AI models side-by-side
- **HTTP API**: Integrate AI capabilities into applications and workflows
- **Configuration cascade**: Consistent settings management across environments

### Which backends are supported?

Currently supported backends:

| Backend | Provider | Description |
|---------|----------|-------------|
| Claude Code | Anthropic | AI coding assistant with excellent code understanding |
| Codex CLI | OpenAI | Code-focused CLI with strong programming capabilities |
| Gemini CLI | Google | Gemini AI CLI with multimodal support |

### Is clinvoker free?

Yes, clinvoker itself is free and open source (MIT License). However, the underlying AI backends may have their own pricing and usage limits. You need valid API credentials for each backend you want to use.

### How does clinvoker compare to other tools?

| Feature | clinvoker | Aider | Continue |
|---------|-----------|-------|----------|
| Multiple backends | Yes | Limited | Via configuration |
| Session management | Built-in | Git-based | Editor-based |
| HTTP API | Yes | No | No |
| Parallel execution | Yes | No | No |
| Backend comparison | Yes | No | No |

## Installation

### How do I install clinvoker?

Multiple installation options are available:

```bash
# Homebrew (macOS/Linux)
brew install signalridge/tap/clinvk

# Go install
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# Nix
nix run github:signalridge/clinvoker

# Download binary from GitHub releases
# Visit: https://github.com/signalridge/clinvoker/releases
```bash

See [Installation Guide](../tutorials/getting-started.md) for detailed instructions.

### Do I need all backends installed?

No. clinvoker works with any combination of backends. Install only the ones you want to use. The tool will automatically detect which backends are available.

### What are the system requirements?

- **Operating System**: macOS, Linux, or Windows
- **Go**: 1.24+ (for building from source)
- **Memory**: 50MB RAM (clinvoker itself)
- **Disk**: 10MB for binary, plus space for sessions

### How do I verify which backends are available?

```bash
clinvk config show
```text

Look for `available: true` under each backend section.

## Usage

### How do I change the default backend?

```bash
# Set in configuration
clinvk config set default_backend codex

# Or use environment variable
export CLINVK_BACKEND=codex

# Or specify per-command
clinvk -b claude "your prompt"
```text

### How do I continue a conversation?

```bash
# Quick continue (resumes last session)
clinvk -c "follow up message"

# Resume specific session
clinvk resume <session-id> "follow up message"

# Resume last session
clinvk resume --last "follow up message"
```text

### Can I use different models?

Yes, specify the model with `--model`:

```bash
# Use specific model
clinvk -b claude -m claude-sonnet-4 "quick task"

# Use model aliases
clinvk -m fast "task"      # Fastest model
clinvk -m balanced "task"  # Balanced speed/quality
clinvk -m best "task"      # Best quality

# Set default in config
clinvk config set backends.claude.model claude-opus-4
```text

### How do I run tasks in parallel?

Create a tasks file and use the parallel command:

```bash
# Create tasks.json
{
  "tasks": [
    {"prompt": "Review this code", "backend": "claude"},
    {"prompt": "Review this code", "backend": "codex"},
    {"prompt": "Review this code", "backend": "gemini"}
  ]
}

# Run in parallel
clinvk parallel --file tasks.json --max-parallel 3
```text

See [Parallel Execution Guide](../guides/parallel.md) for details.

### How do I compare backend responses?

```bash
# Compare all available backends
clinvk compare --all-backends "your prompt"

# Compare specific backends
clinvk compare -b claude -b codex "your prompt"

# Save comparison to file
clinvk compare --all-backends -o comparison.md "your prompt"
```text

See [Backend Comparison](../guides/backends/index.md) for more.

### How do I chain multiple prompts?

```bash
# Create chain.json
{
  "steps": [
    {"backend": "claude", "prompt": "Generate a Python function to sort a list"},
    {"backend": "codex", "prompt": "Review and optimize this code: {{previous}}"},
    {"backend": "claude", "prompt": "Add tests for: {{previous}}"}
  ]
}

# Execute chain
clinvk chain --file chain.json
```bash

See [Chain Execution Guide](../guides/chains.md) for details.

## Configuration

### Where is the config file?

Default location: `~/.clinvk/config.yaml`

You can specify a different location:

```bash
clinvk --config /path/to/config.yaml "prompt"
```text

### What's the configuration priority?

Configuration is resolved in this order (highest to lowest):

1. CLI flags
2. Environment variables
3. Config file
4. Defaults

### How do I see the current configuration?

```bash
clinvk config show
```text

This displays the effective configuration after merging all sources.

### Can I use environment variables for all settings?

Yes, prefix any config key with `CLINVK_`:

```bash
export CLINVK_BACKEND=codex
export CLINVK_TIMEOUT=120
export CLINVK_SERVER_PORT=3000
```text

### How do I set backend-specific options?

```yaml
# ~/.clinvk/config.yaml
backends:
  claude:
    model: claude-sonnet-4
    timeout: 120
  codex:
    model: gpt-5.2
    sandbox_mode: full
```bash

## Sessions

### Where are sessions stored?

Sessions are stored as JSON files in `~/.clinvk/sessions/`.

### How do I list all sessions?

```bash
clinvk sessions list

# Filter by backend
clinvk sessions list --backend claude

# Show detailed info
clinvk sessions list --verbose
```text

### How do I clean up old sessions?

```bash
# Clean sessions older than 30 days
clinvk sessions clean --older-than 30d

# Clean all sessions
clinvk sessions clean --all

# Or manually delete
rm -rf ~/.clinvk/sessions/*
```text

### Can I disable session tracking?

Yes, use ephemeral mode:

```bash
clinvk --ephemeral "prompt"
```bash

This runs without creating or loading any sessions.

### How do I export a session?

```bash
# Export session to file
clinvk sessions export <session-id> > session.json

# Or copy the file directly
cp ~/.clinvk/sessions/<session-id>.json ./backup.json
```text

## HTTP Server

### How do I start the server?

```bash
# Start with default settings
clinvk serve

# Start with custom port
clinvk serve --port 3000

# Start with API key authentication
clinvk serve --api-keys "key1,key2"
```text

### Is the server authenticated?

Authentication is optional. If you configure API keys, all requests must include a valid key:

```bash
# Configure keys
export CLINVK_API_KEYS="key1,key2,key3"

# Or in config
clinvk config set server.api_keys "key1,key2"

# Use in requests
curl -H "Authorization: Bearer key1" http://localhost:8080/api/v1/prompt
```text

If no keys are configured, the server allows all requests.

### How do I expose the server publicly?

Place it behind a reverse proxy (nginx, Apache, Caddy) and enable API keys:

```bash
# Bind to all interfaces (use with caution)
clinvk serve --host 0.0.0.0 --api-keys "your-secret-key"
```text

### Can I use OpenAI client libraries?

Yes, the server provides OpenAI-compatible endpoints:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-api-key"  # Required if API keys are enabled
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Hello!"}]
)
```text

### Can I use Anthropic client libraries?

Yes, Anthropic-compatible endpoints are also available:

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="your-api-key"
)

response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}]
)
```text

## Integration

### How do I integrate with Claude Code?

clinvoker works alongside Claude Code:

```bash
# Use Claude Code through clinvoker
clinvk -b claude "your prompt"

# Or start Claude Code's interactive mode
claude
```text

See [Claude Backend Guide](../guides/backends/claude.md) for details.

### How do I use with LangChain?

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
    model="claude"
)

response = llm.invoke("Hello!")
```text

See [LangChain Integration Guide](../guides/integrations/langchain-langgraph.md) for more.

### How do I use in CI/CD pipelines?

```yaml
# GitHub Actions example
- name: Code Review
  env:
    CLINVK_BACKEND: claude
  run: |
    echo '{"prompt": "Review this PR", "files": ["src/"] }' | \
    clinvk parallel --file - --output-format json
```bash

See [CI/CD Integration Guide](../guides/integrations/ci-cd/index.md) for examples.

## Troubleshooting

### Why is my backend not detected?

Check if the CLI is in your PATH:

```bash
which claude codex gemini
echo $PATH
```text

### Why is my configuration not applied?

Remember the priority order: CLI flags > Environment > Config file. Check for overrides:

```bash
clinvk config show  # Shows effective configuration
env | grep CLINVK   # Shows environment variables
```text

### Where can I get help?

- [Troubleshooting Guide](troubleshooting.md)
- [GitHub Issues](https://github.com/signalridge/clinvoker/issues)
- [GitHub Discussions](https://github.com/signalridge/clinvoker/discussions)

## Contributing

### How can I contribute?

See [Contributing Guide](contributing.md) for:
- Development setup
- Coding standards
- Testing requirements
- PR process

### How do I add a new backend?

1. Implement the `Backend` interface
2. Add to the registry
3. Add unified options mapping
4. Add tests
5. Update documentation

See [Backend System](backend-system.md) for the interface details.

### How do I report a bug?

Open an issue on [GitHub](https://github.com/signalridge/clinvoker/issues) with:

- clinvk version (`clinvk version`)
- OS and version
- Backend versions
- Steps to reproduce
- Error message
- Debug logs (if possible)

## Related Documentation

- [Troubleshooting](troubleshooting.md) - Common issues and solutions
- [Guides](../guides/index.md) - How-to guides
- [Reference](../reference/index.md) - API and CLI reference
- [Concepts](index.md) - Architecture and design
