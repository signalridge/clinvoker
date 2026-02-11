# MCP Server Integration Design

## Summary

- Provide MCP stdio and HTTP transports.
- HTTP runs as a dedicated server but uses the same router/middleware stack as clinvk serve.
- Only /mcp is exposed by default; /health is optional.
- Streaming is explicitly controlled by output_format=stream-json and SSE Accept for HTTP.
- Tool schemas are derived from REST structs; tools include input and output schema.

## Goals

- Reuse executor/service logic.
- Enforce auth, rate limit, CORS, and timeouts for MCP HTTP.
- Avoid schema drift by deriving from REST structs.
- Preserve JSON-RPC semantics for notifications.

## Non-Goals

- MCP client support.
- Changes to backend execution behavior.

## Architecture

- Use server.NewRouter to build a chi router with the shared middleware stack.
- runMCPHTTP mounts only /mcp (configurable) and optional /health; REST/OpenAI/Anthropic/docs/openapi/schemas are not mounted in MCP mode.
- MCP stdio remains standalone.

## Configuration

- New config keys: mcp.transport, mcp.host, mcp.port, mcp.http_path, mcp.expose_health.
- Env: CLINVK_MCP_TRANSPORT, CLINVK_MCP_HOST, CLINVK_MCP_PORT, CLINVK_MCP_HTTP_PATH, CLINVK_MCP_EXPOSE_HEALTH.
- Precedence: flags > env > config > defaults.
- Defaults: transport=stdio, host=127.0.0.1, port=8081, http_path=/mcp, expose_health=false.
- CLI flags default to empty/zero to allow config to apply.

## Routing and Middleware

- /mcp goes through RequestSize, RateLimit, APIKeyAuth, Timeout, CORS, Metrics, and logging middleware.
- SkipAuthPaths uses the shared fixed list (/health, /docs, /openapi.json, /schemas, /metrics). MCP mode only mounts /health when enabled.

## JSON-RPC Semantics

- If request id is missing or null, treat as notification and never return a response.
- tools/call notifications are ignored (no execution) to avoid side effects.
- Unknown notification methods are ignored.

## Initialize Negotiation

- protocolVersion is optional; if provided and unsupported, return CodeInvalidParams.
- clientInfo is logged; capabilities are accepted but ignored (no negotiation yet).
- tools capability is returned; listChanged is omitted (not supported).

## Tool Schema Derivation

- Reflection-based JSON Schema generator derives schema from REST structs:
  - Use json tag; if absent, use query/path tag.
  - omitempty => not required.
  - doc tag => description.
  - time.Time => string (date-time), time.Duration => string.
  - map => object with additionalProperties, slices => array.
- Tool includes inputSchema and outputSchema.

## Tool Output Mapping

- Each tool returns JSON-encoded response body in content[0].text.
- Response bodies match handlers/models.go types (PromptResponseBody, ParallelResponseBody, etc).
- Streaming prompt accumulates output and builds PromptResponseBody using StreamResult for exit_code, token_usage, duration_ms. Errors return JSON-RPC error responses (no body error).

## Streaming Gate

- Streaming only when output_format=stream-json.
- HTTP requires Accept to include text/event-stream; otherwise JSON-RPC error CodeInvalidParams.
- SSE sends notifications as event: notification and final result as event: message.
- Stdio sends notifications only when streaming is requested.

## Error Handling

- Map executor errors using MapExecutorError to JSON-RPC error codes.
- Unknown tool => CodeToolNotFound.
- Execution failures => CodeToolExecutionError or CodeInternalError.
- Streaming errors produce notifications and a final JSON-RPC error response.

## Testing

- Notification semantics (id null) for initialize/tools/list/tools/call/ping.
- SSE Accept variations; stream-json without SSE => error.
- Schema derivation matches REST struct fields and descriptions.
- MCP HTTP inherits auth, rate limit, CORS, and timeouts from shared middleware.
- Streaming error returns JSON-RPC error.
