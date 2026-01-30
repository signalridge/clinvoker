# CI/CD

Use clinvk in CI pipelines for reviews, summaries, or test generation.

## CLI in CI

```yaml
- name: AI review (parallel)
  run: |
    cat > tasks.json <<'JSON'
    {"tasks":[
      {"backend":"claude","prompt":"Review architecture"},
      {"backend":"codex","prompt":"Review performance"}
    ]}
    JSON
    clinvk parallel -f tasks.json
```

Exit code is non‑zero if any task fails.

## HTTP server in CI

Start the server once and call the API from jobs:

```bash
clinvk serve --host 0.0.0.0 --port 8080
```

```bash
curl -sS http://localhost:8080/api/v1/compare \
  -H 'Content-Type: application/json' \
  -d '{"backends":["claude","gemini"],"prompt":"Review diff"}'
```

## Best practices

- Use `--ephemeral` or API `ephemeral: true` for stateless CI runs.
- Store API keys in secrets (if you enable server auth).
