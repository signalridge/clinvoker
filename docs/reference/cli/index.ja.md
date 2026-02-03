# コマンドリファレンス

clinvk の CLI コマンドを網羅したリファレンスです。

## 概要

```bash
clinvk [グローバルフラグ] [プロンプト]
clinvk [コマンド] [サブコマンド] [フラグ]
```

## コマンド一覧

| コマンド | 説明 | よくある用途 |
|---------|-------------|------------|
| [`[prompt]`](prompt.md) | プロンプトを実行（デフォルトコマンド） | 日常的な AI タスク |
| [`resume`](resume.md) | セッションを再開 | 会話を継続 |
| [`sessions`](sessions.md) | セッション管理 | 一覧/表示/削除 |
| [`config`](config.md) | 設定管理 | 設定の確認/変更 |
| [`parallel`](parallel.md) | 並列実行 | 複数タスクを同時実行 |
| [`compare`](compare.md) | バックエンド応答を比較 | 複数 AI の評価 |
| [`chain`](chain.md) | プロンプトチェーンを実行 | 複数ステップのワークフロー |
| [`serve`](serve.md) | HTTP API サーバーを起動 | アプリケーション統合 |
| `version` | バージョン情報を表示 | インストール済み版の確認 |
| `help` | ヘルプを表示 | コマンドの使い方を確認 |

## グローバルフラグ

これらのフラグはすべてのコマンドで利用できます。

| フラグ | 短縮 | 型 | デフォルト | 説明 |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | 使用する AI バックエンド |
| `--model` | `-m` | string | | 使用するモデル |
| `--workdir` | `-w` | string | | 作業ディレクトリ |
| `--output-format` | `-o` | string | `json` | 出力形式 |
| `--config` | | string | | 設定ファイルのパス |
| `--dry-run` | | bool | `false` | 実行せずコマンドのみ表示 |
| `--ephemeral` | | bool | `false` | ステートレスモード |
| `--help` | `-h` | | | ヘルプを表示 |

### プロンプト専用フラグ

これらのフラグはデフォルトのプロンプト実行時にのみ利用できます。

| フラグ | 短縮 | 型 | デフォルト | 説明 |
|------|-------|------|---------|-------------|
| `--continue` | `-c` | bool | `false` | 直近のセッションを継続する |

## フラグ詳細

### --backend, -b

使用する AI バックエンドを選択します。

```bash
clinvk --backend claude "プロンプト"
clinvk -b codex "プロンプト"
clinvk -b gemini "プロンプト"
```

利用可能なバックエンド: `claude`, `codex`, `gemini`

### --model, -m

選択したバックエンドのデフォルトモデルを上書きします。

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "プロンプト"
clinvk -b codex -m o3-mini "プロンプト"
```

### --workdir, -w

AI バックエンドの作業ディレクトリを指定します。

```bash
clinvk --workdir /path/to/project "このコードベースを解析して"
```

### --output-format, -o

出力形式を指定します。

| 値 | 説明 |
|-------|-------------|
| `text` | プレーンテキスト |
| `json` | 構造化 JSON（デフォルト） |
| `stream-json` | JSON イベントをストリーミング出力 |

```bash
clinvk --output-format json "プロンプト"
clinvk -o stream-json "プロンプト"
```

### --config

カスタムの設定ファイルを使用します。

```bash
clinvk --config /path/to/config.yaml "プロンプト"
```

### --dry-run

実行せずに、実行されるコマンドだけを表示します。

```bash
clinvk --dry-run "機能 X を実装して"
# Output: Would execute: claude --model claude-opus-4-5-20251101 "機能 X を実装して"
```

### --ephemeral

セッションを作成せず、ステートレスモードで実行します。

```bash
clinvk --ephemeral "簡単な質問"
```

## コマンドのカテゴリ

### コアコマンド

日常的に使うコマンドです。

- `[prompt]` - プロンプトを実行
- `resume` - セッションを継続
- `sessions` - セッション管理

### 設定コマンド

設定を管理するコマンドです。

- `config` - 設定の表示/変更

### 実行コマンド

高度な実行パターンのためのコマンドです。

- `parallel` - 複数タスクを同時に実行
- `chain` - 逐次パイプラインを実行
- `compare` - 複数バックエンドを比較

### サーバーコマンド

HTTP API を提供するためのコマンドです。

- `serve` - API サーバーを起動

## 使用例

### 基本

```bash
# プロンプトを実行
clinvk "auth.go のバグを修正して"

# バックエンドを指定
clinvk -b codex "機能を実装して"

# モデルを指定
clinvk -b claude -m claude-sonnet-4-20250514 "簡単にレビューして"
```

### セッション管理

```bash
# セッション一覧
clinvk sessions list

# セッション詳細を表示
clinvk sessions show abc123

# 直近のセッションを再開
clinvk resume --last

# 古いセッションを削除
clinvk sessions clean --older-than 30d
```

### 設定

```bash
# 現在の設定を表示
clinvk config show

# 値を設定
clinvk config set default_backend codex
```

### 高度な実行

```bash
# タスクを並列実行
clinvk parallel --file tasks.json

# バックエンドを比較
clinvk compare --all-backends "このコードを説明して"

# チェーンを実行
clinvk chain --file pipeline.json
```

### サーバー

```bash
# サーバーを起動
clinvk serve --port 8080
```

## 終了コード

すべてのコマンドは終了コードを返します。

| コード | 説明 |
|------|-------------|
| 0 | 成功 |
| 1 | 一般エラー |
| 2 | バックエンド利用不可 |
| 3 | 設定不正 |
| 4 | セッションエラー |

[終了コード](../exit-codes.md) に完全な一覧があります。

## ヘルプ

各コマンドのヘルプを表示します。

```bash
# 全体のヘルプ
clinvk --help

# コマンド別ヘルプ
clinvk [command] --help

# 例
clinvk parallel --help
```

## 関連項目

- [設定リファレンス](../configuration.md) - 設定項目
- [環境変数](../environment.md) - 環境変数による設定
- [終了コード](../exit-codes.md) - 終了コードの参照
