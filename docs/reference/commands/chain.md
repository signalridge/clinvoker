# clinvk chain

Execute a sequential pipeline of prompts.

## Synopsis

```bash
clinvk chain [flags]
```

## Description

Execute a series of prompts sequentially, passing output from each step to the next. This enables multi-stage workflows where different backends contribute their strengths.

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
| `stop_on_failure` | bool | `true` | Stop the chain on first failure |
| `pass_working_dir` | bool | `false` | Pass working directory between steps |

### Template Variables

| Variable | Description |
|----------|-------------|
| `{{previous}}` | Output text from the previous step |

!!! note "Ephemeral Only"
    Chain always runs in ephemeral mode - no sessions are persisted, and `{{session}}` is not supported.

## Examples

### Basic Chain

```bash
clinvk chain --file pipeline.json
```

### JSON Output

```bash
clinvk chain --file pipeline.json --json
```

### Example Pipeline

```json
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "Analyze this codebase structure"
    },
    {
      "name": "recommend",
      "backend": "gemini",
      "prompt": "Based on this analysis, recommend improvements: {{previous}}"
    },
    {
      "name": "implement",
      "backend": "codex",
      "prompt": "Implement these recommendations: {{previous}}"
    }
  ]
}
```

## Output

### Text Output (Default)

```yaml
Step 1 (analyze): Starting...
Step 1 (analyze): Completed (2.1s)

Analysis result text...

Step 2 (recommend): Starting...
Step 2 (recommend): Completed (1.8s)

Recommendations text...

Step 3 (implement): Starting...
Step 3 (implement): Completed (3.2s)

Implementation text...

Chain completed successfully
Total time: 7.1s
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
    },
    {
      "step": 2,
      "name": "recommend",
      "backend": "gemini",
      "output": "Recommendations...",
      "duration_seconds": 1.8,
      "exit_code": 0
    },
    {
      "step": 3,
      "name": "implement",
      "backend": "codex",
      "output": "Implementation...",
      "duration_seconds": 3.2,
      "exit_code": 0
    }
  ]
}
```

## Error Handling

If a step fails, the chain stops:

```text
Step 1 (analyze): Completed (2.1s)
Step 2 (implement): Failed - Backend error

Chain failed at step 2
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | All steps succeeded |
| 1 | A step failed |

## See Also

- [parallel](parallel.md) - Concurrent execution
- [compare](compare.md) - Backend comparison
