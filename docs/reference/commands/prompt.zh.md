# clinvk [prompt]

使用 AI 后端执行提示。

## 概要

```bash
clinvk [flags] [prompt]
```

## 描述

根命令使用配置的 AI 后端执行提示。这是与 clinvk 交互的主要方式。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--backend` | `-b` | string | `claude` | 使用的 AI 后端 |
| `--model` | `-m` | string | | 使用的模型 |
| `--workdir` | `-w` | string | cwd | 工作目录 |
| `--output-format` | `-o` | string | `text` | 输出格式 |
| `--continue` | `-c` | bool | `false` | 继续上一个会话 |
| `--dry-run` | | bool | `false` | 只显示命令 |
| `--ephemeral` | | bool | `false` | 无状态模式 |

## 示例

### 基本用法

```bash
clinvk "修复 auth.go 中的 bug"
```bash

### 指定后端

```bash
clinvk --backend codex "实现用户注册"
clinvk -b gemini "解释这个算法"
```

### 指定模型

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "快速审查"
```bash

### 继续会话

```bash
clinvk "实现登录功能"
clinvk -c "现在添加密码验证"
clinvk -c "添加速率限制"
```

### JSON 输出

```bash
clinvk --output-format json "解释这段代码"
```bash

### 试运行

```bash
clinvk --dry-run "实现功能 X"
# 输出：Would execute: claude --model claude-opus-4-5-20251101 "实现功能 X"
```

### 临时模式

```bash
clinvk --ephemeral "2+2 等于多少"
```

## 退出码

| 代码 | 描述 |
|------|------|
| 0 | 成功 |
| 1 | 错误 |
|（后端）| 后端退出码（后端进程以非零退出码结束时会透传） |

详见：[退出码](../exit-codes.md)。

## 另请参阅

- [resume](resume.md) - 恢复会话
- [配置](../configuration.md) - 配置默认值
