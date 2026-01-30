# 自定义 REST API（"/api/v1/*"）

自定义 REST API 提供完整的 clinvk 能力，并显式指定后端。

默认 Base URL：`http://localhost:8080`

API 响应中的耗时均为 **毫秒**（`*_ms`）。

## 鉴权

若配置了 API Key，可使用：

- `X-Api-Key: <key>`
- `Authorization: Bearer <key>`

未配置时不启用鉴权。

## POST /api/v1/prompt

执行单条 prompt。

### 请求

```json
{
  "backend": "claude",
  "prompt": "解释这个函数",
  "model": "claude-opus-4-5-20251101",
  "workdir": "/abs/path",
  "approval_mode": "auto",
  "sandbox_mode": "workspace",
  "output_format": "json",
  "max_tokens": 0,
  "max_turns": 0,
  "system_prompt": "你是严格的代码评审",
  "verbose": false,
  "dry_run": false,
  "ephemeral": false,
  "extra": ["--add-dir", "/abs/path"],
  "metadata": {"request_id": "123"}
}
```

说明：

- `output_format: stream-json` 会返回 NDJSON 的**统一事件流**。
- `workdir` 必须是绝对路径，并通过服务器工作目录限制。
- `extra` 参数会进行后端白名单校验。

### 响应（JSON）

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

并发执行多个 prompt。

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

并行任务始终无状态。响应：

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

顺序执行步骤并使用 `{{previous}}`。

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

说明：

- chain 始终无状态。
- `pass_session_id` 和 `persist_sessions` **不支持**。
- `{{session}}` 占位符会被拒绝。

## POST /api/v1/compare

在多后端对比同一 prompt。

```json
{
  "backends": ["claude", "gemini"],
  "prompt": "解释这段代码",
  "model": "claude-opus-4-5-20251101",
  "workdir": "/abs/path",
  "sequential": false,
  "dry_run": false
}
```

compare 始终无状态。

## GET /api/v1/backends

返回后端可用性：

```json
{ "backends": [{"name": "claude", "available": true}] }
```

## Sessions

- `GET /api/v1/sessions`（支持 `backend` / `status` / `limit` / `offset`）
- `GET /api/v1/sessions/{id}`
- `DELETE /api/v1/sessions/{id}`

## GET /health

返回状态、运行时长、后端可用性与会话存储健康。

示例：

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
