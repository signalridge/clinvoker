---
title: Incident Response War Room
description: Triage production incidents with parallel analysis and a synthesis chain.
---

# Incident Response War Room

When production goes sideways, you need speed and perspective. This use case uses **parallel triage** plus a **chain synthesis** to turn logs into an action plan.

## Scenario

You have:

- A spike in error rates
- Recent deploy changes
- A few log excerpts

You want:

- Root cause hypotheses
- Immediate mitigation steps
- Longer-term fixes

## Workflow

### Step 1: Parallel Triage

```bash
LOGS_FILE=incident-logs.txt
CHANGES_FILE=deploy-notes.md

LOGS=$(cat "$LOGS_FILE")
CHANGES=$(cat "$CHANGES_FILE")

jq -n --arg logs "$LOGS" --arg changes "$CHANGES" '
{
  tasks: [
    {name: "hypotheses", backend: "claude", prompt: "Generate root-cause hypotheses from logs:\n" + $logs},
    {name: "mitigation", backend: "codex", prompt: "List immediate mitigation steps based on:\n" + $logs},
    {name: "change-risk", backend: "gemini", prompt: "Analyze recent deploy notes for risk:\n" + $changes}
  ]
}' > incident-tasks.json

clinvk parallel -f incident-tasks.json --json > incident-triage.json
```

### Step 2: Synthesize Action Plan

```bash
HYP=$(jq -r '.results[] | select(.task_name=="hypotheses").output' incident-triage.json)
MIT=$(jq -r '.results[] | select(.task_name=="mitigation").output' incident-triage.json)
RISK=$(jq -r '.results[] | select(.task_name=="change-risk").output' incident-triage.json)

jq -n --arg hyp "$HYP" --arg mit "$MIT" --arg risk "$RISK" '
{
  steps: [
    {backend: "claude", name: "plan", prompt: "Create an incident action plan using:\nHYPOTHESES:\n" + $hyp + "\n\nMITIGATION:\n" + $mit + "\n\nDEPLOY RISK:\n" + $risk},
    {backend: "gemini", name: "comms", prompt: "Draft a short incident update for stakeholders:\n{{previous}}"}
  ]
}' > incident-plan.json

clinvk chain -f incident-plan.json --json > incident-report.json
```

## Outcome

- Fast hypothesis generation
- Clear mitigations
- A stakeholder-friendly update

## Notes

- For long incidents, move to persistent sessions (`clinvk [prompt]` + `resume`).
- Keep logs and change notes small and focused for best results.
