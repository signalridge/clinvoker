# clinvk

统一的 AI CLI 封装工具，用于编排多个 AI CLI 后端，支持会话持久化、并行任务执行、HTTP API 服务器和统一输出格式化。

## 功能特性

<div class="grid cards" markdown>

-   :material-robot-outline:{ .lg .middle } **多后端支持**

    ---

    在 Claude Code、Codex CLI 和 Gemini CLI 之间无缝切换

-   :material-cog-outline:{ .lg .middle } **统一选项**

    ---

    一致的配置选项适用于所有后端

-   :material-history:{ .lg .middle } **会话持久化**

    ---

    自动会话跟踪，支持恢复功能

-   :material-layers-triple:{ .lg .middle } **并行执行**

    ---

    并发运行多个 AI 任务，支持快速失败

-   :material-compare:{ .lg .middle } **后端对比**

    ---

    并排比较多个后端的响应

-   :material-link-variant:{ .lg .middle } **链式执行**

    ---

    通过多个后端顺序传递提示

-   :material-api:{ .lg .middle } **HTTP API 服务器**

    ---

    RESTful API，兼容 OpenAI 和 Anthropic 端点

-   :material-tune-vertical:{ .lg .middle } **配置级联**

    ---

    CLI 参数 → 环境变量 → 配置文件 → 默认值

</div>

## 快速开始

```bash
# 使用默认后端运行（Claude Code）
clinvk "修复 auth.go 中的 bug"

# 指定后端
clinvk --backend codex "实现用户注册"

# 恢复会话
clinvk resume --last "继续之前的工作"

# 对比后端
clinvk compare --all-backends "解释这段代码"

# 启动 HTTP API 服务器
clinvk serve --port 8080
```

## 支持的后端

| 后端 | CLI 工具 | 描述 |
|------|----------|------|
| Claude Code | `claude` | Anthropic 的 AI 编程助手 |
| Codex CLI | `codex` | OpenAI 的代码专注 CLI |
| Gemini CLI | `gemini` | Google 的 Gemini AI CLI |

## 下一步

- [安装](getting-started/installation.md) - 在您的系统上安装 clinvk
- [快速开始](getting-started/quick-start.md) - 几分钟内上手使用
- [用户指南](user-guide/index.md) - 了解所有功能
- [HTTP API](server/index.md) - 使用 REST API 服务器
