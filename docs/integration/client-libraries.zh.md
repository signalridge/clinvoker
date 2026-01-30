# 客户端 SDK

复用现有 SDK，无需改写客户端。

## 兼容矩阵

| SDK / 客户端 | 端点 | 说明 |
|-------------|------|------|
| OpenAI SDK（Python/TS/Go 等） | `/openai/v1/*` | 无状态，按 `model` 路由 |
| Anthropic SDK | `/anthropic/v1/*` | 无状态，按 `model` 路由 |
| 任意 HTTP 客户端 | `/api/v1/*` | 可显式控制 backend 与 model |

## Python（OpenAI SDK）

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
)

resp = client.chat.completions.create(
    model="claude-opus-4-5-20251101",
    messages=[{"role": "user", "content": "解释这个函数"}],
)
print(resp.choices[0].message.content)
```

### 流式

```python
stream = client.chat.completions.create(
    model="claude-opus-4-5-20251101",
    messages=[{"role": "user", "content": "写一段短文"}],
    stream=True,
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

## Python（Anthropic SDK）

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed",
)

resp = client.messages.create(
    model="claude-opus-4-5-20251101",
    max_tokens=1024,
    messages=[{"role": "user", "content": "总结这段内容"}],
)
print(resp.content[0].text)
```

## 自定义 REST（/api/v1）

```python
import httpx

resp = httpx.post(
    "http://localhost:8080/api/v1/prompt",
    json={
        "backend": "codex",
        "model": "o3",
        "prompt": "重构这段代码",
    },
)
print(resp.json()["output"])
```

## 路由说明（OpenAI/Anthropic）

OpenAI/Anthropic 端点会用 `model` 选择后端，并把该值作为后端模型名。请使用对目标后端有效的模型名，或改用 `/api/v1/prompt` 进行显式控制。
