# Gemini CLI 后端

Gemini CLI 支持 text / JSON 输出，并提供会话清理能力。

## 依赖

- 已安装 `gemini` 且在 `PATH`
- 由 Gemini CLI 自行完成认证

## clinvk 如何调用 Gemini

- 非交互：`gemini --output-format text ...`
- 继续会话：`gemini --resume <session_id> --output-format text ...`

clinvk 会根据需要调整输出格式。

## 支持的能力

- **输出格式**：`text` / `json` / `stream-json`
- **Sandbox**：`--sandbox`
- **Approval**：`--approval-mode` / `--yolo`
- **无状态清理**：尽可能调用 `gemini --delete-session <id>`

## Approval / Sandbox 映射

| 设置 | 映射 |
|---|---|
| `approval_mode: auto` | `--approval-mode auto_edit` |
| `approval_mode: none` | `--yolo` |
| `approval_mode: always` | `--approval-mode default` |
| `sandbox_mode: read-only` | `--sandbox` |
| `sandbox_mode: workspace` | `--sandbox` |
| `sandbox_mode: full` | 不传 sandbox 参数 |

## 模型别名

- `fast` / `quick` → `gemini-2.5-flash`
- `balanced` / `default` / `best` / `powerful` → `gemini-2.5-pro`

## 示例

```bash
clinvk -b gemini "总结这段内容"
clinvk -b gemini -m gemini-2.5-pro "深入分析"
```

配置示例：

```yaml
backends:
  gemini:
    model: gemini-2.5-pro
```
