# Claude Code

深い推論能力と安全性重視を特徴とする、Anthropic の AI コーディングアシスタントです。

## 概要

Claude Code は Anthropic の強力な AI コーディングアシスタントです。次の点が得意です。

- 複雑で多段の推論
- 丁寧なコード分析とレビュー
- 安全で責任ある AI 支援
- 深いコンテキスト理解

## インストール

[Anthropic](https://claude.ai/claude-code) から Claude Code をインストールします。

```bash
# Verify installation
which claude
claude --version
```

## 基本的な使い方

```bash
# Use Claude with clinvk
clinvk --backend claude "fix the bug in auth.go"
clinvk -b claude "explain this codebase"
```

## モデル

| モデル | 説明 |
|-------|-------------|
| `claude-opus-4-5-20251101` | 最も高性能。複雑なタスクに最適 |
| `claude-sonnet-4-20250514` | 性能と速度のバランスが良い |

モデルを指定します。

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "quick review"
```

## 設定

`~/.clinvk/config.yaml` で Claude を設定します。

```yaml
backends:
  claude:
    # Default model
    model: claude-opus-4-5-20251101

    # Tool access (all, or comma-separated list)
    allowed_tools: all

    # Override unified approval mode
    approval_mode: default

    # Override unified sandbox mode
    sandbox_mode: default

    # Enable/disable this backend
    enabled: true

    # Custom system prompt
    system_prompt: ""

    # Extra CLI flags
    extra_flags: []
```

### 環境変数

```bash
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
```

## Approval モード

Claude は複数の approval 挙動をサポートします。

| モード | 説明 |
|------|-------------|
| `default` | 操作のリスクに応じて Claude に判断させる |
| `auto` | 確認プロンプトを減らす / 編集を自動承認（バックエンド依存） |
| `none` | 承認プロンプトを一切出さない（**危険**） |
| `always` | 常に承認を求める（対応している場合） |

Set via config:

```yaml
backends:
  claude:
    approval_mode: auto
```

Or per-command (in tasks/chains):

```json
{
  "backend": "claude",
  "prompt": "refactor the module",
  "approval_mode": "auto"
}
```

## Sandbox モード

Claude のファイルシステムアクセスを制御します。

!!! note
    `sandbox_mode` は共通（unified）の設定です。`claude` バックエンドでは、clinvk は現状 `sandbox_mode` を Claude CLI のフラグへマッピングしていないため、効果がない場合があります。

| モード | 説明 |
|------|-------------|
| `default` | Claude に判断させる |
| `read-only` | 読み取りのみ |
| `workspace` | プロジェクト内のファイルを変更可能 |
| `full` | ファイルシステム全体へアクセス可能 |

## Allowed Tools

Claude が利用できるツールを制御します。

```yaml
backends:
  claude:
    # All tools
    allowed_tools: all

    # Specific tools only
    allowed_tools: read,write,edit
```

## セッション再開

Claude Code はセッションを保存し、再開に対応しています。

```bash
# Resume with clinvk
clinvk resume --last --backend claude
clinvk resume <session-id>
```

内部的には Claude の `--resume` フラグを使用します。

## 追加フラグ

Claude CLI に追加フラグを渡します。

```yaml
backends:
  claude:
    extra_flags:
      - "--add-dir"
      - "./docs"
```

よく使うフラグ:

| フラグ | 説明 |
|------|-------------|
| `--add-dir <path>` | 追加ディレクトリをコンテキストへ追加 |
| `--verbose` | 詳細出力を有効化 |

## ベストプラクティス

!!! tip "Use Opus for Complex Tasks"
    Claude Opus は多段推論、コードアーキテクチャ、丁寧なレビューに最適です。

!!! tip "Leverage Session Continuity"
    Claude は会話をまたいだコンテキスト維持が得意です。`clinvk -c` を使ってセッションを継続してください。

!!! tip "Trust the Defaults"
    Claude のデフォルトの approval / sandbox 設定は、安全性と有用性のバランスが取れています。

## ユースケース

### コードレビュー

```bash
clinvk -b claude "review this PR for security issues and code quality"
```

### 複雑なリファクタリング

```bash
clinvk -b claude "refactor the authentication system to use JWT tokens"
```

### アーキテクチャ分析

```bash
clinvk -b claude "analyze this codebase architecture and suggest improvements"
```

### バグ調査

```bash
clinvk -b claude "investigate why the tests are failing in the CI pipeline"
```

## トラブルシューティング

### Backend Not Available

```bash
# Check if Claude is installed
which claude

# Check clinvk detection
clinvk config show | grep claude
```

### Rate Limits

レート制限に当たる場合は、次を検討してください。

- 別のモデルを使う
- リクエスト間隔を空ける
- 比較は逐次モードで実行する

## 次のステップ

- [Codex CLI ガイド](codex.md)
- [Gemini CLI ガイド](gemini.md)
- [バックエンド比較](../compare.md)
