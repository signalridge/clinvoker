# 配置参考

clinvk 所有配置选项的完整参考。

## 配置文件位置

默认位置：`~/.clinvk/config.yaml`

使用 `--config` 参数指定自定义路径：

```bash
clinvk --config /path/to/config.yaml "提示词"
```

## 完整配置示例

```yaml
# 未指定 --backend 时使用的默认后端
default_backend: claude

# 统一标志适用于所有后端
unified_flags:
  # 工具/动作执行的审批模式
  # 可选值：default, auto, none, always
  approval_mode: default

  # 文件/网络访问的沙盒模式
  # 可选值：default, read-only, workspace, full
  sandbox_mode: default

  # 启用详细输出
  verbose: false

  # 模拟执行模式 - 只显示命令不执行
  dry_run: false

  # 最大 agentic 回合数（0 = 无限制）
  max_turns: 0

  # 最大响应 token 数（0 = 后端默认，当前未映射）
  max_tokens: 0

  # 命令超时时间（秒，0 = 无超时）
  command_timeout_secs: 0

# 后端特定配置
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    approval_mode: ""
    sandbox_mode: ""
    enabled: true
    system_prompt: ""
    extra_flags: []

  codex:
    model: o3
    enabled: true
    extra_flags: []

  gemini:
    model: gemini-2.5-pro
    enabled: true
    extra_flags: []

# 会话管理设置
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []

# 输出显示设置
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
  # 通过 gopass 获取 API Key（留空禁用）
  api_keys_gopass_path: ""
  # 限流设置
  rate_limit_enabled: false
  rate_limit_rps: 10
  rate_limit_burst: 20
  rate_limit_cleanup_secs: 180
  # 安全设置
  trusted_proxies: []
  max_request_body_bytes: 10485760
  # CORS 设置
  cors_allowed_origins: []
  cors_allow_credentials: false
  cors_max_age: 300
  # 工作目录限制
  allowed_workdir_prefixes: []
  blocked_workdir_prefixes: []
  # 可观测性
  metrics_enabled: false

# 并行执行设置
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

---

## 全局设置

### default_backend

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `default_backend` | string | `claude` | 未指定 `--backend` 时的默认后端 |

可选值：`claude`、`codex`、`gemini`

```yaml
default_backend: claude
```

---

## 统一标志

适用于所有后端的全局选项，除非被覆盖。

### approval_mode

控制后端何时在执行动作前请求审批。

| 值 | 描述 | 安全级别 |
|-------|-------------|--------------|
| `default` | 让后端决定 | 中 |
| `auto` | 安全时减少提示/自动批准 | 低-中 |
| `none` | 从不请求审批 | **危险** |
| `always` | 总是请求审批 | 高 |

**后端映射：**

| 后端 | `auto` | `none` | `always` |
|---------|--------|--------|----------|
| Claude | `--permission-mode acceptEdits` | `--permission-mode dontAsk` | `--permission-mode default` |
| Codex | `--ask-for-approval on-request` | `--ask-for-approval never` | `--ask-for-approval untrusted` |
| Gemini | `--approval-mode auto_edit` | `--yolo` | `--approval-mode default` |

!!! warning "安全警告"
    `approval_mode: none` 会禁用审批提示，可能允许在没有确认的情况下执行编辑/命令。请谨慎使用，并考虑与 `sandbox_mode: read-only` 组合以获得更安全的操作。

### sandbox_mode

控制文件系统访问限制。

| 值 | 描述 | 文件访问 |
|-------|-------------|-------------|
| `default` | 让后端决定 | 因后端而异 |
| `read-only` | 只读文件访问 | 只读 |
| `workspace` | 仅访问项目目录 | 项目目录 |
| `full` | 完全文件系统访问 | 无限制 |

**后端说明：**

- **Claude**：`sandbox_mode` 不会映射为 CLI 参数（请改用 `allowed_dirs` 和审批设置）
- **Gemini**：`read-only` 和 `workspace` 都映射为 `--sandbox`（无区别）
- **Codex**：映射为 `--sandbox read-only|workspace-write|danger-full-access`

### verbose

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `verbose` | boolean | `false` | 启用后端的详细输出 |

### dry_run

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `dry_run` | boolean | `false` | 显示将要执行的命令而不运行 |

### max_turns

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `max_turns` | integer | `0` | 最大 agentic 回合数（0 = 无限制） |

### max_tokens

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `max_tokens` | integer | `0` | 最大响应 token 数（0 = 后端默认） |

!!! note
    `max_tokens` 被接受但当前不会映射为任何后端 CLI 参数。它可能被后端忽略。

### command_timeout_secs

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `command_timeout_secs` | integer | `0` | 允许后端命令运行的最长时间（秒，0 = 无超时） |

---

## 后端特定设置

在 `backends` 部分配置各个后端。

### 后端字段

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `model` | string | 后端默认 | 此后端的默认模型 |
| `allowed_tools` | string | `all` | 逗号分隔列表或 `all`（**仅 Claude**） |
| `approval_mode` | string | `""` | 覆盖统一的 `approval_mode`（空 = 使用统一设置） |
| `sandbox_mode` | string | `""` | 覆盖统一的 `sandbox_mode`（空 = 使用统一设置） |
| `enabled` | boolean | `true` | 启用/禁用后端（已存储但当前未强制执行） |
| `system_prompt` | string | `""` | 此后端的默认系统提示词 |
| `extra_flags` | array | `[]` | 传递给后端的额外 CLI 参数 |

### 后端配置示例

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    approval_mode: ""
    sandbox_mode: ""
    enabled: true
    system_prompt: "You are a helpful coding assistant."
    extra_flags:
      - "--add-dir"
      - "./docs"

  codex:
    model: o3
    enabled: true
    extra_flags:
      - "--quiet"

  gemini:
    model: gemini-2.5-pro
    enabled: true
    extra_flags:
      - "--sandbox"
```

!!! note "allowed_tools 限制"
    `allowed_tools` 选项目前仅 Claude 后端支持。为 Codex 或 Gemini 设置将无效，系统会记录警告。

---

## 会话设置

配置会话持久化和管理。

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `auto_resume` | boolean | `true` | 运行 `clinvk [prompt]` 时自动恢复最近可恢复会话 |
| `retention_days` | integer | `30` | 保留会话的天数（0 = 永久） |
| `store_token_usage` | boolean | `true` | 跟踪并存储 token 使用统计 |
| `default_tags` | array | `[]` | 新会话的默认标签 |

```yaml
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []
```

---

## 输出设置

配置输出显示首选项。

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `format` | string | `json` | 默认输出格式（`text`、`json`、`stream-json`） |
| `show_tokens` | boolean | `false` | 在输出中显示 token 使用 |
| `show_timing` | boolean | `false` | 在输出中显示执行时间 |
| `color` | boolean | `true` | 启用彩色输出 |

```yaml
output:
  format: json
  show_tokens: false
  show_timing: false
  color: true
```

---

## 服务器设置

配置 HTTP API 服务器（与 `clinvk serve` 一起使用）。

### 连接设置

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `host` | string | `127.0.0.1` | 服务器绑定地址 |
| `port` | integer | `8080` | 监听端口 |

### 超时设置

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `request_timeout_secs` | integer | `300` | 请求处理超时 |
| `read_timeout_secs` | integer | `30` | 读取超时 |
| `write_timeout_secs` | integer | `300` | 写入超时 |
| `idle_timeout_secs` | integer | `120` | 空闲连接超时 |

### 限流设置

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `rate_limit_enabled` | boolean | `false` | 启用按 IP 限流 |
| `rate_limit_rps` | integer | `10` | 每 IP 每秒请求数 |
| `rate_limit_burst` | integer | `20` | 限流突发值 |
| `rate_limit_cleanup_secs` | integer | `180` | 限流表清理间隔 |

### 安全设置

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `trusted_proxies` | array | `[]` | 可信代理；为空时忽略代理头 |
| `max_request_body_bytes` | integer | `10485760` | 请求体最大大小（0 = 无限制） |
| `api_keys_gopass_path` | string | `""` | gopass 中 API Key 的路径 |

### CORS 设置

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `cors_allowed_origins` | array | `[]` | 允许的 CORS 来源（空 = 仅本地） |
| `cors_allow_credentials` | boolean | `false` | 允许 CORS 请求携带凭证 |
| `cors_max_age` | integer | `300` | CORS 预检缓存最大时间（秒） |

### 工作目录限制

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `allowed_workdir_prefixes` | array | `[]` | 允许的工作目录前缀 |
| `blocked_workdir_prefixes` | array | `[]` | 阻止的工作目录前缀 |

### 可观测性

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `metrics_enabled` | boolean | `false` | 启用 Prometheus `/metrics` 端点 |

!!! note "API Key"
    你可以通过环境变量 `CLINVK_API_KEYS`（逗号分隔）或 `server.api_keys_gopass_path` 提供 API Key。出于安全原因，Key 不会直接存储在配置文件中。

---

## 并行设置

配置并行执行的默认行为。

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `max_workers` | integer | `3` | 最大并发任务数 |
| `fail_fast` | boolean | `false` | 第一个失败时停止 |
| `aggregate_output` | boolean | `true` | 在摘要中组合任务输出 |

```yaml
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

---

## 配置优先级

值按以下顺序解析（从高到低）：

1. **CLI 参数** - `clinvk --backend codex`
2. **环境变量** - `CLINVK_BACKEND=codex`
3. **配置文件** - `~/.clinvk/config.yaml`
4. **默认值** - 内置默认值

## 另请参阅

- [环境变量](environment.md) - 基于环境的配置
- [config 命令](cli/config.md) - 通过 CLI 管理配置
