# Chain Execution

Pipeline prompts through multiple backends sequentially, with each step building on the previous output.

## Overview

The `chain` command executes a series of steps in order, passing output from one step to the next. This enables complex multi-stage workflows where different backends contribute their strengths.

## Basic Usage

### Create a Pipeline File

Create a `pipeline.json` file:

```json
{
  "steps": [
    {
      "name": "initial-review",
      "backend": "claude",
      "prompt": "Review this code for bugs"
    },
    {
      "name": "security-check",
      "backend": "gemini",
      "prompt": "Check for security issues in: {{previous}}"
    },
    {
      "name": "final-summary",
      "backend": "codex",
      "prompt": "Summarize the findings: {{previous}}"
    }
  ]
}
```

### Run the Chain

```bash
clinvk chain --file pipeline.json
```

## Template Variables

Use these placeholders in prompts to reference previous outputs:

| Variable | Description |
|----------|-------------|
| `{{previous}}` | Output text from the previous step |

### Example with Variables

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
      "prompt": "Based on this analysis: {{previous}}\n\nRecommend improvements"
    },
    {
      "name": "implement",
      "backend": "codex",
      "prompt": "Implement these recommendations: {{previous}}"
    }
  ]
}
```

## Step Options

Each step can specify various options:

```json
{
  "steps": [
    {
      "name": "step-name",
      "backend": "claude",
      "prompt": "Task description",
      "model": "claude-opus-4-5-20251101",
      "workdir": "/path/to/project",
      "approval_mode": "auto",
      "sandbox_mode": "workspace",
      "max_turns": 10
    }
  ]
}
```

### Step Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Step identifier |
| `backend` | string | Yes | Backend to use |
| `prompt` | string | Yes | The prompt (supports `{{previous}}`) |
| `model` | string | No | Model override |
| `workdir` | string | No | Working directory |
| `approval_mode` | string | No | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | No | `default`, `read-only`, `workspace`, `full` |
| `max_turns` | int | No | Maximum agentic turns |

## Chain Options

Top-level fields for chain execution:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `stop_on_failure` | bool | `true` | CLI always stops on first failure (field is accepted but `false` is ignored) |
| `pass_working_dir` | bool | `false` | Pass working directory between steps |

!!! note "Ephemeral Only"
    Chain always runs in ephemeral mode - no sessions are persisted, and `{{session}}` is not supported.

## Output Options

### JSON Output

Get structured results for programmatic use:

```bash
clinvk chain --file pipeline.json --json
```

Output:

```json
{
  "total_steps": 2,
  "completed_steps": 2,
  "failed_step": 0,
  "total_duration_seconds": 3.5,
  "results": [
    {
      "step": 1,
      "name": "initial-review",
      "backend": "claude",
      "output": "Found several issues...",
      "duration_seconds": 2.0,
      "exit_code": 0
    },
    {
      "step": 2,
      "name": "security-check",
      "backend": "gemini",
      "output": "No critical vulnerabilities...",
      "duration_seconds": 1.5,
      "exit_code": 0
    }
  ]
}
```

## Use Cases

### Code Review Pipeline

```json
{
  "steps": [
    {
      "name": "functionality-review",
      "backend": "claude",
      "prompt": "Review this code for correctness and logic errors"
    },
    {
      "name": "security-review",
      "backend": "gemini",
      "prompt": "Review the code for security vulnerabilities. Previous analysis: {{previous}}"
    },
    {
      "name": "performance-review",
      "backend": "codex",
      "prompt": "Review the code for performance issues. Previous findings: {{previous}}"
    },
    {
      "name": "summary",
      "backend": "claude",
      "prompt": "Create a summary report from all reviews: {{previous}}"
    }
  ]
}
```

### Documentation Generation

```json
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "Analyze the API structure in this codebase"
    },
    {
      "name": "document",
      "backend": "codex",
      "prompt": "Generate API documentation based on: {{previous}}"
    },
    {
      "name": "examples",
      "backend": "gemini",
      "prompt": "Add usage examples to this documentation: {{previous}}"
    }
  ]
}
```

### Iterative Refinement

```json
{
  "steps": [
    {
      "name": "draft",
      "backend": "codex",
      "prompt": "Write a function to parse CSV files"
    },
    {
      "name": "review",
      "backend": "claude",
      "prompt": "Review and suggest improvements: {{previous}}"
    },
    {
      "name": "refine",
      "backend": "codex",
      "prompt": "Apply these improvements: {{previous}}"
    }
  ]
}
```

## Error Handling

If a step fails, the chain stops and reports the error:

```text
Step 1 (analyze): Completed (2.1s)
Step 2 (implement): Failed - Backend error: rate limit exceeded

Chain failed at step 2
```

With `--json`, failed steps include error information:

```json
{
  "results": [
    {"name": "analyze", "exit_code": 0, "error": ""},
    {"name": "implement", "exit_code": 1, "error": "rate limit exceeded"}
  ]
}
```

## Tips

!!! tip "Use Descriptive Step Names"
    Good step names make the output easier to understand and debug.

!!! tip "Start Simple"
    Begin with 2-3 steps and add more as needed. Complex chains can be harder to debug.

!!! tip "Consider Context Length"
    When using `{{previous}}`, be mindful that output from earlier steps adds to the prompt length.

!!! tip "Use Different Backends"
    Leverage each backend's strengths - Claude for reasoning, Codex for code generation, Gemini for broad knowledge.

## Comparison with Parallel

| Feature | Chain | Parallel |
|---------|-------|----------|
| Execution | Sequential | Concurrent |
| Data flow | Previous output available | Independent |
| Use case | Multi-stage workflows | Independent tasks |
| Speed | Sum of step times | Max step time |

## Next Steps

- [Parallel Execution](parallel-execution.md) - Run independent tasks concurrently
- [Backend Comparison](backend-comparison.md) - Compare responses side-by-side
