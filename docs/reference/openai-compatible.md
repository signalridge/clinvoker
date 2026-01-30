# OpenAI Compatible API

Use clinvk with existing OpenAI client libraries and tools.

## Overview

clinvk provides OpenAI-compatible endpoints that allow you to use any OpenAI client library to interact with your AI backends.

## Base URL

```yaml
http://localhost:8080/openai/v1
```

## Endpoints

### GET /openai/v1/models

List available models (backends mapped as models).

**Response:**

```json
{
  "object": "list",
  "data": [
    {
      "id": "claude",
      "object": "model",
      "created": 1704067200,
      "owned_by": "clinvk"
    },
    {
      "id": "codex",
      "object": "model",
      "created": 1704067200,
      "owned_by": "clinvk"
    },
    {
      "id": "gemini",
      "object": "model",
      "created": 1704067200,
      "owned_by": "clinvk"
    }
  ]
}
```text

### POST /openai/v1/chat/completions

Create a chat completion.

**Request Body:**

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

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `model` | string | Yes | Backend name (claude, codex, gemini) |
| `messages` | array | Yes | Chat messages |
| `max_tokens` | integer | No | Maximum response tokens |
| `temperature` | number | No | Sampling temperature (ignored) |
| `stream` | boolean | No | Enable streaming (SSE) when `true` |

**Response:**

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
```bash

## Client Examples

### Python (openai package)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # clinvk doesn't require API key
)

# List models (backends)
models = client.models.list()
for model in models:
    print(model.id)

# Chat completion
response = client.chat.completions.create(
    model="claude",
    messages=[
        {"role": "system", "content": "You are a helpful coding assistant."},
        {"role": "user", "content": "Write a hello world in Python"}
    ]
)

print(response.choices[0].message.content)

# Streaming (SSE)
stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Write a poem"}],
    stream=True
)
for chunk in stream:
    if chunk.choices and chunk.choices[0].delta and chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
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
      { role: 'user', content: 'Hello!' }
    ]
  });

  console.log(response.choices[0].message.content);
}

main();
```bash

### cURL

```bash
# List models
curl http://localhost:8080/openai/v1/models

# Chat completion
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### LangChain

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
    model="claude"
)

response = llm.invoke("Explain Python decorators")
print(response.content)
```bash

## Model Mapping

OpenAI's `model` parameter maps to clinvk backends:

| OpenAI Model | clinvk Backend |
|--------------|----------------|
| `claude` | Claude Code |
| `codex` | Codex CLI |
| `gemini` | Gemini CLI |

You can also use specific model names:

```python
# Uses claude backend with specific model
response = client.chat.completions.create(
    model="claude-opus-4-5-20251101",
    messages=[...]
)
```

## Differences from OpenAI API

!!! note "Feature Support"
    Not all OpenAI API features are supported. Currently implemented:

    - Chat completions
    - Model listing
    - Basic message format
    - Streaming (`stream: true`, SSE)

    Not yet supported:

    - Function calling
    - Vision/images
    - Embeddings
    - File uploads

!!! note "Session Behavior"
    By default, each request creates a new session. Use the REST API's session features for conversation continuity.

## Configuration

The OpenAI-compatible API uses the same server configuration:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
```text

## Error Handling

Errors are returned as RFC 7807 Problem Details (via Huma), not OpenAI's error schema. For example, schema validation errors may return HTTP 422:

```json
{
  "title": "Unprocessable Entity",
  "status": 422,
  "detail": "model is required"
}
```

## Next Steps

- [Anthropic Compatible](anthropic-compatible.md) - Anthropic client support
- [REST API](rest-api.md) - Full clinvk API access
