# clinvk config

Manage configuration.

## Synopsis

```
clinvk config [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `show` | Display current configuration |
| `set` | Set a configuration value |

---

## clinvk config show

Display the current configuration with all sources resolved.

### Usage

```bash
clinvk config show
```

### Output

```yaml
default_backend: claude
unified_flags:
  approval_mode: default
  sandbox_mode: default
  verbose: false
  max_turns: 0
  max_tokens: 0
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    enabled: true
    available: true
  codex:
    model: o3
    enabled: true
    available: false
  gemini:
    model: gemini-2.5-pro
    enabled: true
    available: true
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
parallel:
  max_workers: 3
  fail_fast: false
```

The `available` field under each backend indicates whether the backend's CLI tool is found in PATH.

---

## clinvk config set

Set a configuration value.

### Usage

```bash
clinvk config set <key> <value>
```

### Examples

```bash
# Set default backend
clinvk config set default_backend gemini

# Set backend-specific model
clinvk config set backends.claude.model claude-sonnet-4-20250514

# Set session retention
clinvk config set session.retention_days 60

# Set server port
clinvk config set server.port 3000

# Enable/disable a backend
clinvk config set backends.gemini.enabled false

# Set parallel workers
clinvk config set parallel.max_workers 5
```

### Key Path Format

Use dot notation to access nested values:

| Key Path | Description |
|----------|-------------|
| `default_backend` | Default backend |
| `backends.<name>.model` | Backend model |
| `backends.<name>.enabled` | Enable/disable backend |
| `session.retention_days` | Session retention |
| `server.port` | Server port |
| `parallel.max_workers` | Parallel workers |

---

## Configuration File

Configuration is stored in `~/.clinvk/config.yaml`.

See [Configuration Reference](../configuration.md) for complete documentation of all options.

## Configuration Priority

Values are resolved in this order:

1. CLI flags (highest)
2. Environment variables
3. Config file
4. Defaults (lowest)

## See Also

- [Configuration Reference](../configuration.md)
- [Environment Variables](../environment.md)
