# HTTP サーバー

clinvk には HTTP API サーバーが組み込まれており、すべての機能を REST エンドポイントとして公開できます。

## 概要

`clinvk serve` コマンドは次の機能を提供する HTTP サーバーを起動します。

- **独自 REST API** - clinvk の全機能にアクセス
- **OpenAI 互換 API** - OpenAI クライアントのドロップイン置換
- **Anthropic 互換 API** - Anthropic クライアントのドロップイン置換

## サーバーを起動する

### 基本

```bash
clinvk serve
```

デフォルトでは `http://127.0.0.1:8080` で起動します。

### ポートを変更

```bash
clinvk serve --port 3000
```

### 全インターフェースにバインド

```bash
clinvk serve --host 0.0.0.0 --port 8080
```

!!! warning "Security"
    `0.0.0.0` にバインドするとサーバーがネットワークに公開されます。本番環境では API キーを有効化し、CORS と workdir を制限してください。

## 認証

API キー認証はオプションです。`CLINVK_API_KEYS` または `server.api_keys_gopass_path` でキーを設定すると有効になります。クライアントは `Authorization: Bearer <key>` または `X-Api-Key: <key>` を送信する必要があります。

## サーバーをテストする

### ヘルスチェック

```bash
curl http://localhost:8080/health
```

レスポンスには `status`、`version`、`uptime`、バックエンドの利用可否、セッションストアの状態が含まれます。

### バックエンド一覧

```bash
curl http://localhost:8080/api/v1/backends
```

### プロンプトを実行する

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello world"}'
```

## エンドポイント概要

### 独自 REST API（`/api/v1/`）

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| POST | `/api/v1/prompt` | 単一プロンプトを実行 |
| POST | `/api/v1/parallel` | 複数プロンプトを実行 |
| POST | `/api/v1/chain` | プロンプトチェーンを実行 |
| POST | `/api/v1/compare` | バックエンド応答を比較 |
| GET | `/api/v1/backends` | バックエンド一覧 |
| GET | `/api/v1/sessions` | セッション一覧 |
| GET | `/api/v1/sessions/{id}` | セッション取得 |
| DELETE | `/api/v1/sessions/{id}` | セッション削除 |

### OpenAI 互換（`/openai/v1/`）

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| GET | `/openai/v1/models` | モデル一覧 |
| POST | `/openai/v1/chat/completions` | チャット補完 |

### Anthropic 互換（`/anthropic/v1/`）

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| POST | `/anthropic/v1/messages` | メッセージ作成 |

### メタエンドポイント

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| GET | `/health` | ヘルスチェック |
| GET | `/openapi.json` | OpenAPI 仕様 |
| GET | `/docs` | API ドキュメント UI |
| GET | `/metrics` | Prometheus メトリクス（有効時） |

## クライアントライブラリで利用する

### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # only required if API keys are enabled
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)
```

### Python (Anthropic SDK)

```python
import anthropic

client = anthropic.Client(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"  # only required if API keys are enabled
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}]
)
print(message.content[0].text)
```

### JavaScript/TypeScript

```typescript
const response = await fetch('http://localhost:8080/api/v1/prompt', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    backend: 'claude',
    prompt: 'hello world'
  })
});

const data = await response.json();
console.log(data.output);
```

### cURL

```bash
# Simple prompt
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "explain async/await"}'

# Parallel execution
curl -X POST http://localhost:8080/api/v1/parallel \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {"backend": "claude", "prompt": "task 1"},
      {"backend": "codex", "prompt": "task 2"}
    ]
  }'
```

## 設定

### 設定ファイルで指定

Edit `~/.clinvk/config.yaml`:

```yaml
server:
  # Bind address
  host: "127.0.0.1"

  # Port number
  port: 8080

  # Request processing timeout (seconds)
  request_timeout_secs: 300

  # Read request timeout (seconds)
  read_timeout_secs: 30

  # Write response timeout (seconds)
  write_timeout_secs: 300

  # Idle connection timeout (seconds)
  idle_timeout_secs: 120

  # Optional API keys via gopass (leave empty to disable)
  api_keys_gopass_path: ""
```

### CLI フラグで指定

```bash
clinvk serve --host 0.0.0.0 --port 3000
```

CLI フラグは設定ファイルの値を上書きします。

## サービスとして実行する

### systemd (Linux)

Create `/etc/systemd/system/clinvk.service`:

```ini
[Unit]
Description=clinvk API Server
After=network.target

[Service]
Type=simple
User=youruser
ExecStart=/usr/local/bin/clinvk serve --port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

有効化して起動します。

```bash
sudo systemctl enable clinvk
sudo systemctl start clinvk
```

### Docker

```bash
docker run -d \
  --name clinvk \
  -p 8080:8080 \
  ghcr.io/signalridge/clinvk serve --host 0.0.0.0
```

### launchd (macOS)

Create `~/Library/LaunchAgents/com.clinvk.server.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.clinvk.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/clinvk</string>
        <string>serve</string>
        <string>--port</string>
        <string>8080</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

サービスをロードします。

```bash
launchctl load ~/Library/LaunchAgents/com.clinvk.server.plist
```

## セキュリティ上の注意

!!! warning "Local Use Only"
    デフォルトではサーバーは `127.0.0.1`（localhost のみ）にバインドされます。`0.0.0.0` にバインドして公開する場合、**認証がない** ことに注意してください。

!!! tip "Production Use"
    本番環境では、サーバーをリバースプロキシ（nginx, Caddy など）の背後に置き、次を処理させることを推奨します。

    - TLS termination
    - Authentication
    - Rate limiting
    - Request logging

## OpenAPI 仕様

サーバーは `/openapi.json` で OpenAPI 仕様を提供します。

```bash
# Download spec
curl http://localhost:8080/openapi.json > openapi.json

# View in Swagger UI or import into API tools
```

## 次のステップ

- [REST API リファレンス](../reference/api/rest.md) - API ドキュメント一式
- [OpenAI 互換](../reference/api/openai-compat.md) - OpenAI クライアントで利用
- [Anthropic 互換](../reference/api/anthropic-compat.md) - Anthropic クライアントで利用
