# Environment Variables

Complete reference for all environment variables supported by clinvk.

## Overview

Environment variables provide a convenient way to configure clinvk without modifying configuration files. They are especially useful for:

- CI/CD pipelines
- Docker containers
- Temporary overrides
- Per-project settings with direnv

## Variable Reference

### Core Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `CLINVK_BACKEND` | No | Default backend to use | `claude`, `codex`, `gemini` |
| `CLINVK_CLAUDE_MODEL` | No | Default model for Claude backend | `claude-opus-4-5-20251101` |
| `CLINVK_CODEX_MODEL` | No | Default model for Codex backend | `o3`, `o3-mini` |
| `CLINVK_GEMINI_MODEL` | No | Default model for Gemini backend | `gemini-2.5-pro` |

### Server Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `CLINVK_API_KEYS` | No | API keys for HTTP server authentication (comma-separated) | `key1,key2,key3` |
| `CLINVK_API_KEYS_GOPASS_PATH` | No | gopass path for retrieving API keys | `myproject/api-keys` |
| `CLINVK_CONFIG` | No | Path to custom configuration file | `/etc/clinvk/config.yaml` |
| `CLINVK_HOME` | No | Directory for clinvk data (sessions, config) | `~/.clinvk` |

### Backend API Keys

The following variables are passed directly to the respective backend CLIs:

| Variable | Backend | Description |
|----------|---------|-------------|
| `ANTHROPIC_API_KEY` | Claude | Anthropic API key |
| `OPENAI_API_KEY` | Codex | OpenAI API key |
| `GOOGLE_API_KEY` | Gemini | Google API key |

## Usage Examples

### Set Default Backend

```bash
export CLINVK_BACKEND=codex
clinvk "implement feature"  # Uses codex
```text

### Set Model per Backend

```bash
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3-mini

clinvk -b claude "complex task"  # Uses claude-sonnet
clinvk -b codex "quick task"     # Uses o3-mini
```text

### Temporary Override

Set a variable for a single command:

```bash
CLINVK_BACKEND=gemini clinvk "explain this"
```text

### API Keys for HTTP Server

```bash
export CLINVK_API_KEYS="prod-key-1,prod-key-2,dev-key-1"
clinvk serve
```text

Clients must include one of the following headers:

```bash
# Option 1: X-Api-Key header
curl -H "X-Api-Key: prod-key-1" http://localhost:8080/api/v1/prompt \
  -d '{"backend":"claude","prompt":"hello"}'

# Option 2: Authorization header
curl -H "Authorization: Bearer prod-key-1" http://localhost:8080/api/v1/prompt \
  -d '{"backend":"claude","prompt":"hello"}'
```text

## Priority

Environment variables have medium priority in the configuration hierarchy:

1. **CLI Flags** (highest priority)
2. **Environment Variables**
3. **Config File**
4. **Defaults** (lowest priority)

Example demonstrating priority:

```bash
export CLINVK_BACKEND=codex
clinvk -b claude "prompt"  # Uses claude (CLI flag wins)
```text

## Shell Configuration

### Bash

Add to `~/.bashrc` or `~/.bash_profile`:

```bash
# clinvk configuration
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
export CLINVK_CODEX_MODEL=o3
```text

### Zsh

Add to `~/.zshrc`:

```zsh
# clinvk configuration
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
export CLINVK_CODEX_MODEL=o3
```text

### Fish

Add to `~/.config/fish/config.fish`:

```fish
# clinvk configuration
set -gx CLINVK_BACKEND claude
set -gx CLINVK_CLAUDE_MODEL claude-opus-4-5-20251101
set -gx CLINVK_CODEX_MODEL o3
```text

## Per-Directory Configuration

Use [direnv](https://direnv.net/) for project-specific settings:

```bash
# .envrc in your project root
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3
```text

When you enter the directory, direnv automatically loads these variables.

## CI/CD Usage

### GitHub Actions

```yaml
name: AI Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    env:
      CLINVK_BACKEND: codex
      CLINVK_CODEX_MODEL: o3
      OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
    steps:
      - uses: actions/checkout@v4
      - name: Review code
        run: clinvk "review this PR for security issues"
```text

### GitLab CI

```yaml
ai-review:
  image: alpine/clinvk
  variables:
    CLINVK_BACKEND: claude
    CLINVK_CLAUDE_MODEL: claude-sonnet-4-20250514
    ANTHROPIC_API_KEY: $ANTHROPIC_API_KEY
  script:
    - clinvk "review the changes"
```text

### Jenkins

```groovy
pipeline {
    agent any
    environment {
        CLINVK_BACKEND = 'gemini'
        CLINVK_GEMINI_MODEL = 'gemini-2.5-pro'
        GOOGLE_API_KEY = credentials('google-api-key')
    }
    stages {
        stage('AI Analysis') {
            steps {
                sh 'clinvk "analyze code quality"'
            }
        }
    }
}
```text

## Docker Usage

### Dockerfile

```dockerfile
FROM alpine:latest
RUN apk add --no-cache clinvk

ENV CLINVK_BACKEND=claude
ENV CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101

ENTRYPOINT ["clinvk"]
```text

### Docker Run

```bash
docker run -e CLINVK_BACKEND=codex -e OPENAI_API_KEY=$OPENAI_API_KEY clinvk "prompt"
```text

### Docker Compose

```yaml
version: '3'
services:
  ai-task:
    image: clinvk
    environment:
      - CLINVK_BACKEND=claude
      - CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    command: clinvk "analyze codebase"
```text

## Kubernetes Usage

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: clinvk-config
data:
  CLINVK_BACKEND: "claude"
  CLINVK_CLAUDE_MODEL: "claude-opus-4-5-20251101"
---
apiVersion: v1
kind: Secret
metadata:
  name: clinvk-secrets
type: Opaque
stringData:
  ANTHROPIC_API_KEY: "your-api-key"
---
apiVersion: v1
kind: Pod
metadata:
  name: clinvk-job
spec:
  containers:
    - name: clinvk
      image: clinvk:latest
      envFrom:
        - configMapRef:
            name: clinvk-config
        - secretRef:
            name: clinvk-secrets
```bash

## Troubleshooting

### Variables Not Taking Effect

1. Verify the variable is exported: `export VAR=value` not just `VAR=value`
2. Check for typos in variable names
3. Ensure CLI flags aren't overriding your variables
4. Check if a config file is setting the same values

### Debug Environment

```bash
# Show all CLINVK variables
env | grep CLINVK

# Show specific variable
echo $CLINVK_BACKEND

# Run with debug output
CLINVK_DEBUG=1 clinvk "prompt"
```text

## See Also

- [Configuration Reference](configuration.md) - Configuration file options
- [config command](cli/config.md) - Manage configuration via CLI
