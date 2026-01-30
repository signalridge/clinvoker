# Environment Variables

Environment variables supported by clinvk.

## Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CLINVK_BACKEND` | Default backend | `claude` |
| `CLINVK_CLAUDE_MODEL` | Claude model | (backend default) |
| `CLINVK_CODEX_MODEL` | Codex model | (backend default) |
| `CLINVK_GEMINI_MODEL` | Gemini model | (backend default) |
| `CLINVK_API_KEYS` | API keys for server auth (comma‑separated) | (unset) |
| `CLINVK_API_KEYS_GOPASS_PATH` | gopass path to load API keys | (unset) |

## Priority

1. CLI flags
2. Environment variables
3. Config file
4. Defaults

## Examples

```bash
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3

clinvk "optimize this"
```

### Server auth

```bash
export CLINVK_API_KEYS="key1,key2"
clinvk serve
```

If `CLINVK_API_KEYS` is set, all API requests must include a key.
