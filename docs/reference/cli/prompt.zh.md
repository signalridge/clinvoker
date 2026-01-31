# clinvk [prompt]

使用指定后端执行提示词。

## 用法

```bash
clinvk [参数] [提示词]
```bash

## 说明

根命令使用当前配置的后端执行提示词，并支持会话持久化、输出格式和自动续聊。

这是默认命令 - 当你运行 `clinvk` 后跟文本时，它会作为提示词执行。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | 选择后端（`claude` / `codex` / `gemini`） |
| `--model` | `-m` | string | | 覆盖后端默认模型 |
| `--workdir` | `-w` | string | | 传给后端的工作目录 |
| `--output-format` | `-o` | string | `json` | 输出格式：`text` / `json` / `stream-json` |
| `--continue` | `-c` | bool | `false` | 继续最近一次可恢复会话 |
| `--dry-run` | | bool | `false` | 仅打印将要执行的后端命令 |
| `--ephemeral` | | bool | `false` | 无状态模式：不保存会话 |
| `--config` | | string | `~/.clinvk/config.yaml` | 自定义配置文件路径 |

## 示例

### 基本使用

执行简单提示词：

```bash
clinvk "修复 auth.go 中的 bug"
```text

### 指定后端

使用特定后端：

```bash
clinvk --backend codex "实现用户注册"
clinvk -b gemini "解释这个算法"
```text

### 指定模型

覆盖默认模型：

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "快速审查"
clinvk -b codex -m o3-mini "简单任务"
```text

### 继续会话

从之前的会话继续：

```bash
# 开始会话
clinvk "实现登录功能"

# 继续会话
clinvk -c "现在添加密码验证"

# 再次继续
clinvk -c "添加限流"
```text

### JSON 输出

获取结构化 JSON 输出：

```bash
clinvk --output-format json "解释这段代码"
```text

### 模拟执行

查看将要执行的命令：

```bash
clinvk --dry-run "实现功能 X"
# 输出：Would execute: claude --model claude-opus-4-5-20251101 "实现功能 X"
```text

### 无状态模式

运行时不创建会话：

```bash
clinvk --ephemeral "2+2 等于多少"
```text

### 设置工作目录

指定工作目录：

```bash
clinvk --workdir /path/to/project "审查代码库"
```text

## 输出

### 文本格式

使用 `--output-format text` 时，仅输出模型回复文本：

```text
代码实现了二分查找算法...
```text

### JSON 格式

```json
{
  "backend": "claude",
  "content": "响应文本...",
  "session_id": "abc123",
  "model": "claude-opus-4-5-20251101",
  "duration_seconds": 2.5,
  "exit_code": 0,
  "usage": {
    "input_tokens": 123,
    "output_tokens": 456,
    "total_tokens": 579
  },
  "raw": {
    "events": []
  }
}
```text

### 流式 JSON 格式

`stream-json` 会直接透传后端的流式输出（NDJSON/JSONL），事件结构由后端决定，并非统一格式。

## 常见错误

| 错误 | 原因 | 解决方案 |
|-------|-------|----------|
| `backend not found` | 后端 CLI 未安装 | 安装后端（例如 `npm install -g @anthropic-ai/claude-code`） |
| `session not resumable` | 会话不支持恢复 | 开始新会话 |
| `timeout` | 命令耗时过长 | 在配置中增加 `command_timeout_secs` |
| `invalid output format` | 指定了未知格式 | 使用 `text`、`json` 或 `stream-json` |

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 错误 |
| 2+ | 后端退出码（后端进程非 0 时透传） |

详见 [退出码](../exit-codes.zh.md)。

## 相关命令

- [resume](resume.md) - 恢复会话
- [sessions](sessions.md) - 管理会话
- [config](config.md) - 配置默认值

## 另请参阅

- [配置参考](../configuration.zh.md) - 配置默认值
- [环境变量](../environment.zh.md) - 基于环境的设置
