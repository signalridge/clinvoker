# clinvk compare

Compare responses from multiple backends.

## Synopsis

```bash
clinvk compare <prompt> [flags]
```

## Description

Send the same prompt to multiple backends and compare their responses. CLI compare runs are always ephemeral (no sessions are persisted).

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--backends` | string | | Comma-separated backend list |
| `--all-backends` | bool | `false` | Compare all registered backends (skips unavailable CLIs) |
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

### Text Output

```text
Comparing 3 backends: claude, codex, gemini
Prompt: explain this algorithm
================================================================================
[claude] This algorithm implements a binary search...
[codex] The algorithm performs a binary search...
[gemini] This is a classic binary search implementation...

================================================================================
COMPARISON SUMMARY
================================================================================
BACKEND      STATUS     DURATION     MODEL
--------------------------------------------------------------------------------
claude       OK         2.50s        claude-opus-4-5-20251101
codex        OK         3.20s        o3
gemini       OK         2.80s        gemini-2.5-pro
--------------------------------------------------------------------------------
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
      "exit_code": 0
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

### Sequential

Run backends one at a time:

```bash
clinvk compare --all-backends --sequential "prompt"
```

## Error Handling

Unavailable backends are skipped with a warning. If any selected backend fails during execution, the command exits with a non-zero status.

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | All selected backends succeeded |
| 1 | Any backend failed or none were available |

## See Also

- [parallel](parallel.md) - Different prompts, concurrent
- [chain](chain.md) - Sequential pipeline
