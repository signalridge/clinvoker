# 用户指南

clinvk 完整使用指南，从安装到高级工作流。

## 快速导航

| 章节 | 描述 |
|------|------|
| [安装](installation.md) | 在您的系统上安装 clinvk |
| [快速开始](quick-start.md) | 5 分钟入门 |
| [配置指南](configuration.md) | 为你的工作流配置 clinvk |
| [基本用法](basic-usage.md) | 核心 CLI 命令 |
| [会话管理](session-management.md) | 跟踪和恢复对话 |
| [并行执行](parallel-execution.md) | 运行并发任务 |
| [链式执行](chain-execution.md) | 构建多步骤流水线 |
| [后端对比](backend-comparison.md) | 对比 AI 后端响应 |
| [HTTP 服务器](http-server.md) | REST API 服务器 |
| [后端配置](backends/index.md) | Claude、Codex、Gemini 指南 |

## 学习路径

### 入门

1. **[安装](installation.md)** - 所有平台的多种安装方法
2. **[快速开始](quick-start.md)** - 您的第一个 clinvk 命令
3. **[配置指南](configuration.md)** - 为你的工作流自定义 clinvk

### 核心用法

4. **[基本用法](basic-usage.md)** - 基本 CLI 命令和选项
5. **[会话管理](session-management.md)** - 持久化和恢复对话

### 高级功能

6. **[并行执行](parallel-execution.md)** - 并发运行多个任务
7. **[链式执行](chain-execution.md)** - 顺序多阶段流水线
8. **[后端对比](backend-comparison.md)** - 并排比较响应

### 集成

9. **[HTTP 服务器](http-server.md)** - 用于编程访问的 REST API
10. **[后端配置](backends/index.md)** - 后端特定配置

## 平台支持

clinvk 支持以下平台：

| 平台 | 架构 | 包格式 |
|------|------|--------|
| Linux | amd64, arm64 | Binary, Deb, RPM, Nix |
| macOS | amd64, arm64 | Binary, Homebrew, Nix |
| Windows | amd64 | Binary, Scoop |

## 前置条件

使用 clinvk 之前，确保您至少安装了一个 AI CLI 工具：

- **Claude Code** - 从 [Anthropic](https://claude.ai/claude-code) 安装
- **Codex CLI** - 从 [OpenAI](https://github.com/openai/codex-cli) 安装
- **Gemini CLI** - 从 [Google](https://github.com/google/gemini-cli) 安装
