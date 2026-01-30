# Configuration Reference

Full reference for `~/.clinvk/config.yaml`.

## Location

Default: `~/.clinvk/config.yaml`

Override with `--config`.

## Full example

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

## Key notes

### approval_mode

Mapped per backend (best‑effort):

| Backend | auto | none | always |
|---------|------|------|--------|
| Claude | `--permission-mode acceptEdits` | `--permission-mode dontAsk` | `--permission-mode default` |
| Codex | `--ask-for-approval on-request` | `--ask-for-approval never` | `--ask-for-approval untrusted` |
| Gemini | `--approval-mode auto_edit` | `--yolo` | `--approval-mode default` |

### sandbox_mode

- Claude: no direct sandbox flag
- Codex: `--sandbox read-only|workspace-write|danger-full-access`
- Gemini: `--sandbox` (read-only/workspace), none for full

### allowed_tools

Only supported by Claude (`--allowedTools`). For other backends it is ignored.

### system_prompt

Only mapped for Claude (`--system-prompt`). Other backends ignore it.

### extra_flags

Extra flags are appended to the backend CLI invocation. For API requests, flags are validated against an allowlist.

Allowlisted flags (case‑insensitive):

- **Claude**: `--model`, `--print`, `--output-format`, `--verbose`, `--max-turns`, `--system-prompt`, `--permission-mode`, `--resume`, `--add-dir`, `--allowedtools`, `--allowed-tools`, `--no-session-persistence`, `--continue`
- **Codex**: `--model`, `--json`, `--sandbox`, `--ask-for-approval`, `--full-auto`, `--quiet`, `--config-dir`
- **Gemini**: `--model`, `--output-format`, `--sandbox`, `--approval-mode`, `--yolo`, `--debug`, `--color`, `--disable-color`
- **Common**: `-v`, `-m`, `-o`, `-q`, `-h`, `--help`, `--version`

### command_timeout_secs

Applies to CLI executions; timeout exits with code `124`.

### parallel.max_workers

Used by the HTTP API when `max_parallel` is not provided. The CLI `parallel` command defaults to `3` unless overridden by `--max-parallel` or file `max_parallel`.

### output.color

Currently not used by the CLI output renderer.

### output.show_tokens / output.show_timing

These are displayed **only for text output**. JSON output is not modified.

## Environment variables

These are read automatically by clinvk:

- `CLINVK_BACKEND`
- `CLINVK_CLAUDE_MODEL`
- `CLINVK_CODEX_MODEL`
- `CLINVK_GEMINI_MODEL`
- `CLINVK_API_KEYS`
- `CLINVK_API_KEYS_GOPASS_PATH`

## API keys

API keys are **not** stored in config for security reasons.

Use:

- `CLINVK_API_KEYS` (comma‑separated)
- `CLINVK_API_KEYS_GOPASS_PATH`

## Workdir restrictions (server)

`allowed_workdir_prefixes` and `blocked_workdir_prefixes` are matched by **path prefix**. Use prefixes **without** trailing slashes to match both the directory and its children (e.g., use `/var/www`, not `/var/www/`).
