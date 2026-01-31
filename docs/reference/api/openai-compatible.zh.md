# OpenAI 兼容 API

使用 OpenAI SDK 访问 clinvk。

## Base URL

```text
http://localhost:8080/openai/v1
```

## 认证

若配置了 API Key，请在请求中携带：

- `Authorization: Bearer <key>`
- 或 `X-Api-Key: <key>`

未配置 Key 时默认放行。

## 模型映射

`model` 字段决定使用哪个后端：

- 精确后端名：`claude` / `codex` / `gemini`
- 包含 `claude` → Claude
- 包含 `gpt` → Codex
- 包含 `gemini` → Gemini
- 其他默认 Claude

**建议：** 直接使用后端名（`codex` / `claude` / `gemini`）。

## 常见差异

- 仅支持 Chat Completions 与模型列表
- 错误为 RFC 7807（非 OpenAI 错误格式）
- 请求无状态（如需会话，请用自定义 REST API）
