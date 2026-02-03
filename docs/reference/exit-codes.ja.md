# 終了コード

clinvk の終了コードと意味に関する完全なリファレンスです。

## 概要

clinvk はコマンド実行結果を終了コードで示します。スクリプト化や自動化を行う場合、これらのコードを理解しておくことが重要です。

## 終了コード一覧

| コード | 名称 | 説明 | 発生タイミング |
|------|------|-------------|----------------|
| 0 | 成功 | コマンドが正常に完了した | 通常の完了 |
| 1 | 一般エラー | CLI/バリデーションエラー、またはサブコマンド失敗 | 入力不正、実行失敗 |
| 2 | バックエンド利用不可 | 指定したバックエンドがインストールされていない | バックエンドのバイナリが見つからない |
| 3 | 設定不正 | 設定ファイルのエラー、または不正な設定値 | 設定ファイル不備 |
| 4 | セッションエラー | セッション操作に失敗した | 再開失敗、セッションが見つからない |
| 5 | API エラー | HTTP API リクエストに失敗した | サーバーエラー、ネットワーク問題 |
| 6 | タイムアウト | コマンド実行がタイムアウトした | タイムアウト上限超過 |
| 7 | キャンセル | ユーザーが操作を中断した | Ctrl+C を押した |
| 8+ | バックエンド終了コード | バックエンド CLI から伝播 | バックエンド固有のエラー |

## 詳細

### 0 - 成功

エラーなく正常に完了しました。

```bash
clinvk "hello world"
echo $?  # Output: 0
```

### 1 - 一般エラー

実行中に一般的なエラーが発生しました。主な原因は次のとおりです。

- 不正なコマンドライン引数
- バックエンド実行の失敗
- ファイルが見つからない
- 権限不足

```bash
clinvk --invalid-flag "prompt"
echo $?  # Output: 1
```

### 2 - バックエンド利用不可

指定したバックエンドが未インストール、または PATH 上に存在しません。

```bash
clinvk -b nonexistent "prompt"
echo $?  # Output: 2
```

### 3 - 設定不正

設定ファイルにエラーがある、または不正な設定値が含まれています。

```bash
clinvk --config /invalid/config.yaml "prompt"
echo $?  # Output: 3
```

### 4 - セッションエラー

セッション関連の操作に失敗しました。

```bash
clinvk resume nonexistent-session
echo $?  # Output: 4
```

### 5 - API エラー

HTTP API リクエストが失敗しました（`clinvk serve` 利用時）。

```bash
# Example: API endpoint returns error
curl -X POST http://localhost:8080/api/v1/prompt \
  -d '{"backend": "invalid"}' 2>/dev/null || echo "API error"
```

### 6 - タイムアウト

設定されたタイムアウトを超えて実行が継続しました。

```bash
# Set timeout via config: unified_flags.command_timeout_secs
clinvk "very long task"
# If command_timeout_secs is set and exceeded, exits with code 6
echo $?  # Output: 6
```

### 7 - キャンセル

ユーザーが操作を中断しました（例: Ctrl+C）。

```bash
clinvk "long running task"
# Press Ctrl+C
echo $?  # Output: 7
```

### バックエンド終了コード（8+）

`clinvk [prompt]` または `clinvk resume` の実行時、clinvk はバックエンド CLI を起動し、バックエンドの終了コードが 0 以外だった場合はその値をそのまま伝播します。これらはバックエンド固有です。

## コマンド別の終了コード

### prompt / resume

| コード | 説明 |
|------|-------------|
| 0 | 成功 |
| 1 | 一般エラー |
| 2+ | バックエンド終了コード（伝播） |

### parallel

| コード | 説明 |
|------|-------------|
| 0 | すべてのタスクが成功 |
| 1 | 1 つ以上のタスクが失敗 |
| 2 | タスクファイルが不正 |

### compare

| コード | 説明 |
|------|-------------|
| 0 | すべてのバックエンドが成功 |
| 1 | 1 つ以上のバックエンドが失敗 |
| 2 | 利用可能なバックエンドがない |

### chain

| コード | 説明 |
|------|-------------|
| 0 | すべてのステップが成功 |
| 1 | いずれかのステップが失敗 |
| 2 | パイプラインファイルが不正 |

### sessions

| コード | 説明 |
|------|-------------|
| 0 | 操作が成功 |
| 1 | 操作が失敗（例: セッションが見つからない） |
| 4 | セッションエラー |

### config

| コード | 説明 |
|------|-------------|
| 0 | 操作が成功 |
| 1 | キーまたは値が不正 |
| 3 | 設定エラー |

### serve

| コード | 説明 |
|------|-------------|
| 0 | 正常終了（SIGINT/SIGTERM） |
| 1 | サーバー起動エラー |
| 5 | 動作中の API エラー |

## スクリプト例

### 成功/失敗を判定する

```bash
if clinvk "implement feature"; then
  echo "Success!"
else
  echo "Failed!"
fi
```

### 特定のコードを扱う

```bash
clinvk -b codex "prompt"
code=$?

case $code in
  0)
    echo "Success"
    ;;
  1)
    echo "General error"
    ;;
  2)
    echo "Backend not available - please install codex"
    ;;
  4)
    echo "Session error"
    ;;
  *)
    echo "Backend error: $code"
    ;;
esac
```

### 失敗時にリトライする

```bash
max_attempts=3
attempt=1

while [ $attempt -le $max_attempts ]; do
  if clinvk "prompt"; then
    echo "Success on attempt $attempt"
    break
  fi

  if [ $attempt -eq $max_attempts ]; then
    echo "Failed after $max_attempts attempts"
    exit 1
  fi

  echo "Attempt $attempt failed, retrying in 5 seconds..."
  sleep 5
  attempt=$((attempt + 1))
done
```

### エラー時に即終了する

```bash
#!/bin/bash
set -e  # Exit on any error

clinvk "step 1"
clinvk "step 2"
clinvk "step 3"

echo "All steps completed successfully"
```

### 特定の失敗は無視する

```bash
#!/bin/bash

# Continue even if this fails
clinvk "optional task" || true

# This must succeed
clinvk "critical task"
```

## CI/CD 統合

### GitHub Actions

```yaml
- name: Run AI task
  run: clinvk "generate tests"
  continue-on-error: true
  id: ai-task

- name: Handle failure
  if: failure() && steps.ai-task.outcome == 'failure'
  run: |
    echo "AI task failed with exit code $?"
    exit 1
```

### GitLab CI

```yaml
ai-task:
  script:
    - clinvk "generate tests" || EXIT_CODE=$?
    - |
      case $EXIT_CODE in
        0) echo "Success" ;;
        2) echo "Backend not installed" ; exit 1 ;;
        *) echo "Error: $EXIT_CODE" ; exit 1 ;;
      esac
```

### Make/Just

```makefile
.PHONY: test lint ai-review

test:
 go test ./...

ai-review:
 clinvk "review the code for issues" || (echo "Review failed" && exit 1)

lint-and-review: lint ai-review
 @echo "All checks passed"
```

## ベストプラクティス

1. スクリプトでは **常に終了コードを確認** し、失敗を適切に扱う
2. bash では `set -e` を使ってエラー時に即終了させる
3. デバッグ時は終了コードをログに残す
4. 必要に応じてコードごとに扱いを変える
5. 失敗しても止めたくない場合は `|| true` を使う

## トラブルシューティング

### 想定外の終了コード

| 症状 | 考えられる原因 | 対処 |
|---------|----------------|----------|
| 常に 1 を返す | バックエンド未設定 | 設定と API キーを確認する |
| 2 を返す | バックエンド未インストール | バックエンド CLI をインストールする |
| 4 を返す | セッションが期限切れ/不正 | `clinvk sessions list` でセッションを確認する |
| 6 を返す | タイムアウトが短すぎる | `command_timeout_secs` を増やす |

### 終了コードを調べる

```bash
# Run with verbose output
clinvk -v "prompt"
echo "Exit code: $?"

# Check backend directly
claude "test"
echo "Backend exit code: $?"
```

## 関連項目

- [コマンドリファレンス](cli/index.md) - コマンドのドキュメント
- [トラブルシューティング](../concepts/troubleshooting.md) - よくある問題と解決策
