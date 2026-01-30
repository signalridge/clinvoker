# LangChain / LangGraph

将 clinvk 作为模型端点接入 LangChain。

## OpenAI 兼容模式

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed",
    model="claude-opus-4-5-20251101",
)

print(llm.invoke("解释这个函数"))
```

## 说明

- `model` 用于路由并作为后端模型名传递。
- 需要显式 backend/model 时，请使用自定义 REST API（`/api/v1/prompt`）。
