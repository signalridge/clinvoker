# MCP 服务器集成

## 概览

clinvk 内置 MCP (Model Context Protocol) 服务器，用 MCP 工具暴露现有能力，支持以下传输：

- **stdio**：本地/桌面使用
- **HTTP + SSE**：服务器部署使用

工具 schema 由现有 REST 模型生成，确保 MCP 与 REST 保持一致。

## 传输方式

### stdio (本地)

```bash
clinvk mcp --transport stdio
```

使用 stdin/stdout 的 line-delimited JSON-RPC。

### HTTP + SSE (服务器)

```bash
clinvk mcp --transport http --host 0.0.0.0 --port 3000 --path /mcp
```

向 `/mcp` POST JSON-RPC，Streaming 时使用 SSE。

## 工具列表

- `clinvk_prompt`
- `clinvk_parallel`
- `clinvk_chain`
- `clinvk_compare`
- `clinvk_backends`
- `clinvk_sessions`
- `clinvk_session_get`
- `clinvk_session_delete`
- `clinvk_health`

每个工具都包含 **input/output schema**。

## Streaming 开关

Streaming 必须显式开启：

- **stdio**：仅当 `output_format=stream-json`
- **HTTP**：`output_format=stream-json` 且 `Accept: text/event-stream`

如果请求 Streaming 但缺少 SSE 的 Accept，会返回 JSON-RPC 错误。

## 配置

### CLI 参数

```bash
clinvk mcp --transport stdio|http
clinvk mcp --transport http --host 127.0.0.1 --port 8081 --path /mcp
clinvk mcp --expose-health
```

### 配置文件 (`~/.clinvk/config.yaml`)

```yaml
mcp:
  transport: stdio
  host: "127.0.0.1"
  port: 8081
  http_path: "/mcp"
  expose_health: false
```

### 环境变量

```text
CLINVK_MCP_TRANSPORT
CLINVK_MCP_HOST
CLINVK_MCP_PORT
CLINVK_MCP_HTTP_PATH
CLINVK_MCP_EXPOSE_HEALTH
```

优先级：CLI 参数 > 环境变量 > 配置文件 > 默认值

## 示例

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

### HTTP Streaming prompt

```bash
curl -N -X POST http://127.0.0.1:8081/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"clinvk_prompt","arguments":{"backend":"claude","prompt":"hello","output_format":"stream-json"}}}'
```

## 认证、限制与 CORS

HTTP 模式复用 `clinvk serve` 的 middleware stack，包括：

- API key 认证（如果配置）
- Rate limiting
- Request size limits
- Timeouts
- CORS

如配置了 API key，请携带 `Authorization: Bearer <key>` 或 `X-API-Key`。

## 相关资源

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Claude Desktop MCP Guide](https://docs.anthropic.com/claude/docs/mcp)
- [REST API Reference](../../reference/api/rest.md)
