# 会话管理

clinvk 自动跟踪会话，以便您可以恢复对话并在调用之间保持上下文。

## 会话工作原理

每次使用 clinvk 运行提示时，都会创建一个会话（除非使用 `--ephemeral` 模式）。会话存储：

- 使用的后端和模型
- 工作目录
- 时间戳信息
- Token 使用量（如果启用）

会话以 JSON 文件形式存储在 `~/.clinvk/sessions/` 中。

## 列出会话

查看所有会话：

```bash
clinvk sessions list
```

输出：

```
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 分钟前        1,234        修复 auth.go 中的 bug
def456    codex     completed  2 小时前        5,678        实现用户注册
```

### 筛选会话

```bash
# 按后端筛选
clinvk sessions list --backend claude

# 限制结果数量
clinvk sessions list --limit 10

# 按状态筛选
clinvk sessions list --status active

# 组合筛选
clinvk sessions list --backend claude --status active --limit 5
```

## 恢复会话

### 恢复上一个会话

继续上一次对话的最快方式：

```bash
clinvk resume --last
```

或带上后续提示：

```bash
clinvk resume --last "添加错误处理"
```

### 交互式选择器

浏览并选择最近的会话：

```bash
clinvk resume --interactive
```

### 按 ID 恢复

恢复特定会话：

```bash
clinvk resume abc123
clinvk resume abc123 "继续测试"
```

### 从当前目录恢复

只显示当前工作目录的会话：

```bash
clinvk resume --here
```

### 按后端筛选

```bash
clinvk resume --backend claude
```

## 快速继续

对于简单的继续，使用 `--continue` 参数：

```bash
clinvk "实现功能"
clinvk -c "现在添加测试"
clinvk -c "更新文档"
```

这会自动恢复最近的会话。

## 会话详情

查看会话的详细信息：

```bash
clinvk sessions show abc123
```

输出：

```
ID:                abc123
Backend:           claude
Model:             claude-opus-4-5-20251101
Status:            active
Created:           2025-01-27T10:00:00Z
Last Used:         2025-01-27T11:30:00Z (30 分钟前)
Working Directory: /projects/myapp
Token Usage:
  Input:           1,234
  Output:          5,678
  Cached:          500
  Total:           6,912
```

## 删除会话

### 删除特定会话

```bash
clinvk sessions delete abc123
```

### 清理旧会话

删除指定时间之前的会话：

```bash
# 删除 30 天前的会话
clinvk sessions clean --older-than 30d

# 删除 7 天前的会话
clinvk sessions clean --older-than 7d

# 使用配置的默认保留期限
clinvk sessions clean
```

## 配置

会话行为可以在 `~/.clinvk/config.yaml` 中配置：

```yaml
session:
  # 在同一目录自动恢复上一个会话
  auto_resume: true

  # 会话保留天数（0 = 永久保留）
  retention_days: 30

  # 在会话元数据中存储 token 使用量
  store_token_usage: true

  # 自动添加到新会话的标签
  default_tags: []
```

## 无状态模式

如果不想创建会话，使用临时模式：

```bash
clinvk --ephemeral "不需要历史记录的快速问题"
```

这适用于：

- 快速一次性查询
- 测试或调试
- 不需要历史记录的自动化脚本

## 提示

!!! tip "使用目录筛选"
    在多个项目上工作时，使用 `clinvk resume --here` 只查看当前目录的会话。

!!! tip "定期清理"
    使用 `clinvk sessions clean` 在 cron 任务中或作为工作流的一部分设置自动清理。

!!! tip "Token 跟踪"
    在配置中启用 `store_token_usage: true` 以跟踪跨会话的使用量。

## 下一步

- [并行执行](parallel-execution.md) - 并发运行多个任务
- [配置](../reference/configuration.md) - 配置会话设置
