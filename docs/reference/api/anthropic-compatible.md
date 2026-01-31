# Anthropic Compatible API

Use clinvk with existing Anthropic client libraries and tools.

## Overview

clinvk provides Anthropic-compatible endpoints that allow you to use Anthropic SDKs with CLI backends.

## Base URL

```text
http://localhost:8080/anthropic/v1
```

## Authentication

If API keys are configured, include:

- `Authorization: Bearer <key>`
- or `X-Api-Key: <key>`

If no keys are configured, requests are allowed.

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
| `stream` | boolean | No | Enable streaming (SSE) when `true` |

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

## Model Mapping

- Exact backend names: `claude`, `codex`, `gemini`
- Any string containing `claude` → Claude backend
- Anything else defaults to Claude

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

### cURL

```bash
curl -X POST http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model": "claude", "max_tokens": 1024, "messages": [{"role": "user", "content": "Hello!"}]}'
```

## Differences from Anthropic API

- Only the Messages API is implemented.
- Errors follow RFC 7807 Problem Details (not Anthropic error schema).
- Requests are stateless; use the custom REST API for session persistence.

## Configuration

Same server configuration applies:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
```

## Next Steps

- [OpenAI Compatible](openai-compatible.md)
- [REST API](rest-api.md)
