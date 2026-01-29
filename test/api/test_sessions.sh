#!/usr/bin/env bash
set -euo pipefail

# Source common test library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test listing sessions
test_list_sessions() {
	local response
	response=$(http_get "/api/v1/sessions")

	# Verify response structure
	assert_json_field "$response" "sessions"
	assert_json_field "$response" "total"
}

# Test listing sessions with pagination
test_list_sessions_pagination() {
	if ! skip_if_missing_backends "claude"; then
		return 0
	fi

	# Create a test session first
	local create_payload
	create_payload=$(jq -n '{
        backend: "claude",
        prompt: "Test session for pagination",
        dry_run: true
    }')
	http_post "/api/v1/prompt" "$create_payload" >/dev/null

	# Test with limit
	local response_limit
	response_limit=$(http_get "/api/v1/sessions?limit=5")
	assert_json_field "$response_limit" "sessions"

	local sessions_count
	sessions_count=$(echo "$response_limit" | jq -r '.sessions | length')
	if [[ "$sessions_count" -gt 5 ]]; then
		log_error "Expected at most 5 sessions with limit=5, got $sessions_count"
		return 1
	fi

	# Test with offset
	local response_offset
	response_offset=$(http_get "/api/v1/sessions?offset=1")
	assert_json_field "$response_offset" "sessions"

	# Test with both limit and offset
	local response_both
	response_both=$(http_get "/api/v1/sessions?limit=3&offset=1")
	assert_json_field "$response_both" "sessions"
}

# Test getting session detail
test_get_session_detail() {
	if ! skip_if_missing_backends "claude"; then
		return 0
	fi

	# Create a test session
	local create_payload
	create_payload=$(jq -n '{
        backend: "claude",
        prompt: "Test session for detail",
        dry_run: true
    }')

	local create_response
	create_response=$(http_post "/api/v1/prompt" "$create_payload")

	local session_id
	session_id=$(echo "$create_response" | jq -r '.session_id')

	if [[ -z "$session_id" || "$session_id" == "null" ]]; then
		log_error "Failed to create test session"
		return 1
	fi

	# Get session detail
	local response
	response=$(http_get "/api/v1/sessions/${session_id}")

	# Verify response structure
	assert_json_equals "$response" "id" "$session_id"
	assert_json_equals "$response" "backend" "claude"
	assert_json_field "$response" "created_at"
}

# Test deleting session
test_delete_session() {
	if ! skip_if_missing_backends "codex"; then
		return 0
	fi

	# Create a test session
	local create_payload
	create_payload=$(jq -n '{
        backend: "codex",
        prompt: "Test session for deletion",
        dry_run: true
    }')

	local create_response
	create_response=$(http_post "/api/v1/prompt" "$create_payload")

	local session_id
	session_id=$(echo "$create_response" | jq -r '.session_id')

	if [[ -z "$session_id" || "$session_id" == "null" ]]; then
		log_error "Failed to create test session"
		return 1
	fi

	# Delete session
	local response
	response=$(http_delete "/api/v1/sessions/${session_id}")

	# Verify deletion response (field is "deleted", not "success")
	assert_json_equals "$response" "deleted" "true"

	# Verify session is gone (should return 404)
	local full_response http_code
	full_response=$(http_get_status "/api/v1/sessions/${session_id}")
	http_code=$(echo "$full_response" | tail -n1)

	if [[ "$http_code" -ne 404 ]]; then
		log_error "Expected 404 for deleted session, got $http_code"
		return 1
	fi
}

# Test session lifecycle
test_session_lifecycle() {
	if ! skip_if_missing_backends "gemini"; then
		return 0
	fi

	# 1. Create session
	local create_payload
	create_payload=$(jq -n '{
        backend: "gemini",
        prompt: "Lifecycle test session",
        dry_run: true
    }')

	local create_response
	create_response=$(http_post "/api/v1/prompt" "$create_payload")
	local session_id
	session_id=$(echo "$create_response" | jq -r '.session_id')

	# 2. Verify it appears in list
	local list_response
	list_response=$(http_get "/api/v1/sessions")
	local found
	found=$(echo "$list_response" | jq -r --arg sid "$session_id" '[.sessions[].id] | any(. == $sid)')

	if [[ "$found" != "true" ]]; then
		log_error "Created session not found in list"
		return 1
	fi

	# 3. Get session detail
	local detail_response
	detail_response=$(http_get "/api/v1/sessions/${session_id}")
	assert_json_equals "$detail_response" "id" "$session_id"

	# 4. Delete session
	local delete_response
	delete_response=$(http_delete "/api/v1/sessions/${session_id}")
	assert_json_equals "$delete_response" "deleted" "true"

	# 5. Verify it's gone from list
	local final_list_response
	final_list_response=$(http_get "/api/v1/sessions")
	local still_found
	still_found=$(echo "$final_list_response" | jq -r --arg sid "$session_id" '[.sessions[].id] | any(. == $sid)')

	if [[ "$still_found" == "true" ]]; then
		log_error "Deleted session still appears in list"
		return 1
	fi
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Sessions API Tests"
	run_test "List sessions" test_list_sessions
	run_test "List sessions with pagination" test_list_sessions_pagination
	run_test "Get session detail" test_get_session_detail
	run_test "Delete session" test_delete_session
	run_test "Session lifecycle" test_session_lifecycle

	print_summary
}

main "$@"
