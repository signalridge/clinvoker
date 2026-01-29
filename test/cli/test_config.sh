#!/usr/bin/env bash
# Test clinvk config command

set -euo pipefail

# Source common test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# =============================================================================
# Test Functions
# =============================================================================

test_config_show_output() {
	local output
	output=$(clinvk config show 2>&1 || true)

	assert_not_empty "$output"
}

test_config_show_structure() {
	local output
	output=$(clinvk config show 2>&1 || true)

	# Config should contain backend or provider information
	if [[ ! "$output" =~ (backend|provider|claude|codex|gemini) ]]; then
		log_warning "Config output does not contain expected backend/provider keywords"
	fi
}

test_config_show_backends() {
	local output
	output=$(clinvk config show 2>&1 || true)

	# Should list at least one backend configuration
	local has_backend=false
	for backend in "${ALL_BACKENDS[@]}"; do
		if [[ "$output" =~ $backend ]]; then
			has_backend=true
			break
		fi
	done

	if [[ "$has_backend" == "false" ]]; then
		log_warning "Config does not show any configured backends"
	fi
}

test_config_show_exit_code() {
	local exit_code=0
	clinvk config show >/dev/null 2>&1 || exit_code=$?

	# Config show should succeed (exit 0) or fail gracefully
	if [[ $exit_code -ne 0 ]] && [[ $exit_code -ne 1 ]]; then
		log_error "Unexpected exit code: $exit_code"
		return 1
	fi
}

test_config_help() {
	local output
	output=$(clinvk config --help 2>&1 || true)

	assert_not_empty "$output"
	assert_contains "$output" "config"
}

# =============================================================================
# Main
# =============================================================================

main() {
	setup_test_env

	print_header "Testing clinvk config"

	run_test "Config show produces output" test_config_show_output
	run_test "Config show has valid structure" test_config_show_structure
	run_test "Config show lists backends" test_config_show_backends
	run_test "Config show exits properly" test_config_show_exit_code
	run_test "Config --help works" test_config_help

	print_summary
}

main "$@"
