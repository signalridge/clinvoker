# clinvk config

設定を管理します。

## 用法

```bash
clinvk config [command] [flags]
```

## 説明

`config` コマンドは、clinvk の設定を表示・変更するためのサブコマンドを提供します。設定はデフォルトで `~/.clinvk/config.yaml` に保存されます。

## サブコマンド

| コマンド | 説明 |
|---------|-------------|
| `show` | 現在の設定を表示 |
| `get` | 特定の設定値を取得 |
| `set` | 設定値を変更 |

---

## clinvk config show

現在の設定を表示します。

### 使い方

```bash
clinvk config show
```

### 出力例

```yaml
default_backend: claude

unified_flags:
  approval_mode: default
  sandbox_mode: default
  verbose: false
  dry_run: false
  max_turns: 0
  max_tokens: 0
  command_timeout_secs: 0

backends:
  claude:
    model: claude-opus-4-5-20251101
    enabled: true
  codex:
    model: o3
    enabled: true
  gemini:
    model: gemini-2.5-pro
    enabled: true

session:
  retention_days: 30
  store_token_usage: true

output:
  format: json
  show_tokens: false
  show_timing: false
  color: true

server:
  host: 127.0.0.1
  port: 8080

parallel:
  max_workers: 3
  fail_fast: false
```

---

## clinvk config get

特定の設定値を取得します。

### 使い方

```bash
clinvk config get <key>
```

### キーの形式

ネストしたキーにはドット記法を使用します。

```bash
clinvk config get default_backend
clinvk config get backends.claude.model
clinvk config get session.retention_days
```

### 例

既定のバックエンドを取得:

```bash
clinvk config get default_backend
# Output: claude
```

Claude のモデルを取得:

```bash
clinvk config get backends.claude.model
# Output: claude-opus-4-5-20251101
```

セッション保持日数を取得:

```bash
clinvk config get session.retention_days
# Output: 30
```

---

## clinvk config set

設定値を変更します。

### 使い方

```bash
clinvk config set <key> <value>
```

### キーの形式

ネストしたキーにはドット記法を使用します。

```bash
clinvk config set default_backend codex
clinvk config set backends.claude.model claude-sonnet-4-20250514
clinvk config set session.retention_days 7
```

### 例

既定のバックエンドを設定:

```bash
clinvk config set default_backend codex
```

Claude のモデルを設定:

```bash
clinvk config set backends.claude.model claude-sonnet-4-20250514
```

セッション保持日数を設定:

```bash
clinvk config set session.retention_days 7
```

並列ワーカー数を設定:

```bash
clinvk config set parallel.max_workers 5
```

詳細出力を有効化:

```bash
clinvk config set unified_flags.verbose true
```

### 値の型

| 型 | 例 | 備考 |
|------|---------|-------|
| 文字列 | `"claude"` | 空白を含む場合などは引用符が必要 |
| 整数 | `30` | 引用符不要 |
| 真偽値 | `true`, `false` | 引用符不要 |
| 配列 | `["item1", "item2"]` | YAML の配列構文 |

---

## 設定ファイルの場所

デフォルトの場所: `~/.clinvk/config.yaml`

別のファイルを使う場合は `--config` フラグを指定します。

```bash
clinvk --config /path/to/config.yaml config show
```

## 設定の優先順位

設定値は次の順序で解決されます（上ほど優先度が高い）。

1. **CLI フラグ** - コマンドライン引数
2. **環境変数** - `CLINVK_*` 変数
3. **設定ファイル** - `~/.clinvk/config.yaml`
4. **デフォルト値** - 組み込み既定値

## よくあるエラー

| エラー | 原因 | 対処 |
|-------|-------|----------|
| `key not found` | 設定キーが存在しない | スペルを確認し、ドット記法を使用 |
| `invalid value` | 値の型が一致しない | キーに合った型で指定 |
| `config file not found` | 設定ファイルが見つからない | `clinvk config set` で作成する |
| `permission denied` | 設定ファイルに書き込めない | ファイル権限を確認 |

## 終了コード

| コード | 説明 |
|------|-------------|
| 0 | 成功 |
| 1 | キーまたは値が不正 |
| 3 | 設定エラー |

## 関連コマンド

- [prompt](prompt.md) - 設定に基づいてプロンプトを実行
- [serve](serve.md) - 設定に基づいてサーバーを起動

## 関連項目

- [設定リファレンス](../configuration.md) - 設定項目の完全版
- [環境変数](../environment.md) - 環境変数による設定
