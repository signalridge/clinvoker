---
title: Contributing Guide
description: Development setup, coding standards, and contribution process for clinvoker.
---

# Contributing Guide

Thank you for your interest in contributing to clinvoker! This guide covers development environment setup, project structure, coding standards, testing requirements, and the contribution process.

## Development Environment Setup

### Prerequisites

- Go 1.24 or later
- Git
- (Optional) Nix for reproducible environment
- (Optional) golangci-lint for linting
- (Optional) pre-commit for git hooks

### Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/clinvoker.git
cd clinvoker
git remote add upstream https://github.com/signalridge/clinvoker.git
```text

### Using Nix (Recommended)

```bash
nix develop
```text

This provides all required tools in a reproducible environment.

### Manual Setup

```bash
go mod download
go build ./cmd/clinvk
./clinvk version
```text

### Pre-commit Hooks

```bash
pre-commit install
pre-commit install --hook-type commit-msg
```text

## Project Structure

```text
clinvoker/
├── cmd/
│   └── clinvk/           # Main application entry point
│       └── main.go
├── internal/
│   ├── app/              # CLI command implementations
│   ├── backend/          # Backend abstraction layer
│   ├── config/           # Configuration management
│   ├── executor/         # Command execution
│   ├── output/           # Output formatting
│   ├── server/           # HTTP API server
│   ├── session/          # Session management
│   ├── auth/             # API key management
│   ├── metrics/          # Prometheus metrics
│   └── resilience/       # Circuit breaker
├── docs/                 # Documentation
├── scripts/              # Build and utility scripts
└── test/                 # Integration tests
```bash

### Package Responsibilities

| Package | Purpose | Key Files |
|---------|---------|-----------|
| `app/` | CLI commands using Cobra | `app.go`, `cmd_*.go` |
| `backend/` | Backend abstraction | `backend.go`, `registry.go`, `claude.go`, `codex.go`, `gemini.go` |
| `config/` | Viper-based configuration | `config.go`, `validate.go` |
| `executor/` | Subprocess execution | `executor.go`, `signal.go` |
| `server/` | HTTP server | `server.go`, `routes.go`, `handlers/`, `middleware/` |
| `session/` | Session persistence | `session.go`, `store.go`, `filelock.go` |

## Coding Standards

### Go Guidelines

Follow [Effective Go](https://golang.org/doc/effective_go.html) and [Google Go Style Guide](https://google.github.io/styleguide/go/):

1. **Formatting**: Use `gofmt` or `goimports`
2. **Linting**: Run `golangci-lint run` before committing
3. **Naming**: Use descriptive, idiomatic names
4. **Comments**: Document all exported types and functions
5. **Error Handling**: Wrap errors with context, avoid naked returns

### Code Style Examples

```go
// Good: Clear function name, proper documentation, error handling
// ExecuteCommand runs the given command and returns the output.
func ExecuteCommand(ctx context.Context, cfg *Config, cmd *exec.Cmd) (*Result, error) {
    if cfg == nil {
        return nil, fmt.Errorf("config is required")
    }

    // Implementation
    result, err := runWithTimeout(ctx, cmd, cfg.Timeout)
    if err != nil {
        return nil, fmt.Errorf("failed to execute command: %w", err)
    }

    return result, nil
}

// Bad: Unclear name, missing documentation, poor error handling
func exec(cfg *Config, c *exec.Cmd) (*Result, error) {
    res, _ := runWithTimeout(context.Background(), c, cfg.Timeout)
    return res, nil
}
```go

### Error Handling

Use the project's error package for consistent error handling:

```go
import apperrors "github.com/signalridge/clinvoker/internal/errors"

// Create error with context
return apperrors.BackendError("claude", err)

// Check error type
if apperrors.IsCode(err, apperrors.ErrCodeBackendUnavailable) {
    // Handle specific error
}
```bash

### Testing Standards

All code must have tests. Follow these guidelines:

1. **File Naming**: `*_test.go` alongside source files
2. **Table-Driven Tests**: Use for multiple test cases
3. **Parallel Tests**: Use `t.Parallel()` for independent tests
4. **Coverage**: Aim for >80% coverage for new code
5. **Mocking**: Use interfaces for testability

```go
func TestExecuteCommand(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        cfg     *Config
        cmd     *exec.Cmd
        wantErr bool
    }{
        {
            name:    "valid command",
            cfg:     &Config{Timeout: 30 * time.Second},
            cmd:     exec.Command("echo", "hello"),
            wantErr: false,
        },
        {
            name:    "nil config",
            cfg:     nil,
            cmd:     exec.Command("echo", "hello"),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            ctx := context.Background()
            _, err := ExecuteCommand(ctx, tt.cfg, tt.cmd)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```text

## Testing Requirements

### Running Tests

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt -o coverage.html

# Short tests only
go test -short ./...

# Specific package
go test ./internal/backend/...
```text

### Integration Tests

Integration tests require actual backend CLIs to be installed:

```bash
# Run integration tests
CLINVK_TEST_INTEGRATION=1 go test ./test/...

# Run with specific backend
CLINVK_TEST_BACKEND=claude go test ./test/...
```text

### Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# Run with memory profiling
go test -bench=. -benchmem ./...
```text

## Documentation Requirements

### Code Comments

Document all exported types and functions:

```go
// Backend represents an AI CLI backend.
type Backend interface {
    // Name returns the backend identifier.
    Name() string

    // IsAvailable checks if the backend CLI is installed.
    IsAvailable() bool
}

// NewStore creates a new session store at the given directory.
// The directory is created if it doesn't exist.
func NewStore(dir string) (*Store, error) {
    // ...
}
```text

### User Documentation

Update documentation for user-facing changes:

1. **Concepts**: Update architecture docs if design changes
2. **Guides**: Add/update how-to guides for new features
3. **Reference**: Update API/CLI reference for changes
4. **Changelog**: Add entry to CHANGELOG.md

### Documentation Style

- Use clear, concise language
- Include code examples
- Add diagrams for complex concepts (Mermaid)
- Keep Chinese and English versions in sync

## PR Process

### Branch Naming

Use conventional branch names:

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feat/` | New features | `feat/add-gemini-backend` |
| `fix/` | Bug fixes | `fix/session-locking` |
| `docs/` | Documentation | `docs/api-examples` |
| `refactor/` | Code refactoring | `refactor/executor` |
| `test/` | Test additions | `test/backend-coverage` |
| `chore/` | Maintenance | `chore/update-deps` |

### Commit Message Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
type(scope): description

[optional body]

[optional footer]
```yaml

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:

```text
feat(backend): add support for new AI provider

Add support for the new XYZ AI CLI tool. This includes:
- Backend implementation
- Flag mapping
- Documentation updates

fix(session): handle concurrent access correctly

Fix race condition in session store when multiple processes
attempt to update the same session simultaneously.

docs(readme): update installation instructions

test(backend): add unit tests for registry

chore(deps): update golang.org/x packages
```text

### Pull Request Template

```markdown
## Summary

Brief description of changes.

## Changes

- Change 1
- Change 2

## Test Plan

How changes were tested:
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing performed

## Related Issues

Fixes #123
Closes #456
```text

### Code Review Checklist

Before submitting PR:

- [ ] All tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] Commit messages follow conventions
- [ ] No unnecessary files included
- [ ] CHANGELOG.md updated (if applicable)

### Review Process

1. PRs require at least one approval
2. CI must pass before merging
3. Address review comments promptly
4. Keep PRs focused and reasonably sized (<500 lines preferred)
5. Rebase on main if there are conflicts

## Release Process

### Versioning

clinvoker follows [Semantic Versioning](https://semver.org/):

- `MAJOR`: Breaking changes
- `MINOR`: New features, backwards compatible
- `PATCH`: Bug fixes, backwards compatible

### Creating a Release

1. Update version in `internal/version/version.go`
2. Update CHANGELOG.md
3. Create git tag: `git tag v1.2.3`
4. Push tag: `git push origin v1.2.3`
5. GitHub Actions builds and publishes release

## Questions?

- Open an issue for bugs or features
- Start a discussion for questions
- Check existing issues before creating new ones

## Code of Conduct

By participating, you agree to:

1. Be respectful and inclusive
2. Accept constructive criticism gracefully
3. Focus on what's best for the community
4. Show empathy towards others

## Related Documentation

- [Architecture Overview](architecture.md) - System architecture
- [Design Decisions](design-decisions.md) - Architectural ADRs
- [Troubleshooting](troubleshooting.md) - Common issues
