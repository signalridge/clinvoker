# clinvk [prompt]

Execute a prompt with an AI backend.

## Synopsis

```bash
clinvk [flags] [prompt]
```

## Description

The root command executes a prompt using the configured backend. It also supports session persistence, output formatting, and auto-resume behavior.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | AI backend to use (`claude`, `codex`, `gemini`) |
| `--model` | `-m` | string | | Model override for the selected backend |
| `--workdir` | `-w` | string | | Working directory passed to the backend |
| `--output-format` | `-o` | string | `json` | Output format: `text`, `json`, `stream-json` (config `output.format` overrides when flag not set) |
| `--continue` | `-c` | bool | `false` | Continue the most recent resumable session |
| `--dry-run` | | bool | `false` | Print the backend command without executing |
| `--ephemeral` | | bool | `false` | Stateless mode: do not persist a session |
| `--config` | | string | `~/.clinvk/config.yaml` | Custom config file path |

## Examples

### Basic Usage

```bash
clinvk "fix the bug in auth.go"
```

### Specify Backend

```bash
clinvk --backend codex "implement user registration"
clinvk -b gemini "explain this algorithm"
```

### Specify Model

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "quick review"
```

### Continue Session

```bash
clinvk "implement the login feature"
clinvk -c "now add password validation"
clinvk -c "add rate limiting"
```

### JSON Output

```bash
clinvk --output-format json "explain this code"
```

### Dry Run

```bash
clinvk --dry-run "implement feature X"
# Output: Would execute: claude --model claude-opus-4-5-20251101 "implement feature X"
```

### Ephemeral Mode

```bash
clinvk --ephemeral "what is 2+2"
```

### Set Working Directory

```bash
clinvk --workdir /path/to/project "review the codebase"
```

## Output

### Text Format

When `--output-format text` is used, only the response text is printed.

### JSON Format

```json
{
  "backend": "claude",
  "content": "The response text...",
  "session_id": "abc123",
  "model": "claude-opus-4-5-20251101",
  "duration_seconds": 2.5,
  "exit_code": 0,
  "usage": {
    "input_tokens": 123,
    "output_tokens": 456,
    "total_tokens": 579
  },
  "raw": {
    "events": []
  }
}
```

### Stream JSON Format

`stream-json` passes through the backend's native streaming format (NDJSON/JSONL). The event shape depends on the backend CLI and is not unified.

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error |
| (backend) | Backend exit code (propagated when the backend process exits non-zero) |

See [Exit Codes](../exit-codes.md) for details.

## See Also

- [resume](resume.md) - Resume a session
- [Configuration](../configuration.md) - Configure defaults
