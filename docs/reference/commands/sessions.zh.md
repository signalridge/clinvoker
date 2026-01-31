# clinvk sessions

管理会话。

## 用法

```bash
clinvk sessions [command] [flags]
```

## 子命令

| 命令 | 说明 |
|------|------|
| `list` | 列出会话 |
| `show` | 查看会话详情 |
| `delete` | 删除会话 |
| `clean` | 清理旧会话 |

---

## clinvk sessions list

列出所有会话。

### 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--backend` | `-b` | string | | 按后端过滤 |
| `--status` | | string | | 按状态过滤（`active` / `completed` / `error` / `paused`） |
| `--limit` | `-n` | int | | 限制数量 |

### 输出示例

```text
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 minutes ago   1234         fix the bug in auth.go
def456    codex     completed  2 hours ago     5678         implement user registration
ghi789    gemini    error      1 day ago       -            failed task
```

---

## clinvk sessions show

查看会话详情。

```bash
clinvk sessions show <session-id>
```

输出示例：

```text
ID:                abc123
Backend:           claude
Model:             claude-opus-4-5-20251101
Status:            active
Created:           2025-01-27T10:00:00Z
Last Used:         2025-01-27T11:30:00Z (30 minutes ago)
Working Directory: /projects/myapp
Backend Session:   session-xyz
Turns:             3
Token Usage:
  Input:           1234
  Output:          5678
  Total:           6912
```

---

## clinvk sessions delete

删除指定会话。

```bash
clinvk sessions delete <session-id>
```

输出：

```text
Session abc123 deleted.
```

---

## clinvk sessions clean

清理旧会话。

### 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--older-than` | string | | 删除超过指定天数的会话（如 `30` 或 `30d`） |

未指定时使用 `session.retention_days`。

### 输出

```text
Deleted 15 session(s) older than 30 days.
```

---

## 会话状态

| 状态 | 说明 |
|------|------|
| `active` | 活跃，可恢复 |
| `completed` | 正常完成 |
| `error` | 发生错误 |
| `paused` | 暂停状态 |

## 另请参阅

- [resume](resume.md) - 恢复会话
- [Configuration](../configuration.zh.md) - 会话配置
