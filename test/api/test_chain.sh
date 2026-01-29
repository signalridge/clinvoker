#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test basic chain execution
test_chain_basic() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n '{
        steps: [
            {backend: "claude", prompt: "What is 2+2?"},
            {backend: "codex", prompt: "Previous answer was {{previous}}"},
            {backend: "gemini", prompt: "Summarize: {{previous}}"}
        ],
        dry_run: true
    }')

	local response
	response=$(http_post "/api/v1/chain" "$payload")

	# Verify response structure
	assert_json_equals "$response" "total_steps" "3"
	assert_json_field "$response" "completed_steps"
	assert_json_field "$response" "results"

	# Verify results array length
	local results_count
	results_count=$(echo "$response" | jq -r '.results | length')
	if [[ "$results_count" -eq 0 ]]; then
		log_error "Expected at least 1 result, got 0"
		return 1
	fi

	# Verify each result has required fields
	for i in $(seq 0 $((results_count - 1))); do
		assert_json_field "$response" "results[$i].backend"
		assert_json_field "$response" "results[$i].exit_code"
	done
}

# Test chain with previous placeholder
test_chain_with_placeholder() {
	if ! skip_if_missing_backends "claude" "codex"; then
		return 0
	fi

	local payload
	payload=$(jq -n '{
        steps: [
            {backend: "claude", prompt: "Step 1: Generate number"},
            {backend: "codex", prompt: "Step 2: Use {{previous}}"}
        ],
        dry_run: true
    }')

	local response
	response=$(http_post "/api/v1/chain" "$payload")

	assert_json_equals "$response" "total_steps" "2"
	assert_json_field "$response" "results"
}

# Test chain with stop_on_failure option
test_chain_stop_on_failure() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n '{
        steps: [
            {backend: "claude", prompt: "Good step"},
            {backend: "codex", prompt: "Another step"},
            {backend: "gemini", prompt: "Final step"}
        ],
        stop_on_failure: true,
        dry_run: true
    }')

	local response
	response=$(http_post "/api/v1/chain" "$payload")

	assert_json_equals "$response" "total_steps" "3"
	assert_json_field "$response" "completed_steps"
	assert_json_field "$response" "results"

	# In dry_run mode, all steps should complete successfully
	# Just verify the structure is correct
	local completed_steps
	completed_steps=$(echo "$response" | jq -r '.completed_steps')
	if [[ "$completed_steps" -lt 1 ]]; then
		log_error "Expected at least 1 completed step, got $completed_steps"
		return 1
	fi
}

# Test chain across all backends
test_chain_all_backends() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n '{
        steps: [
            {backend: "claude", prompt: "Step 1 by Claude"},
            {backend: "codex", prompt: "Step 2 by Codex: {{previous}}"},
            {backend: "gemini", prompt: "Step 3 by Gemini: {{previous}}"}
        ],
        dry_run: true
    }')

	local response
	response=$(http_post "/api/v1/chain" "$payload")

	assert_json_equals "$response" "total_steps" "3"

	# Verify backends are in order
	local backend_0 backend_1 backend_2
	backend_0=$(echo "$response" | jq -r '.results[0].backend // empty')
	backend_1=$(echo "$response" | jq -r '.results[1].backend // empty')
	backend_2=$(echo "$response" | jq -r '.results[2].backend // empty')

	if [[ -n "$backend_0" && "$backend_0" != "claude" ]]; then
		log_error "Expected first backend to be claude, got $backend_0"
		return 1
	fi
	if [[ -n "$backend_1" && "$backend_1" != "codex" ]]; then
		log_error "Expected second backend to be codex, got $backend_1"
		return 1
	fi
	if [[ -n "$backend_2" && "$backend_2" != "gemini" ]]; then
		log_error "Expected third backend to be gemini, got $backend_2"
		return 1
	fi
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Chain Execution Tests"
	run_test "Basic chain execution" test_chain_basic
	run_test "Chain with placeholder" test_chain_with_placeholder
	run_test "Chain with stop_on_failure" test_chain_stop_on_failure
	run_test "Chain execution across all backends" test_chain_all_backends

	print_summary
}

main "$@"
