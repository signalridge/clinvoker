# clinvk CLI Reference

Complete command reference for clinvk.

## Synopsis

```
clinvk [flags] [prompt]
clinvk [command]
```

## Global Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | AI backend to use (claude, codex, gemini) |
| `--model` | `-m` | string | | Model to use for the backend |
| `--workdir` | `-w` | string | | Working directory for the AI backend |
| `--output-format` | `-o` | string | `text` | Output format: text, json, stream-json |
| `--config` | | string | | Config file (default: ~/.clinvk/config.yaml) |
| `--dry-run` | | bool | `false` | Print command without executing |
| `--help` | `-h` | | | Help for clinvk |

### Root Command Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--continue` | `-c` | bool | `false` | Continue the last session |

## Commands

### clinvk [prompt]

Run a prompt with the default or specified backend.

```bash
clinvk "fix the bug in auth.go"
clinvk --backend codex "implement feature X"
clinvk -b gemini -m gemini-2.5-pro "explain this code"
```

### clinvk version

Display version information.

```bash
clinvk version
```

Output:

```
clinvk version v0.1.0
  commit: abc1234
  built:  2025-01-27T00:00:00Z
```

### clinvk resume

Resume a previous session.

**Usage:**

```
clinvk resume [session-id] [prompt] [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--last` | | bool | `false` | Resume the most recent session |
| `--interactive` | `-i` | bool | `false` | Show interactive session picker |
| `--here` | | bool | `false` | Filter sessions by current working directory |
| `--backend` | `-b` | string | | Filter sessions by backend |

**Examples:**

```bash
# Resume last session
clinvk resume --last

# Resume last session with a follow-up prompt
clinvk resume --last "continue from where we left off"

# Interactive session picker
clinvk resume --interactive

# Resume sessions from current directory only
clinvk resume --here

# Filter by backend
clinvk resume --backend claude

# Resume specific session
clinvk resume abc123

# Resume specific session with prompt
clinvk resume abc123 "now add tests"
```

### clinvk sessions

Manage sessions.

**Subcommands:**

#### clinvk sessions list

List all sessions.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | | Filter by backend |
| `--status` | | string | | Filter by status (active, completed, error) |
| `--limit` | `-n` | int | | Limit number of sessions shown |

**Examples:**

```bash
clinvk sessions list
clinvk sessions list --backend claude
clinvk sessions list --status active --limit 10
```

Output:

```
ID        BACKEND   STATUS     LAST USED       TOKENS       TITLE/PROMPT
abc123    claude    active     5 minutes ago   1,234        fix the bug in auth.go
def456    codex     completed  2 hours ago     5,678        implement user registration
```

#### clinvk sessions show

Show details of a specific session.

```bash
clinvk sessions show <session-id>
```

Output:

```
ID:                abc123
Backend:           claude
Model:             claude-opus-4-5-20251101
Status:            active
Created:           2025-01-27T10:00:00Z
Last Used:         2025-01-27T11:30:00Z (30 minutes ago)
Working Directory: /projects/myapp
Token Usage:
  Input:           1,234
  Output:          5,678
  Cached:          500
  Total:           6,912
```

#### clinvk sessions delete

Delete a session.

```bash
clinvk sessions delete <session-id>
```

#### clinvk sessions clean

Remove old sessions.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--older-than` | string | | Delete sessions older than this (e.g., 30d, 7d) |

If `--older-than` is not specified, uses the `session.retention_days` config value.

**Examples:**

```bash
# Delete sessions older than 30 days
clinvk sessions clean --older-than 30d

# Use config default retention
clinvk sessions clean
```

### clinvk config

Manage configuration.

#### clinvk config show

Display current configuration.

```bash
clinvk config show
```

#### clinvk config set

Set a configuration value.

```bash
clinvk config set <key> <value>
```

**Examples:**

```bash
clinvk config set default_backend gemini
clinvk config set backends.claude.model claude-opus-4-5-20251101
clinvk config set session.retention_days 60
```

### clinvk parallel

Execute multiple tasks in parallel.

**Usage:**

```
clinvk parallel [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--file` | `-f` | string | | JSON file containing task definitions |
| `--max-parallel` | | int | 3 | Maximum number of parallel tasks |
| `--fail-fast` | | bool | `false` | Stop all tasks on first failure |
| `--json` | | bool | `false` | Output results as JSON |
| `--quiet` | `-q` | bool | `false` | Suppress task output (show only results) |

**Task File Format:**

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "task prompt",
      "model": "optional-model",
      "workdir": "/optional/path",
      "approval_mode": "auto",
      "sandbox_mode": "workspace",
      "max_turns": 10
    }
  ],
  "max_parallel": 3,
  "fail_fast": true
}
```

**Task Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `backend` | string | Yes | Backend to use (claude, codex, gemini) |
| `prompt` | string | Yes | The prompt to execute |
| `model` | string | No | Model override |
| `workdir` | string | No | Working directory |
| `approval_mode` | string | No | Approval mode (default, auto, none, always) |
| `sandbox_mode` | string | No | Sandbox mode (default, read-only, workspace, full) |
| `output_format` | string | No | Output format (text, json, stream-json) |
| `max_tokens` | int | No | Max response tokens |
| `max_turns` | int | No | Max agentic turns |
| `system_prompt` | string | No | System prompt override |

**Examples:**

```bash
# From file
clinvk parallel --file tasks.json

# From stdin
cat tasks.json | clinvk parallel

# Limit parallel workers
clinvk parallel --file tasks.json --max-parallel 2

# With fail-fast
clinvk parallel --file tasks.json --fail-fast

# JSON output
clinvk parallel --file tasks.json --json
```

### clinvk compare

Compare responses from multiple backends.

**Usage:**

```
clinvk compare [prompt] [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--backends` | string | | Comma-separated list of backends to compare |
| `--all-backends` | bool | `false` | Compare all enabled backends |
| `--sequential` | bool | `false` | Run backends one at a time |
| `--json` | bool | `false` | Output results as JSON |

**Examples:**

```bash
# Compare specific backends
clinvk compare --backends claude,codex "explain this code"

# Compare all backends
clinvk compare --all-backends "what does this function do"

# Sequential execution
clinvk compare --all-backends --sequential "review this PR"

# JSON output
clinvk compare --all-backends --json "analyze performance"
```

**Output Format (JSON):**

```json
{
  "prompt": "explain this code",
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "response": "This code...",
      "duration_ms": 2500,
      "success": true
    },
    {
      "backend": "codex",
      "model": "o3",
      "response": "The code...",
      "duration_ms": 3200,
      "success": true
    }
  ]
}
```

### clinvk chain

Execute a pipeline of prompts through multiple backends.

**Usage:**

```
clinvk chain [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--file` | `-f` | string | | JSON file containing chain definition |
| `--json` | | bool | `false` | Output results as JSON |

**Chain File Format:**

```json
{
  "steps": [
    {
      "name": "step-name",
      "backend": "claude",
      "prompt": "First prompt",
      "model": "optional-model"
    },
    {
      "name": "second-step",
      "backend": "gemini",
      "prompt": "Process this: {{previous}}"
    }
  ]
}
```

**Step Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Step identifier |
| `backend` | string | Yes | Backend to use |
| `prompt` | string | Yes | The prompt (use `{{previous}}` or `{{session}}` for previous session ID) |
| `model` | string | No | Model override |
| `workdir` | string | No | Working directory |
| `approval_mode` | string | No | Approval mode |
| `sandbox_mode` | string | No | Sandbox mode |
| `max_turns` | int | No | Max agentic turns |

**Note:** The `{{previous}}` placeholder is replaced with the session ID from the previous step, allowing the backend to resume context from that session.

**Examples:**

```bash
# Execute chain
clinvk chain --file pipeline.json

# JSON output
clinvk chain --file pipeline.json --json
```

**Output Format (JSON):**

```json
{
  "steps": [
    {
      "name": "initial-review",
      "backend": "claude",
      "prompt": "Review this code",
      "output": "The code has...",
      "duration_ms": 2000,
      "success": true
    },
    {
      "name": "security-check",
      "backend": "gemini",
      "prompt": "Check for security issues in: The code has...",
      "output": "No security issues found.",
      "duration_ms": 1500,
      "success": true
    }
  ],
  "total_duration_ms": 3500
}
```

### clinvk serve

Start the HTTP API server.

**Usage:**

```
clinvk serve [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--host` | | string | `127.0.0.1` | Host to bind to |
| `--port` | `-p` | int | `8080` | Port to listen on |

**Examples:**

```bash
# Start with defaults
clinvk serve

# Custom port
clinvk serve --port 3000

# Bind to all interfaces
clinvk serve --host 0.0.0.0 --port 8080
```

**API Endpoints:**

The server provides three distinct API styles:

**Custom RESTful API (`/api/v1/`):**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/prompt` | Execute single prompt |
| POST | `/api/v1/parallel` | Execute multiple prompts in parallel |
| POST | `/api/v1/chain` | Execute prompts in sequence |
| POST | `/api/v1/compare` | Compare responses across backends |
| GET | `/api/v1/backends` | List available backends |
| GET | `/api/v1/sessions` | List sessions |
| GET | `/api/v1/sessions/{id}` | Get session details |
| DELETE | `/api/v1/sessions/{id}` | Delete session |

**OpenAI Compatible API (`/openai/v1/`):**

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/openai/v1/models` | List available models |
| POST | `/openai/v1/chat/completions` | Create chat completion |

**Anthropic Compatible API (`/anthropic/v1/`):**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/anthropic/v1/messages` | Create message |

**Meta Endpoints:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/openapi.json` | OpenAPI specification |

**Example API Requests:**

```bash
# Execute a prompt
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "explain this code"}'

# List backends
curl http://localhost:8080/api/v1/backends

# OpenAI-compatible chat completion
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Anthropic-compatible message
curl -X POST http://localhost:8080/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

**Server Configuration:**

In `~/.clinvk/config.yaml`:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Command line usage error |
| 126 | Backend not available |
| 127 | Backend command not found |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CLINVK_BACKEND` | Default backend | `claude` |
| `CLINVK_CLAUDE_MODEL` | Claude model | |
| `CLINVK_CODEX_MODEL` | Codex model | |
| `CLINVK_GEMINI_MODEL` | Gemini model | |

## Configuration File

Default location: `~/.clinvk/config.yaml`

See [Configuration Guide](CONFIGURATION.md) for detailed options.

## See Also

- [README](../README.md) - Quick start guide
- [Configuration Guide](CONFIGURATION.md) - Detailed configuration options
