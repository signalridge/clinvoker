# Codex CLI

OpenAI's code-focused CLI tool optimized for code generation and programming tasks.

## Overview

Codex CLI is OpenAI's command-line tool focused on code generation and programming assistance. It excels at:

- Rapid code generation
- Writing tests and boilerplate
- Code transformations
- Quick programming tasks

## Installation

Install Codex CLI from [OpenAI](https://github.com/openai/codex-cli):

```bash
# Verify installation
which codex
codex --version
```

## Basic Usage

```bash
# Use Codex with clinvk
clinvk --backend codex "implement a REST API handler"
clinvk -b codex "generate unit tests for user.go"
```

## Models

| Model | Description |
|-------|-------------|
| `o3` | Latest and most capable model |
| `o3-mini` | Faster, lighter model |

Specify a model:

```bash
clinvk -b codex -m o3-mini "quick code generation"
```

## Configuration

Configure Codex in `~/.clinvk/config.yaml`:

```yaml
backends:
  codex:
    # Default model
    model: o3

    # Enable/disable this backend
    enabled: true

    # Extra CLI flags
    extra_flags: []
```

### Environment Variable

```bash
export CLINVK_CODEX_MODEL=o3-mini
```

## Session Management

Codex resumes sessions via the `codex exec resume` subcommand (handled automatically by `clinvk`):

```bash
# Resume with clinvk
clinvk resume --last --backend codex
clinvk resume <session-id>
```

## Unified Options

These options work with Codex:

| Option | Description |
|--------|-------------|
| `model` | Model to use |
| `max_tokens` | Maximum response tokens |
| `max_turns` | Maximum agentic turns |

## Extra Flags

Pass additional flags to Codex:

```yaml
backends:
  codex:
    extra_flags:
      - "--quiet"
```

Common flags:

| Flag | Description |
|------|-------------|
| `--quiet` | Reduce output verbosity |

## Best Practices

!!! tip "Use for Code Generation"
    Codex is optimized for generating code quickly. It's great for boilerplate and repetitive tasks.

!!! tip "Combine with Other Backends"
    Use Codex to generate code, then Claude to review it - leverage the chain command.

!!! tip "Batch Similar Tasks"
    Run multiple code generation tasks in parallel for efficiency.

## Use Cases

### Generate Boilerplate

```bash
clinvk -b codex "create a CRUD API for the User model"
```

### Write Tests

```bash
clinvk -b codex "generate comprehensive unit tests for the auth module"
```

### Code Transformation

```bash
clinvk -b codex "convert this callback-based code to async/await"
```

### Quick Implementations

```bash
clinvk -b codex "implement a binary search function"
```

## Comparison with Claude

| Aspect | Codex | Claude |
|--------|-------|--------|
| Speed | Faster | More thorough |
| Best for | Code generation | Complex reasoning |
| Context | Good | Excellent |
| Safety focus | Standard | High |

## Workflow Example

Use Codex and Claude together:

```json
{
  "steps": [
    {
      "name": "generate",
      "backend": "codex",
      "prompt": "implement user authentication"
    },
    {
      "name": "review",
      "backend": "claude",
      "prompt": "review this code for security: {{previous}}"
    }
  ]
}
```

## Troubleshooting

### Backend Not Available

```bash
# Check if Codex is installed
which codex

# Check clinvk detection
clinvk config show | grep codex
```

### Model Errors

If a model isn't available:

```bash
# List available models
codex models list

# Update config to use available model
clinvk config set backends.codex.model o3
```

## Next Steps

- [Claude Code Guide](claude.md)
- [Gemini CLI Guide](gemini.md)
- [Backend Comparison](../backend-comparison.md)
