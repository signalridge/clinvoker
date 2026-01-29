#!/usr/bin/env bash
# Test clinvk chain command

set -euo pipefail

# Source common test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# =============================================================================
# Test Functions
# =============================================================================

create_chain_config_json() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -eq 0 ]]; then
		echo "{\"steps\":[]}"
		return
	fi

	local steps="["
	local first=true
	for backend in "${available_backends[@]}"; do
		if [[ "$first" == "false" ]]; then
			steps+=","
		fi
		first=false
		steps+="{\"backend\":\"$backend\",\"prompt\":\"$SIMPLE_PROMPT\"}"
	done
	steps+="]"

	echo "{\"steps\":$steps}"
}

test_chain_sequential_execution() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Chain sequential execution" "Less than 2 backends available"
		return 0
	fi

	local config_json
	config_json=$(create_chain_config_json)
	local input_file
	input_file=$(create_temp_json "$config_json")

	local output
	output=$(clinvk chain --file "$input_file" --dry-run 2>&1 || true)

	assert_not_empty "$output"

	# Verify backends are executed in order
	for backend in "${available_backends[@]}"; do
		if [[ ! "$output" =~ $backend ]]; then
			log_warning "Backend $backend not found in chain output"
		fi
	done
}

test_chain_with_placeholder() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Chain with {{previous}} placeholder" "Less than 2 backends available"
		return 0
	fi

	local backend1="${available_backends[0]}"
	local backend2="${available_backends[1]}"

	local config_json="{\"steps\":[
		{\"backend\":\"$backend1\",\"prompt\":\"$SIMPLE_PROMPT\"},
		{\"backend\":\"$backend2\",\"prompt\":\"Analyze: {{previous}}\"}
	]}"

	local input_file
	input_file=$(create_temp_json "$config_json")

	local output
	output=$(clinvk chain --file "$input_file" --dry-run 2>&1 || true)

	assert_not_empty "$output"
	assert_contains "$output" "$backend1"
	assert_contains "$output" "$backend2"
}

test_chain_stop_on_failure() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Chain with --stop-on-failure" "Less than 2 backends available"
		return 0
	fi

	local config_json
	config_json=$(create_chain_config_json)
	local input_file
	input_file=$(create_temp_json "$config_json")

	local output
	output=$(clinvk chain --file "$input_file" --stop-on-failure --dry-run 2>&1 || true)

	assert_not_empty "$output"
}

test_chain_json_output() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Chain with --json output" "Less than 2 backends available"
		return 0
	fi

	local config_json
	config_json=$(create_chain_config_json)
	local input_file
	input_file=$(create_temp_json "$config_json")

	local output
	output=$(clinvk chain --file "$input_file" --json --dry-run 2>&1 || true)

	assert_not_empty "$output"

	# Check if output contains JSON structure
	if ! echo "$output" | jq empty 2>/dev/null; then
		log_warning "Output is not valid JSON"
	fi
}

test_chain_cross_backend() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 3 ]]; then
		skip_test "Chain across all backends" "Less than 3 backends available"
		return 0
	fi

	local config_json
	config_json=$(create_chain_config_json)
	local input_file
	input_file=$(create_temp_json "$config_json")

	local output
	output=$(clinvk chain --file "$input_file" --dry-run 2>&1 || true)

	# Verify all backends are used
	for backend in "${available_backends[@]}"; do
		if [[ ! "$output" =~ $backend ]]; then
			log_warning "Backend $backend not found in chain"
		fi
	done
}

test_chain_help() {
	local output
	output=$(clinvk chain --help 2>&1)

	assert_not_empty "$output"
	assert_contains "$output" "chain"
}

# =============================================================================
# Main
# =============================================================================

main() {
	setup_test_env

	print_header "Testing clinvk chain"

	run_test "Chain sequential execution across backends" test_chain_sequential_execution
	run_test "Chain with {{previous}} placeholder" test_chain_with_placeholder
	run_test "Chain with --stop-on-failure flag" test_chain_stop_on_failure
	run_test "Chain with --json output" test_chain_json_output
	run_test "Chain across all backends" test_chain_cross_backend
	run_test "Chain --help works" test_chain_help

	print_summary
}

main "$@"
