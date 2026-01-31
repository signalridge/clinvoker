# 参考文档

clinvk 命令、配置和 API 的完整技术参考。

## 概述

本参考部分提供 clinvk 所有功能、选项和行为的详细文档。当你需要关于特定命令、配置选项或 API 端点的精确信息时，请使用本部分。

## 如何使用本参考

- **CLI 命令**：查找特定命令的语法、参数和示例
- **配置**：查找所有可用的配置选项及其默认值
- **环境变量**：发现基于环境的配置选项
- **退出码**：了解程序退出码以便编写脚本
- **API 参考**：将 clinvk 集成到你的应用程序中

## 快速导航

| 章节 | 描述 | 使用场景... |
|---------|-------------|-------------|
| [CLI 命令](cli/index.md) | 命令行界面参考 | 需要命令语法或参数时 |
| [配置](configuration.md) | 配置文件选项 | 设置或修改配置时 |
| [环境变量](environment.md) | 基于环境的设置 | 通过环境配置时 |
| [退出码](exit-codes.md) | 程序退出码 | 编写 clinvk 脚本时 |
| [API 参考](api/index.md) | HTTP API 文档 | 与应用程序集成时 |

## CLI 命令概览

| 命令 | 用途 | 常见使用场景 |
|---------|---------|-----------------|
| [`clinvk [prompt]`](cli/prompt.md) | 执行提示词 | 日常 AI 辅助 |
| [`clinvk resume`](cli/resume.md) | 恢复会话 | 继续对话 |
| [`clinvk sessions`](cli/sessions.md) | 管理会话 | 清理或检查会话 |
| [`clinvk config`](cli/config.md) | 管理配置 | 查看或更改设置 |
| [`clinvk parallel`](cli/parallel.md) | 并行执行 | 运行多个任务 |
| [`clinvk compare`](cli/compare.md) | 对比后端 | 评估不同 AI |
| [`clinvk chain`](cli/chain.md) | 链式执行 | 多步骤工作流 |
| [`clinvk serve`](cli/serve.md) | HTTP API 服务 | 应用程序集成 |

## 配置优先级

配置值按以下顺序解析（从高到低优先级）：

1. **CLI 参数** - 命令行参数覆盖所有其他设置
2. **环境变量** - `CLINVK_*` 变量
3. **配置文件** - `~/.clinvk/config.yaml`
4. **默认值** - 内置默认值

```bash
# 示例：CLI 参数优先于环境变量
export CLINVK_BACKEND=codex
clinvk -b claude "提示词"  # 使用 claude，而不是 codex
```

## 常用参数参考

这些参数适用于大多数命令：

| 参数 | 简写 | 描述 | 示例 |
|------|-------|-------------|---------|
| `--backend` | `-b` | 使用的后端 | `-b codex` |
| `--model` | `-m` | 模型覆盖 | `-m o3-mini` |
| `--workdir` | `-w` | 工作目录 | `-w ./project` |
| `--output-format` | `-o` | 输出格式 | `-o json` |
| `--config` | | 自定义配置路径 | `--config /path/to/config.yaml` |
| `--dry-run` | | 仅显示命令 | `--dry-run` |
| `--help` | `-h` | 显示帮助 | `-h` |

## 后端参考

| 后端 | 二进制文件 | 默认模型 | 最适合 |
|---------|--------|---------------|----------|
| Claude | `claude` | 后端默认 | 复杂推理、代码审查 |
| Codex | `codex` | 后端默认 | 快速编码任务 |
| Gemini | `gemini` | 后端默认 | 通用辅助 |

## 输出格式

| 格式 | 描述 | 最适合 |
|--------|-------------|----------|
| `text` | 纯文本输出 | 人工阅读 |
| `json` | 结构化 JSON | 脚本编写、解析 |
| `stream-json` | 流式 JSON 事件 | 实时处理 |

## API 兼容性

clinvk 提供三种 API 样式用于集成：

| API 样式 | 端点前缀 | 最适合 |
|-----------|-----------------|----------|
| 原生 REST | `/api/v1/` | 完整 clinvk 功能 |
| OpenAI 兼容 | `/openai/v1/` | OpenAI SDK 用户 |
| Anthropic 兼容 | `/anthropic/v1/` | Anthropic SDK 用户 |

## 获取帮助

- 在任何命令后使用 `--help` 获取快速参考
- 查看 [故障排除](../concepts/troubleshooting.md) 了解常见问题
- 参见 [FAQ](../concepts/faq.md) 获取常见问题解答
