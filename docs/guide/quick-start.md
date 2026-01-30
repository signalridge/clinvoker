# Quick Start

Get productive in a few minutes.

## 1) Run your first prompt

```bash
clinvk "explain what this repo does"
```

If you didn’t set a default backend, clinvk falls back to `claude`.

## 2) Pick a backend explicitly

```bash
clinvk -b codex "optimize this function"
clinvk -b gemini "summarize this document"
```

## 3) Try sessions

```bash
clinvk "draft a migration plan"
clinvk -c "now add a rollback strategy"
```

`-c/--continue` continues the most recent resumable session.

## 4) Run parallel tasks

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "Review architecture"},
    {"backend": "codex", "prompt": "Review performance"},
    {"backend": "gemini", "prompt": "Review security"}
  ]
}
```

```bash
clinvk parallel -f tasks.json
```

## 5) Start the API server

```bash
clinvk serve --port 8080
```

Then call it with curl:

```bash
curl -sS http://localhost:8080/api/v1/prompt \
  -H 'Content-Type: application/json' \
  -d '{"backend":"claude","prompt":"Hello"}'
```

## What next

- [Configuration](configuration.md) — set defaults
- [Basic Usage](basic-usage.md) — core CLI flows
- [HTTP Server](http-server.md) — OpenAI/Anthropic compatibility
