# clinvk serve

Start the HTTP API server.

## Synopsis

```bash
clinvk serve [--host <host>] [--port <port>]
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host` | | config / `127.0.0.1` | Bind address |
| `--port` | `-p` | config / `8080` | Listen port |

## Notes

- Endpoints include `/api/v1/*`, `/openai/v1/*`, `/anthropic/v1/*`.
- OpenAPI schema is available at `/openapi.json`.
- Configurable security and limits live under `server:` in config.
