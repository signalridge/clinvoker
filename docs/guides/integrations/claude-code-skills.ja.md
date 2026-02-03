# Claude Code Skills 連携

このガイドでは、clinvk を Claude Code Skills と連携させて、Claude の能力をマルチバックエンド対応で拡張する方法を説明します。

## Skills で clinvk を使う理由

Claude Code Skills は Claude の能力を拡張できますが、次のようなニーズが出ることがあります。

- **他の AI バックエンド**: Gemini はデータ分析、Codex はコード生成が得意
- **マルチモデル協調**: 複雑なタスクは複数の視点が有効
- **並列処理**: 複数の AI タスクを同時に実行したい

## 前提条件

1. **clinvk がインストール済み** で PATH が通っていること
2. **少なくとも 1 つのバックエンド CLI がインストール済み**（`claude`, `codex`, `gemini` のいずれか）

## 基本的な Skill の例

### 単一バックエンド呼び出し

Gemini を使ってデータ分析する Skill を作成します。

<!-- markdownlint-disable MD018 MD022 MD040 -->

```markdown
<!-- ~/.claude/skills/analyze-data/SKILL.md -->
# Data Analysis Skill

Analyzes data using Gemini CLI via clinvk.

## Usage
Run this skill when you need to analyze structured data.

## Script
```
#!/bin/bash
DATA="$1"

clinvk -b gemini -o json --ephemeral "Analyze this data and provide insights: $DATA"
```

```

<!-- markdownlint-enable MD018 MD022 MD040 -->

### Using the Skill

In Claude Code, the skill can be invoked:

```yaml

User: /analyze-data {"sales": [100, 150, 200], "months": ["Jan", "Feb", "Mar"]}

```

## Multi-Model Review Skill

複数バックエンドを使って、より包括的なコードレビューを行う Skill の例です。

<!-- markdownlint-disable MD018 MD022 MD040 -->

```markdown
<!-- ~/.claude/skills/multi-review/SKILL.md -->
# Multi-Model Code Review

Performs comprehensive code review using Claude (architecture),
Codex (performance), and Gemini (security).

## Usage
Provide a file path or code snippet for multi-perspective review.

## Script
```
#!/bin/bash
CODE="$1"

echo "## Multi-Model Code Review Results"
echo ""

echo "### Architecture Review (Claude)"
clinvk -b claude --ephemeral "Review this code for architecture and design patterns:
$CODE"

echo ""
echo "### Performance Review (Codex)"
clinvk -b codex --ephemeral "Review this code for performance issues and optimizations:
$CODE"

echo ""
echo "### Security Review (Gemini)"
clinvk -b gemini --ephemeral "Review this code for security vulnerabilities:
$CODE"
```

```

<!-- markdownlint-enable MD018 MD022 MD040 -->

## Parallel Review Skill

並列実行を使って、マルチモデルレビューを高速化する例です。

<!-- markdownlint-disable MD018 MD022 MD040 -->

```markdown
<!-- ~/.claude/skills/parallel-review/SKILL.md -->
# Parallel Multi-Model Review

Fast parallel code review using all backends simultaneously.

## Script
```
#!/bin/bash
CODE="$1"

# Create tasks file
cat > /tmp/review-tasks.json << EOF
{
  "tasks": [
    {"backend": "claude", "prompt": "Review for architecture and design: $CODE"},
    {"backend": "codex", "prompt": "Review for performance issues: $CODE"},
    {"backend": "gemini", "prompt": "Review for security vulnerabilities: $CODE"}
  ]
}
EOF

clinvk parallel -f /tmp/review-tasks.json --json | jq -r '
  "## Architecture (Claude)\n" + .results[0].output + "\n\n" +
  "## Performance (Codex)\n" + .results[1].output + "\n\n" +
  "## Security (Gemini)\n" + .results[2].output
'
```

```

<!-- markdownlint-enable MD018 MD022 MD040 -->

## Chain Execution Skill

複数バックエンドへ出力を渡しながら処理するパイプライン例です。

<!-- markdownlint-disable MD018 MD022 MD040 -->

```markdown
<!-- ~/.claude/skills/doc-pipeline/SKILL.md -->
# Documentation Pipeline

Generates polished documentation through a multi-step pipeline:
1. Claude analyzes code structure
2. Codex generates documentation
3. Gemini polishes and improves readability

## Script
```
#!/bin/bash
CODE="$1"

# Create pipeline file
cat > /tmp/doc-pipeline.json << EOF
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "Analyze the structure and purpose of this code. List all functions, classes, and their relationships:\n$CODE"
    },
    {
      "name": "document",
      "backend": "codex",
      "prompt": "Based on this analysis, generate comprehensive API documentation in Markdown format:\n{{previous}}"
    },
    {
      "name": "polish",
      "backend": "gemini",
      "prompt": "Improve the readability and add helpful examples to this documentation:\n{{previous}}"
    }
  ]
}
EOF

clinvk chain -f /tmp/doc-pipeline.json --json | jq -r '.results[-1].output'
```

```

<!-- markdownlint-enable MD018 MD022 MD040 -->

## Advanced Patterns

### エラーハンドリング

```bash
#!/bin/bash
set -e

if ! OUTPUT=$(clinvk -b claude --ephemeral "$1" 2>&1); then
  echo "Error executing clinvk: $OUTPUT"
  exit 1
fi

echo "$OUTPUT"
```

### 条件によるバックエンド選択

```bash
#!/bin/bash
TASK_TYPE="$1"
PROMPT="$2"

case "$TASK_TYPE" in
  "analyze")
    BACKEND="claude"
    ;;
  "generate")
    BACKEND="codex"
    ;;
  "research")
    BACKEND="gemini"
    ;;
  *)
    BACKEND="claude"
    ;;
esac

clinvk -b "$BACKEND" --ephemeral "$PROMPT"
```

### バックエンドを比較する

```bash
#!/bin/bash
# Get responses from all backends and compare
clinvk compare --all-backends "$1"
```

## Best Practices

### 1. ステートレスモードを使う

Skill 実行をステートレスにするには、常に `--ephemeral` を使うことを推奨します。

```bash
clinvk -b claude --ephemeral "your prompt"
```

### 2. 適切なバックエンドを選ぶ

| タスク種別 | 推奨バックエンド | 理由 |
|-----------|--------------------|----|
| コードレビュー | Claude | 深い理解と文脈把握 |
| コード生成 | Codex | コード向けに最適化 |
| データ分析 | Gemini | 分析能力が強い |
| ドキュメント | どれでも | いずれも良好 |
| セキュリティ監査 | Claude + Gemini | 異なる視点 |

### 3. パースには JSON 出力を使う

出力を処理する必要がある場合:

```bash
clinvk -b claude -o json --ephemeral "..." | jq -r '.content'
```

### 4. Claude が扱いやすい形式で出力する

Claude が処理しやすいよう、Skill の出力を構造化します。

```bash
echo "## Skill Results"
echo ""
echo "### Summary"
clinvk -b gemini --ephemeral "Summarize: $INPUT"
```

## Skill Directory Structure

```text
~/.claude/skills/
├── analyze-data/
│   └── SKILL.md
├── multi-review/
│   └── SKILL.md
├── doc-pipeline/
│   └── SKILL.md
└── shared/
    └── clinvk-helpers.sh  # Shared functions
```

## Troubleshooting

### Backend Not Available

```bash
# Check available backends
clinvk config show | grep available
```

### Check Version

```bash
clinvk version
```

## Next Steps

- [LangChain/LangGraph 連携](langchain-langgraph.md) - Python ベースのエージェント向け
- [CI/CD](ci-cd/index.md) - パイプラインで自動化
- [CLI コマンドリファレンス](../../reference/cli/index.md) - CLI ドキュメント一式
