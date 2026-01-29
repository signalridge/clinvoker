# clinvk sessions

管理会话。

## 概要

```
clinvk sessions [command] [flags]
```

## 子命令

| 命令 | 描述 |
|------|------|
| `list` | 列出会话 |
| `show` | 显示会话详情 |
| `delete` | 删除会话 |
| `clean` | 删除旧会话 |

---

## clinvk sessions list

列出所有会话。

### 标志

| 标志 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--backend` | `-b` | string | | 按后端筛选 |
| `--status` | | string | | 按状态筛选 |
| `--limit` | `-n` | int | | 显示的最大会话数 |

### 示例

```bash
# 列出所有会话
clinvk sessions list

# 按后端筛选
clinvk sessions list --backend claude

# 限制结果
clinvk sessions list --limit 10
```

---

## clinvk sessions show

显示特定会话的详情。

### 用法

```bash
clinvk sessions show <session-id>
```

---

## clinvk sessions delete

删除特定会话。

### 用法

```bash
clinvk sessions delete <session-id>
```

---

## clinvk sessions clean

删除旧会话。

### 标志

| 标志 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `--older-than` | string | | 删除早于此时间的会话 |

### 示例

```bash
# 删除 30 天前的会话
clinvk sessions clean --older-than 30d

# 使用配置的默认保留期限
clinvk sessions clean
```

## 另请参阅

- [resume](resume.md) - 恢复会话
- [配置](../configuration.md) - 会话设置
