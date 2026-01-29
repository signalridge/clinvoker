#!/usr/bin/env bash
# Test clinvk resume command

set -euo pipefail

# Source common test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# =============================================================================
# Test Functions
# =============================================================================

test_resume_with_session_id() {
	# Try to get a session ID from the list
	local sessions_output
	sessions_output=$(clinvk sessions list 2>&1 || true)

	local session_id
	session_id=$(extract_session_id "$sessions_output" || echo "")

	if [[ -z "$session_id" ]]; then
		skip_test "Resume with session ID" "No sessions available to resume"
		return 0
	fi

	local output
	output=$(clinvk resume "$session_id" --dry-run 2>&1 || true)

	assert_not_empty "$output"
}

test_resume_latest_session() {
	# Check if there are any sessions
	local sessions_output
	sessions_output=$(clinvk sessions list 2>&1 || true)

	if [[ -z "$sessions_output" ]] || [[ "$sessions_output" =~ "No sessions" ]]; then
		skip_test "Resume latest session" "No sessions available"
		return 0
	fi

	local output
	output=$(clinvk resume --latest --dry-run 2>&1 || true)

	assert_not_empty "$output"
}

test_resume_with_backend() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -eq 0 ]]; then
		skip_test "Resume with backend filter" "No backends available"
		return 0
	fi

	local backend="${available_backends[0]}"

	# Get sessions for this backend
	local sessions_output
	sessions_output=$(clinvk sessions list --backend "$backend" 2>&1 || true)

	local session_id
	session_id=$(extract_session_id "$sessions_output" || echo "")

	if [[ -z "$session_id" ]]; then
		skip_test "Resume with backend $backend" "No sessions for $backend"
		return 0
	fi

	local output
	output=$(clinvk resume "$session_id" --backend "$backend" --dry-run 2>&1 || true)

	assert_not_empty "$output"
}

test_resume_with_prompt() {
	# Try to get a session ID
	local sessions_output
	sessions_output=$(clinvk sessions list 2>&1 || true)

	local session_id
	session_id=$(extract_session_id "$sessions_output" || echo "")

	if [[ -z "$session_id" ]]; then
		skip_test "Resume with additional prompt" "No sessions available"
		return 0
	fi

	local output
	output=$(clinvk resume "$session_id" --dry-run "continue with next step" 2>&1 || true)

	assert_not_empty "$output"
}

test_resume_invalid_session() {
	local invalid_id="nonexistent-session-id"

	local exit_code=0
	clinvk resume "$invalid_id" --dry-run >/dev/null 2>&1 || exit_code=$?

	# Should fail with non-zero exit code
	if [[ $exit_code -eq 0 ]]; then
		log_error "Expected resume to fail with invalid session ID"
		return 1
	fi
}

test_resume_help() {
	local output
	output=$(clinvk resume --help 2>&1)

	assert_not_empty "$output"
	assert_contains "$output" "resume"
}

test_resume_all_backends() {
	local available_backends
	IFS=' ' read -ra available_backends <<<"$(get_available_backends)"

	if [[ ${#available_backends[@]} -lt 2 ]]; then
		skip_test "Resume across all backends" "Less than 2 backends available"
		return 0
	fi

	for backend in "${available_backends[@]}"; do
		local sessions_output
		sessions_output=$(clinvk sessions list --backend "$backend" 2>&1 || true)

		local session_id
		session_id=$(extract_session_id "$sessions_output" || echo "")

		if [[ -z "$session_id" ]]; then
			log_warning "No sessions found for backend: $backend"
			continue
		fi

		local output
		output=$(clinvk resume "$session_id" --dry-run 2>&1 || true)

		if [[ -z "$output" ]]; then
			log_warning "Resume failed for $backend session $session_id"
		fi
	done
}

# =============================================================================
# Main
# =============================================================================

main() {
	setup_test_env

	print_header "Testing clinvk resume"

	run_test "Resume with session ID and --dry-run" test_resume_with_session_id
	run_test "Resume latest session" test_resume_latest_session
	run_test "Resume with backend filter" test_resume_with_backend
	run_test "Resume with additional prompt" test_resume_with_prompt
	run_test "Resume with invalid session ID" test_resume_invalid_session
	run_test "Resume across all backends" test_resume_all_backends
	run_test "Resume --help works" test_resume_help

	print_summary
}

main "$@"
