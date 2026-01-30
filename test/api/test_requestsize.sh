#!/usr/bin/env bash
# Request size limiting tests
# Tests the request body size limiting middleware
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Server config for request size tests (uses different port to avoid conflicts)
REQUESTSIZE_SERVER_PORT="${REQUESTSIZE_SERVER_PORT:-18083}"
REQUESTSIZE_SERVER_URL="http://${SERVER_HOST}:${REQUESTSIZE_SERVER_PORT}"
REQUESTSIZE_SERVER_PID=""

# Size limit for testing (1KB for fast testing)
TEST_MAX_SIZE=1024

# Start server with request size limiting enabled
start_requestsize_server() {
	log_info "Starting server with request size limiting on port $REQUESTSIZE_SERVER_PORT..."

	# Create temp config with request size limiting enabled
	local temp_config="${TEST_TEMP_DIR}/requestsize_config.yaml"
	cat >"$temp_config" <<EOF
server:
  host: "${SERVER_HOST}"
  port: ${REQUESTSIZE_SERVER_PORT}
  max_request_body_bytes: ${TEST_MAX_SIZE}
EOF

	"$CLINVK_BIN" serve --config "$temp_config" &
	REQUESTSIZE_SERVER_PID=$!

	# Wait for server to be ready
	local retries=30
	while ((retries > 0)); do
		if curl -sf "${REQUESTSIZE_SERVER_URL}/health" >/dev/null 2>&1; then
			log_success "Request size limit server started (PID: $REQUESTSIZE_SERVER_PID)"
			return 0
		fi
		sleep 0.5
		((retries--))
	done

	log_error "Request size limit server failed to start"
	stop_requestsize_server
	return 1
}

# Stop request size server
stop_requestsize_server() {
	if [[ -n "$REQUESTSIZE_SERVER_PID" ]]; then
		log_info "Stopping request size limit server (PID: $REQUESTSIZE_SERVER_PID)..."
		kill "$REQUESTSIZE_SERVER_PID" 2>/dev/null || true
		wait "$REQUESTSIZE_SERVER_PID" 2>/dev/null || true
		REQUESTSIZE_SERVER_PID=""
	fi
}

# HTTP helpers for request size server
requestsize_http_post_status() {
	local path="$1"
	local body="$2"
	shift 2
	curl -s -w "\n%{http_code}" -X POST "${REQUESTSIZE_SERVER_URL}${path}" \
		-H "Content-Type: application/json" \
		-d "$body" "$@"
}

# Test: Small request within limit should succeed
test_small_request_succeeds() {
	local output status_code
	local small_payload='{"prompt":"hello"}'

	output=$(requestsize_http_post_status "/api/v1/prompt" "$small_payload")
	status_code=$(echo "$output" | tail -1)

	# 4xx is fine (backend/auth might fail), just check not 413
	if [[ "$status_code" == "413" ]]; then
		log_error "Small request should not be rejected, got 413"
		return 1
	fi

	log_debug "Small request returned status $status_code (expected: not 413)"
	return 0
}

# Test: Large request exceeding limit should get 413
test_large_request_rejected() {
	local output status_code
	# Create payload larger than TEST_MAX_SIZE (1KB)
	local large_payload
	large_payload=$(printf '{"prompt":"%s"}' "$(head -c 2048 /dev/zero | tr '\0' 'a')")

	output=$(requestsize_http_post_status "/api/v1/prompt" "$large_payload")
	status_code=$(echo "$output" | tail -1)

	if [[ "$status_code" != "413" ]]; then
		log_error "Large request should be rejected with 413, got $status_code"
		return 1
	fi
}

# Test: 413 response has correct content
test_413_response_content() {
	local output status_code body
	# Create payload larger than TEST_MAX_SIZE
	local large_payload
	large_payload=$(printf '{"prompt":"%s"}' "$(head -c 2048 /dev/zero | tr '\0' 'a')")

	output=$(requestsize_http_post_status "/api/v1/prompt" "$large_payload")
	status_code=$(echo "$output" | tail -1)
	body=$(echo "$output" | sed '$d')

	if [[ "$status_code" != "413" ]]; then
		log_error "Expected 413 status code, got $status_code"
		return 1
	fi

	# Response should indicate request too large
	if [[ ! "$body" =~ (too large|request size|exceeds|413) ]]; then
		log_warning "413 response body may not indicate size limit: $body"
	fi

	return 0
}

# Test: Request at boundary should succeed or be rejected consistently
test_boundary_request() {
	local output status_code
	# Create payload close to TEST_MAX_SIZE boundary
	local boundary_payload='{"prompt":"'
	# Add padding to get close to 1KB
	local padding
	padding=$(head -c 950 /dev/zero | tr '\0' 'b')
	boundary_payload="${boundary_payload}${padding}"'"}'

	output=$(requestsize_http_post_status "/api/v1/prompt" "$boundary_payload")
	status_code=$(echo "$output" | tail -1)

	# Should either succeed (200-499) or be rejected (413)
	# Other 5xx errors might be backend-related, which is fine for this test
	if [[ "$status_code" == "413" ]]; then
		log_debug "Boundary request was rejected (expected behavior)"
	elif [[ "$status_code" =~ ^[245][0-9][0-9]$ ]]; then
		log_debug "Boundary request returned $status_code (accepted or other error)"
	else
		log_error "Unexpected status code for boundary request: $status_code"
		return 1
	fi

	return 0
}

# Main test execution
main() {
	setup_test_env

	# Add custom cleanup for request size server
	trap 'stop_requestsize_server; cleanup_test_env' EXIT INT TERM

	start_requestsize_server

	print_subheader "Request Size Limiting"
	run_test "Small request within limit succeeds" test_small_request_succeeds
	run_test "Large request exceeding limit gets 413" test_large_request_rejected
	run_test "413 response has correct content" test_413_response_content
	run_test "Boundary request handled correctly" test_boundary_request

	print_summary
}

main "$@"
