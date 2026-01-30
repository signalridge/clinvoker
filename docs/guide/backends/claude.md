# Claude Code Backend

Claude Code is the most feature-complete backend in clinvk, offering rich capabilities for complex development workflows.

---

## Overview

| Feature | Support |
|---------|---------|
| **System Prompt** | ✓ Full support |
| **Max Turns** | ✓ Full support |
| **Session Resume** | ✓ Full support |
| **Output Formats** | text, json, stream-json |
| **Tool Control** | ✓ Allowed tools list |
| **Ephemeral Mode** | ✓ `--no-session-persistence` |
| **Sandbox Control** | ✗ Not mapped (Claude handles internally) |

---

## Requirements

1. **Install Claude Code CLI**

   ```bash
   npm install -g @anthropic-ai/claude-code
   ```

2. **Authenticate**

   ```bash
   claude auth login
   ```

3. **Verify Installation**

   ```bash
   claude --version
   clinvk -b claude "test connection"
   ```

---

## How clinvk Uses Claude

### Non-Interactive Execution

```
# clinvk builds this command:
claude --print \
  --model claude-opus-4-5-20251101 \
  --permission-mode acceptEdits \
  "your prompt here"
```

### Session Resume

```
# clinvk builds this command:
claude --resume <session_id> \
  --print \
  "continue from previous context"
```

### JSON Output Mode

```
# For internal parsing (chain, parallel):
claude --print --output-format json
```

---

## Configuration Options

### Full Configuration Example

```
backends:
  claude:
    # Model selection
    model: claude-opus-4-5-20251101

    # Tool restrictions
    allowed_tools: Bash,Edit,Read  # Comma-separated or 'all'

    # Default system prompt
    system_prompt: |
      You are a senior software engineer reviewing code.
      Be thorough but concise. Always explain the 'why'.

    # Additional directories to include
    extra_flags:
      - "--add-dir"
      - "./docs"
      - "--add-dir"
      - "./shared"

    # Enable/disable this backend
    enabled: true
```

### Model Aliases

clinvk provides convenient aliases:

| Alias | Resolves To | Best For |
|-------|-------------|----------|
| `fast`, `quick` | `haiku` | Quick tasks, simple queries |
| `balanced`, `default` | `sonnet` | General development work |
| `best`, `powerful` | `opus` | Complex analysis, architecture |

```
# Use aliases
clinvk -b claude -m fast "quick summary"
clinvk -b claude -m balanced "code review"
clinvk -b claude -m best "architecture design"

# Or full model names
clinvk -b claude -m claude-opus-4-5-20251101 "deep analysis"
```

---

## Approval Mode Mapping

| clinvk `approval_mode` | Claude Flag | Behavior |
|------------------------|-------------|----------|
| `auto` | `--permission-mode acceptEdits` | Auto-accept edits, ask for other actions |
| `none` | `--permission-mode dontAsk` | Never ask for permission |
| `always` | `--permission-mode default` | Always ask for permission |
| `default` | (none) | Use Claude's default |

```
# Auto-approve edits (safe for trusted code)
clinvk -b claude --approval-mode auto "refactor this function"

# Never ask (use with caution)
clinvk -b claude --approval-mode none "apply all fixes"
```

---

## Unique Features

### 1. System Prompts

Claude is the only backend supporting custom system prompts:

```
clinvk -b claude "review this code" \
  --system-prompt "You are a security-focused code reviewer. Highlight all security concerns."
```

Or in config:

```
backends:
  claude:
    system_prompt: "Your custom persona here"
```

### 2. Max Turns Limit

Prevent runaway execution:

```
clinvk -b claude --max-turns 10 "research this topic"
```

```
backends:
  claude:
    max_turns: 15
```

### 3. Tool Control

Restrict which tools Claude can use:

```
backends:
  claude:
    # Specific tools only
    allowed_tools: "Read,Edit,Bash"

    # Or all tools
    allowed_tools: "all"
```

### 4. Extra Flags

Pass additional Claude-specific flags:

```
backends:
  claude:
    extra_flags:
      - "--verbose"
      - "--add-dir"
      - "./config"
```

---

## Use Cases

### When to Choose Claude

| Scenario | Why Claude? |
|----------|-------------|
| **Architecture review** | Strong reasoning and system design |
| **Complex debugging** | Excellent at root cause analysis |
| **Documentation** | High-quality prose and explanations |
| **Learning/Teaching** | Clear explanations with examples |
| **Security audits** | Thorough security analysis |
| **Multi-file changes** | Good context management |

### Claude-Optimized Workflows

```
// Multi-turn analysis in chain
{
  "steps": [
    {
      "backend": "claude",
      "prompt": "Deep architectural analysis of this system",
      "max_turns": 10,
      "name": "deep-analysis"
    }
  ]
}
```

```
// Security review with system prompt
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "Security audit this code",
      "system_prompt": "Focus on OWASP Top 10 and injection vulnerabilities"
    }
  ]
}
```

---

## Limitations

1. **Sandbox Mode**: Not directly mapped (Claude manages its own sandbox)
2. **Rate Limits**: Subject to Anthropic's rate limiting
3. **Cost**: Opus model is more expensive than other backends

---

## Troubleshooting

### "Backend 'claude' is not available"

```
# Check if claude CLI is installed
which claude
claude --version

# If not found, install:
npm install -g @anthropic-ai/claude-code
```

### "No session persistence" warning

This is expected when using `--ephemeral` flag. To persist sessions, remove `--ephemeral`.

### Slow responses with opus model

Opus is powerful but slower. For faster responses:

```
clinvk -b claude -m sonnet "your prompt"  # Balanced speed/quality
clinvk -b claude -m haiku "your prompt"   # Fastest
```

---

## Configuration Reference

| Config Key | Type | Default | Description |
|------------|------|---------|-------------|
| `model` | string | `claude-opus-4-5-20251101` | Default model |
| `allowed_tools` | string | `all` | Comma-separated tool list |
| `system_prompt` | string | - | Default system prompt |
| `max_turns` | int | 0 | Max agentic turns (0 = unlimited) |
| `enabled` | bool | `true` | Enable this backend |
| `extra_flags` | array | [] | Additional CLI flags |

---

## Related

- [Backend Comparison](../backend-comparison.md) - Compare all backends
- [Configuration](../../reference/configuration.md) - Full config reference
- [Claude Code Documentation](https://docs.anthropic.com/en/docs/agents-and-tools/claude-code/overview) - Official docs
