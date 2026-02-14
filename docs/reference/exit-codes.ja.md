# 終了コード

現在の実装挙動に基づく `clinvk` の終了コードリファレンスです。

## 概要

`clinvk` はスクリプトや自動化のためにプロセス終了コードを返します。

## 保証される終了コード契約

| コード | 名称 | 説明 |
|------|------|------|
| 0 | 成功 | コマンドが正常に完了 |
| 1 | 一般エラー | バリデーション / 設定 / 実行時エラー、または集約サブコマンド失敗 |
| 6 | タイムアウト | `unified_flags.command_timeout_secs` 超過でタイムアウトがトップレベルに伝播 |
| 2+ | バックエンド終了コード | `prompt` / `resume` 経路でバックエンド CLI の終了コードを伝播 |

## コマンド別挙動

### prompt / resume

| コード | 意味 |
|------|------|
| 0 | 成功 |
| 1 | バックエンド終了コード伝播前の一般エラー |
| 6 | タイムアウト（`command_timeout_secs`） |
| 2+ | バックエンド CLI 終了コードを伝播 |

### compare / chain / parallel

| コード | 意味 |
|------|------|
| 0 | すべてのユニットが成功 |
| 1 | 1 つ以上のユニットが失敗（ユニット内タイムアウト含む） |

### sessions / config / serve / その他ユーティリティコマンド

| コード | 意味 |
|------|------|
| 0 | 成功 |
| 1 | 失敗 |

## タイムアウトの意味

`unified_flags.command_timeout_secs` を正の値に設定すると:

- バックエンドプロセスのタイムアウトが検出されます。
- トップレベルまで伝播したタイムアウトは終了コード `6` に正規化されます。
- 集約コマンド（`compare` / `chain` / `parallel`）では、各ユニットのタイムアウトはユニット失敗として扱われ、最終的に終了コード `1` になります。

例:

```bash
# config.yaml
unified_flags:
  command_timeout_secs: 5

clinvk "long running task"
echo $?  # トップレベル prompt/resume 経路のタイムアウトなら 6
```

## スクリプト例

### 終了コードで分岐

```bash
clinvk "task"
code=$?

case $code in
  0)
    echo "success"
    ;;
  6)
    echo "timeout"
    ;;
  *)
    echo "failed: $code"
    ;;
esac
```

### Shell で即時失敗

```bash
set -e

clinvk "step 1"
clinvk "step 2"
```

## 注意

- このファイルに記載のない過去のコード対応は前提にしないでください。
- CI で終了コードに依存する場合は、本ファイルの契約を基準にしてください。

## 関連

- [Commands Reference](cli/index.ja.md)
- [Configuration Reference](configuration.ja.md)
