# Exit Codes

Reference for clinvk exit codes and their meanings.

## Exit Code Summary

| Code | Name | Description |
|------|------|-------------|
| 0 | Success | Command completed successfully |
| 1 | General Error | General error or failure |
| 2 | Usage Error | Command line usage error |
| 126 | Backend Unavailable | Backend CLI not available |
| 127 | Backend Not Found | Backend CLI not found in PATH |

## Detailed Descriptions

### 0 - Success

The command completed successfully.

```bash
clinvk "hello world"
echo $?  # 0
```

### 1 - General Error

A general error occurred during execution:

- Backend execution failed
- Network error
- Timeout
- Invalid input

```bash
clinvk "prompt with error"
echo $?  # 1
```

### 2 - Usage Error

Invalid command line arguments or flags:

```bash
clinvk --invalid-flag "prompt"
echo $?  # 2
```

### 126 - Backend Unavailable

The backend is configured but not currently available (e.g., API down):

```bash
clinvk -b codex "prompt"  # Backend API unavailable
echo $?  # 126
```

### 127 - Backend Not Found

The backend's CLI tool is not installed or not in PATH:

```bash
clinvk -b gemini "prompt"  # 'gemini' not in PATH
echo $?  # 127
```

## Command-Specific Exit Codes

### parallel

| Code | Description |
|------|-------------|
| 0 | All tasks succeeded |
| 1 | One or more tasks failed |

### compare

| Code | Description |
|------|-------------|
| 0 | At least one backend succeeded |
| 1 | All backends failed |

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
  126)
    echo "Backend unavailable, trying another..."
    clinvk -b claude "prompt"
    ;;
  127)
    echo "Codex not installed"
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
