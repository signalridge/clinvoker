# clinvk parallel

複数タスクを並列に実行します。

## 用法

```bash
clinvk parallel [flags]
```

## 説明

複数の AI タスクを同時に実行します。タスクは JSON ファイルで定義するか、stdin からパイプで渡します。

**注意:** CLI の parallel 実行は常にステートレスです（セッションは永続化されません）。また、確実にパースできるよう内部で JSON を強制するため、タスク単位の `output_format` は現状無視されます。

## フラグ

| フラグ | 短縮 | 型 | デフォルト | 説明 |
|------|-------|------|---------|-------------|
| `--file` | `-f` | string | | タスクファイル（JSON） |
| `--max-parallel` | | int | `3` | 最大同時実行数（設定を上書き） |
| `--fail-fast` | | bool | `false` | 最初の失敗で停止 |
| `--json` | | bool | `false` | JSON 出力 |
| `--quiet` | `-q` | bool | `false` | タスク出力を抑制 |

## タスクファイル形式

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "task prompt",
      "model": "optional-model",
      "workdir": "/optional/path",
      "approval_mode": "auto",
      "sandbox_mode": "workspace",
      "max_turns": 10
    }
  ],
  "max_parallel": 3,
  "fail_fast": true
}
```

### タスクのフィールド

| フィールド | 型 | 必須 | 説明 |
|-------|------|----------|-------------|
| `backend` | string | はい | 使用するバックエンド |
| `prompt` | string | はい | プロンプト |
| `model` | string | いいえ | モデル上書き |
| `workdir` | string | いいえ | 作業ディレクトリ |
| `approval_mode` | string | いいえ | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | いいえ | `default`, `read-only`, `workspace`, `full` |
| `output_format` | string | いいえ | 受け付けるが CLI parallel では **無視**（予約） |
| `max_tokens` | int | いいえ | 応答トークン上限（まだバックエンドフラグへは未対応） |
| `max_turns` | int | いいえ | 最大エージェントターン数 |
| `system_prompt` | string | いいえ | システムプロンプト |
| `extra` | array | いいえ | バックエンド固有の追加フラグ |
| `verbose` | bool | いいえ | 詳細出力を有効化 |
| `dry_run` | bool | いいえ | 実行をシミュレート |
| `id` | string | いいえ | タスク識別子 |
| `name` | string | いいえ | 表示名 |
| `tags` | array | いいえ | JSON 出力 / `output_dir` の成果物にコピーされるタグ |
| `meta` | object | いいえ | JSON 出力 / `output_dir` の成果物にコピーされる任意メタデータ |

### トップレベルのフィールド

| フィールド | 型 | 説明 |
|-------|------|-------------|
| `tasks` | array | タスク一覧 |
| `max_parallel` | int | 最大同時実行数 |
| `fail_fast` | bool | 最初の失敗で停止 |
| `output_dir` | string | `summary.json` とタスクごとの JSON 出力を保存する任意ディレクトリ |

## 例

### ファイルから実行

```bash
clinvk parallel --file tasks.json
```

### stdin から実行

```bash
cat tasks.json | clinvk parallel
```

### 同時実行数を制限

```bash
clinvk parallel --file tasks.json --max-parallel 2
```

### Fail-Fast モード

```bash
clinvk parallel --file tasks.json --fail-fast
```

### JSON 出力

```bash
clinvk parallel --file tasks.json --json
```

### 出力を保存する

```bash
cat tasks.json | jq '. + {"output_dir": "parallel_runs/run-001"}' | clinvk parallel
```

次のファイルが書き込まれます。

- `summary.json` (aggregate results)
- One JSON file per task (includes `task` + `result`)

### Quiet モード

```bash
clinvk parallel --file tasks.json --quiet
```

## 出力

### テキスト出力

```text
Running 3 tasks (max 3 parallel)...

[1] The auth module looks good...
[2] Added logging statements...
[3] Generated 5 test cases...

Results:
--------------------------------------------------------------------------------
#    BACKEND      STATUS   DURATION   TASK
--------------------------------------------------------------------------------
1    claude       OK       2.50s      review the auth module
2    codex        OK       3.20s      add logging to the API
3    gemini       OK       2.80s      generate tests for utils
--------------------------------------------------------------------------------
Total: 3 tasks, 3 completed, 0 failed (3.20s)
```

### JSON 出力

```json
{
  "total_tasks": 3,
  "completed": 3,
  "failed": 0,
  "total_duration_seconds": 3.2,
  "results": [
    {
      "index": 0,
      "task_id": "task-1",
      "task_name": "Auth Review",
      "backend": "claude",
      "output": "The auth module looks good...",
      "duration_seconds": 2.5,
      "exit_code": 0
    }
  ]
}
```

## 終了コード

| コード | 説明 |
|------|-------------|
| 0 | すべてのタスクが成功 |
| 1 | 1 つ以上のタスクが失敗 |

## 関連項目

- [chain](chain.md) - 逐次実行
- [compare](compare.md) - バックエンド比較
