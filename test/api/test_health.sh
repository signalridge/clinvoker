#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test health endpoint
test_health_endpoint() {
	local response
	response=$(http_get "/health")

	assert_json_equals "$response" "status" "ok"
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Health Endpoint"
	run_test "Health returns OK status" test_health_endpoint

	print_summary
}

main "$@"
