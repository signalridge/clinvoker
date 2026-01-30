# Contributing Guide

Thank you for your interest in contributing to clinvk!

## Code of Conduct

By participating, you agree to maintain a respectful and inclusive environment for everyone.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- (Optional) Nix for reproducible environment
- (Optional) golangci-lint for linting

### Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/clinvoker.git
cd clinvoker
git remote add upstream https://github.com/signalridge/clinvoker.git
```text

## Development Setup

### Using Nix (Recommended)

```bash
nix develop
```bash

Provides all required tools in a reproducible environment.

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
```

## Making Changes

### Branch Naming

Use conventional branch names:

| Prefix | Purpose |
|--------|---------|
| `feat/` | New features |
| `fix/` | Bug fixes |
| `docs/` | Documentation |
| `refactor/` | Code refactoring |
| `test/` | Test additions |
| `chore/` | Maintenance |

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```yaml
type(scope): description

[optional body]

[optional footer]
```yaml

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:

```yaml
feat(backend): add support for new AI provider
fix(session): handle concurrent access correctly
docs(readme): update installation instructions
```

## Testing

### Running Tests

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.txt ./...

# Short tests only
go test -short ./...
```text

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests
- Test both success and error paths
- Use `t.Parallel()` for independent tests

```go
func TestMyFunction(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid", "input", "output", false},
        {"error", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```text

## Code Style

### Go Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Run `golangci-lint run` before committing
- Keep functions focused and reasonably sized
- Use meaningful names
- Document exported functions

### Error Handling

```go
import apperrors "github.com/signalridge/clinvoker/internal/errors"

// Create error with context
return apperrors.BackendError("claude", err)

// Check error type
if apperrors.IsCode(err, apperrors.ErrCodeBackendUnavailable) {
    // Handle specific error
}
```text

## Submitting Changes

### Pull Request Process

1. Ensure all tests pass locally
2. Update documentation if needed
3. Create PR with clear description
4. Reference related issues

### PR Template

```markdown
## Summary

Brief description of changes.

## Changes

- Change 1
- Change 2

## Test Plan

How changes were tested.

## Related Issues

Fixes #123
```

### Review Process

- PRs require at least one approval
- CI must pass before merging
- Address review comments promptly
- Keep PRs focused and reasonably sized

## Documentation

### Code Comments

Document exported types and functions:

```go
// Config represents the application configuration.
type Config struct {
    // ...
}

// LoadConfig loads configuration from the specified path.
func LoadConfig(path string) (*Config, error) {
    // ...
}
```

### User Documentation

- Update docs for user-facing changes
- Add examples for new features
- Keep docs in sync with code

## Questions?

- Open an issue for bugs or features
- Start a discussion for questions
- Check existing issues before creating new ones
