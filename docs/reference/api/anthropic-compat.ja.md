# Anthropic 互換 API

既存の Anthropic クライアントライブラリやツールで clinvk を利用します。

## 概要

clinvk は Anthropic 互換のエンドポイントを提供しており、Anthropic SDK を CLI バックエンドで利用できます。これにより、Anthropic の API 形式を前提とした既存アプリケーションへ統合できます。

## ベース URL

```text
http://localhost:8080/anthropic/v1
```

## 認証

API キー認証はオプションです。キーが設定されている場合は、次のいずれかを付与してください。

- `Authorization: Bearer <key>`
- `X-Api-Key: <key>`

キーが設定されていない場合、認証なしでリクエストできます。

## エンドポイント

### POST /anthropic/v1/messages

メッセージ（チャット補完）を作成します。

**ヘッダー:**

| ヘッダー | 必須 | 説明 |
|--------|----------|-------------|
| `Content-Type` | はい | `application/json` |
| `anthropic-version` | はい | API バージョン（例: `2023-06-01`） |

**リクエストボディ:**

```json
{
  "model": "claude",
  "max_tokens": 1024,
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "system": "You are a helpful assistant.",
  "dry_run": true
}
```

**フィールド:**

| フィールド | 型 | 必須 | 説明 |
|-------|------|----------|-------------|
| `model` | string | はい | バックエンド選択子（後述のマッピング参照） |
| `max_tokens` | integer | はい | 最大応答トークン数（現状、CLI バックエンドでは無視されます） |
| `messages` | array | はい | チャットメッセージ |
| `system` | string | いいえ | システムプロンプト |
| `temperature` | number | いいえ | サンプリング温度（無視されます） |
| `top_p` | number | いいえ | Nucleus サンプリング（無視されます） |
| `top_k` | integer | いいえ | Top-k サンプリング（無視されます） |
| `stop_sequences` | array | いいえ | 停止シーケンス（無視されます） |
| `stream` | boolean | いいえ | `true` の場合ストリーミング（SSE）を有効化します |
| `metadata` | object | いいえ | リクエストメタデータ（無視されます） |
| `dry_run` | boolean | いいえ | **非標準:** コマンドを実行せずにシミュレーションします |

**レスポンス:**

```json
{
  "id": "msg_abc123",
  "type": "message",
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Hello! How can I help you today?"}
  ],
  "model": "claude",
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 15
  }
}
```

**ストリーミングレスポンス:**

`stream: true` の場合、Server-Sent Events (SSE) を返します。

```text
event: message_start
data: {"type":"message_start","message":{"id":"msg_abc123","type":"message","role":"assistant","content":[],"model":"claude","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":15}}

event: message_stop
data: {"type":"message_stop"}
```

## モデルマッピング

`model` フィールドにより、どのバックエンドが使われるかが決まります。

| Model 値 | 使用されるバックエンド |
|-------------|--------------|
| `claude` | Claude |
| `codex` | Codex |
| `gemini` | Gemini |
| `claude` を含む | Claude |
| `gpt` を含む | Codex |
| `gemini` を含む | Gemini |
| それ以外 | Claude（デフォルト） |

**推奨:** Codex や Gemini を狙う場合は、`codex` / `gemini` を明示的に指定してください。

## クライアント例

### Python（anthropic パッケージ）

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"  # Only required if API keys are enabled
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    system="You are a helpful coding assistant.",
    messages=[{"role": "user", "content": "Write a Python function"}]
)

print(message.content[0].text)
```

### TypeScript/JavaScript

```typescript
import Anthropic from '@anthropic-ai/sdk';

const client = new Anthropic({
  baseURL: 'http://localhost:8080/anthropic/v1',
  apiKey: 'not-needed'
});

const message = await client.messages.create({
  model: 'claude',
  max_tokens: 1024,
  messages: [{ role: 'user', content: 'Hello!' }]
});

console.log(message.content[0].text);
```

### cURL

```bash
curl -X POST http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### ストリーミング例（Python）

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

stream = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Tell me a story"}],
    stream=True
)

for text in stream.text_stream:
    print(text, end="", flush=True)
```

## Anthropic API との差分

| 機能 | Anthropic API | clinvk 互換 |
|---------|---------------|-------------------|
| Models | Claude モデル | Claude, Codex, Gemini |
| Completions | 対応 | 未実装 |
| Embeddings | 対応 | 未実装 |
| Images | 対応 | 未実装 |
| Tools | 対応 | 未実装 |
| Error format | Anthropic スキーマ | RFC 7807 Problem Details |
| Sessions | ステートフル | ステートレス（セッションは REST API を利用） |

## 設定

Anthropic 互換 API は、同じサーバー設定を利用します。

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
```

## エラーレスポンス

エラーは RFC 7807 Problem Details 形式に従います。

```json
{
  "type": "https://api.clinvk.dev/errors/backend-not-found",
  "title": "Backend Not Found",
  "status": 400,
  "detail": "The requested backend 'unknown' is not available"
}
```

## 次のステップ

- [OpenAI 互換](openai-compat.md) - OpenAI SDK 互換
- [REST API](rest.md) - フル機能向けのネイティブ REST API
- [serve コマンド](../cli/serve.md) - サーバー設定
