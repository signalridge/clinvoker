# Clinvk Justfile
# Run `just --list` to see all available commands

# Default recipe
default:
    @just --list

# ==================== Build ====================

# Build the binary
[group('build')]
build:
    go build -o clinvk ./cmd/clinvk

# Build with version info
[group('build')]
build-release version="dev" commit="unknown":
    go build -ldflags "-s -w \
        -X github.com/signalridge/clinvoker/internal/app.version={{version}} \
        -X github.com/signalridge/clinvoker/internal/app.commit={{commit}} \
        -X github.com/signalridge/clinvoker/internal/app.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o clinvk ./cmd/clinvk

# Clean build artifacts
[group('build')]
clean:
    rm -f clinvk
    rm -rf dist/
    rm -f coverage.txt coverage.html

# ==================== Test ====================

# Run all tests
[group('test')]
test:
    go test -race -coverprofile=coverage.txt ./...

# Run tests with verbose output
[group('test')]
test-verbose:
    go test -v -race ./...

# Run short tests only
[group('test')]
test-short:
    go test -short -race ./...

# Run tests and show coverage
[group('test')]
test-coverage:
    go test -race -coverprofile=coverage.txt ./...
    go tool cover -func=coverage.txt

# Generate HTML coverage report
[group('test')]
coverage-html: test
    go tool cover -html=coverage.txt -o coverage.html
    @echo "Coverage report: coverage.html"

# Run benchmarks
[group('test')]
bench:
    go test -bench=. -benchmem ./...

# ==================== Lint ====================

# Run all linters
[group('lint')]
lint: lint-go lint-yaml lint-nix

# Run Go linter
[group('lint')]
lint-go:
    golangci-lint run --timeout 5m

# Run YAML linter
[group('lint')]
lint-yaml:
    yamllint -c .yamllint.yaml .

# Run Nix linter/formatter check
[group('lint')]
lint-nix:
    nixfmt --check *.nix || true

# Fix auto-fixable lint issues
[group('lint')]
lint-fix:
    golangci-lint run --fix --timeout 5m

# ==================== Format ====================

# Format all code
[group('format')]
fmt: fmt-go fmt-nix

# Format Go code
[group('format')]
fmt-go:
    gofmt -w .
    goimports -w .

# Format Nix files
[group('format')]
fmt-nix:
    nixfmt *.nix || true

# ==================== Dev ====================

# Build and run with --help
[group('dev')]
dev: build
    ./clinvk --help

# Watch and run tests on changes
[group('dev')]
watch:
    watchexec -e go 'just test-short'

# Run the CLI with arguments
[group('dev')]
run *args: build
    ./clinvk {{args}}

# Start the HTTP server
[group('dev')]
serve port="8080": build
    ./clinvk serve --port {{port}}

# ==================== Release ====================

# Dry run release
[group('release')]
release-dry:
    goreleaser release --snapshot --clean

# Check if ready for release
[group('release')]
release-check:
    goreleaser check

# ==================== Docker ====================

# Build Docker image
[group('docker')]
docker-build version="dev":
    docker build \
        --build-arg VERSION={{version}} \
        --build-arg COMMIT=$(git rev-parse --short HEAD) \
        --build-arg DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
        -t clinvk:{{version}} .

# Run Docker container
[group('docker')]
docker-run version="dev" *args:
    docker run --rm clinvk:{{version}} {{args}}

# ==================== Dependencies ====================

# Download dependencies
[group('deps')]
deps:
    go mod download

# Tidy dependencies
[group('deps')]
deps-tidy:
    go mod tidy

# Update all dependencies
[group('deps')]
deps-update:
    go get -u ./...
    go mod tidy

# Check for vulnerabilities
[group('deps')]
vuln:
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...

# ==================== Setup ====================

# Install development tools
[group('setup')]
setup:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest
    pre-commit install
    pre-commit install --hook-type commit-msg
    @echo "Development environment ready!"

# ==================== CI ====================

# Run CI checks locally
[group('ci')]
ci: deps-tidy lint test build
    @echo "All CI checks passed!"

# Run security checks
[group('ci')]
security: vuln
    @echo "Security checks passed!"
