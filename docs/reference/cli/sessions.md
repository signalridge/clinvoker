# clinvk sessions

Manage sessions.

## Synopsis

```bash
clinvk sessions [command] [flags]
```

## Description

The `sessions` command provides subcommands for managing clinvk sessions. Sessions store conversation history and state, allowing you to resume conversations later.

## Subcommands

| Command | Description |
|---------|-------------|
| `list` | List all sessions |
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
| `--offset` | | int | `0` | Skip this many sessions before returning results |
| `--json` | | bool | `false` | Output machine-readable JSON |

### Examples

List all sessions:

```bash
clinvk sessions list
```

Filter by backend:

```bash
clinvk sessions list --backend claude
```

Filter by status:

```bash
clinvk sessions list --status active
```

Limit results:

```bash
clinvk sessions list --limit 10
```

Combined filters:

```bash
clinvk sessions list --backend claude --status active --limit 5
```

Paginated JSON:

```bash
clinvk sessions list --backend claude --limit 10 --offset 20 --json
```

### Output

```text
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 minutes ago   1234         fix the bug in auth.go
def456    codex     completed  2 hours ago     5678         implement user registration
ghi789    gemini    error      1 day ago       -            failed task
```

JSON output:

```json
{
  "items": [
    {
      "id": "abc123...",
      "backend": "claude",
      "status": "active",
      "last_used": "2026-02-15T01:02:03Z",
      "model": "claude-opus-4-5-20251101",
      "tags": ["feature-auth"],
      "title": "fix auth bug",
      "prompt_preview": "fix the bug in auth.go"
    }
  ],
  "total": 42,
  "limit": 10,
  "offset": 20,
  "filters": {
    "backend": "claude",
    "status": ""
  },
  "sort": {
    "by": "last_used",
    "order": "desc"
  }
}
```

---

## clinvk sessions show

Show details of a specific session.

### Usage

```bash
clinvk sessions show <session-id> [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output session details as machine-readable JSON |

### Example

```bash
clinvk sessions show abc123
```

Show JSON details:

```bash
clinvk sessions show abc123 --json
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
Tags:              [feature-auth, urgent]
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
| `--dry-run` | bool | `false` | Preview cleanup candidates without deleting data |

If not specified, uses the `session.retention_days` config value.

### Examples

Clean sessions older than 30 days:

```bash
clinvk sessions clean --older-than 30d
```

Clean sessions older than 7 days:

```bash
clinvk sessions clean --older-than 7
```

Use config default:

```bash
clinvk sessions clean
```

Preview cleanup without deleting:

```bash
clinvk sessions clean --older-than 30d --dry-run
```

### Output

```text
Deleted 15 session(s) older than 30 days.
```

Dry run output:

```text
Dry run: would delete 15 session(s) older than 30 days.
Sample session IDs: abc123..., def456...
No sessions were deleted. Re-run without --dry-run to apply.
Note: candidate sessions may change between dry-run and actual cleanup.
```

---

## Session Status

| Status | Description |
|--------|-------------|
| `active` | Session is active and can be resumed |
| `completed` | Session completed normally |
| `error` | Session ended with an error |
| `paused` | Session is paused (not currently active) |

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `session not found` | Session ID doesn't exist | Check `clinvk sessions list` |
| `invalid status filter` | Unknown status value | Use `active`, `completed`, `error`, or `paused` |
| `no sessions to clean` | No sessions match criteria | Adjust filters or retention period |

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error (e.g., session not found) |
| 4 | Session error |

## Related Commands

- [resume](resume.md) - Resume a session
- [prompt](prompt.md) - Execute a new prompt

## See Also

- [Session Management](../../guides/sessions.md) - Guide to session management
- [Configuration Reference](../configuration.md) - Session settings
