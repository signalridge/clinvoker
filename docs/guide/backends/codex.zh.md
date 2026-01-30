# Codex CLI 后端

Codex CLI 面向代码任务优化，输出为 JSONL。

## 依赖

- 已安装 `codex` 且在 `PATH`
- 由 Codex CLI 自行完成认证

## clinvk 如何调用 Codex

- 非交互：`codex exec --json ...`
- 继续会话：`codex exec resume <session_id> --json ...`

clinvk 内部始终请求 JSON 输出，再根据你的输出格式打印。

## 支持的能力

- **输出格式**：JSON / JSONL（text 由 clinvk 渲染）
- **Sandbox**：映射到 `--sandbox`
- **Approval**：映射到 `--ask-for-approval`
- **无状态清理**：删除 `~/.codex/sessions` 中的会话文件

## Approval / Sandbox 映射

| 设置 | 映射 |
|---|---|
| `approval_mode: auto` | `--ask-for-approval on-request` |
| `approval_mode: none` | `--ask-for-approval never` |
| `approval_mode: always` | `--ask-for-approval untrusted` |
| `sandbox_mode: read-only` | `--sandbox read-only` |
| `sandbox_mode: workspace` | `--sandbox workspace-write` |
| `sandbox_mode: full` | `--sandbox danger-full-access` |

## 模型别名

- `fast` / `quick` → `gpt-4.1-mini`
- `balanced` / `default` → `gpt-5.2`
- `best` / `powerful` → `gpt-5-codex`

## 示例

```bash
clinvk -b codex "优化这个函数"
clinvk -b codex -m o3 "重构模块"
```

配置示例：

```yaml
backends:
  codex:
    model: o3
```
