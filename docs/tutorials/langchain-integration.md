---
title: LangChain Integration Tutorial
description: Connect clinvoker to LangChain and LangGraph for complex agent workflows.
---

# Tutorial: LangChain Integration

Connect clinvoker to LangChain and LangGraph for building complex agent workflows with multiple AI backends.

## What You'll Build

A LangChain application that:

1. Uses clinvoker's OpenAI-compatible API
2. Routes to different backends based on task type
3. Implements a multi-agent workflow

## Prerequisites

- Python 3.9+
- LangChain installed
- clinvoker server running

## Step 1: Install Dependencies

```bash
pip install langchain langchain-openai
```

## Step 2: Configure LangChain

Create `clinvk_client.py`:

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage

# Configure LangChain to use clinvoker
llm = ChatOpenAI(
    model_name="claude",  # Routes to Claude backend
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="your-api-key",
    temperature=0.7,
)

# Use the LLM
response = llm.invoke([HumanMessage(content="Explain microservices")])
print(response.content)
```

## Step 3: Multi-Backend Chain

Create a chain that routes to different backends:

```python
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser
from langchain_core.runnables import RunnableParallel

# Define LLMs for different backends
claude = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="your-api-key",
)

codex = ChatOpenAI(
    model_name="codex",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="your-api-key",
)

gemini = ChatOpenAI(
    model_name="gemini",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="your-api-key",
)

# Create parallel review chain
review_chain = RunnableParallel(
    architecture=claude.invoke,
    implementation=codex.invoke,
    security=gemini.invoke,
)

# Use the chain
from langchain_core.messages import HumanMessage

code = """
def authenticate(user, password):
    query = f"SELECT * FROM users WHERE user='{user}' AND pass='{password}'"
    return db.execute(query)
"""

results = review_chain.invoke({
    "architecture": [HumanMessage(content=f"Review architecture:\n{code}")],
    "implementation": [HumanMessage(content=f"Review implementation:\n{code}")],
    "security": [HumanMessage(content=f"Security audit:\n{code}")],
})

print("Architecture:", results["architecture"].content)
print("Implementation:", results["implementation"].content)
print("Security:", results["security"].content)
```

## Step 4: LangGraph Workflow

Create a multi-agent workflow with LangGraph:

```python
from typing import TypedDict, Annotated
from langgraph.graph import StateGraph, END
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage

# State definition
class AgentState(TypedDict):
    code: str
    architecture_review: str
    implementation: str
    security_review: str
    final_output: str

# Initialize LLMs
claude = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="your-api-key",
)

codex = ChatOpenAI(
    model_name="codex",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="your-api-key",
)

gemini = ChatOpenAI(
    model_name="gemini",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="your-api-key",
)

# Node functions
def architect_review(state: AgentState):
    prompt = f"Review this code architecture:\n{state['code']}"
    response = claude.invoke([HumanMessage(content=prompt)])
    return {"architecture_review": response.content}

def implement(state: AgentState):
    prompt = f"Implement improved version based on:\n{state['architecture_review']}\n\nOriginal code:\n{state['code']}"
    response = codex.invoke([HumanMessage(content=prompt)])
    return {"implementation": response.content}

def security_check(state: AgentState):
    prompt = f"Security audit:\n{state['implementation']}"
    response = gemini.invoke([HumanMessage(content=prompt)])
    return {"security_review": response.content}

def finalize(state: AgentState):
    prompt = f"Synthesize final solution:\nArchitecture: {state['architecture_review']}\nImplementation: {state['implementation']}\nSecurity: {state['security_review']}"
    response = claude.invoke([HumanMessage(content=prompt)])
    return {"final_output": response.content}

# Build the graph
workflow = StateGraph(AgentState)

workflow.add_node("architect", architect_review)
workflow.add_node("implement", implement)
workflow.add_node("security", security_check)
workflow.add_node("finalize", finalize)

workflow.set_entry_point("architect")
workflow.add_edge("architect", "implement")
workflow.add_edge("implement", "security")
workflow.add_edge("security", "finalize")
workflow.add_edge("finalize", END)

# Compile and run
app = workflow.compile()

# Execute
result = app.invoke({
    "code": """
def process_payment(card_number, amount):
    # Process payment
    return True
"""
})

print(result["final_output"])
```

## Step 5: Custom Tool Integration

Create a custom LangChain tool for clinvoker:

```python
from langchain_core.tools import BaseTool
from pydantic import BaseModel, Field
import requests

class ClinvokerInput(BaseModel):
    backend: str = Field(description="Backend to use (claude, codex, gemini)")
    prompt: str = Field(description="Prompt to send")

class ClinvokerTool(BaseTool):
    name = "clinvoker"
    description = "Execute a prompt on a specific AI backend via clinvoker"
    args_schema = ClinvokerInput

    def _run(self, backend: str, prompt: str) -> str:
        response = requests.post(
            "http://localhost:8080/api/v1/prompt",
            headers={"Authorization": "Bearer your-api-key"},
            json={"backend": backend, "prompt": prompt}
        )
        return response.json()["output"]

# Use the tool
from langchain.agents import AgentExecutor, create_openai_functions_agent
from langchain_core.prompts import ChatPromptTemplate

tools = [ClinvokerTool()]

prompt = ChatPromptTemplate.from_messages([
    ("system", "You are an AI assistant with access to multiple backends."),
    ("user", "{input}"),
])

agent = create_openai_functions_agent(claude, tools, prompt)
agent_executor = AgentExecutor(agent=agent, tools=tools)

result = agent_executor.invoke({
    "input": "Analyze this code for security issues using Gemini"
})
print(result["output"])
```

## Verification

Test your LangChain integration:

```python
# Test basic connection
response = llm.invoke([HumanMessage(content="Hello")])
assert response.content is not None

# Test multi-backend
results = review_chain.invoke({...})
assert "architecture" in results
assert "implementation" in results
assert "security" in results

print("All tests passed!")
```

## Next Steps

- Explore [AI Team Collaboration](../use-cases/ai-team-collaboration.md) for workflow patterns
- Learn about [API Gateway](../use-cases/api-gateway-pattern.md) for production deployment
- See [Building AI Skills](building-ai-skills.md) for Claude Code integration
