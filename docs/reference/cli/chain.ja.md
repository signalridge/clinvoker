# clinvk chain

プロンプトの逐次パイプラインを実行します。

## 用法

```bash
clinvk chain [flags]
```

## 説明

複数のプロンプトを順番に実行し、各ステップの出力を `{{previous}}` を介して次のステップへ渡します。これにより、バックエンドごとの得意分野を活かした多段ワークフローを構築できます。

**注意:** CLI の chain 実行は常にステートレスです（セッションは永続化されません）。`{{session}}`、`pass_session_id`、`persist_sessions` はサポートされておらず、指定するとエラーになります。

## フラグ

| フラグ | 短縮 | 型 | デフォルト | 説明 |
|------|-------|------|---------|-------------|
| `--file` | `-f` | string | | パイプラインファイル（JSON） |
| `--json` | | bool | `false` | JSON 出力 |

## パイプラインファイル形式

```json
{
  "steps": [
    {
      "name": "step-name",
      "backend": "claude",
      "prompt": "First prompt",
      "model": "optional-model"
    },
    {
      "name": "second-step",
      "backend": "gemini",
      "prompt": "Process this: {{previous}}"
    }
  ]
}
```

### ステップのフィールド

| フィールド | 型 | 必須 | 説明 |
|-------|------|----------|-------------|
| `name` | string | いいえ | ステップ識別子 |
| `backend` | string | はい | 使用するバックエンド |
| `prompt` | string | はい | プロンプト |
| `model` | string | いいえ | モデル上書き |
| `workdir` | string | いいえ | 作業ディレクトリ |
| `approval_mode` | string | いいえ | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | いいえ | `default`, `read-only`, `workspace`, `full` |
| `max_turns` | int | いいえ | 最大エージェントターン数 |

### トップレベルのフィールド

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `steps` | array | | ステップ一覧（必須） |
| `stop_on_failure` | bool | `true` | **CLI は常に失敗時に停止**（フィールドは受け付けるが `false` は無視） |
| `pass_working_dir` | bool | `false` | ステップ間で作業ディレクトリを引き継ぐ |

### テンプレート変数

| 変数 | 説明 |
|----------|-------------|
| `{{previous}}` | 直前ステップの出力テキスト |

## 例

### 基本

```bash
clinvk chain --file pipeline.json
```

### JSON 出力

```bash
clinvk chain --file pipeline.json --json
```

## 出力

### テキスト出力

```text
Executing chain with 3 steps
================================================================================

[1/3] analyze (claude)
--------------------------------------------------------------------------------
Analysis result text...

[2/3] recommend (gemini)
--------------------------------------------------------------------------------
Recommendations text...

[3/3] implement (codex)
--------------------------------------------------------------------------------
Implementation text...

================================================================================
CHAIN EXECUTION SUMMARY
================================================================================
STEP   BACKEND      STATUS   DURATION   NAME
--------------------------------------------------------------------------------
1      claude       OK       2.10s      analyze
2      gemini       OK       1.80s      recommend
3      codex        OK       3.20s      implement
--------------------------------------------------------------------------------
Total: 3/3 steps completed (7.10s)
```

### JSON 出力

```json
{
  "total_steps": 3,
  "completed_steps": 3,
  "failed_step": 0,
  "total_duration_seconds": 7.1,
  "results": [
    {
      "step": 1,
      "name": "analyze",
      "backend": "claude",
      "output": "Analysis result...",
      "duration_seconds": 2.1,
      "exit_code": 0
    }
  ]
}
```

## 終了コード

| コード | 説明 |
|------|-------------|
| 0 | すべてのステップが成功 |
| 1 | いずれかのステップが失敗 |

## 関連項目

- [parallel](parallel.md) - 並列実行
- [compare](compare.md) - バックエンド比較
