# Configuration Reference

Complete reference for all clinvk configuration options.

## Configuration File Location

Default location: `~/.clinvk/config.yaml`

Use a custom path with the `--config` flag:

```bash
clinvk --config /path/to/config.yaml "prompt"
```text

## Full Configuration Example

```yaml
# Default backend to use when --backend is not specified
default_backend: claude

# Unified flags apply to all backends
unified_flags:
  # Approval mode for tool/action execution
  # Values: default, auto, none, always
  approval_mode: default

  # Sandbox mode for file/network access
  # Values: default, read-only, workspace, full
  sandbox_mode: default

  # Enable verbose output
  verbose: false

  # Dry run mode - show command without executing
  dry_run: false

  # Maximum agentic turns (0 = unlimited)
  max_turns: 0

  # Maximum response tokens (0 = backend default, currently not mapped)
  max_tokens: 0

  # Command timeout in seconds (0 = no timeout)
  command_timeout_secs: 0

# Backend-specific configuration
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

# Session management settings
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []

# Output display settings
output:
  format: json
  show_tokens: false
  show_timing: false
  color: true

# HTTP server settings
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  # Optional API keys via gopass (leave empty to disable)
  api_keys_gopass_path: ""
  # Rate limiting
  rate_limit_enabled: false
  rate_limit_rps: 10
  rate_limit_burst: 20
  rate_limit_cleanup_secs: 180
  # Security
  trusted_proxies: []
  max_request_body_bytes: 10485760
  # CORS
  cors_allowed_origins: []
  cors_allow_credentials: false
  cors_max_age: 300
  # Working directory restrictions
  allowed_workdir_prefixes: []
  blocked_workdir_prefixes: []
  # Observability
  metrics_enabled: false

# Parallel execution settings
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```yaml

---

## Global Settings

### default_backend

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `default_backend` | string | `claude` | Default backend when `--backend` is not specified |

Available values: `claude`, `codex`, `gemini`

```yaml
default_backend: claude
```yaml

---

## Unified Flags

Global options that apply to all backends unless overridden.

### approval_mode

Controls when the backend asks for approval before executing actions.

| Value | Description | Safety Level |
|-------|-------------|--------------|
| `default` | Let the backend decide | Medium |
| `auto` | Reduce prompts / auto-approve when safe | Low-Medium |
| `none` | Never ask for approval | **Dangerous** |
| `always` | Always ask for approval | High |

**Backend Mappings:**

| Backend | `auto` | `none` | `always` |
|---------|--------|--------|----------|
| Claude | `--permission-mode acceptEdits` | `--permission-mode dontAsk` | `--permission-mode default` |
| Codex | `--ask-for-approval on-request` | `--ask-for-approval never` | `--ask-for-approval untrusted` |
| Gemini | `--approval-mode auto_edit` | `--yolo` | `--approval-mode default` |

!!! warning "Security Warning"
    `approval_mode: none` disables approval prompts and may allow edits/commands without confirmation. Use with caution and consider combining with `sandbox_mode: read-only` for safer operation.

### sandbox_mode

Controls file system access restrictions.

| Value | Description | File Access |
|-------|-------------|-------------|
| `default` | Let the backend decide | Varies by backend |
| `read-only` | Read-only file access | Read only |
| `workspace` | Access to project directory only | Project directory |
| `full` | Full file system access | Unrestricted |

**Backend Notes:**

- **Claude**: `sandbox_mode` is not mapped to a CLI flag (use `allowed_dirs` and approval settings instead)
- **Gemini**: `read-only` and `workspace` both map to `--sandbox` (no distinction)
- **Codex**: Maps to `--sandbox read-only|workspace-write|danger-full-access`

### verbose

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `verbose` | boolean | `false` | Enable verbose output from backends |

### dry_run

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `dry_run` | boolean | `false` | Show the command that would be executed without running it |

### max_turns

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max_turns` | integer | `0` | Maximum number of agentic turns (0 = unlimited) |

### max_tokens

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max_tokens` | integer | `0` | Maximum response tokens (0 = backend default) |

!!! note
    `max_tokens` is accepted but currently not mapped to any backend CLI flags. It may be ignored depending on the backend.

### command_timeout_secs

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `command_timeout_secs` | integer | `0` | Maximum time in seconds to allow a backend command to run (0 = no timeout) |

---

## Backend-Specific Settings

Configure individual backends under the `backends` section.

### Backend Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | Backend default | Default model for this backend |
| `allowed_tools` | string | `all` | Comma-separated list or `all` (**Claude only**) |
| `approval_mode` | string | `""` | Override unified `approval_mode` (empty = use unified) |
| `sandbox_mode` | string | `""` | Override unified `sandbox_mode` (empty = use unified) |
| `enabled` | boolean | `true` | Enable/disable backend (stored but not currently enforced) |
| `system_prompt` | string | `""` | Default system prompt for this backend |
| `extra_flags` | array | `[]` | Additional CLI flags to pass to the backend |

### Example Backend Configuration

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
```bash

!!! note "allowed_tools Limitation"
    The `allowed_tools` option is currently only supported by the Claude backend. Setting it for Codex or Gemini will have no effect, and a warning will be logged.

---

## Session Settings

Configure session persistence and management.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `auto_resume` | boolean | `true` | Auto-resume the most recent resumable session when running `clinvk [prompt]` |
| `retention_days` | integer | `30` | Days to keep sessions (0 = forever) |
| `store_token_usage` | boolean | `true` | Track and store token usage statistics |
| `default_tags` | array | `[]` | Default tags for new sessions |

```yaml
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []
```yaml

---

## Output Settings

Configure output display preferences.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `format` | string | `json` | Default output format (`text`, `json`, `stream-json`) |
| `show_tokens` | boolean | `false` | Show token usage in output |
| `show_timing` | boolean | `false` | Show execution time in output |
| `color` | boolean | `true` | Enable colored output |

```yaml
output:
  format: json
  show_tokens: false
  show_timing: false
  color: true
```yaml

---

## Server Settings

Configure the HTTP API server (used with `clinvk serve`).

### Connection Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `host` | string | `127.0.0.1` | Bind address for the server |
| `port` | integer | `8080` | Listen port |

### Timeout Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `request_timeout_secs` | integer | `300` | Request processing timeout |
| `read_timeout_secs` | integer | `30` | Read timeout |
| `write_timeout_secs` | integer | `300` | Write timeout |
| `idle_timeout_secs` | integer | `120` | Idle connection timeout |

### Rate Limiting

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `rate_limit_enabled` | boolean | `false` | Enable per-IP rate limiting |
| `rate_limit_rps` | integer | `10` | Requests per second per IP |
| `rate_limit_burst` | integer | `20` | Burst size for rate limiting |
| `rate_limit_cleanup_secs` | integer | `180` | Cleanup interval for rate limiter entries |

### Security Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `trusted_proxies` | array | `[]` | Trusted proxies; if empty, proxy headers are ignored |
| `max_request_body_bytes` | integer | `10485760` | Max request body size (0 = unlimited) |
| `api_keys_gopass_path` | string | `""` | gopass path for API keys |

### CORS Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `cors_allowed_origins` | array | `[]` | Allowed CORS origins (empty = localhost only) |
| `cors_allow_credentials` | boolean | `false` | Allow credentials in CORS requests |
| `cors_max_age` | integer | `300` | CORS preflight cache max age in seconds |

### Working Directory Restrictions

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `allowed_workdir_prefixes` | array | `[]` | Allowed working directory prefixes |
| `blocked_workdir_prefixes` | array | `[]` | Blocked working directory prefixes |

### Observability

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `metrics_enabled` | boolean | `false` | Enable Prometheus `/metrics` endpoint |

!!! note "API Keys"
    You can provide API keys via the `CLINVK_API_KEYS` environment variable (comma-separated) or `server.api_keys_gopass_path`. Keys are not stored directly in the config file for security reasons.

---

## Parallel Settings

Configure default behavior for parallel execution.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_workers` | integer | `3` | Maximum concurrent tasks |
| `fail_fast` | boolean | `false` | Stop on first failure |
| `aggregate_output` | boolean | `true` | Combine task output in summary |

```yaml
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```text

---

## Configuration Priority

Values are resolved in this order (highest to lowest):

1. **CLI Flags** - `clinvk --backend codex`
2. **Environment Variables** - `CLINVK_BACKEND=codex`
3. **Config File** - `~/.clinvk/config.yaml`
4. **Defaults** - Built-in defaults

## See Also

- [Environment Variables](environment.md) - Environment-based configuration
- [config command](cli/config.md) - Manage configuration via CLI
