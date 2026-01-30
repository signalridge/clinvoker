# Quick Start

Get up and running with clinvk in just a few minutes.

## Your First Prompt

Run a simple prompt using the default backend (Claude Code):

```bash
clinvk "explain what this project does"
```bash

## Specify a Backend

Use a specific backend with the `--backend` or `-b` flag:

```bash
# Use Claude Code
clinvk --backend claude "fix the bug in auth.go"

# Use Codex CLI
clinvk -b codex "implement user registration"

# Use Gemini CLI
clinvk -b gemini "generate unit tests"
```

## Continue a Session

Resume your last session to continue the conversation:

```bash
# Continue with a follow-up prompt
clinvk --continue "now add error handling"

# Or use the resume command
clinvk resume --last "add tests for the changes"
```bash

## Compare Backends

Get responses from multiple backends for the same prompt:

```bash
# Compare all enabled backends
clinvk compare --all-backends "what does this code do"

# Compare specific backends
clinvk compare --backends claude,codex "explain this algorithm"
```

## Run Tasks in Parallel

Execute multiple tasks concurrently. Create a `tasks.json` file:

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "review the auth module"},
    {"backend": "codex", "prompt": "add logging to the API"},
    {"backend": "gemini", "prompt": "generate tests for utils"}
  ]
}
```yaml

Run the tasks:

```bash
clinvk parallel --file tasks.json
```

## Chain Backends

Pass output through multiple backends sequentially. Create a `pipeline.json`:

```json
{
  "steps": [
    {"name": "review", "backend": "claude", "prompt": "review this code for bugs"},
    {"name": "security", "backend": "gemini", "prompt": "check for security issues in: {{previous}}"},
    {"name": "summary", "backend": "codex", "prompt": "summarize the findings: {{previous}}"}
  ]
}
```yaml

Run the chain:

```bash
clinvk chain --file pipeline.json
```

## Start the HTTP Server

Run clinvk as an HTTP API server:

```bash
# Start on default port (8080)
clinvk serve

# Custom port
clinvk serve --port 3000
```yaml

Then make API requests:

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello world"}'
```

## Common Options

| Option | Short | Description |
|--------|-------|-------------|
| `--backend` | `-b` | Backend to use (claude, codex, gemini) |
| `--model` | `-m` | Model to use |
| `--workdir` | `-w` | Working directory |
| `--output-format` | `-o` | Output format (text, json, stream-json) |
| `--continue` | `-c` | Continue last session |
| `--dry-run` | | Show command without executing |

## Next Steps

- [Basic Usage](basic-usage.md) - Detailed usage guide
- [Session Management](session-management.md) - Work with sessions
- [Configuration](../reference/configuration.md) - Customize settings
