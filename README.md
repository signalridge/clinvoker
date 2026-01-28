# clinvk

Unified AI CLI wrapper for orchestrating multiple AI CLI backends (Claude Code, Codex CLI, Gemini CLI) with session persistence, parallel task execution, HTTP API server, and unified output formatting.

## Features

- **Multi-Backend Support**: Seamlessly switch between Claude Code, Codex CLI, and Gemini CLI
- **Unified Options**: Consistent configuration options work across all backends via config file or task definitions
- **Session Persistence**: Automatic session tracking with resume capability
- **Parallel Execution**: Run multiple AI tasks concurrently with fail-fast support
- **Backend Comparison**: Compare responses from multiple backends side-by-side
- **Chain Execution**: Pipeline prompts through multiple backends sequentially
- **HTTP API Server**: RESTful API with OpenAI and Anthropic compatible endpoints
- **Configuration Cascade**: CLI flags → Environment variables → Config file → Defaults
- **Cross-Platform**: Supports Linux, macOS, and Windows

## Installation

### From Release

Download the latest release from [GitHub Releases](https://github.com/signalridge/clinvoker/releases).

### From Source

```bash
go install github.com/signalridge/clinvoker/cmd/clinvk@latest
```

### Homebrew (macOS/Linux)

```bash
brew install signalridge/tap/clinvk
```

### Nix

```bash
# Run directly
nix run github:signalridge/clinvoker

# Install to profile
nix profile install github:signalridge/clinvoker

# Development shell
nix develop github:signalridge/clinvoker
```

Add to your flake:

```nix
{
  inputs.clinvoker.url = "github:signalridge/clinvoker";

  # Use overlay
  nixpkgs.overlays = [ clinvoker.overlays.default ];
}
```

### Arch Linux (AUR)

```bash
# Using yay
yay -S clinvk-bin

# Or build from source
yay -S clinvk
```

### Scoop (Windows)

```bash
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket
scoop install clinvk
```

### Debian/Ubuntu (apt)

```bash
# Download the .deb package from releases
sudo dpkg -i clinvk_*.deb
```

### RPM-based (Fedora/RHEL)

```bash
# Download the .rpm package from releases
sudo rpm -i clinvk_*.rpm
```

## Quick Start

```bash
# Run with default backend (Claude Code)
clinvk "fix the bug in auth.go"

# Specify a backend
clinvk --backend codex "implement user registration"
clinvk -b gemini "generate unit tests"

# Resume a session
clinvk resume --last "continue working"
clinvk resume abc123 "follow up"

# Compare backends
clinvk compare --all-backends "explain this code"

# Chain backends
clinvk chain --file pipeline.json

# Start HTTP API server
clinvk serve --port 8080
```

## Usage

### Basic Commands

```bash
# Run a prompt
clinvk "your prompt here"

# With specific backend and model
clinvk --backend codex --model o3 "your prompt"

# Dry run (show command without executing)
clinvk --dry-run "your prompt"

# Version info
clinvk version
```

### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--backend` | `-b` | AI backend to use | `claude` |
| `--model` | `-m` | Model to use | (backend default) |
| `--workdir` | `-w` | Working directory | (current dir) |
| `--output-format` | `-o` | Output format: `text`, `json`, `stream-json` | `text` |
| `--continue` | `-c` | Continue the last session | `false` |
| `--dry-run` | | Print command without executing | `false` |

Example:

```bash
# JSON output format
clinvk --output-format json "refactor auth module"

# Continue last session
clinvk -c "follow up on previous task"
```

### Unified Options (Config/Tasks)

These options can be set in the config file (`unified_flags` section) or per-task in parallel/chain definitions:

| Option | Description | Values |
|--------|-------------|--------|
| `approval_mode` | Approval mode | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | Sandbox mode | `default`, `read-only`, `workspace`, `full` |
| `output_format` | Output format | `text`, `json`, `stream-json` |
| `max_turns` | Max agentic turns | integer |
| `max_tokens` | Max response tokens | integer |

### Session Management

```bash
# List all sessions
clinvk sessions list
clinvk sessions list --backend claude --limit 10

# Show session details
clinvk sessions show <session-id>

# Resume last session
clinvk resume --last

# Resume with interactive picker
clinvk resume --interactive

# Resume specific session
clinvk resume <session-id> "follow up prompt"

# Resume sessions from current directory only
clinvk resume --here

# Delete a session
clinvk sessions delete <session-id>

# Clean old sessions
clinvk sessions clean --older-than 30d
```

### Parallel Execution

Create a `tasks.json` file:

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "review auth module",
      "model": "claude-opus-4-5-20251101",
      "approval_mode": "auto"
    },
    {
      "backend": "codex",
      "prompt": "add logging to api"
    },
    {
      "backend": "gemini",
      "prompt": "generate tests for utils"
    }
  ],
  "max_parallel": 3
}
```

Run tasks:

```bash
# From file
clinvk parallel --file tasks.json

# From stdin
cat tasks.json | clinvk parallel

# Limit parallel workers
clinvk parallel --file tasks.json --max-parallel 2

# With fail-fast (stop on first failure)
clinvk parallel --file tasks.json --fail-fast

# JSON output
clinvk parallel --file tasks.json --json
```

### Backend Comparison

Compare responses from multiple backends:

```bash
# Compare specific backends
clinvk compare --backends claude,codex "explain this algorithm"

# Compare all enabled backends
clinvk compare --all-backends "what does this code do"

# Sequential execution (one at a time)
clinvk compare --all-backends --sequential "review this PR"

# JSON output for programmatic processing
clinvk compare --all-backends --json "analyze performance"
```

### Chain Execution

Create a pipeline that passes output through multiple backends:

Create `pipeline.json`:

```json
{
  "steps": [
    {
      "name": "initial-review",
      "backend": "claude",
      "prompt": "Review this code for bugs"
    },
    {
      "name": "security-check",
      "backend": "gemini",
      "prompt": "Check for security issues in: {{previous}}"
    },
    {
      "name": "final-summary",
      "backend": "codex",
      "prompt": "Summarize the findings: {{previous}}"
    }
  ]
}
```

Run chain:

```bash
clinvk chain --file pipeline.json

# JSON output
clinvk chain --file pipeline.json --json
```

### HTTP API Server

Start the HTTP API server to access AI backends via REST APIs:

```bash
# Start with default settings (127.0.0.1:8080)
clinvk serve

# Custom port
clinvk serve --port 3000

# Bind to all interfaces
clinvk serve --host 0.0.0.0 --port 8080
```

The server provides three API styles:

**Custom RESTful API** (`/api/v1/`):
- `POST /api/v1/prompt` - Execute single prompt
- `POST /api/v1/parallel` - Execute multiple prompts in parallel
- `POST /api/v1/chain` - Execute prompts in sequence
- `POST /api/v1/compare` - Compare responses across backends
- `GET /api/v1/backends` - List available backends
- `GET /api/v1/sessions` - List sessions
- `GET /api/v1/sessions/{id}` - Get session details
- `DELETE /api/v1/sessions/{id}` - Delete session

**OpenAI Compatible API** (`/openai/v1/`):
- `GET /openai/v1/models` - List available models
- `POST /openai/v1/chat/completions` - Create chat completion

**Anthropic Compatible API** (`/anthropic/v1/`):
- `POST /anthropic/v1/messages` - Create message

**Meta Endpoints**:
- `GET /health` - Health check
- `GET /openapi.json` - OpenAPI specification

Example API usage:

```bash
# Execute a prompt via API
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "explain this code"}'

# Use OpenAI-compatible endpoint
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Configuration

```bash
# Show current configuration
clinvk config show

# Set default backend
clinvk config set default_backend gemini

# Set backend-specific model
clinvk config set backends.claude.model claude-opus-4-5-20251101
```

## Configuration

Configuration is stored in `~/.clinvk/config.yaml`:

```yaml
default_backend: claude

unified_flags:
  approval_mode: default
  sandbox_mode: default
  output_format: default
  verbose: false
  max_turns: 0
  max_tokens: 0

backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
    approval_mode: ""      # Override unified setting
    sandbox_mode: ""       # Override unified setting
    enabled: true
    system_prompt: ""
    extra_flags: []
  codex:
    model: o3
    enabled: true
  gemini:
    model: gemini-2.5-pro
    enabled: true

session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []

output:
  format: text
  show_tokens: false
  show_timing: false
  color: true

server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300

parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLINVK_BACKEND` | Default backend |
| `CLINVK_CLAUDE_MODEL` | Claude model |
| `CLINVK_CODEX_MODEL` | Codex model |
| `CLINVK_GEMINI_MODEL` | Gemini model |

## Backends

| Backend | Binary | Resume Flag | Key Flags |
|---------|--------|-------------|-----------|
| Claude Code | `claude` | `--resume` | `--allowedTools`, `--model`, `--add-dir` |
| Codex CLI | `codex` | `--session` | `--model`, `--quiet` |
| Gemini CLI | `gemini` | `-s` | `--model`, `--sandbox` |

### Backend Availability

clinvk automatically detects which backends are available in your PATH. Use `clinvk config show` to see which backends are detected.

## Development

### Prerequisites

- Go 1.24+
- (Optional) golangci-lint for linting

### Building

```bash
go build ./cmd/clinvk
```

### Testing

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...
```

### Linting

```bash
golangci-lint run
```

### Using Nix

```bash
# Enter development shell
nix develop

# Build with Nix
nix build
```

## Documentation

- [CLI Reference](docs/CLI.md) - Complete command reference
- [Configuration Guide](docs/CONFIGURATION.md) - Detailed configuration options

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

## License

MIT License - see [LICENSE](LICENSE) for details.
