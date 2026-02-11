#!/usr/bin/env bash
# MCP authentication and SSE auth integration tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

AUTH_SERVER_PORT="${AUTH_SERVER_PORT:-18081}"
MCP_AUTH_SERVER_PATH="${MCP_AUTH_SERVER_PATH:-/mcp}"
MCP_AUTH_SERVER_URL="http://${SERVER_HOST}:${AUTH_SERVER_PORT}"
MCP_AUTH_ENDPOINT_URL="${MCP_AUTH_SERVER_URL}${MCP_AUTH_SERVER_PATH}"
MCP_AUTH_SERVER_PID=""
TEST_API_KEY="test-integration-key-12345"

start_mcp_auth_server() {
	log_info "Starting MCP auth server on port ${AUTH_SERVER_PORT}..."

	CLINVK_API_KEYS="$TEST_API_KEY" "$CLINVK_BIN" mcp \
		--transport http \
		--host "$SERVER_HOST" \
		--port "$AUTH_SERVER_PORT" \
		--path "$MCP_AUTH_SERVER_PATH" &
	MCP_AUTH_SERVER_PID=$!

	local retries=30
	while ((retries > 0)); do
		if curl -sf "${MCP_AUTH_ENDPOINT_URL}" \
			-H "Content-Type: application/json" \
			-H "X-Api-Key: $TEST_API_KEY" \
			-d '{"jsonrpc":"2.0","id":1,"method":"ping"}' >/dev/null 2>&1; then
			log_success "MCP auth server started (PID: $MCP_AUTH_SERVER_PID)"
			return 0
		fi
		sleep 0.5
		((retries--))
	done

	log_error "MCP auth server failed to start"
	stop_mcp_auth_server
	return 1
}

stop_mcp_auth_server() {
	if [[ -n "$MCP_AUTH_SERVER_PID" ]]; then
		log_info "Stopping MCP auth server (PID: $MCP_AUTH_SERVER_PID)..."
		kill "$MCP_AUTH_SERVER_PID" 2>/dev/null || true
		wait "$MCP_AUTH_SERVER_PID" 2>/dev/null || true
		MCP_AUTH_SERVER_PID=""
	fi
}

mcp_auth_post_status() {
	local body="$1"
	shift
	curl -s -w "\n%{http_code}" -X POST "${MCP_AUTH_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-d "$body" "$@"
}

test_mcp_requires_api_key() {
	local output status_code
	output=$(mcp_auth_post_status '{"jsonrpc":"2.0","id":1,"method":"ping"}')
	status_code=$(echo "$output" | tail -1)
	assert_equals "401" "$status_code"
}

test_mcp_accepts_x_api_key() {
	local output status_code body
	output=$(mcp_auth_post_status '{"jsonrpc":"2.0","id":2,"method":"ping"}' -H "X-Api-Key: $TEST_API_KEY")
	status_code=$(echo "$output" | tail -1)
	body=$(echo "$output" | sed '$d')

	assert_equals "200" "$status_code"
	assert_json_equals "$body" "jsonrpc" "2.0"
	assert_json_field "$body" "result"
}

test_mcp_accepts_bearer_token() {
	local output status_code body
	output=$(mcp_auth_post_status '{"jsonrpc":"2.0","id":3,"method":"ping"}' -H "Authorization: Bearer $TEST_API_KEY")
	status_code=$(echo "$output" | tail -1)
	body=$(echo "$output" | sed '$d')

	assert_equals "200" "$status_code"
	assert_json_equals "$body" "jsonrpc" "2.0"
}

test_mcp_rejects_invalid_api_key() {
	local output status_code
	output=$(mcp_auth_post_status '{"jsonrpc":"2.0","id":4,"method":"ping"}' -H "X-Api-Key: invalid-key")
	status_code=$(echo "$output" | tail -1)
	assert_equals "401" "$status_code"
}

test_mcp_sse_requires_api_key() {
	local status
	status=$(curl -sN -o /dev/null -w "%{http_code}" -X POST "${MCP_AUTH_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-H "Accept: text/event-stream" \
		-d '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}')

	assert_equals "401" "$status"
}

test_mcp_sse_with_api_key_succeeds() {
	local headers_file body_file status
	headers_file="${TEST_TEMP_DIR}/mcp_auth_sse_headers.txt"
	body_file="${TEST_TEMP_DIR}/mcp_auth_sse_body.txt"

	status=$(curl -sN -D "$headers_file" -o "$body_file" -w "%{http_code}" -X POST "${MCP_AUTH_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-H "Accept: text/event-stream" \
		-H "X-Api-Key: $TEST_API_KEY" \
		-d '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}')

	assert_equals "200" "$status"
	assert_contains "$(tr -d '\r' <"$headers_file")" "Content-Type: text/event-stream"
	assert_contains "$(cat "$body_file")" "event: message"
}

main() {
	setup_test_env
	trap 'stop_mcp_auth_server; cleanup_test_env' EXIT INT TERM

	start_mcp_auth_server

	print_subheader "MCP Authentication"
	run_test "MCP endpoint requires API key" test_mcp_requires_api_key
	run_test "MCP accepts X-Api-Key header" test_mcp_accepts_x_api_key
	run_test "MCP accepts Bearer token" test_mcp_accepts_bearer_token
	run_test "MCP rejects invalid API key" test_mcp_rejects_invalid_api_key
	run_test "MCP SSE requires API key" test_mcp_sse_requires_api_key
	run_test "MCP SSE succeeds with API key" test_mcp_sse_with_api_key_succeeds

	print_summary
	if ((TESTS_FAILED > 0)); then
		exit 1
	fi
}

main "$@"
