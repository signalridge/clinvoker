# Development Guide

Information for developers who want to contribute to or extend clinvk.

## Overview

clinvk is written in Go and uses a modular architecture that makes it easy to add new backends and features.

## Getting Started

- **[Architecture](architecture.md)** - System architecture and design decisions
- **[Contributing](contributing.md)** - How to contribute to the project
- **[Adding Backends](adding-backends.md)** - Guide to implementing new backends
- **[Testing](testing.md)** - Testing guidelines and practices

## Quick Start

### Using Nix (Recommended)

```bash
nix develop
just ci
```

### Manual Setup

```bash
# Requirements: Go 1.24+
go mod download
go build ./cmd/clinvk
./clinvk version
```

## Pre-commit Hooks

This repo includes `.pre-commit-config.yaml` for formatting and lint checks.

```bash
# Install pre-commit (macOS)
brew install pre-commit

# Install hooks for this repo
pre-commit install

# Run on all files
pre-commit run --all-files
```

## Project Structure

```
clinvoker/
├── cmd/clinvk/           # Entry point
├── internal/
│   ├── app/              # CLI commands
│   ├── backend/          # Backend implementations
│   ├── config/           # Configuration
│   ├── errors/           # Error types
│   ├── executor/         # Execution logic
│   ├── output/           # Output parsing
│   ├── server/           # HTTP server
│   │   ├── handlers/     # API handlers
│   │   └── service/      # Business logic
│   ├── session/          # Session management
│   └── mock/             # Test utilities
├── docs/                 # Documentation
└── testdata/             # Test fixtures
```

## Common Tasks

```bash
# Build
just build

# Test
just test

# Lint
just lint

# Run all checks
just ci

# Start dev server
just serve
```

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24+ |
| CLI Framework | Cobra |
| HTTP Server | Huma/v2 |
| Configuration | Viper |
| Testing | Standard library + testify |
| Linting | golangci-lint |
| Build | Just (task runner) |
| CI/CD | GitHub Actions |
| Package Manager | Nix (optional) |

## Key Concepts

### Backend Abstraction

All backends implement a common interface:

```go
type Backend interface {
    Name() string
    IsAvailable() bool
    BuildCommand(prompt string, opts *Options) *exec.Cmd
    ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd
    ParseOutput(rawOutput string) string
}
```

### Configuration Cascade

Settings are resolved in priority order:

1. CLI flags
2. Environment variables
3. Config file
4. Defaults

### Session Management

Sessions are JSON files in `~/.clinvk/sessions/` containing metadata about each conversation.

## Development Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Huma API Framework](https://huma.rocks/)
- [Just Task Runner](https://just.systems/)
