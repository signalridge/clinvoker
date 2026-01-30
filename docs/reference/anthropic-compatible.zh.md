# Anthropic 兼容 API

Base path：`/anthropic/v1`

## 支持的端点

- `POST /anthropic/v1/messages`

## 路由规则

`model` 用于选择后端，并作为模型名传递。

- 若 `model` 等于 `claude` / `codex` / `gemini`，使用对应后端
- 若包含 `claude`，使用 Claude
- 其它情况默认使用 Claude

如需显式指定后端并设置模型，请使用 `/api/v1/prompt`。

## Messages

### 请求（子集）

```json
{
  "model": "claude-opus-4-5-20251101",
  "max_tokens": 1024,
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "system": "你是助手",
  "stream": false
}
```

### 行为

- **max_tokens 必填**，且必须 > 0。
- 仅支持文本消息。
- 消息会被拼接为单一 prompt。

### 流式输出

`stream=true` 返回 Anthropic 风格 SSE（`message_start` / `content_block_delta` / `message_delta` 等）。

## 无状态

Anthropic 兼容端点为**无状态**，不会持久化会话。

## 限制

- 不支持图片或工具内容块。
- 一些 Anthropic 参数会被忽略。
