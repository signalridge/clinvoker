# Claude Code 后端

Claude Code 是 clinvk 中功能最完整的后端。

## 依赖

- 已安装 `claude` 且在 `PATH`
- 由 Claude Code 自行完成认证

## clinvk 如何调用 Claude

- 非交互执行：`claude --print ...`
- 继续会话：`claude --resume <session_id> --print ...`

## 支持的能力

- **system prompt**（`system_prompt`）
- **最大轮数**（`max_turns`）
- **输出格式**（`text` / `json` / `stream-json`）
- **无状态模式**（使用 `--no-session-persistence`）
- **工具白名单**（`allowed_tools` → `--allowedTools`）

## Approval / Sandbox

`approval_mode` 映射到 Claude 权限参数：

| approval_mode | Flag |
|---|---|
| `auto` | `--permission-mode acceptEdits` |
| `none` | `--permission-mode dontAsk` |
| `always` | `--permission-mode default` |

`sandbox_mode` 在 Claude 中 **无对应参数**。

## 模型别名

clinvk 提供轻量别名：

- `fast` / `quick` → `haiku`
- `balanced` / `default` → `sonnet`
- `best` / `powerful` → `opus`

仍可通过 `--model` 或配置指定完整模型名。

## 示例

```bash
clinvk -b claude "评审这个模块"
clinvk -b claude -m claude-opus-4-5-20251101 "深度评审"
```

配置示例：

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    system_prompt: "你是严格的代码审查员。"
```
