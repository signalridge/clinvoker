# Testing Guide

Guidelines and practices for testing clinvk.

## Running Tests

### All Tests

```bash
go test ./...
```

### With Race Detection

```bash
go test -race ./...
```

### With Coverage

```bash
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt
```

### Short Tests

```bash
go test -short ./...
```

### Verbose Output

```bash
go test -v ./...
```

### Specific Package

```bash
go test ./internal/backend/...
```

### Using Just

```bash
just test           # Run all tests
just test-verbose   # Verbose output
just test-short     # Short tests only
just test-coverage  # Generate coverage
just coverage-html  # View HTML report
```

## Writing Tests

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestParseOutput(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "simple output",
            input: "Hello, world!",
            want:  "Hello, world!",
        },
        {
            name:  "with whitespace",
            input: "  trimmed  ",
            want:  "trimmed",
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

            got, err := ParseOutput(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

### Test Organization

```
internal/backend/
├── claude.go
├── claude_test.go    # Tests alongside code
├── codex.go
├── codex_test.go
└── ...
```

### Parallel Tests

Use `t.Parallel()` for independent tests:

```go
func TestA(t *testing.T) {
    t.Parallel()
    // ...
}

func TestB(t *testing.T) {
    t.Parallel()
    // ...
}
```

## Mock Package

The `internal/mock` package provides testing utilities.

### Mock Backend

```go
import "github.com/signalridge/clinvoker/internal/mock"

func TestWithMockBackend(t *testing.T) {
    mb := mock.NewMockBackend("test",
        mock.WithParseOutput("mocked output"),
        mock.WithAvailable(true),
    )

    output := mb.ParseOutput("any input")
    if output != "mocked output" {
        t.Errorf("expected mocked output")
    }
}
```

### Mock Options

```go
// Set availability
mock.WithAvailable(true)
mock.WithAvailable(false)

// Set parsed output
mock.WithParseOutput("custom output")

// Set command behavior
mock.WithCommand(func(prompt string, opts *Options) *exec.Cmd {
    return exec.Command("echo", prompt)
})
```

### Temporary Directories

```go
func TestWithTempDir(t *testing.T) {
    dir := t.TempDir()

    // Use dir for test files
    configPath := filepath.Join(dir, "config.yaml")
    // ...
}
```

## Testing HTTP Handlers

### Using httptest

```go
import (
    "net/http/httptest"
    "testing"
)

func TestPromptHandler(t *testing.T) {
    // Create test server
    handler := NewPromptHandler(mockService)
    server := httptest.NewServer(handler)
    defer server.Close()

    // Make request
    resp, err := http.Post(
        server.URL+"/api/v1/prompt",
        "application/json",
        strings.NewReader(`{"backend":"claude","prompt":"test"}`),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    // Check response
    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected 200, got %d", resp.StatusCode)
    }
}
```

## Integration Tests

### Testing CLI Commands

```go
func TestCLIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Build binary
    cmd := exec.Command("go", "build", "-o", "clinvk-test", "./cmd/clinvk")
    if err := cmd.Run(); err != nil {
        t.Fatal(err)
    }
    defer os.Remove("clinvk-test")

    // Run command
    out, err := exec.Command("./clinvk-test", "version").Output()
    if err != nil {
        t.Fatal(err)
    }

    if !strings.Contains(string(out), "clinvk") {
        t.Error("expected version output")
    }
}
```

## Test Fixtures

Store test data in `testdata/`:

```
testdata/
├── config/
│   ├── valid.yaml
│   └── invalid.yaml
├── tasks/
│   └── sample.json
└── sessions/
    └── sample.json
```

Load fixtures:

```go
func loadFixture(t *testing.T, name string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", name))
    if err != nil {
        t.Fatal(err)
    }
    return data
}
```

## Benchmarks

```go
func BenchmarkParseOutput(b *testing.B) {
    input := "large output content..."

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ParseOutput(input)
    }
}
```

Run benchmarks:

```bash
go test -bench=. ./...
go test -bench=BenchmarkParseOutput -benchmem ./internal/output/
```

## Coverage Goals

- Aim for >80% coverage on core packages
- Focus on behavior, not line coverage
- Don't skip error paths

## CI Integration

Tests run automatically on:

- Pull requests
- Pushes to main
- Release tags

See `.github/workflows/ci.yaml` for configuration.

## Troubleshooting Tests

### Flaky Tests

- Check for race conditions
- Use deterministic inputs
- Mock external dependencies

### Slow Tests

- Use `t.Parallel()`
- Skip slow tests with `-short`
- Profile with `-cpuprofile`

### Debug Output

```go
t.Logf("debug: value = %v", value)
```

Run with verbose to see logs:

```bash
go test -v ./...
```
