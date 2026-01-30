# Anthropic 兼容 API

使用 clinvk 与现有的 Anthropic 客户端库和工具。

## 概述

clinvk 提供 Anthropic 兼容的端点，允许您使用 Anthropic Python SDK 和其他兼容客户端。

## 基础 URL

```yaml
http://localhost:8080/anthropic/v1
```

## 端点

### POST /anthropic/v1/messages

创建消息（聊天补全）。

**请求头：**

| 请求头 | 必需 | 描述 |
|--------|------|------|
| `Content-Type` | 是 | `application/json` |
| `anthropic-version` | 是 | API 版本（如 `2023-06-01`） |

**请求体：**

```json
{
  "model": "claude",
  "max_tokens": 1024,
  "messages": [
    {"role": "user", "content": "你好！"}
  ],
  "system": "你是一个有帮助的助手。"
}
```yaml

**字段：**

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `model` | string | 是 | 后端名称 (claude, codex, gemini) |
| `max_tokens` | integer | 是 | 最大响应 token 数 |
| `messages` | array | 是 | 聊天消息 |
| `system` | string | 否 | 系统提示 |

**响应：**

```json
{
  "id": "msg_abc123",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "你好！今天我能帮你什么？"
    }
  ],
  "model": "claude",
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 15
  }
}
```

## 客户端示例

### Python (anthropic 包)

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"  # clinvk 不需要 API key
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    system="你是一个有帮助的编程助手。",
    messages=[
        {"role": "user", "content": "写一个排序列表的 Python 函数"}
    ]
)

print(message.content[0].text)
```text

### 带对话历史

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

messages = [
    {"role": "user", "content": "什么是 Python？"},
]

# 第一条消息
response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=messages
)

# 将助手响应添加到历史
messages.append({
    "role": "assistant",
    "content": response.content[0].text
})

# 继续对话
messages.append({
    "role": "user",
    "content": "给我看一个 hello world 示例"
})

response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=messages
)

print(response.content[0].text)
```

### TypeScript/JavaScript

```typescript
import Anthropic from '@anthropic-ai/sdk';

const client = new Anthropic({
  baseURL: 'http://localhost:8080/anthropic/v1',
  apiKey: 'not-needed'
});

async function main() {
  const message = await client.messages.create({
    model: 'claude',
    max_tokens: 1024,
    messages: [
      { role: 'user', content: '你好！' }
    ]
  });

  console.log(message.content[0].text);
}

main();
```bash

### cURL

```bash
curl -X POST http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "你好！"}]
  }'
```

开启流式传输（SSE）：

```bash
curl -N -X POST http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude",
    "max_tokens": 1024,
    "stream": true,
    "messages": [{"role": "user", "content": "你好！"}]
  }'
```

## 模型映射

`model` 参数映射到 clinvk 后端：

| Anthropic 模型 | clinvk 后端 |
|----------------|-------------|
| `claude` | Claude Code |
| `codex` | Codex CLI |
| `gemini` | Gemini CLI |

## 与 Anthropic API 的差异

!!! note "功能支持"
    目前已实现：

    - Messages API
    - 系统提示
    - 多轮对话
    - 流式传输（`stream: true`，SSE）

    尚未支持：

    - 工具使用
    - 视觉/图像
    - 文档理解

## 下一步

- [OpenAI 兼容](openai-compatible.md) - OpenAI 客户端支持
- [REST API](rest-api.md) - 完整 clinvk API 访问
