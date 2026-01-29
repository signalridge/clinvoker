# Claude Code

Anthropic's AI coding assistant with deep reasoning capabilities and a focus on safety.

## Overview

Claude Code is a powerful AI coding assistant from Anthropic. It excels at:

- Complex multi-step reasoning
- Thorough code analysis and review
- Safe and responsible AI assistance
- Understanding context deeply

## Installation

Install Claude Code from [Anthropic](https://claude.ai/claude-code):

```bash
# Verify installation
which claude
claude --version
```

## Basic Usage

```bash
# Use Claude with clinvk
clinvk --backend claude "fix the bug in auth.go"
clinvk -b claude "explain this codebase"
```

## Models

| Model | Description |
|-------|-------------|
| `claude-opus-4-5-20251101` | Most capable, best for complex tasks |
| `claude-sonnet-4-20250514` | Balanced performance and speed |

Specify a model:

```bash
clinvk -b claude -m claude-sonnet-4-20250514 "quick review"
```

## Configuration

Configure Claude in `~/.clinvk/config.yaml`:

```yaml
backends:
  claude:
    # Default model
    model: claude-opus-4-5-20251101

    # Tool access (all, or comma-separated list)
    allowed_tools: all

    # Override unified approval mode
    approval_mode: default

    # Override unified sandbox mode
    sandbox_mode: default

    # Enable/disable this backend
    enabled: true

    # Custom system prompt
    system_prompt: ""

    # Extra CLI flags
    extra_flags: []
```

### Environment Variable

```bash
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
```

## Approval Modes

Claude supports different approval behaviors:

| Mode | Description |
|------|-------------|
| `default` | Let Claude decide based on action risk |
| `auto` | Automatically approve all actions |
| `none` | Never ask for approval (reject risky actions) |
| `always` | Always ask before any action |

Set via config:

```yaml
backends:
  claude:
    approval_mode: auto
```

Or per-command (in tasks/chains):

```json
{
  "backend": "claude",
  "prompt": "refactor the module",
  "approval_mode": "auto"
}
```

## Sandbox Modes

Control Claude's file system access:

| Mode | Description |
|------|-------------|
| `default` | Let Claude decide |
| `read-only` | Can only read files |
| `workspace` | Can modify files in project |
| `full` | Full file system access |

## Allowed Tools

Control which tools Claude can use:

```yaml
backends:
  claude:
    # All tools
    allowed_tools: all

    # Specific tools only
    allowed_tools: read,write,edit
```

## Session Resume

Claude Code stores sessions and supports resuming:

```bash
# Resume with clinvk
clinvk resume --last --backend claude
clinvk resume <session-id>
```

Internally uses Claude's `--resume` flag.

## Extra Flags

Pass additional flags to the Claude CLI:

```yaml
backends:
  claude:
    extra_flags:
      - "--add-dir"
      - "./docs"
```

Common flags:

| Flag | Description |
|------|-------------|
| `--add-dir <path>` | Add additional directory to context |
| `--verbose` | Enable verbose output |

## Best Practices

!!! tip "Use Opus for Complex Tasks"
    Claude Opus is ideal for multi-step reasoning, code architecture, and thorough reviews.

!!! tip "Leverage Session Continuity"
    Claude excels at maintaining context across a conversation. Use `clinvk -c` to continue sessions.

!!! tip "Trust the Defaults"
    Claude's default approval and sandbox modes are well-tuned for safety while being useful.

## Use Cases

### Code Review

```bash
clinvk -b claude "review this PR for security issues and code quality"
```

### Complex Refactoring

```bash
clinvk -b claude "refactor the authentication system to use JWT tokens"
```

### Architecture Analysis

```bash
clinvk -b claude "analyze this codebase architecture and suggest improvements"
```

### Bug Investigation

```bash
clinvk -b claude "investigate why the tests are failing in the CI pipeline"
```

## Troubleshooting

### Backend Not Available

```bash
# Check if Claude is installed
which claude

# Check clinvk detection
clinvk config show | grep claude
```

### Rate Limits

If hitting rate limits, consider:

- Using a different model
- Spacing out requests
- Running in sequential mode for comparisons

## Next Steps

- [Codex CLI Guide](codex.md)
- [Gemini CLI Guide](gemini.md)
- [Backend Comparison](../backend-comparison.md)
