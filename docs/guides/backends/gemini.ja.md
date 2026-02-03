# Gemini CLI

幅広い知識とマルチモーダル能力を備えた、Google の Gemini AI アシスタントです。

## 概要

Gemini CLI は、Gemini AI モデル向けの Google のコマンドラインインターフェースです。次の点が得意です。

- 幅広い知識を使った一般的な質問
- ドキュメント作成と説明
- マルチモーダルタスク（対応している場合）
- リサーチと情報収集

## インストール

[Google](https://github.com/google/gemini-cli) から Gemini CLI をインストールします。

```bash
# Verify installation
which gemini
gemini --version
```

## 基本的な使い方

```bash
# Use Gemini with clinvk
clinvk --backend gemini "explain how this algorithm works"
clinvk -b gemini "write documentation for this API"
```

## モデル

| モデル | 説明 |
|-------|-------------|
| `gemini-2.5-pro` | 最新かつ高性能なモデル |
| `gemini-2.5-flash` | より高速。速度重視で最適化 |

モデルを指定します。

```bash
clinvk -b gemini -m gemini-2.5-flash "quick explanation"
```

## 設定

`~/.clinvk/config.yaml` で Gemini を設定します。

```yaml
backends:
  gemini:
    # Default model
    model: gemini-2.5-pro

    # Enable/disable this backend
    enabled: true

    # Extra CLI flags
    extra_flags: []
```

### 環境変数

```bash
export CLINVK_GEMINI_MODEL=gemini-2.5-flash
```

## セッション管理

Gemini は `--resume` でセッションを再開します。

```bash
# Resume with clinvk
clinvk resume --last --backend gemini
clinvk resume <session-id>
```

## Sandbox モード

Gemini はサンドボックスモード（制御された実行）をサポートします。

```yaml
backends:
  gemini:
    extra_flags:
      - "--sandbox"
```

## 共通オプション

次のオプションは Gemini で利用できます。

| オプション | 説明 |
|--------|-------------|
| `model` | 使用するモデル |
| `max_tokens` | 応答トークン上限 |
| `max_turns` | 最大エージェントターン数 |

## ベストプラクティス

!!! tip "Use for Explanations"
    Gemini の幅広い知識は、概念の説明や背景の提示に適しています。

!!! tip "Leverage for Documentation"
    分かりやすい説明を活かして、ドキュメントの作成/改善に Gemini を活用できます。

!!! tip "Research Tasks"
    情報収集やリサーチ寄りの問い合わせにも Gemini は向いています。

## ユースケース

### ドキュメント

```bash
clinvk -b gemini "write comprehensive documentation for this module"
```

### 説明

```bash
clinvk -b gemini "explain the architecture of this microservice"
```

### リサーチ

```bash
clinvk -b gemini "what are the best practices for implementing rate limiting"
```

### コードレビュー

```bash
clinvk -b gemini "review this code and explain potential issues"
```

## 他バックエンドとの比較

| 観点 | Gemini | Claude | Codex |
|--------|--------|--------|-------|
| 知識の幅 | 非常に広い | 広い | 広い |
| コード生成 | 良い | 非常に良い | 非常に良い |
| 説明 | 非常に良い | 非常に良い | 良い |
| 速度 | 速い | 中程度 | 速い |

## ワークフロー例

Gemini をリサーチとドキュメントに使う例です。

```json
{
  "steps": [
    {
      "name": "research",
      "backend": "gemini",
      "prompt": "research best practices for authentication in Go"
    },
    {
      "name": "implement",
      "backend": "claude",
      "prompt": "implement authentication based on: {{previous}}"
    },
    {
      "name": "document",
      "backend": "gemini",
      "prompt": "write documentation for: {{previous}}"
    }
  ]
}
```

## トラブルシューティング

### Backend Not Available

```bash
# Check if Gemini is installed
which gemini

# Check clinvk detection
clinvk config show | grep gemini
```

### Authentication

Gemini CLI 用に有効な Google Cloud の認証情報が設定されていることを確認してください。

## 次のステップ

- [Claude Code ガイド](claude.md)
- [Codex CLI ガイド](codex.md)
- [バックエンド比較](../compare.md)
