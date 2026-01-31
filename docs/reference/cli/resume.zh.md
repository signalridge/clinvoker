# clinvk resume

恢复之前的会话。

## 用法

```bash
clinvk resume [session-id] [prompt] [flags]
```

## 说明

恢复之前的会话继续对话。只有包含后端会话 ID 且后端支持恢复的会话才可恢复。

## 参数

| 参数 | 说明 |
|------|------|
| `session-id` | 会话 ID 或前缀（可配合 `--last` 或交互选择） |
| `prompt` | 追问内容（可选） |

## 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--last` | | bool | `false` | 恢复最近的会话（会应用过滤条件） |
| `--interactive` | `-i` | bool | `false` | 打开交互选择列表 |
| `--here` | | bool | `false` | 仅显示当前目录会话 |
| `--backend` | `-b` | string | | 按后端过滤 |

## 示例

### 恢复最近会话

恢复最近的可恢复会话：

```bash
clinvk resume --last
```

### 带追问恢复

恢复并立即发送追问：

```bash
clinvk resume --last "从上次中断的地方继续"
```

### 交互选择

使用交互选择器选择会话：

```bash
clinvk resume --interactive
```

如果不带参数执行 `clinvk resume`，将默认进入交互选择。

### 当前目录过滤

仅考虑当前目录的会话：

```bash
clinvk resume --here
```

### 按后端过滤

仅考虑特定后端的会话：

```bash
clinvk resume --backend claude
```

### 指定会话

按 ID 恢复特定会话：

```bash
clinvk resume abc123
clinvk resume abc123 "现在添加测试"
```

### 组合过滤

组合多个过滤器：

```bash
clinvk resume --here --backend claude --last
```

这会恢复当前目录中最近的 Claude 会话。

## 行为优先级

resume 命令遵循以下优先级：

1. 指定 `--last` 时，恢复满足过滤条件的最近可恢复会话
2. 否则如果提供了会话 ID，则恢复该会话
3. 否则进入交互选择（无可恢复会话时返回错误）

## 输出

恢复会话并输出模型响应，输出格式与 root 命令一致。

### 输出示例

```text
Resuming session abc123 (claude)

> 从上次中断的地方继续

我已经审查了你对 auth 模块的更改。以下是我发现的...
```

## 常见错误

| 错误 | 原因 | 解决方案 |
|-------|-------|----------|
| `session not found` | 会话 ID 不存在 | 使用 `clinvk sessions list` 检查有效 ID |
| `session not resumable` | 会话没有后端会话 ID | 开始新会话 |
| `backend not available` | 后端 CLI 未安装 | 安装后端 |
| `no resumable sessions` | 没有可恢复的会话 | 开始新会话 |

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 会话不存在或执行失败 |
| 2 | 后端不可用 |

## 相关命令

- [sessions](sessions.md) - 管理会话
- [prompt](prompt.md) - 新建会话

## 另请参阅

- [会话管理](../../guides/sessions.md) - 会话管理指南
- [配置参考](../configuration.zh.md) - 会话设置
