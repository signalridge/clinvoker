---
title: Anthropic SDK 集成
description: 使用官方 Anthropic SDK 与 clinvoker 集成。
---

# Anthropic SDK 集成

使用官方 Anthropic SDK 与 clinvoker 作为后端进行集成。

## 快速开始

### Python

```python
from anthropic import Anthropic

# 配置客户端使用 clinvoker
client = Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="your-clinvoker-api-key"
)

# 像正常使用 Anthropic 客户端一样使用
response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "Hello, world!"}
    ]
)

print(response.content[0].text)
```

### TypeScript/JavaScript

```typescript
import Anthropic from '@anthropic-ai/sdk';

const client = new Anthropic({
    baseURL: 'http://localhost:8080/anthropic/v1',
    apiKey: 'your-clinvoker-api-key',
});

const response = await client.messages.create({
    model: 'claude',
    max_tokens: 1024,
    messages: [{ role: 'user', content: 'Hello!' }],
});

console.log(response.content[0].text);
```

## 模型支持

Anthropic SDK 可与 clinvoker 的 Claude 后端配合使用：

| 模型 | 后端 |
|------|------|
| `claude` | Claude Code |
| `claude-sonnet-4` | Claude Code |

## 流式响应

```python
# 流式响应
with client.messages.stream(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Write a poem"}]
) as stream:
    for text in stream.text_stream:
        print(text, end="", flush=True)
```

## 系统提示

```python
response = client.messages.create(
    model="claude",
    max_tokens=1024,
    system="You are a helpful coding assistant.",
    messages=[{"role": "user", "content": "Review this code"}]
)
```

## 最佳实践

1. **使用适当的超时时间**：Anthropic SDK 默认值可能太短
2. **处理流式响应**：对长响应使用流式处理
3. **检查 Token 数量**：通过 response.usage 监控使用情况

## 故障排除

### 连接问题

确保 clinvoker 服务器在预期端口上运行：

```bash
curl http://localhost:8080/health
```

### 版本兼容性

使用兼容的 Anthropic SDK 版本（1.0+）。

## 相关文档

- [Anthropic API 参考](../../reference/api/anthropic-compat.md)
- [OpenAI SDK 集成](openai-sdk.md)
