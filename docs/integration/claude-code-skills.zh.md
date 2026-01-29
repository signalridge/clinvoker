# Claude Code Skills 集成

本指南介绍如何将 clinvk 与 Claude Code Skills 集成，扩展 Claude 的多后端 AI 支持能力。

## 为什么在 Skills 中使用 clinvk？

Claude Code Skills 扩展了 Claude 的能力，但有时你需要：

- **其他 AI 后端**：Gemini 擅长数据分析，Codex 擅长代码生成
- **多模型协作**：复杂任务受益于多个视角
- **并行处理**：并发运行多个 AI 任务

## 前置条件

1. **已安装 clinvk** 且在 PATH 中
2. **至少安装一个后端 CLI**（`claude`、`codex` 或 `gemini`）

## 基本 Skill 示例

### 单后端调用

创建一个使用 Gemini 进行数据分析的 skill：

```markdown
<!-- ~/.claude/skills/analyze-data/SKILL.md -->
# 数据分析 Skill

使用 clinvk 调用 Gemini CLI 分析数据。

## 用法
当需要分析结构化数据时运行此 skill。

## Script
```bash
#!/bin/bash
DATA="$1"

clinvk -b gemini -o json --ephemeral "分析此数据并提供洞察：$DATA"
```
```

## 多模型审查 Skill

使用多个后端进行全面代码审查：

```markdown
<!-- ~/.claude/skills/multi-review/SKILL.md -->
# 多模型代码审查

使用 Claude（架构）、Codex（性能）和 Gemini（安全）进行全面代码审查。

## Script
```bash
#!/bin/bash
CODE="$1"

echo "## 多模型代码审查结果"
echo ""

echo "### 架构审查 (Claude)"
clinvk -b claude --ephemeral "审查此代码的架构和设计模式：
$CODE"

echo ""
echo "### 性能审查 (Codex)"
clinvk -b codex --ephemeral "审查此代码的性能问题和优化点：
$CODE"

echo ""
echo "### 安全审查 (Gemini)"
clinvk -b gemini --ephemeral "审查此代码的安全漏洞：
$CODE"
```
```

## 并行审查 Skill

使用并行执行加速多模型审查：

```markdown
<!-- ~/.claude/skills/parallel-review/SKILL.md -->
# 并行多模型审查

使用所有后端同时进行快速代码审查。

## Script
```bash
#!/bin/bash
CODE="$1"

# 创建任务文件
cat > /tmp/review-tasks.json << EOF
{
  "tasks": [
    {"backend": "claude", "prompt": "审查架构和设计：$CODE"},
    {"backend": "codex", "prompt": "审查性能问题：$CODE"},
    {"backend": "gemini", "prompt": "审查安全漏洞：$CODE"}
  ]
}
EOF

clinvk parallel -f /tmp/review-tasks.json -o json | jq -r '
  "## 架构 (Claude)\n" + .results[0].result + "\n\n" +
  "## 性能 (Codex)\n" + .results[1].result + "\n\n" +
  "## 安全 (Gemini)\n" + .results[2].result
'
```
```

## 链式执行 Skill

通过多个后端串联输出：

```markdown
<!-- ~/.claude/skills/doc-pipeline/SKILL.md -->
# 文档生成流水线

通过多步流水线生成精炼的文档：
1. Claude 分析代码结构
2. Codex 生成文档
3. Gemini 润色提高可读性

## Script
```bash
#!/bin/bash
CODE="$1"

# 创建流水线文件
cat > /tmp/doc-pipeline.json << EOF
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "分析此代码的结构和用途。列出所有函数、类及其关系：\n$CODE"
    },
    {
      "name": "document",
      "backend": "codex",
      "prompt": "基于此分析，生成 Markdown 格式的完整 API 文档：\n{{previous}}"
    },
    {
      "name": "polish",
      "backend": "gemini",
      "prompt": "提高可读性并添加有用的示例：\n{{previous}}"
    }
  ]
}
EOF

clinvk chain -f /tmp/doc-pipeline.json -o json | jq -r '.results[-1].result'
```
```

## 高级模式

### 错误处理

```bash
#!/bin/bash
set -e

if ! OUTPUT=$(clinvk -b claude --ephemeral "$1" 2>&1); then
  echo "执行 clinvk 错误：$OUTPUT"
  exit 1
fi

echo "$OUTPUT"
```

### 条件后端选择

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

### 对比后端

```bash
#!/bin/bash
# 获取所有后端的响应并对比
clinvk compare --all-backends "$1"
```

## 最佳实践

### 1. 使用临时模式

对于无状态 skill 执行，始终使用 `--ephemeral`：

```bash
clinvk -b claude --ephemeral "你的提示"
```

### 2. 选择合适的后端

| 任务类型 | 推荐后端 | 原因 |
|---------|---------|-----|
| 代码审查 | Claude | 深度理解、上下文 |
| 代码生成 | Codex | 针对代码优化 |
| 数据分析 | Gemini | 强大的分析能力 |
| 文档编写 | 任意 | 都表现良好 |
| 安全审计 | Claude + Gemini | 不同视角 |

### 3. 使用 JSON 输出便于解析

需要处理输出时：

```bash
clinvk -b claude -o json --ephemeral "..." | jq -r '.result'
```

## Skill 目录结构

```
~/.claude/skills/
├── analyze-data/
│   └── SKILL.md
├── multi-review/
│   └── SKILL.md
├── doc-pipeline/
│   └── SKILL.md
└── shared/
    └── clinvk-helpers.sh  # 共享函数
```

## 故障排除

### 后端不可用

```bash
# 检查可用后端
clinvk config show | grep available
```

## 下一步

- [LangChain/LangGraph 集成](langchain-langgraph.md) - Python 代理
- [CI/CD 集成](ci-cd.md) - 流水线自动化
- [CLI 命令参考](../reference/commands/index.md) - 完整 CLI 文档
