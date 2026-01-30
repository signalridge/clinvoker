# Basic Usage

The daily workflow for `clinvk`.

## Run a prompt

```bash
clinvk "explain this code"
```

If no backend is specified, clinvk uses:

1. `default_backend` from config
2. fallback to `claude`

## Choose a backend

```bash
clinvk -b claude "review this module"
clinvk -b codex "optimize this function"
clinvk -b gemini "summarize this document"
```

## Pick a model

```bash
clinvk -b claude -m claude-opus-4-5-20251101 "deep review"
clinvk -b codex -m o3 "refactor this"
```

## Working directory

```bash
clinvk --workdir /path/to/project "scan for TODOs"
```

If `--workdir` is omitted, the current directory is used.

## Output formats

Output format can be set per command or via config (`output.format`).

### Text

```bash
clinvk --output-format text "quick summary"
```

### JSON (default)

```bash
clinvk --output-format json "quick summary"
```

### Stream JSON (CLI)

```bash
clinvk --output-format stream-json "stream output"
```

Notes:

- CLI streaming passes through the **backend’s native stream format**.
- For server streaming, see [HTTP Server](http-server.md) (unified events).

## Continue the last session

```bash
clinvk "draft a migration plan"
clinvk -c "add rollback strategy"
```

`-c/--continue` resumes the **most recent resumable session**. If no resumable session exists, a new one is created.

## Dry run

```bash
clinvk --dry-run "generate release notes"
```

Dry run prints the exact backend command without executing it.

## Ephemeral mode (stateless)

```bash
clinvk --ephemeral "one-off question"
```

No session is created. For backends that do not support native stateless mode, clinvk attempts cleanup after execution.

## Output extras

Text output can optionally include token usage and timing:

```yaml
output:
  show_tokens: true
  show_timing: true
```

(Only applies to text output.)

## Next steps

- [Session Management](session-management.md)
- [Parallel Execution](parallel-execution.md)
- [Backend Comparison](backend-comparison.md)
