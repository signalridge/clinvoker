# Exit Codes

Reference for `clinvk` exit codes based on current implementation behavior.

## Overview

`clinvk` uses process exit codes for scripting and automation.

## Guaranteed Exit Codes

| Code | Name | Description |
|------|------|-------------|
| 0 | Success | Command completed successfully |
| 1 | General Error | Validation/config/runtime error, or aggregated subcommand failure |
| 6 | Timeout | Command exceeded `unified_flags.command_timeout_secs` and timeout bubbled to top level |
| 2+ | Backend Exit Code | Propagated backend CLI exit code (for `prompt` / `resume` paths) |

## Command Behavior

### prompt / resume

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error before backend exit propagation |
| 6 | Timeout (`command_timeout_secs`) |
| 2+ | Backend CLI exit code is propagated |

### compare / chain / parallel

| Code | Meaning |
|------|---------|
| 0 | All units succeeded |
| 1 | One or more units failed (including timeout inside a unit) |

### sessions / config / serve / other utility commands

| Code | Meaning |
|------|---------|
| 0 | Command succeeded |
| 1 | Command failed |

## Timeout Semantics

When `unified_flags.command_timeout_secs` is set to a positive value:

- Backend process timeout is detected and treated as a timeout error.
- Top-level timeout errors are normalized to exit code `6`.
- In aggregate commands (`compare` / `chain` / `parallel`), per-unit timeout is treated as unit failure and contributes to aggregate exit code `1`.

Example:

```bash
# config.yaml
unified_flags:
  command_timeout_secs: 5

clinvk "long running task"
echo $?  # 6 if timed out at top-level prompt/resume path
```

## Scripting Examples

### Branch by exit code

```bash
clinvk "task"
code=$?

case $code in
  0)
    echo "success"
    ;;
  6)
    echo "timeout"
    ;;
  *)
    echo "failed with code $code"
    ;;
esac
```

### Fail fast in shell

```bash
set -e

clinvk "step 1"
clinvk "step 2"
```

## Notes

- Do not assume historic code mappings that are not listed above.
- If you depend on specific behavior in CI, pin your workflow to the documented contract in this file.

## See Also

- [Commands Reference](cli/index.md)
- [Configuration Reference](configuration.md)
