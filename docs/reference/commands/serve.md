# clinvk serve

Start the HTTP API server.

## Synopsis

```
clinvk serve [flags]
```

## Description

Start an HTTP server that exposes clinvk functionality via REST APIs. The server provides three API styles:

- Custom REST API (`/api/v1/`)
- OpenAI-compatible API (`/openai/v1/`)
- Anthropic-compatible API (`/anthropic/v1/`)

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--host` | | string | `127.0.0.1` | Host to bind to |
| `--port` | `-p` | int | `8080` | Port to listen on |

## Examples

### Start with Defaults

```bash
clinvk serve
# Server running at http://127.0.0.1:8080
```

### Custom Port

```bash
clinvk serve --port 3000
```

### Bind to All Interfaces

```bash
clinvk serve --host 0.0.0.0 --port 8080
```

!!! warning "Security"
    Binding to `0.0.0.0` exposes the server to the network. There is no built-in authentication.

## Endpoints

### Custom REST API

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

### Meta

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/openapi.json` | OpenAPI spec |

## Quick Test

```bash
# Health check
curl http://localhost:8080/health

# Execute prompt
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello"}'

# OpenAI-style
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "claude", "messages": [{"role": "user", "content": "hello"}]}'
```

## Configuration

Server settings in `~/.clinvk/config.yaml`:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
```

## Output

On start:

```
clinvk HTTP server starting
  Listening: http://127.0.0.1:8080
  Endpoints:
    /api/v1/       - Custom REST API
    /openai/v1/    - OpenAI compatible
    /anthropic/v1/ - Anthropic compatible
    /health        - Health check
    /openapi.json  - OpenAPI specification
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Clean shutdown |
| 1 | Server error |

## See Also

- [REST API Reference](../rest-api.md)
- [OpenAI Compatible](../openai-compatible.md)
- [Anthropic Compatible](../anthropic-compatible.md)
