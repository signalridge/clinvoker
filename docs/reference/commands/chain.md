# clinvk chain

Execute a sequential pipeline of prompts.

## Synopsis

```bash
clinvk chain [flags]
```

## Description

Execute a series of prompts sequentially, passing output from each step to the next via `{{previous}}`. This enables multi-stage workflows where different backends contribute their strengths.

**Note:** CLI chain runs are always ephemeral (no sessions are persisted). `{{session}}`, `pass_session_id`, and `persist_sessions` are not supported and will error.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--file` | `-f` | string | | Pipeline file (JSON) |
| `--json` | | bool | `false` | JSON output |

## Pipeline File Format

```json
{
  "steps": [
    {
      "name": "step-name",
      "backend": "claude",
      "prompt": "First prompt",
      "model": "optional-model"
    },
    {
      "name": "second-step",
      "backend": "gemini",
      "prompt": "Process this: {{previous}}"
    }
  ]
}
```

### Step Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Step identifier |
| `backend` | string | Yes | Backend to use |
| `prompt` | string | Yes | The prompt |
| `model` | string | No | Model override |
| `workdir` | string | No | Working directory |
| `approval_mode` | string | No | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | No | `default`, `read-only`, `workspace`, `full` |
| `max_turns` | int | No | Max agentic turns |

### Top-Level Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `steps` | array | | List of steps (required) |
| `stop_on_failure` | bool | `true` | **CLI always stops on failure** (field is accepted but `false` is ignored) |
| `pass_working_dir` | bool | `false` | Pass working directory between steps |

### Template Variables

| Variable | Description |
|----------|-------------|
| `{{previous}}` | Output text from the previous step |

## Examples

### Basic Chain

```bash
clinvk chain --file pipeline.json
```

### JSON Output

```bash
clinvk chain --file pipeline.json --json
```

## Output

### Text Output

```text
Executing chain with 3 steps
================================================================================

[1/3] analyze (claude)
--------------------------------------------------------------------------------
Analysis result text...

[2/3] recommend (gemini)
--------------------------------------------------------------------------------
Recommendations text...

[3/3] implement (codex)
--------------------------------------------------------------------------------
Implementation text...

================================================================================
CHAIN EXECUTION SUMMARY
================================================================================
STEP   BACKEND      STATUS   DURATION   NAME
--------------------------------------------------------------------------------
1      claude       OK       2.10s      analyze
2      gemini       OK       1.80s      recommend
3      codex        OK       3.20s      implement
--------------------------------------------------------------------------------
Total: 3/3 steps completed (7.10s)
```

### JSON Output

```json
{
  "total_steps": 3,
  "completed_steps": 3,
  "failed_step": 0,
  "total_duration_seconds": 7.1,
  "results": [
    {
      "step": 1,
      "name": "analyze",
      "backend": "claude",
      "output": "Analysis result...",
      "duration_seconds": 2.1,
      "exit_code": 0
    }
  ]
}
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | All steps succeeded |
| 1 | A step failed |

## See Also

- [parallel](parallel.md) - Concurrent execution
- [compare](compare.md) - Backend comparison
