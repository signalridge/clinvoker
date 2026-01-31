---
title: API Design
description: REST API architecture, endpoint design, and compatibility layers.
---

# API Design

REST API architecture, endpoint design, and compatibility layers for clinvoker.

## API Architecture

clinvoker provides three API surfaces:

1. **Custom REST API** (`/api/v1/*`) - Native clinvoker operations
2. **OpenAI Compatible** (`/openai/v1/*`) - Drop-in OpenAI API replacement
3. **Anthropic Compatible** (`/anthropic/v1/*`) - Drop-in Anthropic API replacement

```mermaid
flowchart TB
    subgraph APIs
        CUSTOM[/api/v1/*]
        OPENAI[/openai/v1/*]
        ANTH[/anthropic/v1/*]
    end

    subgraph Handlers
        CH[CustomHandlers]
        OH[OpenAIHandlers]
        AH[AnthropicHandlers]
    end

    subgraph Services
        EXEC[Executor]
        STREAM[Streamer]
    end

    CUSTOM --> CH
    OPENAI --> OH
    ANTH --> AH

    CH --> EXEC
    OH --> EXEC
    AH --> EXEC

    CH --> STREAM
    OH --> STREAM
    AH --> STREAM
```

## Custom REST API

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/backends` | List available backends |
| POST | `/api/v1/prompt` | Execute a prompt |
| POST | `/api/v1/parallel` | Execute parallel tasks |
| POST | `/api/v1/chain` | Execute chained tasks |
| POST | `/api/v1/compare` | Compare backends |
| GET | `/api/v1/sessions` | List sessions |
| GET | `/api/v1/sessions/{id}` | Get session details |
| DELETE | `/api/v1/sessions/{id}` | Delete a session |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

### Request/Response Format

All endpoints use JSON:

```json
// POST /api/v1/prompt
{
  "backend": "claude",
  "prompt": "Explain REST APIs",
  "model": "claude-sonnet-4",
  "workdir": "/project",
  "max_tokens": 1000
}

// Response
{
  "session_id": "a1b2c3d4...",
  "backend": "claude",
  "exit_code": 0,
  "duration_ms": 2500,
  "output": "REST APIs are...",
  "token_usage": {
    "input_tokens": 50,
    "output_tokens": 150
  }
}
```

## OpenAI Compatibility

### Implemented Endpoints

| Endpoint | Status | Notes |
|----------|--------|-------|
| GET `/v1/models` | ✅ Complete | Lists backends as models |
| POST `/v1/chat/completions` | ✅ Complete | Supports streaming |
| POST `/v1/completions` | ⚠️ Partial | Mapped to chat completions |

### Model Mapping

clinvoker backends are exposed as OpenAI models:

```json
// GET /openai/v1/models
{
  "object": "list",
  "data": [
    {
      "id": "claude",
      "object": "model",
      "created": 1704067200,
      "owned_by": "clinvoker"
    },
    {
      "id": "codex",
      "object": "model",
      "created": 1704067200,
      "owned_by": "clinvoker"
    }
  ]
}
```

### Chat Completions

```bash
curl http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "claude",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'
```

### Streaming Support

Server-sent events for real-time responses:

```bash
curl http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

## Anthropic Compatibility

### Implemented Endpoints

| Endpoint | Status | Notes |
|----------|--------|-------|
| POST `/v1/messages` | ✅ Complete | Supports streaming |

### Messages API

```bash
curl http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'
```

## Streaming Architecture

### Event Types

```go
type EventType string

const (
    EventMessage   EventType = "message"
    EventError     EventType = "error"
    EventComplete  EventType = "complete"
)

type UnifiedEvent struct {
    Type      EventType       `json:"type"`
    Backend   string          `json:"backend"`
    Content   json.RawMessage `json:"content"`
    Timestamp time.Time       `json:"timestamp"`
}
```

### SSE Format

```
data: {"type": "message", "backend": "claude", "content": {"text": "Hello"}}

data: {"type": "message", "backend": "claude", "content": {"text": " world"}}

data: {"type": "complete", "backend": "claude"}
```

## Authentication

### API Key Methods

1. **Header**: `Authorization: Bearer <key>`
2. **Query**: `?api_key=<key>`
3. **Environment**: `CLINVK_API_KEY`

### Key Sources

```go
// Priority order:
1. Request header
2. Query parameter
3. Environment variable
4. gopass (if configured)
```

## Rate Limiting

### Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704067200
```

### Response Codes

| Code | Meaning |
|------|---------|
| 429 | Rate limit exceeded |
| 401 | Invalid API key |
| 403 | Forbidden |

## Error Handling

### Error Format

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Backend 'unknown' not found",
    "param": "backend",
    "type": "invalid_request_error"
  }
}
```

### HTTP Status Codes

| Code | Usage |
|------|-------|
| 200 | Success |
| 400 | Bad request |
| 401 | Unauthorized |
| 429 | Rate limited |
| 500 | Internal error |
| 503 | Backend unavailable |

## Best Practices

### 1. Request Timeouts

```yaml
server:
  request_timeout_secs: 300  # 5 minutes
  read_timeout_secs: 30
  write_timeout_secs: 300
```

### 2. Request Size Limits

```yaml
server:
  max_request_body_bytes: 10485760  # 10MB
```

### 3. CORS Configuration

```yaml
server:
  cors_allowed_origins:
    - "http://localhost:3000"
    - "https://app.example.com"
```

## Related Documentation

- [REST API Reference](../reference/api/rest-api.md) - Complete API documentation
- [OpenAI Compatible](../reference/api/openai-compatible.md) - OpenAI API details
- [Anthropic Compatible](../reference/api/anthropic-compatible.md) - Anthropic API details
