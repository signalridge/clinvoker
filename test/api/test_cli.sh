#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

PRIMARY_BACKEND=""

select_primary_backend() {
	local available
	available=$(get_available_backends)
	if [[ -n "$available" ]]; then
		PRIMARY_BACKEND="${available%% *}"
	fi
}

require_backend_or_skip() {
	if [[ -z "$PRIMARY_BACKEND" ]]; then
		skip_test "$CURRENT_TEST_NAME" "no backend available"
		return 1
	fi
	return 0
}

test_prompt_dry_run() {
	require_backend_or_skip || return 0
	local output
	output=$(clinvk --backend "$PRIMARY_BACKEND" --dry-run "hello" 2>&1)
	assert_contains "$output" "Would execute:"
}

test_chain_dry_run() {
	require_backend_or_skip || return 0
	local chain_file
	chain_file="${TEST_TEMP_DIR}/chain.json"
	cat >"$chain_file" <<JSON
{"steps":[{"backend":"${PRIMARY_BACKEND}","prompt":"one"},{"backend":"${PRIMARY_BACKEND}","prompt":"two"}]}
JSON
	local output
	output=$(clinvk --dry-run chain --file "$chain_file" 2>&1)
	assert_contains "$output" "Would execute:"
}

test_parallel_dry_run() {
	require_backend_or_skip || return 0
	local tasks_file
	tasks_file="${TEST_TEMP_DIR}/tasks.json"
	cat >"$tasks_file" <<JSON
{"tasks":[{"backend":"${PRIMARY_BACKEND}","prompt":"hello"}]}
JSON
	local output
	output=$(clinvk --dry-run parallel --file "$tasks_file" --json 2>/dev/null)
	assert_json_equals "$output" "total_tasks" "1"
}

test_compare_dry_run() {
	require_backend_or_skip || return 0
	local output
	output=$(clinvk --dry-run compare "hello" --backends "$PRIMARY_BACKEND" 2>&1)
	assert_contains "$output" "Would execute:"
}

test_invalid_prompt_output_format() {
	local output status
	set +e
	output=$("$CLINVK_BIN" --output-format xml "hello" 2>&1)
	status=$?
	set -e
	if ((status == 0)); then
		log_error "expected failure for invalid output format"
		return 1
	fi
	assert_contains "$output" "invalid output format"
}

test_invalid_compare_missing_backends() {
	local output status
	set +e
	output=$("$CLINVK_BIN" compare "hello" 2>&1)
	status=$?
	set -e
	if ((status == 0)); then
		log_error "expected failure for missing backends"
		return 1
	fi
	assert_contains "$output" "specify backends"
}

test_invalid_chain_no_steps() {
	local chain_file output status
	chain_file="${TEST_TEMP_DIR}/chain-empty.json"
	cat >"$chain_file" <<JSON
{"steps":[]}
JSON
	set +e
	output=$("$CLINVK_BIN" chain --file "$chain_file" 2>&1)
	status=$?
	set -e
	if ((status == 0)); then
		log_error "expected failure for empty chain"
		return 1
	fi
	assert_contains "$output" "no steps defined"
}

test_invalid_parallel_no_tasks() {
	local tasks_file output status
	tasks_file="${TEST_TEMP_DIR}/tasks-empty.json"
	cat >"$tasks_file" <<JSON
{"tasks":[]}
JSON
	set +e
	output=$("$CLINVK_BIN" parallel --file "$tasks_file" 2>&1)
	status=$?
	set -e
	if ((status == 0)); then
		log_error "expected failure for empty tasks"
		return 1
	fi
	assert_contains "$output" "no tasks provided"
}

test_mcp_invalid_transport() {
	local output status
	set +e
	output=$("$CLINVK_BIN" mcp --transport nope 2>&1)
	status=$?
	set -e
	if ((status == 0)); then
		log_error "expected failure for invalid transport"
		return 1
	fi
	assert_contains "$output" "unsupported transport"
}

main() {
	setup_test_env
	mkdir -p "${TEST_TEMP_DIR}/home"
	export HOME="${TEST_TEMP_DIR}/home"
	select_primary_backend

	print_subheader "CLI Dry Run Success Paths"
	run_test "Prompt dry run" test_prompt_dry_run
	run_test "Chain dry run" test_chain_dry_run
	run_test "Parallel dry run" test_parallel_dry_run
	run_test "Compare dry run" test_compare_dry_run

	print_subheader "CLI Invalid Input Errors"
	run_test "Invalid output format" test_invalid_prompt_output_format
	run_test "Compare missing backends" test_invalid_compare_missing_backends
	run_test "Chain empty steps" test_invalid_chain_no_steps
	run_test "Parallel empty tasks" test_invalid_parallel_no_tasks
	run_test "MCP invalid transport" test_mcp_invalid_transport

	print_summary
}

main "$@"
