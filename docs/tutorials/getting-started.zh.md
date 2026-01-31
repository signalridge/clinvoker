---
title: 快速开始
description: clinvoker 完整入门指南 - 安装、第一个提示词和核心概念理解。
---

# clinvoker 快速开始

欢迎使用 clinvoker - 一个用于编排多个 AI 编程助手的统一命令行界面。本综合教程将指导您完成安装、配置以及与工具的首次交互。

## 什么是 clinvoker？

clinvoker（发音为 "see-el-in-voker"）是一个通用网关，统一了多个 AI 编程助手的访问，包括 Claude Code、Codex CLI 和 Gemini CLI。您无需为每个 AI 工具学习不同的命令和接口，而是使用单一、一致的界面。

### 核心优势

- **统一接口**：所有后端使用一套命令结构
- **会话管理**：跨会话的持久化对话
- **并行执行**：同时运行多个 AI 任务
- **链式工作流**：将一个后端的输出传递给另一个后端
- **HTTP API**：作为服务部署以进行集成

---

## 前置要求

在安装 clinvoker 之前，请确保您具备以下条件：

### 系统要求

| 要求 | 版本 | 说明 |
|------|------|------|
| Go | 1.24+ | 仅从源码构建时需要 |
| 操作系统 | Linux、macOS、Windows | 支持 AMD64 和 ARM64 |
| 内存 | 最低 512MB | 用于运行 CLI |
| 磁盘空间 | 100MB | 用于二进制文件和配置 |

### 后端前置要求

clinvoker 需要至少一个 AI 后端才能发挥作用：

| 后端 | 安装命令 | 文档 |
|------|---------|------|
| Claude Code | `npm install -g @anthropic-ai/claude-code` | [Claude.ai](https://claude.ai/claude-code) |
| Codex CLI | `npm install -g @openai/codex` | [GitHub](https://github.com/openai/codex-cli) |
| Gemini CLI | `npm install -g @google/gemini-cli` | [GitHub](https://github.com/google/gemini-cli) |

验证后端安装：

```bash
# 检查 Claude Code
which claude && claude --version

# 检查 Codex CLI
which codex && codex --version

# 检查 Gemini CLI
which gemini && gemini --version
```yaml

---

## 安装方法

选择最适合您环境的安装方法：

### 方法对比

| 方法 | 适用场景 | 优点 | 缺点 |
|------|----------|------|------|
| 快速安装脚本 | 首次使用用户 | 设置最快，自动配置 PATH | 需要 curl/PowerShell |
| 包管理器 | 常规使用 | 易于更新，依赖管理 | 可能不是最新版本 |
| 手动下载 | 隔离网络系统 | 完全控制版本 | 需要手动更新 |
| 从源码构建 | 开发者 | 最新功能，可定制 | 需要 Go 工具链 |

### 1. 快速安装（推荐）

最快的开始方式：

=== "macOS/Linux"

    ```bash
    curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
    ```

    此脚本将：
    1. 检测您的操作系统和架构
    2. 下载相应的二进制文件
    3. 安装到 `~/.local/bin`（或使用 sudo 安装到 `/usr/local/bin`）
    4. 如有需要更新您的 PATH

=== "Windows (PowerShell)"

    ```powershell
    irm https://raw.githubusercontent.com/signalridge/clinvoker/main/install.ps1 | iex
    ```

    PowerShell 脚本执行相同的步骤，安装到 `%LOCALAPPDATA%\Programs\clinvk`。

### 2. 包管理器

#### Homebrew (macOS/Linux)

```bash
# 添加 tap
brew tap signalridge/tap

# 安装 clinvoker
brew install clinvk

# 后续升级
brew upgrade clinvk
```text

#### Scoop (Windows)

```bash
# 添加 bucket
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket

# 安装
scoop install clinvk

# 升级
scoop update clinvk
```text

#### Nix (Linux/macOS)

```bash
# 直接运行，无需安装
nix run github:signalridge/clinvoker

# 安装到 profile
nix profile install github:signalridge/clinvoker

# 在 flake.nix 中使用
{
  inputs.clinvoker.url = "github:signalridge/clinvoker";
  nixpkgs.overlays = [ clinvoker.overlays.default ];
}
```text

#### Arch Linux (AUR)

```bash
# 使用 yay（推荐）
yay -S clinvk-bin

# 或从源码构建
yay -S clinvk
```text

#### Debian/Ubuntu

```bash
# 从 releases 页面下载
wget https://github.com/signalridge/clinvoker/releases/download/v0.1.0/clinvk_0.1.0_amd64.deb
sudo dpkg -i clinvk_*.deb
```text

#### RPM 系 (Fedora/RHEL)

```bash
# 从 releases 页面下载
wget https://github.com/signalridge/clinvoker/releases/download/v0.1.0/clinvk-0.1.0-1.x86_64.rpm
sudo rpm -i clinvk-*.rpm
```bash

### 3. 手动下载

从 [GitHub Releases](https://github.com/signalridge/clinvoker/releases) 下载预构建的二进制文件：

=== "Linux (AMD64)"

    ```bash
    VERSION="0.1.0-alpha"
    curl -LO "https://github.com/signalridge/clinvoker/releases/download/v${VERSION}/clinvoker_${VERSION}_linux_amd64.tar.gz"
    tar xzf "clinvoker_${VERSION}_linux_amd64.tar.gz"
    sudo mv clinvk /usr/local/bin/
    ```

=== "macOS (ARM64 - Apple Silicon)"

    ```bash
    VERSION="0.1.0-alpha"
    curl -LO "https://github.com/signalridge/clinvoker/releases/download/v${VERSION}/clinvoker_${VERSION}_darwin_arm64.tar.gz"
    tar xzf "clinvoker_${VERSION}_darwin_arm64.tar.gz"
    sudo mv clinvk /usr/local/bin/
    ```

=== "Windows"

    下载 `clinvoker_<version>_windows_amd64.zip` 并解压到 PATH 中的某个目录。

### 4. 从源码构建

需要 Go 1.24 或更高版本：

```bash
# 使用 go install
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# 或克隆并构建
git clone https://github.com/signalridge/clinvoker.git
cd clinvoker
go build -o clinvk ./cmd/clinvk
sudo mv clinvk /usr/local/bin/
```yaml

---

## 验证安装

安装完成后，验证一切正常：

```bash
# 检查版本
clinvk version
```text

预期输出：

```text
clinvk version v0.1.0-alpha
  commit: abc1234
  built:  2025-01-27T00:00:00Z
```text

检查检测到的后端：

```bash
clinvk config show
```text

您应该会看到根据系统上已安装内容列出的可用后端列表。

---

## 环境设置

### API 密钥

每个后端都需要自己的 API 密钥配置：

| 后端 | 配置方法 | 环境变量 |
|------|---------|---------|
| Claude | `claude config set api_key <key>` | `ANTHROPIC_API_KEY` |
| Codex | `codex config set api_key <key>` | `OPENAI_API_KEY` |
| Gemini | `gemini config set api_key <key>` | `GOOGLE_API_KEY` |

### 默认后端

设置您偏好的默认后端：

```bash
# 通过环境变量（临时）
export CLINVK_BACKEND=claude

# 通过配置（永久）
clinvk config set default_backend claude
```bash

### 配置文件

创建 `~/.clinvk/config.yaml`：

```yaml
# 未指定 -b 时的默认后端
default_backend: claude

# 统一标志适用于所有后端
unified_flags:
  approval_mode: default
  sandbox_mode: default

# 后端特定设置
backends:
  claude:
    model: claude-sonnet-4-20250514
  codex:
    model: o3
  gemini:
    model: gemini-2.5-pro

# 会话管理
session:
  auto_resume: true
  retention_days: 30
```yaml

---

## 您的第一个提示词

### 基本用法

使用默认后端运行您的第一个提示词：

```bash
clinvk "解释软件工程中的 SOLID 原则"
```text

这会将您的提示词发送到默认后端（默认是 Claude Code）并显示响应。

### 指定后端

使用不同的后端以发挥各自的优势：

```bash
# Claude 擅长复杂推理和架构设计
clinvk -b claude "为电商平台设计微服务架构"

# Codex 针对代码生成进行了优化
clinvk -b codex "用 Python 实现快速排序算法"

# Gemini 提供广泛的知识和解释
clinvk -b gemini "解释 SQL 和 NoSQL 数据库之间的权衡"
```text

### 工作目录

通过设置工作目录提供上下文：

```bash
# 审查当前目录中的代码
clinvk -w . "审查此代码库的安全问题"

# 分析特定项目
clinvk -w /path/to/project "解释此应用的架构"
```yaml

---

## 输出格式详解

clinvoker 支持三种输出格式，每种适用于不同的使用场景：

### Text 格式（默认）

适合交互式使用的人类可读输出：

```bash
clinvk -o text "解释 REST API"
```text

**特点：**
- 干净、格式化的文本
- 无元数据或结构
- 最适合终端阅读
- 适合管道传递给其他工具

**何时使用：** 交互式会话、阅读响应、快速查询

### JSON 格式

适合编程使用的结构化输出，包含元数据：

```bash
clinvk -o json "解释 REST API"
```text

**输出结构：**

```json
{
  "output": "REST (Representational State Transfer)...",
  "backend": "claude",
  "model": "claude-sonnet-4-20250514",
  "duration_ms": 2450,
  "tokens_used": 450,
  "session_id": "sess_abc123",
  "timestamp": "2025-01-27T10:30:00Z"
}
```text

**何时使用：** 脚本编写、日志记录、存储结果、API 集成

### Stream JSON 格式

适用于长时间运行任务的实时流式输出：

```bash
clinvk -o stream-json "撰写一份关于 Go 并发的综合指南"
```text

**特点：**
- 在内容可用时立即发出 JSON 对象
- 实时显示进度
- 每个块包含响应的一部分
- 最终对象包含完整的元数据

**何时使用：** 长格式内容、实时应用、进度监控

### 格式对比

| 格式 | 人类可读 | 机器可读 | 实时 | 元数据 |
|------|---------|---------|------|--------|
| text | 是 | 否 | 否 | 否 |
| json | 否 | 是 | 否 | 是 |
| stream-json | 部分 | 是 | 是 | 是 |

---

## 会话管理基础

### 理解会话

会话是与 AI 后端的持久化对话上下文。clinvoker 会自动：

1. 当您运行提示词时创建新会话
2. 将后续提示词与同一会话关联
3. 在多次交互中保持上下文

### 列会话

查看所有活动会话：

```bash
clinvk sessions list
```text

输出：

```text
ID          BACKEND  CREATED              STATUS   TAGS
sess_abc12  claude   2025-01-27 10:00:00  active   project-x
sess_def34  codex    2025-01-27 09:30:00  closed   -
```text

### 继续会话

恢复之前的对话：

```bash
# 继续最近的会话
clinvk --continue "那缓存策略呢？"

# 或使用 resume 命令
clinvk resume --last

# 恢复特定会话
clinvk resume sess_abc12
```text

### 会话最佳实践

- **使用标签** 按项目或主题组织会话
- **定期清理旧会话** 以节省磁盘空间
- **使用 `--ephemeral`** 用于不需要持久化的一次性查询

---

## 故障排除

### 问题 1："Backend not available"

**症状：**
```text
Error: backend "claude" not available
```bash

**原因和解决方案：**

1. **后端未安装**
   ```bash
   # 验证安装
   which claude

   # 如果缺失则安装
   npm install -g @anthropic-ai/claude-code
   ```

2. **后端不在 PATH 中**
   ```bash
   # 查找二进制文件
   find /usr -name "claude" 2>/dev/null

   # 添加到 PATH
   export PATH="$PATH:/path/to/claude"
   ```

3. **后端在配置中被禁用**
   ```bash
   # 检查配置
   clinvk config show

   # 启用后端
   clinvk config set backends.claude.enabled true
   ```

### 问题 2："API key not configured"

**症状：**
```text
Error: authentication failed for backend "claude"
```text

**解决方案：**

```bash
# 为 Claude 配置 API 密钥
claude config set api_key $ANTHROPIC_API_KEY

# 或设置环境变量
export ANTHROPIC_API_KEY="sk-ant-..."
```text

### 问题 3："Session not found"

**症状：**
```text
Error: no sessions found for backend "claude"
```text

**说明：** 首次运行时这是正常的。会话在第一次成功提示词后创建。

**解决方案：**

```bash
# 运行第一个提示词以创建会话
clinvk "Hello, world!"

# 现在会话将可用
clinvk sessions list
```text

如果仍有问题：

```bash
# 检查会话目录
ls -la ~/.clinvk/sessions/

# 如果损坏则重置
clinvk sessions cleanup
```yaml

---

## 后续步骤

现在您已安装并运行 clinvoker，根据您的目标探索以下路径：

### 用于代码审查自动化

1. [多后端代码审查](multi-backend-code-review.zh.md) - 设置并行审查
2. [CI/CD 集成](ci-cd-integration.zh.md) - 在您的流水线中自动化
3. [并行执行](../guides/parallel.zh.md) - 同时运行多个审查

### 用于工具集成

1. [LangChain 集成](langchain-integration.zh.md) - 连接到 LangChain
2. [HTTP 服务器](../guides/http-server.zh.md) - 作为 API 服务部署
3. [Claude Code Skills](../guides/integrations/claude-code-skills.zh.md) - 构建自定义技能

### 用于复杂工作流

1. [链式执行](../guides/chains.zh.md) - 创建多步骤流水线
2. [构建 AI Skills](building-ai-skills.zh.md) - 开发专门的 AI 代理
3. [架构概述](../concepts/architecture.zh.md) - 了解内部结构

### 快速参考

```bash
# 基本用法
clinvk "您的提示词"
clinvk -b codex "生成代码"

# 并行执行
clinvk parallel -f tasks.json

# 链式工作流
clinvk chain -f pipeline.json

# 服务器模式
clinvk serve --port 8080

# 会话管理
clinvk sessions list
clinvk resume --last
```text

---

## 总结

您已成功完成：

- 使用您偏好的方法安装 clinvoker
- 配置 API 密钥和默认后端
- 使用不同后端运行第一个提示词
- 探索输出格式及其使用场景
- 学习会话管理基础
- 识别常见问题的解决方案

clinvoker 将多个 AI 助手统一在一个界面下，支持并行执行、链式处理和 CI/CD 集成等强大工作流。后续教程将向您展示如何利用这些功能应对实际场景。
