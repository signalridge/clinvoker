# OpenAI 兼容 API

Base path：`/openai/v1`

## 支持的端点

- `GET /openai/v1/models`
- `POST /openai/v1/chat/completions`

## 路由规则

`model` 既用于**选择后端**，也会作为后端模型名传递。

路由规则：

1. `model` 等于 `claude` / `codex` / `gemini` → 对应后端
2. 含 `claude` → Claude 后端
3. 含 `gpt` → Codex 后端
4. 含 `gemini` → Gemini 后端
5. 其它 → Claude 后端

**重要：** 请使用对目标后端有效的模型名。若需显式设置 backend + model，建议使用自定义 REST API（`/api/v1/prompt`）。

## Chat completions

### 请求（子集）

```json
{
  "model": "claude-opus-4-5-20251101",
  "messages": [
    {"role": "system", "content": "你是助手"},
    {"role": "user", "content": "Hello"}
  ],
  "stream": false
}
```

### 行为

- 所有 **user** 消息会被拼接为一个 prompt。
- **system** 消息会作为 `system_prompt`（是否生效取决于后端）。
- assistant 消息以上下文提示形式加入。

### 流式输出

`stream=true` 时返回符合 OpenAI chat streaming 的 SSE。

## Models 列表

`GET /openai/v1/models` 返回的是**后端名称**（`claude` / `codex` / `gemini`），而不是完整模型列表。

## 无状态

OpenAI 兼容端点为**无状态**，不会持久化会话。

## 限制

- 仅实现 `chat.completions`。
- 不支持工具调用、函数调用与图片输入。
- 一些 OpenAI 参数会被忽略。
