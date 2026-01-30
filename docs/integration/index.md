# Integration Guide

clinvk is designed to integrate seamlessly with AI development workflows, from Claude Code Skills to LangChain agents to CI/CD pipelines.

## Integration Overview

```mermaid
flowchart LR
    subgraph ai ["AI development"]
        direction TB
        A1["Claude Code Skills"]
        A2["LangChain/LangGraph"]
        A3["Custom AI Agents"]
    end

    subgraph sdks ["SDKs"]
        direction TB
        B1["OpenAI SDK"]
        B2["Anthropic SDK"]
    end

    subgraph auto ["Automation"]
        direction TB
        C1["CI/CD Pipelines"]
        C2["Shell Scripts"]
    end

    A1 --> D["clinvk server"]
    A2 --> D
    A3 --> D
    B1 --> D
    B2 --> D
    C1 --> D
    C2 --> D

    D --> E1["Claude CLI"]
    D --> E2["Codex CLI"]
    D --> E3["Gemini CLI"]

    style ai fill:#e3f2fd,stroke:#1976d2
    style sdks fill:#fff3e0,stroke:#f57c00
    style auto fill:#f3e5f5,stroke:#7b1fa2
    style D fill:#ffecb3,stroke:#ffa000
    style E1 fill:#f3e5f5,stroke:#7b1fa2
    style E2 fill:#e8f5e9,stroke:#388e3c
    style E3 fill:#ffebee,stroke:#c62828
```bash

## Integration Methods

| Method | Use Case | API Endpoint |
|--------|----------|--------------|
| [Claude Code Skills](claude-code-skills.md) | Extend Claude with multi-backend capabilities | `/api/v1/*` |
| [LangChain/LangGraph](langchain-langgraph.md) | AI framework integration | `/openai/v1/*` |
| [CI/CD](ci-cd.md) | Automated code review, documentation | `/api/v1/*` |
| [Client Libraries](client-libraries.md) | Python, TypeScript, Go clients | All endpoints |
| [MCP Server](mcp-server.md) | Model Context Protocol integration | Future |

## Quick Integration Examples

### Claude Code Skill

```bash
# In your skill script
curl -s http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "gemini", "prompt": "Analyze this data"}'
```

### LangChain

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed"
)
```text

### OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)
```

### GitHub Actions

```yaml
- name: AI Code Review
  run: |
    payload=$(jq -n --arg prompt "Review:\n${{ steps.diff.outputs.changes }}" '{backend:"claude", prompt:$prompt}')
    curl -sS -X POST http://localhost:8080/api/v1/prompt \
      -H "Content-Type: application/json" \
      -d "$payload"
```

## Prerequisites

Before integrating, ensure:

1. **clinvk is installed**: See [Installation](../guide/installation.md)
2. **Backend CLIs are available**: At least one of `claude`, `codex`, or `gemini`
3. **Server is running**: `clinvk serve --port 8080`

## Choosing the Right Endpoint

| Your Situation | Recommended Endpoint |
|---------------|---------------------|
| Using OpenAI SDK or LangChain | `/openai/v1/*` |
| Using Anthropic SDK | `/anthropic/v1/*` |
| Building Claude Code Skills | `/api/v1/*` |
| Need parallel/chain execution | `/api/v1/*` |
| Simple REST integration | `/api/v1/*` |

## Next Steps

Choose your integration path:

- **AI Agent Development**: Start with [Claude Code Skills](claude-code-skills.md)
- **Framework Integration**: See [LangChain/LangGraph](langchain-langgraph.md)
- **Automation**: Check [CI/CD Integration](ci-cd.md)
- **Custom Development**: Review [Client Libraries](client-libraries.md)
