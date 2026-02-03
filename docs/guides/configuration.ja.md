# 設定ガイド

ワークフローに合わせて clinvk を設定する方法を説明します。よくあるシナリオとベストプラクティスを中心にまとめています。

## クイックセットアップ

### 1. 現在の設定を確認する

```bash
clinvk config show
```

このコマンドは、システム上で利用可能なバックエンドを含め、すべての設定を表示します。

### 2. デフォルトバックエンドを設定する

```bash
# Claude をデフォルトにする
clinvk config set default_backend claude

# もしくは Gemini を使う
clinvk config set default_backend gemini
```

### 3. 完了

基本セットアップはこれだけです。clinvk は妥当なデフォルトがあるため、そのままでもすぐ使えます。

## 設定ファイル

clinvk の設定は `~/.clinvk/config.yaml` に保存されます。直接編集するか、`clinvk config set` を使って更新できます。

### 最小構成

```yaml
# ~/.clinvk/config.yaml
default_backend: claude
```

### 推奨構成

```yaml
# ~/.clinvk/config.yaml
default_backend: claude

# 出力に実行時間を表示する
output:
  show_timing: true

# セッションを 60 日保持する
session:
  retention_days: 60
```

## よくあるシナリオ

### シナリオ 1: 複数バックエンドを使い分ける

用途ごとに AI モデルを使い分けたい場合:

```yaml
default_backend: claude

backends:
  claude:
    model: claude-opus-4-5-20251101    # 複雑な推論
  codex:
    model: o3                           # コード生成
  gemini:
    model: gemini-2.5-pro              # 一般タスク
```

**使い方:**

```bash
# デフォルト（Claude）
clinvk "analyze this architecture"

# タスクに応じてバックエンドを指定
clinvk -b codex "generate unit tests"
clinvk -b gemini "summarize this document"
```

### シナリオ 2: 自動承認モードで自動化する

CI/CD など、対話プロンプトを出したくない場合:

```yaml
unified_flags:
  approval_mode: auto    # すべてのアクションを自動承認
output:
  format: json           # 機械可読な出力
```

!!! warning "セキュリティ注意"
    `auto` 承認モードは信頼できる環境でのみ使用してください。AI がファイル操作やコマンドを実行できる可能性があります。

### シナリオ 3: 読み取り専用で解析する

コードレビューや解析で、AI にファイルを変更させたくない場合:

```yaml
unified_flags:
  sandbox_mode: read-only    # ファイル変更を禁止
  # approval_mode はプロンプトの出し方を制御するもので、アクション許可そのものではありません。
  # ファイルアクセスの制限には sandbox_mode を使います。最大限安全にするなら always も検討してください。
  approval_mode: always
```

### シナリオ 4: チームで設定を共有する

チームで一貫した設定にしたい場合、プロジェクト単位の設定を用意します。

```bash
# プロジェクトルートで
cat > .clinvk.yaml << 'EOF'
default_backend: claude
unified_flags:
  sandbox_mode: workspace    # プロジェクトファイルのみアクセス
backends:
  claude:
    system_prompt: "You are working on the MyApp project. Follow our coding standards."
EOF

# プロジェクト設定を使用
clinvk --config .clinvk.yaml "review the auth module"
```

### シナリオ 5: 連携のために HTTP サーバーを使う

clinvk を API バックエンドとして利用する場合:

```yaml
server:
  host: "127.0.0.1"          # localhost のみ（安全）
  port: 8080
  request_timeout_secs: 300  # 長いタスク向けに 5 分

# LAN で公開（注意して使用）
# server:
#   host: "0.0.0.0"          # 全インターフェース
#   port: 8080
```

### シナリオ 6: 並列タスクを最適化する

バッチ処理や多面的なレビュー向け:

```yaml
parallel:
  max_workers: 5       # 最大 5 タスク同時実行
  fail_fast: false     # 一部が失敗しても続行
  aggregate_output: true
```

### シナリオ 7: 本番 API サーバー

セキュリティと可観測性を意識した運用例:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  # リクエストサイズ制限は巨大ペイロード攻撃の緩和に有効
  max_request_body_bytes: 5242880  # 5MB
  # レート制限で濫用を抑止
  rate_limit_enabled: true
  rate_limit_rps: 20
  rate_limit_burst: 50
  # Web フロントエンド向けの CORS
  cors_allowed_origins:
    - "https://myapp.example.com"
    - "http://localhost:3000"
  # Prometheus メトリクス
  metrics_enabled: true

# 作業ディレクトリ制限
server:
  allowed_workdir_prefixes:
    - "/home/user/projects"
    - "/var/www"
  blocked_workdir_prefixes:
    - "/etc"
    - "/root"
    - "/usr/bin"
```

### シナリオ 8: レート制限の設定

インターネットへ公開する場合:

```yaml
server:
  # 基本的なレート制限
  rate_limit_enabled: true
  rate_limit_rps: 10
  rate_limit_burst: 20

  # 逆プロキシ配下の場合の信頼プロキシ設定
  trusted_proxies:
    - "127.0.0.1"
    - "10.0.0.0/8"
    - "172.16.0.0/12"
    - "192.168.0.0/16"
```

### シナリオ 9: メトリクスと可観測性

Prometheus で監視する場合:

```yaml
server:
  # Prometheus のメトリクスエンドポイントを有効化
  metrics_enabled: true

  # リソース保護のためのリクエストサイズ制限
  max_request_body_bytes: 10485760  # 10MB
```

メトリクスは `http://localhost:8080/metrics` から参照できます。

## バックエンド別の設定

### Claude Code

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all              # または: read,write,edit
    system_prompt: "Be concise."
    extra_flags:
      - "--add-dir"
      - "./docs"                    # docs をコンテキストに含める
```

**利用可能なモデル:**

| モデル | 得意分野 |
|-------|----------|
| `claude-opus-4-5-20251101` | 複雑な推論、アーキテクチャ |
| `claude-sonnet-4-20250514` | 速度と能力のバランス |

### Codex CLI

```yaml
backends:
  codex:
    model: o3
    extra_flags:
      - "--quiet"                   # 出力を控えめに
```

### Gemini CLI

```yaml
backends:
  gemini:
    model: gemini-2.5-pro
    extra_flags:
      - "--sandbox"
```

## 環境変数

環境変数で任意の設定を上書きできます。

```bash
# デフォルトバックエンドの上書き
export CLINVK_BACKEND=gemini

# モデルの上書き
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3
export CLINVK_GEMINI_MODEL=gemini-2.5-pro
```

**優先順位**（上ほど優先）:

1. CLI フラグ（`--backend codex`）
2. 環境変数（例: `CLINVK_BACKEND`）
3. 設定ファイル（`~/.clinvk/config.yaml`）
4. 組み込みデフォルト

## ベストプラクティス

### 1. まずデフォルトで使う

clinvk はデフォルトでも十分に機能します。必要な部分だけカスタマイズしてください。

### 2. プロジェクト単位の設定を使う

リポジトリに `.clinvk.yaml` を置いて、プロジェクト固有の設定を管理します。

```bash
clinvk --config .clinvk.yaml "your prompt"
```

### 3. サーバーを安全に運用する

HTTP サーバーを公開する場合:

```yaml
server:
  host: "127.0.0.1"    # リバースプロキシなしで 0.0.0.0 を使わない
```

### 4. 適切なタイムアウトを設定する

長時間タスク向け:

```yaml
server:
  request_timeout_secs: 600    # 10 分
```

### 5. レビューは読み取り専用で

解析だけで変更が不要なら:

```yaml
unified_flags:
  sandbox_mode: read-only
```

## トラブルシューティング

### 設定が反映されない

```bash
# 有効な設定を確認
clinvk config show

# 設定ファイルの場所を確認
ls -la ~/.clinvk/config.yaml
```

### バックエンドが利用できない

```bash
# 検出されたバックエンドを確認
clinvk config show | grep available

# CLI が PATH にあるか確認
which claude codex gemini
```

### デフォルトへ戻す

```bash
# 設定ファイルを削除
rm ~/.clinvk/config.yaml

# デフォルトを確認
clinvk config show
```

## 設定テンプレート

### 開発者ワークステーション

```yaml
default_backend: claude

unified_flags:
  sandbox_mode: workspace

output:
  show_timing: true
  color: true

session:
  retention_days: 30
```

### CI/CD パイプライン

```yaml
default_backend: claude

unified_flags:
  approval_mode: auto
output:
  format: json

parallel:
  max_workers: 3
  fail_fast: true
```

### API サーバー

```yaml
default_backend: claude

server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300

output:
  format: json
```

## 次のステップ

- [設定リファレンス](../reference/configuration.md) - すべてのオプションの詳細
- [環境変数](../reference/environment.md) - 環境変数一覧
- [config コマンド](../reference/cli/config.md) - CLI の設定コマンド
