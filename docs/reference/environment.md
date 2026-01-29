# Environment Variables

Reference for all environment variables supported by clinvk.

## Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CLINVK_BACKEND` | Default backend | `claude` |
| `CLINVK_CLAUDE_MODEL` | Claude model | (backend default) |
| `CLINVK_CODEX_MODEL` | Codex model | (backend default) |
| `CLINVK_GEMINI_MODEL` | Gemini model | (backend default) |

!!! note
    Only the variables above are explicitly supported. Other configuration keys are not currently mapped to environment variables.

## Usage Examples

### Set Default Backend

```bash
export CLINVK_BACKEND=codex
clinvk "implement feature"  # Uses codex
```

### Set Model per Backend

```bash
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3-mini

clinvk -b claude "complex task"  # Uses claude-sonnet-4-20250514
clinvk -b codex "quick task"     # Uses o3-mini
```

### Temporary Override

```bash
CLINVK_BACKEND=gemini clinvk "explain this"
```

## Priority

Environment variables have medium priority:

1. **CLI Flags** (highest)
2. **Environment Variables**
3. **Config File**
4. **Defaults** (lowest)

Example:

```bash
export CLINVK_BACKEND=codex
clinvk -b claude "prompt"  # Uses claude (CLI flag wins)
```

## Shell Configuration

### Bash

Add to `~/.bashrc` or `~/.bash_profile`:

```bash
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
```

### Zsh

Add to `~/.zshrc`:

```zsh
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
set -gx CLINVK_BACKEND claude
set -gx CLINVK_CLAUDE_MODEL claude-opus-4-5-20251101
```

## Per-Directory Configuration

Use direnv for project-specific settings:

```bash
# .envrc
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3
```

## CI/CD Usage

### GitHub Actions

```yaml
jobs:
  build:
    env:
      CLINVK_BACKEND: codex
      CLINVK_CODEX_MODEL: o3
    steps:
      - run: clinvk "generate tests"
```

### Docker

```dockerfile
ENV CLINVK_BACKEND=claude
ENV CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
```

Or at runtime:

```bash
docker run -e CLINVK_BACKEND=codex clinvk "prompt"
```

## See Also

- [Configuration Reference](configuration.md)
- [config command](commands/config.md)
