# Claude Code Skills

Use clinvk inside Claude Code Skills to call other backends.

## Prerequisites

- `clinvk` installed
- At least one backend CLI installed

## Minimal skill example

```bash
# ~/.claude/skills/analyze-data/command.sh
#!/bin/bash
DATA="$1"

clinvk -b gemini --ephemeral -o text "Analyze this data: $DATA"
```

## Multi‑model review skill

```bash
# ~/.claude/skills/multi-review/command.sh
#!/bin/bash
CODE="$1"

echo "### Architecture (Claude)"
clinvk -b claude --ephemeral "Review architecture: $CODE"

echo "### Performance (Codex)"
clinvk -b codex --ephemeral "Review performance: $CODE"

echo "### Security (Gemini)"
clinvk -b gemini --ephemeral "Review security: $CODE"
```

## Parallel review skill

```bash
# ~/.claude/skills/parallel-review/command.sh
#!/bin/bash
CODE="$1"

cat > /tmp/review-tasks.json << JSON
{ "tasks": [
  {"backend":"claude","prompt":"Architecture review: $CODE"},
  {"backend":"codex","prompt":"Performance review: $CODE"},
  {"backend":"gemini","prompt":"Security review: $CODE"}
]}
JSON

clinvk parallel -f /tmp/review-tasks.json --json
```

## Notes

- Use `--ephemeral` for stateless skill runs.
- For structured output, prefer `--output-format json` or `--json`.
