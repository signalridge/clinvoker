# REST API リファレンス

clinvk のカスタム REST API の完全なリファレンスです。

## ベース URL

```text
http://localhost:8080/api/v1
```

## 認証

API キー認証は **オプション** です。キーが設定されている場合、すべてのリクエストに次のいずれかを含める必要があります。

- `X-Api-Key: <key>`
- `Authorization: Bearer <key>`

キーは `CLINVK_API_KEYS`（カンマ区切り）または `server.api_keys_gopass_path`（gopass）で提供できます。

---

## プロンプト実行

### POST /api/v1/prompt

単一のプロンプトを実行します。

**リクエストボディ:**

```json
{
  "backend": "claude",
  "prompt": "explain this code",
  "model": "claude-opus-4-5-20251101",
  "workdir": "/path/to/project",
  "ephemeral": false,
  "approval_mode": "auto",
  "sandbox_mode": "workspace",
  "output_format": "json",
  "max_tokens": 4096,
  "max_turns": 10,
  "system_prompt": "You are a helpful assistant.",
  "verbose": false,
  "dry_run": false,
  "extra": ["--some-flag"],
  "metadata": {"project": "demo"}
}
```

**フィールド:**

| フィールド | 型 | 必須 | 説明 |
|-------|------|----------|-------------|
| `backend` | string | はい | 使用するバックエンド |
| `prompt` | string | はい | プロンプト |
| `model` | string | いいえ | モデルの上書き |
| `workdir` | string | いいえ | 作業ディレクトリ（絶対パス。許可/拒否プレフィックスに対して検証されます） |
| `ephemeral` | boolean | いいえ | ステートレスモード（セッションなし） |
| `approval_mode` | string | いいえ | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | いいえ | `default`, `read-only`, `workspace`, `full` |
| `output_format` | string | いいえ | `default`, `text`, `json`, `stream-json` |
| `max_tokens` | integer | いいえ | 最大応答トークン数（現状、バックエンドフラグには未マップ） |
| `max_turns` | integer | いいえ | エージェント的なターン数の上限 |
| `system_prompt` | string | いいえ | システムプロンプト |
| `verbose` | boolean | いいえ | 詳細出力を有効化 |
| `dry_run` | boolean | いいえ | 実行せずにシミュレーション |
| `extra` | array | いいえ | バックエンド固有の追加フラグ |
| `metadata` | object | いいえ | セッションに保存される任意メタデータ |

**レスポンス:**

```json
{
  "session_id": "abc123",
  "backend": "claude",
  "exit_code": 0,
  "duration_ms": 2500,
  "output": "The code explanation...",
  "token_usage": {
    "input_tokens": 123,
    "output_tokens": 456,
    "cached_tokens": 0,
    "reasoning_tokens": 0
  }
}
```

**ストリーミングレスポンス（`output_format: \"stream-json\"`）:**

統一イベントの NDJSON（`application/x-ndjson`）をストリームします。例（構造は簡略化）:

```json
{"type":"init","backend":"claude","session_id":"...","content":{"model":"..."}}
{"type":"message","backend":"claude","session_id":"...","content":{"text":"..."}}
{"type":"done","backend":"claude","session_id":"..."}
```

---

## 並列実行

### POST /api/v1/parallel

複数タスクを並列に実行します。

**リクエストボディ:**

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "task 1"
    },
    {
      "backend": "codex",
      "prompt": "task 2"
    }
  ],
  "max_parallel": 3,
  "fail_fast": false
}
```

各タスクは `/api/v1/prompt` と同じフィールド（`workdir`, `approval_mode`, `output_format` など）を受け付けます。

**レスポンス:**

```json
{
  "total_tasks": 2,
  "completed": 2,
  "failed": 0,
  "total_duration_ms": 2000,
  "results": [
    {
      "backend": "claude",
      "exit_code": 0,
      "duration_ms": 2000,
      "output": "result 1"
    }
  ]
}
```

> 並列タスクは常にステートレスです。`session_id` は省略される場合があります。

---

## チェーン実行

### POST /api/v1/chain

順次パイプラインを実行します。

**リクエストボディ:**

```json
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "analyze the code"
    },
    {
      "name": "improve",
      "backend": "codex",
      "prompt": "improve based on: {{previous}}"
    }
  ],
  "stop_on_failure": false,
  "pass_working_dir": false
}
```

**フィールド:**

| フィールド | 型 | 必須 | 説明 |
|-------|------|----------|-------------|
| `steps` | array | はい | チェーンステップの一覧 |
| `stop_on_failure` | boolean | いいえ | 最初の失敗で停止（API のデフォルトは `false`） |
| `pass_working_dir` | boolean | いいえ | 作業ディレクトリをステップ間で引き継ぐ |

> チェーン実行は常にステートレスです。`pass_session_id` と `persist_sessions` はサポートされません。

**レスポンス:**

```json
{
  "total_steps": 2,
  "completed_steps": 2,
  "failed_step": 0,
  "total_duration_ms": 3500,
  "results": [
    {
      "step": 1,
      "name": "analyze",
      "backend": "claude",
      "exit_code": 0,
      "duration_ms": 2000,
      "output": "analysis result"
    }
  ]
}
```

---

## バックエンド比較

### POST /api/v1/compare

複数バックエンドの応答を比較します。

**リクエストボディ:**

```json
{
  "prompt": "explain this algorithm",
  "backends": ["claude", "codex", "gemini"],
  "sequential": false
}
```

**レスポンス:**

```json
{
  "prompt": "explain this algorithm",
  "backends": ["claude", "codex", "gemini"],
  "total_duration_ms": 3200,
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "exit_code": 0,
      "duration_ms": 2500,
      "output": "explanation from claude"
    }
  ]
}
```

> 比較実行はステートレスです。`session_id` は省略される場合があります。

---

## バックエンド

### GET /api/v1/backends

利用可能なバックエンド一覧を返します。

**レスポンス:**

```json
{
  "backends": [
    {"name": "claude", "available": true},
    {"name": "codex", "available": true},
    {"name": "gemini", "available": false}
  ]
}
```

---

## セッション

### GET /api/v1/sessions

セッション一覧を返します。

**クエリパラメータ:**

| パラメータ | 型 | 説明 |
|-----------|------|-------------|
| `backend` | string | バックエンドで絞り込み |
| `status` | string | ステータスで絞り込み（`active`, `completed`, `error`, `paused`） |
| `limit` | integer | 最大件数 |
| `offset` | integer | ページングのオフセット |

**レスポンス:**

```json
{
  "sessions": [
    {
      "id": "abc123",
      "backend": "claude",
      "created_at": "2025-01-27T10:00:00Z",
      "last_used": "2025-01-27T11:30:00Z",
      "working_dir": "/projects/myapp",
      "model": "claude-opus-4-5-20251101",
      "initial_prompt": "Review auth changes",
      "status": "active",
      "turn_count": 3,
      "token_usage": {"input_tokens": 123, "output_tokens": 456},
      "tags": ["api"],
      "title": "Review auth changes"
    }
  ],
  "total": 42,
  "limit": 100,
  "offset": 0
}
```

### GET /api/v1/sessions/{id}

セッション詳細を取得します。

### DELETE /api/v1/sessions/{id}

セッションを削除します。

---

## ヘルスチェック

### GET /health

サーバーのヘルスステータスを返します。

**レスポンス（抜粋）:**

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "2m31s",
  "uptime_millis": 151000,
  "backends": [
    {"name": "claude", "available": true}
  ],
  "session_store": {
    "available": true,
    "session_count": 15
  }
}
```

**ステータス値:**

| ステータス | 説明 |
|--------|-------------|
| `ok` | すべてのシステムが正常動作 |
| `degraded` | 一部バックエンドが利用不可 |
| `unhealthy` | セッションストアが利用不可 |

---

## メトリクス

### GET /metrics

Prometheus 互換のメトリクスエンドポイントです（設定で `metrics_enabled: true` の場合）。

---

## エラーレスポンス

### Unauthorized（401）

API キーが設定されているにもかかわらず、欠落/不正な場合:

```json
{
  "error": "unauthorized",
  "message": "missing API key"
}
```

### Rate Limiting（429）

レート制限が有効で、上限を超過した場合。

### Request Size Limit（413）

リクエストボディが `max_request_body_bytes` を超過した場合。

注意:

- 既知の `Content-Length` が上限を超える場合、早期に 413 で拒否されます。
- チャンク送信やサイズ不明の場合、ハンドラがボディを読み取る際に `MaxBytesError` を受け取ります。ボディが一度も読み取られない場合、リクエストが拒否されないことがあります。

---

## OpenAPI 仕様

### GET /openapi.json

API の OpenAPI 3.0 仕様を返します。
