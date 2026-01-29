# clinvk compare

Compare responses from multiple backends.

## Synopsis

```
clinvk compare [prompt] [flags]
```

## Description

Send the same prompt to multiple AI backends and compare their responses side-by-side.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--backends` | string | | Comma-separated backend list |
| `--all-backends` | bool | `false` | Compare all enabled backends |
| `--sequential` | bool | `false` | Run one at a time |
| `--json` | bool | `false` | JSON output |

## Examples

### Compare Specific Backends

```bash
clinvk compare --backends claude,codex "explain this code"
```

### Compare All Backends

```bash
clinvk compare --all-backends "what does this function do"
```

### Sequential Execution

```bash
clinvk compare --all-backends --sequential "review this PR"
```

### JSON Output

```bash
clinvk compare --all-backends --json "analyze performance"
```

## Output

### Text Output (Default)

```
Comparing 3 backends: claude, codex, gemini
Prompt: explain this algorithm
============================================================
[claude] This algorithm implements a binary search...
[codex] The algorithm performs a binary search...
[gemini] This is a classic binary search implementation...

============================================================
COMPARISON SUMMARY
============================================================
BACKEND      STATUS     DURATION     SESSION    MODEL
------------------------------------------------------------
claude       OK         2.50s        abc123     claude-opus-4-5-20251101
codex        OK         3.20s        def456     o3
gemini       OK         2.80s        ghi789     gemini-2.5-pro
------------------------------------------------------------
Total time: 3.20s
```

### JSON Output

```json
{
  "prompt": "explain this algorithm",
  "backends": ["claude", "codex", "gemini"],
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "output": "This algorithm implements a binary search...",
      "duration_seconds": 2.5,
      "exit_code": 0,
      "session_id": "abc123"
    }
  ],
  "total_duration_seconds": 3.2
}
```

## Execution Modes

### Parallel (Default)

All backends run simultaneously:

```bash
clinvk compare --all-backends "prompt"
```

Benefits:

- Faster total time
- Results as fast as slowest backend

### Sequential

Run backends one at a time:

```bash
clinvk compare --all-backends --sequential "prompt"
```

Benefits:

- Avoids rate limits
- Lower resource usage
- Watch results come in

## Error Handling

If a backend fails, comparison continues with remaining backends:

```
Comparing 3 backends: claude, codex, gemini
Prompt: explain this code
============================================================
[claude] Response content...
[codex] Error: Backend unavailable
[gemini] Response content...

============================================================
COMPARISON SUMMARY
============================================================
BACKEND      STATUS     DURATION     SESSION    MODEL
------------------------------------------------------------
claude       OK         2.10s        abc123     (default)
codex        FAILED     0.50s        -          (default)
             Error: Backend unavailable
gemini       OK         1.80s        def456     (default)
------------------------------------------------------------
Total time: 2.10s
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | At least one backend succeeded |
| 1 | All backends failed |

## See Also

- [parallel](parallel.md) - Different prompts, concurrent
- [chain](chain.md) - Sequential pipeline
