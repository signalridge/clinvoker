# mcp

启动 MCP (Model Context Protocol) 服务器。

## 概要

```bash
clinvk mcp [flags]
```

## 参数

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `--transport` | string | `stdio` | 传输类型（`stdio` 或 `http`） |
| `--host` | string | `127.0.0.1` | HTTP 传输的绑定地址 |
| `--port` | int | `8081` | HTTP 传输的端口 |
| `--path` | string | `/mcp` | HTTP 端点路径 |
| `--expose-health` | bool | `false` | 在 MCP 模式下暴露 `/health` |

未指定参数时，会从 config / env 解析默认值。

## Streaming gate

Streaming 需要显式开启：

- **stdio**：仅当 `output_format=stream-json`
- **HTTP**：仅当 `output_format=stream-json` 且 `Accept: text/event-stream`

如果没有 SSE Accept，会返回 JSON-RPC 错误。

## 示例

```bash
clinvk mcp --transport stdio
clinvk mcp --transport http --host 0.0.0.0 --port 3000 --path /mcp
clinvk mcp --transport http --expose-health
```

## 备注

HTTP 模式复用 `clinvk serve` 的同一套 router/middleware（auth、rate limits、CORS、timeouts、request size limits）。
