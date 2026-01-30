# Configuration

Use config to set defaults once and keep CLI commands short.

## Location

Default config path:

```
~/.clinvk/config.yaml
```

Use a custom path with `--config`.

## Priority order

1. CLI flags
2. Environment variables
3. Config file
4. Built‑in defaults

## Minimal example

```yaml
default_backend: claude

output:
  format: text

session:
  auto_resume: true
  retention_days: 30
```

## Common settings

### Default backend

```yaml
default_backend: codex
```

### Output format

```yaml
output:
  format: json   # text | json | stream-json
```

### Command timeout

```yaml
unified_flags:
  command_timeout_secs: 600
```

### Session retention

```yaml
session:
  retention_days: 14
  store_token_usage: true
```

### HTTP server defaults

```yaml
server:
  host: "127.0.0.1"
  port: 8080
```

## Configure via CLI

```bash
clinvk config set default_backend codex
clinvk config set output.format text
clinvk config set session.auto_resume true
```

`config set` writes to `~/.clinvk/config.yaml` and accepts dotted keys.

## Environment variables

Common env vars:

- `CLINVK_BACKEND`
- `CLINVK_CLAUDE_MODEL`
- `CLINVK_CODEX_MODEL`
- `CLINVK_GEMINI_MODEL`
- `CLINVK_API_KEYS` (server auth)
- `CLINVK_API_KEYS_GOPASS_PATH`

Full list: [Environment Variables](../reference/environment.md).

## Full reference

See [Configuration Reference](../reference/configuration.md) for all keys and defaults.
