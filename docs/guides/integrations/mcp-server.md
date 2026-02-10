# MCP Server Integration

## Overview

clinvk includes a built-in MCP (Model Context Protocol) server that exposes existing capabilities as MCP tools. It supports:

- **Stdio** for local/desktop usage (e.g., Claude Desktop)
- **HTTP + SSE** for server deployments with streaming notifications

All tool schemas are derived from the existing REST request/response models, so MCP and REST stay in sync.

## Transports

### Stdio (local)

Use stdio for local integrations and desktop clients:

```bash
clinvk mcp --transport stdio
```

Requests are line-delimited JSON-RPC over stdin/stdout.

### HTTP + SSE (server)

Use HTTP + SSE for network clients:

```bash
clinvk mcp --transport http --host 0.0.0.0 --port 3000 --path /mcp
```

Requests are JSON-RPC over HTTP POST to `/mcp`. Streaming notifications are emitted over SSE when enabled (see Streaming Gate below).

## Tool Surface

The MCP server exposes the following tools:

- `clinvk_prompt`
- `clinvk_parallel`
- `clinvk_chain`
- `clinvk_compare`
- `clinvk_backends`
- `clinvk_sessions`
- `clinvk_session_get`
- `clinvk_session_delete`
- `clinvk_health`

Each tool includes **input and output schemas** derived from `internal/server/handlers/models.go`.

## Streaming Gate

Streaming is explicit:

- **Stdio**: streaming only when `output_format=stream-json`
- **HTTP**: streaming only when `output_format=stream-json` **and** `Accept: text/event-stream`

If streaming is requested without `Accept: text/event-stream`, the server returns a JSON-RPC error.

## Configuration

### CLI Flags

```bash
clinvk mcp --transport stdio|http
clinvk mcp --transport http --host 127.0.0.1 --port 8081 --path /mcp
clinvk mcp --expose-health
```

### Config File (`~/.clinvk/config.yaml`)

```yaml
mcp:
  transport: stdio
  host: "127.0.0.1"
  port: 8081
  http_path: "/mcp"
  expose_health: false
```

### Environment Variables

```text
CLINVK_MCP_TRANSPORT
CLINVK_MCP_HOST
CLINVK_MCP_PORT
CLINVK_MCP_HTTP_PATH
CLINVK_MCP_EXPOSE_HEALTH
```

Priority: CLI flags > Environment variables > Config file > Defaults.

## Examples

### Claude Desktop (stdio)

```json
{
  "mcpServers": {
    "clinvk": {
      "command": "clinvk",
      "args": ["mcp", "--transport", "stdio"],
      "env": {
        "CLINVK_BACKEND": "claude"
      }
    }
  }
}
```

### HTTP tools/list

```bash
curl -sS -X POST http://127.0.0.1:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

### HTTP streaming prompt

```bash
curl -N -X POST http://127.0.0.1:8081/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"claude","prompt":"hello","output_format":"stream-json"}}}'
```

## Auth, Limits, and CORS

HTTP mode uses the same router/middleware stack as `clinvk serve`, including:

- API key auth (if configured)
- Rate limiting
- Request size limits
- Timeouts
- CORS

If API keys are configured, include `Authorization: Bearer <key>` or `X-API-Key`.

## Related Resources

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Claude Desktop MCP Guide](https://docs.anthropic.com/claude/docs/mcp)
- [REST API Reference](../../reference/api/rest.md)
