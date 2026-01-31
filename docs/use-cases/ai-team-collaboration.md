---
title: AI Team Collaboration
description: Simulate a multi-AI development team with specialized roles using clinvoker.
---

# AI Team Collaboration

Simulate a development team composed of AI assistants, each with a specialized role. This pattern combines **parallel reviews** with a **chain pipeline** to turn requirements into reviewed, test-backed code.

## Scenario

- **Claude** acts as the architect and reviewer
- **Codex** implements and refines code
- **Gemini** focuses on security and documentation

## Implementation

### Step 1: Architecture from Claude

```bash
clinvk -b claude --output-format json \
  "Design a secure authentication system with JWT, refresh rotation, and RBAC. \
Include API endpoints and data model." > architecture.json

ARCH=$(jq -r '.content' architecture.json)
```

### Step 2: Parallel Reviews

```bash
jq -n --arg arch "$ARCH" '
{
  tasks: [
    {name: "arch-review", backend: "claude", prompt: "Review this architecture for scalability and API design:\n" + $arch},
    {name: "security-review", backend: "gemini", prompt: "Security audit this architecture (OWASP, token handling, authz):\n" + $arch},
    {name: "impl-risk-review", backend: "codex", prompt: "Identify implementation risks and edge cases:\n" + $arch}
  ]
}' > review-tasks.json

clinvk parallel -f review-tasks.json --json > reviews.json
```

### Step 3: Implementation Chain

`chain` only supports `{{previous}}`, so we inject the architecture into the first step.

```bash
jq -n --arg arch "$ARCH" '
{
  steps: [
    {name: "implement", backend: "codex", prompt: "Implement core auth middleware based on:\n" + $arch},
    {name: "tests", backend: "codex", prompt: "Write unit tests for:\n{{previous}}"},
    {name: "review", backend: "claude", prompt: "Review code + tests; list fixes and risks:\n{{previous}}"}
  ]
}' > implement-pipeline.json

clinvk chain -f implement-pipeline.json --json > implement-results.json
```

### Step 4: Documentation Chain

```bash
IMPLEMENTATION=$(jq -r '.results[-1].output' implement-results.json)

jq -n --arg impl "$IMPLEMENTATION" '
{
  steps: [
    {name: "api-docs", backend: "gemini", prompt: "Write API docs for the following implementation:\n" + $impl},
    {name: "usage-guide", backend: "gemini", prompt: "Create a developer guide based on:\n{{previous}}"},
    {name: "final-review", backend: "claude", prompt: "Review the docs for accuracy and clarity:\n{{previous}}"}
  ]
}' > docs-pipeline.json

clinvk chain -f docs-pipeline.json --json > docs-results.json
```

## Notes

- `parallel` and `chain` runs are **ephemeral** (no sessions are saved).
- Use `clinvk [prompt]` + `resume` when you want persistent sessions.

## Why This Works

- Parallel reviews reduce blind spots.
- The chain enforces a clean, auditable handoff between stages.
- Each backend is used where it performs best.
