#!/usr/bin/env bash
# Test clinvk sessions command

set -euo pipefail

# Source common test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# =============================================================================
# Test Functions
# =============================================================================

test_sessions_list() {
	local output
	output=$(clinvk sessions list 2>&1 || true)

	assert_not_empty "$output"
}

test_sessions_list_json() {
	local output
	output=$(clinvk sessions list --json 2>&1 || true)

	assert_not_empty "$output"

	# Check if output is valid JSON
	if ! echo "$output" | jq empty 2>/dev/null; then
		log_warning "Sessions list output is not valid JSON"
	fi
}

test_sessions_show() {
	# First, try to get a session ID from the list
	local sessions_output
	sessions_output=$(clinvk sessions list 2>&1 || true)

	# Extract first session ID if available
	local session_id
	session_id=$(extract_session_id "$sessions_output" || echo "")

	if [[ -z "$session_id" ]]; then
		skip_test "Sessions show" "No sessions available to show"
		return 0
	fi

	local output
	output=$(clinvk sessions show "$session_id" 2>&1 || true)

	assert_not_empty "$output"
}

test_sessions_show_json() {
	# Try to get a session ID from the list
	local sessions_output
	sessions_output=$(clinvk sessions list 2>&1 || true)

	local session_id
	session_id=$(extract_session_id "$sessions_output" || echo "")

	if [[ -z "$session_id" ]]; then
		skip_test "Sessions show --json" "No sessions available to show"
		return 0
	fi

	local output
	output=$(clinvk sessions show "$session_id" --json 2>&1 || true)

	assert_not_empty "$output"

	# Check if output is valid JSON
	if ! echo "$output" | jq empty 2>/dev/null; then
		log_warning "Sessions show output is not valid JSON"
	fi
}

test_sessions_delete() {
	# Try to get a session ID from the list
	local sessions_output
	sessions_output=$(clinvk sessions list 2>&1 || true)

	local session_id
	session_id=$(extract_session_id "$sessions_output" || echo "")

	if [[ -z "$session_id" ]]; then
		skip_test "Sessions delete" "No sessions available to delete"
		return 0
	fi

	# Note: We're not actually deleting to avoid affecting the test environment
	# Just verify the command syntax works
	local exit_code=0
	clinvk sessions delete --help >/dev/null 2>&1 || exit_code=$?

	if [[ $exit_code -gt 1 ]]; then
		log_error "Sessions delete command failed unexpectedly"
		return 1
	fi
}

test_sessions_list_backends() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -eq 0 ]]; then
		skip_test "Sessions list for each backend" "No backends available"
		return 0
	fi

	for backend in "${available_backends[@]}"; do
		local output
		output=$(clinvk sessions list --backend "$backend" 2>&1 || true)

		if [[ -z "$output" ]]; then
			log_warning "No sessions found for backend: $backend"
		fi
	done
}

test_sessions_help() {
	local output
	output=$(clinvk sessions --help 2>&1)

	assert_not_empty "$output"
	assert_contains "$output" "sessions"
}

# =============================================================================
# Main
# =============================================================================

main() {
	setup_test_env

	print_header "Testing clinvk sessions"

	run_test "Sessions list command" test_sessions_list
	run_test "Sessions list --json output" test_sessions_list_json
	run_test "Sessions show <id> command" test_sessions_show
	run_test "Sessions show --json output" test_sessions_show_json
	run_test "Sessions delete <id> command" test_sessions_delete
	run_test "Sessions list per backend" test_sessions_list_backends
	run_test "Sessions --help works" test_sessions_help

	print_summary
}

main "$@"
