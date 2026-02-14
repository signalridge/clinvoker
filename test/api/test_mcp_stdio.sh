#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

test_stdio_basic_responses() {
	local req_ping req_tools req_unknown req_bad_version req_bad_json
	req_ping='{"jsonrpc":"2.0","id":1,"method":"ping"}'
	req_tools='{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
	req_unknown='{"jsonrpc":"2.0","id":3,"method":"unknown_method"}'
	req_bad_version='{"jsonrpc":"1.0","id":4,"method":"ping"}'
	req_bad_json='{bad json'

	local lines=()
	mapfile -t lines < <(printf '%s\n' "$req_ping" "$req_tools" "$req_unknown" "$req_bad_version" "$req_bad_json" | "$CLINVK_BIN" mcp --transport stdio 2>/dev/null)

	if ((${#lines[@]} < 5)); then
		log_error "expected at least 5 responses, got ${#lines[@]}"
		return 1
	fi

	local ping tools unknown bad_version parse_error
	ping=$(printf '%s\n' "${lines[@]}" | jq -c 'select(.id==1)')
	tools=$(printf '%s\n' "${lines[@]}" | jq -c 'select(.id==2)')
	unknown=$(printf '%s\n' "${lines[@]}" | jq -c 'select(.id==3)')
	bad_version=$(printf '%s\n' "${lines[@]}" | jq -c 'select(.id==4)')
	parse_error=$(printf '%s\n' "${lines[@]}" | jq -c 'select(.error.code==-32700)')

	assert_not_empty "$ping"
	assert_json_field "$ping" "result"

	assert_not_empty "$tools"
	if ! echo "$tools" | jq -e '.result.tools | length > 0' >/dev/null 2>&1; then
		log_error "expected tools/list to return tools"
		return 1
	fi

	assert_not_empty "$unknown"
	assert_json_equals "$unknown" "error.code" "-32601"

	assert_not_empty "$bad_version"
	assert_json_equals "$bad_version" "error.code" "-32600"

	assert_not_empty "$parse_error"
	assert_json_equals "$parse_error" "error.code" "-32700"
}

main() {
	setup_test_env
	mkdir -p "${TEST_TEMP_DIR}/home"
	export HOME="${TEST_TEMP_DIR}/home"

	print_subheader "MCP Stdio JSON-RPC"
	run_test "stdio responses" test_stdio_basic_responses

	print_summary
}

main "$@"
