# Basic Usage

Learn the fundamentals of using clinvk for everyday tasks. This guide covers the command structure, global flags, backend selection strategies, and output formats in detail.

## Command Structure Overview

clinvk follows a consistent command structure:

```bash
clinvk [global-flags] [command] [command-flags] [arguments]
```

### Command Types

| Type | Example | Description |
|------|---------|-------------|
| **Direct prompt** | `clinvk "fix the bug"` | Run a prompt with the default backend |
| **Subcommand** | `clinvk sessions list` | Execute a specific command |
| **Resume** | `clinvk resume --last` | Resume a previous session |

## Global Flags Explained

Global flags affect how clinvk operates regardless of the command being run. They can be specified before any subcommand.

### Backend Selection (`--backend`, `-b`)

The `--backend` flag determines which AI backend processes your prompt.

```bash
# Use specific backends
clinvk --backend claude "fix the bug in auth.go"
clinvk -b codex "implement user registration"
clinvk -b gemini "explain this algorithm"
```

**Backend Selection Strategy:**

| Task Type | Recommended Backend | Reason |
|-----------|---------------------|--------|
| Complex reasoning | `claude` | Deep context understanding, safety focus |
| Code generation | `codex` | Optimized for programming tasks |
| Documentation | `gemini` | Broad knowledge, clear explanations |
| Security review | `claude` | Thorough analysis, risk assessment |
| Quick prototypes | `codex` | Fast code generation |

When no backend is specified, clinvk uses the `default_backend` from configuration (defaults to `claude`).

### Model Selection (`--model`, `-m`)

Override the default model for the selected backend:

```bash
clinvk --model claude-opus-4-5-20251101 "complex architecture task"
clinvk -b codex -m o3 "implement feature"
clinvk -b gemini -m gemini-2.5-flash "quick question"
```

**When to Override Models:**

- Use larger models (Opus, o3, Pro) for complex tasks requiring deep reasoning
- Use smaller models (Sonnet, o3-mini, Flash) for faster, simpler tasks
- Consider cost and latency trade-offs

### Working Directory (`--workdir`, `-w`)

Set the working directory for the AI to operate in:

```bash
clinvk --workdir /path/to/project "review the codebase"
clinvk -w ./subproject "fix tests"
```

**Working Directory Behavior:**

- The AI receives the specified directory as its working context
- File operations are relative to this directory
- Different backends handle sandboxing differently (see [Backend Guides](backends/index.md))
- Use absolute paths for clarity in scripts

**Security Considerations:**

```bash
# Good: Explicit, limited scope
clinvk -w /home/user/projects/myapp "analyze code"

# Risky: Full system access (depends on backend sandbox mode)
clinvk -w / "search for files"
```

### Output Format (`--output-format`, `-o`)

Control how output is displayed. The effective default comes from `output.format` in config (built-in default is `json`).

#### Text Format

Human-readable output with formatting:

```bash
clinvk --output-format text "explain this code"
```

**Best for:** Interactive use, reading in terminal, quick checks

#### JSON Format

Structured output for programmatic processing:

```bash
clinvk --output-format json "explain this code"
```

**Output structure:**

```json
{
  "output": "The code implements...",
  "backend": "claude",
  "model": "claude-opus-4-5-20251101",
  "duration_seconds": 2.5,
  "exit_code": 0
}
```

**Best for:** Scripting, CI/CD pipelines, storing results, further processing

#### Streaming JSON Format

```bash
clinvk -o stream-json "explain this code"
```

`stream-json` passes through the backend's native streaming output (NDJSON/JSONL). This provides real-time updates as the AI generates content.

**Best for:** Long-running tasks, real-time monitoring, building interactive tools

**Format Comparison:**

| Format | Human-Readable | Machine-Parsable | Streaming | Use Case |
|--------|---------------|------------------|-----------|----------|
| `text` | Yes | No | No | Interactive use |
| `json` | Somewhat | Yes | No | Scripting, storage |
| `stream-json` | Somewhat | Yes | Yes | Real-time apps |

### Continue Mode (`--continue`, `-c`)

Continue the last session without specifying a session ID:

```bash
clinvk "implement the login feature"
clinvk -c "now add password validation"
clinvk -c "add rate limiting"
```

**How Continue Works:**

1. clinvk looks up the most recent resumable session
2. Appends the new prompt to the conversation history
3. The AI has full context of the previous interaction

**Session Requirements:**

- The previous session must have a backend session ID
- Sessions created with `--ephemeral` cannot be continued
- Only sessions from the same backend can be continued

### Ephemeral Mode (`--ephemeral`)

Run in stateless mode without creating a session:

```bash
clinvk --ephemeral "what is 2+2"
```

**When to Use Ephemeral Mode:**

| Scenario | Why Ephemeral? |
|----------|----------------|
| Quick one-off queries | No need for history |
| CI/CD scripts | Avoid session accumulation |
| Testing/debugging | Clean state every time |
| Public/shared systems | Privacy, no data retention |
| High-volume automation | Reduce storage overhead |

**Trade-offs:**

- **Pros:** No storage, faster execution, privacy
- **Cons:** No conversation history, cannot resume

### Dry Run Mode (`--dry-run`)

Preview the command without executing:

```bash
clinvk --dry-run "implement feature X"
```

**Output shows the exact command that would be run:**

```yaml
Would execute: claude --model claude-opus-4-5-20251101 "implement feature X"
```

**Use Cases:**

- Verify configuration before running expensive operations
- Debug flag parsing and backend selection
- Document expected behavior
- Test in CI/CD without making actual API calls

### Verbose Mode (`--verbose`, `-v`)

Enable detailed logging:

```bash
clinvk --verbose "complex task"
```

**Shows:**

- Configuration loading details
- Backend detection information
- Command construction steps
- API calls and responses (depending on backend)

## Exit Codes Reference

clinvk uses standard exit codes for scripting:

| Code | Meaning | When It Occurs |
|------|---------|----------------|
| 0 | Success | Command completed successfully |
| 1 | General error | CLI error, validation failure, backend error |
| Backend code | Propagated | Backend's own exit code (for prompt/resume) |

**Command-Specific Exit Codes:**

| Command | Exit Code 0 | Exit Code 1 |
|---------|-------------|-------------|
| `prompt` | Success | Backend error |
| `parallel` | All tasks succeeded | One or more tasks failed |
| `compare` | All backends succeeded | One or more backends failed |
| `chain` | All steps succeeded | A step failed |
| `serve` | Clean shutdown | Server error |

**Scripting Example:**

```bash
#!/bin/bash

clinvk "implement feature"
exit_code=$?

case $exit_code in
  0)
    echo "Success - feature implemented"
    ;;
  1)
    echo "Failed - check logs"
    exit 1
    ;;
  *)
    echo "Backend returned code: $exit_code"
    ;;
esac
```

## Environment Variables

Override configuration with environment variables:

```bash
# Set default backend
export CLINVK_BACKEND=codex

# Set models per backend
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3-mini
export CLINVK_GEMINI_MODEL=gemini-2.5-flash

# Run with environment settings
clinvk "prompt"  # Uses codex with o3-mini
```

**Priority Order** (highest to lowest):

1. CLI flags (`--backend codex`)
2. Environment variables (`CLINVK_BACKEND`)
3. Config file (`~/.clinvk/config.yaml`)
4. Built-in defaults

## Continuing Conversations

### Quick Continue

Use `--continue` (or `-c`) to continue the last session:

```bash
clinvk "implement the login feature"
clinvk -c "now add password validation"
clinvk -c "add rate limiting"
```

### Resume Command

For more control, use the `resume` command:

```bash
# Resume last session
clinvk resume --last

# Interactive session picker
clinvk resume --interactive

# Resume with a specific prompt
clinvk resume --last "continue from where we left off"

# Resume by ID
clinvk resume abc123 "add tests"
```

See [Session Management](sessions.md) for complete details.

## Global Flags Summary

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--backend` | `-b` | AI backend to use | `claude` |
| `--model` | `-m` | Model to use | (backend default) |
| `--workdir` | `-w` | Working directory | (current dir) |
| `--output-format` | `-o` | Output format | `json` (configurable) |
| `--continue` | `-c` | Continue last session | `false` |
| `--dry-run` | | Show command only | `false` |
| `--ephemeral` | | Stateless mode | `false` |
| `--verbose` | `-v` | Enable verbose logging | `false` |
| `--config` | | Config file path | `~/.clinvk/config.yaml` |

## Examples

### Quick Bug Fix

```bash
clinvk "there's a null pointer exception in utils.go line 45"
```

### Code Generation

```bash
clinvk -b codex "generate a REST API handler for user CRUD operations"
```

### Code Explanation

```bash
clinvk -b gemini "explain what the main function in cmd/server/main.go does"
```

### Refactoring with Continuation

```bash
clinvk "refactor the database module to use connection pooling"
clinvk -c "now add unit tests for the changes"
clinvk -c "update the documentation"
```

### CI/CD Integration

```bash
# Non-interactive mode with JSON output
clinvk --ephemeral --output-format json \
  --backend codex \
  "generate tests for the auth module"
```

### Multi-Step Workflow

```bash
#!/bin/bash

# Step 1: Analyze
clinvk -o json "analyze the codebase architecture" > analysis.json

# Step 2: Generate based on analysis
clinvk -c "implement the recommended changes"

# Step 3: Verify
clinvk -c "run the tests and fix any failures"
```

## Common Patterns

### Pattern 1: Explore with Text, Automate with JSON

```bash
# Interactive exploration - use text
clinvk -o text "explain this module"

# Once satisfied, switch to JSON for automation
clinvk -o json --ephemeral "generate the implementation"
```

### Pattern 2: Backend per Task Type

```bash
# Architecture decisions - Claude
clinvk -b claude "design the API structure"

# Implementation - Codex
clinvk -b codex "implement the endpoints"

# Documentation - Gemini
clinvk -b gemini "write API documentation"
```

### Pattern 3: Dry Run Before Execution

```bash
# Verify what will happen
clinvk --dry-run --backend codex "refactor the entire codebase"

# If satisfied, run for real
clinvk --backend codex "refactor the entire codebase"
```

## Troubleshooting

### Backend Not Found

```bash
# Check available backends
clinvk config show | grep available

# Verify CLI installation
which claude codex gemini
```

### Configuration Not Applied

```bash
# Check effective configuration
clinvk config show

# Verify file exists
ls -la ~/.clinvk/config.yaml
```

### Session Not Resuming

```bash
# List available sessions
clinvk sessions list

# Check if session has backend ID
clinvk sessions show <session-id>
```

## Next Steps

- [Session Management](sessions.md) - Work with sessions effectively
- [Backend Comparison](compare.md) - Get multiple perspectives
- [Configuration](../reference/configuration.md) - Customize your setup
- [Parallel Execution](parallel.md) - Run multiple tasks concurrently
