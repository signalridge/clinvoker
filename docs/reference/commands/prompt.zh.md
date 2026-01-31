# clinvk [prompt]

使用指定后端执行提示词。

## 用法

```bash
clinvk [flags] [prompt]
```

## 说明

根命令会使用当前配置的后端执行提示词，并支持会话持久化、输出格式、以及自动续聊。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--backend` | `-b` | string | `claude` | 选择后端（`claude` / `codex` / `gemini`） |
| `--model` | `-m` | string | | 覆盖后端默认模型 |
| `--workdir` | `-w` | string | | 传给后端的工作目录 |
| `--output-format` | `-o` | string | `json` | 输出格式：`text` / `json` / `stream-json`（未显式指定时会使用 `output.format` 配置） |
| `--continue` | `-c` | bool | `false` | 继续最近一次可恢复会话 |
| `--dry-run` | | bool | `false` | 仅打印将要执行的后端命令 |
| `--ephemeral` | | bool | `false` | 无状态模式：不保存会话 |
| `--config` | | string | `~/.clinvk/config.yaml` | 自定义配置文件路径 |

## 示例

### 基本使用

```bash
clinvk "fix the bug in auth.go"
```

### 指定后端

```bash
clinvk --backend codex "implement user registration"
clinvk -b gemini "explain this algorithm"
```

### 指定模型

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "quick review"
```

### 继续会话

```bash
clinvk "implement the login feature"
clinvk -c "now add password validation"
clinvk -c "add rate limiting"
```

### JSON 输出

```bash
clinvk --output-format json "explain this code"
```

### Dry Run

```bash
clinvk --dry-run "implement feature X"
```

### 无状态模式

```bash
clinvk --ephemeral "what is 2+2"
```

### 指定工作目录

```bash
clinvk --workdir /path/to/project "review the codebase"
```

## 输出

### 文本输出

当 `--output-format text` 时，仅输出模型回复文本。

### JSON 输出

```json
{
  "backend": "claude",
  "content": "The response text...",
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
```

### Stream JSON

`stream-json` 会直接透传后端的流式输出（NDJSON/JSONL），事件结构由后端决定，并非统一格式。

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 错误 |
|（后端）| 后端退出码（后端进程非 0 时透传） |

详见 [Exit Codes](../exit-codes.zh.md)。

## 另请参阅

- [resume](resume.md) - 恢复会话
- [Configuration](../configuration.zh.md) - 配置说明
