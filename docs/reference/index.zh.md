# 参考文档

clinvk 技术参考文档。

## 命令

所有 clinvk 命令的完整文档：

<div class="grid cards" markdown>

-   :material-console-line:{ .lg .middle } **[命令](commands/index.md)**

    ---

    所有 CLI 命令的完整参考

</div>

### 命令列表

| 命令 | 描述 |
|------|------|
| [`clinvk [prompt]`](commands/prompt.md) | 执行提示 |
| [`clinvk resume`](commands/resume.md) | 恢复会话 |
| [`clinvk sessions`](commands/sessions.md) | 管理会话 |
| [`clinvk config`](commands/config.md) | 管理配置 |
| [`clinvk parallel`](commands/parallel.md) | 并行执行 |
| [`clinvk compare`](commands/compare.md) | 后端对比 |
| [`clinvk chain`](commands/chain.md) | 链式执行 |
| [`clinvk serve`](commands/serve.md) | HTTP API 服务器 |

## 配置

<div class="grid cards" markdown>

-   :material-cog:{ .lg .middle } **[配置](configuration.md)**

    ---

    完整配置参考

-   :material-variable:{ .lg .middle } **[环境变量](environment.md)**

    ---

    环境变量参考

-   :material-numeric:{ .lg .middle } **[退出码](exit-codes.md)**

    ---

    退出码含义

</div>

## 配置优先级

配置值按以下顺序解析（从高到低优先级）：

1. **CLI 参数** - 命令行参数
2. **环境变量** - `CLINVK_*` 变量
3. **配置文件** - `~/.clinvk/config.yaml`
4. **默认值** - 内置默认值

## 快速参考

### 常用参数

| 参数 | 简写 | 描述 |
|------|------|------|
| `--backend` | `-b` | 使用的后端 |
| `--model` | `-m` | 使用的模型 |
| `--workdir` | `-w` | 工作目录 |
| `--output-format` | `-o` | 输出格式 |
| `--continue` | `-c` | 继续会话 |
| `--dry-run` | | 只显示命令 |

### 后端

| 后端 | 二进制文件 | 模型 |
|------|-----------|------|
| Claude | `claude` | claude-opus-4-5-20251101, claude-sonnet-4-20250514 |
| Codex | `codex` | o3, o3-mini |
| Gemini | `gemini` | gemini-2.5-pro, gemini-2.5-flash |

### 输出格式

| 格式 | 描述 |
|------|------|
| `text` | 纯文本输出（默认） |
| `json` | 结构化 JSON |
| `stream-json` | 流式 JSON 事件 |
