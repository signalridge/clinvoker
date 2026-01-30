# Anthropic‑compatible API

Base path: `/anthropic/v1`

## Supported endpoint

- `POST /anthropic/v1/messages`

## Routing behavior

`model` is used to select backend and passed through as the backend model name.

- If `model` equals `claude`, `codex`, or `gemini`, that backend is used.
- Else if it contains `claude`, Claude backend is used.
- Otherwise defaults to Claude.

For explicit backend selection with separate model, use `/api/v1/prompt`.

## Messages

### Request (subset)

```json
{
  "model": "claude-opus-4-5-20251101",
  "max_tokens": 1024,
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "system": "You are helpful",
  "stream": false
}
```

### Behavior

- **max_tokens is required** and must be > 0.
- Only text messages are supported.
- Messages are concatenated to a single prompt.

### Streaming

`stream=true` returns Anthropic‑style SSE events (`message_start`, `content_block_delta`, `message_delta`, ...).

## Stateless behavior

Anthropic‑compatible requests are **stateless**. Sessions are not persisted.

## Limitations

- No image or tool content blocks.
- Some Anthropic parameters are accepted but ignored by backends.
