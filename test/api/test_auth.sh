#!/usr/bin/env bash
# API Key authentication tests
# Tests the API key middleware with various authentication scenarios
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Server config for auth tests (uses different port to avoid conflicts)
AUTH_SERVER_PORT="${AUTH_SERVER_PORT:-18081}"
AUTH_SERVER_URL="http://${SERVER_HOST}:${AUTH_SERVER_PORT}"
AUTH_SERVER_PID=""
TEST_API_KEY="test-integration-key-12345"

# Start server with API key authentication enabled
start_auth_server() {
	log_info "Starting server with API key auth on port $AUTH_SERVER_PORT..."

	# Export API keys via environment variable
	CLINVK_API_KEYS="$TEST_API_KEY" "$CLINVK_BIN" serve \
		--host "$SERVER_HOST" \
		--port "$AUTH_SERVER_PORT" &
	AUTH_SERVER_PID=$!

	# Wait for server to be ready
	local retries=30
	while ((retries > 0)); do
		# Health endpoint should work without auth
		if curl -sf "${AUTH_SERVER_URL}/health" >/dev/null 2>&1; then
			log_success "Auth server started (PID: $AUTH_SERVER_PID)"
			return 0
		fi
		sleep 0.5
		((retries--))
	done

	log_error "Auth server failed to start"
	stop_auth_server
	return 1
}

# Stop auth server
stop_auth_server() {
	if [[ -n "$AUTH_SERVER_PID" ]]; then
		log_info "Stopping auth server (PID: $AUTH_SERVER_PID)..."
		kill "$AUTH_SERVER_PID" 2>/dev/null || true
		wait "$AUTH_SERVER_PID" 2>/dev/null || true
		AUTH_SERVER_PID=""
	fi
}

# HTTP helpers for auth server
auth_http_get() {
	local path="$1"
	shift
	curl -sf "${AUTH_SERVER_URL}${path}" "$@"
}

auth_http_get_status() {
	local path="$1"
	shift
	curl -s -w "\n%{http_code}" "${AUTH_SERVER_URL}${path}" "$@"
}

# Test: Health endpoint should work without authentication
test_health_no_auth_required() {
	local response
	response=$(auth_http_get "/health")

	local status
	status=$(echo "$response" | jq -r '.status')
	if [[ "$status" != "ok" && "$status" != "degraded" ]]; then
		log_error "Health check should work without auth, got status: $status"
		return 1
	fi
}

# Test: Protected endpoint requires authentication
test_protected_endpoint_requires_auth() {
	local output status_code
	output=$(auth_http_get_status "/api/v1/backends")
	status_code=$(echo "$output" | tail -1)

	if [[ "$status_code" != "401" ]]; then
		log_error "Expected 401 for unauthenticated request, got $status_code"
		return 1
	fi
}

# Test: Valid X-Api-Key header grants access
test_valid_x_api_key_header() {
	local response
	response=$(auth_http_get "/api/v1/backends" -H "X-Api-Key: $TEST_API_KEY")

	# Should get backends list
	assert_json_field "$response" "backends"
}

# Test: Valid Bearer token grants access
test_valid_bearer_token() {
	local response
	response=$(auth_http_get "/api/v1/backends" -H "Authorization: Bearer $TEST_API_KEY")

	# Should get backends list
	assert_json_field "$response" "backends"
}

# Test: Invalid API key returns 401
test_invalid_api_key() {
	local output status_code
	output=$(auth_http_get_status "/api/v1/backends" -H "X-Api-Key: invalid-key")
	status_code=$(echo "$output" | tail -1)

	if [[ "$status_code" != "401" ]]; then
		log_error "Expected 401 for invalid API key, got $status_code"
		return 1
	fi
}

# Test: Empty API key returns 401
test_empty_api_key() {
	local output status_code
	output=$(auth_http_get_status "/api/v1/backends" -H "X-Api-Key: ")
	status_code=$(echo "$output" | tail -1)

	if [[ "$status_code" != "401" ]]; then
		log_error "Expected 401 for empty API key, got $status_code"
		return 1
	fi
}

# Test: OpenAPI docs endpoint should work without auth
test_openapi_docs_no_auth() {
	local output status_code
	output=$(auth_http_get_status "/docs")
	status_code=$(echo "$output" | tail -1)

	# Docs endpoint should be accessible (200 or redirect)
	if [[ "$status_code" != "200" && "$status_code" != "301" && "$status_code" != "302" ]]; then
		log_error "Expected docs endpoint to be accessible without auth, got $status_code"
		return 1
	fi
}

# Main test execution
main() {
	setup_test_env

	# Add custom cleanup for auth server
	trap 'stop_auth_server; cleanup_test_env' EXIT INT TERM

	start_auth_server

	print_subheader "API Key Authentication"
	run_test "Health endpoint works without auth" test_health_no_auth_required
	run_test "Protected endpoint requires auth" test_protected_endpoint_requires_auth
	run_test "Valid X-Api-Key header grants access" test_valid_x_api_key_header
	run_test "Valid Bearer token grants access" test_valid_bearer_token
	run_test "Invalid API key returns 401" test_invalid_api_key
	run_test "Empty API key returns 401" test_empty_api_key
	run_test "OpenAPI docs accessible without auth" test_openapi_docs_no_auth

	print_summary
}

main "$@"
