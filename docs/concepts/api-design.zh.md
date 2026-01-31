---
title: API 设计
description: clinvoker 的 REST API 架构、SDK 兼容层和请求/响应转换。
---

# API 设计

本文档解释 clinvoker 的 API 架构，包括 REST API 设计原则、OpenAI 和 Anthropic 兼容层、端点路由、中间件集成以及请求/响应转换。

## API 架构概述

clinvoker 暴露三个 API 层面：

```mermaid
flowchart TB
    subgraph API层["API 层"]
        CUSTOM[原生 API
        /api/v1/*]
        OPENAI[OpenAI 兼容
        /openai/v1/*]
        ANTH[Anthropic 兼容
        /anthropic/v1/*]
    end

    subgraph 核心["核心服务"]
        EXEC[执行器]
        SESSION[会话管理器]
        BACKEND[后端注册表]
    end

    CUSTOM --> EXEC
    OPENAI --> EXEC
    ANTH --> EXEC
    EXEC --> SESSION
    EXEC --> BACKEND
```text

## REST API 设计原则

### 面向资源的设计

原生 API 遵循 REST 原则，使用面向资源的 URL：

| 方法 | 端点 | 描述 |
|--------|----------|-------------|
| GET | `/api/v1/health` | 健康检查 |
| POST | `/api/v1/prompt` | 提交提示 |
| GET | `/api/v1/sessions` | 列出会话 |
| GET | `/api/v1/sessions/{id}` | 获取会话详情 |
| POST | `/api/v1/sessions/{id}/resume` | 恢复会话 |
| DELETE | `/api/v1/sessions/{id}` | 删除会话 |

### HTTP 状态码

| 状态 | 含义 |
|--------|---------|
| 200 OK | 成功的 GET/PUT/DELETE |
| 201 Created | 资源已创建 |
| 400 Bad Request | 无效的请求体/参数 |
| 401 Unauthorized | 缺少/无效的 API 密钥 |
| 429 Too Many Requests | 超出速率限制 |
| 500 Internal Server Error | 服务器错误 |

### 响应格式

所有响应遵循一致的封装格式：

```json
{
  "data": { ... },
  "meta": {
    "request_id": "req-abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```text

## OpenAI 兼容层

OpenAI 兼容 API（`/openai/v1/*`）支持 OpenAI SDK 客户端的即插即用替换。

### 端点映射

| OpenAI 端点 | clinvoker 处理器 |
|-----------------|-------------------|
| `POST /v1/chat/completions` | `POST /openai/v1/chat/completions` |
| `GET /v1/models` | `GET /openai/v1/models` |
| `GET /v1/models/{model}` | `GET /openai/v1/models/{model}` |

### 请求转换

```go
// OpenAI 请求格式
{
  "model": "gpt-4",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "stream": false
}

// 转换为 clinvoker 内部格式
{
  "backend": "claude",
  "prompt": "You are a helpful assistant.\n\nHello!",
  "options": {
    "model": "sonnet"
  }
}
```text

### 响应转换

```go
// clinvoker 内部响应
{
  "content": "Hello! How can I help you today?",
  "session_id": "sess-abc123",
  "usage": {
    "input_tokens": 25,
    "output_tokens": 10
  }
}

// OpenAI 响应格式
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1705317600,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 10,
    "total_tokens": 35
  }
}
```text

### 流式支持

OpenAI 兼容的流式使用服务器发送事件 (SSE)：

```text
data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"!"}}]}

data: {"id":"chatcmpl-123","choices":[{"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```text

## Anthropic 兼容层

Anthropic 兼容 API（`/anthropic/v1/*`）支持 Anthropic SDK 客户端的即插即用替换。

### 端点映射

| Anthropic 端点 | clinvoker 处理器 |
|--------------------|-------------------|
| `POST /v1/messages` | `POST /anthropic/v1/messages` |
| `GET /v1/models` | `GET /anthropic/v1/models` |

### 请求转换

```go
// Anthropic 请求格式
{
  "model": "claude-3-sonnet-20240229",
  "max_tokens": 1024,
  "messages": [
    {"role": "user", "content": "Hello, Claude!"}
  ]
}

// 转换为 clinvoker 内部格式
{
  "backend": "claude",
  "prompt": "Hello, Claude!",
  "options": {
    "model": "sonnet"
  }
}
```text

### 响应转换

```go
// Anthropic 响应格式
{
  "id": "msg_01XgY...",
  "type": "message",
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Hello! How can I help?"}
  ],
  "model": "claude-3-sonnet-20240229",
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 15,
    "output_tokens": 10
  }
}
```bash

## 端点路由架构

### 路由注册

路由在 `internal/server/routes.go` 中注册：

```go
func (s *Server) RegisterRoutes() {
    // 注册自定义 RESTful API 处理器
    customHandlers := handlers.NewCustomHandlersWithHealthInfo(s.executor, healthInfo)
    customHandlers.Register(s.api)

    // 注册 OpenAI 兼容 API 处理器
    openaiHandlers := handlers.NewOpenAIHandlers(service.NewStatelessRunner(s.logger), s.logger)
    openaiHandlers.Register(s.api)

    // 注册 Anthropic 兼容 API 处理器
    anthropicHandlers := handlers.NewAnthropicHandlers(service.NewStatelessRunner(s.logger), s.logger)
    anthropicHandlers.Register(s.api)
}
```text

### Huma 集成

clinvoker 使用 Huma 进行 OpenAPI 生成和请求/响应验证：

```go
huma.Register(s.api, huma.Operation{
    OperationID: "create-chat-completion",
    Method:      http.MethodPost,
    Path:        "/openai/v1/chat/completions",
    Summary:     "Create chat completion",
    Description: "Creates a completion for the chat message",
    Tags:        []string{"OpenAI"},
}, func(ctx context.Context, input *ChatCompletionRequest) (*ChatCompletionResponse, error) {
    // 处理器实现
})
```bash

## 中间件集成

### 中间件栈

中间件栈在 `internal/server/server.go:58-131` 中配置：

```mermaid
flowchart LR
    REQID[RequestID]
    REALIP[RealIP]
    RECOVER[Recoverer]
    LOGGER[RequestLogger]
    SIZE[RequestSize]
    RATE[RateLimiter]
    AUTH[APIKeyAuth]
    TIMEOUT[Timeout]
    CORS[CORS]

    REQID --> REALIP
    REALIP --> RECOVER
    RECOVER --> LOGGER
    LOGGER --> SIZE
    SIZE --> RATE
    RATE --> AUTH
    AUTH --> TIMEOUT
    TIMEOUT --> CORS
```text

### 请求 ID 中间件

为跟踪分配唯一请求 ID：

```go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := generateRequestID()
        ctx := context.WithValue(r.Context(), requestIDKey, requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```text

### 速率限制中间件

实现令牌桶速率限制：

```go
type RateLimiter struct {
    rps     float64
    burst   int
    clients map[string]*clientLimiter
    mu      sync.RWMutex
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientID := getClientID(r)
            if !rl.allow(clientID) {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```text

### 认证中间件

验证来自多个来源的 API 密钥：

```go
func APIKeyAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := extractAPIKey(r)
        if apiKey == "" || !isValidAPIKey(apiKey) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), apiKeyKey, apiKey)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func extractAPIKey(r *http.Request) string {
    // 检查 Authorization 请求头
    if auth := r.Header.Get("Authorization"); auth != "" {
        if strings.HasPrefix(auth, "Bearer ") {
            return strings.TrimPrefix(auth, "Bearer ")
        }
    }
    // 检查 X-Api-Key 请求头
    if key := r.Header.Get("X-Api-Key"); key != "" {
        return key
    }
    return ""
}
```bash

## 认证设计

### API 密钥来源

API 密钥可以通过以下方式提供：

1. **HTTP 请求头**：`Authorization: Bearer <key>` 或 `X-Api-Key: <key>`
2. **环境变量**：`CLINVK_API_KEY`
3. **gopass**：安全密码存储集成
4. **配置文件**：`~/.clinvk/config.yaml`

### 密钥验证

```go
func (s *Server) validateAPIKey(key string) bool {
    // 检查配置的密钥
    for _, validKey := range s.config.APIKeys {
        if subtle.ConstantTimeCompare([]byte(key), []byte(validKey)) == 1 {
            return true
        }
    }
    return false
}
```text

注意：`subtle.ConstantTimeCompare` 可防止时序攻击。

## 错误处理策略

### 错误响应格式

所有错误遵循一致的格式：

```json
{
  "error": {
    "code": "invalid_request",
    "message": "The request body is invalid",
    "details": {
      "field": "model",
      "issue": "required"
    }
  }
}
```text

### 错误类型

| 代码 | HTTP 状态 | 描述 |
|------|-------------|-------------|
| `invalid_request` | 400 | 请求验证失败 |
| `authentication_error` | 401 | 无效或缺少 API 密钥 |
| `rate_limit_exceeded` | 429 | 请求过多 |
| `backend_unavailable` | 503 | 后端不可用 |
| `internal_error` | 500 | 内部服务器错误 |

### 错误处理中间件

```go
func ErrorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                log.Printf("Panic: %v\n%s", rec, debug.Stack())
                respondWithError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```text

## 版本控制方法

### URL 版本控制

API 版本包含在 URL 路径中：

- `/api/v1/*` - 原生 API v1
- `/openai/v1/*` - OpenAI 兼容 v1
- `/anthropic/v1/*` - Anthropic 兼容 v1

### 版本协商

未来版本可能支持基于请求头的协商：

```text
Accept-Version: v2
```text

### 弃用策略

1. **公告**：弃用前 6 个月通知
2. **Sunset 请求头**：在响应中包含 `Sunset` 请求头
3. **宽限期**：新版本发布后支持旧版本 3 个月

## 请求/响应转换

### 统一选项映射

```mermaid
flowchart TB
    subgraph 输入["输入请求"]
        OPENAI_REQ[OpenAI 格式]
        ANTH_REQ[Anthropic 格式]
        NATIVE_REQ[原生格式]
    end

    subgraph 转换["转换层"]
        MAP[选项映射器]
    end

    subgraph 内部["内部格式"]
        UNIFIED[UnifiedOptions]
    end

    OPENAI_REQ --> MAP
    ANTH_REQ --> MAP
    NATIVE_REQ --> MAP
    MAP --> UNIFIED
```text

### 流式转换

对于流式响应，数据逐块转换：

```go
func (h *OpenAIHandler) streamResponse(ctx context.Context, input *ChatCompletionRequest, w http.ResponseWriter) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }

    for chunk := range h.executor.Stream(ctx, input) {
        openaiChunk := transformToOpenAI(chunk)
        data, _ := json.Marshal(openaiChunk)
        fmt.Fprintf(w, "data: %s\n\n", data)
        flusher.Flush()
    }

    fmt.Fprint(w, "data: [DONE]\n\n")
    flusher.Flush()
}
```text

## 相关文档

- [架构概述](architecture.zh.md) - 高级系统架构
- [后端系统](backend-system.zh.md) - 后端抽象层
- [会话系统](session-system.zh.md) - 会话持久化机制
- [参考：REST API](../reference/api/rest.zh.md) - 完整 API 参考
