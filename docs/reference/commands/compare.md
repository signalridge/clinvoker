# clinvk compare

Compare the same prompt across multiple backends.

## Synopsis

```bash
clinvk compare <prompt> --backends claude,gemini
clinvk compare <prompt> --all-backends
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--backends` | | Comma‑separated backend list |
| `--all-backends` | false | Use all backends |
| `--json` | false | Output JSON |
| `--sequential` | false | Run backends sequentially |

## Notes

- Comparison is **always ephemeral**.
- Backends not installed are skipped with a warning.
- You must provide `--backends` or `--all-backends`.

## Exit codes

- `0` all backends succeeded
- `1` one or more backends failed
