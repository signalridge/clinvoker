# REST API Reference

Complete reference for the clinvk custom REST API.

## Base URL

```yaml
http://localhost:8080/api/v1
```

## Authentication

The API does not require authentication by default. For production use, place behind a reverse proxy with authentication.

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
```yaml

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `backend` | string | Yes | Backend to use |
| `prompt` | string | Yes | The prompt |
| `model` | string | No | Model override |
| `workdir` | string | No | Working directory |
| `ephemeral` | boolean | No | Stateless mode (no session) |
| `approval_mode` | string | No | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | No | `default`, `read-only`, `workspace`, `full` |
| `output_format` | string | No | `default`, `text`, `json`, `stream-json` |
| `max_tokens` | integer | No | Maximum response tokens |
| `max_turns` | integer | No | Maximum agentic turns |
| `system_prompt` | string | No | System prompt |
| `verbose` | boolean | No | Enable verbose output |
| `dry_run` | boolean | No | Simulate execution |
| `extra` | array | No | Extra backend-specific flags |
| `metadata` | object | No | Custom metadata |

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

When `output_format` is `stream-json`, the endpoint streams NDJSON (`application/x-ndjson`).
Each line is a unified event:

```json
{"type":"init","backend":"claude","session_id":"...","content":{...}}
{"type":"message","backend":"claude","session_id":"...","content":{...}}
{"type":"done","backend":"claude","session_id":"...","content":{...}}
```yaml

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello world"}'
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
```yaml

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tasks` | array | Yes | List of task objects |
| `max_parallel` | integer | No | Max concurrent tasks |
| `fail_fast` | boolean | No | Stop on first failure |
| `dry_run` | boolean | No | Simulate execution |

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
    },
    {
      "backend": "codex",
      "exit_code": 0,
      "duration_ms": 1800,
      "output": "result 2"
    }
  ]
}
```

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
  "stop_on_failure": true,
  "pass_working_dir": false
}
```yaml

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `steps` | array | Yes | List of chain steps |
| `stop_on_failure` | boolean | No | Stop on first failure (default true) |
| `pass_working_dir` | boolean | No | Pass working directory between steps |

!!! note "Ephemeral Only"
    Chain execution is always ephemeral. Session linking and persistence are not supported.

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
    },
    {
      "step": 2,
      "name": "improve",
      "backend": "codex",
      "exit_code": 0,
      "duration_ms": 1500,
      "output": "improved code"
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
```yaml

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt` | string | Yes | The prompt |
| `backends` | array | Yes | Backends to compare |
| `model` | string | No | Model to use (if applicable) |
| `workdir` | string | No | Working directory |
| `sequential` | boolean | No | Run one at a time |
| `dry_run` | boolean | No | Simulate execution |

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
    },
    {
      "backend": "codex",
      "model": "o3",
      "exit_code": 0,
      "duration_ms": 3200,
      "output": "explanation from codex"
    }
  ]
}
```

---

## Backends

### GET /api/v1/backends

List available backends.

**Response:**

```json
{
  "backends": [
    {
      "name": "claude",
      "available": true
    },
    {
      "name": "codex",
      "available": true
    },
    {
      "name": "gemini",
      "available": false
    }
  ]
}
```yaml

---

## Sessions

### GET /api/v1/sessions

List sessions.

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `backend` | string | Filter by backend |
| `status` | string | Filter by status (`active`, `completed`, `error`) |
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
      "token_usage": {
        "input_tokens": 123,
        "output_tokens": 456
      },
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

**Response:**

```json
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
  "token_usage": {
    "input_tokens": 123,
    "output_tokens": 456
  },
  "tags": ["api"],
  "title": "Review auth changes"
}
```text

### DELETE /api/v1/sessions/{id}

Delete a session.

**Response:**

```json
{
  "deleted": true,
  "id": "abc123"
}
```

---

## Health Check

### GET /health

Server health status.

**Response:**

```json
{
  "status": "ok",
  "backends": [
    {"name": "claude", "available": true},
    {"name": "codex", "available": true},
    {"name": "gemini", "available": false}
  ]
}
```yaml

**Status Values:**

| Status | Description |
|--------|-------------|
| `ok` | All systems operational |
| `degraded` | Some backends unavailable |

---

## Metrics

### GET /metrics

Prometheus-compatible metrics endpoint (when `metrics_enabled: true` in config).

**Response:** Prometheus exposition format

```text
# HELP clinvk_requests_total Total HTTP requests

# TYPE clinvk_requests_total counter

clinvk_requests_total{method="POST",path="/api/v1/prompt",status="200"} 42

# HELP clinvk_request_duration_seconds HTTP request duration

# TYPE clinvk_request_duration_seconds histogram

clinvk_request_duration_seconds_bucket{path="/api/v1/prompt",le="0.1"} 5
...

# HELP clinvk_rate_limit_hits_total Rate limit hits

# TYPE clinvk_rate_limit_hits_total counter

clinvk_rate_limit_hits_total{ip="192.168.1.1"} 3

# HELP clinvk_sessions_total Total sessions

# TYPE clinvk_sessions_total gauge

clinvk_sessions_total 15
```

**Enable in config:**

```yaml
server:
  metrics_enabled: true
```

---

## Error Responses

### Rate Limiting (429)

When rate limiting is enabled and the limit is exceeded:

**Status:** `429 Too Many Requests`

**Headers:**

| Header | Description |
|--------|-------------|
| `Retry-After` | Seconds to wait before retry |

**Response:**

```json
{
  "title": "Too Many Requests",
  "status": 429,
  "detail": "Rate limit exceeded. Retry after 5 seconds."
}
```text

### Request Size Limit (413)

When request body exceeds `max_request_body_bytes`:

**Status:** `413 Request Entity Too Large`

**Response:**

```json
{
  "title": "Request Entity Too Large",
  "status": 413,
  "detail": "Request body exceeds maximum size of 10485760 bytes"
}
```

---

## OpenAPI Specification

### GET /openapi.json

Get the OpenAPI specification.

Returns the full OpenAPI 3.0 specification for the API.

---

## Error Responses

Execution failures are typically reported in the normal response body via `exit_code != 0` and the `error` field.

For request validation errors (for example, missing required fields), the server responds with non-2xx and an RFC 7807 Problem Details body (via Huma). Example:

```json
{
  "title": "Unprocessable Entity",
  "status": 422,
  "detail": "backend is required"
}
```
