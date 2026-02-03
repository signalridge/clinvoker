---
title: OpenAI SDK 連携
description: OpenAI SDK を clinvoker バックエンドとして利用します。
---

# OpenAI SDK 連携

公式の OpenAI SDK を、clinvoker をバックエンドとして利用します。既存の OpenAI ベースのアプリケーションに対して、最小の変更で統合できます。

## クイックスタート

### Python

```python
from openai import OpenAI

# Configure client to use clinvoker
client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-clinvoker-api-key"  # or any non-empty string
)

# Use like normal OpenAI client
response = client.chat.completions.create(
    model="claude",  # Maps to Claude backend
    messages=[
        {"role": "user", "content": "Hello, world!"}
    ]
)

print(response.choices[0].message.content)
```

### TypeScript/JavaScript

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
    baseURL: 'http://localhost:8080/openai/v1',
    apiKey: 'your-clinvoker-api-key',
});

const response = await client.chat.completions.create({
    model: 'claude',
    messages: [{ role: 'user', content: 'Hello!' }],
});

console.log(response.choices[0].message.content);
```

### Go

```go
package main

import (
    "context"
    "fmt"
    "github.com/sashabaranov/go-openai"
)

func main() {
    client := openai.NewClientWithConfig(openai.ClientConfig{
        BaseURL: "http://localhost:8080/openai/v1",
        APIKey:  "your-clinvoker-api-key",
    })

    resp, err := client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: "claude",
            Messages: []openai.ChatCompletionMessage{
                {Role: openai.ChatMessageRoleUser, Content: "Hello!"},
            },
        },
    )

    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## モデルのマッピング

clinvoker のバックエンドは、OpenAI の「モデル」として公開されます。

| clinvoker バックエンド | OpenAI のモデル ID |
|-------------------|-----------------|
| Claude Code | `claude` |
| Codex CLI | `codex` |
| Gemini CLI | `gemini` |

バックエンド固有のモデル名を指定することもできます。

```python
# Use specific Claude model
response = client.chat.completions.create(
    model="claude-sonnet-4",
    messages=[...]
)
```

## Streaming Responses

```python
# Stream responses in real-time
stream = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Write a story"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

## Error Handling

```python
from openai import OpenAIError, RateLimitError

try:
    response = client.chat.completions.create(...)
except RateLimitError:
    print("Rate limit exceeded")
except OpenAIError as e:
    print(f"Error: {e}")
```

## Advanced Configuration

### Custom Headers

```python
client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-key",
    default_headers={
        "X-Custom-Header": "value"
    }
)
```

### Timeout Configuration

```python
import httpx

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="your-key",
    timeout=httpx.Timeout(300.0)  # 5 minutes
)
```

## Best Practices

1. **コネクションプーリング**: クライアントインスタンスを再利用する
2. **リトライ**: 指数バックオフを実装する
3. **タイムアウト**: 長時間タスクに合わせて適切なタイムアウトを設定する
4. **API キー管理**: キーは安全に保管する

## Troubleshooting

### Connection Refused

clinvoker サーバーが起動していることを確認してください。

```bash
clinvk serve --port 8080
```

### Model Not Found

利用可能なモデルを確認します。

```bash
curl http://localhost:8080/openai/v1/models
```

### Authentication Errors

API キーが clinvoker 側で正しく設定されていることを確認してください。

## Related Documentation

- [OpenAI API リファレンス](../../reference/api/openai-compat.md)
- [API ゲートウェイパターン](../http-server.md)
- [LangChain 連携](langchain-langgraph.md)
