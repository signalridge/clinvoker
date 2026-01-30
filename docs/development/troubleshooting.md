# Troubleshooting

Common issues and solutions for clinvk.

## Backend Issues

### Backend Not Found

**Error:** `Backend 'codex' not found in PATH`

**Solution:**

1. Verify the backend CLI is installed:

   ```bash
   which claude codex gemini
   ```

2. Add the binary location to your PATH:

   ```bash
   export PATH="$PATH:/path/to/backend"
   ```

3. Check clinvk detection:

   ```bash
   clinvk config show | grep available
   ```

### Backend Unavailable

**Error:** `Backend unavailable` or exit code 126

**Solution:**

- Check if the backend's API is accessible
- Verify API credentials are configured for the backend CLI
- Try running the backend CLI directly to diagnose

### Model Not Found

**Error:** `Model 'invalid-model' not found`

**Solution:**

1. List available models for the backend:

   ```bash
   # For Claude
   claude models list

   # Check configured model
   clinvk config show | grep model
   ```

2. Update your configuration:

   ```bash
   clinvk config set backends.claude.model claude-opus-4-5-20251101
   ```

## Configuration Issues

### Config File Not Loading

**Symptoms:** Settings not applied

**Solution:**

1. Check config file location:

   ```bash
   ls -la ~/.clinvk/config.yaml
   ```

2. Validate YAML syntax:

   ```bash
   cat ~/.clinvk/config.yaml | python -c "import yaml,sys; yaml.safe_load(sys.stdin)"
   ```

3. Check file permissions:

   ```bash
   chmod 600 ~/.clinvk/config.yaml
   ```

4. View effective configuration:

   ```bash
   clinvk config show
   ```

### Environment Variables Not Applied

**Solution:**

1. Verify the variable is set:

   ```bash
   echo $CLINVK_BACKEND
   ```

2. Check shell configuration is loaded:

   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```

3. Remember CLI flags override environment variables

## Session Issues

### Cannot Resume Session

**Error:** `Session 'abc123' not found`

**Solution:**

1. List available sessions:

   ```bash
   clinvk sessions list
   ```

2. Check session directory:

   ```bash
   ls ~/.clinvk/sessions/
   ```

3. Sessions may have been cleaned. Create a new session instead.

### Session Storage Full

**Solution:**

Clean old sessions:

```bash
# Clean sessions older than 7 days
clinvk sessions clean --older-than 7d

# Or delete all sessions
rm -rf ~/.clinvk/sessions/*
```

## Server Issues

### Port Already in Use

**Error:** `Address already in use`

**Solution:**

1. Find the process using the port:

   ```bash
   lsof -i :8080
   ```

2. Use a different port:

   ```bash
   clinvk serve --port 3000
   ```

3. Or kill the existing process:

   ```bash
   kill -9 <PID>
   ```

### Cannot Connect to Server

**Solution:**

1. Verify server is running:

   ```bash
   curl http://localhost:8080/health
   ```

2. Check bind address:

   ```bash
   # If connecting from another machine, use 0.0.0.0
   clinvk serve --host 0.0.0.0
   ```

3. Check firewall settings

## Execution Issues

### Command Timeout

**Error:** Request or command times out

**Solution:**

1. Increase timeout in config:

   ```yaml
   server:
     request_timeout_secs: 600
   ```

2. For complex tasks, break into smaller prompts

### Output Truncated

**Solution:**

1. Use JSON output for full response:

   ```bash
   clinvk -o json "long prompt"
   ```

2. Check backend-specific output limits

### Rate Limiting

**Error:** Rate limit exceeded

**Solution:**

1. Wait before retrying
2. Use `--sequential` for compare commands
3. Reduce parallel workers:

   ```bash
   clinvk parallel --max-parallel 1 --file tasks.json
   ```

## Platform-Specific Issues

### macOS: Gatekeeper Blocking

**Solution:**

```bash
xattr -d com.apple.quarantine /path/to/clinvk
```

### Windows: PATH Issues

**Solution:**

Add to PATH via System Properties → Environment Variables, or:

```powershell
$env:Path += ";C:\path\to\clinvk"
```

### Linux: Permission Denied

**Solution:**

```bash
chmod +x /path/to/clinvk
```

## Debugging

### Enable Verbose Output

```bash
clinvk --verbose "prompt"
```

### Dry Run Mode

See what command would be executed:

```bash
clinvk --dry-run "prompt"
```

### Check Version

Verify you're running the expected version:

```bash
clinvk version
```

## Getting Help

If issues persist:

1. Check [GitHub Issues](https://github.com/signalridge/clinvoker/issues)
2. Search for similar problems
3. Open a new issue with:
   - clinvk version
   - OS and version
   - Error message
   - Steps to reproduce
