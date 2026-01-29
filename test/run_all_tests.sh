#!/usr/bin/env bash
# Run all integration tests (CLI and API)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

print_header "Running All Integration Tests"

CLI_RESULT=0
API_RESULT=0

# Run CLI tests
echo ""
log_info "Starting CLI tests..."
if "${SCRIPT_DIR}/run_cli_tests.sh"; then
	log_success "CLI tests passed"
else
	log_error "CLI tests failed"
	CLI_RESULT=1
fi

# Run API tests
echo ""
log_info "Starting API tests..."
if "${SCRIPT_DIR}/run_api_tests.sh"; then
	log_success "API tests passed"
else
	log_error "API tests failed"
	API_RESULT=1
fi

# Final summary
echo ""
echo -e "${BOLD}========================================${NC}"
echo -e "${BOLD}Final Summary${NC}"
echo -e "${BOLD}========================================${NC}"

if ((CLI_RESULT == 0)); then
	echo -e "${GREEN}CLI Tests:${NC} PASSED"
else
	echo -e "${RED}CLI Tests:${NC} FAILED"
fi

if ((API_RESULT == 0)); then
	echo -e "${GREEN}API Tests:${NC} PASSED"
else
	echo -e "${RED}API Tests:${NC} FAILED"
fi

echo ""

if ((CLI_RESULT + API_RESULT > 0)); then
	log_error "Some tests failed"
	exit 1
fi

log_success "All tests passed"
exit 0
