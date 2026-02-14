#!/usr/bin/env bash
# Run all CLI tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

print_header "Running CLI Tests"

CLI_TEST_DIR="${SCRIPT_DIR}/cli"
TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_SKIPPED=0

extract_suite_count() {
	local label="$1"
	local output="$2"

	printf '%s\n' "$output" | sed -nE "s/^${label}:[[:space:]]*([0-9]+).*/\\1/p" | tail -n1
}

# Run each CLI test script
for test_script in "${CLI_TEST_DIR}"/test_*.sh; do
	if [[ -x "$test_script" ]]; then
		test_name="$(basename "$test_script" .sh)"
		print_subheader "$test_name"

		set +e
		suite_output="$("$test_script" 2>&1)"
		suite_status=$?
		set -e

		printf '%s\n' "$suite_output"

		suite_passed="$(extract_suite_count "Passed" "$suite_output")"
		suite_failed="$(extract_suite_count "Failed" "$suite_output")"
		suite_skipped="$(extract_suite_count "Skipped" "$suite_output")"
		suite_passed="${suite_passed:-0}"
		suite_failed="${suite_failed:-0}"
		suite_skipped="${suite_skipped:-0}"

		if ((suite_status == 0)); then
			log_success "Suite: $test_name"
		else
			log_error "Suite: $test_name"
		fi

		# Accumulate totals
		((TOTAL_PASSED += suite_passed)) || true
		((TOTAL_FAILED += suite_failed)) || true
		((TOTAL_SKIPPED += suite_skipped)) || true
	fi
done

# Print overall summary
echo ""
echo -e "${BOLD}========================================${NC}"
echo -e "${BOLD}CLI Tests Overall Summary${NC}"
echo -e "${BOLD}========================================${NC}"
echo -e "${GREEN}Passed:${NC}  $TOTAL_PASSED"
echo -e "${RED}Failed:${NC}  $TOTAL_FAILED"
echo -e "${YELLOW}Skipped:${NC} $TOTAL_SKIPPED"
echo -e "Total:   $((TOTAL_PASSED + TOTAL_FAILED + TOTAL_SKIPPED))"
echo ""

if ((TOTAL_FAILED > 0)); then
	exit 1
fi
exit 0
