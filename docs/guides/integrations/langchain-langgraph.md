# LangChain / LangGraph Integration

This guide explains how to integrate clinvk with LangChain and LangGraph for building AI-powered applications and agents.

## Overview

clinvk's OpenAI-compatible endpoint (`/openai/v1/*`) allows seamless integration with LangChain and other AI frameworks that support the OpenAI API format.

```mermaid
flowchart LR
    A["LangChain app"] --> B["ChatOpenAI"]
    B --> C["clinvk /openai/v1"]
    C --> D["Backend CLI (claude/codex/gemini)"]

    style A fill:#e3f2fd,stroke:#1976d2
    style B fill:#fff3e0,stroke:#f57c00
    style C fill:#ffecb3,stroke:#ffa000
    style D fill:#f3e5f5,stroke:#7b1fa2
```bash

## OpenAI SDK Compatibility

clinvk is fully compatible with the OpenAI Python SDK:

```python
from openai import OpenAI

# Configure clinvk as the base URL
client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # only required if API keys are enabled
)

# Use any backend by setting the model name
response = client.chat.completions.create(
    model="claude",  # Backend name: claude, codex, or gemini
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Explain this code snippet."}
    ]
)

print(response.choices[0].message.content)
```text

## LangChain Integration

### Basic Chat Model

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, SystemMessage

# Initialize with clinvk endpoint
llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed",
    temperature=0.7
)

# Simple invocation
messages = [
    SystemMessage(content="You are a code review expert."),
    HumanMessage(content="Review this function for bugs.")
]

response = llm.invoke(messages)
print(response.content)
```text

### Using with Chains

```python
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser

# Setup
llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed"
)

# Create a chain
prompt = ChatPromptTemplate.from_messages([
    ("system", "You are a code documentation expert."),
    ("user", "Generate documentation for this code:\n{code}")
])

chain = prompt | llm | StrOutputParser()

# Execute
result = chain.invoke({"code": "def hello(): return 'world'"})
print(result)
```text

### Multiple Backends in One Application

```python
from langchain_openai import ChatOpenAI

# Create multiple LLM instances for different backends
claude_llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed"
)

codex_llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="codex",
    api_key="not-needed"
)

gemini_llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="gemini",
    api_key="not-needed"
)

# Use different backends for different tasks
architecture_review = claude_llm.invoke("Review the architecture...")
code_generation = codex_llm.invoke("Generate a function that...")
data_analysis = gemini_llm.invoke("Analyze this dataset...")
```text

## LangGraph Agent Integration

### Basic Agent with Tools

```python
from langchain_openai import ChatOpenAI
from langchain.tools import tool
from langgraph.prebuilt import create_react_agent
import httpx

# Define a tool that uses clinvk's parallel execution
@tool
def parallel_code_review(code: str) -> dict:
    """Review code using multiple AI backends in parallel."""
    response = httpx.post(
        "http://localhost:8080/api/v1/parallel",
        json={
            "tasks": [
                {
                    "backend": "claude",
                    "prompt": f"Review architecture and design:\n{code}"
                },
                {
                    "backend": "codex",
                    "prompt": f"Review for performance issues:\n{code}"
                },
                {
                    "backend": "gemini",
                    "prompt": f"Review for security vulnerabilities:\n{code}"
                }
            ]
        },
        timeout=120
    )
    return response.json()

@tool
def chain_documentation(code: str) -> str:
    """Generate documentation through a multi-step pipeline."""
    response = httpx.post(
        "http://localhost:8080/api/v1/chain",
        json={
            "steps": [
                {
                    "name": "analyze",
                    "backend": "claude",
                    "prompt": f"Analyze this code structure:\n{code}"
                },
                {
                    "name": "document",
                    "backend": "codex",
                    "prompt": "Generate API docs based on: {{previous}}"
                }
            ]
        },
        timeout=120
    )
    results = response.json()
    return results["results"][-1]["output"]

# Create the agent
llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed"
)

agent = create_react_agent(llm, [parallel_code_review, chain_documentation])

# Run the agent
result = agent.invoke({
    "messages": [{"role": "user", "content": "Review this code: def add(a, b): return a + b"}]
})
```text

### Custom Agent Graph

```python
from typing import Annotated, TypedDict
from langgraph.graph import StateGraph, END
from langchain_openai import ChatOpenAI

class AgentState(TypedDict):
    code: str
    architecture_review: str
    performance_review: str
    security_review: str
    final_report: str

def review_architecture(state: AgentState) -> AgentState:
    llm = ChatOpenAI(
        base_url="http://localhost:8080/openai/v1",
        model="claude",
        api_key="not-needed"
    )
    result = llm.invoke(f"Review architecture:\n{state['code']}")
    return {"architecture_review": result.content}

def review_performance(state: AgentState) -> AgentState:
    llm = ChatOpenAI(
        base_url="http://localhost:8080/openai/v1",
        model="codex",
        api_key="not-needed"
    )
    result = llm.invoke(f"Review performance:\n{state['code']}")
    return {"performance_review": result.content}

def review_security(state: AgentState) -> AgentState:
    llm = ChatOpenAI(
        base_url="http://localhost:8080/openai/v1",
        model="gemini",
        api_key="not-needed"
    )
    result = llm.invoke(f"Review security:\n{state['code']}")
    return {"security_review": result.content}

def generate_report(state: AgentState) -> AgentState:
    llm = ChatOpenAI(
        base_url="http://localhost:8080/openai/v1",
        model="claude",
        api_key="not-needed"
    )
    prompt = f"""Generate a final report based on these reviews:

Architecture: {state['architecture_review']}
Performance: {state['performance_review']}
Security: {state['security_review']}
"""
    result = llm.invoke(prompt)
    return {"final_report": result.content}

# Build the graph
workflow = StateGraph(AgentState)

workflow.add_node("architecture", review_architecture)
workflow.add_node("performance", review_performance)
workflow.add_node("security", review_security)
workflow.add_node("report", generate_report)

workflow.set_entry_point("architecture")
workflow.add_edge("architecture", "performance")
workflow.add_edge("performance", "security")
workflow.add_edge("security", "report")
workflow.add_edge("report", END)

app = workflow.compile()

# Run
result = app.invoke({"code": "def process(data): return data * 2"})
print(result["final_report"])
```text

## Streaming Responses

### OpenAI SDK Streaming

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

# Stream the response
stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Write a long explanation..."}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="", flush=True)
```text

### LangChain Streaming

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed",
    streaming=True
)

for chunk in llm.stream([HumanMessage(content="Explain recursion...")]):
    print(chunk.content, end="", flush=True)
```text

## Async Support

```python
import asyncio
from openai import AsyncOpenAI

async def main():
    client = AsyncOpenAI(
        base_url="http://localhost:8080/openai/v1",
        api_key="not-needed"
    )

    # Async completion
    response = await client.chat.completions.create(
        model="claude",
        messages=[{"role": "user", "content": "Hello!"}]
    )
    print(response.choices[0].message.content)

    # Async streaming
    stream = await client.chat.completions.create(
        model="claude",
        messages=[{"role": "user", "content": "Tell me a story..."}],
        stream=True
    )

    async for chunk in stream:
        if chunk.choices[0].delta.content:
            print(chunk.choices[0].delta.content, end="")

asyncio.run(main())
```text

## Error Handling

```python
from openai import OpenAI, APIError, APIConnectionError

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

try:
    response = client.chat.completions.create(
        model="claude",
        messages=[{"role": "user", "content": "Hello"}]
    )
except APIConnectionError:
    print("Could not connect to clinvk server. Is it running?")
except APIError as e:
    print(f"API error: {e.message}")
```text

## Best Practices

### 1. Connection Pooling

```python
import httpx

# Reuse client for multiple requests
client = httpx.Client(
    base_url="http://localhost:8080",
    timeout=120
)

# Use in your application
response = client.post("/api/v1/prompt", json={...})
```text

### 2. Timeout Configuration

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed",
    request_timeout=120  # 2 minutes for long tasks
)
```text

### 3. Backend Selection Strategy

| Task | Recommended Backend | Reason |
|------|--------------------|----|
| Complex reasoning | `claude` | Strong analytical capabilities |
| Code generation | `codex` | Optimized for code |
| Data analysis | `gemini` | Good with structured data |
| General tasks | `claude` | Versatile default |

## Next Steps

- [CI/CD Integration](ci-cd.md) - Automate with pipelines
- [Client Libraries](client-libraries.md) - Other language bindings
- [REST API Reference](../reference/api/rest-api.md) - Complete API docs
