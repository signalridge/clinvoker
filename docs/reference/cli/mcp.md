# mcp

Start the MCP (Model Context Protocol) server.

## Synopsis

```bash
clinvk mcp [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--transport` | string | `stdio` | Transport type (`stdio` or `http`) |
| `--host` | string | `127.0.0.1` | Host to bind for HTTP transport |
| `--port` | int | `8081` | Port for HTTP transport |
| `--path` | string | `/mcp` | HTTP endpoint path |
| `--expose-health` | bool | `false` | Expose `/health` in MCP mode |

Defaults are resolved from config/env when flags are not provided.

## Streaming Gate

Streaming is explicit:

- **stdio**: only when `output_format=stream-json`
- **HTTP**: only when `output_format=stream-json` and `Accept: text/event-stream`

If streaming is requested without SSE Accept, the server returns a JSON-RPC error.

## Examples

```bash
clinvk mcp --transport stdio
clinvk mcp --transport http --host 0.0.0.0 --port 3000 --path /mcp
clinvk mcp --transport http --expose-health
```

## Notes

HTTP mode reuses the same router/middleware stack as `clinvk serve` (auth, rate limits, CORS, timeouts, request size limits).
