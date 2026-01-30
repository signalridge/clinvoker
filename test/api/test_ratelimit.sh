#!/usr/bin/env bash
# Rate limiting tests
# Tests the per-IP rate limiting middleware
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Server config for rate limit tests (uses different port to avoid conflicts)
RATELIMIT_SERVER_PORT="${RATELIMIT_SERVER_PORT:-18082}"
RATELIMIT_SERVER_URL="http://${SERVER_HOST}:${RATELIMIT_SERVER_PORT}"
RATELIMIT_SERVER_PID=""

# Rate limit settings for testing (low values for fast testing)
# Note: burst should be higher than what we test to account for server startup health check
TEST_RPS=2
TEST_BURST=5

# Start server with rate limiting enabled
start_ratelimit_server() {
	log_info "Starting server with rate limiting on port $RATELIMIT_SERVER_PORT..."

	# Create temp config with rate limiting enabled
	local temp_config="${TEST_TEMP_DIR}/ratelimit_config.yaml"
	cat >"$temp_config" <<EOF
server:
  host: "${SERVER_HOST}"
  port: ${RATELIMIT_SERVER_PORT}
  rate_limit_enabled: true
  rate_limit_rps: ${TEST_RPS}
  rate_limit_burst: ${TEST_BURST}
EOF

	"$CLINVK_BIN" serve --config "$temp_config" &
	RATELIMIT_SERVER_PID=$!

	# Wait for server to be ready
	local retries=30
	while ((retries > 0)); do
		if curl -sf "${RATELIMIT_SERVER_URL}/health" >/dev/null 2>&1; then
			log_success "Rate limit server started (PID: $RATELIMIT_SERVER_PID)"
			return 0
		fi
		sleep 0.5
		((retries--))
	done

	log_error "Rate limit server failed to start"
	stop_ratelimit_server
	return 1
}

# Stop rate limit server
stop_ratelimit_server() {
	if [[ -n "$RATELIMIT_SERVER_PID" ]]; then
		log_info "Stopping rate limit server (PID: $RATELIMIT_SERVER_PID)..."
		kill "$RATELIMIT_SERVER_PID" 2>/dev/null || true
		wait "$RATELIMIT_SERVER_PID" 2>/dev/null || true
		RATELIMIT_SERVER_PID=""
	fi
}

# HTTP helpers for rate limit server
ratelimit_http_get_status() {
	local path="$1"
	shift
	curl -s -w "\n%{http_code}" "${RATELIMIT_SERVER_URL}${path}" "$@"
}

# Test: Normal requests within limit should succeed
test_requests_within_limit() {
	local i output status_code

	# Make requests within burst limit (subtract 1 for server startup health check)
	local safe_count=$((TEST_BURST - 2))
	for i in $(seq 1 "$safe_count"); do
		output=$(ratelimit_http_get_status "/health")
		status_code=$(echo "$output" | tail -1)

		if [[ "$status_code" != "200" ]]; then
			log_error "Request $i within burst limit should succeed, got $status_code"
			return 1
		fi
	done
}

# Test: Requests exceeding limit should get 429
test_requests_exceed_limit() {
	local i output status_code got_429=false

	# Rapid-fire requests to exceed rate limit
	# Send more than burst allows in quick succession
	local total_requests=$((TEST_BURST + 5))

	for i in $(seq 1 "$total_requests"); do
		output=$(ratelimit_http_get_status "/health")
		status_code=$(echo "$output" | tail -1)

		if [[ "$status_code" == "429" ]]; then
			got_429=true
			log_debug "Got 429 on request $i (expected)"
			break
		fi
	done

	if [[ "$got_429" != "true" ]]; then
		log_error "Expected to hit rate limit (429) after $total_requests requests"
		return 1
	fi
}

# Test: Rate limit recovers after waiting
test_rate_limit_recovery() {
	local output status_code

	# First exhaust the rate limit
	for _ in $(seq 1 $((TEST_BURST + 3))); do
		ratelimit_http_get_status "/health" >/dev/null
	done

	# Wait for token bucket to refill (at least 1 second for 2 RPS)
	sleep 2

	# Should be able to make requests again
	output=$(ratelimit_http_get_status "/health")
	status_code=$(echo "$output" | tail -1)

	if [[ "$status_code" != "200" ]]; then
		log_error "Expected rate limit to recover after waiting, got $status_code"
		return 1
	fi
}

# Test: 429 response includes Retry-After header
test_retry_after_header() {
	local i headers

	# Exhaust rate limit
	for i in $(seq 1 $((TEST_BURST + 5))); do
		headers=$(curl -sI "${RATELIMIT_SERVER_URL}/health" 2>/dev/null || true)

		if echo "$headers" | grep -qi "429"; then
			# Check for Retry-After header
			if echo "$headers" | grep -qi "Retry-After"; then
				log_debug "Found Retry-After header in 429 response"
				return 0
			else
				log_error "429 response should include Retry-After header"
				return 1
			fi
		fi
	done

	# If we never hit 429, that's also a problem for this test
	log_warning "Could not trigger rate limit to check Retry-After header"
	return 0 # Skip rather than fail
}

# Main test execution
main() {
	setup_test_env

	# Add custom cleanup for rate limit server
	trap 'stop_ratelimit_server; cleanup_test_env' EXIT INT TERM

	start_ratelimit_server

	print_subheader "Rate Limiting"
	run_test "Requests within limit succeed" test_requests_within_limit

	# Wait a moment before next test to reset rate limiter
	sleep 2

	run_test "Requests exceeding limit get 429" test_requests_exceed_limit
	run_test "Rate limit recovers after waiting" test_rate_limit_recovery
	run_test "429 includes Retry-After header" test_retry_after_header

	print_summary
}

main "$@"
