# Parallel Execution

Run multiple AI tasks concurrently to save time and increase productivity.

## Overview

The `parallel` command executes multiple tasks simultaneously across one or more backends. This is useful for:

- Running independent tasks faster
- Getting multiple perspectives simultaneously
- Batch processing

## Basic Usage

### Create a Task File

Create a `tasks.json` file:

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "review the auth module"
    },
    {
      "backend": "codex",
      "prompt": "add logging to the API"
    },
    {
      "backend": "gemini",
      "prompt": "generate tests for utils"
    }
  ]
}
```bash

### Run Tasks

```bash
clinvk parallel --file tasks.json
```

### From Stdin

You can also pipe task definitions:

```bash
cat tasks.json | clinvk parallel
```text

## Task Options

Each task can specify various options:

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "review the code",
      "model": "claude-opus-4-5-20251101",
      "workdir": "/path/to/project",
      "approval_mode": "auto",
      "sandbox_mode": "workspace",
      "output_format": "json",
      "max_tokens": 4096,
      "max_turns": 10,
      "system_prompt": "You are a code reviewer."
    }
  ]
}
```

### Task Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `backend` | string | Yes | Backend to use (claude, codex, gemini) |
| `prompt` | string | Yes | The prompt to execute |
| `model` | string | No | Model override |
| `workdir` | string | No | Working directory |
| `approval_mode` | string | No | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | No | `default`, `read-only`, `workspace`, `full` |
| `output_format` | string | No | `text`, `json`, `stream-json` |
| `max_tokens` | int | No | Maximum response tokens |
| `max_turns` | int | No | Maximum agentic turns |
| `system_prompt` | string | No | Custom system prompt |

## Execution Options

### Limit Parallel Workers

Control how many tasks run simultaneously:

```bash
# Run at most 2 tasks at a time
clinvk parallel --file tasks.json --max-parallel 2
```bash

### Fail-Fast Mode

Stop all tasks on the first failure:

```bash
clinvk parallel --file tasks.json --fail-fast
```

### JSON Output

Get structured output for programmatic processing:

```bash
clinvk parallel --file tasks.json --json
```bash

### Quiet Mode

Suppress task output, show only summary:

```bash
clinvk parallel --file tasks.json --quiet
```

## Top-Level Options

You can specify options at the file level that apply to all tasks:

```json
{
  "tasks": [...],
  "max_parallel": 3,
  "fail_fast": true
}
```

CLI flags override file-level settings.

## Output Format

### Text Output (Default)

Shows progress and results as tasks complete:

```text
Running 3 tasks (max 3 parallel)...

[1] The auth module looks good...
[2] Added logging statements...
[3] Generated 5 test cases...

Results
============================================================

BACKEND      STATUS   DURATION   SESSION    TASK

------------------------------------------------------------

1    claude       OK       2.50s      abc123     review the auth module
2    codex        OK       3.20s      def456     add logging to the API
3    gemini       OK       2.80s      ghi789     generate tests for utils
------------------------------------------------------------

Total: 3 tasks, 3 completed, 0 failed (3.20s)

```

### JSON Output

```json
{
  "total_tasks": 3,
  "completed": 3,
  "failed": 0,
  "total_duration_seconds": 3.2,
  "results": [
    {
      "backend": "claude",
      "output": "The auth module looks good...",
      "duration_seconds": 2.5,
      "exit_code": 0,
      "session_id": "abc123"
    },
    {
      "backend": "codex",
      "output": "Added logging statements...",
      "duration_seconds": 3.2,
      "exit_code": 0,
      "session_id": "def456"
    }
  ]
}
```

## Examples

### Code Review from Multiple Perspectives

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "review this code for architectural issues"
    },
    {
      "backend": "gemini",
      "prompt": "review this code for security vulnerabilities"
    },
    {
      "backend": "codex",
      "prompt": "review this code for performance issues"
    }
  ]
}
```text

### Batch Test Generation

```json
{
  "tasks": [
    {"backend": "codex", "prompt": "generate unit tests for auth.go"},
    {"backend": "codex", "prompt": "generate unit tests for user.go"},
    {"backend": "codex", "prompt": "generate unit tests for api.go"}
  ],
  "max_parallel": 3
}
```

### Multi-Project Tasks

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "update dependencies",
      "workdir": "/projects/frontend"
    },
    {
      "backend": "claude",
      "prompt": "update dependencies",
      "workdir": "/projects/backend"
    }
  ]
}
```text

## Configuration

Default parallel execution settings in `~/.clinvk/config.yaml`:

```yaml
parallel:
  # Maximum concurrent tasks
  max_workers: 3

  # Stop on first failure
  fail_fast: false

  # Combine output from all tasks
  aggregate_output: true
```

## Tips

!!! tip "Use Fail-Fast for Dependencies"
    If later tasks depend on earlier ones succeeding, use `--fail-fast` to stop immediately on failure.

!!! tip "Balance Parallelism"
    Running too many tasks in parallel may hit rate limits or resource constraints. Start with 2-3 workers.

!!! tip "Use JSON for Scripting"
    When automating workflows, use `--json` output for reliable parsing.

## Next Steps

- [Chain Execution](chain-execution.md) - Sequential pipelines
- [Backend Comparison](backend-comparison.md) - Compare responses
