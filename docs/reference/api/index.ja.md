# API リファレンス

clinvk の HTTP API に関する完全なリファレンスです。

## 概要

clinvoker は、さまざまなツールや SDK と統合するために複数の API エンドポイントを提供します。HTTP サーバーは 3 つの API スタイルを公開します。

| API スタイル | エンドポイント接頭辞 | 得意な用途 |
|-----------|-----------------|----------|
| **ネイティブ REST** | `/api/v1/` | clinvoker の全機能 |
| **OpenAI 互換** | `/openai/v1/` | OpenAI SDK 利用者 |
| **Anthropic 互換** | `/anthropic/v1/` | Anthropic SDK 利用者 |

## クイックスタート

### サーバーを起動する

```bash
clinvk serve --port 8080
```

### API を試す

```bash
curl http://localhost:8080/health
```

### プロンプトを実行する

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello"}'
```

## API の選び方

| ユースケース | 推奨 API | ドキュメント |
|----------|-----------------|---------------|
| OpenAI SDK を使う | OpenAI 互換 | [openai-compat.md](openai-compat.md) |
| Anthropic SDK を使う | Anthropic 互換 | [anthropic-compat.md](anthropic-compat.md) |
| clinvoker の全機能を使う | ネイティブ REST | [rest.md](rest.md) |
| 独自統合を行う | ネイティブ REST | [rest.md](rest.md) |
| セッション管理を行う | ネイティブ REST | [rest.md](rest.md) |

## ベース URL

```text
http://localhost:8080
```

## 認証

API キー認証はオプションです。有効にした場合、すべてのリクエストに API キーが必要です。

### 設定

API キーは次のいずれかで設定します。

- `CLINVK_API_KEYS` 環境変数（カンマ区切り）
- 設定ファイルの `server.api_keys_gopass_path`（gopass のパス）

### リクエストヘッダー

API キーは次のいずれかのヘッダーで送信します。

```bash
# Option 1: X-Api-Key header
curl -H "X-Api-Key: your-api-key" http://localhost:8080/api/v1/prompt

# Option 2: Authorization header
curl -H "Authorization: Bearer your-api-key" http://localhost:8080/api/v1/prompt
```

キーが設定されていない場合、認証なしでリクエストできます。

## 応答形式

すべての API は一貫した構造の JSON を返します。

```json
{
  "success": true,
  "data": { ... }
}
```

## エラーハンドリング

エラーレスポンスにはエラーメッセージが含まれます。

```json
{
  "success": false,
  "error": "Backend not available"
}
```

### HTTP ステータスコード

| コード | 説明 |
|------|-------------|
| 200 | 成功 |
| 400 | 不正なリクエスト |
| 401 | 未認証（API キー不正/未指定） |
| 404 | 見つからない |
| 429 | レート制限 |
| 500 | サーバーエラー |

## 利用できるエンドポイント

### ネイティブ REST API

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

### メタエンドポイント

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| GET | `/health` | ヘルスチェック |
| GET | `/openapi.json` | OpenAPI 仕様 |
| GET | `/docs` | API ドキュメント（Huma UI） |
| GET | `/metrics` | Prometheus メトリクス |

## SDK 連携例

### Python（OpenAI SDK）

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```

### Python（Anthropic SDK）

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}]
)

print(message.content[0].text)
```

### JavaScript/TypeScript（OpenAI SDK）

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/openai/v1',
  apiKey: 'not-needed'
});

const response = await client.chat.completions.create({
  model: 'claude',
  messages: [{ role: 'user', content: 'Hello!' }]
});

console.log(response.choices[0].message.content);
```

## サーバー設定

サーバーは `~/.clinvk/config.yaml` で設定します。

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  rate_limit_enabled: false
  metrics_enabled: false
```

## セキュリティ上の注意

- 本番環境では API キーを使用してください
- 本番環境では CORS の許可オリジンを制限してください
- 許可する作業ディレクトリのプレフィックスを設定してください
- 本番環境では HTTPS を使用してください（リバースプロキシ経由）
- 公開環境ではレート制限を有効化してください

## 次のステップ

- [REST API ドキュメント](rest.md) - ネイティブ REST API のリファレンス
- [OpenAI 互換 API](openai-compat.md) - OpenAI SDK 互換
- [Anthropic 互換 API](anthropic-compat.md) - Anthropic SDK 互換
- [`serve` コマンド](../cli/serve.md) - サーバー起動コマンドの参照
