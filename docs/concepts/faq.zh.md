---
title: 常见问题
description: 关于 clinvoker 安装、使用、配置和集成的常见问题和答案。
---

# 常见问题

本 FAQ 涵盖关于 clinvoker 的常见问题，按主题组织。如果您在这里找不到答案，请查看[故障排除指南](troubleshooting.zh.md)或 [GitHub Discussions](https://github.com/signalridge/clinvoker/discussions)。

## 一般问题

### 什么是 clinvoker？

clinvoker 是一个统一的 AI CLI 封装工具，为使用多个 AI 编程助手（Claude Code、Codex CLI、Gemini CLI）提供单一接口。它提供会话管理、并行执行、后端对比和 HTTP API 服务器，用于与其他工具集成。

### 为什么使用 clinvoker 而不是单独的 CLI？

- **统一接口**：相同命令适用于所有后端
- **会话管理**：轻松跨会话恢复对话
- **并行执行**：跨后端并发运行多个任务
- **后端对比**：并排比较不同 AI 模型的响应
- **HTTP API**：将 AI 能力集成到应用程序和工作流中
- **配置级联**：跨环境一致的设置管理

### 支持哪些后端？

目前支持的后端：

| 后端 | 提供商 | 描述 |
|---------|----------|-------------|
| Claude Code | Anthropic | 具有出色代码理解能力的 AI 编程助手 |
| Codex CLI | OpenAI | 具有强大编程能力的代码专注 CLI |
| Gemini CLI | Google | 具有多模态支持的 Gemini AI CLI |

### clinvoker 免费吗？

是的，clinvoker 本身是免费开源的（MIT 许可证）。但是，底层 AI 后端可能有自己的定价和使用限制。您需要为要使用的每个后端提供有效的 API 凭证。

### clinvoker 与其他工具相比如何？

| 功能 | clinvoker | Aider | Continue |
|---------|-----------|-------|----------|
| 多后端 | 是 | 有限 | 通过配置 |
| 会话管理 | 内置 | 基于 Git | 基于编辑器 |
| HTTP API | 是 | 否 | 否 |
| 并行执行 | 是 | 否 | 否 |
| 后端对比 | 是 | 否 | 否 |

## 安装

### 如何安装 clinvoker？

有多种安装选项可用：

```bash
# Homebrew（macOS/Linux）
brew install signalridge/tap/clinvk

# Go install
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# Nix
nix run github:signalridge/clinvoker

# 从 GitHub releases 下载二进制文件
# 访问：https://github.com/signalridge/clinvoker/releases
```

详见[安装指南](../tutorials/getting-started.zh.md)。

### 需要安装所有后端吗？

不需要。clinvoker 可以使用任意组合的后端。只安装您想使用的。工具会自动检测哪些后端可用。

### 系统要求是什么？

- **操作系统**：macOS、Linux 或 Windows
- **Go**：1.24+（从源码构建）
- **内存**：50MB RAM（clinvoker 本身）
- **磁盘**：二进制文件 10MB，加上会话空间

### 如何验证哪些后端可用？

```bash
clinvk config show
```

查看每个后端部分下的 `available: true`。

## 使用

### 如何更改默认后端？

```bash
# 在配置中设置
clinvk config set default_backend codex

# 或使用环境变量
export CLINVK_BACKEND=codex

# 或为每个命令指定
clinvk -b claude "your prompt"
```

### 如何继续对话？

```bash
# 快速继续（恢复上次会话）
clinvk -c "follow up message"

# 恢复特定会话
clinvk resume <session-id> "follow up message"

# 恢复上次会话
clinvk resume --last "follow up message"
```

### 可以使用不同的模型吗？

可以，使用 `--model` 指定：

```bash
# 使用特定模型
clinvk -b claude -m claude-sonnet-4 "quick task"

# 使用模型别名
clinvk -m fast "task"      # 最快模型
clinvk -m balanced "task"  # 速度/质量平衡
clinvk -m best "task"      # 最佳质量

# 在配置中设置默认值
clinvk config set backends.claude.model claude-opus-4
```

### 如何并行运行任务？

创建任务文件并使用 parallel 命令：

```bash
# 创建 tasks.json
{
  "tasks": [
    {"prompt": "Review this code", "backend": "claude"},
    {"prompt": "Review this code", "backend": "codex"},
    {"prompt": "Review this code", "backend": "gemini"}
  ]
}

# 并行运行
clinvk parallel --file tasks.json --max-parallel 3
```

详见[并行执行指南](../guides/parallel.zh.md)。

### 如何比较后端响应？

```bash
# 比较所有可用后端
clinvk compare --all-backends "your prompt"

# 比较特定后端
clinvk compare -b claude -b codex "your prompt"

# 保存比较到文件
clinvk compare --all-backends -o comparison.md "your prompt"
```

详见[后端对比](../guides/backends/index.zh.md)。

### 如何链式多个提示？

```bash
# 创建 chain.json
{
  "steps": [
    {"backend": "claude", "prompt": "Generate a Python function to sort a list"},
    {"backend": "codex", "prompt": "Review and optimize this code: {{previous}}"},
    {"backend": "claude", "prompt": "Add tests for: {{previous}}"}
  ]
}

# 执行链
clinvk chain --file chain.json
```

详见[链式执行指南](../guides/chains.zh.md)。

## 配置

### 配置文件在哪里？

默认位置：`~/.clinvk/config.yaml`

您可以指定不同位置：

```bash
clinvk --config /path/to/config.yaml "prompt"
```

### 配置优先级是什么？

配置按此顺序解析（从高到低）：

1. CLI 参数
2. 环境变量
3. 配置文件
4. 默认值

### 如何查看当前配置？

```bash
clinvk config show
```

这显示合并所有来源后的有效配置。

### 可以为所有设置使用环境变量吗？

可以，为任何配置键添加 `CLINVK_` 前缀：

```bash
export CLINVK_BACKEND=codex
export CLINVK_TIMEOUT=120
export CLINVK_SERVER_PORT=3000
```

### 如何设置后端特定选项？

```yaml
# ~/.clinvk/config.yaml
backends:
  claude:
    model: claude-sonnet-4
    timeout: 120
  codex:
    model: gpt-5.2
    sandbox_mode: full
```

## 会话

### 会话存储在哪里？

会话以 JSON 文件形式存储在 `~/.clinvk/sessions/`。

### 如何列出所有会话？

```bash
clinvk sessions list

# 按后端过滤
clinvk sessions list --backend claude

# 显示详细信息
clinvk sessions list --verbose
```

### 如何清理旧会话？

```bash
# 清理 30 天前的会话
clinvk sessions clean --older-than 30d

# 清理所有会话
clinvk sessions clean --all

# 或手动删除
rm -rf ~/.clinvk/sessions/*
```

### 可以禁用会话跟踪吗？

可以，使用临时模式：

```bash
clinvk --ephemeral "prompt"
```

这运行时不创建或加载任何会话。

### 如何导出会话？

```bash
# 导出会话到文件
clinvk sessions export <session-id> > session.json

# 或直接复制文件
cp ~/.clinvk/sessions/<session-id>.json ./backup.json
```

## HTTP 服务器

### 如何启动服务器？

```bash
# 使用默认设置启动
clinvk serve

# 使用自定义端口启动
clinvk serve --port 3000

# 使用 API 密钥认证启动
clinvk serve --api-keys "key1,key2"
```

### 服务器有认证吗？

认证是可选的。如果您配置了 API 密钥，所有请求都必须包含有效密钥：

```bash
# 配置密钥
export CLINVK_API_KEYS="key1,key2,key3"

# 或在配置中
clinvk config set server.api_keys "key1,key2"

# 在请求中使用
curl -H "Authorization: Bearer key1" http://localhost:8080/api/v1/prompt
```

如果未配置密钥，服务器允许所有请求。

### 如何公开暴露服务器？

将其放在反向代理（nginx、Apache、Caddy）后面并启用 API 密钥：

```bash
# 绑定到所有接口（谨慎使用）
clinvk serve --host 0.0.0.0 --api-keys "your-secret-key"
```

### 可以使用 OpenAI 客户端库吗？

可以，服务器提供 OpenAI 兼容端点：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-api-key"  # 如果启用了 API 密钥则需要
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### 可以使用 Anthropic 客户端库吗？

可以，Anthropic 兼容端点也可用：

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="your-api-key"
)

response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}]
)
```

## 集成

### 如何与 Claude Code 集成？

clinvoker 与 Claude Code 一起工作：

```bash
# 通过 clinvoker 使用 Claude Code
clinvk -b claude "your prompt"

# 或启动 Claude Code 的交互模式
claude
```

详见[Claude 后端指南](../guides/backends/claude.zh.md)。

### 如何与 LangChain 一起使用？

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
    model="claude"
)

response = llm.invoke("Hello!")
```

详见[LangChain 集成指南](../guides/integrations/langchain-langgraph.zh.md)。

### 如何在 CI/CD 管道中使用？

```yaml
# GitHub Actions 示例
- name: Code Review
  env:
    CLINVK_BACKEND: claude
  run: |
    echo '{"prompt": "Review this PR", "files": ["src/"] }' | \
    clinvk parallel --file - --output-format json
```

详见[CI/CD 集成指南](../guides/integrations/ci-cd/index.zh.md)。

## 故障排除

### 为什么我的后端未被检测到？

检查 CLI 是否在 PATH 中：

```bash
which claude codex gemini
echo $PATH
```

### 为什么我的配置没有应用？

记住优先级顺序：CLI 参数 > 环境变量 > 配置文件。检查覆盖：

```bash
clinvk config show  # 显示有效配置
env | grep CLINVK   # 显示环境变量
```

### 哪里可以获得帮助？

- [故障排除指南](troubleshooting.zh.md)
- [GitHub Issues](https://github.com/signalridge/clinvoker/issues)
- [GitHub Discussions](https://github.com/signalridge/clinvoker/discussions)

## 贡献

### 如何贡献？

详见[贡献指南](contributing.zh.md)：
- 开发设置
- 编码标准
- 测试要求
- PR 流程

### 如何添加新后端？

1. 实现 `Backend` 接口
2. 添加到注册表
3. 添加统一选项映射
4. 添加测试
5. 更新文档

详见[后端系统](backend-system.zh.md)了解接口详情。

### 如何报告 bug？

在 [GitHub](https://github.com/signalridge/clinvoker/issues) 上开 issue，包含：

- clinvk 版本 (`clinvk version`)
- 操作系统和版本
- 后端版本
- 复现步骤
- 错误消息
- 调试日志（如果可能）

## 相关文档

- [故障排除](troubleshooting.zh.md) - 常见问题和解决方案
- [指南](../guides/index.zh.md) - 操作指南
- [参考](../reference/index.zh.md) - API 和 CLI 参考
- [概念](index.zh.md) - 架构和设计
