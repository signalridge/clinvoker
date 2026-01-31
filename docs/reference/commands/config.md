# clinvk config

Manage configuration.

## Synopsis

```bash
clinvk config [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `show` | Display current configuration summary |
| `set` | Set a configuration value |

---

## clinvk config show

Display a human-readable summary of the current configuration and backend availability.

### Usage

```bash
clinvk config show
```

### Output

```text
Default Backend: claude

Backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
  codex:
    model: o3
  gemini:
    model: gemini-2.5-pro

Session:
  auto_resume: true
  retention_days: 30

Available backends:
  claude: available
  codex: not installed
  gemini: available
```

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

# Store a backend enable flag (not enforced by CLI yet)
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
| `backends.<name>.enabled` | Store enable/disable flag |
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
