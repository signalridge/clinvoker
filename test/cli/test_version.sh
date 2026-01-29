#!/usr/bin/env bash
# Test clinvk version command

set -euo pipefail

# Source common test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# =============================================================================
# Test Functions
# =============================================================================

test_version_output() {
	local output
	output=$(clinvk version 2>&1)

	assert_not_empty "$output"
	# Output should contain "clinvk" or version info
	assert_contains "$output" "clinvk"
}

test_version_format() {
	local output
	output=$(clinvk version 2>&1)

	# Should contain a version number pattern (e.g., v1.0.0, 1.0.0, or dev)
	if [[ ! "$output" =~ [0-9]+\.[0-9]+\.[0-9]+ ]] && [[ ! "$output" =~ dev ]]; then
		log_error "Version output does not contain valid version format: $output"
		return 1
	fi
}

test_version_exit_code() {
	local exit_code=0
	clinvk version >/dev/null 2>&1 || exit_code=$?

	assert_exit_code 0 "$exit_code"
}

test_version_help_flag() {
	local output
	output=$(clinvk version --help 2>&1 || true)

	# Should contain usage information or succeed
	assert_not_empty "$output"
}

# =============================================================================
# Main
# =============================================================================

main() {
	setup_test_env

	print_header "Testing clinvk version"

	run_test "Version output is non-empty" test_version_output
	run_test "Version format is valid" test_version_format
	run_test "Version command exits with code 0" test_version_exit_code
	run_test "Version --help flag works" test_version_help_flag

	print_summary
}

main "$@"
