# clinvk sessions

Manage sessions.

## Synopsis

```
clinvk sessions [command] [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `list` | List sessions |
| `show` | Show session details |
| `delete` | Delete a session |
| `clean` | Remove old sessions |

---

## clinvk sessions list

List all sessions.

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | | Filter by backend |
| `--status` | | string | | Filter by status |
| `--limit` | `-n` | int | | Max sessions to show |

### Examples

```bash
# List all sessions
clinvk sessions list

# Filter by backend
clinvk sessions list --backend claude

# Limit results
clinvk sessions list --limit 10

# Filter by status
clinvk sessions list --status active

# Combine filters
clinvk sessions list --backend claude --status active --limit 5
```

### Output

```
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 minutes ago   1,234        fix the bug in auth.go
def456    codex     completed  2 hours ago     5,678        implement user registration
ghi789    gemini    error      1 day ago       0            failed task
```

---

## clinvk sessions show

Show details of a specific session.

### Usage

```bash
clinvk sessions show <session-id>
```

### Example

```bash
clinvk sessions show abc123
```

### Output

```
ID:                abc123
Backend:           claude
Model:             claude-opus-4-5-20251101
Status:            active
Created:           2025-01-27T10:00:00Z
Last Used:         2025-01-27T11:30:00Z (30 minutes ago)
Working Directory: /projects/myapp
Token Usage:
  Input:           1,234
  Output:          5,678
  Cached:          500
  Total:           6,912
```

---

## clinvk sessions delete

Delete a specific session.

### Usage

```bash
clinvk sessions delete <session-id>
```

### Example

```bash
clinvk sessions delete abc123
```

### Output

```
Session abc123 deleted
```

---

## clinvk sessions clean

Remove old sessions.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--older-than` | string | | Delete sessions older than this |

The `--older-than` format accepts:

- `30d` - 30 days
- `7d` - 7 days
- `24h` - 24 hours

If not specified, uses the `session.retention_days` config value.

### Examples

```bash
# Delete sessions older than 30 days
clinvk sessions clean --older-than 30d

# Delete sessions older than 7 days
clinvk sessions clean --older-than 7d

# Use config default retention
clinvk sessions clean
```

### Output

```
Cleaned 15 sessions older than 30 days
```

---

## Session Status

| Status | Description |
|--------|-------------|
| `active` | Session is active and can be resumed |
| `completed` | Session completed normally |
| `error` | Session ended with an error |

## See Also

- [resume](resume.md) - Resume a session
- [Configuration](../configuration.md) - Session settings
