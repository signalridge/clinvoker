# Session Management

Sessions let you resume conversations across runs and track token usage.

## How sessions work

- Sessions are stored in `~/.clinvk/sessions`.
- A session is resumable **only if the backend returns a session ID**.
- `--ephemeral` disables session persistence.

## Quick continue

```bash
clinvk "design a new API"
clinvk -c "add pagination"
```

`-c/--continue` resumes the most recent resumable session (filtered by `--backend` if provided).

## Resume command

```bash
# Resume the most recent session
clinvk resume --last

# Interactive picker
clinvk resume --interactive

# Filter by backend
clinvk resume --backend codex

# Filter by current working directory
clinvk resume --here
```

You can also pass a prompt when resuming:

```bash
clinvk resume --last "continue from here"
```

## List and inspect sessions

```bash
clinvk sessions list
clinvk sessions list --backend claude
clinvk sessions list --status completed

clinvk sessions show <session-id>
```

## Delete or clean

```bash
clinvk sessions delete <session-id>

# Clean by retention (uses config if omitted)
clinvk sessions clean
clinvk sessions clean --older-than 30d
```

## Config options

```yaml
session:
  auto_resume: true
  retention_days: 30
  store_token_usage: true
  default_tags: []
```

Notes:

- `auto_resume` will attempt to resume the most recent resumable session whenever you run the root command (unless `--ephemeral` is set). The prompt, if provided, is used as the continuation message.
- `store_token_usage` only records tokens when a backend reports usage.

## When sessions are not created

- `clinvk --ephemeral ...`
- `parallel`, `chain`, and `compare` (always ephemeral)
- OpenAI/Anthropic compatible endpoints (stateless server mode)
