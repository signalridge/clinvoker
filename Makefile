.PHONY: all build test lint lint-go lint-yaml lint-all fmt vet clean install dev release snapshot help

# Variables
BINARY_NAME := clinvoker
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/signalridge/clinvoker/internal/app.version=$(VERSION) \
	-X github.com/signalridge/clinvoker/internal/app.commit=$(COMMIT) \
	-X github.com/signalridge/clinvoker/internal/app.date=$(DATE)

# Default target
all: lint-all test build

# Build binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/clinvoker

# Run tests
test:
	@echo "Running tests..."
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run tests with verbose output
test-v:
	@echo "Running tests (verbose)..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run short tests (for pre-commit)
test-short:
	@echo "Running short tests..."
	go test -race -short ./...

# Run Go linter
lint-go:
	@echo "Running Go linter..."
	golangci-lint run --timeout=5m

# Run YAML linter
lint-yaml:
	@echo "Running YAML linter..."
	yamllint -c .yamllint.yaml .

# Run JSON linter (check syntax)
lint-json:
	@echo "Checking JSON files..."
	@find . -name "*.json" -not -path "./vendor/*" -not -path "./.git/*" | xargs -I {} sh -c 'python3 -m json.tool {} > /dev/null && echo "✓ {}" || echo "✗ {}"'

# Run all linters
lint-all: lint-go lint-yaml
	@echo "All linters passed!"

# Alias for lint-go (backward compatibility)
lint: lint-go

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.txt
	rm -rf dist/

# Install binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	go install -ldflags "$(LDFLAGS)" ./cmd/clinvoker

# Development build (no optimizations)
dev:
	@echo "Building for development..."
	go build -o $(BINARY_NAME) ./cmd/clinvoker

# Create release with goreleaser
release:
	@echo "Creating release..."
	goreleaser release --clean

# Create snapshot (for testing release)
snapshot:
	@echo "Creating snapshot..."
	goreleaser release --snapshot --clean

# Run pre-commit hooks
pre-commit:
	@echo "Running pre-commit..."
	pre-commit run --all-files

# Setup development environment
setup:
	@echo "Setting up development environment..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	pip install --user yamllint || pip3 install --user yamllint
	@command -v pre-commit >/dev/null 2>&1 && pre-commit install || echo "pre-commit not installed, skipping hook installation"

# Show coverage in browser
coverage: test
	@echo "Opening coverage report..."
	go tool cover -html=coverage.txt

# Cross-compile for all platforms
cross:
	@echo "Cross-compiling..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/clinvoker
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/clinvoker
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/clinvoker
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/clinvoker
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/clinvoker

# Check Nix flake
nix-check:
	@echo "Checking Nix flake..."
	nix flake check

# Build with Nix
nix-build:
	@echo "Building with Nix..."
	nix build .#clinvoker

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Run all linters, tests, and build (default)"
	@echo "  build      - Build binary with version info"
	@echo "  test       - Run tests with coverage"
	@echo "  test-v     - Run tests with verbose output"
	@echo "  test-short - Run short tests (for pre-commit)"
	@echo "  lint       - Run Go linter (alias for lint-go)"
	@echo "  lint-go    - Run golangci-lint"
	@echo "  lint-yaml  - Run yamllint"
	@echo "  lint-json  - Check JSON syntax"
	@echo "  lint-all   - Run all linters"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  clean      - Remove build artifacts"
	@echo "  install    - Install binary to GOPATH"
	@echo "  dev        - Build for development (fast)"
	@echo "  release    - Create release with goreleaser"
	@echo "  snapshot   - Create snapshot release"
	@echo "  pre-commit - Run pre-commit hooks"
	@echo "  setup      - Setup development environment"
	@echo "  coverage   - Generate and view coverage report"
	@echo "  cross      - Cross-compile for all platforms"
	@echo "  nix-check  - Check Nix flake"
	@echo "  nix-build  - Build with Nix"
	@echo "  help       - Show this help"
