# mcp

MCP (Model Context Protocol) サーバーを起動します。

## 概要

```bash
clinvk mcp [flags]
```

## フラグ

| フラグ | 型 | 既定値 | 説明 |
|-------|----|--------|------|
| `--transport` | string | `stdio` | トランスポート種別 (`stdio` または `http`) |
| `--host` | string | `127.0.0.1` | HTTP トランスポートの bind 先 |
| `--port` | int | `8081` | HTTP トランスポートのポート |
| `--path` | string | `/mcp` | HTTP エンドポイントパス |
| `--expose-health` | bool | `false` | MCP モードで `/health` を公開 |

フラグ未指定時は config/env から既定値を解決します。

## Streaming ゲート

Streaming は明示的に指定します。

- **stdio**: `output_format=stream-json` のときのみ
- **HTTP**: `output_format=stream-json` かつ `Accept: text/event-stream`

SSE の Accept がない場合は JSON-RPC エラーを返します。

## 例

```bash
clinvk mcp --transport stdio
clinvk mcp --transport http --host 0.0.0.0 --port 3000 --path /mcp
clinvk mcp --transport http --expose-health
```

## 備考

HTTP モードは `clinvk serve` と同じ middleware stack を利用します。
