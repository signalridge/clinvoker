# Claude Code

Anthropic 的 AI 编程助手，具有深度推理能力和安全性重点。

## 概述

Claude Code 是 Anthropic 的强大 AI 编程助手。它擅长：

- 复杂的多步骤推理
- 彻底的代码分析和审查
- 安全和负责任的 AI 助手
- 深入理解上下文

## 安装

从 [Anthropic](https://claude.ai/claude-code) 安装 Claude Code：

```bash
# 验证安装
which claude
claude --version
```text

## 基本用法

```bash
# 使用 clinvk 调用 Claude
clinvk --backend claude "修复 auth.go 中的 bug"
clinvk -b claude "解释这个代码库"
```text

## 模型

| 模型 | 描述 |
|------|------|
| `claude-opus-4-5-20251101` | 最强大，适合复杂任务 |
| `claude-sonnet-4-20250514` | 平衡性能和速度 |

指定模型：

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "快速审查"
```bash

## 配置

在 `~/.clinvk/config.yaml` 中配置 Claude：

```yaml
backends:
  claude:
    # 默认模型
    model: claude-opus-4-5-20251101

    # 工具访问（all 或逗号分隔列表）
    allowed_tools: all

    # 覆盖统一审批模式
    approval_mode: default

    # 覆盖统一沙箱模式
    sandbox_mode: default

    # 启用/禁用此后端
    enabled: true

    # 自定义系统提示
    system_prompt: ""

    # 额外 CLI 参数
    extra_flags: []
```text

### 环境变量

```bash
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
```text

## 审批模式

Claude 支持不同的审批行为：

| 模式 | 描述 |
|------|------|
| `default` | 让 Claude 根据操作风险决定 |
| `auto` | 减少提示/自动批准（因后端而异） |
| `none` | 不再弹出批准提示（**危险**） |
| `always` | 尽可能总是请求批准 |

通过配置设置：

```yaml
backends:
  claude:
    approval_mode: auto
```text

或每个命令（在 tasks/chains 中）：

```json
{
  "backend": "claude",
  "prompt": "重构模块",
  "approval_mode": "auto"
}
```text

## 沙箱模式

控制 Claude 的文件系统访问：

!!! note
    `sandbox_mode` 是统一配置项，但对 `claude` 后端目前不会映射为 Claude CLI 参数，因此可能不会生效。

| 模式 | 描述 |
|------|------|
| `default` | 让 Claude 决定 |
| `read-only` | 只能读取文件 |
| `workspace` | 只能修改项目中的文件 |
| `full` | 完全文件系统访问 |

## 允许的工具

控制 Claude 可以使用哪些工具：

```yaml
backends:
  claude:
    # 所有工具
    allowed_tools: all

    # 仅特定工具
    allowed_tools: read,write,edit
```bash

## 最佳实践

!!! tip "复杂任务使用 Opus"
    Claude Opus 适合多步骤推理、代码架构和彻底审查。

!!! tip "利用会话连续性"
    Claude 擅长在对话中保持上下文。使用 `clinvk -c` 继续会话。

!!! tip "信任默认值"
    Claude 的默认审批和沙箱模式已针对安全性和实用性进行了良好调整。

## 使用场景

### 代码审查

```bash
clinvk -b claude "审查这个 PR 的安全问题和代码质量"
```text

### 复杂重构

```bash
clinvk -b claude "重构认证系统以使用 JWT token"
```text

### 架构分析

```bash
clinvk -b claude "分析这个代码库架构并建议改进"
```text

### Bug 调查

```bash
clinvk -b claude "调查 CI 管道中测试失败的原因"
```text

## 下一步

- [Codex CLI 指南](codex.md)
- [Gemini CLI 指南](gemini.md)
- [后端对比](../compare.md)
