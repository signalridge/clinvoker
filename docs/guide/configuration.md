# Configuration Guide

Learn how to configure clinvk for your workflow. This guide covers common scenarios and best practices.

## Quick Setup

### 1. View Current Configuration

```bash
clinvk config show
```

This shows all settings including which backends are available on your system.

### 2. Set Default Backend

```bash
# Use Claude as default
clinvk config set default_backend claude

# Or use Gemini
clinvk config set default_backend gemini
```

### 3. Done!

That's it for basic setup. clinvk works out of the box with sensible defaults.

## Configuration File

clinvk stores configuration in `~/.clinvk/config.yaml`. You can edit it directly or use `clinvk config set`.

### Minimal Configuration

```yaml
# ~/.clinvk/config.yaml
default_backend: claude
```

### Recommended Configuration

```yaml
# ~/.clinvk/config.yaml
default_backend: claude

# Show execution time in output
output:
  show_timing: true

# Keep sessions for 60 days
session:
  retention_days: 60
  auto_resume: true
```

## Common Scenarios

### Scenario 1: Using Multiple Backends

If you work with different AI models for different tasks:

```yaml
default_backend: claude

backends:
  claude:
    model: claude-opus-4-5-20251101    # For complex reasoning
  codex:
    model: o3                           # For code generation
  gemini:
    model: gemini-2.5-pro              # For general tasks
```

**Usage:**

```bash
# Use default (Claude)
clinvk "analyze this architecture"

# Specify backend for specific tasks
clinvk -b codex "generate unit tests"
clinvk -b gemini "summarize this document"
```

### Scenario 2: Auto-Approve Mode for Automation

For CI/CD or scripted workflows where you don't want interactive prompts:

```yaml
unified_flags:
  approval_mode: auto    # Auto-approve all actions
  output_format: json    # Machine-readable output
```

!!! warning "Security Note"
    Only use `auto` approval mode in trusted environments. The AI can execute file operations and commands.

### Scenario 3: Read-Only Analysis

For code review or analysis where the AI should not modify files:

```yaml
unified_flags:
  sandbox_mode: read-only    # No file modifications
  # approval_mode controls prompting behavior, not whether actions are allowed.
  # Use `sandbox_mode` to restrict file access. Use `always` for maximum safety.
  approval_mode: always
```

### Scenario 4: Team Shared Configuration

For consistent settings across a team, create a project-level config:

```bash
# In your project root
cat > .clinvk.yaml << 'EOF'
default_backend: claude
unified_flags:
  sandbox_mode: workspace    # Only access project files
backends:
  claude:
    system_prompt: "You are working on the MyApp project. Follow our coding standards."
EOF

# Use project config
clinvk --config .clinvk.yaml "review the auth module"
```

### Scenario 5: HTTP Server for Integration

For using clinvk as an API backend:

```yaml
server:
  host: "127.0.0.1"          # Localhost only (safe)
  port: 8080
  request_timeout_secs: 300  # 5 minutes for long tasks

# For LAN access (use with caution)
# server:
#   host: "0.0.0.0"          # All interfaces
#   port: 8080
```

### Scenario 6: Parallel Task Optimization

For batch processing or multi-perspective reviews:

```yaml
parallel:
  max_workers: 5       # Run up to 5 tasks simultaneously
  fail_fast: false     # Continue even if some tasks fail
  aggregate_output: true
```

## Backend-Specific Settings

### Claude Code

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all              # Or: read,write,edit
    system_prompt: "Be concise."
    extra_flags:
      - "--add-dir"
      - "./docs"                    # Include docs in context
```

**Available models:**

| Model | Best For |
|-------|----------|
| `claude-opus-4-5-20251101` | Complex reasoning, architecture |
| `claude-sonnet-4-20250514` | Balanced speed and capability |

### Codex CLI

```yaml
backends:
  codex:
    model: o3
    extra_flags:
      - "--quiet"                   # Less verbose output
```

### Gemini CLI

```yaml
backends:
  gemini:
    model: gemini-2.5-pro
    extra_flags:
      - "--sandbox"
```

## Environment Variables

Override any config with environment variables:

```bash
# Override default backend
export CLINVK_BACKEND=gemini

# Override models
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3
export CLINVK_GEMINI_MODEL=gemini-2.5-pro
```

**Priority order** (highest to lowest):

1. CLI flags (`--backend codex`)
2. Environment variables (for example, `CLINVK_BACKEND`)
3. Config file (`~/.clinvk/config.yaml`)
4. Built-in defaults

## Best Practices

### 1. Start with Defaults

clinvk works well out of the box. Only customize what you need.

### 2. Use Project-Level Configs

Keep project-specific settings in `.clinvk.yaml` in your repo:

```bash
clinvk --config .clinvk.yaml "your prompt"
```

### 3. Secure Your Server

If exposing the HTTP server:

```yaml
server:
  host: "127.0.0.1"    # Never use 0.0.0.0 without a reverse proxy
```

### 4. Set Appropriate Timeouts

For long-running tasks:

```yaml
server:
  request_timeout_secs: 600    # 10 minutes
```

### 5. Use Read-Only for Reviews

When you just want analysis without changes:

```yaml
unified_flags:
  sandbox_mode: read-only
```

## Troubleshooting

### Config Not Applied

```bash
# Check effective configuration
clinvk config show

# Verify config file location
ls -la ~/.clinvk/config.yaml
```

### Backend Not Available

```bash
# Check which backends are detected
clinvk config show | grep available

# Verify CLI is in PATH
which claude codex gemini
```

### Reset to Defaults

```bash
# Remove config file
rm ~/.clinvk/config.yaml

# Verify defaults
clinvk config show
```

## Configuration Templates

### Developer Workstation

```yaml
default_backend: claude

unified_flags:
  sandbox_mode: workspace

output:
  show_timing: true
  color: true

session:
  auto_resume: true
  retention_days: 30
```

### CI/CD Pipeline

```yaml
default_backend: claude

unified_flags:
  approval_mode: auto
  output_format: json

parallel:
  max_workers: 3
  fail_fast: true
```

### API Server

```yaml
default_backend: claude

server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300

output:
  format: json
```

## Next Steps

- [Configuration Reference](../reference/configuration.md) - Complete option reference
- [Environment Variables](../reference/environment.md) - All environment variables
- [config Command](../reference/commands/config.md) - CLI configuration commands
