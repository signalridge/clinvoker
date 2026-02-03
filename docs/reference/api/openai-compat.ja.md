# OpenAI 互換 API

既存の OpenAI クライアントライブラリやツールで clinvk を利用します。

## 概要

clinvk は OpenAI 互換のエンドポイントを提供しており、OpenAI SDK を CLI バックエンドで利用できます。これにより、OpenAI の API 形式を前提とした既存アプリケーションへ統合できます。

## ベース URL

```text
http://localhost:8080/openai/v1
```

## 認証

API キー認証はオプションです。キーが設定されている場合は、次のいずれかを付与してください。

- `Authorization: Bearer <key>`
- `X-Api-Key: <key>`

キーが設定されていない場合、認証なしでリクエストできます。

## エンドポイント

### GET /openai/v1/models

利用可能なモデル一覧を返します（バックエンドがモデルとしてマッピングされます）。

**レスポンス:**

```json
{
  "object": "list",
  "data": [
    {
      "id": "claude",
      "object": "model",
      "created": 1704067200,
      "owned_by": "clinvoker"
    },
    {
      "id": "codex",
      "object": "model",
      "created": 1704067200,
      "owned_by": "clinvoker"
    },
    {
      "id": "gemini",
      "object": "model",
      "created": 1704067200,
      "owned_by": "clinvoker"
    }
  ]
}
```

### POST /openai/v1/chat/completions

チャット補完を作成します。

**リクエストボディ:**

```json
{
  "model": "claude",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 4096,
  "temperature": 0.7,
  "stream": false
}
```

**フィールド:**

| フィールド | 型 | 必須 | 説明 |
|-------|------|----------|-------------|
| `model` | string | はい | バックエンド選択（下記マッピング参照） |
| `messages` | array | はい | チャットメッセージ |
| `max_tokens` | integer | いいえ | 応答トークン上限（現状 CLI バックエンドでは無視） |
| `temperature` | number | いいえ | サンプリング温度（無視） |
| `top_p` | number | いいえ | ニュークリアスサンプリング（無視） |
| `n` | integer | いいえ | 生成数（無視。常に 1） |
| `stream` | boolean | いいえ | `true` の場合にストリーミング（SSE）を有効化 |
| `stop` | string/array | いいえ | 停止シーケンス（無視） |
| `presence_penalty` | number | いいえ | presence penalty（無視） |
| `frequency_penalty` | number | いいえ | frequency penalty（無視） |
| `logit_bias` | object | いいえ | logit bias（無視） |
| `user` | string | いいえ | ユーザー識別子（無視） |
| `dry_run` | boolean | いいえ | **非標準:** コマンドを実行せずにシミュレート |

**レスポンス:**

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1704067200,
  "model": "claude",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 15,
    "total_tokens": 25
  }
}
```

**ストリーミングレスポンス:**

`stream: true` の場合、Server-Sent Events（SSE）を返します。

```text
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"claude","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

## モデルマッピング

`model` フィールドは使用するバックエンドを決定します。

| model の値 | 使用されるバックエンド |
|-------------|--------------|
| `claude` | Claude |
| `codex` | Codex |
| `gemini` | Gemini |
| `claude` を含む | Claude |
| `gpt` を含む | Codex |
| `gemini` を含む | Gemini |
| その他 | Claude（デフォルト） |

**推奨:** あいまいさを避けるため、バックエンド名（`codex`, `claude`, `gemini`）をそのまま指定してください。

## クライアント例

### Python（openai パッケージ）

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # Only required if API keys are enabled
)

response = client.chat.completions.create(
    model="claude",
    messages=[
        {"role": "system", "content": "You are a helpful coding assistant."},
        {"role": "user", "content": "Write a hello world in Python"}
    ]
)

print(response.choices[0].message.content)
```

### TypeScript/JavaScript

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

### cURL

```bash
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "claude", "messages": [{"role": "user", "content": "Hello!"}]}'
```

### ストリーミング例（Python）

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Tell me a story"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content is not None:
        print(chunk.choices[0].delta.content, end="")
```

## OpenAI API との差分

| 機能 | OpenAI API | clinvk 互換 |
|---------|------------|-------------------|
| Models | GPT-3.5, GPT-4 | Claude, Codex, Gemini |
| Completions | 対応 | 未実装 |
| Embeddings | 対応 | 未実装 |
| Images | 対応 | 未実装 |
| Audio | 対応 | 未実装 |
| エラーフォーマット | OpenAI スキーマ | RFC 7807 Problem Details |
| Sessions | ステートフル | ステートレス（セッションは REST API を使用） |

## 設定

OpenAI 互換 API はサーバー設定を共有します。

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
```

## エラーレスポンス

エラーは RFC 7807 Problem Details 形式です。

```json
{
  "type": "https://api.clinvk.dev/errors/backend-not-found",
  "title": "Backend Not Found",
  "status": 400,
  "detail": "The requested backend 'unknown' is not available"
}
```

## 次のステップ

- [Anthropic 互換](anthropic-compat.md) - Anthropic SDK 互換
- [REST API](rest.md) - 全機能向けのネイティブ REST API
- [`serve` コマンド](../cli/serve.md) - サーバー設定
