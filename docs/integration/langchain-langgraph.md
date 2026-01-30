# LangChain / LangGraph

Use clinvk as a drop‑in model endpoint.

## OpenAI‑compatible mode

Point LangChain to the OpenAI‑compatible endpoint:

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
    model="claude-opus-4-5-20251101",
)

print(llm.invoke("Explain this function"))
```

## Notes

- `model` is used to route to a backend and forwarded as the backend model name.
- For explicit backend/model control, use the custom REST API (`/api/v1/prompt`).
