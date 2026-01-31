#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test basic parallel execution
test_parallel_basic() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n '{
        tasks: [
            {backend: "claude", prompt: "What is 2+2?"},
            {backend: "codex", prompt: "What is 3+3?"},
            {backend: "gemini", prompt: "What is 4+4?"}
        ],
        dry_run: true
    }')

	local response
	response=$(http_post "/api/v1/parallel" "$payload")

	# Verify response structure
	assert_json_equals "$response" "total_tasks" "3"
	assert_json_field "$response" "completed"
	assert_json_field "$response" "failed"
	assert_json_field "$response" "results"

	# Verify results array length
	local results_count
	results_count=$(echo "$response" | jq -r '.results | length')
	if [[ "$results_count" -ne 3 ]]; then
		log_error "Expected 3 results, got $results_count"
		return 1
	fi

	# Verify each result has required fields
	# Note: session_id is omitted in dry_run mode (omitempty)
	for i in 0 1 2; do
		assert_json_field "$response" "results[$i].backend"
		assert_json_field "$response" "results[$i].exit_code"
	done
}

# Test parallel with max_parallel option
test_parallel_max_parallel() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n '{
        tasks: [
            {backend: "claude", prompt: "Task 1"},
            {backend: "codex", prompt: "Task 2"},
            {backend: "gemini", prompt: "Task 3"}
        ],
        max_parallel: 2,
        dry_run: true
    }')

	local response
	response=$(http_post "/api/v1/parallel" "$payload")

	assert_json_equals "$response" "total_tasks" "3"
	assert_json_field "$response" "results"
}

# Test parallel with fail_fast option
test_parallel_fail_fast() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n '{
        tasks: [
            {backend: "claude", prompt: "Good task"},
            {backend: "codex", prompt: "Another good task"},
            {backend: "gemini", prompt: "Yet another task"}
        ],
        fail_fast: true,
        dry_run: true
    }')

	local response
	response=$(http_post "/api/v1/parallel" "$payload")

	# Should have results (might be partial due to fail_fast)
	assert_json_equals "$response" "total_tasks" "3"
	assert_json_field "$response" "results"
}

# Test parallel with all backends
test_parallel_all_backends() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n '{
        tasks: [
            {backend: "claude", prompt: "Calculate 5+5"},
            {backend: "codex", prompt: "Calculate 6+6"},
            {backend: "gemini", prompt: "Calculate 7+7"}
        ],
        max_parallel: 3,
        dry_run: true
    }')

	local response
	response=$(http_post "/api/v1/parallel" "$payload")

	assert_json_equals "$response" "total_tasks" "3"

	# Verify all backends are represented
	local claude_found codex_found gemini_found
	claude_found=$(echo "$response" | jq -r '[.results[].backend] | any(. == "claude")')
	codex_found=$(echo "$response" | jq -r '[.results[].backend] | any(. == "codex")')
	gemini_found=$(echo "$response" | jq -r '[.results[].backend] | any(. == "gemini")')

	if [[ "$claude_found" != "true" ]]; then
		log_error "Claude backend not found in results"
		return 1
	fi
	if [[ "$codex_found" != "true" ]]; then
		log_error "Codex backend not found in results"
		return 1
	fi
	if [[ "$gemini_found" != "true" ]]; then
		log_error "Gemini backend not found in results"
		return 1
	fi
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Parallel Execution Tests"
	run_test "Basic parallel execution" test_parallel_basic
	run_test "Parallel with max_parallel option" test_parallel_max_parallel
	run_test "Parallel with fail_fast option" test_parallel_fail_fast
	run_test "Parallel execution with all backends" test_parallel_all_backends

	print_summary
}

main "$@"
