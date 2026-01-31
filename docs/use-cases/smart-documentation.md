---
title: Smart Documentation
description: Generate architecture and API docs from your codebase using multiple backends.
---

# Smart Documentation

Generate high-quality documentation by combining parallel extraction with a synthesis chain.

## Workflow

### Step 1: Collect Source Context

Pick a small set of key files and summarize them first:

```bash
clinvk -b claude --output-format json \
  "Summarize the architecture based on: README.md, internal/app, internal/server" \
  > arch-summary.json

SUMMARY=$(jq -r '.content' arch-summary.json)
```

### Step 2: Parallel Doc Sections

```bash
jq -n --arg summary "$SUMMARY" '
{
  tasks: [
    {name: "overview", backend: "claude", prompt: "Write a project overview using:\n" + $summary},
    {name: "api", backend: "gemini", prompt: "Draft REST API docs based on:\n" + $summary},
    {name: "cli", backend: "codex", prompt: "Draft CLI usage docs based on:\n" + $summary}
  ]
}' > docs-tasks.json

clinvk parallel -f docs-tasks.json --json > docs-sections.json
```

### Step 3: Synthesize a Single Document

```bash
OV=$(jq -r '.results[] | select(.task_name=="overview").output' docs-sections.json)
API=$(jq -r '.results[] | select(.task_name=="api").output' docs-sections.json)
CLI=$(jq -r '.results[] | select(.task_name=="cli").output' docs-sections.json)

jq -n --arg ov "$OV" --arg api "$API" --arg cli "$CLI" '
{
  steps: [
    {backend: "claude", name: "draft", prompt: "Combine into a single doc:\nOVERVIEW:\n" + $ov + "\n\nAPI:\n" + $api + "\n\nCLI:\n" + $cli},
    {backend: "gemini", name: "polish", prompt: "Improve clarity and consistency:\n{{previous}}"}
  ]
}' > docs-synthesis.json

clinvk chain -f docs-synthesis.json --json > docs-final.json
```

## Notes

- Use small, curated inputs for better signal.
- For large repos, pre-summarize per module and feed those summaries into the workflow.
