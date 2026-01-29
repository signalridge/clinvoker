#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test listing OpenAI models
test_openai_list_models() {
	local response
	response=$(http_get "/openai/v1/models")

	# Verify OpenAI response format
	assert_json_equals "$response" "object" "list"
	assert_json_field "$response" "data"

	# Verify we have at least one model
	local models_count
	models_count=$(echo "$response" | jq -r '.data | length')
	if [[ "$models_count" -lt 1 ]]; then
		log_error "Expected at least 1 model, got $models_count"
		return 1
	fi

	# Verify each model has required fields
	local model_0_id model_0_object
	model_0_id=$(echo "$response" | jq -r '.data[0].id // empty')
	model_0_object=$(echo "$response" | jq -r '.data[0].object // empty')

	if [[ -z "$model_0_id" ]]; then
		log_error "Model missing 'id' field"
		return 1
	fi
	if [[ "$model_0_object" != "model" ]]; then
		log_error "Model 'object' field should be 'model', got '$model_0_object'"
		return 1
	fi
}

# Test OpenAI chat completions endpoint
test_openai_chat_completions() {
	# Note: We use dry_run via system message hack or just test request format
	local payload
	payload=$(jq -n '{
        model: "gpt-3.5-turbo",
        messages: [
            {role: "system", content: "dry_run: true"},
            {role: "user", content: "What is 2+2?"}
        ]
    }')

	local response
	response=$(http_post "/openai/v1/chat/completions" "$payload")

	# Verify OpenAI response format
	assert_json_field "$response" "id"
	assert_json_equals "$response" "object" "chat.completion"
	assert_json_field "$response" "choices"
	assert_json_field "$response" "usage"

	# Verify choices structure
	assert_json_equals "$response" "choices[0].index" "0"
	assert_json_field "$response" "choices[0].message"
	assert_json_equals "$response" "choices[0].message.role" "assistant"
	assert_json_field "$response" "choices[0].message.content"

	# Verify usage structure
	assert_json_field "$response" "usage.prompt_tokens"
	assert_json_field "$response" "usage.completion_tokens"
	assert_json_field "$response" "usage.total_tokens"
}

# Test OpenAI chat completions with streaming
test_openai_chat_completions_streaming() {
	local payload
	payload=$(jq -n '{
        model: "gpt-3.5-turbo",
        messages: [
            {role: "system", content: "dry_run: true"},
            {role: "user", content: "Count to 3"}
        ],
        stream: true
    }')

	# For streaming, we just verify the request is accepted
	# Full SSE validation would require special handling
	local response http_code
	response=$(http_post "/openai/v1/chat/completions" "$payload" "http_code" || true)
	http_code=$(echo "$response" | tail -n1)

	if [[ "$http_code" -ge 400 ]]; then
		log_error "Streaming request failed with status $http_code"
		return 1
	fi
}

# Test OpenAI chat completions with different parameters
test_openai_chat_completions_parameters() {
	local payload
	payload=$(jq -n '{
        model: "gpt-4",
        messages: [
            {role: "system", content: "dry_run: true"},
            {role: "user", content: "Test"}
        ],
        temperature: 0.7,
        max_tokens: 100,
        top_p: 0.9,
        frequency_penalty: 0.5,
        presence_penalty: 0.3
    }')

	local response
	response=$(http_post "/openai/v1/chat/completions" "$payload")

	# Verify response structure
	assert_json_field "$response" "id"
	assert_json_equals "$response" "object" "chat.completion"
	assert_json_field "$response" "model"
}

# Test OpenAI error handling
test_openai_error_handling() {
	# Test with invalid model
	local payload
	payload=$(jq -n '{
        model: "invalid-model-name",
        messages: [
            {role: "user", content: "Test"}
        ]
    }')

	local response http_code
	response=$(http_post "/openai/v1/chat/completions" "$payload" "http_code" || true)
	http_code=$(echo "$response" | tail -n1)

	if [[ "$http_code" -lt 400 ]]; then
		log_error "Expected error status for invalid model, got $http_code"
		return 1
	fi
}

# Test OpenAI response format compliance
test_openai_response_format() {
	local payload
	payload=$(jq -n '{
        model: "gpt-3.5-turbo",
        messages: [
            {role: "system", content: "dry_run: true"},
            {role: "user", content: "Hello"}
        ]
    }')

	local response
	response=$(http_post "/openai/v1/chat/completions" "$payload")

	# Verify all required top-level fields
	local id object created model choices usage
	id=$(echo "$response" | jq -r '.id // empty')
	object=$(echo "$response" | jq -r '.object // empty')
	created=$(echo "$response" | jq -r '.created // empty')
	model=$(echo "$response" | jq -r '.model // empty')
	choices=$(echo "$response" | jq -r '.choices // empty')
	usage=$(echo "$response" | jq -r '.usage // empty')

	if [[ -z "$id" || -z "$object" || -z "$created" || -z "$model" || -z "$choices" || -z "$usage" ]]; then
		log_error "Response missing required OpenAI fields"
		return 1
	fi

	# Verify choice structure
	local choice_index choice_message choice_finish_reason
	choice_index=$(echo "$response" | jq -r '.choices[0].index // empty')
	choice_message=$(echo "$response" | jq -r '.choices[0].message // empty')
	choice_finish_reason=$(echo "$response" | jq -r '.choices[0].finish_reason // empty')

	if [[ -z "$choice_index" || -z "$choice_message" || -z "$choice_finish_reason" ]]; then
		log_error "Choice structure incomplete"
		return 1
	fi
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "OpenAI API Tests"
	run_test "List models" test_openai_list_models
	run_test "Chat completions" test_openai_chat_completions
	run_test "Chat completions streaming" test_openai_chat_completions_streaming
	run_test "Chat completions with parameters" test_openai_chat_completions_parameters
	run_test "Error handling" test_openai_error_handling
	run_test "Response format compliance" test_openai_response_format

	print_summary
}

main "$@"
