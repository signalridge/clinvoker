---
title: Getting Started
description: Get up and running with clinvoker in minutes.
---

# Getting Started

Welcome to clinvoker! This guide will get you from zero to your first AI-orchestrated workflow in under 5 minutes.

## What You'll Learn

1. [Installation](installation.md) - Install clinvoker on your system
2. [First Prompt](first-prompt.md) - Run your first AI prompt
3. [Next Steps](next-steps.md) - Discover what to learn next

## Quick Preview

Here's what clinvoker enables:

```bash
# Run a simple prompt
clinvk "Explain the benefits of microservices"

# Use a specific backend
clinvk -b codex "Generate a REST API in Python"

# Compare multiple backends
clinvk compare --all-backends "Review this code: $(cat auth.go)"

# Execute tasks in parallel
clinvk parallel -f tasks.json

# Chain multiple operations
clinvk chain -f pipeline.json

# Start the HTTP API server
clinvk serve --port 8080
```

## Installation Options

### macOS

```bash
# Using Homebrew (recommended)
brew install signalridge/tap/clinvoker

# Or using the install script
curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
```

### Linux

```bash
# Using the install script
curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash

# Or download the binary directly
curl -L https://github.com/signalridge/clinvoker/releases/latest/download/clinvk-linux-amd64 -o clinvk
chmod +x clinvk
sudo mv clinvk /usr/local/bin/
```

### From Source

```bash
go install github.com/signalridge/clinvoker/cmd/clinvk@latest
```

## Verify Installation

```bash
clinvk --version
```

You should see the version number and build information.

## Prerequisites

Before using clinvoker, ensure you have at least one AI CLI backend installed:

- **Claude Code**: `npm install -g @anthropic-ai/claude-code`
- **Codex CLI**: Install from OpenAI
- **Gemini CLI**: Install from Google

## Start Learning

Ready to begin? Head to the [Installation guide](installation.md) for detailed setup instructions.
