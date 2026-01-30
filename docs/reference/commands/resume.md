# clinvk resume

Resume a previous session.

## Synopsis

```bash
clinvk resume [session-id] [prompt]
```

## Description

Resumes a previously saved session. A session must have a backend session ID to be resumable.

If no session ID is provided, clinvk opens an interactive picker.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--last` | | bool | `false` | Resume the most recent session |
| `--backend` | `-b` | string | | Filter sessions by backend |
| `--here` | | bool | `false` | Filter by current working directory |
| `--interactive` | `-i` | bool | `false` | Interactive session picker |
| `--output-format` | `-o` | string | config / `json` | `text`, `json`, `stream-json` |

## Examples

```bash
# Resume by prefix or full ID
clinvk resume abc123

# Resume last session
clinvk resume --last

# Interactive picker
clinvk resume --interactive

# Resume with a prompt
clinvk resume --last "continue from here"
```

## Notes

- Sessions without backend session IDs cannot be resumed.
- If `--last` is used, the most recent **resumable** session is chosen.
