---
title: Test Generation Pipeline
description: Design, implement, and review tests using chain + parallel execution.
---

# Test Generation Pipeline

Use a structured pipeline to generate tests, implement them, and validate coverage.

## Workflow

### Step 1: Test Plan

```bash
clinvk -b gemini --output-format json \
  "Design a test plan for the Go package in ./internal/session. Include edge cases." \
  > test-plan.json

PLAN=$(jq -r '.content' test-plan.json)
```

### Step 2: Implement Tests (Chain)

```bash
jq -n --arg plan "$PLAN" '
{
  steps: [
    {name: "generate", backend: "codex", prompt: "Write unit tests based on this plan:\n" + $plan},
    {name: "refine", backend: "codex", prompt: "Refine and fix tests:\n{{previous}}"},
    {name: "review", backend: "claude", prompt: "Review tests for correctness and coverage gaps:\n{{previous}}"}
  ]
}' > test-pipeline.json

clinvk chain -f test-pipeline.json --json > test-results.json
```

### Step 3: Coverage Review (Parallel)

```bash
TESTS=$(jq -r '.results[-1].output' test-results.json)

jq -n --arg tests "$TESTS" '
{
  tasks: [
    {name: "coverage", backend: "claude", prompt: "Review coverage gaps for these tests:\n" + $tests},
    {name: "edge-cases", backend: "gemini", prompt: "Suggest missing edge cases for:\n" + $tests}
  ]
}' > coverage-tasks.json

clinvk parallel -f coverage-tasks.json --json > coverage-results.json
```

## Notes

- `chain` is best for sequential refinement.
- `parallel` is best for independent critiques.
