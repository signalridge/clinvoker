# Exit Codes

Complete reference for clinvk exit codes and their meanings.

## Overview

clinvk uses exit codes to indicate the result of command execution. Understanding these codes is essential for scripting and automation.

## Exit Code Reference

| Code | Name | Description | When It Occurs |
|------|------|-------------|----------------|
| 0 | Success | Command completed successfully | Normal completion |
| 1 | General Error | CLI/validation error or subcommand failure | Invalid input, execution failure |
| 2 | Backend Not Available | The requested backend is not installed | Backend binary not found |
| 3 | Invalid Configuration | Configuration file error or invalid settings | Bad config file |
| 4 | Session Error | Session operation failed | Resume failed, session not found |
| 5 | API Error | HTTP API request failed | Server error, network issue |
| 6 | Timeout | Command execution timed out | Exceeded timeout limit |
| 7 | Cancelled | User cancelled the operation | Ctrl+C pressed |
| 8+ | Backend Exit Code | Propagated from backend CLI | Backend-specific error |

## Detailed Descriptions

### 0 - Success

The command completed successfully without errors.

```bash
clinvk "hello world"
echo $?  # Output: 0
```

### 1 - General Error

A general error occurred during execution. Common causes include:

- Invalid command-line arguments
- Backend execution failed
- File not found
- Permission denied

```bash
clinvk --invalid-flag "prompt"
echo $?  # Output: 1
```

### 2 - Backend Not Available

The requested backend is not installed or not in PATH.

```bash
clinvk -b nonexistent "prompt"
echo $?  # Output: 2
```

### 3 - Invalid Configuration

The configuration file has errors or contains invalid settings.

```bash
clinvk --config /invalid/config.yaml "prompt"
echo $?  # Output: 3
```

### 4 - Session Error

A session-related operation failed.

```bash
clinvk resume nonexistent-session
echo $?  # Output: 4
```

### 5 - API Error

An HTTP API request failed (when using `clinvk serve` or API mode).

```bash
# Server not running
clinvk --api-mode "prompt"
echo $?  # Output: 5
```

### 6 - Timeout

The command execution exceeded the configured timeout.

```bash
clinvk --timeout 5 "very long task"
echo $?  # Output: 6
```

### 7 - Cancelled

The user cancelled the operation (e.g., pressed Ctrl+C).

```bash
clinvk "long running task"
# Press Ctrl+C
echo $?  # Output: 7
```

### Backend Exit Codes (8+)

When running `clinvk [prompt]` or `clinvk resume`, clinvk executes the backend CLI and propagates the backend's exit code when it is non-zero. These codes are backend-specific.

## Command-Specific Exit Codes

### prompt / resume

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2+ | Backend exit code (propagated) |

### parallel

| Code | Description |
|------|-------------|
| 0 | All tasks succeeded |
| 1 | One or more tasks failed |
| 2 | Invalid task file |

### compare

| Code | Description |
|------|-------------|
| 0 | All backends succeeded |
| 1 | One or more backends failed |
| 2 | No backends available |

### chain

| Code | Description |
|------|-------------|
| 0 | All steps succeeded |
| 1 | A step failed |
| 2 | Invalid pipeline file |

### sessions

| Code | Description |
|------|-------------|
| 0 | Operation succeeded |
| 1 | Operation failed (e.g., session not found) |
| 4 | Session error |

### config

| Code | Description |
|------|-------------|
| 0 | Operation succeeded |
| 1 | Invalid key or value |
| 3 | Configuration error |

### serve

| Code | Description |
|------|-------------|
| 0 | Clean shutdown (SIGINT/SIGTERM) |
| 1 | Server startup error |
| 5 | API error during operation |

## Scripting Examples

### Check Success

```bash
if clinvk "implement feature"; then
  echo "Success!"
else
  echo "Failed!"
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
  1)
    echo "General error"
    ;;
  2)
    echo "Backend not available - please install codex"
    ;;
  4)
    echo "Session error"
    ;;
  *)
    echo "Backend error: $code"
    ;;
esac
```

### Retry on Failure

```bash
max_attempts=3
attempt=1

while [ $attempt -le $max_attempts ]; do
  if clinvk "prompt"; then
    echo "Success on attempt $attempt"
    break
  fi

  if [ $attempt -eq $max_attempts ]; then
    echo "Failed after $max_attempts attempts"
    exit 1
  fi

  echo "Attempt $attempt failed, retrying in 5 seconds..."
  sleep 5
  attempt=$((attempt + 1))
done
```

### Exit on Error

```bash
#!/bin/bash
set -e  # Exit on any error

clinvk "step 1"
clinvk "step 2"
clinvk "step 3"

echo "All steps completed successfully"
```

### Ignore Specific Errors

```bash
#!/bin/bash

# Continue even if this fails
clinvk "optional task" || true

# This must succeed
clinvk "critical task"
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run AI task
  run: clinvk "generate tests"
  continue-on-error: true
  id: ai-task

- name: Handle failure
  if: failure() && steps.ai-task.outcome == 'failure'
  run: |
    echo "AI task failed with exit code $?"
    exit 1
```

### GitLab CI

```yaml
ai-task:
  script:
    - clinvk "generate tests" || EXIT_CODE=$?
    - |
      case $EXIT_CODE in
        0) echo "Success" ;;
        2) echo "Backend not installed" ; exit 1 ;;
        *) echo "Error: $EXIT_CODE" ; exit 1 ;;
      esac
```

### Make/Just

```makefile
.PHONY: test lint ai-review

test:
 go test ./...

ai-review:
 clinvk "review the code for issues" || (echo "Review failed" && exit 1)

lint-and-review: lint ai-review
 @echo "All checks passed"
```

## Exit Code Best Practices

1. **Always check exit codes** in scripts to handle failures gracefully
2. **Use `set -e`** in bash scripts to exit immediately on errors
3. **Log the exit code** when debugging issues
4. **Handle specific codes** differently based on your needs
5. **Use `|| true`** when a command failure should not stop the script

## Troubleshooting

### Unexpected Exit Codes

| Symptom | Possible Cause | Solution |
|---------|----------------|----------|
| Always returns 1 | Backend not configured | Check config and API keys |
| Returns 2 | Backend not installed | Install the backend CLI |
| Returns 4 | Session expired or invalid | Check session with `clinvk sessions list` |
| Returns 6 | Timeout too short | Increase `command_timeout_secs` |

### Debug Exit Codes

```bash
# Run with verbose output
clinvk -v "prompt"
echo "Exit code: $?"

# Check backend directly
claude "test"
echo "Backend exit code: $?"
```

## See Also

- [Commands Reference](cli/index.md) - Command documentation
- [Troubleshooting](../concepts/troubleshooting.md) - Common issues and solutions
