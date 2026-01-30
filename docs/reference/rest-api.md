# Custom REST API ("/api/v1/*")

The custom REST API exposes all core clinvk capabilities with explicit backend selection.

Base URL (default): `http://localhost:8080`

All durations in API responses are in **milliseconds** (`*_ms`).

## Authentication

If API keys are configured, send one of:

- `X-Api-Key: <key>`
- `Authorization: Bearer <key>`

If no keys are configured, auth is disabled.

## POST /api/v1/prompt

Execute one prompt.

### Request

```json
{
  "backend": "claude",
  "prompt": "Explain this function",
  "model": "claude-opus-4-5-20251101",
  "workdir": "/abs/path",
  "approval_mode": "auto",
  "sandbox_mode": "workspace",
  "output_format": "json",
  "max_tokens": 0,
  "max_turns": 0,
  "system_prompt": "You are a strict reviewer",
  "verbose": false,
  "dry_run": false,
  "ephemeral": false,
  "extra": ["--add-dir", "/abs/path"],
  "metadata": {"request_id": "123"}
}
```

Notes:

- `output_format: stream-json` streams NDJSON **unified events**.
- `workdir` must be an absolute path and pass server restrictions.
- `extra` flags are validated per backend (allowlist).

### Response (JSON)

```json
{
  "session_id": "...",
  "backend": "claude",
  "exit_code": 0,
  "duration_ms": 1234,
  "output": "...",
  "error": "",
  "token_usage": {"input_tokens": 10, "output_tokens": 20}
}
```

## POST /api/v1/parallel

Execute multiple prompts concurrently.

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "Review"},
    {"backend": "codex", "prompt": "Optimize"}
  ],
  "max_parallel": 3,
  "fail_fast": false,
  "dry_run": false
}
```

Parallel tasks are always ephemeral. Response:

```json
{
  "total_tasks": 2,
  "completed": 2,
  "failed": 0,
  "total_duration_ms": 2000,
  "results": [ ... ]
}
```

## POST /api/v1/chain

Execute steps sequentially with `{{previous}}` substitution.

```json
{
  "steps": [
    {"backend": "claude", "prompt": "Analyze"},
    {"backend": "codex", "prompt": "Fix: {{previous}}"}
  ],
  "stop_on_failure": true,
  "pass_working_dir": false,
  "dry_run": false
}
```

Notes:

- Chain is always ephemeral.
- `pass_session_id` and `persist_sessions` are **not supported**.
- `{{session}}` placeholders are rejected.

## POST /api/v1/compare

Compare a prompt across backends.

```json
{
  "backends": ["claude", "gemini"],
  "prompt": "Explain this code",
  "model": "claude-opus-4-5-20251101",
  "workdir": "/abs/path",
  "sequential": false,
  "dry_run": false
}
```

Compare is always ephemeral.

## GET /api/v1/backends

Returns available backends:

```json
{ "backends": [{"name": "claude", "available": true}] }
```

## Sessions

- `GET /api/v1/sessions` (supports `backend`, `status`, `limit`, `offset`)
- `GET /api/v1/sessions/{id}`
- `DELETE /api/v1/sessions/{id}`

## GET /health

Returns status, uptime, backend availability, and session store health.

Example:

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "10m5s",
  "uptime_millis": 605000,
  "backends": [{"name": "claude", "available": true}],
  "session_store": {"available": true, "session_count": 42, "error": ""}
}
```
