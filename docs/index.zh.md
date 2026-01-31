---
title: clinvoker - 统一的 AI CLI 编排工具
description: 将 AI CLI 工具转化为可编程基础设施，支持 SDK 兼容、会话持久化和多后端编排。
---

# clinvoker

[![GitHub](https://img.shields.io/badge/GitHub-signalridge%2Fclinvoker-blue?logo=github)](https://github.com/signalridge/clinvoker)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](concepts/license.md)

## 什么是 clinvoker？

clinvoker 是一个统一的 AI CLI 包装器，可将多个 AI CLI 后端转化为可编程基础设施。它提供了一个单一界面来编排 Claude Code、Codex CLI 和 Gemini CLI，同时保持与 OpenAI 和 Anthropic SDK 的完全兼容性。无论您是构建 AI 驱动的自动化、与现有应用程序集成，还是管理复杂的多模型工作流，clinvoker 都能弥合交互式 CLI 工具与生产就绪 API 之间的差距。

- **使用熟悉的 SDK 与任何 CLI 后端** - 与 OpenAI 和 Anthropic SDK 的即插即用兼容性意味着现有应用程序无需修改即可工作
- **编排多后端工作流** - 将任务路由到最合适的模型，跨后端运行比较，或将多个 AI 步骤链接到统一管道中
- **维护持久会话** - 跨进程文件锁定确保会话状态在 CLI 调用和服务器重启之间持久保存
- **部署为 HTTP API 服务器** - 将任何 CLI 后端转换为具有内置速率限制、身份验证和指标的 REST API

## 架构

clinvoker 遵循分层架构，将 CLI 关注点与 HTTP API 功能分离，同时共享核心组件。设计强调模块化、可测试性和职责的清晰分离。

```mermaid
flowchart TB
    subgraph CLI["CLI Layer"]
        MAIN["cmd/clinvk/main.go"]
        APP["internal/app/"]
    end

    subgraph CORE["Core Components"]
        REG["Backend Registry<br/>internal/backend/registry.go"]
        SESS["Session Store<br/>internal/session/store.go"]
        EXEC["Executor<br/>internal/executor/executor.go"]
        CFG["Config Manager<br/>internal/config/config.go"]
    end

    subgraph HTTP["HTTP Server"]
        ROUTER["Chi Router<br/>internal/server/server.go"]
        MW["Middleware Stack<br/>internal/server/middleware/"]
        HANDLERS["API Handlers<br/>internal/server/handlers/"]
    end

    subgraph BACKENDS["AI CLI Backends"]
        CLAUDE["Claude Code<br/>internal/backend/claude.go"]
        CODEX["Codex CLI<br/>internal/backend/codex.go"]
        GEMINI["Gemini CLI<br/>internal/backend/gemini.go"]
    end

    MAIN --> APP
    APP --> REG
    APP --> SESS
    APP --> EXEC
    APP --> CFG

    ROUTER --> MW
    MW --> HANDLERS
    HANDLERS --> EXEC
    HANDLERS --> SESS

    EXEC --> CLAUDE
    EXEC --> CODEX
    EXEC --> GEMINI

    REG -.-> CLAUDE
    REG -.-> CODEX
    REG -.-> GEMINI
```bash

**CLI Layer** (`cmd/clinvk/main.go`, `internal/app/`)
: 使用 Cobra 框架的入口点和命令定义。处理标志解析、配置初始化和命令路由，用于提示执行、会话管理和工作流编排。

**Backend Registry** (`internal/backend/registry.go`)
: 线程安全的注册表，管理后端注册和发现。提供缓存的可用性检查，并支持来自多个 goroutine 的并发访问。

**Session Store** (`internal/session/store.go`)
: 具有跨进程文件锁定的持久会话管理。维护会话元数据、用于恢复功能的后端会话 ID，并处理跨 CLI 和 HTTP 上下文的会话生命周期。

**Executor** (`internal/executor/executor.go`)
: 具有 PTY 支持的过程执行引擎，用于交互式 CLI 工具。处理 stdin/stdout/stderr 流、信号管理和超时处理。

**Config Manager** (`internal/config/config.go`)
: 基于 Viper 的配置管理，支持 YAML 配置文件、环境变量和命令行标志。包括验证和热重载功能。

**Chi Router** (`internal/server/server.go`)
: 使用 go-chi/chi 的 HTTP 路由器，具有中间件链，用于请求 ID、真实 IP 提取、恢复、日志记录、速率限制、身份验证和 CORS。

**Middleware Stack** (`internal/server/middleware/`)
: 可组合的中间件组件，包括 API 密钥身份验证、速率限制、请求大小限制、指标收集和分布式跟踪。

**API Handlers** (`internal/server/handlers/`)
: 基于 Huma 的 HTTP 处理程序，提供 OpenAI 兼容和 Anthropic 兼容的端点、自定义 REST API 和流式响应。

## 核心功能

### 多后端编排

clinvoker 将 AI CLI 工具之间的差异抽象为统一接口。后端系统处理每个支持的 CLI 工具的命令构建、输出解析和会话管理。您可以根据后端的优势将任务路由到特定后端，在多个后端上运行相同的提示以进行比较，或构建链式工作流，其中一个后端的输出馈送到另一个后端。

### HTTP API 转换

内置 HTTP 服务器将任何 CLI 后端转换为具有 OpenAI 兼容和 Anthropic 兼容端点的 REST API。使用 OpenAI SDK 的现有应用程序可以指向 clinvoker 并立即访问 Claude Code、Codex CLI 或 Gemini CLI，而无需更改代码。API 支持流式响应、函数调用模式和正确的错误处理。

### 会话持久化

会话通过跨进程文件锁定持久保存到磁盘，允许您通过 CLI 开始对话并通过 HTTP API 继续，或在几天后恢复会话。每个会话维护后端会话 ID、工作目录、模型配置和对话历史元数据。

### 并行和链式执行

`parallel` 命令同时在多个后端上执行提示，聚合结果以进行比较。`chain` 命令创建工作流，其中每个步骤可以使用不同的后端，实现"Claude 架构、Codex 实现、Gemini 审查"等模式。

## 快速开始

### 安装

```bash
curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
```text

### 基本用法

使用默认后端运行提示：

```bash
clinvk "解释此代码库的架构"
```text

指定后端和模型：

```bash
clinvk --backend claude --model claude-opus-4.5 "重构此函数以改进错误处理"
```text

使用最新 GPT 模型的 Codex CLI：

```bash
clinvk --backend codex --model gpt-5.2 "为 auth.go 生成单元测试"
```text

### SDK 集成示例

将 clinvoker 用作 OpenAI API 的即插即用替代品：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-api-key"
)

response = client.chat.completions.create(
    model="claude-opus-4.5",
    messages=[{"role": "user", "content": "Hello, world!"}]
)
print(response.choices[0].message.content)
```text

## 功能对比

| 功能 | clinvoker | AgentAPI | Aider | Direct CLI |
|---------|:---------:|:--------:|:-----:|:----------:|
| 多后端支持 | ✓ | ✓ | ✓ | ✗ |
| OpenAI SDK 兼容 | ✓ | ✓ | ✗ | ✗ |
| Anthropic SDK 兼容 | ✓ | ✗ | ✗ | ✗ |
| 会话持久化 | ✓ | ✗ | ✓ | ✗ |
| 并行执行 | ✓ | ✗ | ✗ | ✗ |
| 链式工作流 | ✓ | ✗ | ✗ | ✗ |
| 后端对比 | ✓ | ✗ | ✗ | ✗ |
| 速率限制 | ✓ | ✗ | ✗ | ✗ |
| API 密钥身份验证 | ✓ | ✗ | ✗ | ✗ |
| 自托管 | ✓ | ✓ | ✓ | N/A |

## 支持的后端

| 后端 | CLI 工具 | 模型 | 最适用于 |
|---------|----------|--------|----------|
| Claude Code | `claude` | claude-opus-4.5, claude-sonnet-4.5, claude-haiku-4.5 | 复杂推理、架构决策、详细分析 |
| Codex CLI | `codex` | gpt-5.2, gpt-5.2-mini, gpt-5.2-nano | 代码生成、快速实现、迭代开发 |
| Gemini CLI | `gemini` | gemini-2.5-pro, gemini-2.5-flash | 研究、摘要、创意任务、多模态输入 |

## 下一步

<div class="grid cards" markdown>

-   **快速开始**

    ---

    在 5 分钟内安装 clinvk 并运行您的第一个提示词。了解后端选择、会话管理和配置的基础知识。

    [:octicons-arrow-right-24: 快速开始](tutorials/getting-started.md)

-   **架构**

    ---

    深入了解 clinvoker 的设计原则、组件交互以及添加新后端的扩展点。

    [:octicons-arrow-right-24: 架构](concepts/architecture.md)

-   **操作指南**

    ---

    针对特定任务的实用指南，包括并行执行、链式工作流、CI/CD 集成和后端配置。

    [:octicons-arrow-right-24: 操作指南](guides/index.md)

-   **API 参考**

    ---

    完整的 REST API 文档，包含 OpenAI 兼容和 Anthropic 兼容的端点、身份验证和示例。

    [:octicons-arrow-right-24: API 参考](reference/api/rest.md)

</div>

## 社区

- **GitHub**: [signalridge/clinvoker](https://github.com/signalridge/clinvoker)
- **Issues**: [报告错误或请求功能](https://github.com/signalridge/clinvoker/issues)
- **贡献**: [开发指南](concepts/contributing.md)
