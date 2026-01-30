# Client Libraries

Use existing SDKs without rewriting clients.

## Compatibility matrix

| SDK / Client | Endpoint | Notes |
|-------------|----------|------|
| OpenAI SDK (Python/TS/Go/etc.) | `/openai/v1/*` | Stateless, routes by `model` |
| Anthropic SDK | `/anthropic/v1/*` | Stateless, routes by `model` |
| Any HTTP client | `/api/v1/*` | Full control over backend & model |

## Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
)

resp = client.chat.completions.create(
    model="claude-opus-4-5-20251101",
    messages=[{"role": "user", "content": "Explain this function"}],
)
print(resp.choices[0].message.content)
```

### Streaming

```python
stream = client.chat.completions.create(
    model="claude-opus-4-5-20251101",
    messages=[{"role": "user", "content": "Write a short note"}],
    stream=True,
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

## Python (Anthropic SDK)

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed",
)

resp = client.messages.create(
    model="claude-opus-4-5-20251101",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Summarize this"}],
)
print(resp.content[0].text)
```

## Direct REST (custom API)

```python
import httpx

resp = httpx.post(
    "http://localhost:8080/api/v1/prompt",
    json={
        "backend": "codex",
        "model": "o3",
        "prompt": "Refactor this",
    },
)
print(resp.json()["output"])
```

## Routing note (OpenAI/Anthropic endpoints)

The server uses `model` to choose a backend **and** forwards it as the backend model name. Use a model string that is valid for the target backend (or use `/api/v1/prompt` for explicit control).
