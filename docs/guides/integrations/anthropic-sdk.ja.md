---
title: Anthropic SDK 連携
description: Anthropic SDK を clinvoker バックエンドとして利用します。
---

# Anthropic SDK 連携

公式の Anthropic SDK を、clinvoker をバックエンドとして利用します。

## クイックスタート

### Python

```python
from anthropic import Anthropic

# Configure client to use clinvoker
client = Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="your-clinvoker-api-key"
)

# Use like normal Anthropic client
response = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "Hello, world!"}
    ]
)

print(response.content[0].text)
```

### TypeScript/JavaScript

```typescript
import Anthropic from '@anthropic-ai/sdk';

const client = new Anthropic({
    baseURL: 'http://localhost:8080/anthropic/v1',
    apiKey: 'your-clinvoker-api-key',
});

const response = await client.messages.create({
    model: 'claude',
    max_tokens: 1024,
    messages: [{ role: 'user', content: 'Hello!' }],
});

console.log(response.content[0].text);
```

## 対応モデル

Anthropic SDK は、clinvoker の Claude バックエンドで動作します。

| モデル | バックエンド |
|-------|---------|
| `claude` | Claude Code |
| `claude-sonnet-4` | Claude Code |

## Streaming

```python
# ストリーミング応答
with client.messages.stream(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Write a poem"}]
) as stream:
    for text in stream.text_stream:
        print(text, end="", flush=True)
```

## System Prompts

```python
response = client.messages.create(
    model="claude",
    max_tokens=1024,
    system="You are a helpful coding assistant.",
    messages=[{"role": "user", "content": "Review this code"}]
)
```

## ベストプラクティス

1. **適切なタイムアウトを設定する**: Anthropic SDK のデフォルトは短すぎる場合があります
2. **ストリーミングを扱う**: 長い応答にはストリーミングを使うと便利です
3. **トークン数を確認する**: `response.usage` で利用状況を監視します

## トラブルシューティング

### Connection Issues

clinvoker サーバーが想定のポートで起動していることを確認してください。

```bash
curl http://localhost:8080/health
```

### Version Compatibility

互換性のある Anthropic SDK バージョン（1.0+）を使用してください。

## 関連ドキュメント

- [Anthropic API リファレンス](../../reference/api/anthropic-compat.md)
- [OpenAI SDK 連携](openai-sdk.md)
