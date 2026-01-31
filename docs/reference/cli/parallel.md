# clinvk parallel

Execute multiple tasks in parallel.

## Synopsis

```bash
clinvk parallel [flags]
```text

## Description

Run multiple AI tasks concurrently. Tasks are defined in a JSON file or piped via stdin.

**Note:** CLI parallel runs are always ephemeral (no sessions are persisted), and task-level `output_format` is currently ignored because the command forces JSON internally for reliable parsing.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--file` | `-f` | string | | Task file (JSON) |
| `--max-parallel` | | int | `3` | Max concurrent tasks (overrides config) |
| `--fail-fast` | | bool | `false` | Stop on first failure |
| `--json` | | bool | `false` | JSON output |
| `--quiet` | `-q` | bool | `false` | Suppress task output |

## Task File Format

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "task prompt",
      "model": "optional-model",
      "workdir": "/optional/path",
      "approval_mode": "auto",
      "sandbox_mode": "workspace",
      "max_turns": 10
    }
  ],
  "max_parallel": 3,
  "fail_fast": true
}
```text

### Task Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `backend` | string | Yes | Backend to use |
| `prompt` | string | Yes | The prompt |
| `model` | string | No | Model override |
| `workdir` | string | No | Working directory |
| `approval_mode` | string | No | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | No | `default`, `read-only`, `workspace`, `full` |
| `output_format` | string | No | Accepted but **ignored** in CLI parallel (reserved) |
| `max_tokens` | int | No | Max response tokens (not mapped to backend flags yet) |
| `max_turns` | int | No | Max agentic turns |
| `system_prompt` | string | No | System prompt |
| `extra` | array | No | Extra backend-specific flags |
| `verbose` | bool | No | Enable verbose output |
| `dry_run` | bool | No | Simulate execution |
| `id` | string | No | Task identifier |
| `name` | string | No | Task display name |
| `tags` | array | No | Tags copied into JSON output / `output_dir` artifacts |
| `meta` | object | No | Custom metadata copied into JSON output / `output_dir` artifacts |

### Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `tasks` | array | List of tasks |
| `max_parallel` | int | Max concurrent tasks |
| `fail_fast` | bool | Stop on first failure |
| `output_dir` | string | Optional directory to persist `summary.json` and per-task JSON outputs |

## Examples

### From File

```bash
clinvk parallel --file tasks.json
```text

### From Stdin

```bash
cat tasks.json | clinvk parallel
```text

### Limit Workers

```bash
clinvk parallel --file tasks.json --max-parallel 2
```text

### Fail-Fast Mode

```bash
clinvk parallel --file tasks.json --fail-fast
```text

### JSON Output

```bash
clinvk parallel --file tasks.json --json
```text

### Persist Outputs

```bash
cat tasks.json | jq '. + {"output_dir": "parallel_runs/run-001"}' | clinvk parallel
```text

This writes:

- `summary.json` (aggregate results)
- One JSON file per task (includes `task` + `result`)

### Quiet Mode

```bash
clinvk parallel --file tasks.json --quiet
```text

## Output

### Text Output

```text
Running 3 tasks (max 3 parallel)...

[1] The auth module looks good...
[2] Added logging statements...
[3] Generated 5 test cases...

Results:
--------------------------------------------------------------------------------
#    BACKEND      STATUS   DURATION   TASK
--------------------------------------------------------------------------------
1    claude       OK       2.50s      review the auth module
2    codex        OK       3.20s      add logging to the API
3    gemini       OK       2.80s      generate tests for utils
--------------------------------------------------------------------------------
Total: 3 tasks, 3 completed, 0 failed (3.20s)
```text

### JSON Output

```json
{
  "total_tasks": 3,
  "completed": 3,
  "failed": 0,
  "total_duration_seconds": 3.2,
  "results": [
    {
      "index": 0,
      "task_id": "task-1",
      "task_name": "Auth Review",
      "backend": "claude",
      "output": "The auth module looks good...",
      "duration_seconds": 2.5,
      "exit_code": 0
    }
  ]
}
```text

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | All tasks succeeded |
| 1 | One or more tasks failed |

## See Also

- [chain](chain.md) - Sequential execution
- [compare](compare.md) - Backend comparison
