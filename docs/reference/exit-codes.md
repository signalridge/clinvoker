# Exit Codes

Reference for clinvk exit codes and their meanings.

## Exit Code Summary

| Code | Name | Description |
|------|------|-------------|
| 0 | Success | Command completed successfully |
| 1 | Error | CLI/validation error or subcommand failure |
| (backend) | Backend exit code | For `clinvk [prompt]` and `clinvk resume`, the backend CLI exit code is propagated |

## Detailed Descriptions

### 0 - Success

The command completed successfully.

```bash
clinvk "hello world"
echo $?  # 0
```

### 1 - Error

A general error occurred during execution, for example:

- Backend execution failed
- Invalid input

```bash
clinvk "prompt with error"
echo $?  # 1
```

### Backend Exit Codes (prompt/resume)

When running `clinvk [prompt]` or `clinvk resume`, clinvk executes the backend CLI and propagates the backend process exit code when it is non-zero.

## Command-Specific Exit Codes

### parallel

| Code | Description |
|------|-------------|
| 0 | All tasks succeeded |
| 1 | One or more tasks failed |

### compare

| Code | Description |
|------|-------------|
| 0 | All backends succeeded |
| 1 | One or more backends failed |

### chain

| Code | Description |
|------|-------------|
| 0 | All steps succeeded |
| 1 | A step failed |

### serve

| Code | Description |
|------|-------------|
| 0 | Clean shutdown (SIGINT/SIGTERM) |
| 1 | Server error |

## Scripting Examples

### Check Success

```bash
if clinvk "implement feature"; then
  echo "Success"
else
  echo "Failed"
fi
```

### Handle Specific Codes

```bash
clinvk -b codex "prompt"
code=$?

case $code in
  0)
    echo "Success"
    ;;
  *)
    echo "Error: $code"
    ;;
esac
```

### Retry on Failure

```bash
max_attempts=3
attempt=1

while [ $attempt -le $max_attempts ]; do
  if clinvk "prompt"; then
    break
  fi
  echo "Attempt $attempt failed, retrying..."
  attempt=$((attempt + 1))
  sleep 2
done
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run AI task
  run: clinvk "generate tests"
  continue-on-error: true
  id: ai-task

- name: Handle failure
  if: steps.ai-task.outcome == 'failure'
  run: echo "AI task failed"
```

### Make/Just

```makefile
test:
 clinvk "generate tests" || (echo "Failed" && exit 1)
```

## See Also

- [Commands Reference](commands/index.md)
- [Troubleshooting](../development/troubleshooting.md)
