---
title: よくある質問（FAQ）
description: clinvoker のインストール、使い方、設定、連携に関するよくある質問と回答。
---

# よくある質問（FAQ）

この FAQ では、clinvoker に関するよくある質問をトピック別にまとめています。ここで解決しない場合は、[トラブルシューティング](troubleshooting.md) または [GitHub Discussions](https://github.com/signalridge/clinvoker/discussions) を参照してください。

## 一般

### clinvoker とは何ですか？

clinvoker は統一 AI CLI ラッパーで、複数の AI コーディングアシスタント（Claude Code, Codex CLI, Gemini CLI）を 1 つのインターフェースで扱えます。セッション管理、並列実行、バックエンド比較、他ツール連携のための HTTP API サーバー機能を提供します。

### 個別 CLI を直接使うのではなく、clinvoker を使う理由は？

- **統一インターフェース**: すべてのバックエンドで同じコマンド体系
- **セッション管理**: 会話を簡単に再開できる
- **並列実行**: 複数バックエンドに同時にタスクを投げられる
- **バックエンド比較**: 異なるモデルの回答を並べて比較できる
- **HTTP API**: アプリやワークフローへ統合しやすい
- **設定カスケード**: 環境をまたいでも一貫した設定管理ができる

### 対応バックエンドは？

現時点での対応バックエンド:

| バックエンド | プロバイダ | 説明 |
|---------|----------|-------------|
| Claude Code | Anthropic | コード理解に強い AI コーディングアシスタント |
| Codex CLI | OpenAI | プログラミング能力に強いコード指向 CLI |
| Gemini CLI | Google | マルチモーダル対応の Gemini AI CLI |

### clinvoker は無料ですか？

はい。clinvoker 自体は無料かつオープンソース（MIT License）です。ただし、利用する AI バックエンドはそれぞれ価格や利用制限がある場合があります。使いたいバックエンドごとに有効な API 認証情報が必要です。

### 他ツールとの比較は？

| 機能 | clinvoker | Aider | Continue |
|---------|-----------|-------|----------|
| 複数バックエンド | はい | 限定的 | 設定により可能 |
| セッション管理 | 組み込み | Git ベース | エディタベース |
| HTTP API | はい | いいえ | いいえ |
| 並列実行 | はい | いいえ | いいえ |
| バックエンド比較 | はい | いいえ | いいえ |

## インストール

### clinvoker のインストール方法は？

複数のインストール方法があります。

```bash
# Homebrew（macOS/Linux）
brew install signalridge/tap/clinvk

# Go install
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# Nix
nix run github:signalridge/clinvoker

# GitHub Releases からバイナリをダウンロード
# https://github.com/signalridge/clinvoker/releases
```

詳しくは [インストールガイド](../tutorials/getting-started.md) を参照してください。

### すべてのバックエンドを入れる必要がありますか？

いいえ。clinvoker は利用可能なバックエンドだけを使えます。使いたいものだけインストールしてください。clinvoker が自動で利用可否を検出します。

### システム要件は？

- **OS**: macOS / Linux / Windows
- **Go**: 1.24+（ソースからビルドする場合）
- **メモリ**: 50MB RAM（clinvoker 自体）
- **ディスク**: バイナリ 10MB + セッション保存領域

### 利用可能なバックエンドを確認するには？

```bash
clinvk config show
```

各バックエンドの `available: true` を確認してください。

## 使い方

### デフォルトバックエンドを変更するには？

```bash
# 設定で指定
clinvk config set default_backend codex

# 環境変数で指定
export CLINVK_BACKEND=codex

# コマンドごとに指定
clinvk -b claude "your prompt"
```

### 会話を継続するには？

```bash
# クイック継続（直前セッションを再開）
clinvk -c "follow up message"

# 特定セッションを再開
clinvk resume <session-id> "follow up message"

# 最後のセッションを再開
clinvk resume --last "follow up message"
```

### 別のモデルを使えますか？

はい。`--model` でモデルを指定できます。

```bash
# 特定モデルを使用（バックエンド CLI にそのまま渡されます）
clinvk -b claude -m sonnet "quick task"
clinvk -b gemini -m gemini-2.5-flash "quick task"
clinvk -b codex -m o3 "quick task"

# 設定でデフォルトを指定
clinvk config set backends.claude.model sonnet
```

### 並列でタスクを実行するには？

タスクファイルを作成して parallel コマンドを使います。

```bash
# tasks.json を作成
{
  "tasks": [
    {"prompt": "Review this code", "backend": "claude"},
    {"prompt": "Review this code", "backend": "codex"},
    {"prompt": "Review this code", "backend": "gemini"}
  ]
}

# 並列実行
clinvk parallel --file tasks.json --max-parallel 3
```

詳細は [並列実行ガイド](../guides/parallel.md) を参照してください。

### バックエンドの応答を比較するには？

```bash
# 利用可能なバックエンドをすべて比較
clinvk compare --all-backends "your prompt"

# 特定バックエンドを比較
clinvk compare -b claude -b codex "your prompt"

# 比較結果をファイルに保存
clinvk compare --all-backends -o comparison.md "your prompt"
```

詳細は [バックエンド比較](../guides/backends/index.md) を参照してください。

### 複数プロンプトをチェーン実行するには？

```bash
# chain.json を作成
{
  "steps": [
    {"backend": "claude", "prompt": "Generate a Python function to sort a list"},
    {"backend": "codex", "prompt": "Review and optimize this code: {{previous}}"},
    {"backend": "claude", "prompt": "Add tests for: {{previous}}"}
  ]
}

# チェーン実行
clinvk chain --file chain.json
```

詳細は [チェーン実行ガイド](../guides/chains.md) を参照してください。

## 設定

### 設定ファイルはどこですか？

デフォルトの場所: `~/.clinvk/config.yaml`

別の場所を指定することもできます。

```bash
clinvk --config /path/to/config.yaml "prompt"
```

### 設定の優先順位は？

設定は次の順（上ほど優先）で解決されます。

1. CLI フラグ
2. 環境変数
3. 設定ファイル
4. デフォルト

### 現在の設定を見るには？

```bash
clinvk config show
```

これは複数ソースをマージした「有効な設定」を表示します。

### すべての設定を環境変数で指定できますか？

はい。任意の設定キーに `CLINVK_` を付けることで上書きできます。

```bash
export CLINVK_BACKEND=codex
export CLINVK_TIMEOUT=120
export CLINVK_SERVER_PORT=3000
```

### バックエンド別の設定を行うには？

```yaml
# ~/.clinvk/config.yaml
backends:
  claude:
    model: claude-sonnet-4
    timeout: 120
  codex:
    model: gpt-5.2
    sandbox_mode: full
```

## セッション

### セッションはどこに保存されますか？

セッションは `~/.clinvk/sessions/` に JSON ファイルとして保存されます。

### セッション一覧を表示するには？

```bash
clinvk sessions list

# バックエンドで絞り込み
clinvk sessions list --backend claude

# 詳細を表示
clinvk sessions list --verbose
```

### 古いセッションを掃除するには？

```bash
# 30 日より古いセッションを削除
clinvk sessions clean --older-than 30d

# すべて削除
clinvk sessions clean --all

# 手動削除
rm -rf ~/.clinvk/sessions/*
```

### セッション追跡を無効にできますか？

はい。ephemeral モードを使います。

```bash
clinvk --ephemeral "prompt"
```

このモードではセッションを作成/ロードしません。

### セッションをエクスポートするには？

```bash
# セッションをファイルへエクスポート
clinvk sessions export <session-id> > session.json

# もしくはファイルを直接コピー
cp ~/.clinvk/sessions/<session-id>.json ./backup.json
```

## HTTP サーバー

### サーバーの起動方法は？

```bash
# デフォルト設定で起動
clinvk serve

# ポートを指定
clinvk serve --port 3000

# API キー認証を有効化
clinvk serve --api-keys "key1,key2"
```

### サーバーは認証されますか？

認証はオプションです。API キーを設定した場合、すべてのリクエストに有効なキーが必要です。

```bash
# キーを設定
export CLINVK_API_KEYS="key1,key2,key3"

# もしくは設定で（例）
clinvk config set server.api_keys "key1,key2"

# リクエストで使用
curl -H "Authorization: Bearer key1" http://localhost:8080/api/v1/prompt
```

キーが設定されていない場合、サーバーはすべてのリクエストを許可します。

### サーバーを公開するには？

リバースプロキシ（nginx/Apache/Caddy）配下に置き、API キーを有効化してください。

```bash
# 全インターフェースでバインド（注意して使用）
clinvk serve --host 0.0.0.0 --api-keys "your-secret-key"
```

### OpenAI のクライアントライブラリを使えますか？

はい。サーバーは OpenAI 互換エンドポイントを提供します。

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-api-key"  # API キーを有効化している場合に必要
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### Anthropic のクライアントライブラリを使えますか？

はい。Anthropic 互換エンドポイントも利用できます。

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="your-api-key"
)

response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}]
)
```

## 連携

### Claude Code と連携するには？

clinvoker は Claude Code と併用できます。

```bash
# clinvoker 経由で Claude Code を使う
clinvk -b claude "your prompt"

# Claude Code の対話モードを起動
claude
```

詳細は [Claude バックエンドガイド](../guides/backends/claude.md) を参照してください。

### LangChain と一緒に使うには？

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
    model="claude"
)

response = llm.invoke("Hello!")
```

詳細は [LangChain 連携ガイド](../guides/integrations/langchain-langgraph.md) を参照してください。

### CI/CD パイプラインで使うには？

```yaml
# GitHub Actions 例
- name: Code Review
  env:
    CLINVK_BACKEND: claude
  run: |
    echo '{"prompt": "Review this PR", "files": ["src/"] }' | \
    clinvk parallel --file - --output-format json
```

例は [CI/CD 連携ガイド](../guides/integrations/ci-cd/index.md) を参照してください。

## トラブルシューティング

### バックエンドが検出されないのはなぜ？

CLI が PATH にあるか確認します。

```bash
which claude codex gemini
echo $PATH
```

### 設定が反映されないのはなぜ？

優先順位（CLI フラグ > 環境変数 > 設定ファイル）を確認し、上書きがないか調べてください。

```bash
clinvk config show  # 有効な設定を表示
env | grep CLINVK   # 環境変数を表示
```

### どこで助けを得られますか？

- [トラブルシューティング](troubleshooting.md)
- [GitHub Issues](https://github.com/signalridge/clinvoker/issues)
- [GitHub Discussions](https://github.com/signalridge/clinvoker/discussions)

## コントリビュート

### どう貢献できますか？

[コントリビューティングガイド](contributing.md) を参照してください。

- 開発環境セットアップ
- コーディング規約
- テスト要件
- PR プロセス

### 新しいバックエンドを追加するには？

1. `Backend` インターフェースを実装
2. レジストリへ追加
3. 統一オプションのマッピングを追加
4. テストを追加
5. ドキュメントを更新

インターフェースの詳細は [バックエンドシステム](backend-system.md) を参照してください。

### バグ報告はどうすればよいですか？

[GitHub](https://github.com/signalridge/clinvoker/issues) に issue を作成し、次を含めてください。

- clinvk バージョン（`clinvk version`）
- OS とバージョン
- バックエンドのバージョン
- 再現手順
- エラーメッセージ
- デバッグログ（可能なら）

## 関連ドキュメント

- [トラブルシューティング](troubleshooting.md) - よくある問題と対処
- [ガイド](../guides/index.md) - How-to ガイド
- [リファレンス](../reference/index.md) - API/CLI リファレンス
- [コンセプト](index.md) - アーキテクチャと設計
