# clinvoker CLI Reference

Complete command reference for clinvoker.

## Synopsis

```
clinvoker [flags] [prompt]
clinvoker [command]
```

## Global Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | `claude` | AI backend to use (claude, codex, gemini) |
| `--model` | `-m` | string | | Model to use for the backend |
| `--workdir` | `-w` | string | | Working directory for the AI backend |
| `--config` | | string | | Config file (default: ~/.clinvoker/config.yaml) |
| `--dry-run` | | bool | `false` | Print command without executing |
| `--help` | `-h` | | | Help for clinvoker |

## Commands

### clinvoker [prompt]

Run a prompt with the default or specified backend.

```bash
clinvoker "fix the bug in auth.go"
clinvoker --backend codex "implement feature X"
clinvoker -b gemini -m gemini-2.5-pro "explain this code"
```

### clinvoker version

Display version information.

```bash
clinvoker version
```

Output:

```
clinvoker version v0.1.0
  commit: abc1234
  built:  2025-01-27T00:00:00Z
```

### clinvoker resume

Resume a previous session.

**Usage:**

```
clinvoker resume [session-id] [prompt] [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--last` | `-l` | bool | `false` | Resume the most recent session |

**Examples:**

```bash
# Resume last session
clinvoker resume --last

# Resume last session with a follow-up prompt
clinvoker resume --last "continue from where we left off"

# Resume specific session
clinvoker resume abc123

# Resume specific session with prompt
clinvoker resume abc123 "now add tests"
```

### clinvoker sessions

Manage sessions.

**Subcommands:**

#### clinvoker sessions list

List all sessions.

```bash
clinvoker sessions list
```

Output:

```
ID        Backend   Created              Last Used            Status     WorkDir
abc123    claude    2025-01-27 10:00:00  2025-01-27 11:30:00  active     /projects/myapp
def456    codex     2025-01-26 15:00:00  2025-01-26 16:00:00  completed  /projects/api
```

#### clinvoker sessions show

Show details of a specific session.

```bash
clinvoker sessions show <session-id>
```

Output:

```
Session: abc123
  Backend:      claude
  Created:      2025-01-27 10:00:00
  Last Used:    2025-01-27 11:30:00
  Status:       active
  Working Dir:  /projects/myapp
  Tags:         important, feature
  Token Usage:
    Input:      1,234
    Output:     5,678
```

#### clinvoker sessions delete

Delete a session.

```bash
clinvoker sessions delete <session-id>
```

#### clinvoker sessions clean

Remove old sessions.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--older-than` | duration | | Delete sessions older than this (e.g., 30d, 7d, 24h) |
| `--status` | string | | Delete sessions with this status (completed, failed) |

**Examples:**

```bash
# Delete sessions older than 30 days
clinvoker sessions clean --older-than 30d

# Delete completed sessions older than 7 days
clinvoker sessions clean --older-than 7d --status completed
```

### clinvoker config

Manage configuration.

#### clinvoker config show

Display current configuration.

```bash
clinvoker config show
```

#### clinvoker config set

Set a configuration value.

```bash
clinvoker config set <key> <value>
```

**Examples:**

```bash
clinvoker config set default_backend gemini
clinvoker config set backends.claude.model claude-opus-4-5-20251101
clinvoker config set session.retention_days 60
```

### clinvoker parallel

Execute multiple tasks in parallel.

**Usage:**

```
clinvoker parallel [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--file` | `-f` | string | | JSON file containing task definitions |
| `--fail-fast` | | bool | `false` | Stop all tasks on first failure |
| `--json` | | bool | `false` | Output results as JSON |
| `--quiet` | `-q` | bool | `false` | Minimal output |

**Task File Format:**

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "task prompt",
      "model": "optional-model",
      "work_dir": "/optional/path",
      "approval_mode": "auto",
      "sandbox_mode": "workspace"
    }
  ],
  "max_parallel": 3
}
```

**Task Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `backend` | string | Yes | Backend to use (claude, codex, gemini) |
| `prompt` | string | Yes | The prompt to execute |
| `model` | string | No | Model override |
| `work_dir` | string | No | Working directory |
| `approval_mode` | string | No | Approval mode (default, auto, none, always) |
| `sandbox_mode` | string | No | Sandbox mode (default, read-only, workspace, full) |

**Examples:**

```bash
# From file
clinvoker parallel --file tasks.json

# From stdin
cat tasks.json | clinvoker parallel

# With fail-fast
clinvoker parallel --file tasks.json --fail-fast

# JSON output
clinvoker parallel --file tasks.json --json
```

### clinvoker compare

Compare responses from multiple backends.

**Usage:**

```
clinvoker compare [prompt] [flags]
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
clinvoker compare --backends claude,codex "explain this code"

# Compare all backends
clinvoker compare --all-backends "what does this function do"

# Sequential execution
clinvoker compare --all-backends --sequential "review this PR"

# JSON output
clinvoker compare --all-backends --json "analyze performance"
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

### clinvoker chain

Execute a pipeline of prompts through multiple backends.

**Usage:**

```
clinvoker chain [flags]
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
| `prompt` | string | Yes | The prompt (use `{{previous}}` for previous output) |
| `model` | string | No | Model override |

**Examples:**

```bash
# Execute chain
clinvoker chain --file pipeline.json

# JSON output
clinvoker chain --file pipeline.json --json
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
| `CLINVOKER_BACKEND` | Default backend | `claude` |
| `CLINVOKER_CLAUDE_MODEL` | Claude model | |
| `CLINVOKER_CODEX_MODEL` | Codex model | |
| `CLINVOKER_GEMINI_MODEL` | Gemini model | |

## Configuration File

Default location: `~/.clinvoker/config.yaml`

See [Configuration Guide](CONFIGURATION.md) for detailed options.

## See Also

- [README](../README.md) - Quick start guide
- [Configuration Guide](CONFIGURATION.md) - Detailed configuration options
