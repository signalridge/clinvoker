# Contributing to clinvk

Thank you for your interest in contributing to clinvk! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Documentation](#documentation)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- (Optional) Nix for reproducible development environment
- (Optional) golangci-lint for linting

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:

```bash
git clone https://github.com/YOUR_USERNAME/clinvoker.git
cd clinvoker
```

3. Add upstream remote:

```bash
git remote add upstream https://github.com/signalridge/clinvoker.git
```

## Development Setup

### Using Nix (Recommended)

```bash
nix develop
```

This provides all required tools in a reproducible environment.

### Manual Setup

```bash
# Download dependencies
go mod download

# Build
go build ./cmd/clinvk

# Verify installation
./clinvk version
```

### Pre-commit Hooks

Install pre-commit hooks to ensure code quality before committing:

```bash
pre-commit install
pre-commit install --hook-type commit-msg
```

## Making Changes

### Branch Naming

Use conventional branch names:

- `feat/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation changes
- `refactor/description` - Code refactoring
- `test/description` - Test additions or modifications
- `chore/description` - Maintenance tasks

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:

```
feat(backend): add support for new AI provider
fix(session): handle concurrent access correctly
docs(readme): update installation instructions
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.txt ./...

# Short tests only (for quick feedback)
go test -short ./...
```

### Writing Tests

- Place tests in `*_test.go` files alongside the code
- Use table-driven tests for multiple scenarios
- Test both success and error paths
- Use `t.Parallel()` for independent tests
- Mock external dependencies

Example:

```go
func TestMyFunction(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "result",
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. Ensure all tests pass locally
2. Update documentation if needed
3. Create a pull request with a clear description
4. Reference any related issues

### Pull Request Template

```markdown
## Summary

Brief description of the changes.

## Changes

- Change 1
- Change 2

## Test Plan

How the changes were tested.

## Related Issues

Fixes #123
```

### Review Process

- PRs require at least one approval
- CI must pass before merging
- Address review comments promptly
- Keep PRs focused and reasonably sized

## Code Style

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Run `golangci-lint run` before committing
- Keep functions focused and reasonably sized
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

### File Organization

```
clinvoker/
├── cmd/clinvk/         # Main entry point
├── internal/
│   ├── app/            # Application commands
│   ├── backend/        # Backend implementations
│   ├── config/         # Configuration handling
│   ├── executor/       # Execution logic
│   ├── output/         # Output parsing
│   ├── server/         # HTTP server
│   └── session/        # Session management
├── docs/               # Documentation
└── testdata/           # Test fixtures
```

### Error Handling

- Return errors, don't panic
- Wrap errors with context
- Use custom error types for domain errors

```go
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

## Documentation

### Code Comments

- Document all exported types and functions
- Use godoc format for documentation comments

```go
// Config represents the application configuration.
// It is loaded from the config file and environment variables.
type Config struct {
    // ...
}

// LoadConfig loads configuration from the specified path.
// It returns an error if the file cannot be read or parsed.
func LoadConfig(path string) (*Config, error) {
    // ...
}
```

### User Documentation

- Update README.md for user-facing changes
- Update docs/CLI.md for command changes
- Add examples for new features

## Adding a New Backend

1. Create a new file in `internal/backend/`
2. Implement the `Backend` interface
3. Register in `internal/backend/registry.go`
4. Add configuration support in `internal/config/`
5. Add tests
6. Update documentation

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions
- Check existing issues before creating new ones

Thank you for contributing!
