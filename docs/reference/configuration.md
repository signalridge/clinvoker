# Configuration Reference

Complete reference for clinvk configuration options.

## Configuration File

Default location: `~/.clinvk/config.yaml`

Use a custom path with `--config`:

```bash
clinvk --config /path/to/config.yaml "prompt"
```

## Full Configuration Example

```yaml
# Default backend to use
default_backend: claude

# Unified flags apply to all backends
unified_flags:
  # Approval mode for tool/action execution
  # Values: default, auto, none, always
  approval_mode: default

  # Sandbox mode for file/network access
  # Values: default, read-only, workspace, full
  sandbox_mode: default

  # Output format
  # Values: default, text, json, stream-json
  output_format: default

  # Enable verbose output
  verbose: false

  # Dry run mode
  dry_run: false

  # Maximum agentic turns (0 = unlimited)
  max_turns: 0

  # Maximum response tokens (0 = backend default)
  max_tokens: 0

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

# Session management
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []

# Output settings
output:
  format: text
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

# Parallel execution settings
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

## Section Reference

### default_backend

```yaml
default_backend: claude
```

The backend to use when `--backend` is not specified.

| Value | Description |
|-------|-------------|
| `claude` | Claude Code (default) |
| `codex` | Codex CLI |
| `gemini` | Gemini CLI |

---

### unified_flags

Global options that apply to all backends:

```yaml
unified_flags:
  approval_mode: default
  sandbox_mode: default
  output_format: default
  verbose: false
  dry_run: false
  max_turns: 0
  max_tokens: 0
```

#### approval_mode

| Value | Description |
|-------|-------------|
| `default` | Let backend decide |
| `auto` | Reduce prompts / auto-approve (backend-specific) |
| `none` | Never ask for approval prompts (**dangerous**) |
| `always` | Always ask for approval (when supported) |

**Backend mappings** (best-effort):

| Backend | `auto` | `none` | `always` |
|---------|--------|--------|----------|
| Claude | `--permission-mode acceptEdits` | `--permission-mode dontAsk` | `--permission-mode default` |
| Codex | `--ask-for-approval on-request` | `--ask-for-approval never` | `--ask-for-approval untrusted` |
| Gemini | `--approval-mode auto_edit` | `--yolo` | `--approval-mode default` |

!!! warning "Safety"
    `approval_mode: none` disables approval prompts and may allow edits/commands without confirmation (depending on the backend). Prefer `sandbox_mode: read-only` and `approval_mode: always` for audits.

#### sandbox_mode

| Value | Description |
|-------|-------------|
| `default` | Let backend decide |
| `read-only` | Read-only file access |
| `workspace` | Access to project directory only |
| `full` | Full file system access |

**Backend notes**:

- Claude: `sandbox_mode` is not mapped to a CLI flag (use `allowed_dirs` / approval settings instead).
- Gemini: `read-only` and `workspace` both map to `--sandbox` (no distinction).
- Codex: maps to `--sandbox read-only|workspace-write|danger-full-access`.

#### output_format

| Value | Description |
|-------|-------------|
| `default` | Use backend's default |
| `text` | Plain text |
| `json` | Structured JSON |
| `stream-json` | Streaming JSON events |

---

### backends

Backend-specific configuration:

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    approval_mode: ""
    sandbox_mode: ""
    enabled: true
    system_prompt: ""
    extra_flags: []
```

#### Backend Fields

| Field | Type | Description |
|-------|------|-------------|
| `model` | string | Default model |
| `allowed_tools` | string | `all` or comma-separated list |
| `approval_mode` | string | Override unified setting |
| `sandbox_mode` | string | Override unified setting |
| `enabled` | bool | Enable/disable backend |
| `system_prompt` | string | Default system prompt |
| `extra_flags` | array | Additional CLI flags |

#### extra_flags Examples

```yaml
backends:
  claude:
    extra_flags:
      - "--add-dir"
      - "./docs"

  codex:
    extra_flags:
      - "--quiet"

  gemini:
    extra_flags:
      - "--sandbox"
```

---

### session

Session management settings:

```yaml
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `auto_resume` | bool | `true` | Auto-resume in same directory |
| `retention_days` | int | `30` | Days to keep sessions (0 = forever) |
| `store_token_usage` | bool | `true` | Track token usage |
| `default_tags` | array | `[]` | Tags for new sessions |

---

### output

Output display settings:

```yaml
output:
  format: text
  show_tokens: false
  show_timing: false
  color: true
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `format` | string | `text` | Default format |
| `show_tokens` | bool | `false` | Show token usage |
| `show_timing` | bool | `false` | Show execution time |
| `color` | bool | `true` | Colored output |

---

### server

HTTP server settings:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `host` | string | `127.0.0.1` | Bind address |
| `port` | int | `8080` | Listen port |
| `request_timeout_secs` | int | `300` | Request processing timeout |
| `read_timeout_secs` | int | `30` | Read timeout |
| `write_timeout_secs` | int | `300` | Write timeout |
| `idle_timeout_secs` | int | `120` | Idle connection timeout |

---

### parallel

Parallel execution settings:

```yaml
parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_workers` | int | `3` | Max concurrent tasks |
| `fail_fast` | bool | `false` | Stop on first failure |
| `aggregate_output` | bool | `true` | Combine task output |

---

## Configuration Priority

Values are resolved in this order (highest to lowest):

1. **CLI Flags** - `clinvk --backend codex`
2. **Environment Variables** - `CLINVK_BACKEND=codex`
3. **Config File** - `~/.clinvk/config.yaml`
4. **Defaults** - Built-in defaults

## See Also

- [Environment Variables](environment.md)
- [config command](commands/config.md)
