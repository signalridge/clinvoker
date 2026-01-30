# Parallel Execution

Run multiple AI tasks concurrently across different backends. Perfect for comprehensive code reviews, security audits, and any scenario where you need multiple perspectives simultaneously.

---

## Overview

```
flowchart LR
    A[clinvk parallel] --> B[Task 1<br/>Claude]
    A --> C[Task 2<br/>Codex]
    A --> D[Task 3<br/>Gemini]
    B --> E[Aggregate Results]
    C --> E
    D --> E
```

**Key Characteristics:**

- **Concurrent execution** - Tasks run in parallel, not sequentially
- **Ephemeral mode** - No sessions persisted (clean state)
- **Configurable concurrency** - Control max parallel workers
- **Fail-fast option** - Stop all tasks on first failure
- **Output aggregation** - Consolidated results table

---

## Basic Usage

```
# From file
clinvk parallel --file tasks.json

# From stdin
cat tasks.json | clinvk parallel

# Using shorthand
clinvk parallel -f tasks.json
```

---

## Task File Format

```
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "Review architecture and design patterns",
      "name": "Architecture Review"
    },
    {
      "backend": "codex",
      "prompt": "Check for performance bottlenecks",
      "name": "Performance Analysis"
    },
    {
      "backend": "gemini",
      "prompt": "Identify security vulnerabilities",
      "name": "Security Audit"
    }
  ],
  "max_parallel": 3,
  "fail_fast": false,
  "aggregate_output": true,
  "output_dir": "./parallel-results"
}
```

---

## Task Fields Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `backend` | string | Backend to use: `claude`, `codex`, `gemini` |
| `prompt` | string | The prompt to send to the backend |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `workdir` | string | - | Working directory for the task |
| `model` | string | - | Model override (e.g., `claude-opus-4-5-20251101`) |
| `approval_mode` | string | `default` | `default`, `auto`, `none`, `always` (best-effort) |
| `sandbox_mode` | string | `default` | `default`, `read-only`, `workspace`, `full` (best-effort) |
| `max_turns` | int | 0 | Maximum agentic turns (Claude only, 0 = unlimited) |
| `max_tokens` | int | 0 | Maximum response tokens (currently not mapped) |
| `system_prompt` | string | - | System prompt (Claude only) |
| `verbose` | bool | false | Enable verbose output |
| `dry_run` | bool | false | Show command without executing |
| `extra` | array | [] | Backend-specific extra flags |
| `id` | string | auto | Task identifier |
| `name` | string | - | Human-readable name for the task |
| `tags` | array | [] | Tags for tracking |
| `meta` | object | {} | Free-form metadata |

---

## Execution Controls

### Command-Line Flags

```
# Limit concurrency (overrides config)
clinvk parallel -f tasks.json --max-parallel 2

# Stop on first failure
clinvk parallel -f tasks.json --fail-fast

# Output as JSON for tooling
clinvk parallel -f tasks.json --json

# Quiet mode (summary only, no task output)
clinvk parallel -f tasks.json --quiet
```

### Configuration Priority

Concurrency settings are resolved in this order:

1. `--max-parallel` CLI flag (highest priority)
2. `max_parallel` in task file
3. `parallel.max_workers` in config file
4. Default = 3 (lowest priority)

### Configuration File Options

```
# ~/.clinvk/config.yaml
parallel:
  max_workers: 5          # Default max parallel tasks
  fail_fast: false        # Stop all on first failure
  aggregate_output: true  # Show summary table
```

---

## Output Options

### Standard Output

By default, parallel shows:

1. Task execution output (if not `--quiet`)
2. Summary table with status, duration, and results

```
Executing 3 tasks with max 3 parallel workers
========================================

[claude] Architecture Review
Analyzing code structure...

[codex] Performance Analysis
Checking algorithmic complexity...

[gemini] Security Audit
Scanning for vulnerabilities...

========================================
PARALLEL EXECUTION SUMMARY
========================================
TASK                BACKEND     STATUS   DURATION   OUTPUT
Architecture Review claude      OK       12.34s     Complete
Performance Analysis codex      OK       8.21s      Complete
Security Audit      gemini      OK       15.67s     Complete
----------------------------------------
Total: 3/3 tasks completed (15.67s)
```

### JSON Output

Use `--json` for programmatic processing:

```
clinvk parallel -f tasks.json --json > results.json
```

```
{
  "total_tasks": 3,
  "completed_tasks": 3,
  "failed_tasks": 0,
  "results": [
    {
      "task_id": "task-1",
      "name": "Architecture Review",
      "backend": "claude",
      "exit_code": 0,
      "content": "The code follows good separation of concerns...",
      "duration_seconds": 12.34
    }
  ]
}
```

### Output Directory

Save individual task outputs to files:

```
{
  "tasks": [...],
  "output_dir": "./parallel-results"
}
```

This creates:

```
./parallel-results/
├── summary.json          # Overall execution summary
├── task-1-claude.md      # Individual task output
├── task-2-codex.md
└── task-3-gemini.md
```

---

## Practical Examples

### Example 1: Multi-Perspective Code Review

```
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "Review this code for architecture and design patterns. Focus on: single responsibility principle, dependency management, and API design.",
      "name": "Architecture Review"
    },
    {
      "backend": "codex",
      "prompt": "Analyze this code for performance issues. Look for: algorithmic complexity, memory leaks, unnecessary allocations, and I/O bottlenecks.",
      "name": "Performance Analysis"
    },
    {
      "backend": "gemini",
      "prompt": "Security audit: check for injection risks, unsafe deserialization, insecure dependencies, and data exposure vulnerabilities.",
      "name": "Security Audit"
    }
  ],
  "max_parallel": 3,
  "fail_fast": false,
  "output_dir": "./code-review-results"
}
```

```
# Run the review with the code as context
clinvk parallel -f review.json < ./src/auth.js
```

**Why this works:** Each backend analyzes from a different perspective simultaneously.

---

### Example 2: Dependency Update Assessment

```
{
  "tasks": [
    {
      "backend": "gemini",
      "prompt": "Review the changelog for express@5.0.0 and identify breaking changes that might affect a typical REST API",
      "name": "Changelog Review"
    },
    {
      "backend": "codex",
      "prompt": "Analyze package.json dependencies and suggest a safe update order to minimize conflicts",
      "name": "Dependency Analysis",
      "workdir": "./"
    },
    {
      "backend": "claude",
      "prompt": "Create a migration plan for updating to the latest versions with rollback strategy",
      "name": "Migration Plan"
    }
  ]
}
```

---

### Example 3: Documentation Generation

```
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "Generate API documentation from the source code",
      "name": "API Docs"
    },
    {
      "backend": "gemini",
      "prompt": "Create a getting started guide for new users",
      "name": "Getting Started"
    },
    {
      "backend": "codex",
      "prompt": "Write code examples in Python, JavaScript, and Go",
      "name": "Code Examples"
    }
  ],
  "output_dir": "./docs-output"
}
```

---

### Example 4: Backend Performance Comparison

```
#!/bin/bash
# benchmark.sh

PROMPT="Refactor this recursive function to iterative"
FILE="fibonacci.py"

echo '{"tasks": [' > tasks.json
first=true
for backend in claude codex gemini; do
  if [ "$first" = true ]; then
    first=false
  else
    echo ',' >> tasks.json
  fi
  echo "{\"backend\": \"$backend\", \"prompt\": \"$PROMPT\", \"name\": \"$backend Benchmark\"}" >> tasks.json
done
echo '], "output_dir": "./benchmark-results"}' >> tasks.json

clinvk parallel -f tasks.json --json | jq -r '
  .results[] |
  "\(.backend): \(.duration_seconds)s - \(.content[:50])..."
'
```

---

### Example 5: Confidence Voting

```
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "Should we migrate from REST to GraphQL? Rate confidence 1-10 and explain your reasoning.",
      "name": "Claude Vote"
    },
    {
      "backend": "codex",
      "prompt": "Should we migrate from REST to GraphQL? Rate confidence 1-10 and explain your reasoning.",
      "name": "Codex Vote"
    },
    {
      "backend": "gemini",
      "prompt": "Should we migrate from REST to GraphQL? Rate confidence 1-10 and explain your reasoning.",
      "name": "Gemini Vote"
    }
  ]
}
```

```
# Extract and analyze confidence scores
clinvk parallel -f vote.json --json | jq -r '
  .results[] |
  "\(.backend): \(.content)"
' | grep -oE '[0-9]+/10' | awk '{sum+=$1; count++} END {
  print "Average confidence:", sum/count "/10"
  if (sum/count >= 7) print "Recommendation: Proceed"
  else if (sum/count >= 4) print "Recommendation: Proceed with caution"
  else print "Recommendation: Do not proceed"
}'
```

---

## Best Practices

### 1. Use Descriptive Names

Always include `name` for better output:

```
{
  "backend": "claude",
  "prompt": "Review code",
  "name": "Architecture Review"  // Clear and descriptive
}
```

### 2. Handle Failures Gracefully

```
{
  "tasks": [...],
  "fail_fast": false  // Get all results even if one fails
}
```

### 3. Use Output Directory for Complex Workflows

```
{
  "tasks": [...],
  "output_dir": "./results"  // Save for later analysis
}
```

### 4. Tag Your Tasks

```
{
  "backend": "claude",
  "prompt": "Review code",
  "tags": ["review", "architecture", "sprint-23"]
}
```

### 5. Set Appropriate Timeouts

For long-running tasks, adjust backend timeouts:

```
# ~/.clinvk/config.yaml
backends:
  claude:
    max_turns: 10  # Limit agentic turns
```

---

## Limitations

1. **Ephemeral Only** - Parallel tasks never persist sessions
2. **No Cross-Task Communication** - Tasks are isolated (use `chain` for context passing)
3. **Resource Limits** - Max concurrency limited by your system and API rate limits
4. **Output Format** - Task-level `output_format` is ignored; use CLI `--json` flag instead

---

## HTTP API Equivalent

The HTTP server supports parallel execution via POST `/api/v1/parallel`:

```
curl -sS http://localhost:8080/api/v1/parallel \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: your-key' \
  -d '{
    "tasks": [
      {"backend": "claude", "prompt": "Review architecture"},
      {"backend": "codex", "prompt": "Check performance"}
    ]
  }'
```

See [REST API Reference](../reference/rest-api.md) for full details.

---

## Related Topics

- [Chain Execution](chain-execution.md) - Sequential workflows with context passing
- [Backend Comparison](backend-comparison.md) - Side-by-side backend responses
- [Use Cases](use-cases.md) - More real-world scenarios
