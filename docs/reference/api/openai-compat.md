# OpenAI Compatible API

Use clinvk with existing OpenAI client libraries and tools.

## Overview

clinvk provides OpenAI-compatible endpoints that allow you to use OpenAI SDKs with CLI backends. This enables integration with existing applications that use OpenAI's API format.

## Base URL

```text
http://localhost:8080/openai/v1
```text

## Authentication

API key authentication is optional. If keys are configured, include one of:

- `Authorization: Bearer <key>`
- `X-Api-Key: <key>`

If no keys are configured, requests are allowed without authentication.

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
      "owned_by": "anthropic"
    },
    {
      "id": "codex",
      "object": "model",
      "created": 1704067200,
      "owned_by": "openai"
    },
    {
      "id": "gemini",
      "object": "model",
      "created": 1704067200,
      "owned_by": "google"
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
```text

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `model` | string | Yes | Backend selector (see mapping below) |
| `messages` | array | Yes | Chat messages |
| `max_tokens` | integer | No | Maximum response tokens (ignored by CLI backends today) |
| `temperature` | number | No | Sampling temperature (ignored) |
| `top_p` | number | No | Nucleus sampling (ignored) |
| `n` | integer | No | Number of completions (ignored, always 1) |
| `stream` | boolean | No | Enable streaming (SSE) when `true` |
| `stop` | string/array | No | Stop sequences (ignored) |
| `presence_penalty` | number | No | Presence penalty (ignored) |
| `frequency_penalty` | number | No | Frequency penalty (ignored) |
| `logit_bias` | object | No | Logit bias (ignored) |
| `user` | string | No | User identifier (ignored) |

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
```text

**Streaming Response:**

When `stream: true`, returns Server-Sent Events (SSE):

```text
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```bash

## Model Mapping

The `model` field determines which backend is used:

| Model Value | Backend Used |
|-------------|--------------|
| `claude` | Claude |
| `codex` | Codex |
| `gemini` | Gemini |
| Contains `claude` | Claude |
| Contains `gpt` | Codex |
| Contains `gemini` | Gemini |
| Anything else | Claude (default) |

**Recommendation:** Use the exact backend name (`codex`, `claude`, `gemini`) to avoid ambiguity.

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
```text

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
```text

### cURL

```bash
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "claude", "messages": [{"role": "user", "content": "Hello!"}]}'
```text

### Streaming Example (Python)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Tell me a story"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content is not None:
        print(chunk.choices[0].delta.content, end="")
```bash

## Differences from OpenAI API

| Feature | OpenAI API | clinvk Compatible |
|---------|------------|-------------------|
| Models | GPT-3.5, GPT-4 | Claude, Codex, Gemini |
| Completions | Supported | Not implemented |
| Embeddings | Supported | Not implemented |
| Images | Supported | Not implemented |
| Audio | Supported | Not implemented |
| Error format | OpenAI schema | RFC 7807 Problem Details |
| Sessions | Stateful | Stateless (use REST API for sessions) |

## Configuration

OpenAI-compatible API uses the same server configuration:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
```text

## Error Responses

Errors follow RFC 7807 Problem Details format:

```json
{
  "type": "https://api.clinvk.dev/errors/backend-not-found",
  "title": "Backend Not Found",
  "status": 400,
  "detail": "The requested backend 'unknown' is not available"
}
```text

## Next Steps

- [Anthropic Compatible](anthropic-compat.md) - Anthropic SDK compatibility
- [REST API](rest.md) - Native REST API for full features
- [serve command](../cli/serve.md) - Server configuration
