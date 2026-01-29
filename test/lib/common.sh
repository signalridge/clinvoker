#!/usr/bin/env bash
# Common test utilities for clinvk integration tests
# shellcheck disable=SC2034

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export PROJECT_ROOT

# Binary path (built or installed)
CLINVK_BIN="${CLINVK_BIN:-${PROJECT_ROOT}/bin/clinvk}"
export CLINVK_BIN

# Server configuration
SERVER_HOST="${SERVER_HOST:-127.0.0.1}"
SERVER_PORT="${SERVER_PORT:-18080}"
SERVER_URL="http://${SERVER_HOST}:${SERVER_PORT}"
export SERVER_HOST SERVER_PORT SERVER_URL

# Test configuration
TEST_TIMEOUT="${TEST_TIMEOUT:-60}"
SIMPLE_PROMPT="${SIMPLE_PROMPT:-echo hello world}"
export TEST_TIMEOUT SIMPLE_PROMPT

# All backends to test
ALL_BACKENDS=("claude" "codex" "gemini")
export ALL_BACKENDS

# =============================================================================
# Colors and Formatting
# =============================================================================

if [[ -t 1 ]]; then
	RED='\033[0;31m'
	GREEN='\033[0;32m'
	YELLOW='\033[0;33m'
	BLUE='\033[0;34m'
	CYAN='\033[0;36m'
	BOLD='\033[1m'
	NC='\033[0m' # No Color
else
	RED=''
	GREEN=''
	YELLOW=''
	BLUE=''
	CYAN=''
	BOLD=''
	NC=''
fi

# =============================================================================
# Test Counters
# =============================================================================

TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0
CURRENT_TEST_NAME=""

# =============================================================================
# Output Functions
# =============================================================================

log_info() {
	echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
	echo -e "${GREEN}[PASS]${NC} $*"
}

log_error() {
	echo -e "${RED}[FAIL]${NC} $*"
}

log_warning() {
	echo -e "${YELLOW}[WARN]${NC} $*"
}

log_skip() {
	echo -e "${YELLOW}[SKIP]${NC} $*"
}

log_debug() {
	if [[ "${DEBUG:-}" == "1" ]]; then
		echo -e "${CYAN}[DEBUG]${NC} $*"
	fi
}

print_header() {
	echo ""
	echo -e "${BOLD}========================================${NC}"
	echo -e "${BOLD}$*${NC}"
	echo -e "${BOLD}========================================${NC}"
	echo ""
}

print_subheader() {
	echo ""
	echo -e "${CYAN}--- $* ---${NC}"
}

# =============================================================================
# Backend Utilities
# =============================================================================

# Check if a backend CLI is available
backend_available() {
	local backend="$1"
	command -v "$backend" &>/dev/null
}

# Get available backends
get_available_backends() {
	local available=()
	for backend in "${ALL_BACKENDS[@]}"; do
		if backend_available "$backend"; then
			available+=("$backend")
		fi
	done
	echo "${available[*]}"
}

# Check if at least one backend is available
any_backend_available() {
	for backend in "${ALL_BACKENDS[@]}"; do
		if backend_available "$backend"; then
			return 0
		fi
	done
	return 1
}

# Skip current test if required backends are missing.
# Relies on CURRENT_TEST_NAME set by run_test.
skip_if_missing_backends() {
	local missing=()
	local backend
	for backend in "$@"; do
		if ! backend_available "$backend"; then
			missing+=("$backend")
		fi
	done
	if (( ${#missing[@]} > 0 )); then
		skip_test "${CURRENT_TEST_NAME:-unnamed test}" "backends not available: ${missing[*]}"
		return 1
	fi
	return 0
}

# =============================================================================
# Binary Management
# =============================================================================

# Build the clinvk binary if not exists
ensure_binary() {
	if [[ ! -x "$CLINVK_BIN" ]]; then
		log_info "Building clinvk binary..."
		(cd "$PROJECT_ROOT" && go build -o bin/clinvk ./cmd/clinvk)
		if [[ ! -x "$CLINVK_BIN" ]]; then
			log_error "Failed to build clinvk binary"
			return 1
		fi
		log_success "Binary built: $CLINVK_BIN"
	fi
}

# Run clinvk command
clinvk() {
	"$CLINVK_BIN" "$@"
}

# =============================================================================
# Server Management
# =============================================================================

SERVER_PID=""

# Start the test server
start_server() {
	if server_running; then
		log_info "Server already running on $SERVER_URL"
		return 0
	fi

	log_info "Starting server on $SERVER_URL..."

	"$CLINVK_BIN" serve --host "$SERVER_HOST" --port "$SERVER_PORT" &
	SERVER_PID=$!

	# Wait for server to be ready
	local retries=30
	while ((retries > 0)); do
		if curl -sf "${SERVER_URL}/health" >/dev/null 2>&1; then
			log_success "Server started (PID: $SERVER_PID)"
			return 0
		fi
		sleep 0.5
		((retries--))
	done

	log_error "Server failed to start"
	stop_server
	return 1
}

# Stop the test server
stop_server() {
	if [[ -n "$SERVER_PID" ]]; then
		log_info "Stopping server (PID: $SERVER_PID)..."
		kill "$SERVER_PID" 2>/dev/null || true
		wait "$SERVER_PID" 2>/dev/null || true
		SERVER_PID=""
	fi
}

# Check if server is running
server_running() {
	curl -sf "${SERVER_URL}/health" >/dev/null 2>&1
}

# =============================================================================
# HTTP Request Helpers
# =============================================================================

# GET request
http_get() {
	local path="$1"
	shift
	curl -sf "${SERVER_URL}${path}" "$@"
}

# GET request that returns HTTP status code
# Usage: http_get_status "/path"
# Output: HTTP_STATUS_CODE on last line, response body before
http_get_status() {
	local path="$1"
	shift
	curl -s -w "\n%{http_code}" "${SERVER_URL}${path}" "$@"
}

# POST request with JSON body
http_post() {
	local path="$1"
	local body="$2"
	shift 2
	curl -sf -X POST "${SERVER_URL}${path}" \
		-H "Content-Type: application/json" \
		-d "$body" "$@"
}

# POST request that returns HTTP status code
# Usage: http_post_status "/path" '{"json":"body"}'
# Output: HTTP_STATUS_CODE on last line, response body before
http_post_status() {
	local path="$1"
	local body="$2"
	shift 2
	curl -s -w "\n%{http_code}" -X POST "${SERVER_URL}${path}" \
		-H "Content-Type: application/json" \
		-d "$body" "$@"
}

# DELETE request
http_delete() {
	local path="$1"
	shift
	curl -sf -X DELETE "${SERVER_URL}${path}" "$@"
}

# POST request for streaming (SSE)
http_post_stream() {
	local path="$1"
	local body="$2"
	shift 2
	curl -sN -X POST "${SERVER_URL}${path}" \
		-H "Content-Type: application/json" \
		-H "Accept: text/event-stream" \
		-d "$body" "$@"
}

# =============================================================================
# Test Framework
# =============================================================================

# Setup test environment
setup_test_env() {
	ensure_binary

	# Create temp directory for test artifacts
	TEST_TEMP_DIR="$(mktemp -d)"
	export TEST_TEMP_DIR

	# Trap for cleanup
	trap cleanup_test_env EXIT INT TERM
}

# Cleanup test environment
cleanup_test_env() {
	stop_server
	if [[ -n "${TEST_TEMP_DIR:-}" && -d "$TEST_TEMP_DIR" ]]; then
		rm -rf "$TEST_TEMP_DIR"
	fi
}

# Run a test function
# Always returns 0 to not exit script on test failure (set -e compatible)
run_test() {
	local test_name="$1"
	local test_func="$2"
	shift 2

	CURRENT_TEST_NAME="$test_name"
	log_info "Running: $test_name"

	# Temporarily disable errexit to capture test result
	local result=0
	set +e
	"$test_func" "$@"
	result=$?
	set -e

	if ((result == 0)); then
		((TESTS_PASSED++)) || true
		log_success "$test_name"
	else
		((TESTS_FAILED++)) || true
		log_error "$test_name"
	fi
	return 0
}

# Skip a test
skip_test() {
	local test_name="$1"
	local reason="${2:-No reason provided}"

	((TESTS_SKIPPED++)) || true
	log_skip "$test_name: $reason"
}

# Print test summary
print_summary() {
	echo ""
	echo -e "${BOLD}========================================${NC}"
	echo -e "${BOLD}Test Summary${NC}"
	echo -e "${BOLD}========================================${NC}"
	echo -e "${GREEN}Passed:${NC}  $TESTS_PASSED"
	echo -e "${RED}Failed:${NC}  $TESTS_FAILED"
	echo -e "${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
	echo -e "Total:   $((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))"
	echo ""

	if ((TESTS_FAILED > 0)); then
		return 1
	fi
	return 0
}

# =============================================================================
# Assertion Functions
# =============================================================================

# Assert command succeeds
assert_success() {
	local cmd="$*"
	if ! eval "$cmd"; then
		log_error "Command failed: $cmd"
		return 1
	fi
}

# Assert command fails
assert_failure() {
	local cmd="$*"
	if eval "$cmd" 2>/dev/null; then
		log_error "Command should have failed: $cmd"
		return 1
	fi
}

# Assert string contains substring
assert_contains() {
	local haystack="$1"
	local needle="$2"
	if [[ "$haystack" != *"$needle"* ]]; then
		log_error "Expected to contain '$needle' but got: $haystack"
		return 1
	fi
}

# Assert string does not contain substring
assert_not_contains() {
	local haystack="$1"
	local needle="$2"
	if [[ "$haystack" == *"$needle"* ]]; then
		log_error "Expected not to contain '$needle' but got: $haystack"
		return 1
	fi
}

# Assert strings are equal
assert_equals() {
	local expected="$1"
	local actual="$2"
	if [[ "$expected" != "$actual" ]]; then
		log_error "Expected '$expected' but got '$actual'"
		return 1
	fi
}

# Assert string is not empty
assert_not_empty() {
	local value="$1"
	if [[ -z "$value" ]]; then
		log_error "Expected non-empty value but got empty string"
		return 1
	fi
}

# Assert JSON field exists
assert_json_field() {
	local json="$1"
	local field="$2"
	if ! echo "$json" | jq -e ".$field" >/dev/null 2>&1; then
		log_error "JSON field '$field' not found in: $json"
		return 1
	fi
}

# Assert JSON field equals value
assert_json_equals() {
	local json="$1"
	local field="$2"
	local expected="$3"
	local actual
	actual=$(echo "$json" | jq -r ".$field")
	if [[ "$actual" != "$expected" ]]; then
		log_error "JSON field '$field': expected '$expected' but got '$actual'"
		return 1
	fi
}

# Assert exit code
assert_exit_code() {
	local expected="$1"
	local actual="$2"
	if [[ "$expected" != "$actual" ]]; then
		log_error "Expected exit code $expected but got $actual"
		return 1
	fi
}

# Assert HTTP status code
assert_http_status() {
	local expected="$1"
	local url="$2"
	local actual
	actual=$(curl -sf -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
	if [[ "$expected" != "$actual" ]]; then
		log_error "Expected HTTP status $expected but got $actual for $url"
		return 1
	fi
}

# =============================================================================
# Utility Functions
# =============================================================================

# Generate a unique test ID
generate_test_id() {
	local seed
	seed="$(date +%s%N 2>/dev/null || date +%s)"
	seed="${seed}-${$}-${RANDOM}"

	if command -v sha256sum &>/dev/null; then
		echo "$seed" | sha256sum | awk '{print substr($1,1,8)}'
		return 0
	fi
	if command -v shasum &>/dev/null; then
		echo "$seed" | shasum -a 256 | awk '{print substr($1,1,8)}'
		return 0
	fi
	if command -v md5 &>/dev/null; then
		echo "$seed" | md5 -q | awk '{print substr($1,1,8)}'
		return 0
	fi
	echo "$seed" | tr -cd '0-9a-f' | head -c 8
}

# Create a temporary JSON file
create_temp_json() {
	local content="$1"
	local temp_file="${TEST_TEMP_DIR}/$(generate_test_id).json"
	echo "$content" >"$temp_file"
	echo "$temp_file"
}

# Wait for a condition with timeout
wait_for() {
	local condition="$1"
	local timeout="${2:-30}"
	local interval="${3:-1}"

	local elapsed=0
	while ! eval "$condition" 2>/dev/null; do
		if ((elapsed >= timeout)); then
			return 1
		fi
		sleep "$interval"
		((elapsed += interval))
	done
	return 0
}

# Run command with timeout
run_with_timeout() {
	local timeout="$1"
	shift
	timeout "$timeout" "$@"
}

# Extract session ID from output
extract_session_id() {
	local output="$1"
	echo "$output" | sed -nE 's/.*session[_-]?id["[:space:]:]*([A-Za-z0-9-]+).*/\1/p' | head -1
}

# Check if jq is available
require_jq() {
	if ! command -v jq &>/dev/null; then
		log_error "jq is required but not installed"
		exit 1
	fi
}

# =============================================================================
# Initialization
# =============================================================================

require_jq
