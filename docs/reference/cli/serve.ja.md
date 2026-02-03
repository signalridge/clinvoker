# clinvk serve

HTTP API サーバーを起動します。

## 用法

```bash
clinvk serve [flags]
```

## 説明

clinvk の機能を REST API 経由で公開する HTTP サーバーを起動します。サーバーは 3 つの API スタイルを提供します。

- 独自 REST API（`/api/v1/`）
- OpenAI 互換 API（`/openai/v1/`）
- Anthropic 互換 API（`/anthropic/v1/`）

## フラグ

| フラグ | 短縮 | 型 | デフォルト | 説明 |
|------|-------|------|---------|-------------|
| `--host` | | string | `127.0.0.1` | バインドするホスト（設定値へフォールバック） |
| `--port` | `-p` | int | `8080` | 待ち受けポート（設定値へフォールバック） |

## 例

### デフォルトで起動

```bash
clinvk serve
# Server running at http://127.0.0.1:8080
```

### ポートを変更

```bash
clinvk serve --port 3000
```

### 全インターフェースにバインド

```bash
clinvk serve --host 0.0.0.0 --port 8080
```

!!! warning "Security"
    `0.0.0.0` へバインドすると、サーバーがネットワークに公開されます。本番環境では API キーを有効化し、CORS と workdir を制限してください。

## 認証

API キー認証は **オプション** です。キーを設定した場合に有効になります。

- `CLINVK_API_KEYS` environment variable (comma-separated)
- `server.api_keys_gopass_path` in config (gopass)

クライアントは次のいずれかを送信する必要があります。

- `X-Api-Key: <key>`
- `Authorization: Bearer <key>`

キーが設定されていない場合、デフォルトでリクエストは許可されます。

## エンドポイント

### 独自 REST API

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| POST | `/api/v1/prompt` | プロンプトを実行 |
| POST | `/api/v1/parallel` | 並列実行 |
| POST | `/api/v1/chain` | チェーン実行 |
| POST | `/api/v1/compare` | バックエンド比較 |
| GET | `/api/v1/backends` | バックエンド一覧 |
| GET | `/api/v1/sessions` | セッション一覧 |
| GET | `/api/v1/sessions/{id}` | セッション取得 |
| DELETE | `/api/v1/sessions/{id}` | セッション削除 |

### OpenAI 互換

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| GET | `/openai/v1/models` | モデル一覧 |
| POST | `/openai/v1/chat/completions` | チャット補完 |

### Anthropic 互換

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| POST | `/anthropic/v1/messages` | メッセージ作成 |

### Meta

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| GET | `/health` | ヘルスチェック |
| GET | `/openapi.json` | OpenAPI 仕様 |
| GET | `/docs` | API ドキュメント（Huma UI） |
| GET | `/metrics` | Prometheus メトリクス（有効時） |

## 動作確認

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

## 設定

サーバー設定は `~/.clinvk/config.yaml` にあります。

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  api_keys_gopass_path: "myproject/server/api-keys"
  rate_limit_enabled: false
  metrics_enabled: false
```

## 出力

起動時の例:

```text
clinvk API server starting on http://127.0.0.1:8080

Available endpoints:
  Custom API:     /api/v1/prompt, /api/v1/parallel, /api/v1/chain, /api/v1/compare
  OpenAI:         /openai/v1/models, /openai/v1/chat/completions
  Anthropic:      /anthropic/v1/messages
  Docs:           /openapi.json
  Health:         /health

Press Ctrl+C to stop
```

## 終了コード

| コード | 説明 |
|------|-------------|
| 0 | 正常終了 |
| 1 | サーバーエラー |

## 関連項目

- [REST API リファレンス](../api/rest.md)
- [OpenAI 互換](../api/openai-compat.md)
- [Anthropic 互換](../api/anthropic-compat.md)
