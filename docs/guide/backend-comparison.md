# Backend Comparison

Compare the same prompt across multiple backends.

## Basic usage

```bash
clinvk compare --backends claude,gemini "Explain this algorithm"
```

Use all backends:

```bash
clinvk compare --all-backends "Review this PR"
```

If a backend CLI is not installed, clinvk warns and skips it.

## Sequential vs parallel

```bash
clinvk compare --all-backends --sequential "Check for security risks"
```

By default, comparisons run in parallel.

## JSON output

```bash
clinvk compare --all-backends --json "Summarize this patch"
```

## Notes

- Compare is always **ephemeral** (no sessions persisted).
- Exit code is non‑zero if any backend fails.
