# Development Guide

Information for developers who want to contribute to or extend clinvk.

## Overview

clinvk is written in Go and uses a modular architecture that makes it easy to add new backends and features.

## Getting Started

<div class="grid cards" markdown>

-   :material-cog:{ .lg .middle } **[Architecture](architecture.md)**

    ---

    System architecture and design decisions

-   :material-account-group:{ .lg .middle } **[Contributing](contributing.md)**

    ---

    How to contribute to the project

-   :material-plus-box:{ .lg .middle } **[Adding Backends](adding-backends.md)**

    ---

    Guide to implementing new backends

-   :material-test-tube:{ .lg .middle } **[Testing](testing.md)**

    ---

    Testing guidelines and practices

</div>

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
