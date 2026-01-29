# 客户端库

本指南介绍如何在各种编程语言和 SDK 中使用 clinvk。

## 概述

clinvk 提供与现有客户端库兼容的 SDK 端点：

| 语言 | 库 | 端点 |
|------|---|------|
| Python | `openai` | `/openai/v1/*` |
| Python | `anthropic` | `/anthropic/v1/*` |
| Python | `httpx`/`requests` | `/api/v1/*` |
| TypeScript/JS | `openai` | `/openai/v1/*` |
| TypeScript/JS | `@anthropic-ai/sdk` | `/anthropic/v1/*` |
| Go | `sashabaranov/go-openai` | `/openai/v1/*` |
| Go | `net/http` | `/api/v1/*` |
| Rust | `async-openai` | `/openai/v1/*` |
| curl | - | 所有端点 |

## Python

### OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

# 聊天完成
response = client.chat.completions.create(
    model="claude",  # 后端名称
    messages=[
        {"role": "system", "content": "你是一个有帮助的助手。"},
        {"role": "user", "content": "你好！"}
    ]
)
print(response.choices[0].message.content)

# 流式传输
stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "写一首诗"}],
    stream=True
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

### Anthropic SDK

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:8080/anthropic",
    api_key="not-needed"
)

response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "解释递归。"}
    ]
)
print(response.content[0].text)
```

### httpx（直接 API）

```python
import httpx

# 简单提示
response = httpx.post(
    "http://localhost:8080/api/v1/prompt",
    json={
        "backend": "claude",
        "prompt": "你好，世界！"
    },
    timeout=60
)
print(response.json()["result"])

# 并行执行
response = httpx.post(
    "http://localhost:8080/api/v1/parallel",
    json={
        "tasks": [
            {"backend": "claude", "prompt": "审查架构"},
            {"backend": "codex", "prompt": "审查性能"}
        ]
    },
    timeout=120
)
for result in response.json()["results"]:
    print(f"{result['backend']}: {result['result']}")

# 链式执行
response = httpx.post(
    "http://localhost:8080/api/v1/chain",
    json={
        "steps": [
            {"name": "analyze", "backend": "claude", "prompt": "分析这个"},
            {"name": "improve", "backend": "codex", "prompt": "改进: {{previous}}"}
        ]
    },
    timeout=120
)
print(response.json()["results"][-1]["result"])
```

### 异步 Python

```python
import asyncio
from openai import AsyncOpenAI

async def main():
    client = AsyncOpenAI(
        base_url="http://localhost:8080/openai/v1",
        api_key="not-needed"
    )

    response = await client.chat.completions.create(
        model="claude",
        messages=[{"role": "user", "content": "你好！"}]
    )
    print(response.choices[0].message.content)

asyncio.run(main())
```

## TypeScript / JavaScript

### OpenAI SDK

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/openai/v1',
  apiKey: 'not-needed',
});

// 聊天完成
const response = await client.chat.completions.create({
  model: 'claude',
  messages: [
    { role: 'system', content: '你是一个有帮助的助手。' },
    { role: 'user', content: '你好！' },
  ],
});
console.log(response.choices[0].message.content);

// 流式传输
const stream = await client.chat.completions.create({
  model: 'claude',
  messages: [{ role: 'user', content: '写一首诗' }],
  stream: true,
});

for await (const chunk of stream) {
  process.stdout.write(chunk.choices[0]?.delta?.content || '');
}
```

### Anthropic SDK

```typescript
import Anthropic from '@anthropic-ai/sdk';

const client = new Anthropic({
  baseURL: 'http://localhost:8080/anthropic',
  apiKey: 'not-needed',
});

const response = await client.messages.create({
  model: 'claude',
  max_tokens: 1024,
  messages: [{ role: 'user', content: '解释递归。' }],
});
console.log(response.content[0].text);
```

### fetch（直接 API）

```typescript
// 简单提示
const response = await fetch('http://localhost:8080/api/v1/prompt', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    backend: 'claude',
    prompt: '你好，世界！',
  }),
});
const data = await response.json();
console.log(data.result);

// 并行执行
const parallelResponse = await fetch('http://localhost:8080/api/v1/parallel', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    tasks: [
      { backend: 'claude', prompt: '审查架构' },
      { backend: 'codex', prompt: '审查性能' },
    ],
  }),
});
const parallelData = await parallelResponse.json();
parallelData.results.forEach((r: any) => {
  console.log(`${r.backend}: ${r.result}`);
});
```

## Go

### go-openai

```go
package main

import (
    "context"
    "fmt"
    openai "github.com/sashabaranov/go-openai"
)

func main() {
    config := openai.DefaultConfig("not-needed")
    config.BaseURL = "http://localhost:8080/openai/v1"

    client := openai.NewClientWithConfig(config)

    resp, err := client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: "claude",
            Messages: []openai.ChatCompletionMessage{
                {Role: "user", Content: "你好！"},
            },
        },
    )
    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

### net/http（直接 API）

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type PromptRequest struct {
    Backend string `json:"backend"`
    Prompt  string `json:"prompt"`
}

type PromptResponse struct {
    Result string `json:"result"`
}

func main() {
    reqBody, _ := json.Marshal(PromptRequest{
        Backend: "claude",
        Prompt:  "你好，世界！",
    })

    resp, err := http.Post(
        "http://localhost:8080/api/v1/prompt",
        "application/json",
        bytes.NewBuffer(reqBody),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    var result PromptResponse
    json.Unmarshal(body, &result)

    fmt.Println(result.Result)
}
```

## Rust

### async-openai

```rust
use async_openai::{
    config::OpenAIConfig,
    types::{CreateChatCompletionRequestArgs, ChatCompletionRequestUserMessageArgs},
    Client,
};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = OpenAIConfig::new()
        .with_api_base("http://localhost:8080/openai/v1")
        .with_api_key("not-needed");

    let client = Client::with_config(config);

    let request = CreateChatCompletionRequestArgs::default()
        .model("claude")
        .messages([
            ChatCompletionRequestUserMessageArgs::default()
                .content("你好！")
                .build()?
                .into(),
        ])
        .build()?;

    let response = client.chat().create(request).await?;

    println!("{}", response.choices[0].message.content.as_ref().unwrap());

    Ok(())
}
```

## curl

```bash
# 简单提示
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "你好，世界！"}'

# OpenAI 兼容
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "messages": [{"role": "user", "content": "你好！"}]
  }'

# 并行执行
curl -X POST http://localhost:8080/api/v1/parallel \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {"backend": "claude", "prompt": "审查架构"},
      {"backend": "codex", "prompt": "审查性能"}
    ]
  }'

# 链式执行
curl -X POST http://localhost:8080/api/v1/chain \
  -H "Content-Type: application/json" \
  -d '{
    "steps": [
      {"name": "analyze", "backend": "claude", "prompt": "分析这个"},
      {"name": "improve", "backend": "codex", "prompt": "改进: {{previous}}"}
    ]
  }'
```

## 错误处理

所有客户端应处理这些常见错误：

| HTTP 状态码 | 含义 | 操作 |
|------------|------|------|
| 400 | 请求错误 | 检查请求格式 |
| 404 | 未找到 | 检查端点/后端名称 |
| 500 | 服务器错误 | 检查 clinvk 日志 |
| 503 | 服务不可用 | 后端 CLI 不可用 |
| 504 | 网关超时 | 增加超时时间 |

### Python 示例

```python
import httpx

try:
    response = httpx.post(
        "http://localhost:8080/api/v1/prompt",
        json={"backend": "claude", "prompt": "你好"},
        timeout=60
    )
    response.raise_for_status()
    print(response.json()["result"])
except httpx.ConnectError:
    print("无法连接到 clinvk 服务器")
except httpx.TimeoutException:
    print("请求超时")
except httpx.HTTPStatusError as e:
    print(f"HTTP 错误: {e.response.status_code}")
    print(e.response.json())
```

## 下一步

- [REST API 参考](../reference/rest-api.md) - 完整 API 文档
- [OpenAI 兼容](../reference/openai-compatible.md) - OpenAI 端点详情
- [Anthropic 兼容](../reference/anthropic-compatible.md) - Anthropic 端点详情
