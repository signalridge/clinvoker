# clinvk [prompt]

Execute a prompt with an AI backend.

## Synopsis

```bash
clinvk [flags] [prompt]
```bash

## Description

The root command executes a prompt using the configured backend. It supports session persistence, output formatting, and auto-resume behavior.

This is the default command - when you run `clinvk` followed by text, it executes as a prompt.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | AI backend to use (`claude`, `codex`, `gemini`) |
| `--model` | `-m` | string | | Model override for the selected backend |
| `--workdir` | `-w` | string | | Working directory passed to the backend |
| `--output-format` | `-o` | string | `json` | Output format: `text`, `json`, `stream-json` |
| `--continue` | `-c` | bool | `false` | Continue the most recent resumable session |
| `--dry-run` | | bool | `false` | Print the backend command without executing |
| `--ephemeral` | | bool | `false` | Stateless mode: do not persist a session |
| `--config` | | string | `~/.clinvk/config.yaml` | Custom config file path |

## Examples

### Basic Usage

Execute a simple prompt:

```bash
clinvk "fix the bug in auth.go"
```text

### Specify Backend

Use a specific backend:

```bash
clinvk --backend codex "implement user registration"
clinvk -b gemini "explain this algorithm"
```text

### Specify Model

Override the default model:

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "quick review"
clinvk -b codex -m o3-mini "simple task"
```text

### Continue Session

Continue from a previous session:

```bash
# Start a session
clinvk "implement the login feature"

# Continue the session
clinvk -c "now add password validation"

# Continue again
clinvk -c "add rate limiting"
```text

### JSON Output

Get structured JSON output:

```bash
clinvk --output-format json "explain this code"
```text

### Dry Run

See what command would be executed:

```bash
clinvk --dry-run "implement feature X"
# Output: Would execute: claude --model claude-opus-4-5-20251101 "implement feature X"
```text

### Ephemeral Mode

Run without creating a session:

```bash
clinvk --ephemeral "what is 2+2"
```text

### Set Working Directory

Specify the working directory:

```bash
clinvk --workdir /path/to/project "review the codebase"
```text

## Output

### Text Format

When `--output-format text` is used, only the response text is printed:

```text
The code implements a binary search algorithm...
```text

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
```text

### Stream JSON Format

`stream-json` passes through the backend's native streaming format (NDJSON/JSONL). The event shape depends on the backend CLI and is not unified.

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `backend not found` | Backend CLI not installed | Install the backend (e.g., `npm install -g @anthropic-ai/claude-code`) |
| `session not resumable` | Session doesn't support resuming | Start a new session |
| `timeout` | Command took too long | Increase `command_timeout_secs` in config |
| `invalid output format` | Unknown format specified | Use `text`, `json`, or `stream-json` |

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error |
| 2+ | Backend exit code (propagated when the backend process exits non-zero) |

See [Exit Codes](../exit-codes.md) for details.

## Related Commands

- [resume](resume.md) - Resume a session
- [sessions](sessions.md) - Manage sessions
- [config](config.md) - Configure defaults

## See Also

- [Configuration Reference](../configuration.md) - Configure defaults
- [Environment Variables](../environment.md) - Environment-based settings
