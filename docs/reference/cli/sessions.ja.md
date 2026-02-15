# clinvk sessions

セッションを管理します。

## 用法

```bash
clinvk sessions [command] [flags]
```

## 説明

`sessions` コマンドは、clinvk のセッションを管理するサブコマンドを提供します。セッションは会話履歴と状態を保存し、後で会話を再開できるようにします。

## サブコマンド

| コマンド | 説明 |
|---------|-------------|
| `list` | セッション一覧 |
| `show` | セッション詳細 |
| `delete` | セッション削除 |
| `clean` | 古いセッションを削除 |

---

## clinvk sessions list

セッション一覧を表示します。

### フラグ

| フラグ | 短縮 | 型 | デフォルト | 説明 |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | | バックエンドで絞り込み |
| `--status` | | string | | 状態で絞り込み（`active`, `completed`, `error`, `paused`） |
| `--limit` | `-n` | int | | 表示する最大件数 |
| `--offset` | | int | `0` | 結果を返す前にこの件数をスキップ |
| `--json` | | bool | `false` | machine-readable JSON を出力 |

### 例

すべてのセッションを表示:

```bash
clinvk sessions list
```

バックエンドで絞り込み:

```bash
clinvk sessions list --backend claude
```

状態で絞り込み:

```bash
clinvk sessions list --status active
```

件数を制限:

```bash
clinvk sessions list --limit 10
```

フィルターを組み合わせ:

```bash
clinvk sessions list --backend claude --status active --limit 5
```

ページ付き JSON:

```bash
clinvk sessions list --backend claude --limit 10 --offset 20 --json
```

### 出力例

```text
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 minutes ago   1234         fix the bug in auth.go
def456    codex     completed  2 hours ago     5678         implement user registration
ghi789    gemini    error      1 day ago       -            failed task
```

---

## clinvk sessions show

特定のセッションの詳細を表示します。

### 使い方

```bash
clinvk sessions show <session-id> [flags]
```

### フラグ

| フラグ | 型 | デフォルト | 説明 |
|------|------|---------|-------------|
| `--json` | bool | `false` | セッション詳細を machine-readable JSON で出力 |

### 例

```bash
clinvk sessions show abc123
```

JSON 出力:

```bash
clinvk sessions show abc123 --json
```

### 出力例

```text
ID:                abc123
Backend:           claude
Model:             claude-opus-4-5-20251101
Status:            active
Created:           2025-01-27T10:00:00Z
Last Used:         2025-01-27T11:30:00Z (30 minutes ago)
Working Directory: /projects/myapp
Backend Session:   session-xyz
Turns:             3
Token Usage:
  Input:           1234
  Output:          5678
  Total:           6912
Tags:              [feature-auth, urgent]
```

---

## clinvk sessions delete

特定のセッションを削除します。

### 使い方

```bash
clinvk sessions delete <session-id>
```

### 例

```bash
clinvk sessions delete abc123
```

### 出力例

```text
Session abc123 deleted.
```

---

## clinvk sessions clean

古いセッションを削除します。

### フラグ

| フラグ | 型 | デフォルト | 説明 |
|------|------|---------|-------------|
| `--older-than` | string | | 指定日数より古いセッションを削除（例: `30` または `30d`） |
| `--dry-run` | bool | `false` | 削除せずにクリーンアップ候補をプレビュー |

指定しない場合は、設定の `session.retention_days` を使用します。

### 例

30 日より古いセッションを削除:

```bash
clinvk sessions clean --older-than 30d
```

7 日より古いセッションを削除:

```bash
clinvk sessions clean --older-than 7
```

設定のデフォルトを使用:

```bash
clinvk sessions clean
```

削除せずに候補をプレビュー:

```bash
clinvk sessions clean --older-than 30d --dry-run
```

### 出力例

```text
Deleted 15 session(s) older than 30 days.
```

dry-run 出力:

```text
Dry run: would delete 15 session(s) older than 30 days.
Sample session IDs: abc123..., def456...
No sessions were deleted. Re-run without --dry-run to apply.
Note: candidate sessions may change between dry-run and actual cleanup.
```

---

## セッション状態

| 状態 | 説明 |
|--------|-------------|
| `active` | セッションが有効で再開できる |
| `completed` | 正常に完了した |
| `error` | エラーで終了した |
| `paused` | 一時停止中（現在は非アクティブ） |

## よくあるエラー

| エラー | 原因 | 対処 |
|-------|-------|----------|
| `session not found` | セッション ID が存在しない | `clinvk sessions list` を確認 |
| `invalid status filter` | 不明な状態値 | `active`, `completed`, `error`, `paused` を使用 |
| `no sessions to clean` | 条件に一致するセッションがない | フィルターや保持期間を調整 |

## 終了コード

| コード | 説明 |
|------|-------------|
| 0 | 成功 |
| 1 | エラー（例: セッションが見つからない） |
| 4 | セッションエラー |

## 関連コマンド

- [resume](resume.md) - セッションを再開
- [prompt](prompt.md) - 新しいプロンプトを実行

## 関連項目

- [セッション管理](../../guides/sessions.md) - セッション管理ガイド
- [設定リファレンス](../configuration.md) - セッション関連の設定
