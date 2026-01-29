# Backends

clinvk supports multiple AI CLI backends, each with unique strengths and characteristics.

## Supported Backends

<div class="grid cards" markdown>

-   :material-robot:{ .lg .middle } **[Claude Code](claude.md)**

    ---

    Anthropic's AI coding assistant with deep reasoning capabilities

-   :material-code-tags:{ .lg .middle } **[Codex CLI](codex.md)**

    ---

    OpenAI's code-focused CLI tool optimized for code generation

-   :material-google:{ .lg .middle } **[Gemini CLI](gemini.md)**

    ---

    Google's Gemini AI with broad knowledge and capabilities

</div>

## Backend Comparison

| Feature | Claude Code | Codex CLI | Gemini CLI |
|---------|-------------|-----------|------------|
| Binary | `claude` | `codex` | `gemini` |
| Default Model | claude-opus-4-5-20251101 | o3 | gemini-2.5-pro |
| Session Resume | `--resume` | `--session` | `-s` |
| Strengths | Complex reasoning, safety | Code generation | Broad knowledge |

## Backend Detection

clinvk automatically detects available backends by checking for their binaries in your PATH:

```bash
clinvk config show
```

Output shows which backends are available:

```yaml
backends:
  claude:
    enabled: true
    available: true  # 'claude' found in PATH
  codex:
    enabled: true
    available: false  # 'codex' not found
  gemini:
    enabled: true
    available: true  # 'gemini' found in PATH
```

## Selecting Backends

### Via CLI

```bash
clinvk --backend claude "prompt"
clinvk -b codex "prompt"
clinvk -b gemini "prompt"
```

### Via Configuration

Set a default backend in `~/.clinvk/config.yaml`:

```yaml
default_backend: claude
```

### Via Environment Variable

```bash
export CLINVK_BACKEND=codex
clinvk "prompt"  # Uses codex
```

## Backend-Specific Options

Each backend supports the unified options, plus its own specific flags:

### Unified Options

These work across all backends:

| Option | Description |
|--------|-------------|
| `model` | Model to use |
| `approval_mode` | Approval behavior |
| `sandbox_mode` | File access permissions |
| `max_turns` | Maximum agentic turns |
| `max_tokens` | Maximum response tokens |

### Backend-Specific Flags

Pass additional flags via `extra_flags` in config:

```yaml
backends:
  claude:
    extra_flags: ["--add-dir", "./docs"]
  codex:
    extra_flags: ["--quiet"]
  gemini:
    extra_flags: ["--sandbox"]
```

## Choosing a Backend

### Use Claude Code when:

- Working on complex, multi-step tasks
- Needing thorough code review and analysis
- Safety and accuracy are paramount

### Use Codex CLI when:

- Generating boilerplate code
- Writing tests
- Quick code transformations

### Use Gemini CLI when:

- Needing broad knowledge context
- Working with documentation
- General explanations

## Tips

!!! tip "Try Multiple Backends"
    Use `clinvk compare --all-backends` to see how different backends approach the same problem.

!!! tip "Match Backend to Task"
    Different backends excel at different tasks. Experiment to find the best fit for your workflow.

!!! tip "Configure Defaults"
    Set backend-specific models and options in your config file for a personalized experience.

## Next Steps

- [Claude Code Guide](claude.md)
- [Codex CLI Guide](codex.md)
- [Gemini CLI Guide](gemini.md)
