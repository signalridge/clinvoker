---
title: Next Steps
description: Continue your clinvoker journey with these resources.
---

# Next Steps

You've successfully run your first prompts with clinvoker. Here's where to go next based on your interests.

## By Use Case

### I want to automate code reviews

1. Read the [Automated Code Review](../use-cases/automated-code-review.md) use case
2. Follow the [CI/CD Integration tutorial](../tutorials/ci-cd-integration.md)
3. Set up parallel execution for multi-backend reviews

### I want to integrate with my existing tools

1. Learn about [LangChain integration](../tutorials/langchain-integration.md)
2. Explore the [HTTP API](../reference/api/rest-api.md)
3. Check out [Claude Code Skills](../integration/claude-code-skills.md)

### I want to build complex AI workflows

1. Study [Chain Execution](../how-to/chain-execution.md)
2. Read about [Parallel Execution](../how-to/parallel-execution.md)
3. See the [AI Team Collaboration](../use-cases/ai-team-collaboration.md) use case

### I want to deploy as a service

1. Review the [API Gateway Pattern](../use-cases/api-gateway-pattern.md)
2. Learn about [Server configuration](../reference/configuration.md)
3. Set up [Monitoring and Observability](https://github.com/signalridge/clinvoker)

## By Skill Level

### Beginner

- [Basic Usage](../how-to/basic-usage.md) - Complete command reference
- [Session Management](../how-to/session-management.md) - Working with sessions
- [Configuration](../how-to/configuration.md) - Customize clinvoker

### Intermediate

- [Parallel Execution](../how-to/parallel-execution.md) - Run multiple tasks
- [Chain Execution](../how-to/chain-execution.md) - Sequential workflows
- [Backend Comparison](../how-to/backend-comparison.md) - Compare backends

### Advanced

- [Building AI Skills](../tutorials/building-ai-skills.md) - Create Claude Code Skills
- [Adding Backends](../development/adding-backends.md) - Extend clinvoker
- [Architecture Overview](../architecture/overview.md) - Deep dive into internals

## Learning Paths

### Path 1: CLI Power User

1. Master all [CLI commands](../reference/commands/index.md)
2. Learn [Configuration](../reference/configuration.md) options
3. Set up [Shell completion](https://github.com/signalridge/clinvoker)
4. Create aliases and shortcuts

### Path 2: Integration Developer

1. Understand the [REST API](../reference/api/rest-api.md)
2. Implement [OpenAI-compatible](../reference/api/openai-compatible.md) clients
3. Build [LangChain integrations](../tutorials/langchain-integration.md)
4. Create custom SDK wrappers

### Path 3: DevOps Engineer

1. Deploy the [API Gateway](../use-cases/api-gateway-pattern.md)
2. Set up [CI/CD pipelines](../tutorials/ci-cd-integration.md)
3. Configure [Monitoring](https://github.com/signalridge/clinvoker)
4. Implement security policies

## Quick Reference

### Common Commands

```bash
# Basic prompt
clinvk "Your prompt here"

# Specific backend
clinvk -b codex "Generate code"

# Parallel tasks
clinvk parallel -f tasks.json

# Chain workflow
clinvk chain -f pipeline.json

# Compare backends
clinvk compare --all-backends "Question"

# Start server
clinvk serve --port 8080

# Session management
clinvk sessions list
clinvk resume --last
```

### Configuration File

```yaml
# ~/.clinvk/config.yaml
default_backend: claude

unified_flags:
  approval_mode: default
  sandbox_mode: default

backends:
  claude:
    model: claude-sonnet-4
  codex:
    model: gpt-4o
  gemini:
    model: gemini-pro

session:
  auto_resume: true
  retention_days: 30
```

## Community Resources

- [GitHub Repository](https://github.com/signalridge/clinvoker) - Source code and issues
- [Examples](https://github.com/signalridge/clinvoker/tree/main/examples) - Sample configurations
- [FAQ](../development/faq.md) - Common questions
- [Troubleshooting](../development/troubleshooting.md) - Problem solving

## Getting Help

- **Documentation**: Browse the full [documentation](../index.md)
- **Issues**: Report bugs on [GitHub](https://github.com/signalridge/clinvoker/issues)
- **Contributing**: See the [Contributing Guide](../development/contributing.md)

## What's Next?

Ready to dive deeper? Here are three recommended next steps:

1. **Try Parallel Execution** - Run multiple AI tasks simultaneously
2. **Set Up a Chain** - Create a multi-step workflow
3. **Explore Use Cases** - See real-world patterns in action

Pick one that interests you and get started!
