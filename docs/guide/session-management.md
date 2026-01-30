# Session Management

clinvk automatically tracks sessions so you can resume conversations and maintain context across invocations.

## How Sessions Work

Every time you run a prompt with clinvk, a session is created (unless using `--ephemeral` mode). Sessions store:

- Backend and model used
- Working directory
- Timestamp information
- Token usage (if enabled)

Sessions are stored as JSON files in `~/.clinvk/sessions/`.

## Listing Sessions

View all your sessions:

```bash
clinvk sessions list
```

Output:

```text
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 minutes ago   1,234        fix the bug in auth.go
def456    codex     completed  2 hours ago     5,678        implement user registration
```

### Filtering Sessions

```bash
# Filter by backend
clinvk sessions list --backend claude

# Limit number of results
clinvk sessions list --limit 10

# Filter by status
clinvk sessions list --status active

# Combine filters
clinvk sessions list --backend claude --status active --limit 5
```

## Resuming Sessions

### Resume Last Session

The quickest way to continue your last conversation:

```bash
clinvk resume --last
```

Or with a follow-up prompt:

```bash
clinvk resume --last "add error handling"
```

### Interactive Picker

Browse and select from recent sessions:

```bash
clinvk resume --interactive
```

### Resume by ID

Resume a specific session:

```bash
clinvk resume abc123
clinvk resume abc123 "continue with tests"
```

### Resume from Current Directory

Only show sessions from the current working directory:

```bash
clinvk resume --here
```

### Filter by Backend

```bash
clinvk resume --backend claude
```

## Quick Continue

For simple continuation, use the `--continue` flag:

```bash
clinvk "implement the feature"
clinvk -c "now add tests"
clinvk -c "update the documentation"
```

This automatically resumes the most recent session.

## Session Details

View detailed information about a session:

```bash
clinvk sessions show abc123
```

Output:

```yaml
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

## Deleting Sessions

### Delete a Specific Session

```bash
clinvk sessions delete abc123
```

### Clean Old Sessions

Remove sessions older than a specified period:

```bash
# Delete sessions older than 30 days
clinvk sessions clean --older-than 30d

# Delete sessions older than 7 days
clinvk sessions clean --older-than 7d

# Use config default retention period
clinvk sessions clean
```

## Configuration

Session behavior can be configured in `~/.clinvk/config.yaml`:

```yaml
session:
  # Automatically resume last session in the same directory
  auto_resume: true

  # Days to keep sessions (0 = keep forever)
  retention_days: 30

  # Store token usage in session metadata
  store_token_usage: true

  # Tags automatically added to new sessions
  default_tags: []
```

## Stateless Mode

If you don't want to create a session, use ephemeral mode:

```bash
clinvk --ephemeral "quick question that doesn't need history"
```

This is useful for:

- Quick one-off queries
- Testing or debugging
- Automated scripts where history isn't needed

## Tips

!!! tip "Use Directory Filtering"
    When working on multiple projects, use `clinvk resume --here` to only see sessions from the current directory.

!!! tip "Clean Regularly"
    Set up automatic cleanup with `clinvk sessions clean` in a cron job or as part of your workflow.

!!! tip "Token Tracking"
    Enable `store_token_usage: true` in config to track your usage across sessions.

## Next Steps

- [Parallel Execution](parallel-execution.md) - Run multiple tasks concurrently
- [Configuration](../reference/configuration.md) - Configure session settings
