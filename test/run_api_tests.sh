#!/usr/bin/env bash
# Run all API tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

print_header "Running API Tests"

# Setup environment and start server for non-MCP API suites
setup_test_env
start_server

API_TEST_DIR="${SCRIPT_DIR}/api"
SUITES_PASSED=0
SUITES_FAILED=0

run_suite() {
	local suite_path="$1"
	local test_name
	test_name="$(basename "$suite_path" .sh)"

	print_subheader "$test_name"
	if "$suite_path"; then
		log_success "Suite: $test_name"
		((SUITES_PASSED += 1)) || true
	else
		log_error "Suite: $test_name"
		((SUITES_FAILED += 1)) || true
	fi
}

# Run regular API suites (exclude MCP suites; they have dedicated lifecycle)
for test_script in "${API_TEST_DIR}"/test_*.sh; do
	if [[ -x "$test_script" ]]; then
		test_name="$(basename "$test_script" .sh)"
		if [[ "$test_name" == "test_mcp" || "$test_name" == "test_mcp_auth" ]]; then
			continue
		fi
		run_suite "$test_script"
	fi
done

# Stop default API server before MCP suites
stop_server

# Run dedicated MCP suites
for mcp_suite in "${API_TEST_DIR}/test_mcp.sh" "${API_TEST_DIR}/test_mcp_auth.sh"; do
	if [[ -x "$mcp_suite" ]]; then
		run_suite "$mcp_suite"
	fi
done

# Print overall summary
echo ""
echo -e "${BOLD}========================================${NC}"
echo -e "${BOLD}API Test Suites Summary${NC}"
echo -e "${BOLD}========================================${NC}"
echo -e "${GREEN}Passed:${NC}  $SUITES_PASSED"
echo -e "${RED}Failed:${NC}  $SUITES_FAILED"
echo -e "Total:   $((SUITES_PASSED + SUITES_FAILED))"
echo ""

if ((SUITES_FAILED > 0)); then
	exit 1
fi
exit 0
