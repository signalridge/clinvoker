---
title: API Design
description: REST API architecture, SDK compatibility layers, and request/response transformation in clinvoker.
---

# API Design

This document explains clinvoker's API architecture, including the REST API design principles, OpenAI and Anthropic compatibility layers, endpoint routing, middleware integration, and request/response transformation.

## API Architecture Overview

clinvoker exposes three API surfaces:

```mermaid
flowchart TB
    subgraph API_Layers["API Layers"]
        CUSTOM[Native API
        /api/v1/*]
        OPENAI[OpenAI Compatible
        /openai/v1/*]
        ANTH[Anthropic Compatible
        /anthropic/v1/*]
    end

    subgraph Core["Core Services"]
        EXEC[Executor]
        SESSION[Session Manager]
        BACKEND[Backend Registry]
    end

    CUSTOM --> EXEC
    OPENAI --> EXEC
    ANTH --> EXEC
    EXEC --> SESSION
    EXEC --> BACKEND
```text

## REST API Design Principles

### Resource-Oriented Design

The native API follows REST principles with resource-oriented URLs:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| POST | `/api/v1/prompt` | Submit a prompt |
| GET | `/api/v1/sessions` | List sessions |
| GET | `/api/v1/sessions/{id}` | Get session details |
| POST | `/api/v1/sessions/{id}/resume` | Resume a session |
| DELETE | `/api/v1/sessions/{id}` | Delete a session |

### HTTP Status Codes

| Status | Meaning |
|--------|---------|
| 200 OK | Successful GET/PUT/DELETE |
| 201 Created | Resource created |
| 400 Bad Request | Invalid request body/params |
| 401 Unauthorized | Missing/invalid API key |
| 429 Too Many Requests | Rate limit exceeded |
| 500 Internal Server Error | Server error |

### Response Format

All responses follow a consistent envelope:

```json
{
  "data": { ... },
  "meta": {
    "request_id": "req-abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```text

## OpenAI Compatibility Layer

The OpenAI-compatible API (`/openai/v1/*`) enables drop-in replacement for OpenAI SDK clients.

### Endpoint Mapping

| OpenAI Endpoint | clinvoker Handler |
|-----------------|-------------------|
| `POST /v1/chat/completions` | `POST /openai/v1/chat/completions` |
| `GET /v1/models` | `GET /openai/v1/models` |
| `GET /v1/models/{model}` | `GET /openai/v1/models/{model}` |

### Request Transformation

```go
// OpenAI request format
{
  "model": "gpt-4",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "stream": false
}

// Transformed to clinvoker internal format
{
  "backend": "claude",
  "prompt": "You are a helpful assistant.\n\nHello!",
  "options": {
    "model": "sonnet"
  }
}
```text

### Response Transformation

```go
// clinvoker internal response
{
  "content": "Hello! How can I help you today?",
  "session_id": "sess-abc123",
  "usage": {
    "input_tokens": 25,
    "output_tokens": 10
  }
}

// OpenAI response format
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

### Streaming Support

OpenAI-compatible streaming uses Server-Sent Events (SSE):

```text
data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"!"}}]}

data: {"id":"chatcmpl-123","choices":[{"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```text

## Anthropic Compatibility Layer

The Anthropic-compatible API (`/anthropic/v1/*`) enables drop-in replacement for Anthropic SDK clients.

### Endpoint Mapping

| Anthropic Endpoint | clinvoker Handler |
|--------------------|-------------------|
| `POST /v1/messages` | `POST /anthropic/v1/messages` |
| `GET /v1/models` | `GET /anthropic/v1/models` |

### Request Transformation

```go
// Anthropic request format
{
  "model": "claude-3-sonnet-20240229",
  "max_tokens": 1024,
  "messages": [
    {"role": "user", "content": "Hello, Claude!"}
  ]
}

// Transformed to clinvoker internal format
{
  "backend": "claude",
  "prompt": "Hello, Claude!",
  "options": {
    "model": "sonnet"
  }
}
```text

### Response Transformation

```go
// Anthropic response format
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

## Endpoint Routing Architecture

### Route Registration

Routes are registered in `internal/server/routes.go`:

```go
func (s *Server) RegisterRoutes() {
    // Register custom RESTful API handlers
    customHandlers := handlers.NewCustomHandlersWithHealthInfo(s.executor, healthInfo)
    customHandlers.Register(s.api)

    // Register OpenAI-compatible API handlers
    openaiHandlers := handlers.NewOpenAIHandlers(service.NewStatelessRunner(s.logger), s.logger)
    openaiHandlers.Register(s.api)

    // Register Anthropic-compatible API handlers
    anthropicHandlers := handlers.NewAnthropicHandlers(service.NewStatelessRunner(s.logger), s.logger)
    anthropicHandlers.Register(s.api)
}
```text

### Huma Integration

clinvoker uses Huma for OpenAPI generation and request/response validation:

```go
huma.Register(s.api, huma.Operation{
    OperationID: "create-chat-completion",
    Method:      http.MethodPost,
    Path:        "/openai/v1/chat/completions",
    Summary:     "Create chat completion",
    Description: "Creates a completion for the chat message",
    Tags:        []string{"OpenAI"},
}, func(ctx context.Context, input *ChatCompletionRequest) (*ChatCompletionResponse, error) {
    // Handler implementation
})
```bash

## Middleware Integration

### Middleware Stack

The middleware stack is configured in `internal/server/server.go:58-131`:

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

### Request ID Middleware

Assigns a unique request ID for tracing:

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

### Rate Limiting Middleware

Implements token bucket rate limiting:

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

### Authentication Middleware

Validates API keys from multiple sources:

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
    // Check Authorization header
    if auth := r.Header.Get("Authorization"); auth != "" {
        if strings.HasPrefix(auth, "Bearer ") {
            return strings.TrimPrefix(auth, "Bearer ")
        }
    }
    // Check X-Api-Key header
    if key := r.Header.Get("X-Api-Key"); key != "" {
        return key
    }
    return ""
}
```bash

## Authentication Design

### API Key Sources

API keys can be provided via:

1. **HTTP Header**: `Authorization: Bearer <key>` or `X-Api-Key: <key>`
2. **Environment Variable**: `CLINVK_API_KEY`
3. **gopass**: Secure password store integration
4. **Config File**: `~/.clinvk/config.yaml`

### Key Validation

```go
func (s *Server) validateAPIKey(key string) bool {
    // Check against configured keys
    for _, validKey := range s.config.APIKeys {
        if subtle.ConstantTimeCompare([]byte(key), []byte(validKey)) == 1 {
            return true
        }
    }
    return false
}
```yaml

Note: `subtle.ConstantTimeCompare` prevents timing attacks.

## Error Handling Strategy

### Error Response Format

All errors follow a consistent format:

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

### Error Types

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_request` | 400 | Request validation failed |
| `authentication_error` | 401 | Invalid or missing API key |
| `rate_limit_exceeded` | 429 | Too many requests |
| `backend_unavailable` | 503 | Backend not available |
| `internal_error` | 500 | Internal server error |

### Error Handling Middleware

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

## Versioning Approach

### URL Versioning

API versions are included in the URL path:

- `/api/v1/*` - Native API v1
- `/openai/v1/*` - OpenAI-compatible v1
- `/anthropic/v1/*` - Anthropic-compatible v1

### Version Negotiation

Future versions may support header-based negotiation:

```text
Accept-Version: v2
```text

### Deprecation Strategy

1. **Announcement**: 6 months notice before deprecation
2. **Sunset Header**: Include `Sunset` header in responses
3. **Grace Period**: Support old version for 3 months after new version release

## Request/Response Transformation

### Unified Options Mapping

```mermaid
flowchart TB
    subgraph Input["Input Request"]
        OPENAI_REQ[OpenAI Format]
        ANTH_REQ[Anthropic Format]
        NATIVE_REQ[Native Format]
    end

    subgraph Transform["Transformation Layer"]
        MAP[Options Mapper]
    end

    subgraph Internal["Internal Format"]
        UNIFIED[UnifiedOptions]
    end

    OPENAI_REQ --> MAP
    ANTH_REQ --> MAP
    NATIVE_REQ --> MAP
    MAP --> UNIFIED
```text

### Streaming Transformation

For streaming responses, data is transformed chunk by chunk:

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

## Related Documentation

- [Architecture Overview](architecture.md) - High-level system architecture
- [Backend System](backend-system.md) - Backend abstraction layer
- [Session System](session-system.md) - Session persistence mechanisms
- [Reference: REST API](../reference/api/rest.md) - Complete API reference
