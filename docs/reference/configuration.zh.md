# 配置参考

clinvk 配置选项完整参考。

## 配置文件

默认位置：`~/.clinvk/config.yaml`

使用 `--config` 指定自定义路径：

```bash
clinvk --config /path/to/config.yaml "提示"
```

## 完整配置示例

```yaml
# 默认使用的后端
default_backend: claude

# 统一标志适用于所有后端
unified_flags:
  approval_mode: default
  sandbox_mode: default
  output_format: default
  verbose: false
  dry_run: false
  max_turns: 0
  max_tokens: 0

# 后端特定配置
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    enabled: true
  codex:
    model: o3
    enabled: true
  gemini:
    model: gemini-2.5-pro
    enabled: true

# 会话管理
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []

# 输出设置
output:
  format: text
  show_tokens: false
  show_timing: false
  color: true

# HTTP 服务器设置
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300

# 并行执行设置
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

## 配置项参考

### default_backend

默认使用的后端：

| 值 | 描述 |
|---|------|
| `claude` | Claude Code（默认） |
| `codex` | Codex CLI |
| `gemini` | Gemini CLI |

### unified_flags

适用于所有后端的全局选项。

#### approval_mode

| 值 | 描述 |
|---|------|
| `default` | 让后端决定 |
| `auto` | 自动批准操作 |
| `none` | 永不批准（拒绝风险操作） |
| `always` | 总是请求批准 |

#### sandbox_mode

| 值 | 描述 |
|---|------|
| `default` | 让后端决定 |
| `read-only` | 只读文件访问 |
| `workspace` | 仅访问项目目录 |
| `full` | 完全文件系统访问 |

### backends

后端特定配置：

| 字段 | 类型 | 描述 |
|------|------|------|
| `model` | string | 默认模型 |
| `allowed_tools` | string | `all` 或逗号分隔列表 |
| `enabled` | bool | 启用/禁用后端 |
| `extra_flags` | array | 额外 CLI 参数 |

### session

会话管理设置：

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `auto_resume` | bool | `true` | 在同一目录自动恢复 |
| `retention_days` | int | `30` | 会话保留天数 |
| `store_token_usage` | bool | `true` | 跟踪 token 使用量 |

### server

HTTP 服务器设置：

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `host` | string | `127.0.0.1` | 绑定地址 |
| `port` | int | `8080` | 监听端口 |
| `request_timeout_secs` | int | `300` | 请求超时 |

### parallel

并行执行设置：

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `max_workers` | int | `3` | 最大并发任务数 |
| `fail_fast` | bool | `false` | 第一个失败时停止 |

## 配置优先级

值按以下顺序解析（从高到低）：

1. **CLI 参数**（最高）
2. **环境变量**
3. **配置文件**
4. **默认值**（最低）

## 另请参阅

- [环境变量](environment.md)
- [config 命令](commands/config.md)
