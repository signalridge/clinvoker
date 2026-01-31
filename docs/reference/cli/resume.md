# clinvk resume

Resume a previous session.

## Synopsis

```bash
clinvk resume [session-id] [prompt] [flags]
```text

## Description

Resume a previous session to continue the conversation. A session can only be resumed if it has a backend session ID recorded and the backend supports resuming.

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

Resume the most recent resumable session:

```bash
clinvk resume --last
```text

### Resume with Follow-up

Resume and immediately send a follow-up prompt:

```bash
clinvk resume --last "continue from where we left off"
```text

### Interactive Picker

Use the interactive picker to select a session:

```bash
clinvk resume --interactive
```bash

If you run `clinvk resume` with no arguments and no `--last`, the interactive picker opens by default.

### Resume from Current Directory

Filter sessions to only those from the current directory:

```bash
clinvk resume --here
```text

### Filter by Backend

Only consider sessions from a specific backend:

```bash
clinvk resume --backend claude
```text

### Resume Specific Session

Resume a specific session by ID:

```bash
clinvk resume abc123
clinvk resume abc123 "now add tests"
```text

### Combine Filters

Combine multiple filters:

```bash
clinvk resume --here --backend claude --last
```text

This resumes the most recent Claude session from the current directory.

## Behavior

The resume command follows this priority:

1. If `--last` is specified, resume the most recent resumable session that matches filters
2. Else if a session ID is provided, resume that session
3. Else open the interactive picker (or error if no resumable sessions exist)

## Output

Resumes the session and displays the AI's response. The same output format options as the root command apply.

### Example Output

```text
Resuming session abc123 (claude)

> continue from where we left off

I've reviewed the changes you made to the auth module. Here's what I found...
```text

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `session not found` | Session ID doesn't exist | Check `clinvk sessions list` for valid IDs |
| `session not resumable` | Session has no backend session ID | Start a new session |
| `backend not available` | Backend CLI not installed | Install the backend |
| `no resumable sessions` | No sessions can be resumed | Start a new session |

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Session not found or error |
| 2 | Backend not available |

## Related Commands

- [sessions](sessions.md) - List and manage sessions
- [prompt](prompt.md) - Execute a new prompt

## See Also

- [Session Management](../../guides/sessions.md) - Guide to session management
- [Configuration Reference](../configuration.md) - Session settings
