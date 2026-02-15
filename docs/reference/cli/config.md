# clinvk config

Manage configuration.

## Synopsis

```bash
clinvk config [command] [flags]
```

## Description

The `config` command provides subcommands for viewing and modifying clinvk configuration. The configuration is stored in `~/.clinvk/config.yaml` by default.

## Subcommands

| Command | Description |
|---------|-------------|
| `show` | Display current configuration |
| `set` | Set a configuration value |
| `lint` | Validate configuration |

---

## clinvk config show

Display the current configuration.

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

## clinvk config set

Set a configuration value.

### Usage

```bash
clinvk config set <key> <value>
```

### Key Format

Use dot notation for nested keys:

```bash
clinvk config set default_backend codex
clinvk config set backends.claude.model claude-sonnet-4-20250514
clinvk config set session.retention_days 7
```

### Examples

Set default backend:

```bash
clinvk config set default_backend codex
```

Set Claude model:

```bash
clinvk config set backends.claude.model claude-sonnet-4-20250514
```

Set session retention:

```bash
clinvk config set session.retention_days 7
```

Set parallel workers:

```bash
clinvk config set parallel.max_workers 5
```

Enable verbose output:

```bash
clinvk config set unified_flags.verbose true
```

### Value Types

| Type | Example | Notes |
|------|---------|-------|
| String | `"claude"` | Quotes optional unless containing spaces |
| Integer | `30` | No quotes |
| Boolean | `true`, `false` | No quotes |
| Array | `["item1", "item2"]` | YAML array syntax |

---

## clinvk config lint

Validate configuration syntax and semantics.

### Usage

```bash
clinvk config lint [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | | Validate a specific config file path |
| `--json` | bool | `false` | Output validation report as JSON |

### Examples

Validate current loaded config:

```bash
clinvk config lint
```

Validate a specific file:

```bash
clinvk config lint --config ./config.example.yaml
```

Machine-readable output for CI:

```bash
clinvk config lint --json
```

JSON output example:

```json
{
  "schema_version": "v1",
  "valid": false,
  "error_count": 1,
  "errors": [
    "config validation: default_backend: invalid backend \"unknown\" (valid: claude, codex, gemini)"
  ]
}
```

---

## Configuration File Location

Default location: `~/.clinvk/config.yaml`

Use `--config` flag to specify a different file:

```bash
clinvk --config /path/to/config.yaml config show
```

## Configuration Priority

Configuration values are resolved in this order (highest to lowest):

1. **CLI Flags** - Command-line arguments
2. **Environment Variables** - `CLINVK_*` variables
3. **Config File** - `~/.clinvk/config.yaml`
4. **Defaults** - Built-in defaults

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `key not found` | Configuration key doesn't exist | Check spelling and use dot notation |
| `invalid value` | Value type doesn't match | Use correct type for the key |
| `config file not found` | Config file missing | Create one with `clinvk config set` |
| `permission denied` | Can't write config file | Check file permissions |

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General command error |
| 2 | Validation failed (`config lint` or `doctor`) |

## Related Commands

- [prompt](prompt.md) - Execute prompts with config settings
- [serve](serve.md) - Start server with config settings

## See Also

- [Configuration Reference](../configuration.md) - Complete configuration options
- [Environment Variables](../environment.md) - Environment-based configuration
