# clinvk sessions

管理已保存的会话。

## 语法

```bash
clinvk sessions [command]
```

## 子命令

### list

```bash
clinvk sessions list [--backend <name>] [--status <status>] [--limit N]
```

- `--backend`, `-b`：按后端过滤
- `--status`：`active` / `completed` / `error`
- `--limit`, `-n`：限制数量

### show

```bash
clinvk sessions show <session-id>
```

支持前缀 ID。

### delete

```bash
clinvk sessions delete <session-id>
```

### clean

```bash
clinvk sessions clean [--older-than 30d]
```

不传 `--older-than` 时使用配置 `session.retention_days`。

## 示例

```bash
clinvk sessions list --backend claude
clinvk sessions show abc123
clinvk sessions delete abc123
clinvk sessions clean --older-than 7d
```
