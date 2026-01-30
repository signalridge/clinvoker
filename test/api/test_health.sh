#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test health endpoint
test_health_endpoint() {
	local response status
	response=$(http_get "/health")

	# Status can be "ok" or "degraded" depending on backend availability
	status=$(echo "$response" | jq -r '.status')
	if [[ "$status" != "ok" && "$status" != "degraded" ]]; then
		log_error "Expected status 'ok' or 'degraded', got '$status'"
		return 1
	fi
}

# Test health endpoint returns backend status
test_health_backends() {
	local response
	response=$(http_get "/health")

	# Verify backends field exists
	assert_json_field "$response" "backends"

	# Verify backends array has at least one entry
	local backend_count
	backend_count=$(echo "$response" | jq -r '.backends | length')
	if [[ "$backend_count" -lt 1 ]]; then
		log_error "Expected at least 1 backend in health response, got $backend_count"
		return 1
	fi
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Health Endpoint"
	run_test "Health returns valid status" test_health_endpoint
	run_test "Health returns backend status" test_health_backends

	print_summary
}

main "$@"
