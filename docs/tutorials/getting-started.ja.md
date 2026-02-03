---
title: はじめに
description: clinvoker のインストール、最初のプロンプト実行、コア概念の理解までを網羅したガイド。
---

# clinvoker を始める

clinvoker へようこそ。clinvoker は、複数の AI コーディングアシスタントをオーケストレーションするための統合コマンドラインインターフェースです。このチュートリアルでは、インストール、設定、そして最初の操作までを順を追って解説します。

## clinvoker とは？

clinvoker（発音: "see-el-in-voker"）は、Claude Code / Codex CLI / Gemini CLI など複数の AI コーディングアシスタントへのアクセスを 1 つにまとめるユニバーサルゲートウェイです。各ツールごとに異なるコマンドや UI を覚える代わりに、1 つの一貫したインターフェースで利用できます。

### 主な利点

- **統一インターフェース**: すべてのバックエンドで同じコマンド構造
- **セッション管理**: 会話コンテキストを永続化して再開
- **並列実行**: 複数の AI タスクを同時に実行
- **チェーンワークフロー**: あるバックエンドの出力を別のバックエンドへ渡す
- **HTTP API**: 連携のためにサービスとしてデプロイ可能

---

## 前提条件

clinvoker をインストールする前に、次を確認してください。

### システム要件

| 要件 | バージョン | 備考 |
|-------------|---------|-------|
| Go | 1.24+ | ソースからビルドする場合のみ必要 |
| OS | Linux / macOS / Windows | AMD64 / ARM64 をサポート |
| メモリ | 最低 512MB | CLI 実行のため |
| ディスク容量 | 100MB | バイナリと設定のため |

### バックエンドの前提条件

clinvoker を有効に使うには、少なくとも 1 つの AI バックエンドが必要です。

| バックエンド | インストールコマンド | ドキュメント |
|---------|---------------------|---------------|
| Claude Code | `npm install -g @anthropic-ai/claude-code` | [Claude.ai](https://claude.ai/claude-code) |
| Codex CLI | `npm install -g @openai/codex` | [GitHub](https://github.com/openai/codex-cli) |
| Gemini CLI | `npm install -g @google/gemini-cli` | [GitHub](https://github.com/google/gemini-cli) |

バックエンドのインストールを確認します。

```bash
# Claude Code を確認
which claude && claude --version

# Codex CLI を確認
which codex && codex --version

# Gemini CLI を確認
which gemini && gemini --version
```

---

## インストール方法

環境に合う方法を選んでください。

### 方法の比較

| 方法 | 適している人 | 長所 | 短所 |
|--------|----------|------|------|
| クイックインストールスクリプト | 初めての方 | 最速で導入、PATH も自動設定 | curl/PowerShell が必要 |
| パッケージマネージャー | 継続利用 | 更新が容易、依存関係を管理 | 最新版でない場合がある |
| 手動ダウンロード | 閉域/隔離環境 | バージョンを完全に制御 | 手動更新が必要 |
| ソースからビルド | 開発者 | 最新機能、カスタマイズ | Go ツールチェーンが必要 |

### 1. クイックインストール（推奨）

最速で始める方法です。

=== "macOS/Linux"

    ```bash
    curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
    ```

    このスクリプトは次を行います。
    1. OS とアーキテクチャを判定
    2. 適切なバイナリをダウンロード
    3. `~/.local/bin`（または sudo を使う場合は `/usr/local/bin`）へインストール
    4. 必要に応じて PATH を更新

=== "Windows (PowerShell)"

    ```powershell
    irm https://raw.githubusercontent.com/signalridge/clinvoker/main/install.ps1 | iex
    ```

    PowerShell スクリプトも同様の手順を実行し、`%LOCALAPPDATA%\Programs\clinvk` にインストールします。

### 2. パッケージマネージャー

#### Homebrew (macOS/Linux)

```bash
# tap を追加
brew tap signalridge/tap

# clinvoker をインストール
brew install clinvk

# 後で更新
brew upgrade clinvk
```

#### Scoop（Windows）

```bash
# bucket を追加
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket

# インストール
scoop install clinvk

# 更新
scoop update clinvk
```

#### Nix（Linux/macOS）

```bash
# インストールせずに直接実行
nix run github:signalridge/clinvoker

# profile にインストール
nix profile install github:signalridge/clinvoker

# flake.nix で使う
{
  inputs.clinvoker.url = "github:signalridge/clinvoker";
  nixpkgs.overlays = [ clinvoker.overlays.default ];
}
```

#### Arch Linux（AUR）

```bash
# yay を使う（推奨）
yay -S clinvk-bin

# またはソースからビルド
yay -S clinvk
```

#### Debian/Ubuntu

```bash
# Releases ページからダウンロード
wget https://github.com/signalridge/clinvoker/releases/download/v0.1.0/clinvk_0.1.0_amd64.deb
sudo dpkg -i clinvk_*.deb
```

#### RPM 系（Fedora/RHEL）

```bash
# Releases ページからダウンロード
wget https://github.com/signalridge/clinvoker/releases/download/v0.1.0/clinvk-0.1.0-1.x86_64.rpm
sudo rpm -i clinvk-*.rpm
```

### 3. 手動ダウンロード

[GitHub Releases](https://github.com/signalridge/clinvoker/releases) からビルド済みバイナリをダウンロードします。

=== "Linux (AMD64)"

    ```bash
    VERSION="0.1.0-alpha"
    curl -LO "https://github.com/signalridge/clinvoker/releases/download/v${VERSION}/clinvoker_${VERSION}_linux_amd64.tar.gz"
    tar xzf "clinvoker_${VERSION}_linux_amd64.tar.gz"
    sudo mv clinvk /usr/local/bin/
    ```

=== "macOS (ARM64 - Apple Silicon)"

    ```bash
    VERSION="0.1.0-alpha"
    curl -LO "https://github.com/signalridge/clinvoker/releases/download/v${VERSION}/clinvoker_${VERSION}_darwin_arm64.tar.gz"
    tar xzf "clinvoker_${VERSION}_darwin_arm64.tar.gz"
    sudo mv clinvk /usr/local/bin/
    ```

=== "Windows"

    `clinvoker_<version>_windows_amd64.zip` をダウンロードし、PATH が通っているディレクトリへ展開してください。

### 4. ソースからビルド

Go 1.24 以上が必要です。

```bash
# go install を使う
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# または clone してビルド
git clone https://github.com/signalridge/clinvoker.git
cd clinvoker
go build -o clinvk ./cmd/clinvk
sudo mv clinvk /usr/local/bin/
```

---

## インストール確認

インストール後、動作確認を行います。

```bash
# バージョンを確認
clinvk version
```

期待される出力例:

```text
clinvk version v0.1.0-alpha
  commit: abc1234
  built:  2025-01-27T00:00:00Z
```

検出されたバックエンドを確認します。

```bash
clinvk config show
```

システムにインストールされている内容に応じて、利用可能なバックエンド一覧が表示されます。

---

## 環境設定

### API キー

各バックエンドには、それぞれの API キー設定が必要です。

| バックエンド | 設定方法 | 環境変数 |
|---------|---------------------|---------------------|
| Claude | `claude config set api_key <key>` | `ANTHROPIC_API_KEY` |
| Codex | `codex config set api_key <key>` | `OPENAI_API_KEY` |
| Gemini | `gemini config set api_key <key>` | `GOOGLE_API_KEY` |

### 既定のバックエンド

よく使うバックエンドを既定値として設定します。

```bash
# 環境変数で指定（暫定）
export CLINVK_BACKEND=claude

# 設定で指定（永続）
clinvk config set default_backend claude
```

### 設定ファイル

`~/.clinvk/config.yaml` を作成します。

```yaml
# -b を指定しない場合の既定バックエンド
default_backend: claude

# unified_flags はすべてのバックエンドに適用される共通フラグ
unified_flags:
  approval_mode: default
  sandbox_mode: default

# バックエンド固有の設定
backends:
  claude:
    model: claude-sonnet-4-20250514
  codex:
    model: o3
  gemini:
    model: gemini-2.5-pro

# セッション管理
session:
  retention_days: 30
```

---

## 最初のプロンプト

### 基本的な使い方

まずは既定のバックエンドで最初のプロンプトを実行してみましょう。

```bash
clinvk "ソフトウェア工学における SOLID 原則を説明して"
```

このコマンドはプロンプトを既定のバックエンド（デフォルトでは Claude Code）へ送信し、応答を表示します。

### バックエンドを指定する

得意分野に応じてバックエンドを使い分けることができます。

```bash
# Claude は複雑な推論やアーキテクチャ設計が得意
clinvk -b claude "EC プラットフォーム向けのマイクロサービスアーキテクチャを設計して"

# Codex はコード生成に最適化
clinvk -b codex "Python でクイックソートを実装して"

# Gemini は幅広い知識と説明が得意
clinvk -b gemini "SQL と NoSQL データベースのトレードオフを説明して"
```

### 作業ディレクトリ

作業ディレクトリを指定してコンテキストを与えられます。

```bash
# 現在のディレクトリのコードをレビュー
clinvk -w . "このコードベースをセキュリティ観点でレビューして"

# 特定プロジェクトを解析
clinvk -w /path/to/project "このアプリケーションのアーキテクチャを説明して"
```

---

## 出力形式について

clinvoker は 3 つの出力形式をサポートしており、用途に応じて使い分けられます。

### text 形式（デフォルト）

対話的な利用に向いた、人が読みやすい出力です。

```bash
clinvk -o text "REST API を説明して"
```

**特徴:**
- 整形された読みやすいテキスト
- メタデータや構造を含まない
- ターミナルで読む用途に最適
- 他ツールへのパイプにも向く

**使いどころ:** 対話的な利用、応答の閲覧、簡単な問い合わせ

### json 形式

プログラムから扱いやすい、メタデータ付きの構造化出力です。

```bash
clinvk -o json "REST API を説明して"
```

**出力構造:**

```json
{
  "output": "REST (Representational State Transfer)...",
  "backend": "claude",
  "model": "claude-sonnet-4-20250514",
  "duration_ms": 2450,
  "tokens_used": 450,
  "session_id": "sess_abc123",
  "timestamp": "2025-01-27T10:30:00Z"
}
```

**使いどころ:** スクリプト、ログ、結果の保存、API 連携

### stream-json 形式

長時間タスク向けのリアルタイムストリーミング出力です。

```bash
clinvk -o stream-json "Go の並行処理について包括的なガイドを書いて"
```

**特徴:**
- 利用可能になった順に JSON オブジェクトを出力
- 進捗をリアルタイムに表示
- 各チャンクが応答の一部を含む
- 最終オブジェクトに完全なメタデータが含まれる

**使いどころ:** 長文生成、リアルタイムアプリケーション、進捗監視

### 形式の比較

| 形式 | 人間が読める | 機械が読める | リアルタイム | メタデータ |
|--------|---------------|------------------|-----------|----------|
| text | はい | いいえ | いいえ | いいえ |
| json | いいえ | はい | いいえ | はい |
| stream-json | 一部 | はい | はい | はい |

---

## セッション管理の基本

### セッションとは

セッションは、AI バックエンドとの会話コンテキストを永続化したものです。clinvoker は自動的に次を行います。

1. プロンプト実行時に新しいセッションを作成する
2. 以降のプロンプトを同じセッションへ関連付ける
3. 複数回のやり取りをまたいでコンテキストを維持する

### セッション一覧

有効なセッションを一覧表示します。

```bash
clinvk sessions list
```

出力例:

```text
ID          BACKEND  CREATED              STATUS   TAGS
sess_abc12  claude   2025-01-27 10:00:00  active   project-x
sess_def34  codex    2025-01-27 09:30:00  closed   -
```

### セッションを継続する

以前の会話を再開します。

```bash
# 直近のセッションを継続
clinvk --continue "キャッシュ戦略はどうすればいい？"

# または resume コマンドを使う
clinvk resume --last

# 特定のセッションを再開
clinvk resume sess_abc12
```

### セッション運用のベストプラクティス

- **タグを活用** して、プロジェクト/トピックごとに整理する
- **古いセッションを定期的に削除** してディスク容量を節約する
- 永続化が不要な単発問い合わせには **`--ephemeral` を使用** する

---

## トラブルシューティング

### 問題 1: "Backend not available"

**症状:**
```text
Error: backend "claude" not available
```

**原因と対処:**

1. **バックエンドが未インストール**
   ```bash
   # インストール確認
   which claude

   # 未導入ならインストール
   npm install -g @anthropic-ai/claude-code
   ```

2. **バックエンドが PATH にない**
   ```bash
   # バイナリを探す
   find /usr -name "claude" 2>/dev/null

   # PATH に追加
   export PATH="$PATH:/path/to/claude"
   ```

3. **設定でバックエンドが無効化されている**
   ```bash
   # 設定を確認
   clinvk config show

   # バックエンドを有効化
   clinvk config set backends.claude.enabled true
   ```

### 問題 2: "API key not configured"

**症状:**
```text
Error: authentication failed for backend "claude"
```

**対処:**

```bash
# Claude の API キーを設定
claude config set api_key $ANTHROPIC_API_KEY

# または環境変数で設定
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 問題 3: "Session not found"

**症状:**
```text
Error: no sessions found for backend "claude"
```

**説明:** 初回実行時は正常です。セッションは最初にプロンプトが成功した後に作成されます。

**対処:**

```bash
# 最初のプロンプトを実行してセッションを作成
clinvk "Hello, world!"

# 以降、セッションが利用可能になります
clinvk sessions list
```

それでも問題が続く場合は次を確認してください。

```bash
# セッションディレクトリを確認
ls -la ~/.clinvk/sessions/

# 破損している場合はクリーンアップ
clinvk sessions cleanup
```

---

## 次のステップ

clinvoker のインストールと動作確認ができたら、目的に応じて次の道筋を進めてください。

### コードレビュー自動化をしたい場合

1. [マルチバックエンドのコードレビュー](multi-backend-code-review.md) - 並列レビューを構築
2. [CI/CD 統合](ci-cd-integration.md) - パイプラインで自動化
3. [並列実行](../guides/parallel.md) - 複数レビューを同時に実行

### ツール連携をしたい場合

1. [LangChain 連携](langchain-integration.md) - LangChain と接続する
2. [HTTP サーバー](../guides/http-server.md) - API サービスとしてデプロイする
3. [Claude Code Skills](../guides/integrations/claude-code-skills.md) - カスタムスキルを作る

### 複雑なワークフローを構築したい場合

1. [チェーン実行](../guides/chains.md) - 複数ステップのパイプラインを作る
2. [AI スキルの構築](building-ai-skills.md) - 特化型 AI エージェントを開発する
3. [アーキテクチャ概要](../concepts/architecture.md) - 内部構造を理解する

### クイックリファレンス

```bash
# 基本
clinvk "ここにプロンプトを入力"
clinvk -b codex "コードを生成して"

# 並列実行
clinvk parallel -f tasks.json

# チェーンワークフロー
clinvk chain -f pipeline.json

# サーバーモード
clinvk serve --port 8080

# セッション管理
clinvk sessions list
clinvk resume --last
```

---

## まとめ

ここまでで、次を完了しました。

- 好みの方法で clinvoker をインストール
- API キーと既定バックエンドを設定
- 複数バックエンドで最初のプロンプトを実行
- 出力形式と用途を把握
- セッション管理の基本を理解
- よくある問題の対処方法を確認

clinvoker は複数の AI アシスタントを 1 つのインターフェースで統合し、並列実行、チェーン処理、CI/CD 統合といった強力なワークフローを実現します。次のチュートリアルでは、これらの機能を実際のシナリオで活用する方法を紹介します。
