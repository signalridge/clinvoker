#!/usr/bin/env bash
# Test clinvk basic prompt execution

set -euo pipefail

# Source common test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# =============================================================================
# Test Functions
# =============================================================================

test_prompt_dry_run() {
	local backend="$1"

	if ! backend_available "$backend"; then
		skip_test "Dry run with $backend" "$backend not available"
		return 0
	fi

	local output
	output=$(clinvk --backend "$backend" --dry-run "$SIMPLE_PROMPT" 2>&1)

	assert_not_empty "$output"
	# Dry run should show "Would execute:" message
	assert_contains "$output" "Would execute"
}

test_prompt_json_output() {
	local backend="$1"

	if ! backend_available "$backend"; then
		skip_test "JSON output with $backend" "$backend not available"
		return 0
	fi

	local output
	output=$(clinvk --backend "$backend" --output-format json --dry-run "$SIMPLE_PROMPT" 2>&1 || true)

	assert_not_empty "$output"

	# Check if output is valid JSON
	if ! echo "$output" | jq empty 2>/dev/null; then
		log_warning "Output is not valid JSON: $output"
	fi
}

test_prompt_ephemeral_flag() {
	local backend="$1"

	if ! backend_available "$backend"; then
		skip_test "Ephemeral flag with $backend" "$backend not available"
		return 0
	fi

	local exit_code=0
	clinvk --backend "$backend" --ephemeral --dry-run "$SIMPLE_PROMPT" >/dev/null 2>&1 || exit_code=$?

	# Should succeed or fail gracefully (not crash)
	if [[ $exit_code -gt 1 ]]; then
		log_error "Unexpected exit code: $exit_code"
		return 1
	fi
}

test_prompt_model_flag() {
	local backend="$1"

	if ! backend_available "$backend"; then
		skip_test "Model flag with $backend" "$backend not available"
		return 0
	fi

	local model=""
	case "$backend" in
	claude)
		model="claude-3-5-sonnet-20241022"
		;;
	codex)
		model="gpt-4o"
		;;
	gemini)
		model="gemini-1.5-pro"
		;;
	esac

	local exit_code=0
	clinvk --backend "$backend" --model "$model" --dry-run "$SIMPLE_PROMPT" >/dev/null 2>&1 || exit_code=$?

	# Should succeed or fail gracefully
	if [[ $exit_code -gt 1 ]]; then
		log_error "Unexpected exit code: $exit_code"
		return 1
	fi
}

test_prompt_basic_execution() {
	local backend="$1"

	if ! backend_available "$backend"; then
		skip_test "Basic execution with $backend" "$backend not available"
		return 0
	fi

	local exit_code=0
	clinvk --backend "$backend" --dry-run "$SIMPLE_PROMPT" >/dev/null 2>&1 || exit_code=$?

	# Dry run should succeed with exit code 0
	assert_exit_code 0 "$exit_code"
}

# =============================================================================
# Main
# =============================================================================

main() {
	setup_test_env

	print_header "Testing clinvk prompt execution"

	# Test each backend
	for backend in "${ALL_BACKENDS[@]}"; do
		print_subheader "Testing with backend: $backend"

		run_test "[$backend] Basic prompt with --dry-run" test_prompt_basic_execution "$backend"
		run_test "[$backend] Prompt with --output-format json" test_prompt_json_output "$backend"
		run_test "[$backend] Prompt with --ephemeral flag" test_prompt_ephemeral_flag "$backend"
		run_test "[$backend] Prompt with --model flag" test_prompt_model_flag "$backend"
		run_test "[$backend] Dry run validation" test_prompt_dry_run "$backend"
	done

	print_summary
}

main "$@"
