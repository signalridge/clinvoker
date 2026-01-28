# Development Guide

This guide covers how to set up and work with the clinvk codebase.

## Quick Start

### Using Nix (Recommended)

```bash
# Enter development shell with all tools
nix develop

# Build and test
just ci
```

### Manual Setup

```bash
# Requirements: Go 1.24+, golangci-lint, just

# Install tools
just setup

# Build and test
just ci
```

## Project Structure

```
clinvoker/
├── cmd/clinvk/           # Main entry point
├── internal/
│   ├── app/              # CLI commands and execution
│   ├── backend/          # Backend implementations (claude, codex, gemini)
│   ├── config/           # Configuration management
│   ├── errors/           # Structured error types
│   ├── executor/         # Command execution logic
│   ├── output/           # Output parsing and formatting
│   ├── server/           # HTTP API server
│   │   ├── handlers/     # API handlers (custom, OpenAI, Anthropic)
│   │   └── service/      # Business logic
│   ├── session/          # Session persistence
│   └── testutil/         # Test utilities
├── docs/                 # Documentation
└── .github/              # CI/CD workflows
```

## Common Tasks

### Building

```bash
# Simple build
just build

# Build with version info
just build-release v1.0.0 abc1234

# Clean build artifacts
just clean
```

### Testing

```bash
# Run all tests
just test

# Run with verbose output
just test-verbose

# Run short tests only (quick)
just test-short

# Generate coverage report
just test-coverage

# View HTML coverage report
just coverage-html
```

### Linting

```bash
# Run all linters
just lint

# Go linter only
just lint-go

# Auto-fix issues
just lint-fix
```

### Development Server

```bash
# Start HTTP server on default port
just serve

# Start on custom port
just serve 3000
```

### Docker

```bash
# Build container
just docker-build

# Run container
just docker-run dev --help
```

## Adding a New Backend

1. Create a new file in `internal/backend/`:

```go
package backend

type NewBackend struct{}

func (b *NewBackend) Name() string { return "new" }
func (b *NewBackend) IsAvailable() bool {
    _, err := exec.LookPath("new-cli")
    return err == nil
}
// ... implement remaining interface methods
```

2. Register in `internal/backend/registry.go`

3. Add configuration in `internal/config/config.go`:
   - Add to `BackendConfig` map
   - Add environment variable binding

4. Add tests in `internal/backend/new_test.go`

5. Update documentation:
   - `README.md` - Add to features list
   - `docs/CLI.md` - Add command examples
   - `config.example.yaml` - Add configuration example

## Code Patterns

### Error Handling

Use structured errors from `internal/errors`:

```go
import apperrors "github.com/signalridge/clinvoker/internal/errors"

// Create error with context
return apperrors.BackendError("claude", err)

// Check error type
if apperrors.IsCode(err, apperrors.ErrCodeBackendUnavailable) {
    // Handle specific error
}
```

### Configuration

Access configuration through the `config` package:

```go
import "github.com/signalridge/clinvoker/internal/config"

cfg := config.Get()
backend := cfg.DefaultBackend

// Validate configuration
if err := config.ValidateConfig(); err != nil {
    return err
}
```

### Testing

Use table-driven tests and the testutil package:

```go
import "github.com/signalridge/clinvoker/internal/testutil"

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
            testutil.AssertEqual(t, got, tt.want)
        })
    }
}
```

### Mock Backends

Use `testutil.MockBackend` for testing:

```go
mock := testutil.NewMockBackend("test",
    testutil.WithParseOutput("mocked output"),
    testutil.WithAvailable(true),
)
```

## Release Process

1. Update `CHANGELOG.md` with new version

2. Create a version tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. CI will automatically:
   - Build binaries for all platforms
   - Build Docker images
   - Generate SBOM
   - Create GitHub release
   - Update Homebrew/Scoop/AUR

4. Verify release:
   ```bash
   # Check Nix build
   nix run github:signalridge/clinvoker -- version

   # Check container
   docker run --rm ghcr.io/signalridge/clinvk:v1.0.0 version
   ```

## Debugging

### Enable Verbose Output

```bash
clinvk --verbose "prompt"
```

### Dry Run Mode

```bash
clinvk --dry-run "prompt"
```

### Check Configuration

```bash
clinvk config show
```

### View Session Details

```bash
clinvk sessions list
clinvk sessions show <session-id>
```

## Performance

### Run Benchmarks

```bash
just bench
```

### Profile CPU/Memory

```bash
go test -cpuprofile=cpu.out -memprofile=mem.out -bench=. ./...
go tool pprof cpu.out
```

## Troubleshooting

### Backend Not Found

Ensure the backend CLI is in your PATH:

```bash
which claude codex gemini
```

### Configuration Not Loading

Check configuration file location and format:

```bash
cat ~/.clinvk/config.yaml
clinvk config show
```

### Session Issues

Clear old sessions:

```bash
clinvk sessions clean --older-than 1d
```

## Resources

- [CLI Reference](CLI.md)
- [Architecture Overview](ARCHITECTURE.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
- [Security Policy](../SECURITY.md)
