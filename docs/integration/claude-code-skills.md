# Claude Code Skills Integration

This guide explains how to integrate clinvk with Claude Code Skills to extend Claude's capabilities with multi-backend AI support.

## Why Use clinvk in Skills?

Claude Code Skills extend Claude's capabilities, but sometimes you need:

- **Other AI Backends**: Gemini excels at data analysis, Codex at code generation
- **Multi-Model Collaboration**: Complex tasks benefit from multiple perspectives
- **Parallel Processing**: Run multiple AI tasks concurrently

## Prerequisites

1. **clinvk installed** and in your PATH
2. **At least one backend CLI installed** (`claude`, `codex`, or `gemini`)

## Basic Skill Example

### Single Backend Call

Create a skill that uses Gemini for data analysis:

```markdown
<!-- ~/.claude/skills/analyze-data/SKILL.md -->
# Data Analysis Skill

Analyzes data using Gemini CLI via clinvk.

## Usage
Run this skill when you need to analyze structured data.

## Script
```bash
#!/bin/bash
DATA="$1"

clinvk -b gemini -o json --ephemeral "Analyze this data and provide insights: $DATA"
```

```

### Using the Skill

In Claude Code, the skill can be invoked:

```

User: /analyze-data {"sales": [100, 150, 200], "months": ["Jan", "Feb", "Mar"]}

```

## Multi-Model Review Skill

A more powerful skill that uses multiple backends for comprehensive code review:

```markdown
<!-- ~/.claude/skills/multi-review/SKILL.md -->
# Multi-Model Code Review

Performs comprehensive code review using Claude (architecture),
Codex (performance), and Gemini (security).

## Usage
Provide a file path or code snippet for multi-perspective review.

## Script
```bash
#!/bin/bash
CODE="$1"

echo "## Multi-Model Code Review Results"
echo ""

echo "### Architecture Review (Claude)"
clinvk -b claude --ephemeral "Review this code for architecture and design patterns:
$CODE"

echo ""
echo "### Performance Review (Codex)"
clinvk -b codex --ephemeral "Review this code for performance issues and optimizations:
$CODE"

echo ""
echo "### Security Review (Gemini)"
clinvk -b gemini --ephemeral "Review this code for security vulnerabilities:
$CODE"
```

```

## Parallel Review Skill

For faster multi-model review using parallel execution:

```markdown
<!-- ~/.claude/skills/parallel-review/SKILL.md -->
# Parallel Multi-Model Review

Fast parallel code review using all backends simultaneously.

## Script
```bash
#!/bin/bash
CODE="$1"

# Create tasks file
cat > /tmp/review-tasks.json << EOF
{
  "tasks": [
    {"backend": "claude", "prompt": "Review for architecture and design: $CODE"},
    {"backend": "codex", "prompt": "Review for performance issues: $CODE"},
    {"backend": "gemini", "prompt": "Review for security vulnerabilities: $CODE"}
  ]
}
EOF

clinvk parallel -f /tmp/review-tasks.json --json | jq -r '
  "## Architecture (Claude)\n" + .results[0].output + "\n\n" +
  "## Performance (Codex)\n" + .results[1].output + "\n\n" +
  "## Security (Gemini)\n" + .results[2].output
'
```

```

## Chain Execution Skill

A skill that pipelines output through multiple backends:

```markdown
<!-- ~/.claude/skills/doc-pipeline/SKILL.md -->
# Documentation Pipeline

Generates polished documentation through a multi-step pipeline:
1. Claude analyzes code structure
2. Codex generates documentation
3. Gemini polishes and improves readability

## Script
```bash
#!/bin/bash
CODE="$1"

# Create pipeline file
cat > /tmp/doc-pipeline.json << EOF
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "Analyze the structure and purpose of this code. List all functions, classes, and their relationships:\n$CODE"
    },
    {
      "name": "document",
      "backend": "codex",
      "prompt": "Based on this analysis, generate comprehensive API documentation in Markdown format:\n{{previous}}"
    },
    {
      "name": "polish",
      "backend": "gemini",
      "prompt": "Improve the readability and add helpful examples to this documentation:\n{{previous}}"
    }
  ]
}
EOF

clinvk chain -f /tmp/doc-pipeline.json --json | jq -r '.results[-1].output'
```

```

## Advanced Patterns

### Error Handling

```bash
#!/bin/bash
set -e

if ! OUTPUT=$(clinvk -b claude --ephemeral "$1" 2>&1); then
  echo "Error executing clinvk: $OUTPUT"
  exit 1
fi

echo "$OUTPUT"
```

### Conditional Backend Selection

```bash
#!/bin/bash
TASK_TYPE="$1"
PROMPT="$2"

case "$TASK_TYPE" in
  "analyze")
    BACKEND="claude"
    ;;
  "generate")
    BACKEND="codex"
    ;;
  "research")
    BACKEND="gemini"
    ;;
  *)
    BACKEND="claude"
    ;;
esac

clinvk -b "$BACKEND" --ephemeral "$PROMPT"
```

### Compare Backends

```bash
#!/bin/bash
# Get responses from all backends and compare
clinvk compare --all-backends "$1"
```

## Best Practices

### 1. Use Ephemeral Mode

For stateless skill execution, always use `--ephemeral`:

```bash
clinvk -b claude --ephemeral "your prompt"
```

### 2. Choose the Right Backend

| Task Type | Recommended Backend | Why |
|-----------|--------------------|----|
| Code Review | Claude | Deep understanding, context |
| Code Generation | Codex | Optimized for code |
| Data Analysis | Gemini | Strong analytical capabilities |
| Documentation | Any | All perform well |
| Security Audit | Claude + Gemini | Different perspectives |

### 3. Use JSON Output for Parsing

When you need to process the output:

```bash
clinvk -b claude -o json --ephemeral "..." | jq -r '.content'
```

### 4. Format Output for Claude

Structure skill output for Claude to process:

```bash
echo "## Skill Results"
echo ""
echo "### Summary"
clinvk -b gemini --ephemeral "Summarize: $INPUT"
```

## Skill Directory Structure

```
~/.claude/skills/
├── analyze-data/
│   └── SKILL.md
├── multi-review/
│   └── SKILL.md
├── doc-pipeline/
│   └── SKILL.md
└── shared/
    └── clinvk-helpers.sh  # Shared functions
```

## Troubleshooting

### Backend Not Available

```bash
# Check available backends
clinvk config show | grep available
```

### Check Version

```bash
clinvk version
```

## Next Steps

- [LangChain/LangGraph Integration](langchain-langgraph.md) - For Python-based agents
- [CI/CD Integration](ci-cd.md) - Automate in pipelines
- [CLI Commands Reference](../reference/commands/index.md) - Complete CLI documentation
