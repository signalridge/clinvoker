# clinvk [prompt]

Execute a prompt with the selected backend. This is the root command.

## Synopsis

```bash
clinvk [flags] [prompt]
```

## Description

Runs a single prompt on the chosen backend. By default, a session is created and saved (unless `--ephemeral` is set).

If `session.auto_resume` is enabled and a resumable session exists, clinvk will continue it when appropriate.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | config / `claude` | Backend to use |
| `--model` | `-m` | string | | Model override |
| `--workdir` | `-w` | string | current dir | Working directory |
| `--output-format` | `-o` | string | config / `json` | `text`, `json`, `stream-json` |
| `--continue` | `-c` | bool | `false` | Continue most recent session |
| `--dry-run` | | bool | `false` | Print command without executing |
| `--ephemeral` | | bool | `false` | Stateless mode (no session) |
| `--config` | | string | `~/.clinvk/config.yaml` | Config file path |

## Examples

```bash
clinvk "fix the bug in auth.go"
clinvk -b codex "optimize this function"
clinvk -b gemini -m gemini-2.5-pro "summarize this"
clinvk -c "continue from last session"
```

## Output

### JSON (default)

```json
{
  "backend": "claude",
  "content": "...",
  "session_id": "abc123...",
  "model": "claude-opus-4-5-20251101",
  "duration_seconds": 2.5,
  "exit_code": 0,
  "error": "",
  "usage": {
    "input_tokens": 123,
    "output_tokens": 456,
    "total_tokens": 579
  },
  "raw": {}
}
```

### Text

```bash
clinvk --output-format text "explain this"
```

### Stream JSON (CLI)

```bash
clinvk --output-format stream-json "stream output"
```

Notes:

- CLI streaming passes through backend‑native stream lines.
- Server streaming uses unified events (see HTTP Server docs).

## Exit codes

- `0` on success
- `1` on errors
- Backend exit code is propagated if non‑zero
- `124` when a command timeout occurs (if configured)
