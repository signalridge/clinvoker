# Anthropic 兼容 API

使用 Anthropic SDK 访问 clinvk。

## Base URL

```text
http://localhost:8080/anthropic/v1
```

## 认证

若配置了 API Key，请在请求中携带：

- `Authorization: Bearer <key>`
- 或 `X-Api-Key: <key>`

未配置 Key 时默认放行。

## 模型映射

- 精确后端名：`claude` / `codex` / `gemini`
- 包含 `claude` → Claude
- 其他默认 Claude

**建议：** 如需 Codex/Gemini，请显式使用 `codex` 或 `gemini`。

## 常见差异

- 仅支持 Messages API
- 错误为 RFC 7807
- 请求无状态（如需会话，请用自定义 REST API）
