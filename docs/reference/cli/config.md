# clinvk config

Manage configuration.

## Synopsis

```bash
clinvk config [command] [flags]
```bash

## Description

The `config` command provides subcommands for viewing and modifying clinvk configuration. The configuration is stored in `~/.clinvk/config.yaml` by default.

## Subcommands

| Command | Description |
|---------|-------------|
| `show` | Display current configuration |
| `get` | Get a specific configuration value |
| `set` | Set a configuration value |

---

## clinvk config show

Display the current configuration.

### Usage

```bash
clinvk config show
```text

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
  auto_resume: true
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
```yaml

---

## clinvk config get

Get a specific configuration value.

### Usage

```bash
clinvk config get <key>
```text

### Key Format

Use dot notation for nested keys:

```bash
clinvk config get default_backend
clinvk config get backends.claude.model
clinvk config get session.auto_resume
```text

### Examples

Get default backend:

```bash
clinvk config get default_backend
# Output: claude
```text

Get Claude model:

```bash
clinvk config get backends.claude.model
# Output: claude-opus-4-5-20251101
```text

Get session retention:

```bash
clinvk config get session.retention_days
# Output: 30
```yaml

---

## clinvk config set

Set a configuration value.

### Usage

```bash
clinvk config set <key> <value>
```text

### Key Format

Use dot notation for nested keys:

```bash
clinvk config set default_backend codex
clinvk config set backends.claude.model claude-sonnet-4-20250514
clinvk config set session.retention_days 7
```text

### Examples

Set default backend:

```bash
clinvk config set default_backend codex
```text

Set Claude model:

```bash
clinvk config set backends.claude.model claude-sonnet-4-20250514
```text

Set session retention:

```bash
clinvk config set session.retention_days 7
```text

Set parallel workers:

```bash
clinvk config set parallel.max_workers 5
```text

Enable verbose output:

```bash
clinvk config set unified_flags.verbose true
```bash

### Value Types

| Type | Example | Notes |
|------|---------|-------|
| String | `"claude"` | Quotes optional unless containing spaces |
| Integer | `30` | No quotes |
| Boolean | `true`, `false` | No quotes |
| Array | `["item1", "item2"]` | YAML array syntax |

---

## Configuration File Location

Default location: `~/.clinvk/config.yaml`

Use `--config` flag to specify a different file:

```bash
clinvk --config /path/to/config.yaml config show
```text

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
| 1 | Invalid key or value |
| 3 | Configuration error |

## Related Commands

- [prompt](prompt.md) - Execute prompts with config settings
- [serve](serve.md) - Start server with config settings

## See Also

- [Configuration Reference](../configuration.md) - Complete configuration options
- [Environment Variables](../environment.md) - Environment-based configuration
