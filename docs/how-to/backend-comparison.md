# Backend Comparison

Compare responses from multiple AI backends side-by-side to get different perspectives on the same problem.

## Overview

The `compare` command runs the same prompt against multiple backends simultaneously and displays their responses together. This is useful for:

- Getting diverse perspectives
- Evaluating backend strengths
- Making informed decisions
- Learning how different AI models approach problems

## Basic Usage

### Compare All Backends

Run against all enabled backends:

```bash
clinvk compare --all-backends "explain this algorithm"
```

### Compare Specific Backends

Select which backends to compare:

```bash
clinvk compare --backends claude,codex "what does this code do"
clinvk compare --backends claude,gemini "review this PR"
```

## Execution Modes

### Parallel (Default)

Run all backends simultaneously:

```bash
clinvk compare --all-backends "explain this code"
```

### Sequential

Run backends one at a time:

```bash
clinvk compare --all-backends --sequential "review this implementation"
```

Sequential mode is useful when:

- You want to avoid rate limits
- System resources are constrained
- You prefer watching responses come in one by one

## Output Formats

### Text Output (Default)

Displays each backend's response with clear separation:

```yaml
Prompt: explain this algorithm

━━━ claude ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: claude-opus-4-5-20251101
Duration: 2.5s

This algorithm implements a binary search...

━━━ codex ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: o3
Duration: 3.2s

The algorithm performs a binary search...

━━━ gemini ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: gemini-2.5-pro
Duration: 2.8s

This is a classic binary search implementation...
```

### JSON Output

Get structured data for programmatic processing:

```bash
clinvk compare --all-backends --json "explain this code"
```

Output:

```json
{
  "prompt": "explain this code",
  "backends": ["claude", "codex", "gemini"],
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "output": "This algorithm implements a binary search...",
      "duration_seconds": 2.5,
      "exit_code": 0
    },
    {
      "backend": "codex",
      "model": "o3",
      "output": "The algorithm performs a binary search...",
      "duration_seconds": 3.2,
      "exit_code": 0
    },
    {
      "backend": "gemini",
      "model": "gemini-2.5-pro",
      "output": "This is a classic binary search implementation...",
      "duration_seconds": 2.8,
      "exit_code": 0
    }
  ],
  "total_duration_seconds": 3.2
}
```

## Use Cases

### Code Review

Get multiple perspectives on code quality:

```bash
clinvk compare --all-backends "review this code for bugs and improvements"
```

### Architecture Decisions

Compare recommendations for design choices:

```bash
clinvk compare --backends claude,gemini "what's the best way to implement caching here?"
```

### Learning

See how different AI models explain concepts:

```bash
clinvk compare --all-backends "explain how async/await works in JavaScript"
```

### Validation

Cross-check important decisions:

```bash
clinvk compare --all-backends "is this implementation secure?"
```

## Handling Failures

If a backend fails, the comparison continues with remaining backends:

```yaml
Comparing 3 backends: claude, codex, gemini
Prompt: explain this code
============================================================
[claude] Response content...
[codex] Error: Backend unavailable
[gemini] Response content...

============================================================
COMPARISON SUMMARY
============================================================
BACKEND      STATUS     DURATION     MODEL
------------------------------------------------------------
claude       OK         2.50s        claude-opus-4-5-20251101
codex        FAILED     0.50s        o3
             Error: Backend unavailable
gemini       OK         2.80s        gemini-2.5-pro
------------------------------------------------------------
Total time: 2.80s
```

## Command Options

| Flag | Description | Default |
|------|-------------|---------|
| `--backends` | Comma-separated backend list | - |
| `--all-backends` | Compare all registered backends (skips unavailable) | `false` |
| `--sequential` | Run one at a time | `false` |
| `--json` | JSON output | `false` |

## Configuration

`backends.<name>.enabled` is stored in config but is not currently enforced by `compare`.

## Tips

!!! tip "Use for Important Decisions"
    For critical code changes or architecture decisions, comparing multiple backends can reveal issues one might miss.

!!! tip "Note the Differences"
    Pay attention to where backends agree (high confidence) and where they differ (worth investigating).

!!! tip "Consider Response Time"
    JSON output includes duration, useful for benchmarking backend performance.

!!! tip "Combine with Other Features"
    Use comparison results to inform which backend to use for follow-up work.

## Comparison vs. Other Commands

| Command | Use Case |
|---------|----------|
| `compare` | Same prompt, different backends, side-by-side |
| `parallel` | Different prompts, any backends, concurrent |
| `chain` | Sequential pipeline, output flows between steps |

## Next Steps

- [Parallel Execution](parallel-execution.md) - Run independent tasks concurrently
- [Chain Execution](chain-execution.md) - Sequential multi-backend pipelines
- [Backends Guide](backends/index.md) - Learn about each backend
