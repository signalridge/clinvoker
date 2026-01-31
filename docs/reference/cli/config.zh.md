# clinvk config

管理配置。

## 用法

```bash
clinvk config [command] [flags]
```

## 说明

`config` 命令提供用于查看和修改 clinvk 配置的子命令。配置默认存储在 `~/.clinvk/config.yaml`。

## 子命令

| 命令 | 说明 |
|---------|-------------|
| `show` | 显示当前配置 |
| `get` | 获取特定配置值 |
| `set` | 设置配置值 |

---

## clinvk config show

显示当前配置。

### 用法

```bash
clinvk config show
```

### 输出

```yaml
default_backend: claude

unified_flags:
  approval_mode: default
  sandbox_mode: default
  verbose: false
  dry_run: false
  max_turns: 0
  max_tokens: 0
  command_timeout_secs: 0

backends:
  claude:
    model: claude-opus-4-5-20251101
    enabled: true
  codex:
    model: o3
    enabled: true
  gemini:
    model: gemini-2.5-pro
    enabled: true

session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true

output:
  format: json
  show_tokens: false
  show_timing: false
  color: true

server:
  host: 127.0.0.1
  port: 8080

parallel:
  max_workers: 3
  fail_fast: false
```

---

## clinvk config get

获取特定配置值。

### 用法

```bash
clinvk config get <key>
```

### 键格式

使用点符号表示嵌套键：

```bash
clinvk config get default_backend
clinvk config get backends.claude.model
clinvk config get session.auto_resume
```

### 示例

获取默认后端：

```bash
clinvk config get default_backend
# 输出：claude
```

获取 Claude 模型：

```bash
clinvk config get backends.claude.model
# 输出：claude-opus-4-5-20251101
```

获取会话保留时间：

```bash
clinvk config get session.retention_days
# 输出：30
```

---

## clinvk config set

设置配置值。

### 用法

```bash
clinvk config set <key> <value>
```

### 键格式

使用点符号表示嵌套键：

```bash
clinvk config set default_backend codex
clinvk config set backends.claude.model claude-sonnet-4-20250514
clinvk config set session.retention_days 7
```

### 示例

设置默认后端：

```bash
clinvk config set default_backend codex
```

设置 Claude 模型：

```bash
clinvk config set backends.claude.model claude-sonnet-4-20250514
```

设置会话保留时间：

```bash
clinvk config set session.retention_days 7
```

设置并行工作线程：

```bash
clinvk config set parallel.max_workers 5
```

启用详细输出：

```bash
clinvk config set unified_flags.verbose true
```

### 值类型

| 类型 | 示例 | 说明 |
|------|---------|-------|
| 字符串 | `"claude"` | 除非包含空格，否则引号可选 |
| 整数 | `30` | 无需引号 |
| 布尔值 | `true`、`false` | 无需引号 |
| 数组 | `["item1", "item2"]` | YAML 数组语法 |

---

## 配置文件位置

默认位置：`~/.clinvk/config.yaml`

使用 `--config` 参数指定其他文件：

```bash
clinvk --config /path/to/config.yaml config show
```

## 配置优先级

配置值按以下顺序解析（从高到低）：

1. **CLI 参数** - 命令行参数
2. **环境变量** - `CLINVK_*` 变量
3. **配置文件** - `~/.clinvk/config.yaml`
4. **默认值** - 内置默认值

## 常见错误

| 错误 | 原因 | 解决方案 |
|-------|-------|----------|
| `key not found` | 配置键不存在 | 检查拼写并使用点符号 |
| `invalid value` | 值类型不匹配 | 使用该键的正确类型 |
| `config file not found` | 配置文件缺失 | 使用 `clinvk config set` 创建 |
| `permission denied` | 无法写入配置文件 | 检查文件权限 |

## 退出码

| 代码 | 说明 |
|------|-------------|
| 0 | 成功 |
| 1 | 无效的键或值 |
| 3 | 配置错误 |

## 相关命令

- [prompt](prompt.md) - 使用配置设置执行提示词
- [serve](serve.md) - 使用配置设置启动服务器

## 另请参阅

- [配置参考](../configuration.zh.md) - 完整配置选项
- [环境变量](../environment.zh.md) - 基于环境的配置
