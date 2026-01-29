# 命令参考

所有 clinvk CLI 命令的完整参考。

## 概要

```
clinvk [flags] [prompt]
clinvk [command] [flags]
```

## 命令

| 命令 | 描述 |
|------|------|
| [`[prompt]`](prompt.md) | 执行提示（默认命令） |
| [`resume`](resume.md) | 恢复之前的会话 |
| [`sessions`](sessions.md) | 管理会话 |
| [`config`](config.md) | 管理配置 |
| [`parallel`](parallel.md) | 并行执行任务 |
| [`compare`](compare.md) | 对比后端响应 |
| [`chain`](chain.md) | 链式执行提示 |
| [`serve`](serve.md) | 启动 HTTP API 服务器 |
| `version` | 显示版本信息 |
| `help` | 显示帮助 |

## 全局参数

这些参数适用于所有命令：

| 参数 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--backend` | `-b` | string | `claude` | 使用的 AI 后端 |
| `--model` | `-m` | string | | 使用的模型 |
| `--workdir` | `-w` | string | | 工作目录 |
| `--output-format` | `-o` | string | `text` | 输出格式 |
| `--config` | | string | | 配置文件路径 |
| `--dry-run` | | bool | `false` | 只显示命令 |
| `--ephemeral` | | bool | `false` | 无状态模式 |
| `--help` | `-h` | | | 显示帮助 |

## 参数详情

### --backend, -b

选择要使用的 AI 后端：

```bash
clinvk --backend claude "提示"
clinvk -b codex "提示"
clinvk -b gemini "提示"
```

可用后端：`claude`, `codex`, `gemini`

### --model, -m

覆盖所选后端的默认模型：

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "提示"
clinvk -b codex -m o3-mini "提示"
```

### --workdir, -w

设置 AI 后端的工作目录：

```bash
clinvk --workdir /path/to/project "分析这个代码库"
```

### --output-format, -o

控制输出格式：

| 值 | 描述 |
|---|------|
| `text` | 纯文本（默认） |
| `json` | 结构化 JSON |
| `stream-json` | 流式 JSON 事件 |

### --dry-run

显示将要执行的命令而不运行：

```bash
clinvk --dry-run "实现功能 X"
# 输出：Would execute: claude --model claude-opus-4-5-20251101 "实现功能 X"
```

### --ephemeral

在无状态模式下运行，不创建会话：

```bash
clinvk --ephemeral "快速问题"
```

## 示例

```bash
# 基本提示
clinvk "修复 auth.go 中的 bug"

# 带选项
clinvk -b codex -m o3 -w ./project "实现功能"

# 只显示命令
clinvk --dry-run -b claude "复杂任务"

# JSON 输出
clinvk -o json "解释这个"

# 无状态查询
clinvk --ephemeral "2+2 等于多少"
```
