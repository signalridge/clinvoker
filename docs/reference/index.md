# Reference

Technical reference documentation for clinvk.

## Commands

Complete documentation for all clinvk commands:

<div class="grid cards" markdown>

-   :material-console-line:{ .lg .middle } **[Commands](commands/index.md)**

    ---

    Full reference for all CLI commands

</div>

### Command List

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

<div class="grid cards" markdown>

-   :material-cog:{ .lg .middle } **[Configuration](configuration.md)**

    ---

    Complete configuration reference

-   :material-variable:{ .lg .middle } **[Environment Variables](environment.md)**

    ---

    Environment variable reference

-   :material-numeric:{ .lg .middle } **[Exit Codes](exit-codes.md)**

    ---

    Exit code meanings

</div>

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

| Backend | Binary | Models |
|---------|--------|--------|
| Claude | `claude` | claude-opus-4-5-20251101, claude-sonnet-4-20250514 |
| Codex | `codex` | o3, o3-mini |
| Gemini | `gemini` | gemini-2.5-pro, gemini-2.5-flash |

### Output Formats

| Format | Description |
|--------|-------------|
| `text` | Plain text output (default) |
| `json` | Structured JSON |
| `stream-json` | Streaming JSON events |
