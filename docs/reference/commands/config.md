# clinvk config

Manage configuration.

## Synopsis

```bash
clinvk config [command]
```

## Subcommands

### show

```bash
clinvk config show
```

Prints a **summary** of config values and backend availability.

### set

```bash
clinvk config set <key> <value>
```

Writes the value into `~/.clinvk/config.yaml` using dotted keys.

Examples:

```bash
clinvk config set default_backend codex
clinvk config set output.format text
clinvk config set session.auto_resume true
```

Notes:

- Values are written as strings; for complex types edit the YAML file directly.
- For the full schema, see [Configuration Reference](../configuration.md).
