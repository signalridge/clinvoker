# REST API 参考

clinvk 自定义 REST API 完整参考。

## 基础 URL

```
http://localhost:8080/api/v1
```

## 认证

API 默认不需要认证。生产环境使用时，请放在带认证的反向代理后面。

---

## 提示执行

### POST /api/v1/prompt

执行单个提示。

**请求体：**

```json
{
  "backend": "claude",
  "prompt": "解释这段代码",
  "model": "claude-opus-4-5-20251101",
  "workdir": "/path/to/project",
  "ephemeral": false,
  "approval_mode": "auto",
  "sandbox_mode": "workspace",
  "output_format": "json",
  "max_tokens": 4096,
  "max_turns": 10,
  "system_prompt": "你是一个有帮助的助手。",
  "verbose": false,
  "dry_run": false,
  "extra": ["--some-flag"],
  "metadata": {"project": "demo"}
}
```

**字段：**

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `backend` | string | 是 | 使用的后端 |
| `prompt` | string | 是 | 提示内容 |
| `model` | string | 否 | 模型覆盖 |
| `workdir` | string | 否 | 工作目录 |
| `ephemeral` | boolean | 否 | 无状态模式（不创建会话） |
| `approval_mode` | string | 否 | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | 否 | `default`, `read-only`, `workspace`, `full` |
| `output_format` | string | 否 | `default`, `text`, `json`, `stream-json` |
| `max_tokens` | integer | 否 | 最大响应 token 数 |
| `max_turns` | integer | 否 | 最大代理轮次 |
| `system_prompt` | string | 否 | 系统提示 |
| `verbose` | boolean | 否 | 启用详细输出 |
| `dry_run` | boolean | 否 | 模拟执行 |
| `extra` | array | 否 | 额外后端参数 |
| `metadata` | object | 否 | 自定义元数据 |

**响应：**

```json
{
  "session_id": "abc123",
  "backend": "claude",
  "exit_code": 0,
  "duration_ms": 2500,
  "output": "代码解释...",
  "token_usage": {
    "input_tokens": 123,
    "output_tokens": 456,
    "cached_tokens": 0,
    "reasoning_tokens": 0
  }
}
```

---

## 并行执行

### POST /api/v1/parallel

并行执行多个任务。

**请求体：**

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "任务 1"
    },
    {
      "backend": "codex",
      "prompt": "任务 2"
    }
  ],
  "max_parallel": 3,
  "fail_fast": false
}
```

**字段：**

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `tasks` | array | 是 | 任务对象列表 |
| `max_parallel` | integer | 否 | 最大并发任务数 |
| `fail_fast` | boolean | 否 | 第一个失败时停止 |
| `dry_run` | boolean | 否 | 模拟执行 |

**响应：**

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
      "output": "结果 1"
    }
  ]
}
```

---

## 链式执行

### POST /api/v1/chain

执行顺序管道。

**请求体：**

```json
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "分析代码"
    },
    {
      "name": "improve",
      "backend": "codex",
      "prompt": "基于以下内容改进：{{previous}}"
    }
  ],
  "stop_on_failure": true,
  "pass_working_dir": false
}
```

**字段：**

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `steps` | array | 是 | 步骤列表 |
| `stop_on_failure` | boolean | 否 | 失败即停止（默认 true） |
| `pass_working_dir` | boolean | 否 | 在步骤间传递工作目录 |

!!! note "仅临时模式"
    链式执行始终为临时模式，不支持会话关联或持久化。

**响应：**

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
      "output": "分析结果",
      "duration_ms": 2000,
      "exit_code": 0
    }
  ]
}
```

---

## 后端对比

### POST /api/v1/compare

比较多个后端的响应。

**请求体：**

```json
{
  "prompt": "解释这个算法",
  "backends": ["claude", "codex", "gemini"],
  "sequential": false
}
```

**字段：**

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `prompt` | string | 是 | 提示内容 |
| `backends` | array | 是 | 要对比的后端 |
| `model` | string | 否 | 模型（如适用） |
| `workdir` | string | 否 | 工作目录 |
| `sequential` | boolean | 否 | 顺序执行 |
| `dry_run` | boolean | 否 | 模拟执行 |

**响应：**

```json
{
  "prompt": "解释这个算法",
  "backends": ["claude", "codex", "gemini"],
  "total_duration_ms": 3200,
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "exit_code": 0,
      "duration_ms": 2500,
      "output": "来自 claude 的解释"
    },
    {
      "backend": "codex",
      "model": "o3",
      "exit_code": 0,
      "duration_ms": 3200,
      "output": "来自 codex 的解释"
    }
  ]
}
```

---

## 后端

### GET /api/v1/backends

列出可用后端。

**响应：**

```json
{
  "backends": [
    {
      "name": "claude",
      "available": true,
      "enabled": true,
      "default_model": "claude-opus-4-5-20251101"
    }
  ]
}
```

---

## 会话

### GET /api/v1/sessions

列出会话。

### GET /api/v1/sessions/{id}

获取会话详情。

### DELETE /api/v1/sessions/{id}

删除会话。

---

## 健康检查

### GET /health

服务器健康状态。

**响应：**

```json
{
  "status": "ok"
}
```

---

## 错误响应

错误遵循以下格式：

```json
{
  "error": {
    "code": "BACKEND_UNAVAILABLE",
    "message": "后端 'codex' 不可用",
    "details": {}
  }
}
```

### 错误码

| 代码 | HTTP 状态 | 描述 |
|------|-----------|------|
| `INVALID_REQUEST` | 400 | 请求格式错误 |
| `BACKEND_UNAVAILABLE` | 503 | 后端不可用 |
| `SESSION_NOT_FOUND` | 404 | 会话不存在 |
| `EXECUTION_ERROR` | 500 | 后端执行失败 |
| `TIMEOUT` | 504 | 请求超时 |
