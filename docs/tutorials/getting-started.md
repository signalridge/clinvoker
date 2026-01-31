---
title: Getting Started
description: Complete guide to installing clinvoker, running your first prompt, and understanding core concepts.
---

# Getting Started with clinvoker

Welcome to clinvoker - a unified command-line interface for orchestrating multiple AI coding assistants. This comprehensive tutorial will guide you through installation, configuration, and your first interactions with the tool.

## What is clinvoker?

clinvoker (pronounced "see-el-in-voker") is a universal gateway that unifies access to multiple AI coding assistants including Claude Code, Codex CLI, and Gemini CLI. Instead of learning different commands and interfaces for each AI tool, you use a single, consistent interface.

### Key Benefits

- **Unified Interface**: One command structure for all backends
- **Session Management**: Persistent conversations across sessions
- **Parallel Execution**: Run multiple AI tasks simultaneously
- **Chain Workflows**: Pass output from one backend to another
- **HTTP API**: Deploy as a service for integrations

---

## Prerequisites

Before installing clinvoker, ensure you have:

### System Requirements

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.24+ | Only needed for building from source |
| Operating System | Linux, macOS, Windows | AMD64 and ARM64 supported |
| Memory | 512MB minimum | For running the CLI |
| Disk Space | 100MB | For binary and configuration |

### Backend Prerequisites

clinvoker requires at least one AI backend to be useful:

| Backend | Installation Command | Documentation |
|---------|---------------------|---------------|
| Claude Code | `npm install -g @anthropic-ai/claude-code` | [Claude.ai](https://claude.ai/claude-code) |
| Codex CLI | `npm install -g @openai/codex` | [GitHub](https://github.com/openai/codex-cli) |
| Gemini CLI | `npm install -g @google/gemini-cli` | [GitHub](https://github.com/google/gemini-cli) |

Verify backend installation:

```bash
# Check Claude Code
which claude && claude --version

# Check Codex CLI
which codex && codex --version

# Check Gemini CLI
which gemini && gemini --version
```yaml

---

## Installation Methods

Choose the installation method that best fits your environment:

### Method Comparison

| Method | Best For | Pros | Cons |
|--------|----------|------|------|
| Quick Install Script | First-time users | Fastest setup, automatic PATH config | Requires curl/PowerShell |
| Package Manager | Regular use | Easy updates, dependency management | May not have latest version |
| Manual Download | Air-gapped systems | Full control over version | Manual updates required |
| Build from Source | Developers | Latest features, customization | Requires Go toolchain |

### 1. Quick Install (Recommended)

The fastest way to get started:

=== "macOS/Linux"

    ```bash
    curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
    ```

    This script will:
    1. Detect your operating system and architecture
    2. Download the appropriate binary
    3. Install to `~/.local/bin` (or `/usr/local/bin` with sudo)
    4. Update your PATH if needed

=== "Windows (PowerShell)"

    ```powershell
    irm https://raw.githubusercontent.com/signalridge/clinvoker/main/install.ps1 | iex
    ```

    The PowerShell script performs the same steps, installing to `%LOCALAPPDATA%\Programs\clinvk`.

### 2. Package Managers

#### Homebrew (macOS/Linux)

```bash
# Add the tap
brew tap signalridge/tap

# Install clinvoker
brew install clinvk

# Upgrade later
brew upgrade clinvk
```text

#### Scoop (Windows)

```bash
# Add the bucket
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket

# Install
scoop install clinvk

# Upgrade
scoop update clinvk
```text

#### Nix (Linux/macOS)

```bash
# Run directly without installing
nix run github:signalridge/clinvoker

# Install to profile
nix profile install github:signalridge/clinvoker

# Use in a flake.nix
{
  inputs.clinvoker.url = "github:signalridge/clinvoker";
  nixpkgs.overlays = [ clinvoker.overlays.default ];
}
```text

#### Arch Linux (AUR)

```bash
# Using yay (recommended)
yay -S clinvk-bin

# Or build from source
yay -S clinvk
```text

#### Debian/Ubuntu

```bash
# Download from releases page
wget https://github.com/signalridge/clinvoker/releases/download/v0.1.0/clinvk_0.1.0_amd64.deb
sudo dpkg -i clinvk_*.deb
```text

#### RPM-based (Fedora/RHEL)

```bash
# Download from releases page
wget https://github.com/signalridge/clinvoker/releases/download/v0.1.0/clinvk-0.1.0-1.x86_64.rpm
sudo rpm -i clinvk-*.rpm
```bash

### 3. Manual Download

Download pre-built binaries from [GitHub Releases](https://github.com/signalridge/clinvoker/releases):

=== "Linux (AMD64)"

    ```bash
    VERSION="0.1.0-alpha"
    curl -LO "https://github.com/signalridge/clinvoker/releases/download/v${VERSION}/clinvoker_${VERSION}_linux_amd64.tar.gz"
    tar xzf "clinvoker_${VERSION}_linux_amd64.tar.gz"
    sudo mv clinvk /usr/local/bin/
    ```

=== "macOS (ARM64 - Apple Silicon)"

    ```bash
    VERSION="0.1.0-alpha"
    curl -LO "https://github.com/signalridge/clinvoker/releases/download/v${VERSION}/clinvoker_${VERSION}_darwin_arm64.tar.gz"
    tar xzf "clinvoker_${VERSION}_darwin_arm64.tar.gz"
    sudo mv clinvk /usr/local/bin/
    ```

=== "Windows"

    Download `clinvoker_<version>_windows_amd64.zip` and extract to a directory in your PATH.

### 4. Build from Source

Requires Go 1.24 or later:

```bash
# Using go install
go install github.com/signalridge/clinvoker/cmd/clinvk@latest

# Or clone and build
git clone https://github.com/signalridge/clinvoker.git
cd clinvoker
go build -o clinvk ./cmd/clinvk
sudo mv clinvk /usr/local/bin/
```yaml

---

## Verify Installation

After installation, verify everything is working:

```bash
# Check version
clinvk version
```text

Expected output:

```text
clinvk version v0.1.0-alpha
  commit: abc1234
  built:  2025-01-27T00:00:00Z
```text

Check detected backends:

```bash
clinvk config show
```text

You should see a list of available backends based on what's installed on your system.

---

## Environment Setup

### API Keys

Each backend requires its own API key configuration:

| Backend | Configuration Method | Environment Variable |
|---------|---------------------|---------------------|
| Claude | `claude config set api_key <key>` | `ANTHROPIC_API_KEY` |
| Codex | `codex config set api_key <key>` | `OPENAI_API_KEY` |
| Gemini | `gemini config set api_key <key>` | `GOOGLE_API_KEY` |

### Default Backend

Set your preferred default backend:

```bash
# Via environment variable (temporary)
export CLINVK_BACKEND=claude

# Via config (permanent)
clinvk config set default_backend claude
```bash

### Configuration File

Create `~/.clinvk/config.yaml`:

```yaml
# Default backend when -b is not specified
default_backend: claude

# Unified flags apply to all backends
unified_flags:
  approval_mode: default
  sandbox_mode: default

# Backend-specific settings
backends:
  claude:
    model: claude-sonnet-4-20250514
  codex:
    model: o3
  gemini:
    model: gemini-2.5-pro

# Session management
session:
  auto_resume: true
  retention_days: 30
```yaml

---

## Your First Prompt

### Basic Usage

Run your first prompt with the default backend:

```bash
clinvk "Explain the SOLID principles in software engineering"
```text

This sends your prompt to the default backend (Claude Code by default) and displays the response.

### Specifying a Backend

Use different backends for different strengths:

```bash
# Claude excels at complex reasoning and architecture
clinvk -b claude "Design a microservices architecture for an e-commerce platform"

# Codex is optimized for code generation
clinvk -b codex "Implement a quicksort algorithm in Python"

# Gemini provides broad knowledge and explanations
clinvk -b gemini "Explain the trade-offs between SQL and NoSQL databases"
```text

### Working Directory

Provide context by setting the working directory:

```bash
# Review code in current directory
clinvk -w . "Review this codebase for security issues"

# Analyze a specific project
clinvk -w /path/to/project "Explain the architecture of this application"
```yaml

---

## Output Formats Explained

clinvoker supports three output formats, each suited for different use cases:

### Text Format (Default)

Human-readable output ideal for interactive use:

```bash
clinvk -o text "Explain REST APIs"
```text

**Characteristics:**
- Clean, formatted text
- No metadata or structure
- Best for terminal reading
- Suitable for piping to other tools

**When to use:** Interactive sessions, reading responses, quick queries

### JSON Format

Structured output with metadata for programmatic use:

```bash
clinvk -o json "Explain REST APIs"
```text

**Output structure:**

```json
{
  "output": "REST (Representational State Transfer)...",
  "backend": "claude",
  "model": "claude-sonnet-4-20250514",
  "duration_ms": 2450,
  "tokens_used": 450,
  "session_id": "sess_abc123",
  "timestamp": "2025-01-27T10:30:00Z"
}
```text

**When to use:** Scripting, logging, storing results, API integrations

### Stream JSON Format

Real-time streaming for long-running tasks:

```bash
clinvk -o stream-json "Write a comprehensive guide to Go concurrency"
```text

**Characteristics:**
- Emits JSON objects as they become available
- Shows progress in real-time
- Each chunk contains a portion of the response
- Final object contains complete metadata

**When to use:** Long-form content, real-time applications, progress monitoring

### Format Comparison

| Format | Human Readable | Machine Readable | Real-time | Metadata |
|--------|---------------|------------------|-----------|----------|
| text | Yes | No | No | No |
| json | No | Yes | No | Yes |
| stream-json | Partial | Yes | Yes | Yes |

---

## Session Management Basics

### Understanding Sessions

A session is a persistent conversation context with an AI backend. clinvoker automatically:

1. Creates a new session when you run a prompt
2. Associates subsequent prompts with the same session
3. Maintains context across multiple interactions

### Listing Sessions

View all active sessions:

```bash
clinvk sessions list
```text

Output:

```text
ID          BACKEND  CREATED              STATUS   TAGS
sess_abc12  claude   2025-01-27 10:00:00  active   project-x
sess_def34  codex    2025-01-27 09:30:00  closed   -
```text

### Continuing Sessions

Resume a previous conversation:

```bash
# Continue the most recent session
clinvk --continue "What about caching strategies?"

# Or use the resume command
clinvk resume --last

# Resume a specific session
clinvk resume sess_abc12
```text

### Session Best Practices

- **Use tags** to organize sessions by project or topic
- **Clean up old sessions** periodically to save disk space
- **Use `--ephemeral`** for one-off queries that don't need persistence

---

## Troubleshooting

### Issue 1: "Backend not available"

**Symptoms:**
```text
Error: backend "claude" not available
```bash

**Causes and Solutions:**

1. **Backend not installed**
   ```bash
   # Verify installation
   which claude

   # Install if missing
   npm install -g @anthropic-ai/claude-code
   ```

2. **Backend not in PATH**
   ```bash
   # Find the binary
   find /usr -name "claude" 2>/dev/null

   # Add to PATH
   export PATH="$PATH:/path/to/claude"
   ```

3. **Backend disabled in config**
   ```bash
   # Check config
   clinvk config show

   # Enable backend
   clinvk config set backends.claude.enabled true
   ```

### Issue 2: "API key not configured"

**Symptoms:**
```text
Error: authentication failed for backend "claude"
```text

**Solution:**

```bash
# Configure API key for Claude
claude config set api_key $ANTHROPIC_API_KEY

# Or set environment variable
export ANTHROPIC_API_KEY="sk-ant-..."
```text

### Issue 3: "Session not found"

**Symptoms:**
```text
Error: no sessions found for backend "claude"
```text

**Explanation:** This is normal for your first run. Sessions are created after your first successful prompt.

**Solution:**

```bash
# Run your first prompt to create a session
clinvk "Hello, world!"

# Now sessions will be available
clinvk sessions list
```text

If you still see issues:

```bash
# Check session directory
ls -la ~/.clinvk/sessions/

# Reset if corrupted
clinvk sessions cleanup
```yaml

---

## Next Steps

Now that you have clinvoker installed and working, explore these paths based on your goals:

### For Code Review Automation

1. [Multi-Backend Code Review](multi-backend-code-review.md) - Set up parallel reviews
2. [CI/CD Integration](ci-cd-integration.md) - Automate in your pipeline
3. [Parallel Execution](../guides/parallel.md) - Run multiple reviews simultaneously

### For Tool Integration

1. [LangChain Integration](langchain-integration.md) - Connect to LangChain
2. [HTTP Server](../guides/http-server.md) - Deploy as an API service
3. [Claude Code Skills](../guides/integrations/claude-code-skills.md) - Build custom skills

### For Complex Workflows

1. [Chain Execution](../guides/chains.md) - Create multi-step pipelines
2. [Building AI Skills](building-ai-skills.md) - Develop specialized AI agents
3. [Architecture Overview](../concepts/architecture.md) - Understand internals

### Quick Reference

```bash
# Basic usage
clinvk "Your prompt here"
clinvk -b codex "Generate code"

# Parallel execution
clinvk parallel -f tasks.json

# Chain workflow
clinvk chain -f pipeline.json

# Server mode
clinvk serve --port 8080

# Session management
clinvk sessions list
clinvk resume --last
```text

---

## Summary

You have successfully:

- Installed clinvoker using your preferred method
- Configured API keys and default backend
- Run your first prompts with different backends
- Explored output formats and their use cases
- Learned session management basics
- Identified solutions to common issues

clinvoker unifies multiple AI assistants under one interface, enabling powerful workflows like parallel execution, chain processing, and CI/CD integration. The next tutorials will show you how to leverage these capabilities for real-world scenarios.
