# 常见问题

## 一般问题

### 什么是 clinvoker？

clinvoker 是一个统一的 AI CLI 封装工具，让您可以通过单一接口使用多个 AI 编程助手（Claude Code、Codex CLI、Gemini CLI）。它提供会话管理、并行执行、后端对比和 HTTP API 服务器。

### 为什么使用 clinvoker 而不是单独的 CLI？

- **统一接口** - 相同命令适用于所有后端
- **会话管理** - 轻松恢复对话
- **并行执行** - 并发运行多个任务
- **后端对比** - 并排比较响应
- **HTTP API** - 将 AI 功能集成到其他工具
- **配置级联** - 一致的设置管理

### 支持哪些后端？

目前支持：

- **Claude Code** - Anthropic 的 AI 编程助手
- **Codex CLI** - OpenAI 的代码专注 CLI
- **Gemini CLI** - Google 的 Gemini AI CLI

### clinvoker 免费吗？

clinvoker 本身是免费开源的。但底层 AI 后端可能有自己的定价和使用限制。

## 安装

### 如何安装 clinvoker？

多种选项：

```bash
# Homebrew
brew install signalridge/tap/clinvk

# Go
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# Nix
nix run github:signalridge/clinvoker
```

详见 [安装](../guide/installation.md)。

### 需要安装所有后端吗？

不需要。clinvoker 可以使用任意组合的后端。只安装您想使用的。

### 如何验证哪些后端可用？

```bash
clinvk config show
```

查看每个后端下的 `available: true`。

## 使用

### 如何更改默认后端？

```bash
clinvk config set default_backend codex
```

或设置环境变量：

```bash
export CLINVK_BACKEND=codex
```

### 如何继续对话？

使用 `--continue` 或 resume 命令：

```bash
# 快速继续
clinvk -c "后续消息"

# Resume 命令
clinvk resume --last "后续消息"
```

### 可以使用不同的模型吗？

可以，使用 `--model` 指定：

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "快速任务"
```

### 如何并行运行任务？

创建任务文件并使用 parallel 命令：

```bash
clinvk parallel --file tasks.json
```

### 如何比较后端响应？

```bash
clinvk compare --all-backends "您的提示"
```

## 配置

### 配置文件在哪里？

默认位置：`~/.clinvk/config.yaml`

### 配置优先级是什么？

1. CLI 参数（最高）
2. 环境变量
3. 配置文件
4. 默认值（最低）

### 如何查看当前配置？

```bash
clinvk config show
```

## HTTP 服务器

### 服务器有认证吗？

没有。服务器没有内置认证，设计用于本地使用。

### 可以使用 OpenAI 客户端库吗？

可以，服务器在 `/openai/v1/` 提供 OpenAI 兼容端点：

```python
from openai import OpenAI
client = OpenAI(base_url="http://localhost:8080/openai/v1", api_key="not-needed")
```

## 故障排除

### 为什么我的后端未被检测到？

检查 CLI 是否在 PATH 中：

```bash
which claude codex gemini
```

### 我的配置为什么没有应用？

检查优先级：CLI 参数覆盖环境变量，环境变量覆盖配置文件。

### 哪里可以获得帮助？

- [故障排除指南](troubleshooting.md)
- [GitHub Issues](https://github.com/signalridge/clinvoker/issues)

## 贡献

### 如何贡献？

详见 [贡献指南](../development/contributing.md)。

### 如何添加新后端？

详见 [添加后端](../development/adding-backends.md)。
