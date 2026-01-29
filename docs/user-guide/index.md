# User Guide

This guide covers all features of clinvk in detail.

## Overview

clinvk provides a unified interface to multiple AI coding assistants, with powerful orchestration features for complex workflows.

## Core Features

<div class="grid cards" markdown>

-   :material-console:{ .lg .middle } **[Basic Usage](basic-usage.md)**

    ---

    Learn the fundamentals of running prompts and using backends

-   :material-history:{ .lg .middle } **[Session Management](session-management.md)**

    ---

    Track and resume conversations across sessions

-   :material-layers-triple:{ .lg .middle } **[Parallel Execution](parallel-execution.md)**

    ---

    Run multiple tasks concurrently for faster workflows

-   :material-link-variant:{ .lg .middle } **[Chain Execution](chain-execution.md)**

    ---

    Pipeline prompts through multiple backends sequentially

-   :material-compare:{ .lg .middle } **[Backend Comparison](backend-comparison.md)**

    ---

    Compare responses from different AI backends

</div>

## Backend Guides

Learn about each supported backend:

<div class="grid cards" markdown>

-   :material-robot:{ .lg .middle } **[Claude Code](backends/claude.md)**

    ---

    Anthropic's AI coding assistant

-   :material-code-tags:{ .lg .middle } **[Codex CLI](backends/codex.md)**

    ---

    OpenAI's code-focused CLI tool

-   :material-google:{ .lg .middle } **[Gemini CLI](backends/gemini.md)**

    ---

    Google's Gemini AI assistant

</div>

## Workflow Examples

### Solo Development

```bash
# Start working on a feature
clinvk "implement user authentication"

# Continue the conversation
clinvk -c "add password hashing"

# Get a different perspective
clinvk -b gemini "review the implementation"
```

### Code Review

```bash
# Get reviews from multiple backends
clinvk compare --all-backends "review this PR for issues"
```

### Complex Tasks

```bash
# Run multiple independent tasks
clinvk parallel --file tasks.json

# Chain multiple perspectives
clinvk chain --file review-pipeline.json
```

## Tips

!!! tip "Use Session Continuity"
    Always use `--continue` or `clinvk resume` to maintain context in longer conversations.

!!! tip "Backend Selection"
    Different backends excel at different tasks. Claude is great for complex reasoning, Codex for code generation, and Gemini for broad knowledge.

!!! tip "Dry Run First"
    Use `--dry-run` to see what command would be executed without actually running it.
