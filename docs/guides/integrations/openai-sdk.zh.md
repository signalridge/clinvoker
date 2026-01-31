---
title: OpenAI SDK 集成
description: 使用官方 OpenAI SDK 与 clinvoker 作为后端。
---

# OpenAI SDK 集成

使用官方 OpenAI SDK 与 clinvoker 作为后端，实现与现有基于 OpenAI 的应用程序的无缝集成。

## 快速开始

### Python

```python
from openai import OpenAI

# 配置客户端使用 clinvoker
client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-clinvoker-api-key"  # 或任何非空字符串
)

# 像正常使用 OpenAI 客户端一样使用
response = client.chat.completions.create(
    model="claude",  # 映射到 Claude 后端
    messages=[
        {"role": "user", "content": "Hello, world!"}
    ]
)

print(response.choices[0].message.content)
```text

### TypeScript/JavaScript

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
    baseURL: 'http://localhost:8080/openai/v1',
    apiKey: 'your-clinvoker-api-key',
});

const response = await client.chat.completions.create({
    model: 'claude',
    messages: [{ role: 'user', content: 'Hello!' }],
});

console.log(response.choices[0].message.content);
```text

### Go

```go
package main

import (
    "context"
    "fmt"
    "github.com/sashabaranov/go-openai"
)

func main() {
    client := openai.NewClientWithConfig(openai.ClientConfig{
        BaseURL: "http://localhost:8080/openai/v1",
        APIKey:  "your-clinvoker-api-key",
    })

    resp, err := client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: "claude",
            Messages: []openai.ChatCompletionMessage{
                {Role: openai.ChatMessageRoleUser, Content: "Hello!"},
            },
        },
    )

    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```text

## 模型映射

clinvoker 后端作为 OpenAI 模型暴露：

| clinvoker 后端 | OpenAI 模型 ID |
|----------------|----------------|
| Claude Code | `claude` |
| Codex CLI | `codex` |
| Gemini CLI | `gemini` |

您也可以使用后端特定的模型名称：

```python
# 使用特定的 Claude 模型
response = client.chat.completions.create(
    model="claude-sonnet-4",
    messages=[...]
)
```text

## 流式响应

```python
# 实时流式响应
stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Write a story"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```text

## 错误处理

```python
from openai import OpenAIError, RateLimitError

try:
    response = client.chat.completions.create(...)
except RateLimitError:
    print("Rate limit exceeded")
except OpenAIError as e:
    print(f"Error: {e}")
```text

## 高级配置

### 自定义请求头

```python
client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-key",
    default_headers={
        "X-Custom-Header": "value"
    }
)
```text

### 超时配置

```python
import httpx

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-key",
    timeout=httpx.Timeout(300.0)  # 5 分钟
)
```text

## 最佳实践

1. **连接池**：复用客户端实例
2. **错误重试**：实现指数退避
3. **请求超时**：为长时间运行的任务设置适当的超时
4. **API 密钥管理**：安全存储密钥

## 故障排除

### 连接被拒绝

确保 clinvoker 服务器正在运行：

```bash
clinvk serve --port 8080
```text

### 模型未找到

检查可用模型：

```bash
curl http://localhost:8080/openai/v1/models
```text

### 认证错误

验证您的 API 密钥是否在 clinvoker 中正确配置。

## 相关文档

- [OpenAI API 参考](../../reference/api/openai-compat.md)
- [LangChain 集成](langchain-langgraph.md)
