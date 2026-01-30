# HTTP Server

Expose clinvk as an API server with three API styles:

- **Custom REST API**: `/api/v1/*`
- **OpenAI‑compatible API**: `/openai/v1/*`
- **Anthropic‑compatible API**: `/anthropic/v1/*`

## Start the server

```bash
clinvk serve --host 127.0.0.1 --port 8080
```

Defaults come from `server.host` and `server.port` in config.

## Endpoints

### Custom REST API

- `POST /api/v1/prompt`
- `POST /api/v1/parallel`
- `POST /api/v1/chain`
- `POST /api/v1/compare`
- `GET /api/v1/backends`
- `GET /api/v1/sessions`
- `GET /api/v1/sessions/{id}`
- `DELETE /api/v1/sessions/{id}`
- `GET /health`

### OpenAI‑compatible

- `GET /openai/v1/models`
- `POST /openai/v1/chat/completions`

### Anthropic‑compatible

- `POST /anthropic/v1/messages`

### OpenAPI schema

- `GET /openapi.json`

## Example: custom prompt

```bash
curl -sS http://localhost:8080/api/v1/prompt \
  -H 'Content-Type: application/json' \
  -d '{"backend":"claude","prompt":"Hello"}'
```

## Streaming (custom API)

Set `output_format` to `stream-json` to get NDJSON unified events.

```bash
curl -N http://localhost:8080/api/v1/prompt \
  -H 'Content-Type: application/json' \
  -d '{"backend":"claude","prompt":"Stream please","output_format":"stream-json"}'
```

## Sessions and stateless behavior

- **Custom REST API** creates sessions unless `ephemeral: true` is set.
- **OpenAI/Anthropic endpoints** are **stateless** (no session persistence).

## Security & limits

Configure in `server:`

- **API keys**: set `CLINVK_API_KEYS` or `CLINVK_API_KEYS_GOPASS_PATH`
- **Rate limiting**: `rate_limit_enabled`, `rate_limit_rps`, `rate_limit_burst`
- **Request size**: `max_request_body_bytes`
- **CORS**: `cors_allowed_origins`, `cors_allow_credentials`, `cors_max_age`
- **Workdir restrictions**: `allowed_workdir_prefixes`, `blocked_workdir_prefixes`
- **Timeouts**: `request_timeout_secs`, `read_timeout_secs`, `write_timeout_secs`, `idle_timeout_secs`
- **Metrics**: `metrics_enabled` (exposes `/metrics`)

See [Configuration Reference](../reference/configuration.md) for details.
