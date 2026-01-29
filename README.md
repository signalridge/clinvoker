# clinvoker

> **Command:** `clinvk`

Unified AI CLI wrapper for orchestrating multiple AI backends (Claude Code, Codex CLI, Gemini CLI).

[![CI](https://github.com/signalridge/clinvoker/actions/workflows/ci.yaml/badge.svg)](https://github.com/signalridge/clinvoker/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/signalridge/clinvoker)](https://goreportcard.com/report/github.com/signalridge/clinvoker)
[![Go Reference](https://pkg.go.dev/badge/github.com/signalridge/clinvoker.svg)](https://pkg.go.dev/github.com/signalridge/clinvoker)
[![GitHub release](https://img.shields.io/github/v/release/signalridge/clinvoker)](https://github.com/signalridge/clinvoker/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/signalridge/clinvoker/blob/main/LICENSE)

## Features

- **Multi-Backend** - Switch between Claude, Codex, and Gemini seamlessly
- **Session Management** - Persist and resume conversations
- **Parallel Execution** - Run tasks concurrently across backends
- **HTTP API Server** - REST API with OpenAI/Anthropic compatible endpoints
- **Cross-Platform** - Linux, macOS, Windows

## Installation

```bash
# Homebrew
brew install signalridge/tap/clinvk

# Go
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# Nix
nix run github:signalridge/clinvoker
```

See [Installation Guide](https://signalridge.github.io/clinvoker/getting-started/installation/) for more options.

## Quick Start

```bash
# Run with default backend
clinvk "fix the bug in auth.go"

# Use specific backend
clinvk -b codex "implement user registration"

# Resume last session
clinvk resume --last "continue"

# Compare backends
clinvk compare --all-backends "explain this code"

# Start HTTP server
clinvk serve --port 8080
```

## Documentation

Full documentation: **[signalridge.github.io/clinvoker](https://signalridge.github.io/clinvoker/)**

- [Getting Started](https://signalridge.github.io/clinvoker/getting-started/)
- [User Guide](https://signalridge.github.io/clinvoker/user-guide/)
- [HTTP API](https://signalridge.github.io/clinvoker/server/)
- [Reference](https://signalridge.github.io/clinvoker/reference/)

## Contributing

Contributions welcome! See [Contributing Guide](https://signalridge.github.io/clinvoker/development/contributing/).

## License

MIT License - see [LICENSE](LICENSE).
