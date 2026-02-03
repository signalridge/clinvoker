# バックエンド

clinvk は複数の AI CLI バックエンドをサポートしており、それぞれに特徴と得意分野があります。

## 対応バックエンド

- **[Claude Code](claude.md)** - 深い推論力を備えた Anthropic の AI コーディングアシスタント
- **[Codex CLI](codex.md)** - コード生成に最適化された OpenAI の CLI ツール
- **[Gemini CLI](gemini.md)** - 幅広い知識と能力を持つ Google の Gemini

## バックエンド比較

| 機能 | Claude Code | Codex CLI | Gemini CLI |
|---------|-------------|-----------|------------|
| バイナリ | `claude` | `codex` | `gemini` |
| デフォルトモデル | バックエンド既定（または設定で上書き） | バックエンド既定（または設定で上書き） | バックエンド既定（または設定で上書き） |
| セッション再開 | `--resume` | `codex exec resume` | `--resume` |
| 得意分野 | 複雑な推論、安全性 | コード生成 | 幅広い知識 |

## バックエンド検出

clinvk は PATH 上のバイナリを確認して、利用可能なバックエンドを自動検出します。

```bash
clinvk config show
```

出力にはバックエンドの利用可否が含まれます。

```text
Default Backend: claude

Backends:
  claude:
    model: claude-opus-4-5-20251101
  codex:
    model: o3
  gemini:
    model: gemini-2.5-pro

Available backends:
  claude: available
  codex: not installed
  gemini: available
```

## バックエンドの選択

### CLI で指定する

```bash
clinvk --backend claude "プロンプト"
clinvk -b codex "プロンプト"
clinvk -b gemini "プロンプト"
```

### 設定で指定する

`~/.clinvk/config.yaml` で既定のバックエンドを設定します。

```yaml
default_backend: claude
```

### 環境変数で指定する

```bash
export CLINVK_BACKEND=codex
clinvk "プロンプト"  # codex を使用
```

## バックエンド固有のオプション

各バックエンドは共通オプション（unified options）に加えて、固有のフラグもサポートします。

### 共通オプション

すべてのバックエンドで共通して使えます。

| オプション | 説明 |
|--------|-------------|
| `model` | 使用するモデル |
| `approval_mode` | 承認（approval）の扱い |
| `sandbox_mode` | ファイルアクセス権限 |
| `max_turns` | 最大エージェントターン数 |
| `max_tokens` | 応答トークン数の上限 |

### バックエンド固有フラグ

追加フラグは設定の `extra_flags` で渡せます。

```yaml
backends:
  claude:
    extra_flags: ["--add-dir", "./docs"]
  codex:
    extra_flags: ["--quiet"]
  gemini:
    extra_flags: ["--sandbox"]
```

## バックエンドの選び方

### Claude Code を選ぶ場面

- 複雑で多段のタスクに取り組むとき
- 丁寧なコードレビューや分析が必要なとき
- 安全性と正確性が最重要なとき

### Codex CLI を選ぶ場面

- ボイラープレートコードを生成したいとき
- テストを書きたいとき
- 手早くコード変換したいとき

### Gemini CLI を選ぶ場面

- 幅広い知識のコンテキストが必要なとき
- ドキュメント作成に取り組むとき
- 一般的な説明がほしいとき

## ヒント

!!! tip "Try Multiple Backends"
    `clinvk compare --all-backends` を使うと、同じ問題に対して各バックエンドがどうアプローチするかを比較できます。

!!! tip "Match Backend to Task"
    バックエンドごとに得意なタスクが異なります。試しながらワークフローに合うものを見つけてください。

!!! tip "Configure Defaults"
    設定ファイルでバックエンドごとのモデルやオプションを指定すると、より自分向けに調整できます。

## 次のステップ

- [Claude Code ガイド](claude.md)
- [Codex CLI ガイド](codex.md)
- [Gemini CLI ガイド](gemini.md)
