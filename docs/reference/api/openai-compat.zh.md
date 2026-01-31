# OpenAI 兼容 API

使用 OpenAI SDK 访问 clinvk。

## 概述

clinvk 提供 OpenAI 兼容端点，允许你使用 OpenAI SDK 与 CLI 后端。这支持与使用 OpenAI API 格式的现有应用程序集成。

## 基础 URL

```text
http://localhost:8080/openai/v1
```

## 认证

API Key 认证是可选的。如果配置了 Key，请包含以下之一：

- `Authorization: Bearer <key>`
- `X-Api-Key: <key>`

如果未配置 Key，则允许无认证请求。

## 端点

### GET /openai/v1/models

列出可用模型（后端映射为模型）。

**响应：**

```json
{
  "object": "list",
  "data": [
    {
      "id": "claude",
      "object": "model",
      "created": 1704067200,
      "owned_by": "anthropic"
    },
    {
      "id": "codex",
      "object": "model",
      "created": 1704067200,
      "owned_by": "openai"
    },
    {
      "id": "gemini",
      "object": "model",
      "created": 1704067200,
      "owned_by": "google"
    }
  ]
}
```

### POST /openai/v1/chat/completions

创建对话补全。

**请求体：**

```json
{
  "model": "claude",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 4096,
  "temperature": 0.7,
  "stream": false
}
```

**字段：**

| 字段 | 类型 | 必填 | 说明 |
|-------|------|----------|-------------|
| `model` | string | 是 | 后端选择器（见下方映射） |
| `messages` | array | 是 | 对话消息 |
| `max_tokens` | integer | 否 | 最大响应 token 数（当前被 CLI 后端忽略） |
| `temperature` | number | 否 | 采样温度（被忽略） |
| `top_p` | number | 否 | 核心采样（被忽略） |
| `n` | integer | 否 | 补全数量（被忽略，始终为 1） |
| `stream` | boolean | 否 | 启用流式（SSE）时为 `true` |
| `stop` | string/array | 否 | 停止序列（被忽略） |
| `presence_penalty` | number | 否 | 存在惩罚（被忽略） |
| `frequency_penalty` | number | 否 | 频率惩罚（被忽略） |
| `logit_bias` | object | 否 | Logit 偏置（被忽略） |
| `user` | string | 否 | 用户标识符（被忽略） |

**响应：**

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1704067200,
  "model": "claude",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 15,
    "total_tokens": 25
  }
}
```

**流式响应：**

当 `stream: true` 时，返回 Server-Sent Events (SSE)：

```text
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

## 模型映射

`model` 字段决定使用哪个后端：

| 模型值 | 使用的后端 |
|-------------|--------------|
| `claude` | Claude |
| `codex` | Codex |
| `gemini` | Gemini |
| 包含 `claude` | Claude |
| 包含 `gpt` | Codex |
| 包含 `gemini` | Gemini |
| 其他 | Claude（默认） |

**建议：** 使用精确的后端名称（`codex`、`claude`、`gemini`）以避免歧义。

## 客户端示例

### Python（openai 包）

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # 仅在启用 API Key 时需要
)

response = client.chat.completions.create(
    model="claude",
    messages=[
        {"role": "system", "content": "You are a helpful coding assistant."},
        {"role": "user", "content": "Write a hello world in Python"}
    ]
)

print(response.choices[0].message.content)
```

### TypeScript/JavaScript

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/openai/v1',
  apiKey: 'not-needed'
});

const response = await client.chat.completions.create({
  model: 'claude',
  messages: [{ role: 'user', content: 'Hello!' }]
});

console.log(response.choices[0].message.content);
```

### cURL

```bash
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "claude", "messages": [{"role": "user", "content": "Hello!"}]}'
```

### 流式示例（Python）

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Tell me a story"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content is not None:
        print(chunk.choices[0].delta.content, end="")
```

## 与 OpenAI API 的差异

| 功能 | OpenAI API | clinvk 兼容 |
|---------|------------|-------------------|
| 模型 | GPT-3.5, GPT-4 | Claude, Codex, Gemini |
| 补全 | 支持 | 未实现 |
| 嵌入 | 支持 | 未实现 |
| 图像 | 支持 | 未实现 |
| 音频 | 支持 | 未实现 |
| 错误格式 | OpenAI 模式 | RFC 7807 Problem Details |
| 会话 | 有状态 | 无状态（使用 REST API 获取会话） |

## 配置

OpenAI 兼容 API 使用相同的服务器配置：

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
```

## 错误响应

错误遵循 RFC 7807 Problem Details 格式：

```json
{
  "type": "https://api.clinvk.dev/errors/backend-not-found",
  "title": "Backend Not Found",
  "status": 400,
  "detail": "The requested backend 'unknown' is not available"
}
```

## 下一步

- [Anthropic 兼容](anthropic-compat.md) - Anthropic SDK 兼容性
- [REST API](rest.md) - 原生 REST API 获取完整功能
- [serve 命令](../cli/serve.md) - 服务器配置
