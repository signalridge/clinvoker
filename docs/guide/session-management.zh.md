# 会话管理

会话用于跨多次运行继续对话，并记录 token 使用情况。

## 会话如何工作

- 会话保存在 `~/.clinvk/sessions`。
- 只有 **后端返回了 session id** 的会话才可恢复。
- `--ephemeral` 会禁用会话持久化。

## 快速继续

```bash
clinvk "设计新的 API"
clinvk -c "加上分页"
```

`-c/--continue` 会继续最近可恢复的会话（若指定 `--backend` 会先过滤）。

## resume 命令

```bash
# 继续最近会话
clinvk resume --last

# 交互式选择
clinvk resume --interactive

# 按后端过滤
clinvk resume --backend codex

# 仅当前工作目录
clinvk resume --here
```

继续时也可附带 prompt：

```bash
clinvk resume --last "继续处理"
```

## 列表与查看详情

```bash
clinvk sessions list
clinvk sessions list --backend claude
clinvk sessions list --status completed

clinvk sessions show <session-id>
```

## 删除与清理

```bash
clinvk sessions delete <session-id>

# 按保留策略清理（不传则使用配置）
clinvk sessions clean
clinvk sessions clean --older-than 30d
```

## 配置项

```yaml
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []
```

说明：

- `auto_resume` 会在运行根命令时自动尝试恢复最近可恢复会话（非 `--ephemeral`）。若提供了 prompt，则作为继续对话的内容。
- `store_token_usage` 仅在后端提供 token 使用信息时生效。

## 不创建会话的情况

- `clinvk --ephemeral ...`
- `parallel` / `chain` / `compare`（始终无状态）
- OpenAI/Anthropic 兼容端点（服务器无状态模式）
