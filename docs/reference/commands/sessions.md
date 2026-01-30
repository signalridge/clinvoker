# clinvk sessions

Manage saved sessions.

## Synopsis

```bash
clinvk sessions [command]
```

## Subcommands

### list

```bash
clinvk sessions list [--backend <name>] [--status <status>] [--limit N]
```

- `--backend`, `-b`: filter by backend
- `--status`: `active`, `completed`, `error`
- `--limit`, `-n`: limit number of sessions

### show

```bash
clinvk sessions show <session-id>
```

Supports prefix IDs.

### delete

```bash
clinvk sessions delete <session-id>
```

### clean

```bash
clinvk sessions clean [--older-than 30d]
```

If `--older-than` is omitted, uses `session.retention_days` from config.

## Examples

```bash
clinvk sessions list --backend claude
clinvk sessions show abc123
clinvk sessions delete abc123
clinvk sessions clean --older-than 7d
```
