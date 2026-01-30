# Commands Reference

Complete reference for the clinvk CLI.

## Synopsis

```bash
clinvk [flags] [prompt]
clinvk [command] [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| [`[prompt]`](prompt.md) | Execute a prompt (root command) |
| [`resume`](resume.md) | Resume a previous session |
| [`sessions`](sessions.md) | Manage sessions |
| [`config`](config.md) | Manage configuration |
| [`parallel`](parallel.md) | Execute tasks in parallel |
| [`compare`](compare.md) | Compare backend responses |
| [`chain`](chain.md) | Execute a prompt chain |
| [`serve`](serve.md) | Start HTTP API server |
| `version` | Show version information |
| `help` | Show help |

## Persistent flags

These flags apply to all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | config / `claude` | Backend to use |
| `--model` | `-m` | string | | Model override |
| `--workdir` | `-w` | string | current dir | Working directory |
| `--output-format` | `-o` | string | config / `json` | `text`, `json`, `stream-json` |
| `--config` | | string | `~/.clinvk/config.yaml` | Config file path |
| `--dry-run` | | bool | `false` | Print command without executing |
| `--ephemeral` | | bool | `false` | Stateless mode (no session) |

Notes:

- Defaults come from config when flags are not explicitly set.
- `--output-format` is normalized to lowercase.

## Root‑command flags only

| Flag | Short | Description |
|------|-------|-------------|
| `--continue` | `-c` | Continue the most recent resumable session |
