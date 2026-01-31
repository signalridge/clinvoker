# clinvk resume

Resume a previous session.

## Synopsis

```bash
clinvk resume [session-id] [prompt] [flags]
```

## Description

Resume a previous session to continue the conversation. A session can only be resumed if it has a backend session ID recorded.

## Arguments

| Argument | Description |
|----------|-------------|
| `session-id` | Session ID or prefix (optional when using `--last` or interactive picker) |
| `prompt` | Follow-up prompt (optional) |

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--last` | | bool | `false` | Resume the most recent session (filtered by other flags) |
| `--interactive` | `-i` | bool | `false` | Show interactive session picker |
| `--here` | | bool | `false` | Filter sessions by current working directory |
| `--backend` | `-b` | string | | Filter sessions by backend |

## Examples

### Resume Last Session

```bash
clinvk resume --last
```

### Resume with Follow-up

```bash
clinvk resume --last "continue from where we left off"
```

### Interactive Picker

```bash
clinvk resume --interactive
```

If you run `clinvk resume` with no arguments and no `--last`, the interactive picker opens by default.

### Resume from Current Directory

```bash
clinvk resume --here
```

### Filter by Backend

```bash
clinvk resume --backend claude
```

### Resume Specific Session

```bash
clinvk resume abc123
clinvk resume abc123 "now add tests"
```

### Combine Filters

```bash
clinvk resume --here --backend claude --last
```

## Behavior

1. If `--last` is specified, resume the most recent resumable session that matches filters
2. Else if a session ID is provided, resume that session
3. Else open the interactive picker (or error if no resumable sessions exist)

## Output

Resumes the session and displays the AI's response. The same output format options as the root command apply.

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Session not found or error |

## See Also

- [sessions](sessions.md) - List and manage sessions
- [prompt](prompt.md) - Execute a new prompt
