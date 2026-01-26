# clinvoker

Unified AI CLI wrapper for orchestrating multiple AI CLI backends (Claude Code, Codex CLI, Gemini CLI) with session persistence, parallel task execution, and unified output formatting.

## Features

- **Multi-Backend Support**: Seamlessly switch between Claude Code, Codex CLI, and Gemini CLI
- **Unified Flags**: Common flags (`--approval`, `--sandbox`, `--output`) work across all backends
- **Session Persistence**: Automatic session tracking with resume, fork, and tagging
- **Parallel Execution**: Run multiple AI tasks concurrently with fail-fast support
- **Backend Comparison**: Compare responses from multiple backends side-by-side
- **Chain Execution**: Pipeline prompts through multiple backends sequentially
- **Configuration Cascade**: CLI flags → Environment variables → Config file → Defaults
- **Cross-Platform**: Supports Linux, macOS, and Windows

## Installation

### From Release

Download the latest release from [GitHub Releases](https://github.com/signalridge/clinvoker/releases).

### From Source

```bash
go install github.com/signalridge/clinvoker/cmd/clinvoker@latest
```

### Homebrew (macOS/Linux)

```bash
brew install signalridge/tap/clinvoker
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
yay -S clinvoker-bin

# Or build from source
yay -S clinvoker
```

### Scoop (Windows)

```powershell
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket
scoop install clinvoker
```

### Debian/Ubuntu (apt)

```bash
# Download the .deb package from releases
sudo dpkg -i clinvoker_*.deb
```

### RPM-based (Fedora/RHEL)

```bash
# Download the .rpm package from releases
sudo rpm -i clinvoker_*.rpm
```

## Quick Start

```bash
# Run with default backend (Claude Code)
clinvoker "fix the bug in auth.go"

# Specify a backend
clinvoker --backend codex "implement user registration"
clinvoker -b gemini "generate unit tests"

# Resume a session
clinvoker resume --last "continue working"
clinvoker resume abc123 "follow up"

# Compare backends
clinvoker compare --all-backends "explain this code"

# Chain backends
clinvoker chain --file pipeline.json
```

## Usage

### Basic Commands

```bash
# Run a prompt
clinvoker "your prompt here"

# With specific backend and model
clinvoker --backend codex --model o3 "your prompt"

# Dry run (show command without executing)
clinvoker --dry-run "your prompt"

# Version info
clinvoker version
```

### Unified Flags

These flags work consistently across all backends:

| Flag | Description | Values |
|------|-------------|--------|
| `--approval` | Approval mode | `default`, `auto`, `none`, `always` |
| `--sandbox` | Sandbox mode | `default`, `read-only`, `workspace`, `full` |
| `--output` | Output format | `default`, `text`, `json`, `stream-json` |
| `--verbose` | Verbose output | boolean |
| `--max-turns` | Max agentic turns | integer |
| `--max-tokens` | Max response tokens | integer |

Example:

```bash
# Auto-approve all tool calls with JSON output
clinvoker --approval auto --output json "refactor auth module"
```

### Session Management

```bash
# List all sessions
clinvoker sessions list

# Show session details
clinvoker sessions show <session-id>

# Resume last session
clinvoker resume --last

# Resume specific session
clinvoker resume <session-id> "follow up prompt"

# Fork a session (create a branch)
clinvoker sessions fork <session-id>

# Tag a session
clinvoker sessions tag <session-id> important

# Delete a session
clinvoker sessions delete <session-id>

# Clean old sessions
clinvoker sessions clean --older-than 30d
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
clinvoker parallel --file tasks.json

# From stdin
cat tasks.json | clinvoker parallel

# With fail-fast (stop on first failure)
clinvoker parallel --file tasks.json --fail-fast

# JSON output
clinvoker parallel --file tasks.json --json
```

### Backend Comparison

Compare responses from multiple backends:

```bash
# Compare specific backends
clinvoker compare --backends claude,codex "explain this algorithm"

# Compare all enabled backends
clinvoker compare --all-backends "what does this code do"

# Sequential execution (one at a time)
clinvoker compare --all-backends --sequential "review this PR"

# JSON output for programmatic processing
clinvoker compare --all-backends --json "analyze performance"
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
clinvoker chain --file pipeline.json

# JSON output
clinvoker chain --file pipeline.json --json
```

### Configuration

```bash
# Show current configuration
clinvoker config show

# Set default backend
clinvoker config set default_backend gemini

# Set backend-specific model
clinvoker config set backends.claude.model claude-opus-4-5-20251101
```

## Configuration

Configuration is stored in `~/.clinvoker/config.yaml`:

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

parallel:
  max_workers: 3
  fail_fast: false
  aggregate_output: true
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLINVOKER_BACKEND` | Default backend |
| `CLINVOKER_CLAUDE_MODEL` | Claude model |
| `CLINVOKER_CODEX_MODEL` | Codex model |
| `CLINVOKER_GEMINI_MODEL` | Gemini model |

## Backends

| Backend | Binary | Resume Flag | Key Flags |
|---------|--------|-------------|-----------|
| Claude Code | `claude` | `--resume` | `--allowedTools`, `--model`, `--add-dir` |
| Codex CLI | `codex` | `--session` | `--model`, `--quiet` |
| Gemini CLI | `gemini` | `-s` | `--model`, `--sandbox` |

### Backend Availability

clinvoker automatically detects which backends are available in your PATH. Use `clinvoker config show` to see which backends are detected.

## Development

### Prerequisites

- Go 1.24+
- (Optional) golangci-lint for linting

### Building

```bash
go build ./cmd/clinvoker
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
