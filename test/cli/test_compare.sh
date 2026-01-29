#!/usr/bin/env bash
# Test clinvk compare command

set -euo pipefail

# Source common test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# =============================================================================
# Test Functions
# =============================================================================

test_compare_two_backends() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Compare with two backends" "Less than 2 backends available"
		return 0
	fi

	local backend1="${available_backends[0]}"
	local backend2="${available_backends[1]}"

	local output
	output=$(clinvk compare --backends "$backend1,$backend2" --dry-run "$SIMPLE_PROMPT" 2>&1 || true)

	assert_not_empty "$output"
	assert_contains "$output" "$backend1"
	assert_contains "$output" "$backend2"
}

test_compare_all_backends() {
	if ! any_backend_available; then
		skip_test "Compare with --all-backends" "No backends available"
		return 0
	fi

	local output
	output=$(clinvk compare --all-backends --dry-run "$SIMPLE_PROMPT" 2>&1 || true)

	assert_not_empty "$output"
}

test_compare_sequential_flag() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Compare with --sequential" "Less than 2 backends available"
		return 0
	fi

	local backend1="${available_backends[0]}"
	local backend2="${available_backends[1]}"

	local output
	output=$(clinvk compare --backends "$backend1,$backend2" --sequential --dry-run "$SIMPLE_PROMPT" 2>&1 || true)

	assert_not_empty "$output"
}

test_compare_json_output() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Compare with --json" "Less than 2 backends available"
		return 0
	fi

	local backend1="${available_backends[0]}"
	local backend2="${available_backends[1]}"

	local output
	output=$(clinvk compare --backends "$backend1,$backend2" --json --dry-run "$SIMPLE_PROMPT" 2>&1 || true)

	assert_not_empty "$output"

	# Check if output contains JSON structure
	if ! echo "$output" | jq empty 2>/dev/null; then
		log_warning "Output is not valid JSON"
	fi
}

test_compare_exit_code() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Compare exit code validation" "Less than 2 backends available"
		return 0
	fi

	local backend1="${available_backends[0]}"
	local backend2="${available_backends[1]}"

	local exit_code=0
	clinvk compare --backends "$backend1,$backend2" --dry-run "$SIMPLE_PROMPT" >/dev/null 2>&1 || exit_code=$?

	# Should succeed or fail gracefully
	if [[ $exit_code -gt 1 ]]; then
		log_error "Unexpected exit code: $exit_code"
		return 1
	fi
}

test_compare_help() {
	local output
	output=$(clinvk compare --help 2>&1)

	assert_not_empty "$output"
	assert_contains "$output" "compare"
}

# =============================================================================
# Main
# =============================================================================

main() {
	setup_test_env

	print_header "Testing clinvk compare"

	run_test "Compare with --backends flag (2 backends)" test_compare_two_backends
	run_test "Compare with --all-backends" test_compare_all_backends
	run_test "Compare with --sequential flag" test_compare_sequential_flag
	run_test "Compare with --json output" test_compare_json_output
	run_test "Compare exit code validation" test_compare_exit_code
	run_test "Compare --help works" test_compare_help

	print_summary
}

main "$@"
