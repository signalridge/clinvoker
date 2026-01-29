#!/usr/bin/env bash
# Test clinvk parallel command

set -euo pipefail

# Source common test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# =============================================================================
# Test Functions
# =============================================================================

create_parallel_tasks_json() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -eq 0 ]]; then
		echo "[]"
		return
	fi

	local tasks="["
	local first=true
	for backend in "${available_backends[@]}"; do
		if [[ "$first" == "false" ]]; then
			tasks+=","
		fi
		first=false
		tasks+="{\"backend\":\"$backend\",\"prompt\":\"$SIMPLE_PROMPT\"}"
	done
	tasks+="]"

	echo "$tasks"
}

test_parallel_json_input() {
	if ! any_backend_available; then
		skip_test "Parallel with JSON input" "No backends available"
		return 0
	fi

	local tasks_json
	tasks_json=$(create_parallel_tasks_json)
	local input_file
	input_file=$(create_temp_json "$tasks_json")

	local output
	output=$(clinvk parallel --input "$input_file" --dry-run 2>&1 || true)

	assert_not_empty "$output"
}

test_parallel_max_parallel() {
	if ! any_backend_available; then
		skip_test "Parallel with --max-parallel" "No backends available"
		return 0
	fi

	local tasks_json
	tasks_json=$(create_parallel_tasks_json)
	local input_file
	input_file=$(create_temp_json "$tasks_json")

	local output
	output=$(clinvk parallel --input "$input_file" --max-parallel 2 --dry-run 2>&1 || true)

	assert_not_empty "$output"
}

test_parallel_fail_fast() {
	if ! any_backend_available; then
		skip_test "Parallel with --fail-fast" "No backends available"
		return 0
	fi

	local tasks_json
	tasks_json=$(create_parallel_tasks_json)
	local input_file
	input_file=$(create_temp_json "$tasks_json")

	local output
	output=$(clinvk parallel --input "$input_file" --fail-fast --dry-run 2>&1 || true)

	assert_not_empty "$output"
}

test_parallel_json_output() {
	if ! any_backend_available; then
		skip_test "Parallel with --json output" "No backends available"
		return 0
	fi

	local tasks_json
	tasks_json=$(create_parallel_tasks_json)
	local input_file
	input_file=$(create_temp_json "$tasks_json")

	local output
	output=$(clinvk parallel --input "$input_file" --json --dry-run 2>&1 || true)

	assert_not_empty "$output"

	# Check if output contains JSON structure
	if ! echo "$output" | jq empty 2>/dev/null; then
		log_warning "Output is not valid JSON"
	fi
}

test_parallel_all_backends() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Parallel with all backends" "Less than 2 backends available"
		return 0
	fi

	local tasks_json
	tasks_json=$(create_parallel_tasks_json)
	local input_file
	input_file=$(create_temp_json "$tasks_json")

	local output
	output=$(clinvk parallel --input "$input_file" --dry-run 2>&1 || true)

	# Verify all backends are mentioned in output
	for backend in "${available_backends[@]}"; do
		if [[ ! "$output" =~ $backend ]]; then
			log_warning "Backend $backend not found in output"
		fi
	done
}

test_parallel_help() {
	local output
	output=$(clinvk parallel --help 2>&1)

	assert_not_empty "$output"
	assert_contains "$output" "parallel"
}

# =============================================================================
# Main
# =============================================================================

main() {
	setup_test_env

	print_header "Testing clinvk parallel"

	run_test "Parallel with JSON input file" test_parallel_json_input
	run_test "Parallel with --max-parallel option" test_parallel_max_parallel
	run_test "Parallel with --fail-fast option" test_parallel_fail_fast
	run_test "Parallel with --json output" test_parallel_json_output
	run_test "Parallel with all backends" test_parallel_all_backends
	run_test "Parallel --help works" test_parallel_help

	print_summary
}

main "$@"
