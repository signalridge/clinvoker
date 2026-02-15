# Policy Governance (v0.7)

This guide describes how to operate the v0.7 policy engine safely in internal-first deployments.

## Scope

Policy governance adds centralized allow/deny/quota controls for HTTP and MCP HTTP requests.

- rule evaluation is deterministic (`priority -> specificity -> id`)
- decisions are auditable (`request_id`, `decision_id`, matched rules)
- rollout is controlled with `shadow` and `enforce` modes

## Response Contract

Policy rejections return stable machine-readable JSON:

```json
{
  "code": "policy_denied",
  "message": "request denied by policy",
  "request_id": "req-...",
  "decision_id": "dec-...",
  "reason": "policy_deny",
  "matched_rule_ids": ["deny-example"]
}
```

### Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `policy_denied` | `403` | Request matched a deny rule in `enforce` mode |
| `policy_quota_exceeded` | `429` | Request exceeded policy quota in `enforce` mode |
| `policy_engine_unavailable` | `503` | Policy evaluation failed and `failure_mode=fail-closed` |

### Reason Values

| Reason | Meaning |
|--------|---------|
| `policy_deny` | Explicit deny rule matched |
| `rate_exceeded` | Rate quota exceeded |
| `concurrency_exceeded` | In-flight quota exceeded |
| `token_budget_exceeded` | Token budget exceeded |
| `engine_error` | Runtime evaluation fallback |

## Explain Contract

Explain output is opt-in and only returned when both conditions are true:

1. `server.policy.explain_enabled: true`
2. request header `X-Policy-Explain: true`

Returned headers:

| Header | Description |
|--------|-------------|
| `X-Policy-Decision` | Final decision (`allow`, `deny`, `quota_reject`, `fallback`) |
| `X-Policy-Reason` | Normalized reason |
| `X-Policy-Decision-ID` | Decision trace id |
| `X-Policy-Matched-Rules` | Comma-separated matched rule ids |
| `X-Policy-Decision-Path` | Comma-separated evaluation trace steps |

Sensitive-field handling:

- raw API keys are never exposed by explain headers
- subject identity uses sanitized IDs only
- policy metrics and audit events use bounded/normalized label values

## Audit Contract

Each policy decision emits a structured audit log event (`policy_decision`) with:

| Field | Description |
|-------|-------------|
| `request_id` | Request correlation id |
| `decision_id` | Policy decision id |
| `decision` | Final decision |
| `mode` | `shadow` or `enforce` |
| `reason` | Normalized decision reason |
| `matched_rule_ids` | Matched rules (if any) |
| `status` | `allowed`, `blocked`, or `shadow` |
| `failure_mode` | Included for fallback events |
| `error` | Included for runtime fallback errors |
| `path`, `method` | Request context |

## Rollout Runbook

### Phase 1: Shadow Baseline

- configure `mode: shadow`, `failure_mode: fail-open`
- enable metrics and audit collection
- validate that expected requests produce expected decisions without blocking traffic

Rollback criteria:

- fallback events spike unexpectedly
- evaluation latency regresses beyond acceptable SLO

Action:

- set `server.policy.enabled: false` or remove problematic rules and restart

### Phase 2: Scoped Enforce

- move selected low-risk routes/identities to enforceable rules
- keep broad/default behavior in `allow` until confidence is high
- keep `failure_mode: fail-open` during early enforce rollout

Rollback criteria:

- false-positive deny/quota rejects on trusted internal callers
- elevated `policy_quota_exceeded` on normal traffic

Action:

- reduce scope selectors or revert to `mode: shadow`

### Phase 3: Broader Enforce

- expand selectors gradually (route by route, tenant by tenant)
- consider `failure_mode: fail-closed` only after fallback rate is near-zero and on-call procedures are ready

Rollback criteria:

- fail-closed fallback events appear in normal operation
- critical workflows impacted by policy enforcement

Action:

- switch to `fail-open` immediately, then narrow rule scope

## Minimal Operational Checklist

- `openspec validate --changes --strict`
- `go test ./...`
- `test/api/test_policy.sh`
- verify dashboards:
  - `clinvk_policy_decisions_total`
  - `clinvk_policy_eval_duration_seconds`
  - `clinvk_policy_fallback_total`
  - `clinvk_policy_quota_rejections_total`
