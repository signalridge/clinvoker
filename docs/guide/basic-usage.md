# Basic Usage

Learn the fundamentals of using clinvk for everyday tasks.

## Running Prompts

The simplest way to use clinvk is to run a prompt with the default backend:

```bash
clinvk "your prompt here"
```

### Specifying a Backend

Use the `--backend` (or `-b`) flag to choose a specific backend:

```bash
clinvk --backend claude "fix the bug in auth.go"
clinvk -b codex "implement user registration"
clinvk -b gemini "explain this algorithm"
```

### Specifying a Model

Override the default model with `--model` (or `-m`):

```bash
clinvk --model claude-opus-4-5-20251101 "complex task"
clinvk -b codex -m o3 "implement feature"
```

### Working Directory

Set the working directory for the AI to operate in:

```bash
clinvk --workdir /path/to/project "review the codebase"
clinvk -w ./subproject "fix tests"
```

## Output Formats

Control how output is displayed:

### Text (Default)

```bash
clinvk "explain this code"
```

### JSON

```bash
clinvk --output-format json "explain this code"
```

### Streaming JSON

```bash
clinvk -o stream-json "explain this code"
```

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
```

See [Session Management](session-management.md) for more details.

## Dry Run Mode

Preview the command without executing:

```bash
clinvk --dry-run "implement feature X"
```

Output shows the exact command that would be run:

```yaml
Would execute: claude --model claude-opus-4-5-20251101 "implement feature X"
```

## Ephemeral Mode

Run in stateless mode without creating a session:

```bash
clinvk --ephemeral "what is 2+2"
```

This is useful for quick one-off queries where you don't need conversation history.

## Global Flags Summary

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--backend` | `-b` | AI backend to use | `claude` |
| `--model` | `-m` | Model to use | (backend default) |
| `--workdir` | `-w` | Working directory | (current dir) |
| `--output-format` | `-o` | Output format | `json` |
| `--continue` | `-c` | Continue last session | `false` |
| `--dry-run` | | Show command only | `false` |
| `--ephemeral` | | Stateless mode | `false` |
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

### Refactoring

```bash
clinvk "refactor the database module to use connection pooling"
clinvk -c "now add unit tests for the changes"
```

## Next Steps

- [Session Management](session-management.md) - Work with sessions effectively
- [Backend Comparison](backend-comparison.md) - Get multiple perspectives
- [Configuration](../reference/configuration.md) - Customize your setup
