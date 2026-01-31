# Anthropic Compatible API

Use clinvk with existing Anthropic client libraries and tools.

## Overview

clinvk provides Anthropic-compatible endpoints that allow you to use Anthropic SDKs with CLI backends. This enables integration with existing applications that use Anthropic's API format.

## Base URL

```text
http://localhost:8080/anthropic/v1
```

## Authentication

API key authentication is optional. If keys are configured, include one of:

- `Authorization: Bearer <key>`
- `X-Api-Key: <key>`

If no keys are configured, requests are allowed without authentication.

## Endpoints

### POST /anthropic/v1/messages

Create a message (chat completion).

**Headers:**

| Header | Required | Description |
|--------|----------|-------------|
| `Content-Type` | Yes | `application/json` |
| `anthropic-version` | Yes | API version (e.g., `2023-06-01`) |

**Request Body:**

```json
{
  "model": "claude",
  "max_tokens": 1024,
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "system": "You are a helpful assistant."
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `model` | string | Yes | Backend selector (see mapping below) |
| `max_tokens` | integer | Yes | Maximum response tokens (ignored by CLI backends today) |
| `messages` | array | Yes | Chat messages |
| `system` | string | No | System prompt |
| `temperature` | number | No | Sampling temperature (ignored) |
| `top_p` | number | No | Nucleus sampling (ignored) |
| `top_k` | integer | No | Top-k sampling (ignored) |
| `stop_sequences` | array | No | Stop sequences (ignored) |
| `stream` | boolean | No | Enable streaming (SSE) when `true` |
| `metadata` | object | No | Request metadata (ignored) |

**Response:**

```json
{
  "id": "msg_abc123",
  "type": "message",
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Hello! How can I help you today?"}
  ],
  "model": "claude",
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 15
  }
}
```

**Streaming Response:**

When `stream: true`, returns Server-Sent Events (SSE):

```text
event: message_start
data: {"type":"message_start","message":{"id":"msg_abc123","type":"message","role":"assistant","content":[],"model":"claude","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":15}}

event: message_stop
data: {"type":"message_stop"}
```

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

**Recommendation:** Use `codex` or `gemini` explicitly when targeting those backends.

## Client Examples

### Python (anthropic package)

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"  # Only required if API keys are enabled
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    system="You are a helpful coding assistant.",
    messages=[{"role": "user", "content": "Write a Python function"}]
)

print(message.content[0].text)
```

### TypeScript/JavaScript

```typescript
import Anthropic from '@anthropic-ai/sdk';

const client = new Anthropic({
  baseURL: 'http://localhost:8080/anthropic/v1',
  apiKey: 'not-needed'
});

const message = await client.messages.create({
  model: 'claude',
  max_tokens: 1024,
  messages: [{ role: 'user', content: 'Hello!' }]
});

console.log(message.content[0].text);
```

### cURL

```bash
curl -X POST http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Streaming Example (Python)

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

stream = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Tell me a story"}],
    stream=True
)

for text in stream.text_stream:
    print(text, end="", flush=True)
```

## Differences from Anthropic API

| Feature | Anthropic API | clinvk Compatible |
|---------|---------------|-------------------|
| Models | Claude models | Claude, Codex, Gemini |
| Completions | Supported | Not implemented |
| Embeddings | Supported | Not implemented |
| Images | Supported | Not implemented |
| Tools | Supported | Not implemented |
| Error format | Anthropic schema | RFC 7807 Problem Details |
| Sessions | Stateful | Stateless (use REST API for sessions) |

## Configuration

Anthropic-compatible API uses the same server configuration:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
```

## Error Responses

Errors follow RFC 7807 Problem Details format:

```json
{
  "type": "https://api.clinvk.dev/errors/backend-not-found",
  "title": "Backend Not Found",
  "status": 400,
  "detail": "The requested backend 'unknown' is not available"
}
```

## Next Steps

- [OpenAI Compatible](openai-compat.md) - OpenAI SDK compatibility
- [REST API](rest.md) - Native REST API for full features
- [serve command](../cli/serve.md) - Server configuration
