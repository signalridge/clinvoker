# バックエンド比較

同じ問題に対して複数の AI バックエンドの応答を並べて比較し、異なる視点を得られます。

## 概要

`compare` コマンドは同じプロンプトを複数バックエンドへ同時に投げ、応答をまとめて表示します。次の用途に便利です。

- 多様な視点を得る
- バックエンドの強みを評価する
- 根拠のある意思決定をする
- AI モデルごとのアプローチの違いを学ぶ

## 基本的な使い方

### すべてのバックエンドを比較

有効なバックエンドすべてに対して実行します。

```bash
clinvk compare --all-backends "explain this algorithm"
```

### 特定のバックエンドを比較

比較するバックエンドを指定します。

```bash
clinvk compare --backends claude,codex "what does this code do"
clinvk compare --backends claude,gemini "review this PR"
```

## 実行モード

### 並列（デフォルト）

すべてのバックエンドを同時に実行します。

```bash
clinvk compare --all-backends "explain this code"
```

### 逐次

バックエンドを 1 つずつ順に実行します。

```bash
clinvk compare --all-backends --sequential "review this implementation"
```

逐次モードは次の場面で有効です。

- レート制限を避けたい
- システムリソースが限られている
- 応答を 1 つずつ見ながら進めたい

## 出力形式

### テキスト出力（デフォルト）

各バックエンドの応答が分かりやすく区切って表示されます。

```yaml
Prompt: explain this algorithm

━━━ claude ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: claude-opus-4-5-20251101
Duration: 2.5s

This algorithm implements a binary search...

━━━ codex ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: o3
Duration: 3.2s

The algorithm performs a binary search...

━━━ gemini ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: gemini-2.5-pro
Duration: 2.8s

This is a classic binary search implementation...
```

### JSON 出力

プログラムから扱える構造化データを出力します。

```bash
clinvk compare --all-backends --json "explain this code"
```

出力例:

```json
{
  "prompt": "explain this code",
  "backends": ["claude", "codex", "gemini"],
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "output": "This algorithm implements a binary search...",
      "duration_seconds": 2.5,
      "exit_code": 0
    },
    {
      "backend": "codex",
      "model": "o3",
      "output": "The algorithm performs a binary search...",
      "duration_seconds": 3.2,
      "exit_code": 0
    },
    {
      "backend": "gemini",
      "model": "gemini-2.5-pro",
      "output": "This is a classic binary search implementation...",
      "duration_seconds": 2.8,
      "exit_code": 0
    }
  ],
  "total_duration_seconds": 3.2
}
```

## ユースケース

### コードレビュー

コード品質について複数の観点を得ます。

```bash
clinvk compare --all-backends "review this code for bugs and improvements"
```

### アーキテクチャ判断

設計上の選択について推奨案を比較します。

```bash
clinvk compare --backends claude,gemini "what's the best way to implement caching here?"
```

### 学習

概念を各 AI がどう説明するかを比較します。

```bash
clinvk compare --all-backends "explain how async/await works in JavaScript"
```

### 妥当性確認

重要な判断をクロスチェックします。

```bash
clinvk compare --all-backends "is this implementation secure?"
```

## 失敗時の扱い

いずれかのバックエンドが失敗しても、比較は残りのバックエンドで継続します。

```yaml
Comparing 3 backends: claude, codex, gemini
Prompt: explain this code
============================================================
[claude] Response content...
[codex] Error: Backend unavailable
[gemini] Response content...

============================================================
COMPARISON SUMMARY
============================================================
BACKEND      STATUS     DURATION     MODEL
------------------------------------------------------------
claude       OK         2.50s        claude-opus-4-5-20251101
codex        FAILED     0.50s        o3
             Error: Backend unavailable
gemini       OK         2.80s        gemini-2.5-pro
------------------------------------------------------------
Total time: 2.80s
```

## コマンドオプション

| フラグ | 説明 | デフォルト |
|------|-------------|---------|
| `--backends` | カンマ区切りのバックエンド一覧 | - |
| `--all-backends` | 登録済みバックエンドをすべて比較（利用不可はスキップ） | `false` |
| `--sequential` | 逐次実行 | `false` |
| `--json` | JSON 出力 | `false` |

## 設定

`backends.<name>.enabled` は設定ファイルに保存されますが、現状 `compare` では強制されません。

## ヒント

!!! tip "Use for Important Decisions"
    重要なコード変更やアーキテクチャ判断では、複数バックエンドの比較により見落としがちな問題が見つかることがあります。

!!! tip "Note the Differences"
    バックエンド同士が一致する点（高い確度）と、食い違う点（要調査）に注目してください。

!!! tip "Consider Response Time"
    JSON 出力には所要時間が含まれるため、バックエンド性能のベンチマークに役立ちます。

!!! tip "Combine with Other Features"
    比較結果を踏まえて、次の作業で使うバックエンドを選べます。

## 他コマンドとの違い

| コマンド | ユースケース |
|---------|----------|
| `compare` | 同一プロンプトを複数バックエンドで比較 |
| `parallel` | 異なるプロンプトを並列実行 |
| `chain` | 逐次パイプライン（出力を次ステップへ渡す） |

## 次のステップ

- [並列実行](parallel.md) - 独立タスクを同時に実行
- [チェーン実行](chains.md) - マルチバックエンドの逐次パイプライン
- [バックエンドガイド](backends/index.md) - 各バックエンドの詳細
