# Anthropic Compatible API

Use clinvk with existing Anthropic client libraries and tools.

## Overview

clinvk provides Anthropic-compatible endpoints that allow you to use the Anthropic Python SDK and other compatible clients.

## Base URL

```yaml
http://localhost:8080/anthropic/v1
```

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
| `model` | string | Yes | Backend name (claude, codex, gemini) |
| `max_tokens` | integer | Yes | Maximum response tokens |
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
    {
      "type": "text",
      "text": "Hello! How can I help you today?"
    }
  ],
  "model": "claude",
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 15
  }
}
```

## Client Examples

### Python (anthropic package)

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"  # clinvk doesn't require API key
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    system="You are a helpful coding assistant.",
    messages=[
        {"role": "user", "content": "Write a Python function to sort a list"}
    ]
)

print(message.content[0].text)
```

### With Conversation History

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

messages = [
    {"role": "user", "content": "What is Python?"},
]

# First message
response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=messages
)

# Add assistant response to history
messages.append({
    "role": "assistant",
    "content": response.content[0].text
})

# Continue conversation
messages.append({
    "role": "user",
    "content": "Show me a hello world example"
})

response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=messages
)

print(response.content[0].text)
```

### TypeScript/JavaScript

```typescript
import Anthropic from '@anthropic-ai/sdk';

const client = new Anthropic({
  baseURL: 'http://localhost:8080/anthropic/v1',
  apiKey: 'not-needed'
});

async function main() {
  const message = await client.messages.create({
    model: 'claude',
    max_tokens: 1024,
    messages: [
      { role: 'user', content: 'Hello!' }
    ]
  });

  console.log(message.content[0].text);
}

main();
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

To stream responses:

```bash
curl -N -X POST http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude",
    "max_tokens": 1024,
    "stream": true,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## Model Mapping

The `model` parameter maps to clinvk backends:

| Anthropic Model | clinvk Backend |
|-----------------|----------------|
| `claude` | Claude Code |
| `codex` | Codex CLI |
| `gemini` | Gemini CLI |

Specific model names also work:

```python
message = client.messages.create(
    model="claude-opus-4-5-20251101",  # Uses Claude backend with this model
    max_tokens=1024,
    messages=[...]
)
```

## Differences from Anthropic API

!!! note "Feature Support"
    Currently implemented:

    - Messages API
    - System prompts
    - Multi-turn conversations
    - Streaming (`stream: true`, SSE)

    Not yet supported:

    - Tool use
    - Vision/images
    - Document understanding

!!! note "Session Behavior"
    Each request is independent. For session continuity, use the clinvk REST API's session features.

## Message Format

### User Messages

```json
{
  "role": "user",
  "content": "Your message here"
}
```

### Assistant Messages (in history)

```json
{
  "role": "assistant",
  "content": "Previous assistant response"
}
```

### System Prompt

System prompts are passed as a top-level field:

```json
{
  "model": "claude",
  "max_tokens": 1024,
  "system": "You are a helpful assistant.",
  "messages": [...]
}
```

## Error Handling

Errors are returned as RFC 7807 Problem Details (via Huma), not Anthropic's error schema. For example, schema validation errors may return HTTP 422:

```json
{
  "title": "Unprocessable Entity",
  "status": 422,
  "detail": "max_tokens is required"
}
```

## Configuration

Uses the same server configuration:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
```

## Next Steps

- [OpenAI Compatible](openai-compatible.md) - OpenAI client support
- [REST API](rest-api.md) - Full clinvk API access
