# HTTP API Server

clinvk includes a built-in HTTP API server that exposes all functionality via REST endpoints.

## Overview

The `clinvk serve` command starts an HTTP server that provides:

- **Custom REST API** - Full access to all clinvk features
- **OpenAI Compatible API** - Drop-in replacement for OpenAI clients
- **Anthropic Compatible API** - Drop-in replacement for Anthropic clients

## Quick Start

```bash
# Start with defaults (127.0.0.1:8080)
clinvk serve

# Custom port
clinvk serve --port 3000

# Bind to all interfaces
clinvk serve --host 0.0.0.0 --port 8080
```

## API Styles

- **[REST API](rest-api.md)** - Full-featured custom API for all clinvk operations
- **[OpenAI Compatible](openai-compatible.md)** - Use with existing OpenAI client libraries
- **[Anthropic Compatible](anthropic-compatible.md)** - Use with existing Anthropic client libraries

## Endpoints Overview

### Custom REST API (`/api/v1/`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/prompt` | Execute single prompt |
| POST | `/api/v1/parallel` | Execute multiple prompts |
| POST | `/api/v1/chain` | Execute prompt chain |
| POST | `/api/v1/compare` | Compare backend responses |
| GET | `/api/v1/backends` | List backends |
| GET | `/api/v1/sessions` | List sessions |
| GET | `/api/v1/sessions/{id}` | Get session |
| DELETE | `/api/v1/sessions/{id}` | Delete session |

### OpenAI Compatible (`/openai/v1/`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/openai/v1/models` | List models |
| POST | `/openai/v1/chat/completions` | Chat completion |

### Anthropic Compatible (`/anthropic/v1/`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/anthropic/v1/messages` | Create message |

### Meta Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/openapi.json` | OpenAPI spec |

## Configuration

Server settings in `~/.clinvk/config.yaml`:

```yaml
server:
  # Bind address
  host: "127.0.0.1"

  # Port number
  port: 8080

  # Request processing timeout (seconds)
  request_timeout_secs: 300

  # Read request timeout (seconds)
  read_timeout_secs: 30

  # Write response timeout (seconds)
  write_timeout_secs: 300

  # Idle connection timeout (seconds)
  idle_timeout_secs: 120
```

## Security Notes

!!! warning "Local Use Only"
    By default, the server binds to `127.0.0.1` (localhost only). If you bind to `0.0.0.0` to expose it publicly, be aware that **there is no authentication**.

!!! tip "Production Use"
    For production deployments, place the server behind a reverse proxy (nginx, Caddy) that handles:

    - TLS termination
    - Authentication
    - Rate limiting
    - Request logging

## Example Requests

### Execute a Prompt

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "explain this code"}'
```

### OpenAI-Style Chat

```bash
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Health Check

```bash
curl http://localhost:8080/health
```

## OpenAPI Specification

The server provides an OpenAPI specification at `/openapi.json`:

```bash
# Download spec
curl http://localhost:8080/openapi.json > openapi.json

# View in Swagger UI or import into API tools
```

## Next Steps

- [Getting Started](getting-started.md) - Step-by-step server setup
- [REST API Reference](rest-api.md) - Full API documentation
- [OpenAI Compatible](openai-compatible.md) - Use with OpenAI clients
