# OpenAI‑compatible API

Base path: `/openai/v1`

## Supported endpoints

- `GET /openai/v1/models`
- `POST /openai/v1/chat/completions`

## Routing behavior

`model` is used to **select a backend** and is also forwarded as the backend model name.

Routing rules:

1. If `model` equals `claude`, `codex`, or `gemini`, use that backend.
2. Else if it contains `claude` → Claude backend.
3. Else if it contains `gpt` → Codex backend.
4. Else if it contains `gemini` → Gemini backend.
5. Otherwise → Claude backend.

**Important:** choose a model string that is valid for the target backend. If you want explicit backend+model control, use the custom REST API (`/api/v1/prompt`).

## Chat completions

### Request (subset)

```json
{
  "model": "claude-opus-4-5-20251101",
  "messages": [
    {"role": "system", "content": "You are helpful"},
    {"role": "user", "content": "Hello"}
  ],
  "stream": false
}
```

### Behavior

- All **user** messages are concatenated into a single prompt.
- The **system** message is passed as `system_prompt` (backend‑dependent support).
- Assistant messages are added as context markers.

### Streaming

If `stream=true`, the server emits SSE chunks compatible with OpenAI chat streaming.

## Models list

`GET /openai/v1/models` returns **backend names** (`claude`, `codex`, `gemini`) as models.

## Stateless behavior

OpenAI‑compatible requests are **stateless**. Sessions are not persisted.

## Limitations

- Only `chat.completions` is implemented.
- Tool/function calling and images are not supported.
- Some OpenAI parameters are accepted but ignored by backends.
