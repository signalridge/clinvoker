# API 参考

clinvk HTTP API 的完整参考。

## 概述

clinvoker 为与各种工具和 SDK 集成提供了多个 API 端点。HTTP 服务器暴露三种 API 样式：

| API 样式 | 端点前缀 | 最适合 |
|-----------|-----------------|----------|
| **原生 REST** | `/api/v1/` | 完整 clinvoker 功能 |
| **OpenAI 兼容** | `/openai/v1/` | OpenAI SDK 用户 |
| **Anthropic 兼容** | `/anthropic/v1/` | Anthropic SDK 用户 |

## 快速开始

### 启动服务器

```bash
clinvk serve --port 8080
```

### 测试 API

```bash
curl http://localhost:8080/health
```

### 执行提示词

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello"}'
```

## 选择 API

| 使用场景 | 推荐 API | 文档 |
|----------|-----------------|---------------|
| 使用 OpenAI SDK | OpenAI 兼容 | [openai-compat.md](openai-compat.md) |
| 使用 Anthropic SDK | Anthropic 兼容 | [anthropic-compat.md](anthropic-compat.md) |
| 完整 clinvoker 功能 | 原生 REST | [rest.md](rest.md) |
| 自定义集成 | 原生 REST | [rest.md](rest.md) |
| 会话管理 | 原生 REST | [rest.md](rest.md) |

## 基础 URL

```text
http://localhost:8080
```

## 认证

API Key 认证是可选的。配置后，每个请求都必须包含 API Key。

### 配置

通过以下方式设置 API Key：

- `CLINVK_API_KEYS` 环境变量（逗号分隔）
- 配置中的 `server.api_keys_gopass_path`（gopass 路径）

### 请求头

使用以下头之一包含 API Key：

```bash
# 选项 1：X-Api-Key 头
curl -H "X-Api-Key: your-api-key" http://localhost:8080/api/v1/prompt

# 选项 2：Authorization 头
curl -H "Authorization: Bearer your-api-key" http://localhost:8080/api/v1/prompt
```

如果未配置 Key，则允许无认证请求。

## 响应格式

所有 API 返回具有统一结构的 JSON 响应：

```json
{
  "success": true,
  "data": { ... }
}
```

## 错误处理

错误响应包含错误消息：

```json
{
  "success": false,
  "error": "Backend not available"
}
```

### HTTP 状态码

| 代码 | 说明 |
|------|-------------|
| 200 | 成功 |
| 400 | 错误请求 |
| 401 | 未授权（无效/缺失 API Key） |
| 404 | 未找到 |
| 429 | 限流 |
| 500 | 服务器错误 |

## 可用端点

### 原生 REST API

| 方法 | 端点 | 说明 |
|--------|----------|-------------|
| POST | `/api/v1/prompt` | 执行提示词 |
| POST | `/api/v1/parallel` | 并行执行 |
| POST | `/api/v1/chain` | 链式执行 |
| POST | `/api/v1/compare` | 后端对比 |
| GET | `/api/v1/backends` | 列出后端 |
| GET | `/api/v1/sessions` | 列出会话 |
| GET | `/api/v1/sessions/{id}` | 获取会话 |
| DELETE | `/api/v1/sessions/{id}` | 删除会话 |

### OpenAI 兼容

| 方法 | 端点 | 说明 |
|--------|----------|-------------|
| GET | `/openai/v1/models` | 列出模型 |
| POST | `/openai/v1/chat/completions` | 对话补全 |

### Anthropic 兼容

| 方法 | 端点 | 说明 |
|--------|----------|-------------|
| POST | `/anthropic/v1/messages` | 创建消息 |

### 元端点

| 方法 | 端点 | 说明 |
|--------|----------|-------------|
| GET | `/health` | 健康检查 |
| GET | `/openapi.json` | OpenAPI 规范 |
| GET | `/docs` | API 文档（Huma UI） |
| GET | `/metrics` | Prometheus 指标 |

## SDK 集成示例

### Python 与 OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```

### Python 与 Anthropic SDK

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}]
)

print(message.content[0].text)
```

### JavaScript/TypeScript 与 OpenAI SDK

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/openai/v1',
  apiKey: 'not-needed'
});

const response = await client.chat.completions.create({
  model: 'claude',
  messages: [{ role: 'user', content: 'Hello!' }]
});

console.log(response.choices[0].message.content);
```

## 服务器配置

在 `~/.clinvk/config.yaml` 中配置服务器：

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  rate_limit_enabled: false
  metrics_enabled: false
```

## 安全考虑

- 生产部署使用 API Key
- 生产中限制 CORS 来源
- 配置允许的工作目录前缀
- 生产中使用 HTTPS（通过反向代理）
- 为公共部署启用限流

## 下一步

- [REST API 文档](rest.md) - 原生 REST API 参考
- [OpenAI 兼容 API](openai-compat.md) - OpenAI SDK 兼容性
- [Anthropic 兼容 API](anthropic-compat.md) - Anthropic SDK 兼容性
- [serve 命令](../cli/serve.md) - 服务器命令参考
