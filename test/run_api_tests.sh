#!/usr/bin/env bash
# Run all API tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

print_header "Running API Tests"

# Setup environment and start server
setup_test_env
start_server

API_TEST_DIR="${SCRIPT_DIR}/api"
TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_SKIPPED=0

# Run each API test script
for test_script in "${API_TEST_DIR}"/test_*.sh; do
	if [[ -x "$test_script" ]]; then
		test_name="$(basename "$test_script" .sh)"
		print_subheader "$test_name"

		# Reset counters for this test
		TESTS_PASSED=0
		TESTS_FAILED=0
		TESTS_SKIPPED=0

		if "$test_script"; then
			log_success "Suite: $test_name"
		else
			log_error "Suite: $test_name"
		fi

		# Accumulate totals
		((TOTAL_PASSED += TESTS_PASSED)) || true
		((TOTAL_FAILED += TESTS_FAILED)) || true
		((TOTAL_SKIPPED += TESTS_SKIPPED)) || true
	fi
done

# Stop server
stop_server

# Print overall summary
echo ""
echo -e "${BOLD}========================================${NC}"
echo -e "${BOLD}API Tests Overall Summary${NC}"
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
