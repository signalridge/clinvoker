---
title: 构建 AI Skills
description: 创建利用 clinvoker 编排多个 AI 后端的 Claude Code Skills，实现专业任务的复杂多代理工作流。
---

# 教程：构建 AI Skills

学习如何创建利用 clinvoker 调用其他 AI 后端的 Claude Code Skills，在 Claude Code 本身内部实现复杂的多代理工作流。

## 什么是 Claude Code Skills？

Claude Code Skills 是您可以通过 JSON 配置文件添加到 Claude Code 的自定义功能。它们允许您：

- 定义专业行为和提示词
- 授予对特定工具和命令的访问权限
- 为常见任务创建可重用工作流

### Skills 架构

```text
┌─────────────────────────────────────────────────────────┐
│                    Claude Code                           │
│  ┌─────────────────────────────────────────────────┐   │
│  │              您的自定义 Skill                    │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │   │
│  │  │  架构师     │  │  实现者     │  │  安全专家│ │   │
│  │  │  (Claude)   │  │  (Codex)    │  │ (Gemini) │ │   │
│  │  └──────┬──────┘  └──────┬──────┘  └────┬─────┘ │   │
│  │         │                │              │       │   │
│  │         └────────────────┴──────────────┘       │   │
│  │                      │                          │   │
│  │                 clinvoker                       │   │
│  └──────────────────────┼──────────────────────────┘   │
└─────────────────────────┼───────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
    ┌──────────┐   ┌──────────┐   ┌──────────┐
    │  Claude  │   │   Codex  │   │  Gemini  │
    │   CLI    │   │   CLI    │   │   CLI    │
    └──────────┘   └──────────┘   └──────────┘
```

### 为什么将 Skills 与 clinvoker 结合？

通过将 Claude Code Skills 与 clinvoker 结合，您可以：

1. **访问多个后端**：使用 Claude Code 作为编排器，同时利用 Codex 进行代码生成，利用 Gemini 进行安全分析
2. **保持上下文**：Claude Code 的会话管理在后端调用之间保留上下文
3. **统一界面**：用户与单一技能交互，而多个 AI 在后台工作
4. **专业 expertise**：自动将任务路由到最合适的后端

---

## 前置要求

在构建 AI Skills 之前，请确保您具备以下条件：

- Claude Code 已安装和配置
- clinvoker 已安装（`clinvk` 命令可用）
- 至少配置了两个后端（Claude、Codex 和/或 Gemini）
- 对 JSON 配置有基本了解

验证您的设置：

```bash
# 检查 Claude Code
claude --version

# 检查 clinvoker
clinvk version

# 检查可用后端
clinvk config show
```

---

## Claude Code Skills 架构

### Skill 结构

一个 Claude Code Skill 包含：

```json
{
  "name": "skill-name",
  "description": "此 skill 的作用",
  "prompt": "使用此 skill 时给 Claude 的指令",
  "tools": ["allowed", "tools"],
  "mention": {
    "prefix": "@mention-prefix"
  }
}
```

| 字段 | 描述 | 必需 |
|------|------|------|
| `name` | Skill 的唯一标识符 | 是 |
| `description` | 人类可读的描述 | 是 |
| `prompt` | Skill 的系统提示词 | 是 |
| `tools` | 允许的工具/命令数组 | 否 |
| `mention` | @提及的配置 | 否 |

### 工具集成

当您在 `tools` 数组中包含 `clinvk` 时，Claude Code 可以执行 clinvoker 命令：

```json
{
  "tools": ["clinvk", "jq", "git"]
}
```

这授予 skill 在您的环境中运行这些命令的权限。

---

## 步骤 1：创建基本的多后端 Skill

创建您的第一个协调多个后端的 skill：

```bash
mkdir -p .claude/skills
```

创建 `.claude/skills/ai-team.json`：

```json
{
  "name": "ai-team",
  "description": "与 AI 团队成员（Claude、Codex、Gemini）协作解决复杂问题",
  "prompt": "您正在协调一个拥有三位专家的 AI 开发团队：\n\n1. **Claude**（系统架构师）- 最擅长设计、架构和复杂推理\n2. **Codex**（实现者）- 最擅长编写、重构和优化代码\n3. **Gemini**（安全与研究）- 最擅长安全分析、文档和研究\n\n当用户要求您实现功能或解决问题时：\n\n1. **首先**，咨询 Claude 创建架构/设计计划\n2. **然后**，使用 Codex 实现解决方案\n3. **最后**，使用 Gemini 审查安全问题\n4. **综合**所有反馈形成最终建议\n\n使用 clinvk 命令调用不同的后端：\n- `clinvk -b claude \"<prompt>\"` - 用于架构和设计\n- `clinvk -b codex \"<prompt>\"` - 用于实现\n- `clinvk -b gemini \"<prompt>\"` - 用于安全和研究\n\n始终解释您正在咨询哪位团队成员以及原因。呈现最终解决方案时包括：\n- 架构决策和理由\n- 实现细节\n- 安全考虑\n- 做出的任何权衡",
  "tools": ["clinvk"]
}
```

### 此 Skill 的工作原理

1. 当您调用 `/ai-team` 时，Claude Code 加载此 skill 的提示词
2. Skill 指示 Claude 充当协调者
3. Claude 使用 `clinvk` 调用其他后端执行专业任务
4. 结果综合成全面的解决方案

---

## 步骤 2：测试 AI Team Skill

在项目目录中启动 Claude Code：

```bash
claude
```

调用 skill：

```text
/ai-team 我需要在 Go 中实现一个安全的用户认证系统
```

Claude 应该：

1. 调用 `clinvk -b claude "在 Go 中设计一个安全的认证系统架构"`
2. 调用 `clinvk -b codex "基于[架构]实现认证系统"`
3. 调用 `clinvk -b gemini "审查此认证代码的安全漏洞"`
4. 呈现结合所有输入的综合解决方案

---

## 步骤 3：创建代码审查 Skill

创建一个用于多后端代码审查的专业 skill：

创建 `.claude/skills/multi-review.json`：

```json
{
  "name": "multi-review",
  "description": "从多个 AI 视角获取全面的代码审查",
  "prompt": "您是一位代码审查协调员。当用户分享代码时：\n\n1. **创建并行审查任务**用于：\n   - **Claude**：架构、设计模式和可维护性\n   - **Codex**：实现质量、性能和优化\n   - **Gemini**：安全漏洞和最佳实践\n\n2. **使用 clinvoker 执行并行审查**：\n   ```bash\n   echo '{\"tasks\":[\n     {\"backend\":\"claude\",\"prompt\":\"审查架构...\"},\n     {\"backend\":\"codex\",\"prompt\":\"审查实现...\"},\n     {\"backend\":\"gemini\",\"prompt\":\"安全审计...\"}\n   ]}' | clinvk parallel -f -\n   ```\n\n3. **将反馈综合**成结构化报告：\n   - 执行摘要（关键发现）\n   - 按类别分类的详细反馈\n   - 优先建议\n   - 积极亮点\n\n4. **突出显示**多个后端发现的严重问题\n\n将您的响应格式化为：\n```\n## 代码审查报告\n\n### 摘要\n[2-3 句话的关键发现]\n\n### 严重问题\n- [问题] - 发现者：[后端]\n\n### 架构 (Claude)\n[反馈]\n\n### 实现 (Codex)\n[反馈]\n\n### 安全 (Gemini)\n[反馈]\n\n### 建议\n1. [优先级] [建议]\n```",
  "tools": ["clinvk", "jq"]
}
```

### 使用示例

```text
/multi-review
```

然后在提示时粘贴您的代码。Claude 将：

1. 创建并行审查任务文件
2. 执行 `clinvk parallel` 同时运行所有审查
3. 解析 JSON 结果
4. 呈现统一的审查报告

---

## 步骤 4：创建后端路由 Skill

创建一个自动将任务路由到最合适后端的 skill：

创建 `.claude/skills/backend-router.json`：

```json
{
  "name": "backend-router",
  "description": "自动将任务路由到最适合该工作的 AI 后端",
  "prompt": "您是一个智能任务路由器。分析用户的请求并将其路由到最合适的 AI 后端。\n\n## 路由规则\n\n| 任务类型 | 最佳后端 | 原因 |\n|---------|---------|------|\n| 架构、设计、复杂推理 | Claude | 深度推理，注重安全\n| 代码生成、重构、调试 | Codex | 针对编码任务优化\n| 安全分析、研究、文档 | Gemini | 知识广泛，注重安全\n| 快速问题、解释 | 默认 | 最快响应\n\n## 如何路由\n\n1. **分析**用户的请求\n2. **使用上述规则**确定最佳后端\n3. **使用 clinvoker 执行**：\n   ```bash\n   clinvk -b <backend> \"<optimized prompt>\"\n   ```\n4. **呈现**结果并说明为什么选择该后端\n\n## 示例\n\n用户：\"设计微服务架构\"\n- 路由到：Claude\n- 原因：需要架构思维和权衡分析\n\n用户：\"实现快速排序算法\"\n- 路由到：Codex\n- 原因：直接的实现任务\n\n用户：\"检查此代码的 SQL 注入\"\n- 路由到：Gemini\n- 原因：注重安全的分析\n\n始终解释您的路由决策，帮助用户理解后端优势。",
  "tools": ["clinvk"]
}
```

---

## 步骤 5：具有会话管理的高级 Skill

创建一个在多个会话中维护上下文的 skill：

创建 `.claude/skills/long-term-project.json`：

```json
{
  "name": "long-term-project",
  "description": "使用 clinvoker 的会话持久性和多个后端管理长期项目",
  "prompt": "您帮助管理使用 clinvoker 会话持久性和多个后端的长期编码项目。\n\n## 能力\n\n1. **会话管理**：\n   - 列出活动会话：`clinvk sessions list`\n   - 恢复之前的工作：`clinvk resume --last`\n   - 继续特定会话：`clinvk resume <session-id>`\n\n2. **多后端协调**：\n   - 根据任务阶段切换后端\n   - 在后端切换之间保持上下文\n   - 聚合来自多个来源的结果\n\n## 工作流程\n\n开始工作时：\n1. 检查现有会话：`clinvk sessions list`\n2. 如果找到，询问用户是否恢复\n3. 如果恢复：`clinvk resume --last`\n4. 如果是新的：继续新会话\n\n工作期间：\n- 使用 Claude 进行规划和架构决策\n- 使用 Codex 进行实现阶段\n- 使用 Gemini 进行安全审查和文档\n- 适当标记会话：`clinvk config set session.default_tags [\"project-x\"]`\n\n会话结束时：\n- 总结进度\n- 记录任何阻碍或下一步\n- 建议接下来使用哪个后端\n\n## 最佳实践\n\n- 用项目名称标记会话\n- 在每个会话结束时总结进度\n- 对重要工作使用特定会话 ID\n- 定期清理旧会话：`clinvk sessions cleanup`",
  "tools": ["clinvk"]
}
```

---

## 步骤 6：Skill 中的错误处理

### 常见错误模式

在构建调用 clinvoker 的 skills 时，处理以下场景：

#### 后端不可用

```json
{
  "prompt": "调用 clinvoker 时，处理后端的错误：\n\n如果后端不可用：\n1. 尝试替代后端：\n   - 如果 Claude 失败，尝试 Gemini 进行分析\n   - 如果 Codex 失败，尝试 Claude 进行实现\n2. 告知用户有关回退的信息\n3. 根据可用后端调整期望\n\n错误处理示例：\n```bash\n# 尝试主后端\nresult=$(clinvk -b claude \"analyze this\" 2>&1) || {\n  # 回退到 Gemini\n  echo \"Claude 不可用，使用 Gemini...\"\n  result=$(clinvk -b gemini \"analyze this\" 2>&1)\n}\n```"
}
```

#### 超时处理

```json
{
  "prompt": "优雅地处理超时：\n\n1. 为长任务设置适当的超时：\n   ```bash\n   clinvk -b claude --timeout 300 \"complex analysis\"\n   ```\n\n2. 如果发生超时：\n   - 将任务分成更小的块\n   - 使用更快的后端\n   - 询问用户是否要继续\n\n3. 对于关键任务，使用指数退避重试"
}
```

---

## 步骤 7：本地测试 Skills

### 测试每个 Skill

创建测试脚本 `test-skills.sh`：

```bash
#!/bin/bash
# AI Skills 测试脚本

echo "测试 AI Team Skill..."
claude -c "echo 'Test'" 2>/dev/null || echo "Claude Code 未运行"

echo ""
echo "测试 clinvoker 集成..."
clinvk -b claude --dry-run "test" > /dev/null && echo "clinvk OK" || echo "clinvk 失败"

echo ""
echo "测试后端可用性..."
clinvk config show | grep -E "claude|codex|gemini"

echo ""
echo "所有测试完成！"
```

### 手动测试清单

使用以下场景测试每个 skill：

| Skill | 测试用例 | 预期结果 |
|-------|---------|---------|
| ai-team | "设计一个 API" | 使用所有 3 个后端 |
| multi-review | 粘贴代码 | 执行并行审查 |
| backend-router | 各种任务 | 选择正确的后端 |
| long-term-project | 恢复会话 | 会话已恢复 |

---

## 步骤 8：集成模式

### 模式 1：顺序链

按顺序执行任务，在后端之间传递输出：

```json
{
  "prompt": "对于顺序工作流：\n\n1. 步骤 1 - 设计 (Claude)：\n   ```bash\n   design=$(clinvk -b claude \"设计缓存层\")\n   ```\n\n2. 步骤 2 - 实现 (Codex)：\n   ```bash\n   code=$(clinvk -b codex \"实现：$design\")\n   ```\n\n3. 步骤 3 - 审查 (Gemini)：\n   ```bash\n   review=$(clinvk -b gemini \"审查：$code\")\n   ```\n\n4. 呈现所有三个结果"
}
```

### 模式 2：并行聚合

并行运行任务并组合结果：

```json
{
  "prompt": "对于并行分析：\n\n```bash\necho '{\"tasks\":[\n  {\"backend\":\"claude\",\"prompt\":\"分析架构\"},\n  {\"backend\":\"codex\",\"prompt\":\"分析性能\"},\n  {\"backend\":\"gemini\",\"prompt\":\"分析安全\"}\n]}' | clinvk parallel -f - -o json | jq '.results[]'\n```\n\n聚合发现并呈现统一视图"
}
```

### 模式 3：回退链

按顺序尝试后端直到一个成功：

```json
{
  "prompt": "对于弹性执行：\n\n```bash\n# 按顺序尝试后端\nfor backend in claude gemini codex; do\n  result=$(clinvk -b $backend \"task\" 2>&1) && break\ndone\n```\n\n呈现成功的结果"
}
```

---

## 部署考虑

### 版本控制

在您的仓库中存储 skills：

```bash
# 创建 skills 目录
mkdir -p .claude/skills

# 添加 skills
git add .claude/skills/
git commit -m "为多后端工作流添加 AI team skills"
```

### 共享 Skills

与您的团队共享 skills：

```bash
# 导出 skills
tar czf ai-skills.tar.gz .claude/skills/

# 其他人导入
tar xzf ai-skills.tar.gz
```

### 组织范围的 Skills

对于组织范围的部署：

1. 创建共享 skills 仓库
2. 使用符号链接链接 skills：
   ```bash
   ln -s /shared/skills/* .claude/skills/
   ```
3. 在您的团队 wiki 中记录 skill 用法

---

## 最佳实践

### 1. 清晰的角色定义

定义每个后端最擅长的方面：

```json
{
  "prompt": "后端角色：\n- Claude：架构、推理、安全关键决策\n- Codex：实现、代码生成、调试\n- Gemini：安全、研究、文档"
}
```

### 2. 显式错误处理

始终包含回退指令：

```json
{
  "prompt": "如果后端失败：\n1. 尝试替代后端\n2. 告知用户变更\n3. 相应调整方法"
}
```

### 3. 结构化输出

请求结构化输出以便更容易解析：

```json
{
  "prompt": "将结果格式化为：\n## 后端：<name>\n### 输出\n<content>\n### 置信度\n<high/medium/low>"
}
```

### 4. 会话感知

将会话用于上下文保留：

```json
{
  "prompt": "对于多步骤任务：\n1. 检查现有会话\n2. 如果相关则恢复\n3. 适当标记新会话"
}
```

---

## 故障排除

### Skill 未出现

如果您的 skill 未出现在 Claude Code 中：

```bash
# 检查文件位置
ls -la .claude/skills/

# 验证 JSON
jq . .claude/skills/your-skill.json

# 重启 Claude Code
exit
claude
```

### 找不到 clinvoker

如果 Claude 找不到 clinvoker：

```bash
# 验证 clinvoker 在 PATH 中
which clinvoker

# 添加到 skill 工具
{
  "tools": ["clinvk", "jq"]
}

# 如果需要使用完整路径
{
  "prompt": "使用完整路径：/usr/local/bin/clinvk"
}
```

### 后端错误

处理后端特定的错误：

```json
{
  "prompt": "如果您收到 'backend not available'：\n1. 检查可用后端：clinvk config show\n2. 使用可用后端\n3. 告知用户限制"
}
```

---

## 后续步骤

- 了解[会话管理](../guides/sessions.zh.md)以获取持久工作流
- 探索[LangChain 集成](langchain-integration.zh.md)以获取编程工作流
- 查看[多后端代码审查](multi-backend-code-review.zh.md)以获取审查自动化
- 查看[架构概述](../concepts/architecture.zh.md)以了解内部原理

---

## 总结

您已经学会如何：

1. 创建利用 clinvoker 的 Claude Code Skills
2. 在 Claude Code 内部协调多个 AI 后端
3. 实现错误处理和回退策略
4. 为不同用例设计 skills（审查、路由、项目管理）
5. 有效测试和部署 skills

通过将 Claude Code Skills 与 clinvoker 结合，您可以创建强大的多代理工作流，利用每个 AI 助手的独特优势，同时保持统一的用户体验。
