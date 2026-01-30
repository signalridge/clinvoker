# clinvk resume

恢复已保存的会话。

## 语法

```bash
clinvk resume [session-id] [prompt]
```

## 说明

恢复历史会话。会话必须包含后端 session id 才能恢复。

不提供会话 ID 时会进入交互选择。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|-------|------|---------|-------------|
| `--last` | | bool | `false` | 恢复最近会话 |
| `--backend` | `-b` | string | | 按后端过滤 |
| `--here` | | bool | `false` | 按当前工作目录过滤 |
| `--interactive` | `-i` | bool | `false` | 交互式选择器 |
| `--output-format` | `-o` | string | 配置 / `json` | `text` / `json` / `stream-json` |

## 示例

```bash
# 使用前缀或完整 ID
clinvk resume abc123

# 继续最近会话
clinvk resume --last

# 交互式选择
clinvk resume --interactive

# 继续时附带 prompt
clinvk resume --last "继续处理"
```

## 说明

- 没有后端 session id 的会话无法恢复。
- 使用 `--last` 时会选择最近**可恢复**会话。
