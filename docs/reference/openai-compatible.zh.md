# OpenAI 兼容 API

使用 clinvk 与现有的 OpenAI 客户端库和工具。

## 概述

clinvk 提供 OpenAI 兼容的端点，允许您使用任何 OpenAI 客户端库与您的 AI 后端交互。

## 基础 URL

```
http://localhost:8080/openai/v1
```

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
      "owned_by": "clinvk"
    }
  ]
}
```

### POST /openai/v1/chat/completions

创建聊天补全。

**请求体：**

```json
{
  "model": "claude",
  "messages": [
    {"role": "system", "content": "你是一个有帮助的助手。"},
    {"role": "user", "content": "你好！"}
  ],
  "max_tokens": 4096
}
```

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
        "content": "你好！今天我能帮你什么？"
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

## 客户端示例

### Python (openai 包)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # clinvk 不需要 API key
)

# 列出模型（后端）
models = client.models.list()
for model in models:
    print(model.id)

# 聊天补全
response = client.chat.completions.create(
    model="claude",
    messages=[
        {"role": "system", "content": "你是一个有帮助的编程助手。"},
        {"role": "user", "content": "用 Python 写一个 hello world"}
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

async function main() {
  const response = await client.chat.completions.create({
    model: 'claude',
    messages: [
      { role: 'user', content: '你好！' }
    ]
  });

  console.log(response.choices[0].message.content);
}

main();
```

### LangChain

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
    model="claude"
)

response = llm.invoke("解释 Python 装饰器")
print(response.content)
```

## 模型映射

OpenAI 的 `model` 参数映射到 clinvk 后端：

| OpenAI 模型 | clinvk 后端 |
|-------------|-------------|
| `claude` | Claude Code |
| `codex` | Codex CLI |
| `gemini` | Gemini CLI |

## 与 OpenAI API 的差异

!!! note "功能支持"
    目前已实现：

    - 聊天补全
    - 模型列表
    - 基本消息格式
    - 流式传输（`stream: true`，SSE）

    尚未支持：

    - 函数调用
    - 视觉/图像
    - 嵌入
    - 文件上传

## 下一步

- [Anthropic 兼容](anthropic-compatible.md) - Anthropic 客户端支持
- [REST API](rest-api.md) - 完整 clinvk API 访问
