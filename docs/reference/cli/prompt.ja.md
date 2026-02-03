# clinvk [prompt]

AI バックエンドでプロンプトを実行します。

## 用法

```bash
clinvk [flags] [prompt]
```

## 説明

ルートコマンドは、設定されたバックエンドを使ってプロンプトを実行します。セッションの永続化と出力形式の指定をサポートします。直近のセッションを継続するには `-c` フラグを使用します。

これはデフォルトコマンドです。`clinvk` の後にテキストを続けて実行すると、そのままプロンプトとして扱われます。

## フラグ

| フラグ | 短縮 | 型 | デフォルト | 説明 |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | 使用する AI バックエンド（`claude`, `codex`, `gemini`） |
| `--model` | `-m` | string | | 選択したバックエンドのモデルを上書き |
| `--workdir` | `-w` | string | | バックエンドへ渡す作業ディレクトリ |
| `--output-format` | `-o` | string | `json` | 出力形式: `text`, `json`, `stream-json` |
| `--continue` | `-c` | bool | `false` | 直近の再開可能なセッションを継続 |
| `--dry-run` | | bool | `false` | 実行せずにバックエンドのコマンドを表示 |
| `--ephemeral` | | bool | `false` | ステートレスモード: セッションを永続化しない |
| `--config` | | string | `~/.clinvk/config.yaml` | カスタム設定ファイルのパス |

## 例

### 基本

シンプルなプロンプトを実行します。

```bash
clinvk "fix the bug in auth.go"
```

### バックエンドを指定する

特定のバックエンドを使用します。

```bash
clinvk --backend codex "implement user registration"
clinvk -b gemini "explain this algorithm"
```

### モデルを指定する

デフォルトモデルを上書きします。

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "quick review"
clinvk -b codex -m o3-mini "simple task"
```

### セッションを継続する

以前のセッションの続きから実行します。

```bash
# Start a session
clinvk "implement the login feature"

# Continue the session
clinvk -c "now add password validation"

# Continue again
clinvk -c "add rate limiting"
```

### JSON 出力

構造化された JSON を出力します。

```bash
clinvk --output-format json "explain this code"
```

### Dry Run

実際に実行せず、実行されるコマンドを確認します。

```bash
clinvk --dry-run "implement feature X"
# Output: Would execute: claude --model claude-opus-4-5-20251101 "implement feature X"
```

### ステートレスモード

セッションを作成せずに実行します。

```bash
clinvk --ephemeral "what is 2+2"
```

### 作業ディレクトリを指定する

作業ディレクトリを指定します。

```bash
clinvk --workdir /path/to/project "review the codebase"
```

## 出力

### text 形式

`--output-format text` を指定すると、応答テキストのみが表示されます。

```text
The code implements a binary search algorithm...
```

### json 形式

```json
{
  "backend": "claude",
  "content": "The response text...",
  "session_id": "abc123",
  "model": "claude-opus-4-5-20251101",
  "duration_seconds": 2.5,
  "exit_code": 0,
  "usage": {
    "input_tokens": 123,
    "output_tokens": 456,
    "total_tokens": 579
  },
  "raw": {
    "events": []
  }
}
```

### stream-json 形式

`stream-json` はバックエンドのネイティブなストリーミング形式（NDJSON/JSONL）をそのまま通します。イベントの形はバックエンド CLI に依存し、統一されていません。

## よくあるエラー

| エラー | 原因 | 対処 |
|-------|-------|----------|
| `backend not found` | バックエンド CLI が未インストール | バックエンドをインストール（例: `npm install -g @anthropic-ai/claude-code`） |
| `session not resumable` | セッションが再開に対応していない | 新しいセッションを開始する |
| `timeout` | 実行に時間がかかりすぎた | 設定の `command_timeout_secs` を増やす |
| `invalid output format` | 不明な形式が指定された | `text`, `json`, `stream-json` のいずれかを使う |

## 終了コード

| コード | 説明 |
|------|-------------|
| 0 | 成功 |
| 1 | エラー |
| 2+ | バックエンドの終了コード（バックエンドプロセスが非 0 で終了した場合に伝播） |

詳細は [終了コード](../exit-codes.md) を参照してください。

## 関連コマンド

- [resume](resume.md) - セッションを再開
- [sessions](sessions.md) - セッション管理
- [config](config.md) - 既定値を設定

## 関連項目

- [設定リファレンス](../configuration.md) - 既定値の設定
- [環境変数](../environment.md) - 環境変数による設定
