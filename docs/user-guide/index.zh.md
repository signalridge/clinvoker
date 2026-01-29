# 用户指南

本指南详细介绍 clinvk 的所有功能。

## 概述

clinvk 为多个 AI 编程助手提供统一接口，具有强大的编排功能以支持复杂的工作流程。

## 核心功能

- **[基本用法](basic-usage.md)** - 学习运行提示和使用后端的基础知识
- **[会话管理](session-management.md)** - 跨会话跟踪和恢复对话
- **[并行执行](parallel-execution.md)** - 并发运行多个任务以加快工作流程
- **[链式执行](chain-execution.md)** - 通过多个后端顺序传递提示
- **[后端对比](backend-comparison.md)** - 比较不同 AI 后端的响应

## 后端指南

了解每个支持的后端：

- **[Claude Code](backends/claude.md)** - Anthropic 的 AI 编程助手
- **[Codex CLI](backends/codex.md)** - OpenAI 的代码专注 CLI 工具
- **[Gemini CLI](backends/gemini.md)** - Google 的 Gemini AI 助手

## 工作流示例

### 独立开发

```bash
# 开始开发功能
clinvk "实现用户认证"

# 继续对话
clinvk -c "添加密码哈希"

# 获取不同视角
clinvk -b gemini "审查实现"
```

### 代码审查

```bash
# 从多个后端获取审查
clinvk compare --all-backends "审查这个 PR 的问题"
```

### 复杂任务

```bash
# 运行多个独立任务
clinvk parallel --file tasks.json

# 链接多个视角
clinvk chain --file review-pipeline.json
```

## 提示

!!! tip "使用会话连续性"
    始终使用 `--continue` 或 `clinvk resume` 在较长对话中保持上下文。

!!! tip "后端选择"
    不同后端擅长不同任务。Claude 擅长复杂推理，Codex 擅长代码生成，Gemini 擅长广泛知识。

!!! tip "先试运行"
    使用 `--dry-run` 查看将执行的命令而不实际运行。
