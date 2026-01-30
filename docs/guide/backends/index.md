# Backends

clinvk wraps external CLIs. Install the backends you need and configure defaults per backend.

## Supported backends

| Backend | CLI command | Notes |
|---------|-------------|-------|
| Claude Code | `claude` | Best support for sessions and system prompts |
| Codex CLI | `codex` | Uses `codex exec --json` for non‑interactive runs |
| Gemini CLI | `gemini` | Supports `--output-format` and session cleanup |

## Configure per backend

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
  codex:
    model: o3
  gemini:
    model: gemini-2.5-pro
```

See:

- [Claude Code](claude.md)
- [Codex CLI](codex.md)
- [Gemini CLI](gemini.md)
