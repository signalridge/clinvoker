# clinvk sessions

Manage sessions.

## Synopsis

```bash
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
| `--status` | | string | | Filter by status (`active`, `completed`, `error`, `paused`) |
| `--limit` | `-n` | int | | Max sessions to show |

### Examples

```bash
clinvk sessions list
clinvk sessions list --backend claude
clinvk sessions list --limit 10
clinvk sessions list --status active
clinvk sessions list --backend claude --status active --limit 5
```

### Output

```text
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 minutes ago   1234         fix the bug in auth.go
def456    codex     completed  2 hours ago     5678         implement user registration
ghi789    gemini    error      1 day ago       -            failed task
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

```text
ID:                abc123
Backend:           claude
Model:             claude-opus-4-5-20251101
Status:            active
Created:           2025-01-27T10:00:00Z
Last Used:         2025-01-27T11:30:00Z (30 minutes ago)
Working Directory: /projects/myapp
Backend Session:   session-xyz
Turns:             3
Token Usage:
  Input:           1234
  Output:          5678
  Total:           6912
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

```text
Session abc123 deleted.
```

---

## clinvk sessions clean

Remove old sessions.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--older-than` | string | | Delete sessions older than this many days (e.g. `30` or `30d`) |

If not specified, uses the `session.retention_days` config value.

### Examples

```bash
clinvk sessions clean --older-than 30d
clinvk sessions clean --older-than 7
clinvk sessions clean
```

### Output

```text
Deleted 15 session(s) older than 30 days.
```

---

## Session Status

| Status | Description |
|--------|-------------|
| `active` | Session is active and can be resumed |
| `completed` | Session completed normally |
| `error` | Session ended with an error |
| `paused` | Session is paused (not currently active) |

## See Also

- [resume](resume.md) - Resume a session
- [Configuration](../configuration.md) - Session settings
