# Codex CLI Backend

Codex CLI (OpenAI) is optimized for code generation and refactoring tasks, with strong JSON structured output support.

---

## Overview

| Feature | Support |
|---------|---------|
| **JSON Output** | ✓ Native JSONL support |
| **Sandbox Control** | ✓ Full mapping |
| **Approval Mode** | ✓ Full mapping |
| **Session Resume** | ✗ Not supported by Codex CLI |
| **Output Formats** | json, stream-json (text rendered by clinvk) |
| **Ephemeral Mode** | ✓ Session cleanup |

---

## Requirements

1. **Install Codex CLI**

   ```bash
   npm install -g @openai/codex
   ```

2. **Set API Key**

   ```bash
   export OPENAI_API_KEY="sk-..."
   ```

   Or configure in Codex:

   ```bash
   codex config set apiKey sk-...
   ```

3. **Verify Installation**

   ```bash
   codex --version
   clinvk -b codex "test connection"
   ```

---

## How clinvk Uses Codex

### Non-Interactive Execution

```
# clinvk builds this command:
codex exec --json \
  --model o3 \
  --ask-for-approval on-request \
  --sandbox workspace-write \
  "your prompt here"
```

### Session Handling

```
# Resume (if supported by backend):
codex exec resume <session_id> --json ...
```

**Note:** Codex CLI has limited session persistence. clinvk handles cleanup for ephemeral mode.

### JSON Parsing

clinvk always requests JSON output internally:

```
# clinvk parses JSONL responses for:
# - Content extraction
# - Session ID tracking
# - Error detection
```

---

## Configuration Options

### Full Configuration Example

```
backends:
  codex:
    # Model selection
    model: o3

    # Default approval mode
    approval_mode: auto

    # Default sandbox mode
    sandbox_mode: workspace

    # Enable/disable this backend
    enabled: true
```

### Model Aliases

clinvk provides convenient aliases:

| Alias | Resolves To | Best For |
|-------|-------------|----------|
| `fast`, `quick` | `gpt-4.1-mini` | Quick tasks, simple edits |
| `balanced`, `default` | `gpt-5.2` | General development |
| `best`, `powerful` | `gpt-5-codex` | Complex refactoring |

```
# Use aliases
clinvk -b codex -m fast "fix typo"
clinvk -b codex -m balanced "add logging"
clinvk -b codex -m best "refactor architecture"

# Or full model names
clinvk -b codex -m o3 "complex task"
```

---

## Approval Mode Mapping

| clinvk `approval_mode` | Codex Flag | Behavior |
|------------------------|------------|----------|
| `auto` | `--ask-for-approval on-request` | Ask only for non-read operations |
| `none` | `--ask-for-approval never` | Never ask (dangerous) |
| `always` | `--ask-for-approval untrusted` | Always ask for confirmation |
| `default` | (none) | Use Codex default |

```
# Safe for trusted code
clinvk -b codex --approval-mode auto "refactor this module"

# CI/CD automation (use with caution)
clinvk -b codex --approval-mode none --ephemeral "apply all fixes"
```

---

## Sandbox Mode Mapping

| clinvk `sandbox_mode` | Codex Flag | Behavior |
|-----------------------|------------|----------|
| `read-only` | `--sandbox read-only` | Read files only |
| `workspace` | `--sandbox workspace-write` | Write within workspace |
| `full` | `--sandbox danger-full-access` | Full filesystem access |
| `default` | (none) | Use Codex default |

```
# Safe sandbox for CI
clinvk -b codex --sandbox read-only "analyze this code"

# Normal development
clinvk -b codex --sandbox workspace "implement feature"

# Careful - full access
clinvk -b codex --sandbox full "system-wide changes"
```

---

## Unique Features

### 1. Structured JSON Output

Codex excels at producing parseable JSON:

```
clinvk -b codex -o json "Generate a list of TODOs from this code" < src/app.js
```

Output:

```
{
  "backend": "codex",
  "content": "1. Add error handling...",
  "exit_code": 0,
  "duration_seconds": 4.32
}
```

### 2. Fast Code Generation

Best for:

- Quick edits and fixes
- Boilerplate generation
- Test file creation
- Documentation comments

```
clinvk -b codex "Generate Jest tests for auth.js"
```

### 3. Inline Code Changes

Codex is optimized for in-place code modifications:

```
clinvk -b codex --approval-mode auto "Add TypeScript types to this file" < utils.ts
```

---

## Use Cases

### When to Choose Codex

| Scenario | Why Codex? |
|----------|------------|
| **Quick edits** | Fast response times |
| **Code generation** | Strong at boilerplate |
| **Refactoring** | Good at pattern matching |
| **JSON processing** | Native structured output |
| **CI/CD automation** | Reliable non-interactive mode |
| **Test generation** | Good test coverage |

### Codex-Optimized Workflows

```
// Quick fixes in parallel
{
  "tasks": [
    {
      "backend": "codex",
      "prompt": "Fix linting errors",
      "approval_mode": "auto"
    },
    {
      "backend": "codex",
      "prompt": "Add missing imports",
      "approval_mode": "auto"
    }
  ]
}
```

```
// Code generation chain
{
  "steps": [
    {
      "name": "generate",
      "backend": "codex",
      "prompt": "Generate API client code",
      "approval_mode": "auto"
    },
    {
      "name": "test",
      "backend": "codex",
      "prompt": "Generate tests for: {{previous}}",
      "approval_mode": "auto"
    }
  ]
}
```

---

## Limitations

1. **No Session Persistence** - Codex CLI sessions cannot be resumed across invocations
2. **No System Prompts** - Cannot customize system behavior like Claude
3. **Limited Tool Control** - No fine-grained tool restrictions
4. **OpenAI Dependency** - Requires OpenAI API key and rate limits

---

## Troubleshooting

### "Backend 'codex' is not available"

```
# Check if codex CLI is installed
which codex
codex --version

# If not found, install:
npm install -g @openai/codex
```

### "Authentication failed"

```
# Set your OpenAI API key
export OPENAI_API_KEY="sk-..."

# Or configure via codex CLI
codex config set apiKey sk-...
```

### "Rate limit exceeded"

You're hitting OpenAI's rate limits:

```
# Slow down in config
parallel:
  max_workers: 2  # Reduce concurrent requests
```

Or use a different backend temporarily:

```
clinvk -b claude "same prompt"
```

---

## Configuration Reference

| Config Key | Type | Default | Description |
|------------|------|---------|-------------|
| `model` | string | `o3` | Default model |
| `approval_mode` | string | `default` | Permission behavior |
| `sandbox_mode` | string | `default` | Sandbox restrictions |
| `enabled` | bool | `true` | Enable this backend |

---

## Related

- [Backend Comparison](../backend-comparison.md) - Compare all backends
- [Configuration](../../reference/configuration.md) - Full config reference
- [OpenAI Codex Documentation](https://platform.openai.com/docs/guides/codex) - Official docs
