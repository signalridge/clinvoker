---
title: Multi-Model Research
description: Run research across multiple backends and synthesize a single report.
---

# Multi-Model Research

Use multiple backends to gather diverse perspectives, then synthesize a unified report.

## Workflow

### Step 1: Parallel Research

```bash
QUESTION="How should we design a multi-tenant rate limiter for an API gateway?"

jq -n --arg q "$QUESTION" '
{
  tasks: [
    {backend: "claude", name: "architecture", prompt: "Propose a system design for: " + $q},
    {backend: "codex", name: "implementation", prompt: "Provide an implementation approach + pseudocode for: " + $q},
    {backend: "gemini", name: "tradeoffs", prompt: "List tradeoffs, risks, and alternatives for: " + $q}
  ]
}' > research-tasks.json

clinvk parallel -f research-tasks.json --json > research-results.json
```

### Step 2: Synthesize Report (Chain)

```bash
ARCH=$(jq -r '.results[] | select(.task_name=="architecture").output' research-results.json)
IMPL=$(jq -r '.results[] | select(.task_name=="implementation").output' research-results.json)
TRD=$(jq -r '.results[] | select(.task_name=="tradeoffs").output' research-results.json)

jq -n --arg arch "$ARCH" --arg impl "$IMPL" --arg trd "$TRD" '
{
  steps: [
    {backend: "claude", name: "draft", prompt: "Draft a structured report using:\nARCHITECTURE:\n" + $arch + "\n\nIMPLEMENTATION:\n" + $impl + "\n\nTRADEOFFS:\n" + $trd},
    {backend: "gemini", name: "improve", prompt: "Improve clarity and add missing caveats:\n{{previous}}"}
  ]
}' > synthesis-pipeline.json

clinvk chain -f synthesis-pipeline.json --json > research-report.json
```

### Step 3: Follow-up Questions

```bash
REPORT=$(jq -r '.results[-1].output' research-report.json)

clinvk -b claude --output-format json \
  "Based on this report, list 5 follow-up research questions:\n$REPORT" > followups.json
```

## Notes

- Use `parallel` for breadth, `chain` for synthesis.
- When you need persistent sessions, switch to `clinvk [prompt]` + `resume`.
