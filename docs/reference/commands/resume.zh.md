# clinvk resume

恢复之前的会话。

## 概要

```
clinvk resume [session-id] [prompt] [flags]
```

## 描述

恢复之前的会话以继续对话。会话保持之前交互的上下文。

## 参数

| 参数 | 描述 |
|------|------|
| `session-id` | 要恢复的会话 ID（使用 `--last` 或 `--interactive` 时可选） |
| `prompt` | 后续提示（可选） |

## 标志

| 标志 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--last` | | bool | `false` | 恢复最近的会话 |
| `--interactive` | `-i` | bool | `false` | 交互式会话选择器 |
| `--here` | | bool | `false` | 按当前目录筛选 |
| `--backend` | `-b` | string | | 按后端筛选 |

## 示例

### 恢复上一个会话

```bash
clinvk resume --last
```

### 恢复并带上后续提示

```bash
clinvk resume --last "从上次中断的地方继续"
```

### 交互式选择器

```bash
clinvk resume --interactive
```

### 从当前目录恢复

```bash
clinvk resume --here
```

### 按后端筛选

```bash
clinvk resume --backend claude
```

### 恢复特定会话

```bash
clinvk resume abc123
clinvk resume abc123 "现在添加测试"
```

## 另请参阅

- [sessions](sessions.md) - 列出和管理会话
- [prompt](prompt.md) - 执行新提示
