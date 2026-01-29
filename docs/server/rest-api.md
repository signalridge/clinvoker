# REST API Reference

Complete reference for the clinvk custom REST API.

## Base URL

```
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
```

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
```

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
```

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
      "name": "analyze",
      "backend": "claude",
      "exit_code": 0,
      "duration_ms": 2000,
      "output": "analysis result"
    },
    {
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
```

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
      "available": true,
      "enabled": true,
      "default_model": "claude-opus-4-5-20251101"
    },
    {
      "name": "codex",
      "available": true,
      "enabled": true,
      "default_model": "o3"
    },
    {
      "name": "gemini",
      "available": false,
      "enabled": true,
      "default_model": "gemini-2.5-pro"
    }
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
| `limit` | integer | Maximum results |
| `offset` | integer | Pagination offset |

**Response:**

```json
{
  "sessions": [
    {
      "id": "abc123",
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "created_at": "2025-01-27T10:00:00Z",
      "updated_at": "2025-01-27T11:30:00Z",
      "workdir": "/projects/myapp"
    }
  ],
  "total": 42
}
```

### GET /api/v1/sessions/{id}

Get session details.

**Response:**

```json
{
  "id": "abc123",
  "backend": "claude",
  "model": "claude-opus-4-5-20251101",
  "created_at": "2025-01-27T10:00:00Z",
  "updated_at": "2025-01-27T11:30:00Z",
  "workdir": "/projects/myapp",
  "metadata": {
    "tokens_input": 1234,
    "tokens_output": 5678
  }
}
```

### DELETE /api/v1/sessions/{id}

Delete a session.

**Response:**

```json
{
  "deleted": true
}
```

---

## Health Check

### GET /health

Server health status.

**Response:**

```json
{
  "status": "ok"
}
```

---

## OpenAPI Specification

### GET /openapi.json

Get the OpenAPI specification.

Returns the full OpenAPI 3.0 specification for the API.

---

## Error Responses

Errors follow this format:

```json
{
  "error": {
    "code": "BACKEND_UNAVAILABLE",
    "message": "Backend 'codex' is not available",
    "details": {}
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | Malformed request |
| `BACKEND_UNAVAILABLE` | 503 | Backend not available |
| `SESSION_NOT_FOUND` | 404 | Session doesn't exist |
| `EXECUTION_ERROR` | 500 | Backend execution failed |
| `TIMEOUT` | 504 | Request timed out |
