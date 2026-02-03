# clinvk compare

複数バックエンドの応答を比較します。

## 用法

```bash
clinvk compare <prompt> [flags]
```

## 説明

同じプロンプトを複数バックエンドへ送信し、応答を比較します。CLI の compare は常にステートレス（セッションは永続化されません）で実行されます。

## フラグ

| フラグ | 型 | デフォルト | 説明 |
|------|------|---------|-------------|
| `--backends` | string | | カンマ区切りのバックエンド一覧 |
| `--all-backends` | bool | `false` | 登録済みのバックエンドをすべて比較（未インストールの CLI はスキップ） |
| `--sequential` | bool | `false` | 逐次実行 |
| `--json` | bool | `false` | JSON 出力 |

## 例

### 特定のバックエンドを比較

```bash
clinvk compare --backends claude,codex "explain this code"
```

### すべてのバックエンドを比較

```bash
clinvk compare --all-backends "what does this function do"
```

### 逐次実行

```bash
clinvk compare --all-backends --sequential "review this PR"
```

### JSON 出力

```bash
clinvk compare --all-backends --json "analyze performance"
```

## 出力

### テキスト出力

```text
Comparing 3 backends: claude, codex, gemini
Prompt: explain this algorithm
================================================================================
[claude] This algorithm implements a binary search...
[codex] The algorithm performs a binary search...
[gemini] This is a classic binary search implementation...

================================================================================
COMPARISON SUMMARY
================================================================================
BACKEND      STATUS     DURATION     MODEL
--------------------------------------------------------------------------------
claude       OK         2.50s        claude-opus-4-5-20251101
codex        OK         3.20s        o3
gemini       OK         2.80s        gemini-2.5-pro
--------------------------------------------------------------------------------
Total time: 3.20s
```

### JSON 出力

```json
{
  "prompt": "explain this algorithm",
  "backends": ["claude", "codex", "gemini"],
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "output": "This algorithm implements a binary search...",
      "duration_seconds": 2.5,
      "exit_code": 0
    }
  ],
  "total_duration_seconds": 3.2
}
```

## 実行モード

### 並列（デフォルト）

すべてのバックエンドを同時に実行します。

```bash
clinvk compare --all-backends "prompt"
```

### 逐次

バックエンドを 1 つずつ順に実行します。

```bash
clinvk compare --all-backends --sequential "prompt"
```

## エラーハンドリング

利用できないバックエンドは警告付きでスキップされます。選択したバックエンドのいずれかが実行中に失敗した場合、このコマンドは非 0 の終了ステータスで終了します。

## 終了コード

| コード | 説明 |
|------|-------------|
| 0 | 選択したバックエンドがすべて成功 |
| 1 | いずれかのバックエンドが失敗、または利用可能なバックエンドがない |

## 関連項目

- [parallel](parallel.md) - 異なるプロンプトを並列実行
- [chain](chain.md) - 逐次パイプライン
