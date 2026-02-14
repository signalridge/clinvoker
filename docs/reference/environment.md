# Environment Variables

Authoritative reference for environment variables currently supported by `clinvk`.

## Scope

This page lists variables that are actually read by the current codebase.
If a variable is not listed here, do not rely on it.

## Supported Variables

### Core Runtime

| Variable | Description | Example |
|---|---|---|
| `CLINVK_BACKEND` | Default backend | `claude`, `codex`, `gemini` |
| `CLINVK_CLAUDE_MODEL` | Default model for Claude backend | `claude-opus-4-5-20251101` |
| `CLINVK_CODEX_MODEL` | Default model for Codex backend | `o3`, `o3-mini` |
| `CLINVK_GEMINI_MODEL` | Default model for Gemini backend | `gemini-2.5-pro` |

### MCP Server

| Variable | Description | Example |
|---|---|---|
| `CLINVK_MCP_TRANSPORT` | MCP transport type | `stdio`, `http` |
| `CLINVK_MCP_HOST` | Host for MCP HTTP transport | `127.0.0.1`, `0.0.0.0` |
| `CLINVK_MCP_PORT` | Port for MCP HTTP transport | `8081` |
| `CLINVK_MCP_HTTP_PATH` | HTTP path for MCP endpoint | `/mcp` |
| `CLINVK_MCP_EXPOSE_HEALTH` | Expose `/health` in MCP mode | `true`, `false` |

### HTTP API Authentication

| Variable | Description | Example |
|---|---|---|
| `CLINVK_API_KEYS` | Comma-separated API keys for `serve`/`mcp` auth | `key1,key2,key3` |
| `CLINVK_API_KEYS_GOPASS_PATH` | gopass path to load API keys | `myproject/clinvk/api-keys` |

### Backend Provider Keys

| Variable | Backend | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | Claude | Provider API key |
| `OPENAI_API_KEY` | Codex | Provider API key |
| `GOOGLE_API_KEY` | Gemini | Provider API key |

## Commonly Mistaken Variables

The following are not supported as first-class runtime config variables in current `clinvk` releases:

- `CLINVK_TIMEOUT`
- `CLINVK_DEBUG`
- `CLINVK_SERVER_PORT`
- `CLINVK_HOME`
- `CLINVK_CONFIG`

Use these instead:

- command timeout: `unified_flags.command_timeout_secs` in config file
- debug/inspection: `--dry-run`, `--output-format json`, and regular stderr logs
- config file path: `--config /path/to/config.yaml`

## Examples

### Backend and Model

```bash
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3-mini
clinvk "review this patch"
```

### MCP via Environment

```bash
export CLINVK_MCP_TRANSPORT=http
export CLINVK_MCP_HOST=0.0.0.0
export CLINVK_MCP_PORT=8081
export CLINVK_MCP_HTTP_PATH=/mcp
export CLINVK_MCP_EXPOSE_HEALTH=true

clinvk mcp
```

### API Key Auth for HTTP/MCP

```bash
export CLINVK_API_KEYS="prod-key-1,prod-key-2"
clinvk serve --port 8080
```

## Precedence

For supported keys, precedence is:

1. CLI flags
2. Environment variables
3. Config file
4. Defaults

## Troubleshooting

```bash
# Confirm exported variables
env | grep '^CLINVK_'

# Verify provider key exists
echo "${OPENAI_API_KEY:+set}"

# Inspect effective execution command
clinvk --dry-run "check behavior"
```

## See Also

- [Configuration Reference](configuration.md)
- [CLI Config Command](cli/config.md)
- [MCP CLI Reference](cli/mcp.md)
