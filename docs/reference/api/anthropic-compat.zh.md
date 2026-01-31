# Anthropic 兼容 API

使用 Anthropic SDK 访问 clinvk。

## 概述

clinvk 提供 Anthropic 兼容端点，允许你使用 Anthropic SDK 与 CLI 后端。这支持与使用 Anthropic API 格式的现有应用程序集成。

## 基础 URL

```text
http://localhost:8080/anthropic/v1
```

## 认证

API Key 认证是可选的。如果配置了 Key，请包含以下之一：

- `Authorization: Bearer <key>`
- `X-Api-Key: <key>`

如果未配置 Key，则允许无认证请求。

## 端点

### POST /anthropic/v1/messages

创建消息（对话补全）。

**请求头：**

| 头 | 必填 | 说明 |
|--------|----------|-------------|
| `Content-Type` | 是 | `application/json` |
| `anthropic-version` | 是 | API 版本（例如 `2023-06-01`） |

**请求体：**

```json
{
  "model": "claude",
  "max_tokens": 1024,
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "system": "You are a helpful assistant."
}
```

**字段：**

| 字段 | 类型 | 必填 | 说明 |
|-------|------|----------|-------------|
| `model` | string | 是 | 后端选择器（见下方映射） |
| `max_tokens` | integer | 是 | 最大响应 token 数（当前被 CLI 后端忽略） |
| `messages` | array | 是 | 对话消息 |
| `system` | string | 否 | 系统提示词 |
| `temperature` | number | 否 | 采样温度（被忽略） |
| `top_p` | number | 否 | 核心采样（被忽略） |
| `top_k` | integer | 否 | Top-k 采样（被忽略） |
| `stop_sequences` | array | 否 | 停止序列（被忽略） |
| `stream` | boolean | 否 | 启用流式（SSE）时为 `true` |
| `metadata` | object | 否 | 请求元数据（被忽略） |

**响应：**

```json
{
  "id": "msg_abc123",
  "type": "message",
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Hello! How can I help you today?"}
  ],
  "model": "claude",
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 15
  }
}
```

**流式响应：**

当 `stream: true` 时，返回 Server-Sent Events (SSE)：

```text
event: message_start
data: {"type":"message_start","message":{"id":"msg_abc123","type":"message","role":"assistant","content":[],"model":"claude","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":15}}

event: message_stop
data: {"type":"message_stop"}
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

**建议：** 如需 Codex/Gemini，请显式使用 `codex` 或 `gemini`。

## 客户端示例

### Python（anthropic 包）

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"  # 仅在启用 API Key 时需要
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    system="You are a helpful coding assistant.",
    messages=[{"role": "user", "content": "Write a Python function"}]
)

print(message.content[0].text)
```

### TypeScript/JavaScript

```typescript
import Anthropic from '@anthropic-ai/sdk';

const client = new Anthropic({
  baseURL: 'http://localhost:8080/anthropic/v1',
  apiKey: 'not-needed'
});

const message = await client.messages.create({
  model: 'claude',
  max_tokens: 1024,
  messages: [{ role: 'user', content: 'Hello!' }]
});

console.log(message.content[0].text);
```

### cURL

```bash
curl -X POST http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### 流式示例（Python）

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

stream = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Tell me a story"}],
    stream=True
)

for text in stream.text_stream:
    print(text, end="", flush=True)
```

## 与 Anthropic API 的差异

| 功能 | Anthropic API | clinvk 兼容 |
|---------|---------------|-------------------|
| 模型 | Claude 模型 | Claude, Codex, Gemini |
| 补全 | 支持 | 未实现 |
| 嵌入 | 支持 | 未实现 |
| 图像 | 支持 | 未实现 |
| 工具 | 支持 | 未实现 |
| 错误格式 | Anthropic 模式 | RFC 7807 Problem Details |
| 会话 | 有状态 | 无状态（使用 REST API 获取会话） |

## 配置

Anthropic 兼容 API 使用相同的服务器配置：

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

- [OpenAI 兼容](openai-compat.md) - OpenAI SDK 兼容性
- [REST API](rest.md) - 原生 REST API 获取完整功能
- [serve 命令](../cli/serve.md) - 服务器配置
