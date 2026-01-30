# Gemini CLI

Google's Gemini AI assistant with broad knowledge and multimodal capabilities.

## Overview

Gemini CLI is Google's command-line interface for the Gemini AI model. It excels at:

- Broad knowledge and general questions
- Documentation and explanations
- Multimodal tasks (when supported)
- Research and information gathering

## Installation

Install Gemini CLI from [Google](https://github.com/google/gemini-cli):

```bash
# Verify installation
which gemini
gemini --version
```bash

## Basic Usage

```bash
# Use Gemini with clinvk
clinvk --backend gemini "explain how this algorithm works"
clinvk -b gemini "write documentation for this API"
```

## Models

| Model | Description |
|-------|-------------|
| `gemini-2.5-pro` | Latest and most capable model |
| `gemini-2.5-flash` | Faster, optimized for speed |

Specify a model:

```bash
clinvk -b gemini -m gemini-2.5-flash "quick explanation"
```text

## Configuration

Configure Gemini in `~/.clinvk/config.yaml`:

```yaml
backends:
  gemini:
    # Default model
    model: gemini-2.5-pro

    # Enable/disable this backend
    enabled: true

    # Extra CLI flags
    extra_flags: []
```

### Environment Variable

```bash
export CLINVK_GEMINI_MODEL=gemini-2.5-flash
```bash

## Session Management

Gemini uses `--resume` for session resume:

```bash
# Resume with clinvk
clinvk resume --last --backend gemini
clinvk resume <session-id>
```

## Sandbox Mode

Gemini supports sandbox mode for controlled execution:

```yaml
backends:
  gemini:
    extra_flags:
      - "--sandbox"
```bash

## Unified Options

These options work with Gemini:

| Option | Description |
|--------|-------------|
| `model` | Model to use |
| `max_tokens` | Maximum response tokens |
| `max_turns` | Maximum agentic turns |

## Best Practices

!!! tip "Use for Explanations"
    Gemini's broad knowledge makes it excellent for explaining concepts and providing context.

!!! tip "Leverage for Documentation"
    Use Gemini to write or improve documentation with its clear explanations.

!!! tip "Research Tasks"
    Gemini is great for gathering information and research-oriented queries.

## Use Cases

### Documentation

```bash
clinvk -b gemini "write comprehensive documentation for this module"
```

### Explanations

```bash
clinvk -b gemini "explain the architecture of this microservice"
```bash

### Research

```bash
clinvk -b gemini "what are the best practices for implementing rate limiting"
```

### Code Review

```bash
clinvk -b gemini "review this code and explain potential issues"
```text

## Comparison with Other Backends

| Aspect | Gemini | Claude | Codex |
|--------|--------|--------|-------|
| Knowledge breadth | Excellent | Good | Good |
| Code generation | Good | Excellent | Excellent |
| Explanations | Excellent | Excellent | Good |
| Speed | Fast | Moderate | Fast |

## Workflow Example

Use Gemini for research and documentation:

```json
{
  "steps": [
    {
      "name": "research",
      "backend": "gemini",
      "prompt": "research best practices for authentication in Go"
    },
    {
      "name": "implement",
      "backend": "claude",
      "prompt": "implement authentication based on: {{previous}}"
    },
    {
      "name": "document",
      "backend": "gemini",
      "prompt": "write documentation for: {{previous}}"
    }
  ]
}
```

## Troubleshooting

### Backend Not Available

```bash
# Check if Gemini is installed
which gemini

# Check clinvk detection
clinvk config show | grep gemini
```

### Authentication

Ensure you have valid Google Cloud credentials configured for the Gemini CLI.

## Next Steps

- [Claude Code Guide](claude.md)
- [Codex CLI Guide](codex.md)
- [Backend Comparison](../backend-comparison.md)
