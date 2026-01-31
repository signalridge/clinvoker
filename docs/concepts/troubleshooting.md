---
title: Troubleshooting
description: Common issues, diagnostic methods, and solutions for clinvoker.
---

# Troubleshooting

This guide covers common issues you may encounter when using clinvoker, along with diagnostic methods and solutions. Issues are organized by category for easy reference.

## Diagnostic Methods

Before diving into specific issues, here are general diagnostic approaches:

### Enable Debug Mode

```bash
# Verbose output shows detailed execution information
clinvk --verbose "your prompt"

# Debug mode includes backend command output
CLINVK_DEBUG=1 clinvk "your prompt"
```text

### Check System Status

```bash
# View version and build info
clinvk version

# Check available backends
clinvk config show

# Verify configuration file
clinvk config validate
```text

### Dry Run Mode

See what command would be executed without actually running it:

```bash
clinvk --dry-run "your prompt"
```text

## Backend Not Available

### Symptoms

- Error: `Backend 'claude' not found in PATH`
- Error: `No backends available`
- Exit code: 126

### Causes

1. Backend CLI not installed
2. Backend CLI not in PATH
3. Backend CLI has incorrect permissions
4. Backend CLI is a different version than expected

### Solutions

**1. Verify Installation**

```bash
# Check if backend CLIs are installed
which claude codex gemini

# Check versions
claude --version
codex --version
gemini --version
```text

**2. Add to PATH**

```bash
# Add to shell profile (~/.bashrc, ~/.zshrc, etc.)
export PATH="$PATH:/usr/local/bin"

# Reload profile
source ~/.bashrc  # or ~/.zshrc
```text

**3. Check Permissions**

```bash
# Verify executable permissions
ls -la $(which claude)

# Fix if needed
chmod +x /path/to/claude
```bash

**4. Verify clinvk Detection**

```bash
# Check which backends clinvk can find
clinvk config show | grep -A2 "available"
```text

## Session Issues

### Session Corruption

**Symptoms:**
- Error: `Failed to load session: invalid JSON`
- Session file appears empty or truncated
- Cannot resume a previously working session

**Causes:**
1. Process killed during session write
2. Disk full during write
3. Concurrent modification without proper locking

**Solutions:**

```bash
# Check session file integrity
ls -la ~/.clinvk/sessions/

# View session content (JSON should be valid)
cat ~/.clinvk/sessions/<session-id>.json | python -m json.tool

# If corrupted, remove the session
clinvk sessions delete <session-id>

# Or clean all sessions
rm -rf ~/.clinvk/sessions/*
```bash

### Session Locking Issues

**Symptoms:**
- Error: `Failed to acquire store lock: timeout`
- Operations hang indefinitely
- "Resource temporarily unavailable" errors

**Causes:**
1. Another clinvk process holding the lock
2. Previous process crashed without releasing lock
3. Stale lock file

**Solutions:**

```bash
# Check for running clinvk processes
ps aux | grep clinvk

# Kill stale processes if necessary
kill -9 <pid>

# Remove stale lock file (use with caution)
rm -f ~/.clinvk/.store.lock

# Verify lock file permissions
ls -la ~/.clinvk/
```text

### Session Not Found

**Symptoms:**
- Error: `Session 'abc123' not found`
- Cannot resume previous session

**Solutions:**

```bash
# List all available sessions
clinvk sessions list

# Check session directory
ls -la ~/.clinvk/sessions/

# Search for session by partial ID
find ~/.clinvk/sessions/ -name "*abc*"
```text

## API Server Problems

### Server Startup Issues

**Symptoms:**
- Error: `Address already in use`
- Error: `Permission denied` when binding to port
- Server exits immediately

**Solutions:**

**Port Already in Use:**

```bash
# Find process using the port
lsof -i :8080
# or
netstat -tlnp | grep 8080

# Use a different port
clinvk serve --port 3000

# Or kill the existing process
kill -9 <pid>
```text

**Permission Denied (Low Ports):**

```bash
# Use port > 1024 (no root required)
clinvk serve --port 8080

# Or run with sudo (not recommended)
sudo clinvk serve --port 80
```text

### Authentication Issues

**Symptoms:**
- Error: `Unauthorized` (401)
- Error: `Invalid API key`

**Solutions:**

```bash
# Check if API keys are configured
clinvk config show | grep api_keys

# Verify API key in request
curl -H "Authorization: Bearer YOUR_KEY" http://localhost:8080/api/v1/health

# Check environment variable
echo $CLINVK_API_KEY

# Test without authentication (if no keys configured)
curl http://localhost:8080/health
```text

### Rate Limiting Issues

**Symptoms:**
- Error: `Rate limit exceeded` (429)
- Requests being rejected after certain volume

**Solutions:**

```bash
# Check rate limit configuration
clinvk config show | grep rate_limit

# Adjust rate limits in config
clinvk config set server.rate_limit_rps 100

# For parallel execution, reduce concurrency
clinvk parallel --max-parallel 2 --file tasks.json
```text

## Performance Issues

### Slow Response Times

**Symptoms:**
- Commands take longer than expected
- Timeouts occurring

**Causes:**
1. Backend rate limiting
2. Large context/prompt size
3. Network latency to backend APIs
4. Insufficient system resources

**Solutions:**

```bash
# Check backend status
curl http://localhost:8080/api/v1/backends

# Monitor system resources
top
iostat -x 1

# Increase timeout
clinvk config set timeout 300

# Use faster model
clinvk -m fast "prompt"
```bash

### High Memory Usage

**Symptoms:**
- clinvk process using excessive memory
- System becoming unresponsive

**Solutions:**

```bash
# Check memory usage
ps aux | grep clinvk

# Limit parallel execution
clinvk parallel --max-parallel 2 --file tasks.json

# Clean old sessions
clinvk sessions clean --older-than 7d

# Use ephemeral mode (no session storage)
clinvk --ephemeral "prompt"
```text

### Disk Space Issues

**Symptoms:**
- Error: `No space left on device`
- Session operations failing

**Solutions:**

```bash
# Check disk space
df -h

# Check session storage size
du -sh ~/.clinvk/sessions/

# Clean old sessions
clinvk sessions clean --older-than 30d

# Or manually clean
rm -rf ~/.clinvk/sessions/*.json
```text

## Configuration Issues

### Config File Not Loading

**Symptoms:**
- Settings not applied
- Default values being used

**Solutions:**

```bash
# Check config file location
ls -la ~/.clinvk/config.yaml

# Validate YAML syntax
cat ~/.clinvk/config.yaml | python -c "import yaml,sys; yaml.safe_load(sys.stdin)"

# Check file permissions
chmod 600 ~/.clinvk/config.yaml

# View effective configuration
clinvk config show

# Check for environment variable overrides
echo $CLINVK_BACKEND
echo $CLINVK_TIMEOUT
```text

### Environment Variables Not Applied

**Symptoms:**
- Environment settings ignored
- CLI flags work but env vars don't

**Solutions:**

```bash
# Verify variable is exported
export CLINVK_BACKEND=codex

# Check current shell variables
env | grep CLINVK

# Remember priority: CLI flags > Environment > Config file
```text

## Platform-Specific Issues

### macOS

**Gatekeeper Blocking:**

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /path/to/clinvk

# Or allow in System Preferences > Security & Privacy
```text

**Notarization Issues:**

```bash
# If downloaded binary won't run
sudo spctl --master-disable  # Temporarily disable (use with caution)
# Then re-enable after first run
sudo spctl --master-enable
```text

### Windows

**PATH Issues:**

```powershell
# Add to PATH via PowerShell
$env:Path += ";C:\path\to\clinvk"

# Or permanently via System Properties
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\path\to\clinvk", "User")
```bash

**Antivirus False Positives:**

Some antivirus software may flag clinvk. Add an exclusion if necessary.

### Linux

**Permission Denied:**

```bash
chmod +x /path/to/clinvk
```text

**SELinux Issues:**

```bash
# Check SELinux status
getenforce

# If enforcing, check for denials
ausearch -m avc -ts recent

# Create policy module if needed
audit2allow -a -M clinvk
semodule -i clinvk.pp
```text

## Debug Mode Usage

### Enable Comprehensive Logging

```bash
# Set debug environment variable
export CLINVK_DEBUG=1

# Run with verbose output
clinvk --verbose "prompt"

# For server mode
CLINVK_DEBUG=1 clinvk serve
```bash

### Log Locations

**CLI Mode:**
- Logs go to stderr by default
- Redirect to file: `clinvk "prompt" 2> debug.log`

**Server Mode:**
- Logs to stdout/stderr
- Use systemd/journald for persistent logs: `journalctl -u clinvk`
- Or redirect: `clinvk serve > server.log 2>&1`

### Common Log Patterns

**Backend Command:**
```text
[DEBUG] Executing: claude --print --model sonnet "prompt"
[DEBUG] Working directory: /home/user/project
[DEBUG] Exit code: 0
```text

**Session Operations:**
```text
[DEBUG] Loading session: abc123
[DEBUG] Acquiring file lock
[DEBUG] Session saved successfully
```text

**API Requests:**
```text
[DEBUG] POST /api/v1/prompt
[DEBUG] Request body: {...}
[DEBUG] Response: 200 OK
```text

## Getting Help

If issues persist after trying the above solutions:

1. **Check Documentation**
   - [FAQ](faq.md)
   - [Guides](../guides/index.md)
   - [Reference](../reference/index.md)

2. **Search Issues**
   - [GitHub Issues](https://github.com/signalridge/clinvoker/issues)
   - Use search with error message

3. **Open a New Issue**
   Include:
   - clinvk version (`clinvk version`)
   - Operating system and version
   - Backend versions (`claude --version`, etc.)
   - Complete error message
   - Steps to reproduce
   - Debug logs (if possible)

4. **Community Support**
   - Start a GitHub Discussion
   - Check existing discussions for similar problems

## Related Documentation

- [FAQ](faq.md) - Frequently asked questions
- [Design Decisions](design-decisions.md) - Architecture explanations
- [Contributing](contributing.md) - Development setup
