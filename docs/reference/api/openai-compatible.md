# OpenAI Compatible API

Use clinvk with existing OpenAI client libraries and tools.

## Overview

clinvk provides OpenAI-compatible endpoints that allow you to use OpenAI SDKs with CLI backends.

## Base URL

```text
http://localhost:8080/openai/v1
```

## Authentication

If API keys are configured, include:

- `Authorization: Bearer <key>`
- or `X-Api-Key: <key>`

If no keys are configured, requests are allowed.

## Endpoints

### GET /openai/v1/models

List available models (backends mapped as models).

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
| `model` | string | Yes | Backend selector (see mapping below) |
| `messages` | array | Yes | Chat messages |
| `max_tokens` | integer | No | Maximum response tokens (ignored by CLI backends today) |
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
```

## Model Mapping

The `model` field determines which backend is used:

- Exact backend names: `claude`, `codex`, `gemini`
- Any string containing `claude` → Claude backend
- Any string containing `gpt` → Codex backend
- Any string containing `gemini` → Gemini backend
- Anything else defaults to Claude

**Recommendation:** Use the backend name (`codex`, `claude`, `gemini`) to avoid ambiguity.

## Client Examples

### Python (openai package)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # Only required if API keys are enabled
)

response = client.chat.completions.create(
    model="claude",
    messages=[
        {"role": "system", "content": "You are a helpful coding assistant."},
        {"role": "user", "content": "Write a hello world in Python"}
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

const response = await client.chat.completions.create({
  model: 'claude',
  messages: [{ role: 'user', content: 'Hello!' }]
});

console.log(response.choices[0].message.content);
```

### cURL

```bash
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "claude", "messages": [{"role": "user", "content": "Hello!"}]}'
```

## Differences from OpenAI API

- Only chat completions and model listing are implemented.
- Errors follow RFC 7807 Problem Details (not OpenAI error schema).
- Requests are stateless; use the custom REST API for session persistence.

## Configuration

OpenAI-compatible API uses the same server configuration:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
```

## Next Steps

- [Anthropic Compatible](anthropic-compatible.md)
- [REST API](rest-api.md)
