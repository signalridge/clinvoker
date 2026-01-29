#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test backends listing endpoint
test_list_backends() {
	local response
	response=$(http_get "/api/v1/backends")

	# Verify backends array exists
	assert_json_field "$response" "backends"

	# Check that we have at least one backend
	local backend_count
	backend_count=$(echo "$response" | jq -r '.backends | length')
	if [[ "$backend_count" -lt 1 ]]; then
		log_error "Expected at least 1 backend, got $backend_count"
		return 1
	fi

	# Verify each backend has required fields
	local backend_names=("claude" "codex" "gemini")
	for backend_name in "${backend_names[@]}"; do
		local backend_data
		backend_data=$(echo "$response" | jq -r ".backends[] | select(.name == \"$backend_name\")")

		if [[ -z "$backend_data" ]]; then
			log_error "Backend '$backend_name' not found in response"
			return 1
		fi

		# Check required fields
		local name available
		name=$(echo "$backend_data" | jq -r '.name')
		available=$(echo "$backend_data" | jq -r '.available')

		if [[ "$name" != "$backend_name" ]]; then
			log_error "Backend name mismatch: expected '$backend_name', got '$name'"
			return 1
		fi

		if [[ "$available" != "true" && "$available" != "false" ]]; then
			log_error "Backend '$backend_name' has invalid 'available' field: $available"
			return 1
		fi
	done
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Backends Endpoint"
	run_test "List all available backends" test_list_backends

	print_summary
}

main "$@"
