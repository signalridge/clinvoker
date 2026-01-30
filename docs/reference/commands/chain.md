# clinvk chain

Execute a chain of prompts in sequence.

## Synopsis

```bash
clinvk chain --file chain.json [--json]
cat chain.json | clinvk chain
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Chain definition file (JSON) |
| `--input` | | | Deprecated alias for `--file` |
| `--json` | | false | Emit JSON summary |

## Chain format

```json
{
  "steps": [
    {"backend": "claude", "prompt": "Analyze"},
    {"backend": "codex", "prompt": "Fix: {{previous}}"}
  ],
  "stop_on_failure": true,
  "pass_working_dir": false
}
```

## Notes

- CLI chain is **always ephemeral** (no sessions).
- `{{previous}}` is the only supported placeholder.
- Current CLI behavior always stops on the first failure.

## Exit codes

- `0` all steps succeeded
- `1` a step failed
