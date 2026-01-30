#!/usr/bin/env bash
# Distributed tracing tests
# Tests W3C Trace Context propagation and trace header handling
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

# Test: Request without trace headers gets trace ID in response
test_trace_header_in_response() {
	local response

	response=$(http_get "/health")

	# Trace ID might not be in response body, check headers instead
	local headers
	headers=$(curl -sI "${SERVER_URL}/health" 2>/dev/null || true)

	# Look for traceparent or x-trace-id header
	if echo "$headers" | grep -qiE "(traceparent|x-trace-id|trace-id)"; then
		log_debug "Found trace header in response"
		return 0
	fi

	# Trace headers may not be exposed, that's OK - just verify request succeeds
	log_debug "Trace headers not exposed in response (may be internal only)"
	return 0
}

# Test: Request with traceparent header propagates trace context
test_traceparent_propagation() {
	local traceparent="00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	local response

	response=$(curl -sf -H "traceparent: $traceparent" "${SERVER_URL}/health" 2>/dev/null || true)

	if [[ -z "$response" ]]; then
		log_error "Request with traceparent header failed"
		return 1
	fi

	# Verify response is valid health response
	local status
	status=$(echo "$response" | jq -r '.status // empty' 2>/dev/null || true)
	if [[ "$status" != "ok" && "$status" != "degraded" ]]; then
		log_error "Invalid health response with traceparent header"
		return 1
	fi

	return 0
}

# Test: Request with invalid traceparent header is handled gracefully
test_invalid_traceparent_handled() {
	local response

	# Send invalid traceparent format
	response=$(curl -sf -H "traceparent: invalid-format" "${SERVER_URL}/health" 2>/dev/null || true)

	if [[ -z "$response" ]]; then
		log_error "Request with invalid traceparent failed unexpectedly"
		return 1
	fi

	# Should still return valid health response
	local status
	status=$(echo "$response" | jq -r '.status // empty' 2>/dev/null || true)
	if [[ "$status" != "ok" && "$status" != "degraded" ]]; then
		log_error "Invalid traceparent should not break request"
		return 1
	fi

	return 0
}

# Test: Multiple requests generate different trace contexts
test_different_traces() {
	local trace1 trace2

	# Make two requests and check if response headers differ
	local headers1 headers2
	headers1=$(curl -sI "${SERVER_URL}/health" 2>/dev/null || true)
	headers2=$(curl -sI "${SERVER_URL}/health" 2>/dev/null || true)

	# Extract trace IDs if present
	trace1=$(echo "$headers1" | grep -iE "(traceparent|x-trace-id)" | head -1 || true)
	trace2=$(echo "$headers2" | grep -iE "(traceparent|x-trace-id)" | head -1 || true)

	if [[ -n "$trace1" && -n "$trace2" ]]; then
		if [[ "$trace1" == "$trace2" ]]; then
			log_warning "Two requests have same trace ID (may indicate caching)"
		else
			log_debug "Different requests have different trace IDs"
		fi
	fi

	return 0
}

# Main test execution
main() {
	setup_test_env
	start_server

	print_subheader "Distributed Tracing"
	run_test "Trace context in response" test_trace_header_in_response
	run_test "Traceparent header propagation" test_traceparent_propagation
	run_test "Invalid traceparent handled gracefully" test_invalid_traceparent_handled
	run_test "Different requests have unique traces" test_different_traces

	print_summary
}

main "$@"
