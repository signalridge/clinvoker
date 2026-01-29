#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test Anthropic messages endpoint
test_anthropic_messages() {
	local payload
	payload=$(jq -n '{
        model: "claude-3-opus-20240229",
        messages: [
            {role: "user", content: "What is 2+2?"}
        ],
        max_tokens: 100,
        system: "dry_run: true"
    }')

	local response
	response=$(http_post "/anthropic/v1/messages" "$payload")

	# Verify Anthropic response format
	assert_json_field "$response" "id"
	assert_json_equals "$response" "type" "message"
	assert_json_equals "$response" "role" "assistant"
	assert_json_field "$response" "content"
	assert_json_field "$response" "model"
	assert_json_field "$response" "stop_reason"
	assert_json_field "$response" "usage"

	# Verify content structure
	assert_json_equals "$response" "content[0].type" "text"
	assert_json_field "$response" "content[0].text"

	# Verify usage structure
	assert_json_field "$response" "usage.input_tokens"
	assert_json_field "$response" "usage.output_tokens"
}

# Test Anthropic messages with system prompt
test_anthropic_messages_with_system() {
	local payload
	payload=$(jq -n '{
        model: "claude-3-sonnet-20240229",
        messages: [
            {role: "user", content: "Hello"}
        ],
        max_tokens: 50,
        system: "You are a helpful assistant. dry_run: true"
    }')

	local response
	response=$(http_post "/anthropic/v1/messages" "$payload")

	assert_json_field "$response" "id"
	assert_json_equals "$response" "type" "message"
	assert_json_field "$response" "content"
}

# Test Anthropic messages with streaming
test_anthropic_messages_streaming() {
	local payload
	payload=$(jq -n '{
        model: "claude-3-opus-20240229",
        messages: [
            {role: "user", content: "Count to 3"}
        ],
        max_tokens: 100,
        stream: true,
        system: "dry_run: true"
    }')

	# For streaming, we just verify the request is accepted
	local response http_code
	response=$(http_post "/anthropic/v1/messages" "$payload" "http_code" || true)
	http_code=$(echo "$response" | tail -n1)

	if [[ "$http_code" -ge 400 ]]; then
		log_error "Streaming request failed with status $http_code"
		return 1
	fi
}

# Test Anthropic messages with temperature
test_anthropic_messages_with_temperature() {
	local payload
	payload=$(jq -n '{
        model: "claude-3-opus-20240229",
        messages: [
            {role: "user", content: "Test"}
        ],
        max_tokens: 100,
        temperature: 0.7,
        system: "dry_run: true"
    }')

	local response
	response=$(http_post "/anthropic/v1/messages" "$payload")

	assert_json_field "$response" "id"
	assert_json_equals "$response" "type" "message"
}

# Test Anthropic messages with multiple content blocks
test_anthropic_messages_multi_content() {
	local payload
	payload=$(jq -n '{
        model: "claude-3-opus-20240229",
        messages: [
            {role: "user", content: "First question"},
            {role: "assistant", content: "First answer"},
            {role: "user", content: "Second question"}
        ],
        max_tokens: 100,
        system: "dry_run: true"
    }')

	local response
	response=$(http_post "/anthropic/v1/messages" "$payload")

	assert_json_field "$response" "id"
	assert_json_equals "$response" "type" "message"
	assert_json_field "$response" "content"
}

# Test Anthropic error handling
test_anthropic_error_handling() {
	# Test with missing required field (messages)
	local payload
	payload=$(jq -n '{
        model: "claude-3-opus-20240229",
        max_tokens: 100
    }')

	local response http_code
	response=$(http_post "/anthropic/v1/messages" "$payload" "http_code" || true)
	http_code=$(echo "$response" | tail -n1)

	if [[ "$http_code" -lt 400 ]]; then
		log_error "Expected error status for missing messages, got $http_code"
		return 1
	fi
}

# Test Anthropic response format compliance
test_anthropic_response_format() {
	local payload
	payload=$(jq -n '{
        model: "claude-3-opus-20240229",
        messages: [
            {role: "user", content: "Hello"}
        ],
        max_tokens: 50,
        system: "dry_run: true"
    }')

	local response
	response=$(http_post "/anthropic/v1/messages" "$payload")

	# Verify all required top-level fields
	local id type role content model stop_reason usage
	id=$(echo "$response" | jq -r '.id // empty')
	type=$(echo "$response" | jq -r '.type // empty')
	role=$(echo "$response" | jq -r '.role // empty')
	content=$(echo "$response" | jq -r '.content // empty')
	model=$(echo "$response" | jq -r '.model // empty')
	stop_reason=$(echo "$response" | jq -r '.stop_reason // empty')
	usage=$(echo "$response" | jq -r '.usage // empty')

	if [[ -z "$id" || -z "$type" || -z "$role" || -z "$content" || -z "$model" || -z "$stop_reason" || -z "$usage" ]]; then
		log_error "Response missing required Anthropic fields"
		return 1
	fi

	# Verify type and role values
	if [[ "$type" != "message" ]]; then
		log_error "Expected type 'message', got '$type'"
		return 1
	fi
	if [[ "$role" != "assistant" ]]; then
		log_error "Expected role 'assistant', got '$role'"
		return 1
	fi

	# Verify content block structure
	local content_type content_text
	content_type=$(echo "$response" | jq -r '.content[0].type // empty')
	content_text=$(echo "$response" | jq -r '.content[0].text // empty')

	if [[ "$content_type" != "text" ]]; then
		log_error "Expected content type 'text', got '$content_type'"
		return 1
	fi
	if [[ -z "$content_text" ]]; then
		log_error "Content text is empty"
		return 1
	fi

	# Verify usage structure
	local input_tokens output_tokens
	input_tokens=$(echo "$response" | jq -r '.usage.input_tokens // empty')
	output_tokens=$(echo "$response" | jq -r '.usage.output_tokens // empty')

	if [[ -z "$input_tokens" || -z "$output_tokens" ]]; then
		log_error "Usage structure incomplete"
		return 1
	fi
}

# Test Anthropic with stop sequences
test_anthropic_stop_sequences() {
	local payload
	payload=$(jq -n '{
        model: "claude-3-opus-20240229",
        messages: [
            {role: "user", content: "Count to 10"}
        ],
        max_tokens: 100,
        stop_sequences: ["5", "six"],
        system: "dry_run: true"
    }')

	local response
	response=$(http_post "/anthropic/v1/messages" "$payload")

	assert_json_field "$response" "id"
	assert_json_equals "$response" "type" "message"
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Anthropic API Tests"
	run_test "Messages endpoint" test_anthropic_messages
	run_test "Messages with system prompt" test_anthropic_messages_with_system
	run_test "Messages streaming" test_anthropic_messages_streaming
	run_test "Messages with temperature" test_anthropic_messages_with_temperature
	run_test "Messages with multiple content" test_anthropic_messages_multi_content
	run_test "Error handling" test_anthropic_error_handling
	run_test "Response format compliance" test_anthropic_response_format
	run_test "Messages with stop sequences" test_anthropic_stop_sequences

	print_summary
}

main "$@"
