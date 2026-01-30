# Chain Execution

Run AI backends sequentially, passing context between steps. Perfect for multi-stage workflows like analysis → fix → verify → document.

---

## Overview

```
flowchart LR
    A[Step 1<br/>Claude] --"{{previous}}"--> B[Step 2<br/>Codex]
    B --"{{previous}}"--> C[Step 3<br/>Gemini]
    C --"{{previous}}"--> D[Output]
```

**Key Characteristics:**

- **Sequential execution** - Steps run one after another
- **Context passing** - Previous output flows to next step via `{{previous}}`
- **Ephemeral mode** - No sessions persisted (clean state)
- **Failure handling** - Stop or continue on step failure
- **Working directory inheritance** - Optionally pass workdir between steps

---

## Basic Usage

```
# From file
clinvk chain --file chain.json

# From stdin
cat chain.json | clinvk chain

# Using shorthand
clinvk chain -f chain.json

# JSON output for programmatic use
clinvk chain -f chain.json --json
```

---

## Chain Format

```
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "Analyze this code and identify issues",
      "max_turns": 5
    },
    {
      "name": "fix",
      "backend": "codex",
      "prompt": "Fix these issues: {{previous}}",
      "approval_mode": "auto"
    },
    {
      "name": "verify",
      "backend": "gemini",
      "prompt": "Verify the fixes and write tests: {{previous}}"
    }
  ],
  "stop_on_failure": true,
  "pass_working_dir": true
}
```

---

## Step Fields Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `backend` | string | Backend to use: `claude`, `codex`, `gemini` |
| `prompt` | string | The prompt to send (can include `{{previous}}`) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | `Step N` | Human-readable step name |
| `model` | string | - | Model override |
| `workdir` | string | - | Working directory for this step |
| `approval_mode` | string | `default` | `default`, `auto`, `none`, `always` (best-effort) |
| `sandbox_mode` | string | `default` | `default`, `read-only`, `workspace`, `full` (best-effort) |
| `max_turns` | int | 0 | Maximum agentic turns (Claude only, 0 = unlimited) |

---

## Placeholders

### `{{previous}}`

The text output from the previous step is substituted wherever `{{previous}}` appears in the prompt.

```
{
  "steps": [
    {"backend": "claude", "prompt": "Analyze this code"},
    {"backend": "codex", "prompt": "Fix the issues identified: {{previous}}"}
  ]
}
```

**Example flow:**

1. Step 1 (Claude) outputs: "Found 3 issues: null pointer, race condition, memory leak"
2. Step 2 (Codex) receives: "Fix the issues identified: Found 3 issues: null pointer..."

### Limitations

- Only `{{previous}}` is supported (no other placeholders)
- The placeholder is replaced with raw text content
- Cannot access earlier steps (only immediate predecessor)

---

## Top-Level Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `stop_on_failure` | bool | `true` | Stop chain if a step fails |
| `pass_working_dir` | bool | `false` | Inherit workdir from previous step |

### Stop on Failure

```
{
  "steps": [...],
  "stop_on_failure": true
}
```

- `true` (default) - Chain stops at first failing step
- `false` - Continue with subsequent steps even if one fails

### Pass Working Directory

```
{
  "steps": [
    {"backend": "claude", "prompt": "Analyze", "workdir": "./src"},
    {"backend": "codex", "prompt": "Fix: {{previous}}"}  // Inherits ./src
  ],
  "pass_working_dir": true
}
```

When enabled:
- Step 2 inherits Step 1's `workdir` if not explicitly set
- Step 3 inherits Step 2's `workdir` (which may be inherited from Step 1)
- Explicit `workdir` always takes precedence

---

## Output Formats

### Standard Output (Default)

Shows step-by-step progress with clear headers:

```
Executing chain with 3 steps
============================================

[1/3] analyze (claude)
--------------------------------------------
<Step output here>

[2/3] fix (codex)
--------------------------------------------
<Step output here>

[3/3] verify (gemini)
--------------------------------------------
<Step output here>

============================================
CHAIN EXECUTION SUMMARY
============================================
STEP  BACKEND     STATUS   DURATION   NAME
1     claude      OK       8.45s      analyze
2     codex       OK       12.30s     fix
3     gemini      OK       6.12s      verify
--------------------------------------------
Total: 3/3 steps completed (26.87s)
```

### JSON Output

Use `--json` for programmatic processing:

```
clinvk chain -f chain.json --json > results.json
```

```
{
  "total_steps": 3,
  "completed_steps": 3,
  "failed_step": 0,
  "results": [
    {
      "step": 1,
      "name": "analyze",
      "backend": "claude",
      "exit_code": 0,
      "output": "Analysis results...",
      "duration_seconds": 8.45,
      "start_time": "2026-01-31T10:00:00Z",
      "end_time": "2026-01-31T10:00:08Z"
    }
  ],
  "total_duration_seconds": 26.87,
  "start_time": "2026-01-31T10:00:00Z",
  "end_time": "2026-01-31T10:00:27Z"
}
```

---

## Practical Examples

### Example 1: Bug Fix Pipeline

Complete workflow from analysis to documentation.

```
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "Analyze the following error logs and code to find the root cause. Explain what's happening and why in detail.",
      "max_turns": 5
    },
    {
      "name": "fix",
      "backend": "codex",
      "prompt": "Based on this analysis, implement the minimal fix required. Do not over-engineer: {{previous}}",
      "approval_mode": "auto"
    },
    {
      "name": "test",
      "backend": "gemini",
      "prompt": "Write comprehensive unit tests to verify this fix and prevent regression. Include edge cases: {{previous}}"
    },
    {
      "name": "document",
      "backend": "claude",
      "prompt": "Document the fix in CHANGELOG format, explaining what was fixed, why, and the impact: {{previous}}"
    }
  ],
  "stop_on_failure": true,
  "pass_working_dir": true
}
```

```
# Run with error logs as context
clinvk chain -f bugfix.json < error.log
```

---

### Example 2: Feature Implementation

End-to-end feature development.

```
{
  "steps": [
    {
      "name": "design-api",
      "backend": "claude",
      "prompt": "Design the REST API schema for a user notification system with email, push, and in-app notifications. Include endpoints, request/response formats, and authentication."
    },
    {
      "name": "implement-backend",
      "backend": "codex",
      "prompt": "Implement the backend service in Node.js/Express based on this API design: {{previous}}",
      "workdir": "./backend",
      "approval_mode": "auto"
    },
    {
      "name": "implement-frontend",
      "backend": "gemini",
      "prompt": "Create React components for this notification system, including a notification bell, settings panel, and toast notifications: {{previous}}",
      "workdir": "./frontend"
    },
    {
      "name": "integration-tests",
      "backend": "claude",
      "prompt": "Write integration tests using Jest and Supertest for the complete notification flow: {{previous}}"
    }
  ],
  "pass_working_dir": true
}
```

---

### Example 3: Code Review and Refactor

Review, refactor, and verify.

```
{
  "steps": [
    {
      "name": "review",
      "backend": "claude",
      "prompt": "Review this code for code smells, anti-patterns, and areas for improvement. Be specific about what should change and why."
    },
    {
      "name": "refactor",
      "backend": "codex",
      "prompt": "Refactor the code according to these suggestions: {{previous}}",
      "approval_mode": "auto"
    },
    {
      "name": "verify",
      "backend": "gemini",
      "prompt": "Verify that the refactored code maintains the same behavior and that all edge cases are still handled correctly: {{previous}}"
    }
  ],
  "stop_on_failure": true
}
```

---

### Example 4: Documentation Pipeline

Generate documentation at multiple levels.

```
{
  "steps": [
    {
      "name": "api-spec",
      "backend": "claude",
      "prompt": "Generate an OpenAPI 3.0 specification from the source code and route definitions"
    },
    {
      "name": "readme",
      "backend": "gemini",
      "prompt": "Create a comprehensive README.md based on this OpenAPI spec, including usage examples: {{previous}}"
    },
    {
      "name": "code-examples",
      "backend": "codex",
      "prompt": "Generate working code examples in Python, JavaScript, Go, and curl based on the API: {{previous}}"
    },
    {
      "name": "chinese-docs",
      "backend": "claude",
      "prompt": "Translate and adapt the README for Chinese developers, keeping technical terms in English: {{previous}}"
    }
  ],
  "output_dir": "./generated-docs"
}
```

---

### Example 5: Learning Path Generation

Create personalized learning materials.

```
{
  "steps": [
    {
      "name": "skill-assessment",
      "backend": "claude",
      "prompt": "Analyze this codebase and identify the key technical skills a developer needs to contribute effectively. Group by: frontend, backend, DevOps, and domain knowledge."
    },
    {
      "name": "curriculum",
      "backend": "gemini",
      "prompt": "Create a 12-week learning curriculum for these skills, with weekly milestones: {{previous}}"
    },
    {
      "name": "resources",
      "backend": "codex",
      "prompt": "Recommend specific books, online courses, and documentation for each week of this curriculum: {{previous}}"
    },
    {
      "name": "projects",
      "backend": "claude",
      "prompt": "Suggest hands-on projects that apply these skills directly to this codebase, increasing in complexity: {{previous}}"
    }
  ]
}
```

---

## Advanced Patterns

### Pattern 1: Gate-Based Workflows

Use step failures to create conditional flows.

```
{
  "steps": [
    {
      "name": "security-scan",
      "backend": "gemini",
      "prompt": "Security audit: list any critical or high severity vulnerabilities"
    },
    {
      "name": "fix-critical",
      "backend": "codex",
      "prompt": "Fix any critical vulnerabilities found: {{previous}}",
      "approval_mode": "auto"
    },
    {
      "name": "re-scan",
      "backend": "gemini",
      "prompt": "Re-scan to verify all critical issues are resolved: {{previous}}"
    }
  ],
  "stop_on_failure": true
}
```

If security-scan finds no critical issues, the chain stops early (exit code 0).

---

### Pattern 2: Multi-Backend Consensus

Use the same backend at different steps for verification.

```
{
  "steps": [
    {
      "name": "initial-analysis",
      "backend": "claude",
      "prompt": "Analyze this architecture and identify potential issues"
    },
    {
      "name": "solution-design",
      "backend": "codex",
      "prompt": "Design solutions for these issues: {{previous}}"
    },
    {
      "name": "review-solutions",
      "backend": "claude",
      "prompt": "Review these solutions for feasibility and potential downsides: {{previous}}"
    },
    {
      "name": "implementation",
      "backend": "codex",
      "prompt": "Implement the approved solutions: {{previous}}",
      "approval_mode": "auto"
    }
  ]
}
```

---

### Pattern 3: Data Processing Pipeline

Process data through multiple transformations.

```
{
  "steps": [
    {
      "name": "extract",
      "backend": "gemini",
      "prompt": "Extract all function names and their purposes from this code"
    },
    {
      "name": "categorize",
      "backend": "claude",
      "prompt": "Categorize these functions by: public API, internal utility, data access, and business logic: {{previous}}"
    },
    {
      "name": "document",
      "backend": "codex",
      "prompt": "Generate JSDoc comments for each function based on its category: {{previous}}"
    }
  ]
}
```

---

## Best Practices

### 1. Name Your Steps

Always use descriptive names:

```
{
  "name": "security-audit",  // Good
  "backend": "gemini",
  "prompt": "..."
}
```

Not:

```
{
  "backend": "gemini",  // Will show as "Step 1"
  "prompt": "..."
}
```

### 2. Handle Failures Appropriately

Use `stop_on_failure` based on your workflow:

```
{
  "steps": [
    {"name": "compile", "backend": "codex", "prompt": "Compile the code"},
    {"name": "test", "backend": "codex", "prompt": "Run tests"},
    {"name": "lint", "backend": "codex", "prompt": "Run linter"}  // Non-critical
  ],
  "stop_on_failure": true  // Stop if compile or test fails
}
```

### 3. Use Approval Modes Wisely

```
{
  "name": "auto-fix",
  "backend": "codex",
  "prompt": "Fix the bugs",
  "approval_mode": "auto"  // Safe for low-risk changes
}
```

### 4. Limit Agentic Turns

Prevent runaway execution:

```
{
  "name": "research",
  "backend": "claude",
  "prompt": "Research this topic",
  "max_turns": 10  // Prevent infinite loops
}
```

### 5. Use Working Directories

Keep steps organized:

```
{
  "steps": [
    {"name": "backend-tests", "backend": "codex", "prompt": "Run tests", "workdir": "./backend"},
    {"name": "frontend-tests", "backend": "codex", "prompt": "Run tests", "workdir": "./frontend"}
  ]
}
```

---

## Limitations

1. **Ephemeral Only** - Chain never persists sessions
2. **Single Predecessor** - Only `{{previous}}` is supported (not `{{step1}}`, etc.)
3. **Text Only** - Context is passed as text (not structured data)
4. **No Branching** - Linear execution only (no if/else)
5. **No Loops** - No iteration support

---

## Error Handling

### Exit Codes

- `0` - All steps completed successfully
- `1` - One or more steps failed (if `stop_on_failure: true`, indicates which step)

### Failed Step Identification

```
clinvk chain -f chain.json --json | jq '.failed_step'
# Output: 2 (if step 2 failed)
```

### Continuing After Failure

With `stop_on_failure: false`:

```
clinvk chain -f chain.json --json | jq '.results[] | select(.exit_code != 0)'
```

---

## HTTP API Equivalent

The HTTP server supports chain execution via POST `/api/v1/chain`:

```
curl -sS http://localhost:8080/api/v1/chain \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: your-key' \
  -d '{
    "steps": [
      {"backend": "claude", "prompt": "Analyze"},
      {"backend": "codex", "prompt": "Fix: {{previous}}"}
    ]
  }'
```

See [REST API Reference](../reference/rest-api.md) for full details.

---

## Related Topics

- [Parallel Execution](parallel-execution.md) - Concurrent multi-backend tasks
- [Backend Comparison](backend-comparison.md) - Side-by-side analysis
- [Use Cases](use-cases.md) - More real-world workflows
