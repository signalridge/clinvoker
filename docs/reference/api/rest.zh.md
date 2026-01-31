# REST API 参考

clinvk 自定义 REST API 的完整参考。

## Base URL

```text
http://localhost:8080/api/v1
```text

## 认证

API Key 认证为可选项。配置了 Key 后，所有请求需携带：

- `X-Api-Key: <key>`
- 或 `Authorization: Bearer <key>`

Key 可通过 `CLINVK_API_KEYS`（逗号分隔）或 `server.api_keys_gopass_path`（gopass）提供。

---

## 执行提示词

### POST /api/v1/prompt

执行一次提示词。

**请求体：**

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
```text

**字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `backend` | string | 是 | 后端名称 |
| `prompt` | string | 是 | 提示词 |
| `model` | string | 否 | 模型覆盖 |
| `workdir` | string | 否 | 工作目录（绝对路径，受允许/阻止前缀限制） |
| `ephemeral` | boolean | 否 | 无状态（不保存会话） |
| `approval_mode` | string | 否 | `default` / `auto` / `none` / `always` |
| `sandbox_mode` | string | 否 | `default` / `read-only` / `workspace` / `full` |
| `output_format` | string | 否 | `default` / `text` / `json` / `stream-json` |
| `max_tokens` | integer | 否 | 最大 token（当前不映射到后端参数） |
| `max_turns` | integer | 否 | 最大回合数 |
| `system_prompt` | string | 否 | 系统提示词 |
| `verbose` | boolean | 否 | 详细输出 |
| `dry_run` | boolean | 否 | 模拟执行 |
| `extra` | array | 否 | 额外后端参数 |
| `metadata` | object | 否 | 写入会话的自定义元数据 |

**响应：**

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
```text

**流式响应（`output_format: "stream-json"`）：**

使用 NDJSON（`application/x-ndjson`）输出统一事件。例如（简化）：

```json
{"type":"init","backend":"claude","session_id":"...","content":{"model":"..."}}
{"type":"message","backend":"claude","session_id":"...","content":{"text":"..."}}
{"type":"done","backend":"claude","session_id":"..."}
```yaml

---

## 并行执行

### POST /api/v1/parallel

并行执行多个任务。

请求体示例：

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "task 1"},
    {"backend": "codex", "prompt": "task 2"}
  ],
  "max_parallel": 3,
  "fail_fast": false
}
```text

每个任务与 `/api/v1/prompt` 使用相同字段。

**响应：**

```json
{
  "total_tasks": 2,
  "completed": 2,
  "failed": 0,
  "total_duration_ms": 2000,
  "results": [
    {"backend": "claude", "exit_code": 0, "duration_ms": 2000, "output": "result 1"}
  ]
}
```text

> 并行任务始终为无状态，`session_id` 可能为空。

---

## 串行执行

### POST /api/v1/chain

顺序执行流水线。

```json
{
  "steps": [
    {"name": "analyze", "backend": "claude", "prompt": "analyze the code"},
    {"name": "improve", "backend": "codex", "prompt": "improve based on: {{previous}}"}
  ],
  "stop_on_failure": false,
  "pass_working_dir": false
}
```text

> API 默认 `stop_on_failure=false`。

---

## 后端对比

### POST /api/v1/compare

比较多个后端输出。

```json
{
  "prompt": "explain this algorithm",
  "backends": ["claude", "codex", "gemini"],
  "sequential": false
}
```yaml

---

## 会话

### GET /api/v1/sessions

支持 `backend` / `status` / `limit` / `offset` 查询，`status` 可为 `active` / `completed` / `error` / `paused`。

---

## 健康检查

### GET /health

返回 `status`、`version`、`uptime`、后端可用性与会话存储状态。

---

## 错误响应

### 401 未授权

当开启 API Key 且未提供/无效：

```json
{
  "error": "unauthorized",
  "message": "missing API key"
}
```text

---

## OpenAPI

### GET /openapi.json

返回 OpenAPI 3.0 规范。
