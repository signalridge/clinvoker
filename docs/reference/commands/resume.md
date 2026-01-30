# clinvk resume

Resume a previous session.

## Synopsis

```bash
clinvk resume [session-id] [prompt] [flags]
```

## Description

Resume a previous session to continue the conversation. Sessions maintain context from previous interactions.

## Arguments

| Argument | Description |
|----------|-------------|
| `session-id` | Session ID to resume (optional with `--last` or `--interactive`) |
| `prompt` | Follow-up prompt (optional) |

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--last` | | bool | `false` | Resume most recent session |
| `--interactive` | `-i` | bool | `false` | Interactive session picker |
| `--here` | | bool | `false` | Filter by current directory |
| `--backend` | `-b` | string | | Filter by backend |

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

This displays a list of recent sessions to choose from.

### Resume from Current Directory

```bash
clinvk resume --here
```

Only shows sessions created in the current working directory.

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

1. If `--last` is specified, resumes the most recent session (filtered if other flags present)
2. If `--interactive` is specified, shows a picker UI
3. If a session ID is provided, resumes that specific session
4. If a prompt is provided, it's sent as a follow-up message

## Session Resolution

Sessions are resolved in this priority:

1. Explicit session ID argument
2. `--last` flag (most recent matching session)
3. `--interactive` flag (user selection)

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
