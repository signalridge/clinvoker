---
title: Anthropic SDK Integration
description: Use the Anthropic SDK with clinvoker as the backend.
---

# Anthropic SDK Integration

Use the official Anthropic SDK with clinvoker as the backend.

## Quick Start

### Python

```python
from anthropic import Anthropic

# Configure client to use clinvoker
client = Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="your-clinvoker-api-key"
)

# Use like normal Anthropic client
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

## Model Support

The Anthropic SDK works with clinvoker's Claude backend:

| Model | Backend |
|-------|---------|
| `claude` | Claude Code |
| `claude-sonnet-4` | Claude Code |

## Streaming

```python
# Stream responses
with client.messages.stream(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Write a poem"}]
) as stream:
    for text in stream.text_stream:
        print(text, end="", flush=True)
```

## System Prompts

```python
response = client.messages.create(
    model="claude",
    max_tokens=1024,
    system="You are a helpful coding assistant.",
    messages=[{"role": "user", "content": "Review this code"}]
)
```

## Best Practices

1. **Use Appropriate Timeouts**: Anthropic SDK defaults may be too short
2. **Handle Streaming**: Use streaming for long responses
3. **Check Token Counts**: Monitor usage via response.usage

## Troubleshooting

### Connection Issues

Ensure clinvoker server is running on the expected port:

```bash
curl http://localhost:8080/health
```

### Version Compatibility

Use a compatible Anthropic SDK version (1.0+).

## Related Documentation

- [Anthropic API Reference](../reference/api/anthropic-compatible.md)
- [OpenAI SDK Integration](openai-sdk.md)
