#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test basic prompt execution with dry_run
test_basic_prompt() {
	local backend="$1"

	if ! backend_available "$backend"; then
		skip_test "$CURRENT_TEST_NAME" "$backend not available"
		return 0
	fi

	local payload
	payload=$(jq -n \
		--arg backend "$backend" \
		--arg prompt "What is 2+2?" \
		'{
            backend: $backend,
            prompt: $prompt,
            dry_run: true
        }')

	local response
	response=$(http_post "/api/v1/prompt" "$payload")

	# Verify required fields
	assert_json_field "$response" "session_id"
	assert_json_equals "$response" "backend" "$backend"
	assert_json_field "$response" "exit_code"
}

# Test prompt with model option
test_prompt_with_model() {
	local backend="$1"
	local model="$2"

	if ! backend_available "$backend"; then
		skip_test "$CURRENT_TEST_NAME" "$backend not available"
		return 0
	fi

	local payload
	payload=$(jq -n \
		--arg backend "$backend" \
		--arg prompt "What is 2+2?" \
		--arg model "$model" \
		'{
            backend: $backend,
            prompt: $prompt,
            model: $model,
            dry_run: true
        }')

	local response
	response=$(http_post "/api/v1/prompt" "$payload")

	assert_json_field "$response" "session_id"
	assert_json_equals "$response" "backend" "$backend"
}

# Test prompt with ephemeral option
test_prompt_with_ephemeral() {
	local backend="$1"

	if ! backend_available "$backend"; then
		skip_test "$CURRENT_TEST_NAME" "$backend not available"
		return 0
	fi

	local payload
	payload=$(jq -n \
		--arg backend "$backend" \
		--arg prompt "What is 2+2?" \
		'{
            backend: $backend,
            prompt: $prompt,
            ephemeral: true,
            dry_run: true
        }')

	local response
	response=$(http_post "/api/v1/prompt" "$payload")

	# In ephemeral mode, session_id may still be returned but won't be persisted
	assert_json_equals "$response" "backend" "$backend"
	assert_json_field "$response" "exit_code"
}

# Test error handling for empty prompt
test_empty_prompt_error() {
	local backend="$1"

	if ! backend_available "$backend"; then
		skip_test "$CURRENT_TEST_NAME" "$backend not available"
		return 0
	fi

	local payload
	payload=$(jq -n \
		--arg backend "$backend" \
		'{
            backend: $backend,
            prompt: "",
            dry_run: true
        }')

	local full_response http_code
	full_response=$(http_post_status "/api/v1/prompt" "$payload")
	http_code=$(echo "$full_response" | tail -n1)

	# Should return error status (4xx)
	if [[ "$http_code" -lt 400 ]]; then
		log_error "Expected error status for empty prompt, got $http_code"
		return 1
	fi
}

# Main test execution
main() {
	setup_test_env
	start_server

	local backends=("claude" "codex" "gemini")

	# Test each backend
	for backend in "${backends[@]}"; do
		print_subheader "Testing Backend: $backend"
		run_test "Basic prompt for $backend" test_basic_prompt "$backend"
		run_test "Prompt with ephemeral for $backend" test_prompt_with_ephemeral "$backend"
		run_test "Empty prompt error for $backend" test_empty_prompt_error "$backend"
	done

	# Test with model options
	print_subheader "Testing Model Options"
	run_test "Prompt with model for claude" test_prompt_with_model "claude" "claude-3-opus-20240229"
	run_test "Prompt with model for codex" test_prompt_with_model "codex" "gpt-4"
	run_test "Prompt with model for gemini" test_prompt_with_model "gemini" "gemini-pro"

	print_summary
}

main "$@"
