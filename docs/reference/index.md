# Reference

Technical reference documentation for clinvk.

## Commands

Complete documentation for all clinvk commands:

| Command | Description |
|---------|-------------|
| [`clinvk [prompt]`](commands/prompt.md) | Execute a prompt |
| [`clinvk resume`](commands/resume.md) | Resume a session |
| [`clinvk sessions`](commands/sessions.md) | Manage sessions |
| [`clinvk config`](commands/config.md) | Manage configuration |
| [`clinvk parallel`](commands/parallel.md) | Parallel execution |
| [`clinvk compare`](commands/compare.md) | Backend comparison |
| [`clinvk chain`](commands/chain.md) | Chain execution |
| [`clinvk serve`](commands/serve.md) | HTTP API server |

## Configuration

- **[Configuration](configuration.md)** - Complete configuration reference
- **[Environment Variables](environment.md)** - Environment variable reference
- **[Exit Codes](exit-codes.md)** - Exit code meanings

## Configuration Priority

Configuration values are resolved in this order (highest to lowest priority):

1. **CLI Flags** - Command-line arguments
2. **Environment Variables** - `CLINVK_*` variables
3. **Config File** - `~/.clinvk/config.yaml`
4. **Defaults** - Built-in default values

## Quick Reference

### Common Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--backend` | `-b` | Backend to use |
| `--model` | `-m` | Model to use |
| `--workdir` | `-w` | Working directory |
| `--output-format` | `-o` | Output format |
| `--continue` | `-c` | Continue session |
| `--dry-run` | | Show command only |

### Backends

| Backend | Binary | Example Models |
|---------|--------|--------|
| Claude | `claude` | (backend default or configured) |
| Codex | `codex` | (backend default or configured) |
| Gemini | `gemini` | (backend default or configured) |

### Output Formats

| Format | Description |
|--------|-------------|
| `text` | Plain text output |
| `json` | Structured JSON (default via `output.format`) |
| `stream-json` | Streaming JSON (backend-specific) |
