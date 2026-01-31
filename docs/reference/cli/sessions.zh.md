# clinvk sessions

管理会话。

## 用法

```bash
clinvk sessions [command] [flags]
```bash

## 说明

`sessions` 命令提供用于管理 clinvk 会话的子命令。会话存储对话历史和状态，允许你以后恢复对话。

## 子命令

| 命令 | 说明 |
|---------|-------------|
| `list` | 列出所有会话 |
| `show` | 查看会话详情 |
| `delete` | 删除会话 |
| `clean` | 清理旧会话 |

---

## clinvk sessions list

列出所有会话。

### 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | | 按后端过滤 |
| `--status` | | string | | 按状态过滤（`active` / `completed` / `error` / `paused`） |
| `--limit` | `-n` | int | | 限制数量 |

### 示例

列出所有会话：

```bash
clinvk sessions list
```text

按后端过滤：

```bash
clinvk sessions list --backend claude
```text

按状态过滤：

```bash
clinvk sessions list --status active
```text

限制结果：

```bash
clinvk sessions list --limit 10
```text

组合过滤：

```bash
clinvk sessions list --backend claude --status active --limit 5
```text

### 输出

```text
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 minutes ago   1234         fix the bug in auth.go
def456    codex     completed  2 hours ago     5678         implement user registration
ghi789    gemini    error      1 day ago       -            failed task
```yaml

---

## clinvk sessions show

查看会话详情。

### 用法

```bash
clinvk sessions show <session-id>
```text

### 示例

```bash
clinvk sessions show abc123
```text

### 输出

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
Tags:              [feature-auth, urgent]
```yaml

---

## clinvk sessions delete

删除指定会话。

### 用法

```bash
clinvk sessions delete <session-id>
```text

### 示例

```bash
clinvk sessions delete abc123
```text

### 输出

```text
Session abc123 deleted.
```yaml

---

## clinvk sessions clean

清理旧会话。

### 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|---------|-------------|
| `--older-than` | string | | 删除超过指定天数的会话（如 `30` 或 `30d`） |

未指定时使用 `session.retention_days` 配置值。

### 示例

清理超过 30 天的会话：

```bash
clinvk sessions clean --older-than 30d
```text

清理超过 7 天的会话：

```bash
clinvk sessions clean --older-than 7
```text

使用配置默认值：

```bash
clinvk sessions clean
```text

### 输出

```text
Deleted 15 session(s) older than 30 days.
```text

---

## 会话状态

| 状态 | 说明 |
|--------|-------------|
| `active` | 活跃，可恢复 |
| `completed` | 正常完成 |
| `error` | 发生错误 |
| `paused` | 暂停状态 |

## 常见错误

| 错误 | 原因 | 解决方案 |
|-------|-------|----------|
| `session not found` | 会话 ID 不存在 | 使用 `clinvk sessions list` 检查 |
| `invalid status filter` | 未知状态值 | 使用 `active`、`completed`、`error` 或 `paused` |
| `no sessions to clean` | 没有会话匹配条件 | 调整过滤条件或保留期限 |

## 退出码

| 代码 | 说明 |
|------|-------------|
| 0 | 成功 |
| 1 | 错误（例如会话未找到） |
| 4 | 会话错误 |

## 相关命令

- [resume](resume.md) - 恢复会话
- [prompt](prompt.md) - 新建会话

## 另请参阅

- [会话管理](../../guides/sessions.md) - 会话管理指南
- [配置参考](../configuration.zh.md) - 会话设置
