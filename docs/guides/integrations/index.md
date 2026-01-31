# Integration Guide

clinvk is designed to integrate seamlessly with AI development workflows, from Claude Code Skills to LangChain agents to CI/CD pipelines.

## Integration Overview

clinvk integrates with various tools and frameworks through its HTTP API:

**Integration Methods:**

| Category | Tools | API Endpoint |
|----------|-------|--------------|
| AI Development | Claude Code Skills, LangChain/LangGraph, Custom Agents | `/api/v1/*` or `/openai/v1/*` |
| SDKs | OpenAI SDK, Anthropic SDK | `/openai/v1/*` or `/anthropic/v1/*` |
| Automation | CI/CD Pipelines, Shell Scripts | `/api/v1/*` |

**Data Flow:**

```bash
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  AI Development │     │      SDKs       │     │   Automation    │
│  ─────────────  │     │  ────────────   │     │  ────────────   │
│  Claude Skills  │     │   OpenAI SDK    │     │  CI/CD Pipelines│
│  LangChain      │────▶│   Anthropic SDK │────▶│  Shell Scripts  │
│  Custom Agents  │     └─────────────────┘     └─────────────────┘
└─────────────────┘              │                       │
                                 └───────────┬───────────┘
                                             ▼
                                   ┌─────────────────┐
                                   │  clinvk server  │
                                   └────────┬────────┘
                                            │
                       ┌────────────────────┼────────────────────┐
                       ▼                    ▼                    ▼
               ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
               │  Claude CLI  │   │  Codex CLI   │   │  Gemini CLI  │
               └──────────────┘   └──────────────┘   └──────────────┘
```

## Integration Methods

| Method | Use Case | API Endpoint |
|--------|----------|--------------|
| [Claude Code Skills](claude-code-skills.md) | Extend Claude with multi-backend capabilities | `/api/v1/*` |
| [LangChain/LangGraph](langchain-langgraph.md) | AI framework integration | `/openai/v1/*` |
| [CI/CD](ci-cd/index.md) | Automated code review, documentation | `/api/v1/*` |
| [Client Libraries](../../reference/api/index.md) | Python, TypeScript, Go clients | All endpoints |
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
```

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

1. **clinvk is installed**: See [Installation](../../tutorials/getting-started.md)
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
- **Automation**: Check [CI/CD Integration](ci-cd/index.md)
- **Custom Development**: Review [Client Libraries](../../reference/api/index.md)
