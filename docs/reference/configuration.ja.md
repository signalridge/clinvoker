# 設定リファレンス

clinvk のすべての設定オプションを網羅したリファレンスです。

## 設定ファイルの場所

デフォルトの場所: `~/.clinvk/config.yaml`

`--config` フラグで任意のパスを指定できます。

```bash
clinvk --config /path/to/config.yaml "prompt"
```

## 設定例（フル）

```yaml
# --backend が指定されない場合に使用するデフォルトのバックエンド
default_backend: claude

# 統一フラグ（unified_flags）はすべてのバックエンドに適用されます
unified_flags:
  # ツール/アクション実行時の承認モード
  # 値: default, auto, none, always
  approval_mode: default

  # ファイル/ネットワークアクセスのサンドボックスモード
  # 値: default, read-only, workspace, full
  sandbox_mode: default

  # 詳細出力を有効化
  verbose: false

  # ドライラン（実行せずにコマンドを表示）
  dry_run: false

  # エージェント的なターン数の上限（0 = 無制限）
  max_turns: 0

  # 最大応答トークン数（0 = バックエンド既定。現状未マップ）
  max_tokens: 0

  # コマンドタイムアウト（秒、0 = タイムアウトなし）
  command_timeout_secs: 0

# バックエンド別の設定
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    approval_mode: ""
    sandbox_mode: ""
    enabled: true
    system_prompt: ""
    extra_flags: []

  codex:
    model: o3
    enabled: true
    extra_flags: []

  gemini:
    model: gemini-2.5-pro
    enabled: true
    extra_flags: []

# セッション管理
session:
  retention_days: 30
  store_token_usage: true
  default_tags: []

# 出力表示
output:
  format: json
  show_tokens: false
  show_timing: false
  color: true

# HTTP サーバー設定
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  # gopass で API キーを提供（空なら無効）
  api_keys_gopass_path: ""
  # レート制限
  rate_limit_enabled: false
  rate_limit_rps: 10
  rate_limit_burst: 20
  rate_limit_cleanup_secs: 180
  # セキュリティ
  trusted_proxies: []
  max_request_body_bytes: 10485760
  # CORS
  cors_allowed_origins: []
  cors_allow_credentials: false
  cors_max_age: 300
  # 作業ディレクトリ制限
  allowed_workdir_prefixes: []
  blocked_workdir_prefixes: []
  # 可観測性
  metrics_enabled: false

# 並列実行
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

---

## グローバル設定

### default_backend

| オプション | 型 | デフォルト | 説明 |
|--------|------|---------|-------------|
| `default_backend` | string | `claude` | `--backend` が指定されない場合のデフォルトバックエンド |

利用可能な値: `claude`, `codex`, `gemini`

```yaml
default_backend: claude
```

---

## 統一フラグ（unified_flags）

バックエンド別に上書きされない限り、すべてのバックエンドに適用されるグローバルオプションです。

### approval_mode

バックエンドがアクション実行前に承認を求めるタイミングを制御します。

| 値 | 説明 | 安全度 |
|-------|-------------|--------------|
| `default` | バックエンドに委ねる | 中 |
| `auto` | プロンプトを減らす / 安全な範囲で自動承認 | 低〜中 |
| `none` | 承認を一切求めない | **危険** |
| `always` | 常に承認を求める | 高 |

**バックエンドへのマッピング:**

| バックエンド | `auto` | `none` | `always` |
|---------|--------|--------|----------|
| Claude | `--permission-mode acceptEdits` | `--permission-mode dontAsk` | `--permission-mode default` |
| Codex | `--ask-for-approval on-request` | `--ask-for-approval never` | `--ask-for-approval untrusted` |
| Gemini | `--approval-mode auto_edit` | `--yolo` | `--approval-mode default` |

!!! warning "セキュリティ警告"
    `approval_mode: none` は承認プロンプトを無効化し、確認なしに編集/コマンド実行を許可する可能性があります。慎重に使用し、より安全に運用したい場合は `sandbox_mode: read-only` の併用も検討してください。

### sandbox_mode

ファイルシステムアクセスの制限を制御します。

| 値 | 説明 | ファイルアクセス |
|-------|-------------|-------------|
| `default` | バックエンドに委ねる | バックエンドに依存 |
| `read-only` | 読み取り専用 | 読み取りのみ |
| `workspace` | プロジェクトディレクトリのみ | プロジェクト内 |
| `full` | フルアクセス | 制限なし |

**バックエンド別の注意点:**

- **Claude**: `sandbox_mode` は CLI フラグにマップされません（`allowed_dirs` と承認設定を利用してください）
- **Gemini**: `read-only` と `workspace` はどちらも `--sandbox` にマップされます（区別なし）
- **Codex**: `--sandbox read-only|workspace-write|danger-full-access` にマップされます

### verbose

| オプション | 型 | デフォルト | 説明 |
|--------|------|---------|-------------|
| `verbose` | boolean | `false` | バックエンドの詳細出力を有効化 |

### dry_run

| オプション | 型 | デフォルト | 説明 |
|--------|------|---------|-------------|
| `dry_run` | boolean | `false` | 実行せずに、実行されるコマンドを表示 |

### max_turns

| オプション | 型 | デフォルト | 説明 |
|--------|------|---------|-------------|
| `max_turns` | integer | `0` | エージェント的なターン数の最大（0 = 無制限） |

### max_tokens

| オプション | 型 | デフォルト | 説明 |
|--------|------|---------|-------------|
| `max_tokens` | integer | `0` | 最大応答トークン数（0 = バックエンド既定） |

!!! note
    `max_tokens` は受け付けますが、現状はどのバックエンドの CLI フラグにもマッピングされていません。バックエンドによっては無視される場合があります。

### command_timeout_secs

| オプション | 型 | デフォルト | 説明 |
|--------|------|---------|-------------|
| `command_timeout_secs` | integer | `0` | バックエンドコマンドの最大実行時間（秒、0 = タイムアウトなし） |

---

## バックエンド別設定

`backends` セクションでバックエンド個別の設定を行います。

### バックエンドフィールド

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `model` | string | バックエンド既定 | このバックエンドのデフォルトモデル |
| `allowed_tools` | string | `all` | カンマ区切りの一覧または `all`（**Claude のみ**） |
| `approval_mode` | string | `\"\"` | 統一 `approval_mode` の上書き（空 = 統一設定を使用） |
| `sandbox_mode` | string | `\"\"` | 統一 `sandbox_mode` の上書き（空 = 統一設定を使用） |
| `enabled` | boolean | `true` | バックエンド有効/無効（無効化されたバックエンドは実行対象から除外されます） |
| `system_prompt` | string | `\"\"` | このバックエンドのデフォルトシステムプロンプト |
| `extra_flags` | array | `[]` | バックエンドへ渡す追加 CLI フラグ |

### バックエンド設定例

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    approval_mode: ""
    sandbox_mode: ""
    enabled: true
    system_prompt: "You are a helpful coding assistant."
    extra_flags:
      - "--add-dir"
      - "./docs"

  codex:
    model: o3
    enabled: true
    extra_flags:
      - "--quiet"

  gemini:
    model: gemini-2.5-pro
    enabled: true
    extra_flags:
      - "--sandbox"
```

!!! note "allowed_tools の制限"
    `allowed_tools` は現状 Claude バックエンドでのみサポートされます。Codex や Gemini に設定しても効果はなく、警告がログに出力されます。

---

## セッション設定

セッション永続化と管理を設定します。直前のセッションを再開するには `-c` または `--continue` を使います。

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `retention_days` | integer | `30` | セッション保持日数（0 = 無期限） |
| `store_token_usage` | boolean | `true` | トークン使用量統計を追跡/保存 |
| `default_tags` | array | `[]` | 新規セッションのデフォルトタグ |

```yaml
session:
  retention_days: 30
  store_token_usage: true
  default_tags: []
```

---

## 出力設定

出力表示の好みを設定します。

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `format` | string | `json` | デフォルト出力形式（`text`, `json`, `stream-json`） |
| `show_tokens` | boolean | `false` | 出力にトークン使用量を表示 |
| `show_timing` | boolean | `false` | 実行時間を表示 |
| `color` | boolean | `true` | 色付き出力を有効化 |

```yaml
output:
  format: json
  show_tokens: false
  show_timing: false
  color: true
```

---

## サーバー設定

HTTP API サーバー（`clinvk serve`）の設定です。

### 接続設定

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `host` | string | `127.0.0.1` | バインドアドレス |
| `port` | integer | `8080` | リッスンポート |

### タイムアウト設定

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `request_timeout_secs` | integer | `300` | リクエスト処理タイムアウト |
| `read_timeout_secs` | integer | `30` | 読み取りタイムアウト |
| `write_timeout_secs` | integer | `300` | 書き込みタイムアウト |
| `idle_timeout_secs` | integer | `120` | アイドル接続タイムアウト |

### レート制限

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `rate_limit_enabled` | boolean | `false` | IP 単位のレート制限を有効化 |
| `rate_limit_rps` | integer | `10` | IP あたりの毎秒リクエスト数 |
| `rate_limit_burst` | integer | `20` | バーストサイズ |
| `rate_limit_cleanup_secs` | integer | `180` | レート制限エントリのクリーンアップ間隔 |

### セキュリティ設定

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `trusted_proxies` | array | `[]` | 信頼するプロキシ（空ならプロキシヘッダーを無視） |
| `max_request_body_bytes` | integer | `10485760` | リクエストボディ最大サイズ（0 = 無制限） |
| `api_keys_gopass_path` | string | `\"\"` | API キーの gopass パス |

### CORS 設定

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `cors_allowed_origins` | array | `[]` | 許可する CORS オリジン（空 = localhost のみ） |
| `cors_allow_credentials` | boolean | `false` | CORS リクエストで資格情報を許可 |
| `cors_max_age` | integer | `300` | プリフライトキャッシュ最大秒数 |

### 作業ディレクトリ制限

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `allowed_workdir_prefixes` | array | `[]` | 許可する作業ディレクトリのプレフィックス |
| `blocked_workdir_prefixes` | array | `[]` | ブロックする作業ディレクトリのプレフィックス |

### 可観測性

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `metrics_enabled` | boolean | `false` | Prometheus の `/metrics` を有効化 |

!!! note "API キー"
    API キーは `CLINVK_API_KEYS` 環境変数（カンマ区切り）または `server.api_keys_gopass_path` で提供できます。セキュリティの観点から、キーは設定ファイルへ直接保存しません。

---

## 並列設定

並列実行のデフォルト挙動を設定します。

| フィールド | 型 | デフォルト | 説明 |
|-------|------|---------|-------------|
| `max_workers` | integer | `3` | 同時タスク数の上限 |
| `fail_fast` | boolean | `false` | 最初の失敗で停止 |
| `aggregate_output` | boolean | `true` | まとめの出力にタスク結果を集約 |

```yaml
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

---

## 設定の優先順位

値は次の順序（上ほど優先）で解決されます。

1. **CLI フラグ** - `clinvk --backend codex`
2. **環境変数** - `CLINVK_BACKEND=codex`
3. **設定ファイル** - `~/.clinvk/config.yaml`
4. **デフォルト** - 組み込みデフォルト

## 関連項目

- [環境変数](environment.md) - 環境変数ベースの設定
- [config コマンド](cli/config.md) - CLI による設定管理
