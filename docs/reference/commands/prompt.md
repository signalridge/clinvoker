# clinvk [prompt]

Execute a prompt with an AI backend.

## Synopsis

```
clinvk [flags] [prompt]
```

## Description

The root command executes a prompt using the configured AI backend. This is the primary way to interact with clinvk.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | AI backend to use |
| `--model` | `-m` | string | | Model to use |
| `--workdir` | `-w` | string | cwd | Working directory |
| `--output-format` | `-o` | string | `text` | Output format |
| `--continue` | `-c` | bool | `false` | Continue last session |
| `--dry-run` | | bool | `false` | Show command only |
| `--ephemeral` | | bool | `false` | Stateless mode |

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

### Text Format (Default)

The AI's response is printed to stdout.

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
  }
}
```

### Stream JSON Format

Stream JSON passes through the backend's native streaming format (NDJSON/JSONL).
The exact event shape depends on the backend.

```json
{"type": "start", "backend": "claude"}
{"type": "content", "text": "chunk of response"}
{"type": "end", "session_id": "abc123"}
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 126 | Backend not available |
| 127 | Backend not found |

## See Also

- [resume](resume.md) - Resume a session
- [Configuration](../configuration.md) - Configure defaults
