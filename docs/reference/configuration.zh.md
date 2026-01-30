# 配置参考

`~/.clinvk/config.yaml` 的完整说明。

## 路径

默认：`~/.clinvk/config.yaml`

可用 `--config` 指定。

## 完整示例

```yaml
default_backend: claude

unified_flags:
  approval_mode: default   # default | auto | none | always
  sandbox_mode: default    # default | read-only | workspace | full
  verbose: false
  dry_run: false
  max_turns: 0
  max_tokens: 0
  command_timeout_secs: 0

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

session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []

output:
  format: json
  show_tokens: false
  show_timing: false
  color: true

parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true

server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  api_keys_gopass_path: ""
  rate_limit_enabled: false
  rate_limit_rps: 10
  rate_limit_burst: 20
  rate_limit_cleanup_secs: 180
  trusted_proxies: []
  max_request_body_bytes: 10485760
  cors_allowed_origins: []
  cors_allow_credentials: false
  cors_max_age: 300
  allowed_workdir_prefixes: []
  blocked_workdir_prefixes: []
  metrics_enabled: false
```

## 关键说明

### approval_mode

各后端映射（尽力）：

| 后端 | auto | none | always |
|---------|------|------|--------|
| Claude | `--permission-mode acceptEdits` | `--permission-mode dontAsk` | `--permission-mode default` |
| Codex | `--ask-for-approval on-request` | `--ask-for-approval never` | `--ask-for-approval untrusted` |
| Gemini | `--approval-mode auto_edit` | `--yolo` | `--approval-mode default` |

### sandbox_mode

- Claude：无对应参数
- Codex：`--sandbox read-only|workspace-write|danger-full-access`
- Gemini：`--sandbox`（read-only/workspace），full 不传参数

### allowed_tools

仅 Claude 支持（`--allowedTools`），其它后端会忽略。

### system_prompt

仅 Claude 映射（`--system-prompt`），其它后端忽略。

### extra_flags

`extra_flags` 会直接追加到后端 CLI 命令。API 请求会进行白名单校验。

白名单（不区分大小写）：

- **Claude**：`--model`, `--print`, `--output-format`, `--verbose`, `--max-turns`, `--system-prompt`, `--permission-mode`, `--resume`, `--add-dir`, `--allowedtools`, `--allowed-tools`, `--no-session-persistence`, `--continue`
- **Codex**：`--model`, `--json`, `--sandbox`, `--ask-for-approval`, `--full-auto`, `--quiet`, `--config-dir`
- **Gemini**：`--model`, `--output-format`, `--sandbox`, `--approval-mode`, `--yolo`, `--debug`, `--color`, `--disable-color`
- **通用**：`-v`, `-m`, `-o`, `-q`, `-h`, `--help`, `--version`

### command_timeout_secs

作用于 CLI 执行；超时退出码为 `124`。

### parallel.max_workers

HTTP API 在未提供 `max_parallel` 时使用该值。CLI 的 `parallel` 命令默认并发为 `3`，除非使用 `--max-parallel` 或文件中的 `max_parallel` 覆盖。

### output.color

目前 CLI 输出未使用该配置。

### output.show_tokens / output.show_timing

仅在 **text 输出** 时显示；JSON 输出不会增加字段。
## 环境变量

clinvk 会读取：

- `CLINVK_BACKEND`
- `CLINVK_CLAUDE_MODEL`
- `CLINVK_CODEX_MODEL`
- `CLINVK_GEMINI_MODEL`
- `CLINVK_API_KEYS`
- `CLINVK_API_KEYS_GOPASS_PATH`

## API Key

为安全起见，API Key 不会写入配置文件。

请使用：

- `CLINVK_API_KEYS`（逗号分隔）
- `CLINVK_API_KEYS_GOPASS_PATH`

## 工作目录限制（服务器）

`allowed_workdir_prefixes` 与 `blocked_workdir_prefixes` 使用**路径前缀匹配**。建议不要以斜杠结尾，这样可同时匹配目录本身及其子路径（例如使用 `/var/www` 而不是 `/var/www/`）。
