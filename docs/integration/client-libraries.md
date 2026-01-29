# Client Libraries

This guide covers how to use clinvk with various programming languages and SDKs.

## Overview

clinvk provides SDK-compatible endpoints that work with existing client libraries:

| Language | Library | Endpoint |
|----------|---------|----------|
| Python | `openai` | `/openai/v1/*` |
| Python | `anthropic` | `/anthropic/v1/*` |
| Python | `httpx`/`requests` | `/api/v1/*` |
| TypeScript/JS | `openai` | `/openai/v1/*` |
| TypeScript/JS | `@anthropic-ai/sdk` | `/anthropic/v1/*` |
| Go | `sashabaranov/go-openai` | `/openai/v1/*` |
| Go | `net/http` | `/api/v1/*` |
| Rust | `async-openai` | `/openai/v1/*` |
| curl | - | All endpoints |

## Python

### OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

# Chat completion
response = client.chat.completions.create(
    model="claude",  # Backend name
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Hello!"}
    ]
)
print(response.choices[0].message.content)

# Streaming
stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Write a poem"}],
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
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "Explain recursion."}
    ]
)
print(response.content[0].text)
```

### httpx (Direct API)

```python
import httpx

# Simple prompt
response = httpx.post(
    "http://localhost:8080/api/v1/prompt",
    json={
        "backend": "claude",
        "prompt": "Hello, world!"
    },
    timeout=60
)
print(response.json()["output"])

# Parallel execution
response = httpx.post(
    "http://localhost:8080/api/v1/parallel",
    json={
        "tasks": [
            {"backend": "claude", "prompt": "Review architecture"},
            {"backend": "codex", "prompt": "Review performance"}
        ]
    },
    timeout=120
)
for result in response.json()["results"]:
    print(f"{result['backend']}: {result['output']}")

# Chain execution
response = httpx.post(
    "http://localhost:8080/api/v1/chain",
    json={
        "steps": [
            {"name": "analyze", "backend": "claude", "prompt": "Analyze this"},
            {"name": "improve", "backend": "codex", "prompt": "Improve: {{previous}}"}
        ]
    },
    timeout=120
)
print(response.json()["results"][-1]["output"])
```

### Async Python

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
        messages=[{"role": "user", "content": "Hello!"}]
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

// Chat completion
const response = await client.chat.completions.create({
  model: 'claude',
  messages: [
    { role: 'system', content: 'You are a helpful assistant.' },
    { role: 'user', content: 'Hello!' },
  ],
});
console.log(response.choices[0].message.content);

// Streaming
const stream = await client.chat.completions.create({
  model: 'claude',
  messages: [{ role: 'user', content: 'Write a poem' }],
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
  baseURL: 'http://localhost:8080/anthropic/v1',
  apiKey: 'not-needed',
});

const response = await client.messages.create({
  model: 'claude',
  max_tokens: 1024,
  messages: [{ role: 'user', content: 'Explain recursion.' }],
});
console.log(response.content[0].text);
```

### fetch (Direct API)

```typescript
// Simple prompt
const response = await fetch('http://localhost:8080/api/v1/prompt', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    backend: 'claude',
    prompt: 'Hello, world!',
  }),
});
const data = await response.json();
console.log(data.output);

// Parallel execution
const parallelResponse = await fetch('http://localhost:8080/api/v1/parallel', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    tasks: [
      { backend: 'claude', prompt: 'Review architecture' },
      { backend: 'codex', prompt: 'Review performance' },
    ],
  }),
});
const parallelData = await parallelResponse.json();
parallelData.results.forEach((r: any) => {
  console.log(`${r.backend}: ${r.output}`);
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
                {Role: "user", Content: "Hello!"},
            },
        },
    )
    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

### net/http (Direct API)

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
    Output string `json:"output"`
}

func main() {
    reqBody, _ := json.Marshal(PromptRequest{
        Backend: "claude",
        Prompt:  "Hello, world!",
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

    fmt.Println(result.Output)
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
                .content("Hello!")
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
# Simple prompt
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "Hello, world!"}'

# OpenAI-compatible
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Parallel execution
curl -X POST http://localhost:8080/api/v1/parallel \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {"backend": "claude", "prompt": "Review architecture"},
      {"backend": "codex", "prompt": "Review performance"}
    ]
  }'

# Chain execution
curl -X POST http://localhost:8080/api/v1/chain \
  -H "Content-Type: application/json" \
  -d '{
    "steps": [
      {"name": "analyze", "backend": "claude", "prompt": "Analyze this"},
      {"name": "improve", "backend": "codex", "prompt": "Improve: {{previous}}"}
    ]
  }'
```

## Error Handling

All clients should handle these common errors:

| HTTP Status | Meaning | Action |
|-------------|---------|--------|
| 400 | Bad Request | Check request format |
| 404 | Not Found | Check endpoint/backend name |
| 500 | Server Error | Check clinvk logs |
| 503 | Service Unavailable | Backend CLI not available |
| 504 | Gateway Timeout | Increase timeout |

### Python Example

```python
import httpx

try:
    response = httpx.post(
        "http://localhost:8080/api/v1/prompt",
        json={"backend": "claude", "prompt": "Hello"},
        timeout=60
    )
    response.raise_for_status()
    print(response.json()["output"])
except httpx.ConnectError:
    print("Cannot connect to clinvk server")
except httpx.TimeoutException:
    print("Request timed out")
except httpx.HTTPStatusError as e:
    print(f"HTTP error: {e.response.status_code}")
    print(e.response.json())
```

## Next Steps

- [REST API Reference](../reference/rest-api.md) - Complete API documentation
- [OpenAI Compatible](../reference/openai-compatible.md) - OpenAI endpoint details
- [Anthropic Compatible](../reference/anthropic-compatible.md) - Anthropic endpoint details
