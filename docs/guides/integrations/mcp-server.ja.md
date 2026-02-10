# MCP サーバー連携

## 概要

clinvk には MCP (Model Context Protocol) サーバーが組み込まれており、既存の機能を MCP ツールとして公開します。以下のトランスポートをサポートします。

- **stdio**: ローカル/デスクトップ用途
- **HTTP + SSE**: サーバー配信用途

ツールの schema は既存の REST モデルから生成されるため、MCP と REST の整合性が保たれます。

## トランスポート

### stdio (ローカル)

```bash
clinvk mcp --transport stdio
```

stdin/stdout 上の line-delimited JSON-RPC を使用します。

### HTTP + SSE (サーバー)

```bash
clinvk mcp --transport http --host 0.0.0.0 --port 3000 --path /mcp
```

`/mcp` へ JSON-RPC を POST し、Streaming 時は SSE を利用します。

## ツール一覧

- `clinvk_prompt`
- `clinvk_parallel`
- `clinvk_chain`
- `clinvk_compare`
- `clinvk_backends`
- `clinvk_sessions`
- `clinvk_session_get`
- `clinvk_session_delete`
- `clinvk_health`

各ツールは **input/output schema** を含みます。

## Streaming ゲート

Streaming は明示的に指定します。

- **stdio**: `output_format=stream-json` のときのみ Streaming
- **HTTP**: `output_format=stream-json` かつ `Accept: text/event-stream` のときのみ Streaming

SSE の Accept がない場合は JSON-RPC エラーを返します。

## 設定

### CLI フラグ

```bash
clinvk mcp --transport stdio|http
clinvk mcp --transport http --host 127.0.0.1 --port 8081 --path /mcp
clinvk mcp --expose-health
```

### 設定ファイル (`~/.clinvk/config.yaml`)

```yaml
mcp:
  transport: stdio
  host: "127.0.0.1"
  port: 8081
  http_path: "/mcp"
  expose_health: false
```

### 環境変数

```text
CLINVK_MCP_TRANSPORT
CLINVK_MCP_HOST
CLINVK_MCP_PORT
CLINVK_MCP_HTTP_PATH
CLINVK_MCP_EXPOSE_HEALTH
```

優先順位: CLI フラグ > 環境変数 > 設定ファイル > 既定値

## 例

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

## 認証/制限/CORS

HTTP モードは `clinvk serve` と同じ middleware stack を使用します。

- API key 認証 (設定時)
- Rate limiting
- Request size limits
- Timeouts
- CORS

API key を設定している場合は `Authorization: Bearer <key>` または `X-API-Key` を送信してください。

## 参考リンク

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Claude Desktop MCP Guide](https://docs.anthropic.com/claude/docs/mcp)
- [REST API Reference](../../reference/api/rest.md)
