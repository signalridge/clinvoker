# API Reference

Complete reference for clinvk HTTP APIs.

## Overview

clinvoker provides multiple API endpoints for integration with various tools and SDKs. The HTTP server exposes three API styles:

| API Style | Endpoint Prefix | Best For |
|-----------|-----------------|----------|
| **Native REST** | `/api/v1/` | Full clinvoker features |
| **OpenAI Compatible** | `/openai/v1/` | OpenAI SDK users |
| **Anthropic Compatible** | `/anthropic/v1/` | Anthropic SDK users |

## Quick Start

### Start the Server

```bash
clinvk serve --port 8080
```text

### Test the API

```bash
curl http://localhost:8080/health
```text

### Execute a Prompt

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello"}'
```text

## Choosing an API

| Use Case | Recommended API | Documentation |
|----------|-----------------|---------------|
| Using OpenAI SDK | OpenAI Compatible | [openai-compat.md](openai-compat.md) |
| Using Anthropic SDK | Anthropic Compatible | [anthropic-compat.md](anthropic-compat.md) |
| Full clinvoker features | Native REST | [rest.md](rest.md) |
| Custom integrations | Native REST | [rest.md](rest.md) |
| Session management | Native REST | [rest.md](rest.md) |

## Base URL

```text
http://localhost:8080
```text

## Authentication

API key authentication is optional. When configured, every request must include an API key.

### Configuration

Set API keys via:

- `CLINVK_API_KEYS` environment variable (comma-separated)
- `server.api_keys_gopass_path` in config (gopass path)

### Request Headers

Include the API key using one of these headers:

```bash
# Option 1: X-Api-Key header
curl -H "X-Api-Key: your-api-key" http://localhost:8080/api/v1/prompt

# Option 2: Authorization header
curl -H "Authorization: Bearer your-api-key" http://localhost:8080/api/v1/prompt
```text

If no keys are configured, requests are allowed without authentication.

## Response Format

All APIs return JSON responses with a consistent structure:

```json
{
  "success": true,
  "data": { ... }
}
```text

## Error Handling

Error responses include an error message:

```json
{
  "success": false,
  "error": "Backend not available"
}
```text

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad request |
| 401 | Unauthorized (invalid/missing API key) |
| 404 | Not found |
| 429 | Rate limited |
| 500 | Server error |

## Available Endpoints

### Native REST API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/prompt` | Execute prompt |
| POST | `/api/v1/parallel` | Parallel execution |
| POST | `/api/v1/chain` | Chain execution |
| POST | `/api/v1/compare` | Backend comparison |
| GET | `/api/v1/backends` | List backends |
| GET | `/api/v1/sessions` | List sessions |
| GET | `/api/v1/sessions/{id}` | Get session |
| DELETE | `/api/v1/sessions/{id}` | Delete session |

### OpenAI Compatible

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/openai/v1/models` | List models |
| POST | `/openai/v1/chat/completions` | Chat completion |

### Anthropic Compatible

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/anthropic/v1/messages` | Create message |

### Meta Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/openapi.json` | OpenAPI specification |
| GET | `/docs` | API documentation (Huma UI) |
| GET | `/metrics` | Prometheus metrics |

## SDK Integration Examples

### Python with OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```text

### Python with Anthropic SDK

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}]
)

print(message.content[0].text)
```text

### JavaScript/TypeScript with OpenAI SDK

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
```bash

## Server Configuration

Configure the server in `~/.clinvk/config.yaml`:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  rate_limit_enabled: false
  metrics_enabled: false
```text

## Security Considerations

- Use API keys for production deployments
- Restrict CORS origins in production
- Configure allowed working directory prefixes
- Use HTTPS in production (via reverse proxy)
- Enable rate limiting for public deployments

## Next Steps

- [REST API Documentation](rest.md) - Native REST API reference
- [OpenAI Compatible API](openai-compat.md) - OpenAI SDK compatibility
- [Anthropic Compatible API](anthropic-compat.md) - Anthropic SDK compatibility
- [serve command](../cli/serve.md) - Server command reference
