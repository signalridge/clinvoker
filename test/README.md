# Integration Tests

Integration tests for clinvk CLI and HTTP API.

## Structure

```text
test/
├── lib/
│   └── common.sh       # Shared test utilities
├── cli/                # CLI integration tests
│   ├── test_version.sh
│   ├── test_config.sh
│   ├── test_prompt.sh
│   ├── test_resume.sh
│   ├── test_sessions.sh
│   ├── test_chain.sh
│   ├── test_compare.sh
│   └── test_parallel.sh
├── api/                # HTTP API tests
│   ├── test_health.sh
│   ├── test_backends.sh
│   ├── test_auth.sh       # API key authentication
│   ├── test_ratelimit.sh  # Rate limiting
│   ├── test_requestsize.sh # Request body size limiting
│   ├── test_tracing.sh    # Distributed tracing
│   ├── test_prompt.sh
│   ├── test_sessions.sh
│   ├── test_chain.sh
│   ├── test_compare.sh
│   ├── test_parallel.sh
│   ├── test_openai.sh
│   └── test_anthropic.sh
├── run_cli_tests.sh    # Run all CLI tests
├── run_api_tests.sh    # Run all API tests
└── run_all_tests.sh    # Run all tests
```

## Running Tests

### Prerequisites

- Built `clinvk` binary (run `just build` first)
- `jq` for JSON parsing (optional, tests degrade gracefully)
- `curl` for API tests

### Using Just

```bash
# Run all integration tests
just integration

# Run CLI tests only
just integration-cli

# Run API tests only
just integration-api

# Run specific test file
just integration-file cli/test_version
```

### Direct Execution

```bash
# Build first
go build -o bin/clinvk ./cmd/clinvk

# Run all tests
./test/run_all_tests.sh

# Run CLI tests
./test/run_cli_tests.sh

# Run API tests
./test/run_api_tests.sh

# Run single test file
./test/cli/test_version.sh
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `CLINVK_BIN` | `./bin/clinvk` | Path to clinvk binary |
| `SERVER_HOST` | `127.0.0.1` | API server host |
| `SERVER_PORT` | `18080` | API server port |
| `AUTH_SERVER_PORT` | `18081` | Auth test server port |
| `RATELIMIT_SERVER_PORT` | `18082` | Rate limit test server port |
| `TEST_TIMEOUT` | `60` | Test timeout in seconds |
| `DEBUG` | `0` | Enable debug output |

## Writing Tests

### Test Structure

Each test file should:

1. Source `lib/common.sh`
2. Use `run_test "name" command` for individual tests
3. Use `skip_test "name" "reason"` for skipped tests
4. Exit with appropriate code

### Example

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../lib/common.sh"

# Setup
setup_test_env

# Tests
run_test "version shows output" \
    "${CLINVK_BIN}" version

run_test "version contains clinvk" \
    "${CLINVK_BIN}" version | grep -q "clinvk"

# Summary
print_summary
exit_with_status
```

### Available Functions

From `lib/common.sh`:

- `run_test "name" command...` - Run a test case
- `skip_test "name" "reason"` - Skip a test
- `assert_equals "expected" "actual"` - Assert equality
- `assert_contains "haystack" "needle"` - Assert substring
- `assert_json_field "json" "field" "expected"` - Assert JSON field
- `setup_test_env` - Initialize test environment
- `start_server` / `stop_server` - Manage test server
- `api_get/api_post/api_delete` - HTTP helpers

## Test Coverage

### CLI Tests

| Test File | Coverage |
|-----------|----------|
| `test_version.sh` | Version command, help flag |
| `test_config.sh` | Config show command |
| `test_prompt.sh` | Basic prompt execution, dry-run, output formats |
| `test_sessions.sh` | Session list, show, delete commands |
| `test_resume.sh` | Resume session functionality |
| `test_chain.sh` | Chain execution, placeholders, flags |
| `test_compare.sh` | Compare across backends, output formats |
| `test_parallel.sh` | Parallel execution, JSON input, fail-fast |

### API Tests

| Test File | Coverage |
|-----------|----------|
| `test_health.sh` | Health endpoint, backend status |
| `test_backends.sh` | Backend listing |
| `test_auth.sh` | API key authentication |
| `test_ratelimit.sh` | Rate limiting, burst limits, Retry-After header |
| `test_requestsize.sh` | Request body size limiting, 413 responses |
| `test_tracing.sh` | W3C Trace Context propagation |
| `test_prompt.sh` | Prompt API endpoint |
| `test_sessions.sh` | Sessions API (list, get, delete) |
| `test_chain.sh` | Chain API endpoint |
| `test_compare.sh` | Compare API endpoint |
| `test_parallel.sh` | Parallel API endpoint |
| `test_openai.sh` | OpenAI-compatible endpoints |
| `test_anthropic.sh` | Anthropic-compatible endpoints |

## CI Integration

Tests run automatically on:

- Push to main
- Pull requests

See `.github/workflows/ci.yaml` for configuration.
