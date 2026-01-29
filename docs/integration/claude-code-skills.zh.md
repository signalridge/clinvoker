# Claude Code Skills 集成

本指南解释如何将 clinvk 与 Claude Code Skills 集成，以多后端 AI 支持扩展 Claude 的能力。

## 为什么在 Skills 中使用 clinvk？

Claude Code Skills 扩展了 Claude 的能力，但有时你需要：

- **其他 AI 后端**：Gemini 擅长数据分析，Codex 擅长代码生成
- **多模型协作**：复杂任务受益于多个视角
- **标准化 API**：所有后端使用单一 HTTP 接口
- **并行处理**：并发运行多个 AI 任务

## 前提条件

1. **clinvk 服务器正在运行**：
   ```bash
   clinvk serve --port 8080 &
   ```

2. **至少安装一个后端 CLI**（`claude`、`codex` 或 `gemini`）

## 基础 Skill 示例

### 单后端调用

创建一个使用 Gemini 进行数据分析的 skill：

```markdown
<!-- ~/.claude/skills/analyze-data/SKILL.md -->
# 数据分析 Skill

通过 clinvk 使用 Gemini CLI 分析数据。

## 用法
当你需要分析结构化数据时运行此 skill。

## 脚本
```bash
#!/bin/bash
DATA="$1"

curl -s http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d "{
    \"backend\": \"gemini\",
    \"prompt\": \"分析这个数据并提供洞察: $DATA\",
    \"output_format\": \"json\"
  }" | jq -r '.result'
```
```

### 使用 Skill

在 Claude Code 中，可以这样调用 skill：

```
User: /analyze-data {"sales": [100, 150, 200], "months": ["Jan", "Feb", "Mar"]}
```

## 多模型审查 Skill

一个更强大的 skill，使用多个后端进行全面代码审查：

```markdown
<!-- ~/.claude/skills/multi-review/SKILL.md -->
# 多模型代码审查

使用 Claude（架构）、Codex（性能）和 Gemini（安全）进行全面代码审查。

## 用法
提供文件路径或代码片段进行多视角审查。

## 脚本
```bash
#!/bin/bash
CODE="$1"

# 使用所有后端并行审查
RESULT=$(curl -s http://localhost:8080/api/v1/parallel \
  -H "Content-Type: application/json" \
  -d "{
    \"tasks\": [
      {
        \"backend\": \"claude\",
        \"prompt\": \"审查此代码的架构和设计模式：\\n$CODE\"
      },
      {
        \"backend\": \"codex\",
        \"prompt\": \"审查此代码的性能问题和优化：\\n$CODE\"
      },
      {
        \"backend\": \"gemini\",
        \"prompt\": \"审查此代码的安全漏洞：\\n$CODE\"
      }
    ]
  }")

echo "## 多模型代码审查结果"
echo ""
echo "### 架构审查 (Claude)"
echo "$RESULT" | jq -r '.results[0].result // .results[0].error'
echo ""
echo "### 性能审查 (Codex)"
echo "$RESULT" | jq -r '.results[1].result // .results[1].error'
echo ""
echo "### 安全审查 (Gemini)"
echo "$RESULT" | jq -r '.results[2].result // .results[2].error'
```
```

## 链式执行 Skill

一个通过多个后端传递输出的 skill：

```markdown
<!-- ~/.claude/skills/doc-pipeline/SKILL.md -->
# 文档流水线

通过多步骤流水线生成精美文档：
1. Claude 分析代码结构
2. Codex 生成文档
3. Gemini 润色和提高可读性

## 脚本
```bash
#!/bin/bash
CODE="$1"

curl -s http://localhost:8080/api/v1/chain \
  -H "Content-Type: application/json" \
  -d "{
    \"steps\": [
      {
        \"name\": \"analyze\",
        \"backend\": \"claude\",
        \"prompt\": \"分析此代码的结构和目的。列出所有函数、类及其关系：\\n$CODE\"
      },
      {
        \"name\": \"document\",
        \"backend\": \"codex\",
        \"prompt\": \"基于此分析，以 Markdown 格式生成全面的 API 文档：\\n{{previous}}\"
      },
      {
        \"name\": \"polish\",
        \"backend\": \"gemini\",
        \"prompt\": \"提高此文档的可读性并添加有用的示例：\\n{{previous}}\"
      }
    ]
  }" | jq -r '.results[-1].result'
```
```

## 高级模式

### 错误处理

```bash
#!/bin/bash
RESPONSE=$(curl -s -w "\n%{http_code}" http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "'"$1"'"}')

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" != "200" ]; then
  echo "错误: API 返回 $HTTP_CODE"
  echo "$BODY" | jq -r '.error.message // .error // "未知错误"'
  exit 1
fi

echo "$BODY" | jq -r '.result'
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

curl -s http://localhost:8080/api/v1/prompt \
  -d "{\"backend\": \"$BACKEND\", \"prompt\": \"$PROMPT\"}" | jq -r '.result'
```

### 流式响应

对于长时间运行的任务，使用流式传输：

```bash
#!/bin/bash
curl -N http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{
    "backend": "claude",
    "prompt": "'"$1"'",
    "stream": true
  }' | while read -r line; do
    # 处理每个 SSE 事件
    if [[ $line == data:* ]]; then
      echo "${line#data: }" | jq -r '.content // empty'
    fi
  done
```

## 最佳实践

### 1. 使用 Ephemeral 模式

对于无状态的 skill 执行，使用 ephemeral 会话：

```bash
curl -s http://localhost:8080/api/v1/prompt \
  -d '{
    "backend": "claude",
    "prompt": "...",
    "session_mode": "ephemeral"
  }'
```

### 2. 选择正确的后端

| 任务类型 | 推荐后端 | 原因 |
|---------|---------|------|
| 代码审查 | Claude | 深度理解、上下文 |
| 代码生成 | Codex | 针对代码优化 |
| 数据分析 | Gemini | 强大的分析能力 |
| 文档 | 任意 | 都表现良好 |
| 安全审计 | Claude + Gemini | 不同视角 |

### 3. 处理超时

为长时间任务设置适当的超时：

```bash
curl -s --max-time 120 http://localhost:8080/api/v1/prompt \
  -d '{"backend": "claude", "prompt": "...", "timeout": 60}'
```

### 4. 格式化输出

结构化 skill 输出以便 Claude 处理：

```bash
echo "## Skill 结果"
echo ""
echo "### 摘要"
echo "$SUMMARY"
echo ""
echo "### 详情"
echo '```json'
echo "$RESULT" | jq '.'
echo '```'
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

### 服务器未运行

```bash
# 检查服务器是否运行
curl -s http://localhost:8080/health

# 如果没有，启动它
clinvk serve --port 8080 &
```

### 后端不可用

```bash
# 检查可用后端
curl -s http://localhost:8080/api/v1/backends | jq '.backends'
```

### 权限问题

确保 skill 脚本可执行：

```bash
chmod +x ~/.claude/skills/*/SKILL.md
```

## 下一步

- [LangChain/LangGraph 集成](langchain-langgraph.md) - 用于基于 Python 的 agent
- [CI/CD 集成](ci-cd.md) - 在流水线中自动化
- [REST API 参考](../reference/rest-api.md) - 完整的 API 文档
