# 命令参考

所有 clinvk CLI 命令的完整参考。

## 概要

```bash
clinvk [全局参数] [提示词]
clinvk [命令] [子命令] [参数]
```

## 命令概览

| 命令 | 描述 | 常见用途 |
|---------|-------------|------------|
| [`[prompt]`](prompt.md) | 执行提示词（默认命令） | 日常 AI 任务 |
| [`resume`](resume.md) | 恢复之前的会话 | 继续对话 |
| [`sessions`](sessions.md) | 管理会话 | 列出、查看、删除会话 |
| [`config`](config.md) | 管理配置 | 查看或更改设置 |
| [`parallel`](parallel.md) | 并行执行任务 | 运行多个任务 |
| [`compare`](compare.md) | 对比后端响应 | 评估不同 AI |
| [`chain`](chain.md) | 链式执行提示词 | 多步骤工作流 |
| [`serve`](serve.md) | 启动 HTTP API 服务器 | 应用程序集成 |
| `version` | 显示版本信息 | 检查已安装版本 |
| `help` | 显示帮助 | 获取命令帮助 |

## 全局参数

这些参数适用于所有命令：

| 参数 | 简写 | 类型 | 默认值 | 描述 |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | 使用的 AI 后端 |
| `--model` | `-m` | string | | 使用的模型 |
| `--workdir` | `-w` | string | | 工作目录 |
| `--output-format` | `-o` | string | `json` | 输出格式 |
| `--config` | | string | | 配置文件路径 |
| `--dry-run` | | bool | `false` | 仅显示命令 |
| `--ephemeral` | | bool | `false` | 无状态模式 |
| `--help` | `-h` | | | 显示帮助 |

## 参数详情

### --backend, -b

选择要使用的 AI 后端：

```bash
clinvk --backend claude "提示词"
clinvk -b codex "提示词"
clinvk -b gemini "提示词"
```

可用后端：`claude`、`codex`、`gemini`

### --model, -m

覆盖所选后端的默认模型：

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "提示词"
clinvk -b codex -m o3-mini "提示词"
```

### --workdir, -w

设置 AI 后端的工作目录：

```bash
clinvk --workdir /path/to/project "分析这个代码库"
```

### --output-format, -o

控制输出格式：

| 值 | 描述 |
|-------|-------------|
| `text` | 纯文本 |
| `json` | 结构化 JSON（默认） |
| `stream-json` | 流式 JSON 事件 |

```bash
clinvk --output-format json "提示词"
clinvk -o stream-json "提示词"
```

### --config

使用自定义配置文件：

```bash
clinvk --config /path/to/config.yaml "提示词"
```

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

## 命令分类

### 核心命令

日常使用的命令：

- `[prompt]` - 执行提示词
- `resume` - 继续会话
- `sessions` - 管理会话

### 配置命令

管理设置的命令：

- `config` - 查看和修改配置

### 执行命令

高级执行模式的命令：

- `parallel` - 并发运行多个任务
- `chain` - 执行顺序流水线
- `compare` - 对比多个后端

### 服务器命令

运行 HTTP API 的命令：

- `serve` - 启动 API 服务器

## 使用示例

### 基本用法

```bash
# 执行提示词
clinvk "修复 auth.go 中的 bug"

# 指定后端
clinvk -b codex "实现功能"

# 指定模型
clinvk -b claude -m claude-sonnet-4-20250514 "快速审查"
```

### 会话管理

```bash
# 列出会话
clinvk sessions list

# 恢复最近会话
clinvk resume --last

# 删除旧会话
clinvk sessions clean --older-than 30d
```

### 配置

```bash
# 显示当前配置
clinvk config show

# 设置值
clinvk config set default_backend codex
```

### 高级执行

```bash
# 并行运行任务
clinvk parallel --file tasks.json

# 对比后端
clinvk compare --all-backends "解释这段代码"

# 执行链
clinvk chain --file pipeline.json
```

### 服务器

```bash
# 启动服务器
clinvk serve --port 8080
```

## 退出码

所有命令都返回退出码：

| 代码 | 描述 |
|------|-------------|
| 0 | 成功 |
| 1 | 一般错误 |
| 2 | 后端不可用 |
| 3 | 配置无效 |
| 4 | 会话错误 |

参见 [退出码](../exit-codes.md) 获取完整参考。

## 获取帮助

获取任何命令的帮助：

```bash
# 一般帮助
clinvk --help

# 命令帮助
clinvk [命令] --help

# 示例
clinvk parallel --help
```

## 另请参阅

- [配置参考](../configuration.md) - 配置选项
- [环境变量](../environment.md) - 基于环境的设置
- [退出码](../exit-codes.md) - 退出码参考
