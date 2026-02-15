#!/usr/bin/env bash
# Policy engine governance integration tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/../lib/common.sh"

POLICY_SERVER_PORT="${POLICY_SERVER_PORT:-18086}"
POLICY_SERVER_URL="http://${SERVER_HOST}:${POLICY_SERVER_PORT}"
POLICY_SERVER_PID=""
POLICY_SERVER_LOG=""

policy_http_get_status() {
	local path="$1"
	shift
	curl -s -w "\n%{http_code}" "${POLICY_SERVER_URL}${path}" "$@"
}

stop_policy_server() {
	if [[ -n "$POLICY_SERVER_PID" ]]; then
		log_info "Stopping policy server (PID: $POLICY_SERVER_PID)..."
		kill "$POLICY_SERVER_PID" 2>/dev/null || true
		wait "$POLICY_SERVER_PID" 2>/dev/null || true
		POLICY_SERVER_PID=""
	fi
}

start_policy_server() {
	local config_file="$1"
	local enable_forced_error="${2:-false}"
	POLICY_SERVER_LOG="${TEST_TEMP_DIR}/policy-server.log"
	: >"$POLICY_SERVER_LOG"

	log_info "Starting policy server on ${POLICY_SERVER_URL}..."
	if [[ "$enable_forced_error" == "true" ]]; then
		CLINVK_POLICY_TEST_FORCE_EVAL_ERROR=true "$CLINVK_BIN" --config "$config_file" serve --host "$SERVER_HOST" --port "$POLICY_SERVER_PORT" >"$POLICY_SERVER_LOG" 2>&1 &
	else
		"$CLINVK_BIN" --config "$config_file" serve --host "$SERVER_HOST" --port "$POLICY_SERVER_PORT" >"$POLICY_SERVER_LOG" 2>&1 &
	fi
	POLICY_SERVER_PID=$!

	local retries=30
	while ((retries > 0)); do
		if curl -sf "${POLICY_SERVER_URL}/health" >/dev/null 2>&1; then
			log_success "Policy server started (PID: $POLICY_SERVER_PID)"
			return 0
		fi
		sleep 0.5
		((retries--))
	done

	log_error "Policy server failed to start"
	log_error "Last server logs:"
	tail -n 20 "$POLICY_SERVER_LOG" || true
	stop_policy_server
	return 1
}

write_policy_config() {
	local config_path="$1"
	local rules_path="$2"
	local mode="$3"
	local failure_mode="${4:-fail-open}"
	cat >"$config_path" <<EOF
server:
  policy:
    enabled: true
    mode: ${mode}
    failure_mode: ${failure_mode}
    rules_file: "${rules_path}"
    explain_enabled: true
    default_decision: allow
    quota_store: memory
EOF
}

policy_fallback_count() {
	if [[ -z "$POLICY_SERVER_LOG" || ! -f "$POLICY_SERVER_LOG" ]]; then
		echo 0
		return
	fi
	grep -c "reason=engine_error" "$POLICY_SERVER_LOG" || true
}

test_startup_compile_gate_rejects_missing_rules_file() {
	local config_path="${TEST_TEMP_DIR}/policy-bad-config.yaml"
	cat >"$config_path" <<EOF
server:
  policy:
    enabled: true
    mode: shadow
    failure_mode: fail-open
    rules_file: "${TEST_TEMP_DIR}/missing-policy-rules.yaml"
    explain_enabled: false
    default_decision: allow
    quota_store: memory
EOF

	set +e
	local output
	output=$("$CLINVK_BIN" --config "$config_path" serve --host "$SERVER_HOST" --port "$POLICY_SERVER_PORT" 2>&1)
	local exit_code=$?
	set -e

	if [[ $exit_code -eq 0 ]]; then
		log_error "Expected startup failure for missing rules file"
		return 1
	fi
	if ! echo "$output" | grep -q "policy initialization failed"; then
		log_error "Expected policy initialization failure message, got: $output"
		return 1
	fi
}

test_enforce_mode_deny_contract() {
	local rules_path="${TEST_TEMP_DIR}/policy-deny-rules.yaml"
	local config_path="${TEST_TEMP_DIR}/policy-deny-config.yaml"

	cat >"$rules_path" <<EOF
version: v1
rules:
  - id: deny-backends
    enabled: true
    priority: 10
    selectors:
      path_prefix: /api/v1/backends
      methods: [GET]
    action:
      type: deny
EOF
	write_policy_config "$config_path" "$rules_path" "enforce"

	start_policy_server "$config_path"
	local output status_code body
	output=$(policy_http_get_status "/api/v1/backends")
	status_code=$(echo "$output" | tail -1)
	body=$(echo "$output" | sed '$d')
	stop_policy_server

	if [[ "$status_code" != "403" ]]; then
		log_error "Expected 403 policy deny, got $status_code"
		return 1
	fi
	assert_json_field "$body" "code"
	assert_json_field "$body" "request_id"
	assert_json_field "$body" "decision_id"
	assert_json_field "$body" "reason"

	local code reason
	code=$(echo "$body" | jq -r '.code')
	reason=$(echo "$body" | jq -r '.reason')
	if [[ "$code" != "policy_denied" ]]; then
		log_error "Expected policy_denied code, got $code"
		return 1
	fi
	if [[ "$reason" != "policy_deny" ]]; then
		log_error "Expected policy_deny reason, got $reason"
		return 1
	fi
}

test_shadow_mode_does_not_block() {
	local rules_path="${TEST_TEMP_DIR}/policy-shadow-rules.yaml"
	local config_path="${TEST_TEMP_DIR}/policy-shadow-config.yaml"

	cat >"$rules_path" <<EOF
version: v1
rules:
  - id: deny-backends
    enabled: true
    priority: 10
    selectors:
      path_prefix: /api/v1/backends
      methods: [GET]
    action:
      type: deny
EOF
	write_policy_config "$config_path" "$rules_path" "shadow"

	start_policy_server "$config_path"
	local output status_code body
	output=$(policy_http_get_status "/api/v1/backends" -H "X-Policy-Explain: true")
	status_code=$(echo "$output" | tail -1)
	body=$(echo "$output" | sed '$d')
	stop_policy_server

	if [[ "$status_code" != "200" ]]; then
		log_error "Expected 200 in shadow mode, got $status_code"
		return 1
	fi
	assert_json_field "$body" "backends"
}

test_quota_reject_contract() {
	local rules_path="${TEST_TEMP_DIR}/policy-quota-rules.yaml"
	local config_path="${TEST_TEMP_DIR}/policy-quota-config.yaml"

	cat >"$rules_path" <<EOF
version: v1
rules:
  - id: quota-backends
    enabled: true
    priority: 10
    selectors:
      path_prefix: /api/v1/backends
      methods: [GET]
    action:
      type: quota
      quota:
        rate_per_minute: 1
        scopes: [source]
EOF
	write_policy_config "$config_path" "$rules_path" "enforce"

	start_policy_server "$config_path"

	local first_output first_status
	first_output=$(policy_http_get_status "/api/v1/backends")
	first_status=$(echo "$first_output" | tail -1)
	if [[ "$first_status" != "200" ]]; then
		stop_policy_server
		log_error "Expected first quota request to pass with 200, got $first_status"
		return 1
	fi

	local second_output second_status second_body
	second_output=$(policy_http_get_status "/api/v1/backends")
	second_status=$(echo "$second_output" | tail -1)
	second_body=$(echo "$second_output" | sed '$d')
	stop_policy_server

	if [[ "$second_status" != "429" ]]; then
		log_error "Expected 429 for quota exceeded, got $second_status"
		return 1
	fi

	local code reason
	code=$(echo "$second_body" | jq -r '.code')
	reason=$(echo "$second_body" | jq -r '.reason')
	if [[ "$code" != "policy_quota_exceeded" ]]; then
		log_error "Expected policy_quota_exceeded code, got $code"
		return 1
	fi
	if [[ "$reason" != "rate_exceeded" ]]; then
		log_error "Expected rate_exceeded reason, got $reason"
		return 1
	fi
}

test_fail_open_fallback_allows_request() {
	local rules_path="${TEST_TEMP_DIR}/policy-fail-open-rules.yaml"
	local config_path="${TEST_TEMP_DIR}/policy-fail-open-config.yaml"

	cat >"$rules_path" <<EOF
version: v1
rules:
  - id: allow-backends
    enabled: true
    priority: 10
    selectors:
      path_prefix: /api/v1/backends
      methods: [GET]
    action:
      type: allow
EOF
	write_policy_config "$config_path" "$rules_path" "enforce" "fail-open"

	start_policy_server "$config_path" "true"
	local output status_code body
	output=$(policy_http_get_status "/api/v1/backends" -H "X-Policy-Test-Force-Eval-Error: true")
	status_code=$(echo "$output" | tail -1)
	body=$(echo "$output" | sed '$d')
	local fallback_count
	fallback_count=$(policy_fallback_count)
	stop_policy_server

	if [[ "$status_code" != "200" ]]; then
		log_error "Expected 200 in fail-open fallback, got $status_code"
		return 1
	fi
	assert_json_field "$body" "backends"
	if [[ "$fallback_count" -lt 1 ]]; then
		log_error "Expected at least one engine_error fallback audit event in fail-open mode"
		return 1
	fi
}

test_fail_closed_fallback_blocks_request() {
	local rules_path="${TEST_TEMP_DIR}/policy-fail-closed-rules.yaml"
	local config_path="${TEST_TEMP_DIR}/policy-fail-closed-config.yaml"

	cat >"$rules_path" <<EOF
version: v1
rules:
  - id: allow-backends
    enabled: true
    priority: 10
    selectors:
      path_prefix: /api/v1/backends
      methods: [GET]
    action:
      type: allow
EOF
	write_policy_config "$config_path" "$rules_path" "enforce" "fail-closed"

	start_policy_server "$config_path" "true"
	local output status_code body
	output=$(policy_http_get_status "/api/v1/backends" -H "X-Policy-Test-Force-Eval-Error: true")
	status_code=$(echo "$output" | tail -1)
	body=$(echo "$output" | sed '$d')
	local fallback_count
	fallback_count=$(policy_fallback_count)
	stop_policy_server

	if [[ "$status_code" != "503" ]]; then
		log_error "Expected 503 in fail-closed fallback, got $status_code"
		return 1
	fi
	assert_json_field "$body" "code"
	assert_json_field "$body" "request_id"
	assert_json_field "$body" "decision_id"
	assert_json_field "$body" "reason"

	local code reason
	code=$(echo "$body" | jq -r '.code')
	reason=$(echo "$body" | jq -r '.reason')
	if [[ "$code" != "policy_engine_unavailable" ]]; then
		log_error "Expected policy_engine_unavailable code, got $code"
		return 1
	fi
	if [[ "$reason" != "engine_error" ]]; then
		log_error "Expected engine_error reason, got $reason"
		return 1
	fi
	if [[ "$fallback_count" -lt 1 ]]; then
		log_error "Expected at least one engine_error fallback audit event in fail-closed mode"
		return 1
	fi
}

test_non_functional_budget_guardrails() {
	local rules_path="${TEST_TEMP_DIR}/policy-budget-rules.yaml"
	local config_path="${TEST_TEMP_DIR}/policy-budget-config.yaml"
	local latency_file="${TEST_TEMP_DIR}/policy-latency.txt"
	: >"$latency_file"

	cat >"$rules_path" <<EOF
version: v1
rules:
  - id: allow-backends
    enabled: true
    priority: 10
    selectors:
      path_prefix: /api/v1/backends
      methods: [GET]
    action:
      type: allow
EOF
	write_policy_config "$config_path" "$rules_path" "shadow" "fail-open"

	start_policy_server "$config_path"

	local total=20
	local i
	for ((i = 1; i <= total; i++)); do
		local result status latency
		result=$(curl -s -o /dev/null -w "%{http_code} %{time_total}" "${POLICY_SERVER_URL}/api/v1/backends")
		status=$(echo "$result" | awk '{print $1}')
		latency=$(echo "$result" | awk '{print $2}')
		if [[ "$status" != "200" ]]; then
			stop_policy_server
			log_error "Expected 200 during latency budget sampling, got $status"
			return 1
		fi
		echo "$latency" >>"$latency_file"
	done

	local count idx p95 fallback_count
	count=$(wc -l <"$latency_file")
	idx=$(( (count * 95 + 99) / 100 ))
	p95=$(sort -n "$latency_file" | sed -n "${idx}p")
	fallback_count=$(policy_fallback_count)
	stop_policy_server

	# Generous local budget for CI stability while still detecting obvious regressions.
	if ! awk -v value="$p95" -v budget="1.0" 'BEGIN { exit !(value <= budget) }'; then
		log_error "Policy latency p95 budget exceeded: p95=${p95}s budget=1.0s"
		return 1
	fi
	if [[ "$fallback_count" -ne 0 ]]; then
		log_error "Fallback-rate budget exceeded: expected 0 engine_error fallbacks in steady-state, got $fallback_count"
		return 1
	fi
}

main() {
	setup_test_env
	trap 'stop_policy_server; cleanup_test_env' EXIT INT TERM

	print_subheader "Policy Engine Governance"
	run_test "Startup compile gate rejects missing rules file" test_startup_compile_gate_rejects_missing_rules_file
	run_test "Enforce mode deny contract" test_enforce_mode_deny_contract
	run_test "Shadow mode does not block requests" test_shadow_mode_does_not_block
	run_test "Quota reject contract in enforce mode" test_quota_reject_contract
	run_test "Fail-open fallback allows request" test_fail_open_fallback_allows_request
	run_test "Fail-closed fallback blocks request" test_fail_closed_fallback_blocks_request
	run_test "Non-functional budget guardrails" test_non_functional_budget_guardrails

	print_summary
}

main "$@"
