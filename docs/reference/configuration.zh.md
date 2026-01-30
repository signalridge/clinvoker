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
  format: json
  show_tokens: false
  show_timing: false
  color: true

# HTTP 服务器设置
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  rate_limit_enabled: false
  rate_limit_rps: 10
  rate_limit_burst: 20
  rate_limit_cleanup_secs: 180
  trusted_proxies: []
  max_request_body_bytes: 10485760

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
| `auto` | 减少提示/自动批准（因后端而异） |
| `none` | 不再弹出批准提示（**危险**） |
| `always` | 尽可能总是请求批准 |

**后端映射（best-effort）**：

| 后端 | `auto` | `none` | `always` |
|------|--------|--------|----------|
| Claude | `--permission-mode acceptEdits` | `--permission-mode dontAsk` | `--permission-mode default` |
| Codex | `--ask-for-approval on-request` | `--ask-for-approval never` | `--ask-for-approval untrusted` |
| Gemini | `--approval-mode auto_edit` | `--yolo` | `--approval-mode default` |

!!! warning "安全说明"
    `approval_mode: none` 会关闭批准提示，可能导致在没有确认的情况下执行编辑/命令（取决于后端）。做审查/只读分析时，推荐使用 `sandbox_mode: read-only` 并搭配 `approval_mode: always`。

#### sandbox_mode

| 值 | 描述 |
|---|------|
| `default` | 让后端决定 |
| `read-only` | 只读文件访问 |
| `workspace` | 仅访问项目目录 |
| `full` | 完全文件系统访问 |

**后端说明**：

- Claude：目前 `sandbox_mode` 不会映射为 CLI 参数（建议改用 `allowed_dirs`/审批设置控制）。
- Gemini：`read-only` 与 `workspace` 都会映射为 `--sandbox`（无法区分）。
- Codex：映射为 `--sandbox read-only|workspace-write|danger-full-access`。

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
| `read_timeout_secs` | int | `30` | 读取超时 |
| `write_timeout_secs` | int | `300` | 写入超时 |
| `idle_timeout_secs` | int | `120` | 空闲连接超时 |
| `rate_limit_enabled` | bool | `false` | 启用按 IP 限流 |
| `rate_limit_rps` | int | `10` | 每个 IP 每秒请求数 |
| `rate_limit_burst` | int | `20` | 限流突发值 |
| `rate_limit_cleanup_secs` | int | `180` | 限流表清理间隔 |
| `trusted_proxies` | array | `[]` | 可信代理；为空时忽略代理头 |
| `max_request_body_bytes` | int | `10485760` | 请求体最大大小（0 表示不限制） |

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
