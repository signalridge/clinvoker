#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test basic compare functionality
test_compare_basic() {
	if ! skip_if_missing_backends "claude" "codex"; then
		return 0
	fi

	local payload
	payload=$(jq -n \
		--arg prompt "What is 2+2?" \
		'{
            prompt: $prompt,
            backends: ["claude", "codex"],
            dry_run: true
        }')

	local response
	response=$(http_post "/api/v1/compare" "$payload")

	# Verify response structure
	assert_json_equals "$response" "prompt" "What is 2+2?"
	assert_json_field "$response" "backends"
	assert_json_field "$response" "results"

	# Verify we have results for both backends
	local results_count
	results_count=$(echo "$response" | jq -r '.results | length')
	if [[ "$results_count" -ne 2 ]]; then
		log_error "Expected 2 results, got $results_count"
		return 1
	fi

	# Verify each result has required fields
	# Note: session_id is omitted in dry_run mode (omitempty)
	for i in 0 1; do
		assert_json_field "$response" "results[$i].backend"
		assert_json_field "$response" "results[$i].exit_code"
	done
}

# Test compare with all backends
test_compare_all_backends() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n \
		--arg prompt "Calculate 5+5" \
		'{
            prompt: $prompt,
            backends: ["claude", "codex", "gemini"],
            dry_run: true
        }')

	local response
	response=$(http_post "/api/v1/compare" "$payload")

	assert_json_field "$response" "backends"
	assert_json_field "$response" "results"

	# Verify we have results for all backends
	local results_count
	results_count=$(echo "$response" | jq -r '.results | length')
	if [[ "$results_count" -ne 3 ]]; then
		log_error "Expected 3 results, got $results_count"
		return 1
	fi

	# Verify all requested backends are in results
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

# Test compare with sequential option
test_compare_sequential() {
	if ! skip_if_missing_backends "claude" "codex" "gemini"; then
		return 0
	fi

	local payload
	payload=$(jq -n \
		--arg prompt "What is 3+3?" \
		'{
            prompt: $prompt,
            backends: ["claude", "codex", "gemini"],
            sequential: true,
            dry_run: true
        }')

	local response
	response=$(http_post "/api/v1/compare" "$payload")

	assert_json_field "$response" "backends"
	assert_json_field "$response" "results"

	# Verify results exist
	local results_count
	results_count=$(echo "$response" | jq -r '.results | length')
	if [[ "$results_count" -lt 1 ]]; then
		log_error "Expected at least 1 result, got $results_count"
		return 1
	fi
}

# Test compare response structure
test_compare_response_structure() {
	if ! skip_if_missing_backends "claude" "codex"; then
		return 0
	fi

	local payload
	payload=$(jq -n \
		--arg prompt "Test prompt" \
		'{
            prompt: $prompt,
            backends: ["claude", "codex"],
            dry_run: true
        }')

	local response
	response=$(http_post "/api/v1/compare" "$payload")

	# Verify top-level fields
	assert_json_equals "$response" "prompt" "Test prompt"

	# Verify backends array matches request
	local backends_match
	backends_match=$(echo "$response" | jq -r '.backends | sort | join(",")' | grep -E "claude|codex")
	if [[ -z "$backends_match" ]]; then
		log_error "Backends in response don't match request"
		return 1
	fi

	# Verify each result has complete structure
	# Note: session_id is omitted in dry_run mode (omitempty)
	local result_0_backend result_0_exit
	result_0_backend=$(echo "$response" | jq -r '.results[0].backend // empty')
	result_0_exit=$(echo "$response" | jq -r '.results[0].exit_code // empty')

	if [[ -z "$result_0_backend" || -z "$result_0_exit" ]]; then
		log_error "Result structure incomplete"
		return 1
	fi
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Compare Tests"
	run_test "Basic compare functionality" test_compare_basic
	run_test "Compare with all backends" test_compare_all_backends
	run_test "Compare with sequential option" test_compare_sequential
	run_test "Compare response structure" test_compare_response_structure

	print_summary
}

main "$@"
