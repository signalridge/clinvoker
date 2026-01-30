# clinvk [prompt]

执行一条 prompt（根命令）。

## 语法

```bash
clinvk [flags] [prompt]
```

## 说明

对选定后端执行一条 prompt。默认会创建并保存会话（除非 `--ephemeral`）。

如果启用了 `session.auto_resume` 且存在可恢复会话，clinvk 会在合适时机自动继续。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | 配置 / `claude` | 后端选择 |
| `--model` | `-m` | string | | 模型覆盖 |
| `--workdir` | `-w` | string | 当前目录 | 工作目录 |
| `--output-format` | `-o` | string | 配置 / `json` | `text` / `json` / `stream-json` |
| `--continue` | `-c` | bool | `false` | 继续最近会话 |
| `--dry-run` | | bool | `false` | 仅打印命令，不执行 |
| `--ephemeral` | | bool | `false` | 无状态执行（不保存会话） |
| `--config` | | string | `~/.clinvk/config.yaml` | 配置文件路径 |

## 示例

```bash
clinvk "修复 auth.go 的 bug"
clinvk -b codex "优化这个函数"
clinvk -b gemini -m gemini-2.5-pro "总结这段内容"
clinvk -c "继续最近会话"
```

## 输出

### JSON（默认）

```json
{
  "backend": "claude",
  "content": "...",
  "session_id": "abc123...",
  "model": "claude-opus-4-5-20251101",
  "duration_seconds": 2.5,
  "exit_code": 0,
  "error": "",
  "usage": {
    "input_tokens": 123,
    "output_tokens": 456,
    "total_tokens": 579
  },
  "raw": {}
}
```

### Text

```bash
clinvk --output-format text "解释这个"
```

### Stream JSON（CLI）

```bash
clinvk --output-format stream-json "流式输出"
```

说明：

- CLI 流式输出为后端原生格式。
- 服务器流式输出使用统一事件（见 HTTP 服务器文档）。

## 退出码

- 成功为 `0`
- 错误为 `1`
- 后端非零退出码会被透传
- 若配置了超时，超时退出码为 `124`
