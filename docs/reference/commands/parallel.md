# clinvk parallel

Run multiple tasks in parallel from a JSON file or stdin.

## Synopsis

```bash
clinvk parallel --file tasks.json [--max-parallel N] [--fail-fast] [--json] [--quiet]
cat tasks.json | clinvk parallel
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Task file path (JSON) |
| `--max-parallel` | | 3 | Max concurrent tasks |
| `--fail-fast` | | false | Stop all tasks on first failure |
| `--json` | | false | Emit JSON summary |
| `--quiet` | `-q` | false | Suppress per‑task output |

## Input format

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "Review auth module"}
  ],
  "max_parallel": 3,
  "fail_fast": false,
  "output_dir": "./parallel-results"
}
```

### Task fields (CLI)

- `backend` (required)
- `prompt` (required)
- `workdir`, `model`, `approval_mode`, `sandbox_mode`, `max_turns`, `max_tokens`, `system_prompt`, `verbose`, `dry_run`, `extra`
- `id`, `name`, `tags`, `meta`

Notes:

- Parallel execution is **always ephemeral** (no sessions persisted).
- `output_format` in tasks is currently ignored in CLI mode.
- `parallel.aggregate_output=false` suppresses the summary table.

## Output files

If `output_dir` is set, clinvk writes `summary.json` and one file per task.

## Exit codes

- `0` all tasks succeeded
- `1` one or more tasks failed
