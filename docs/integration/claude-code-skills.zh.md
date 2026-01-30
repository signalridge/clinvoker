# Claude Code Skills

在 Claude Code Skills 中调用 clinvk，以扩展能力到其它后端。

## 前置条件

- 已安装 `clinvk`
- 至少安装一个后端 CLI

## 最小示例

```bash
# ~/.claude/skills/analyze-data/command.sh
#!/bin/bash
DATA="$1"

clinvk -b gemini --ephemeral -o text "分析数据: $DATA"
```

## 多模型评审

```bash
# ~/.claude/skills/multi-review/command.sh
#!/bin/bash
CODE="$1"

echo "### 架构（Claude）"
clinvk -b claude --ephemeral "评审架构: $CODE"

echo "### 性能（Codex）"
clinvk -b codex --ephemeral "评审性能: $CODE"

echo "### 安全（Gemini）"
clinvk -b gemini --ephemeral "评审安全: $CODE"
```

## 并行评审

```bash
# ~/.claude/skills/parallel-review/command.sh
#!/bin/bash
CODE="$1"

cat > /tmp/review-tasks.json << JSON
{ "tasks": [
  {"backend":"claude","prompt":"评审架构: $CODE"},
  {"backend":"codex","prompt":"评审性能: $CODE"},
  {"backend":"gemini","prompt":"评审安全: $CODE"}
]}
JSON

clinvk parallel -f /tmp/review-tasks.json --json
```

## 说明

- Skills 场景建议使用 `--ephemeral`。
- 需要结构化输出时使用 `--output-format json` 或 `--json`。
