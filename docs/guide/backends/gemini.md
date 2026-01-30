# Gemini CLI Backend

Gemini CLI (Google) offers balanced performance with strong multilingual support and efficient resource usage.

---

## Overview

| Feature | Support |
|---------|---------|
| **Output Formats** | text, json, stream-json |
| **Sandbox Control** | ✓ Mapped to `--sandbox` |
| **Approval Mode** | ✓ Mapped to `--approval-mode` / `--yolo` |
| **Session Resume** | ✓ Supported |
| **Session Cleanup** | ✓ `--delete-session` support |
| **Ephemeral Mode** | ✓ Full support |

---

## Requirements

1. **Install Gemini CLI**

   ```bash
   npm install -g @google/gemini-cli
   ```

2. **Set API Key**

   ```bash
   export GEMINI_API_KEY="your-api-key"
   ```

   Or configure in Gemini:

   ```bash
   gemini config set api.key your-api-key
   ```

3. **Verify Installation**

   ```bash
   gemini --version
   clinvk -b gemini "test connection"
   ```

---

## How clinvk Uses Gemini

### Non-Interactive Execution

```
# clinvk builds this command:
gemini --output-format text \
  --model gemini-2.5-pro \
  --approval-mode auto_edit \
  "your prompt here"
```

### Session Resume

```
# clinvk builds this command:
gemini --resume <session_id> \
  --output-format text \
  "continue from previous context"
```

### Ephemeral Cleanup

```
# After ephemeral execution, clinvk cleans up:
gemini --delete-session <session_id>
```

---

## Configuration Options

### Full Configuration Example

```
backends:
  gemini:
    # Model selection
    model: gemini-2.5-pro

    # Default approval mode
    approval_mode: auto

    # Enable/disable this backend
    enabled: true
```

### Model Aliases

clinvk provides convenient aliases:

| Alias | Resolves To | Best For |
|-------|-------------|----------|
| `fast`, `quick` | `gemini-2.5-flash` | Quick tasks, high throughput |
| `balanced`, `default`, `best`, `powerful` | `gemini-2.5-pro` | General development |

```
# Use aliases
clinvk -b gemini -m fast "quick summary"
clinvk -b gemini -m balanced "code review"

# Or full model names
clinvk -b gemini -m gemini-2.5-pro "deep analysis"
```

---

## Approval Mode Mapping

| clinvk `approval_mode` | Gemini Flag | Behavior |
|------------------------|-------------|----------|
| `auto` | `--approval-mode auto_edit` | Auto-accept edits |
| `none` | `--yolo` | Skip all confirmations |
| `always` | `--approval-mode default` | Always ask |
| `default` | (none) | Use Gemini default |

```
# Auto-approve edits
clinvk -b gemini --approval-mode auto "refactor this code"

# Skip confirmations (use with caution)
clinvk -b gemini --approval-mode none --ephemeral "apply fixes"
```

---

## Sandbox Mode Mapping

| clinvk `sandbox_mode` | Gemini Flag | Behavior |
|-----------------------|-------------|----------|
| `read-only` | `--sandbox` | Enable sandbox (read-only) |
| `workspace` | `--sandbox` | Enable sandbox (workspace) |
| `full` | (none) | No sandbox flag |
| `default` | (none) | Use Gemini default |

**Note:** Gemini's sandbox is less granular than Codex. Both `read-only` and `workspace` map to `--sandbox`.

```
# Sandboxed execution
clinvk -b gemini --sandbox read-only "analyze this code"

# Full access
clinvk -b gemini --sandbox full "make system changes"
```

---

## Unique Features

### 1. Multilingual Support

Gemini excels at non-English content:

```
# Chinese code review
clinvk -b gemini "审查这段代码的安全问题" < auth.js

# Japanese documentation
clinvk -b gemini "このコードのドキュメントを作成してください"

# Mixed language context
clinvk -b gemini "Explain this Chinese comment: // 用户验证失败"
```

### 2. Efficient Resource Usage

Gemini Flash model is cost-effective for high-volume tasks:

```
{
  "tasks": [
    {
      "backend": "gemini",
      "prompt": "Process these 100 files",
      "model": "gemini-2.5-flash"
    }
  ]
}
```

### 3. Session Management

Gemini provides explicit session cleanup:

```
# clinvk automatically cleans up ephemeral sessions:
gemini --delete-session <id>
```

---

## Use Cases

### When to Choose Gemini

| Scenario | Why Gemini? |
|----------|-------------|
| **Multilingual projects** | Strong non-English support |
| **Cost-sensitive tasks** | Efficient Flash model |
| **Documentation** | Good at explaining concepts |
| **Security audits** | Thorough pattern matching |
| **Learning resources** | Clear explanations |
| **High-volume processing** | Fast and cost-effective |

### Gemini-Optimized Workflows

```
// Multilingual documentation
{
  "steps": [
    {
      "name": "english-docs",
      "backend": "claude",
      "prompt": "Write API documentation"
    },
    {
      "name": "chinese-docs",
      "backend": "gemini",
      "prompt": "Translate to Chinese: {{previous}}"
    },
    {
      "name": "japanese-docs",
      "backend": "gemini",
      "prompt": "Translate to Japanese: {{previous}}"
    }
  ]
}
```

```
// High-volume security scan
{
  "tasks": [
    {
      "backend": "gemini",
      "prompt": "Security audit 1000 files",
      "model": "gemini-2.5-flash"
    }
  ]
}
```

---

## Limitations

1. **Limited Model Options** - Only Flash and Pro models available
2. **Sandbox Granularity** - Less granular than Codex sandbox
3. **Tool Control** - No fine-grained tool restrictions
4. **Regional Availability** - Subject to Google's regional restrictions

---

## Troubleshooting

### "Backend 'gemini' is not available"

```
# Check if gemini CLI is installed
which gemini
gemini --version

# If not found, install:
npm install -g @google/gemini-cli
```

### "Authentication failed"

```
# Set your Gemini API key
export GEMINI_API_KEY="your-key"

# Or configure via gemini CLI
gemini config set api.key your-key
```

### "Model not found"

Check available models:

```
gemini models list
```

Then update your config:

```
backends:
  gemini:
    model: gemini-2.5-pro  # Use exact model name
```

---

## Configuration Reference

| Config Key | Type | Default | Description |
|------------|------|---------|-------------|
| `model` | string | `gemini-2.5-pro` | Default model |
| `approval_mode` | string | `default` | Permission behavior |
| `sandbox_mode` | string | `default` | Sandbox restrictions |
| `enabled` | bool | `true` | Enable this backend |

---

## Related

- [Backend Comparison](../backend-comparison.md) - Compare all backends
- [Configuration](../../reference/configuration.md) - Full config reference
- [Gemini CLI Documentation](https://github.com/google-gemini/gemini-cli) - Official docs
