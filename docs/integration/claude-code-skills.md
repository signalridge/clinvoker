# Claude Code Skills Integration

This guide explains how to integrate clinvk with Claude Code Skills to extend Claude's capabilities with multi-backend AI support.

## Why Use clinvk in Skills?

Claude Code Skills extend Claude's capabilities, but sometimes you need:

- **Other AI Backends**: Gemini excels at data analysis, Codex at code generation
- **Multi-Model Collaboration**: Complex tasks benefit from multiple perspectives
- **Standardized API**: Single HTTP interface for all backends
- **Parallel Processing**: Run multiple AI tasks concurrently

## Prerequisites

1. **clinvk server running**:
   ```bash
   clinvk serve --port 8080 &
   ```

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

curl -s http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d "{
    \"backend\": \"gemini\",
    \"prompt\": \"Analyze this data and provide insights: $DATA\",
    \"output_format\": \"json\"
  }" | jq -r '.result'
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

# Parallel review using all backends
RESULT=$(curl -s http://localhost:8080/api/v1/parallel \
  -H "Content-Type: application/json" \
  -d "{
    \"tasks\": [
      {
        \"backend\": \"claude\",
        \"prompt\": \"Review this code for architecture and design patterns:\\n$CODE\"
      },
      {
        \"backend\": \"codex\",
        \"prompt\": \"Review this code for performance issues and optimizations:\\n$CODE\"
      },
      {
        \"backend\": \"gemini\",
        \"prompt\": \"Review this code for security vulnerabilities:\\n$CODE\"
      }
    ]
  }")

echo "## Multi-Model Code Review Results"
echo ""
echo "### Architecture Review (Claude)"
echo "$RESULT" | jq -r '.results[0].result // .results[0].error'
echo ""
echo "### Performance Review (Codex)"
echo "$RESULT" | jq -r '.results[1].result // .results[1].error'
echo ""
echo "### Security Review (Gemini)"
echo "$RESULT" | jq -r '.results[2].result // .results[2].error'
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

curl -s http://localhost:8080/api/v1/chain \
  -H "Content-Type: application/json" \
  -d "{
    \"steps\": [
      {
        \"name\": \"analyze\",
        \"backend\": \"claude\",
        \"prompt\": \"Analyze the structure and purpose of this code. List all functions, classes, and their relationships:\\n$CODE\"
      },
      {
        \"name\": \"document\",
        \"backend\": \"codex\",
        \"prompt\": \"Based on this analysis, generate comprehensive API documentation in Markdown format:\\n{{previous}}\"
      },
      {
        \"name\": \"polish\",
        \"backend\": \"gemini\",
        \"prompt\": \"Improve the readability and add helpful examples to this documentation:\\n{{previous}}\"
      }
    ]
  }" | jq -r '.results[-1].result'
```
```

## Advanced Patterns

### Error Handling

```bash
#!/bin/bash
RESPONSE=$(curl -s -w "\n%{http_code}" http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "'"$1"'"}')

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" != "200" ]; then
  echo "Error: API returned $HTTP_CODE"
  echo "$BODY" | jq -r '.error.message // .error // "Unknown error"'
  exit 1
fi

echo "$BODY" | jq -r '.result'
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

curl -s http://localhost:8080/api/v1/prompt \
  -d "{\"backend\": \"$BACKEND\", \"prompt\": \"$PROMPT\"}" | jq -r '.result'
```

### Streaming Response

For long-running tasks, use streaming:

```bash
#!/bin/bash
curl -N http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{
    "backend": "claude",
    "prompt": "'"$1"'",
    "stream": true
  }' | while read -r line; do
    # Process each SSE event
    if [[ $line == data:* ]]; then
      echo "${line#data: }" | jq -r '.content // empty'
    fi
  done
```

## Best Practices

### 1. Use Ephemeral Mode

For stateless skill execution, use ephemeral sessions:

```bash
curl -s http://localhost:8080/api/v1/prompt \
  -d '{
    "backend": "claude",
    "prompt": "...",
    "session_mode": "ephemeral"
  }'
```

### 2. Choose the Right Backend

| Task Type | Recommended Backend | Why |
|-----------|--------------------|----|
| Code Review | Claude | Deep understanding, context |
| Code Generation | Codex | Optimized for code |
| Data Analysis | Gemini | Strong analytical capabilities |
| Documentation | Any | All perform well |
| Security Audit | Claude + Gemini | Different perspectives |

### 3. Handle Timeouts

Set appropriate timeouts for long tasks:

```bash
curl -s --max-time 120 http://localhost:8080/api/v1/prompt \
  -d '{"backend": "claude", "prompt": "...", "timeout": 60}'
```

### 4. Format Output

Structure skill output for Claude to process:

```bash
echo "## Skill Results"
echo ""
echo "### Summary"
echo "$SUMMARY"
echo ""
echo "### Details"
echo '```json'
echo "$RESULT" | jq '.'
echo '```'
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

### Server Not Running

```bash
# Check if server is running
curl -s http://localhost:8080/health

# If not, start it
clinvk serve --port 8080 &
```

### Backend Not Available

```bash
# Check available backends
curl -s http://localhost:8080/api/v1/backends | jq '.backends'
```

### Permission Issues

Ensure skill scripts are executable:

```bash
chmod +x ~/.claude/skills/*/SKILL.md
```

## Next Steps

- [LangChain/LangGraph Integration](langchain-langgraph.md) - For Python-based agents
- [CI/CD Integration](ci-cd.md) - Automate in pipelines
- [REST API Reference](../reference/rest-api.md) - Complete API documentation
