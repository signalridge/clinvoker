# 基本的な使い方

日々の作業で clinvk を使うための基礎を学びます。このガイドでは、コマンド構造、グローバルフラグ、バックエンドの選び方、出力形式を詳しく解説します。

## コマンド構造の概要

clinvk は一貫したコマンド構造に従います。

```bash
clinvk [global-flags] [command] [command-flags] [arguments]
```

### コマンド種別

| 種別 | 例 | 説明 |
|------|---------|-------------|
| **直接プロンプト** | `clinvk "fix the bug"` | 既定のバックエンドでプロンプトを実行 |
| **サブコマンド** | `clinvk sessions list` | 特定のサブコマンドを実行 |
| **再開** | `clinvk resume --last` | 以前のセッションを再開 |

## グローバルフラグの説明

グローバルフラグは、実行するコマンドに関係なく clinvk の挙動に影響します。サブコマンドの前に指定できます。

### バックエンド選択（`--backend`, `-b`）

`--backend` フラグは、どの AI バックエンドでプロンプトを処理するかを決定します。

```bash
# Use specific backends
clinvk --backend claude "fix the bug in auth.go"
clinvk -b codex "implement user registration"
clinvk -b gemini "explain this algorithm"
```

**バックエンド選択の目安:**

| タスク種別 | 推奨バックエンド | 理由 |
|-----------|---------------------|--------|
| 複雑な推論 | `claude` | 文脈理解が深く、安全性重視 |
| コード生成 | `codex` | プログラミングタスク向けに最適化 |
| ドキュメント | `gemini` | 知識の幅が広く、説明が明快 |
| セキュリティレビュー | `claude` | 丁寧な分析とリスク評価 |
| 迅速な試作 | `codex` | 素早いコード生成 |

バックエンドを指定しない場合、clinvk は設定の `default_backend` を使用します（デフォルトは `claude`）。

### モデル選択（`--model`, `-m`）

選択したバックエンドのデフォルトモデルを上書きします。

```bash
clinvk --model claude-opus-4-5-20251101 "complex architecture task"
clinvk -b codex -m o3 "implement feature"
clinvk -b gemini -m gemini-2.5-flash "quick question"
```

**モデルを上書きする場面:**

- 深い推論が必要な複雑タスクには大きいモデル（Opus / o3 / Pro）
- 速さ重視の簡単タスクには小さいモデル（Sonnet / o3-mini / Flash）
- コストとレイテンシのトレードオフを考慮

### 作業ディレクトリ（`--workdir`, `-w`）

AI が操作対象とする作業ディレクトリを指定します。

```bash
clinvk --workdir /path/to/project "review the codebase"
clinvk -w ./subproject "fix tests"
```

**作業ディレクトリの挙動:**

- AI は指定したディレクトリを作業コンテキストとして受け取ります
- ファイル操作はこのディレクトリを基準に行われます
- バックエンドによってサンドボックスの扱いが異なります（[バックエンドガイド](backends/index.md) を参照）
- スクリプトでは可読性のため、絶対パスの使用を推奨します

**セキュリティ上の注意:**

```bash
# Good: Explicit, limited scope
clinvk -w /home/user/projects/myapp "analyze code"

# Risky: Full system access (depends on backend sandbox mode)
clinvk -w / "search for files"
```

### 出力形式（`--output-format`, `-o`）

出力の表示形式を指定します。実効デフォルトは設定の `output.format` に従います（組み込みデフォルトは `json`）。

#### text 形式

人が読みやすい整形済みテキスト出力です。

```bash
clinvk --output-format text "explain this code"
```

**向いている用途:** 対話的な利用、ターミナルで読む、簡単な確認

#### json 形式

プログラムから扱いやすい構造化出力です。

```bash
clinvk --output-format json "explain this code"
```

**出力構造:**

```json
{
  "output": "The code implements...",
  "backend": "claude",
  "model": "claude-opus-4-5-20251101",
  "duration_seconds": 2.5,
  "exit_code": 0
}
```

**向いている用途:** スクリプト、CI/CD パイプライン、結果の保存、後続処理

#### stream-json 形式

```bash
clinvk -o stream-json "explain this code"
```

`stream-json` はバックエンドのネイティブなストリーミング出力（NDJSON/JSONL）をそのまま通します。AI が生成している途中経過をリアルタイムで受け取れます。

**向いている用途:** 長時間タスク、リアルタイム監視、対話的ツールの構築

**形式比較:**

| 形式 | 人間が読める | 機械が読める | ストリーミング | 使いどころ |
|--------|---------------|------------------|-----------|----------|
| `text` | はい | いいえ | いいえ | 対話的な利用 |
| `json` | ある程度 | はい | いいえ | スクリプト、保存 |
| `stream-json` | ある程度 | はい | はい | リアルタイム用途 |

### 継続モード（`--continue`, `-c`）

セッション ID を指定せずに、直近のセッションを継続します。

```bash
clinvk "implement the login feature"
clinvk -c "now add password validation"
clinvk -c "add rate limiting"
```

**継続モードの仕組み:**

1. clinvk が直近の再開可能なセッションを探します
2. 新しいプロンプトを会話履歴に追加します
3. AI は過去のやり取りを含む完全な文脈を持った状態で応答します

**セッション要件:**

- 直前のセッションにバックエンドセッション ID が必要です
- `--ephemeral` で作成したセッションは継続できません
- 同じバックエンドのセッションのみ継続できます

### エフェメラルモード（`--ephemeral`）

セッションを作成せず、ステートレスに実行します。

```bash
clinvk --ephemeral "what is 2+2"
```

**エフェメラルモードが向く場面:**

| シナリオ | エフェメラルが向く理由 |
|----------|----------------|
| 単発の質問 | 履歴が不要 |
| CI/CD スクリプト | セッションが溜まるのを防ぐ |
| テスト/デバッグ | 常にクリーンな状態 |
| 公開/共有環境 | プライバシー、保持しない |
| 高頻度の自動化 | 保存コストを削減 |

**トレードオフ:**

- **Pros:** 保存しない、速い、プライバシー
- **Cons:** 履歴が残らない、再開できない

### Dry Run モード（`--dry-run`）

実行せずに、実行されるコマンドを確認します。

```bash
clinvk --dry-run "implement feature X"
```

**実行されるコマンドがそのまま表示されます:**

```yaml
Would execute: claude --model claude-opus-4-5-20251101 "implement feature X"
```

**用途:**

- 高コストな実行の前に設定を確認する
- フラグ解釈やバックエンド選択のデバッグ
- 期待される挙動のドキュメント化
- CI/CD で実 API を叩かずにテストする

### Verbose モード（`--verbose`, `-v`）

詳細ログを有効化します。

```bash
clinvk --verbose "complex task"
```

**表示される内容:**

- 設定読み込みの詳細
- バックエンド検出情報
- コマンド構築の過程
- API の呼び出しと応答（バックエンドによります）

## 終了コードリファレンス

clinvk はスクリプト向けに標準的な終了コードを使用します。

| コード | 意味 | 発生タイミング |
|------|---------|----------------|
| 0 | 成功 | コマンドが正常に完了 |
| 1 | 一般エラー | CLI エラー、バリデーション失敗、バックエンドエラー |
| バックエンドコード | 伝播 | バックエンド固有の終了コード（prompt/resume） |

**コマンド別の終了コード:**

| コマンド | 終了コード 0 | 終了コード 1 |
|---------|-------------|-------------|
| `prompt` | 成功 | バックエンドエラー |
| `parallel` | 全タスク成功 | 1 つ以上のタスクが失敗 |
| `compare` | 全バックエンド成功 | 1 つ以上のバックエンドが失敗 |
| `chain` | 全ステップ成功 | いずれかのステップが失敗 |
| `serve` | 正常終了 | サーバーエラー |

**スクリプト例:**

```bash
#!/bin/bash

clinvk "implement feature"
exit_code=$?

case $exit_code in
  0)
    echo "Success - feature implemented"
    ;;
  1)
    echo "Failed - check logs"
    exit 1
    ;;
  *)
    echo "Backend returned code: $exit_code"
    ;;
esac
```

## 環境変数

環境変数で設定を上書きできます。

```bash
# Set default backend
export CLINVK_BACKEND=codex

# Set models per backend
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3-mini
export CLINVK_GEMINI_MODEL=gemini-2.5-flash

# Run with environment settings
clinvk "prompt"  # Uses codex with o3-mini
```

**優先順位**（上ほど優先度が高い）:

1. CLI フラグ（`--backend codex`）
2. 環境変数（`CLINVK_BACKEND`）
3. 設定ファイル（`~/.clinvk/config.yaml`）
4. 組み込みのデフォルト値

## 会話を継続する

### すぐに継続する

`--continue`（または `-c`）で直近のセッションを継続できます。

```bash
clinvk "implement the login feature"
clinvk -c "now add password validation"
clinvk -c "add rate limiting"
```

### resume コマンド

より細かく制御したい場合は `resume` コマンドを使用します。

```bash
# Resume last session
clinvk resume --last

# Interactive session picker
clinvk resume --interactive

# Resume with a specific prompt
clinvk resume --last "continue from where we left off"

# Resume by ID
clinvk resume abc123 "add tests"
```

詳細は [セッション管理](sessions.md) を参照してください。

## グローバルフラグ一覧

| フラグ | 短縮 | 説明 | デフォルト |
|------|-------|-------------|---------|
| `--backend` | `-b` | 使用する AI バックエンド | `claude` |
| `--model` | `-m` | 使用するモデル | （バックエンド既定） |
| `--workdir` | `-w` | 作業ディレクトリ | （カレント） |
| `--output-format` | `-o` | 出力形式 | `json`（設定可） |
| `--continue` | `-c` | 直近セッションを継続 | `false` |
| `--dry-run` | | コマンドのみ表示 | `false` |
| `--ephemeral` | | ステートレスモード | `false` |
| `--verbose` | `-v` | 詳細ログを有効化 | `false` |
| `--config` | | 設定ファイルのパス | `~/.clinvk/config.yaml` |

## 例

### すぐにバグ修正

```bash
clinvk "there's a null pointer exception in utils.go line 45"
```

### コード生成

```bash
clinvk -b codex "generate a REST API handler for user CRUD operations"
```

### コードの説明

```bash
clinvk -b gemini "explain what the main function in cmd/server/main.go does"
```

### 継続しながらリファクタリング

```bash
clinvk "refactor the database module to use connection pooling"
clinvk -c "now add unit tests for the changes"
clinvk -c "update the documentation"
```

### CI/CD 連携

```bash
# Non-interactive mode with JSON output
clinvk --ephemeral --output-format json \
  --backend codex \
  "generate tests for the auth module"
```

### 複数ステップのワークフロー

```bash
#!/bin/bash

# Step 1: Analyze
clinvk -o json "analyze the codebase architecture" > analysis.json

# Step 2: Generate based on analysis
clinvk -c "implement the recommended changes"

# Step 3: Verify
clinvk -c "run the tests and fix any failures"
```

## よくあるパターン

### パターン 1: まず text で探索し、json で自動化する

```bash
# Interactive exploration - use text
clinvk -o text "explain this module"

# Once satisfied, switch to JSON for automation
clinvk -o json --ephemeral "generate the implementation"
```

### パターン 2: タスク種別ごとにバックエンドを使い分ける

```bash
# Architecture decisions - Claude
clinvk -b claude "design the API structure"

# Implementation - Codex
clinvk -b codex "implement the endpoints"

# Documentation - Gemini
clinvk -b gemini "write API documentation"
```

### パターン 3: 実行前に Dry Run で確認する

```bash
# Verify what will happen
clinvk --dry-run --backend codex "refactor the entire codebase"

# If satisfied, run for real
clinvk --backend codex "refactor the entire codebase"
```

## トラブルシューティング

### バックエンドが見つからない

```bash
# Check available backends
clinvk config show | grep available

# Verify CLI installation
which claude codex gemini
```

### 設定が反映されない

```bash
# Check effective configuration
clinvk config show

# Verify file exists
ls -la ~/.clinvk/config.yaml
```

### セッションが再開できない

```bash
# List available sessions
clinvk sessions list

# Check if session has backend ID
clinvk sessions show <session-id>
```

## 次のステップ

- [セッション管理](sessions.md) - セッションを効果的に扱う
- [バックエンド比較](compare.md) - 複数の視点を得る
- [設定](../reference/configuration.md) - セットアップをカスタマイズ
