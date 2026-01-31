# REST API Reference

Complete reference for the clinvk custom REST API.

## Base URL

```text
http://localhost:8080/api/v1
```

## Authentication

API key auth is **optional**. If keys are configured, every request must include one of:

- `X-Api-Key: <key>`
- `Authorization: Bearer <key>`

Keys can be provided via `CLINVK_API_KEYS` (comma-separated) or `server.api_keys_gopass_path` (gopass).

---

## Prompt Execution

### POST /api/v1/prompt

Execute a single prompt.

**Request Body:**

```json
{
  "backend": "claude",
  "prompt": "explain this code",
  "model": "claude-opus-4-5-20251101",
  "workdir": "/path/to/project",
  "ephemeral": false,
  "approval_mode": "auto",
  "sandbox_mode": "workspace",
  "output_format": "json",
  "max_tokens": 4096,
  "max_turns": 10,
  "system_prompt": "You are a helpful assistant.",
  "verbose": false,
  "dry_run": false,
  "extra": ["--some-flag"],
  "metadata": {"project": "demo"}
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `backend` | string | Yes | Backend to use |
| `prompt` | string | Yes | The prompt |
| `model` | string | No | Model override |
| `workdir` | string | No | Working directory (absolute path; validated against allowed/blocked prefixes) |
| `ephemeral` | boolean | No | Stateless mode (no session) |
| `approval_mode` | string | No | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | No | `default`, `read-only`, `workspace`, `full` |
| `output_format` | string | No | `default`, `text`, `json`, `stream-json` |
| `max_tokens` | integer | No | Maximum response tokens (not mapped to backend flags yet) |
| `max_turns` | integer | No | Maximum agentic turns |
| `system_prompt` | string | No | System prompt |
| `verbose` | boolean | No | Enable verbose output |
| `dry_run` | boolean | No | Simulate execution |
| `extra` | array | No | Extra backend-specific flags |
| `metadata` | object | No | Custom metadata stored with session |

**Response:**

```json
{
  "session_id": "abc123",
  "backend": "claude",
  "exit_code": 0,
  "duration_ms": 2500,
  "output": "The code explanation...",
  "token_usage": {
    "input_tokens": 123,
    "output_tokens": 456,
    "cached_tokens": 0,
    "reasoning_tokens": 0
  }
}
```

**Streaming Response (`output_format: "stream-json"`):**

Streams NDJSON (`application/x-ndjson`) of unified events. Example (structure abbreviated):

```json
{"type":"init","backend":"claude","session_id":"...","content":{"model":"..."}}
{"type":"message","backend":"claude","session_id":"...","content":{"text":"..."}}
{"type":"done","backend":"claude","session_id":"..."}
```

---

## Parallel Execution

### POST /api/v1/parallel

Execute multiple tasks in parallel.

**Request Body:**

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "task 1"
    },
    {
      "backend": "codex",
      "prompt": "task 2"
    }
  ],
  "max_parallel": 3,
  "fail_fast": false
}
```

Each task accepts the same fields as `/api/v1/prompt` (including `workdir`, `approval_mode`, `output_format`, etc.).

**Response:**

```json
{
  "total_tasks": 2,
  "completed": 2,
  "failed": 0,
  "total_duration_ms": 2000,
  "results": [
    {
      "backend": "claude",
      "exit_code": 0,
      "duration_ms": 2000,
      "output": "result 1"
    }
  ]
}
```

> Parallel tasks are always ephemeral; `session_id` may be omitted.

---

## Chain Execution

### POST /api/v1/chain

Execute a sequential pipeline.

**Request Body:**

```json
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "analyze the code"
    },
    {
      "name": "improve",
      "backend": "codex",
      "prompt": "improve based on: {{previous}}"
    }
  ],
  "stop_on_failure": false,
  "pass_working_dir": false
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `steps` | array | Yes | List of chain steps |
| `stop_on_failure` | boolean | No | Stop on first failure (default `false` for API) |
| `pass_working_dir` | boolean | No | Pass working directory between steps |

> Chain execution is always ephemeral. `pass_session_id` and `persist_sessions` are not supported.

**Response:**

```json
{
  "total_steps": 2,
  "completed_steps": 2,
  "failed_step": 0,
  "total_duration_ms": 3500,
  "results": [
    {
      "step": 1,
      "name": "analyze",
      "backend": "claude",
      "exit_code": 0,
      "duration_ms": 2000,
      "output": "analysis result"
    }
  ]
}
```

---

## Backend Comparison

### POST /api/v1/compare

Compare responses from multiple backends.

**Request Body:**

```json
{
  "prompt": "explain this algorithm",
  "backends": ["claude", "codex", "gemini"],
  "sequential": false
}
```

**Response:**

```json
{
  "prompt": "explain this algorithm",
  "backends": ["claude", "codex", "gemini"],
  "total_duration_ms": 3200,
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "exit_code": 0,
      "duration_ms": 2500,
      "output": "explanation from claude"
    }
  ]
}
```

> Compare runs are ephemeral; `session_id` may be omitted.

---

## Backends

### GET /api/v1/backends

List available backends.

**Response:**

```json
{
  "backends": [
    {"name": "claude", "available": true},
    {"name": "codex", "available": true},
    {"name": "gemini", "available": false}
  ]
}
```

---

## Sessions

### GET /api/v1/sessions

List sessions.

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `backend` | string | Filter by backend |
| `status` | string | Filter by status (`active`, `completed`, `error`, `paused`) |
| `limit` | integer | Maximum results |
| `offset` | integer | Pagination offset |

**Response:**

```json
{
  "sessions": [
    {
      "id": "abc123",
      "backend": "claude",
      "created_at": "2025-01-27T10:00:00Z",
      "last_used": "2025-01-27T11:30:00Z",
      "working_dir": "/projects/myapp",
      "model": "claude-opus-4-5-20251101",
      "initial_prompt": "Review auth changes",
      "status": "active",
      "turn_count": 3,
      "token_usage": {"input_tokens": 123, "output_tokens": 456},
      "tags": ["api"],
      "title": "Review auth changes"
    }
  ],
  "total": 42,
  "limit": 100,
  "offset": 0
}
```

### GET /api/v1/sessions/{id}

Get session details.

### DELETE /api/v1/sessions/{id}

Delete a session.

---

## Health Check

### GET /health

Server health status.

**Response (abridged):**

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "2m31s",
  "uptime_millis": 151000,
  "backends": [
    {"name": "claude", "available": true}
  ],
  "session_store": {
    "available": true,
    "session_count": 15
  }
}
```

**Status Values:**

| Status | Description |
|--------|-------------|
| `ok` | All systems operational |
| `degraded` | Some backends unavailable |
| `unhealthy` | Session store unavailable |

---

## Metrics

### GET /metrics

Prometheus-compatible metrics endpoint (when `metrics_enabled: true` in config).

---

## Error Responses

### Unauthorized (401)

If API keys are configured and missing/invalid:

```json
{
  "error": "unauthorized",
  "message": "missing API key"
}
```

### Rate Limiting (429)

When rate limiting is enabled and the limit is exceeded.

### Request Size Limit (413)

When request body exceeds `max_request_body_bytes`.

---

## OpenAPI Specification

### GET /openapi.json

Returns the OpenAPI 3.0 specification for the API.
