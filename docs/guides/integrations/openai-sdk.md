---
title: OpenAI SDK Integration
description: Use the OpenAI SDK with clinvoker as the backend.
---

# OpenAI SDK Integration

Use the official OpenAI SDK with clinvoker as the backend, enabling seamless integration with existing OpenAI-based applications.

## Quick Start

### Python

```python
from openai import OpenAI

# Configure client to use clinvoker
client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-clinvoker-api-key"  # or any non-empty string
)

# Use like normal OpenAI client
response = client.chat.completions.create(
    model="claude",  # Maps to Claude backend
    messages=[
        {"role": "user", "content": "Hello, world!"}
    ]
)

print(response.choices[0].message.content)
```

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
```

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
```

## Model Mapping

clinvoker backends are exposed as OpenAI models:

| clinvoker Backend | OpenAI Model ID |
|-------------------|-----------------|
| Claude Code | `claude` |
| Codex CLI | `codex` |
| Gemini CLI | `gemini` |

You can also use backend-specific model names:

```python
# Use specific Claude model
response = client.chat.completions.create(
    model="claude-sonnet-4",
    messages=[...]
)
```

## Streaming Responses

```python
# Stream responses in real-time
stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Write a story"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

## Error Handling

```python
from openai import OpenAIError, RateLimitError

try:
    response = client.chat.completions.create(...)
except RateLimitError:
    print("Rate limit exceeded")
except OpenAIError as e:
    print(f"Error: {e}")
```

## Advanced Configuration

### Custom Headers

```python
client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-key",
    default_headers={
        "X-Custom-Header": "value"
    }
)
```

### Timeout Configuration

```python
import httpx

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-key",
    timeout=httpx.Timeout(300.0)  # 5 minutes
)
```

## Best Practices

1. **Connection Pooling**: Reuse the client instance
2. **Error Retry**: Implement exponential backoff
3. **Request Timeouts**: Set appropriate timeouts for long-running tasks
4. **API Key Management**: Store keys securely

## Troubleshooting

### Connection Refused

Ensure the clinvoker server is running:

```bash
clinvk serve --port 8080
```

### Model Not Found

Check available models:

```bash
curl http://localhost:8080/openai/v1/models
```

### Authentication Errors

Verify your API key is configured correctly in clinvoker.

## Related Documentation

- [OpenAI API Reference](../../reference/api/openai-compat.md)
- [API Gateway Pattern](../http-server.md)
- [LangChain Integration](langchain-langgraph.md)
