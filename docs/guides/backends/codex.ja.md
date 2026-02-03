# Codex CLI

コード生成やプログラミング支援に最適化された、OpenAI のコード特化 CLI ツールです。

## 概要

Codex CLI は、コード生成とプログラミング支援に焦点を当てた OpenAI のコマンドラインツールです。次の点が得意です。

- 迅速なコード生成
- テストやボイラープレートの作成
- コード変換
- 手早い実装

## インストール

[OpenAI](https://github.com/openai/codex-cli) から Codex CLI をインストールします。

```bash
# Verify installation
which codex
codex --version
```

## 基本的な使い方

```bash
# Use Codex with clinvk
clinvk --backend codex "implement a REST API handler"
clinvk -b codex "generate unit tests for user.go"
```

## モデル

| モデル | 説明 |
|-------|-------------|
| `o3` | 最新かつ高性能なモデル |
| `o3-mini` | より高速で軽量なモデル |

モデルを指定します。

```bash
clinvk -b codex -m o3-mini "quick code generation"
```

## 設定

`~/.clinvk/config.yaml` で Codex を設定します。

```yaml
backends:
  codex:
    # Default model
    model: o3

    # Enable/disable this backend
    enabled: true

    # Extra CLI flags
    extra_flags: []
```

### 環境変数

```bash
export CLINVK_CODEX_MODEL=o3-mini
```

## セッション管理

Codex は `codex exec resume` サブコマンドでセッションを再開します（`clinvk` が自動的に処理します）。

```bash
# Resume with clinvk
clinvk resume --last --backend codex
clinvk resume <session-id>
```

## 共通オプション

次のオプションは Codex で利用できます。

| オプション | 説明 |
|--------|-------------|
| `model` | 使用するモデル |
| `max_tokens` | 応答トークン上限 |
| `max_turns` | 最大エージェントターン数 |

## 追加フラグ

Codex に追加フラグを渡します。

```yaml
backends:
  codex:
    extra_flags:
      - "--quiet"
```

よく使うフラグ:

| フラグ | 説明 |
|------|-------------|
| `--quiet` | 出力を簡素化 |

## ベストプラクティス

!!! tip "Use for Code Generation"
    Codex は素早いコード生成に最適化されています。ボイラープレートや反復的な作業に向いています。

!!! tip "Combine with Other Backends"
    Codex でコードを生成し、Claude でレビューする、といった組み合わせが有効です。`chain` コマンドを活用してください。

!!! tip "Batch Similar Tasks"
    類似タスクは並列実行して効率化しましょう。

## ユースケース

### ボイラープレート生成

```bash
clinvk -b codex "create a CRUD API for the User model"
```

### テスト作成

```bash
clinvk -b codex "generate comprehensive unit tests for the auth module"
```

### コード変換

```bash
clinvk -b codex "convert this callback-based code to async/await"
```

### 手早い実装

```bash
clinvk -b codex "implement a binary search function"
```

## Claude との比較

| 観点 | Codex | Claude |
|--------|-------|--------|
| 速度 | 速い | より丁寧 |
| 得意分野 | コード生成 | 複雑な推論 |
| コンテキスト理解 | 良い | 非常に良い |
| 安全性重視 | 標準 | 高い |

## ワークフロー例

Codex と Claude を組み合わせて使います。

```json
{
  "steps": [
    {
      "name": "generate",
      "backend": "codex",
      "prompt": "implement user authentication"
    },
    {
      "name": "review",
      "backend": "claude",
      "prompt": "review this code for security: {{previous}}"
    }
  ]
}
```

## トラブルシューティング

### Backend Not Available

```bash
# Check if Codex is installed
which codex

# Check clinvk detection
clinvk config show | grep codex
```

### Model Errors

モデルが利用できない場合:

```bash
# List available models
codex models list

# Update config to use available model
clinvk config set backends.codex.model o3
```

## 次のステップ

- [Claude Code ガイド](claude.md)
- [Gemini CLI ガイド](gemini.md)
- [バックエンド比較](../compare.md)
