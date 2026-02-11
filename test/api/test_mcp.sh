#!/usr/bin/env bash
# MCP HTTP endpoint integration tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

test_mcp_ping() {
	local response
	response=$(curl -sf -X POST "${MCP_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"ping"}')

	assert_json_equals "$response" "jsonrpc" "2.0"
	assert_json_field "$response" "result"
}

test_mcp_initialize() {
	local response
	response=$(curl -sf -X POST "${MCP_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":2,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}')

	assert_json_equals "$response" "jsonrpc" "2.0"
	assert_json_equals "$response" "result.protocolVersion" "2024-11-05"
	assert_json_equals "$response" "result.serverInfo.name" "clinvoker"
}

test_mcp_initialize_unsupported_protocol() {
	local response
	response=$(curl -sf -X POST "${MCP_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":3,"method":"initialize","params":{"protocolVersion":"9999-01-01"}}')

	assert_json_equals "$response" "jsonrpc" "2.0"
	assert_json_equals "$response" "error.code" "-32602"

	local message
	message=$(echo "$response" | jq -r '.error.message')
	assert_contains "$message" "unsupported protocol version"
}

test_mcp_tools_list() {
	local response
	response=$(curl -sf -X POST "${MCP_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":4,"method":"tools/list"}')

	assert_json_equals "$response" "jsonrpc" "2.0"
	assert_json_field "$response" "result.tools"
}

test_mcp_tools_call_unknown_tool() {
	local response
	response=$(curl -sf -X POST "${MCP_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}')

	assert_json_equals "$response" "jsonrpc" "2.0"
	assert_json_equals "$response" "error.code" "-32001"

	local message
	message=$(echo "$response" | jq -r '.error.message')
	assert_contains "$message" "unknown tool"
}

test_mcp_tools_call_notification_no_content() {
	local body_file status
	body_file="${TEST_TEMP_DIR}/mcp_notification_body.txt"
	status=$(curl -s -o "$body_file" -w "%{http_code}" -X POST "${MCP_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"clinvk_backends","arguments":{}}}')

	assert_equals "204" "$status"
	assert_equals "0" "$(wc -c <"$body_file" | tr -d ' ')"
}

test_mcp_stream_request_requires_sse_accept() {
	local response
	response=$(curl -sf -X POST "${MCP_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}')

	assert_json_equals "$response" "jsonrpc" "2.0"
	assert_json_equals "$response" "error.code" "-32602"

	local message
	message=$(echo "$response" | jq -r '.error.message')
	assert_contains "$message" "streaming requires Accept: text/event-stream"
}

test_mcp_stream_request_with_sse() {
	local headers_file body_file status
	headers_file="${TEST_TEMP_DIR}/mcp_sse_headers.txt"
	body_file="${TEST_TEMP_DIR}/mcp_sse_body.txt"

	status=$(curl -sN -D "$headers_file" -o "$body_file" -w "%{http_code}" -X POST "${MCP_ENDPOINT_URL}" \
		-H "Content-Type: application/json" \
		-H "Accept: text/event-stream" \
		-d '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"invalid-backend","prompt":"hello","output_format":"stream-json"}}}')

	assert_equals "200" "$status"
	assert_contains "$(tr -d '\r' <"$headers_file")" "Content-Type: text/event-stream"
	assert_contains "$(cat "$body_file")" "event: message"
}

main() {
	setup_test_env
	start_mcp_http_server

	print_subheader "MCP API"
	run_test "MCP ping returns JSON-RPC result" test_mcp_ping
	run_test "MCP initialize accepts supported protocol" test_mcp_initialize
	run_test "MCP initialize rejects unsupported protocol" test_mcp_initialize_unsupported_protocol
	run_test "MCP tools/list returns tools array" test_mcp_tools_list
	run_test "MCP unknown tool returns CodeToolNotFound" test_mcp_tools_call_unknown_tool
	run_test "MCP tools/call notification returns 204 with empty body" test_mcp_tools_call_notification_no_content
	run_test "MCP stream request requires SSE Accept header" test_mcp_stream_request_requires_sse_accept
	run_test "MCP stream request returns SSE when Accept is set" test_mcp_stream_request_with_sse

	print_summary
	if ((TESTS_FAILED > 0)); then
		exit 1
	fi
}

main "$@"
