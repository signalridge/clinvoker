# HTTP 服务器

将 clinvk 作为 API 服务器暴露，提供三种 API 形态：

- **自定义 REST API**：`/api/v1/*`
- **OpenAI 兼容 API**：`/openai/v1/*`
- **Anthropic 兼容 API**：`/anthropic/v1/*`

## 启动服务器

```bash
clinvk serve --host 127.0.0.1 --port 8080
```

默认值来自配置 `server.host` 和 `server.port`。

## 端点一览

### 自定义 REST API

- `POST /api/v1/prompt`
- `POST /api/v1/parallel`
- `POST /api/v1/chain`
- `POST /api/v1/compare`
- `GET /api/v1/backends`
- `GET /api/v1/sessions`
- `GET /api/v1/sessions/{id}`
- `DELETE /api/v1/sessions/{id}`
- `GET /health`

### OpenAI 兼容

- `GET /openai/v1/models`
- `POST /openai/v1/chat/completions`

### Anthropic 兼容

- `POST /anthropic/v1/messages`

### OpenAPI Schema

- `GET /openapi.json`

## 示例：自定义 prompt

```bash
curl -sS http://localhost:8080/api/v1/prompt \
  -H 'Content-Type: application/json' \
  -d '{"backend":"claude","prompt":"Hello"}'
```

## 流式输出（自定义 API）

将 `output_format` 设为 `stream-json`，获得 NDJSON 的统一事件流。

```bash
curl -N http://localhost:8080/api/v1/prompt \
  -H 'Content-Type: application/json' \
  -d '{"backend":"claude","prompt":"Stream please","output_format":"stream-json"}'
```

## 会话与无状态

- **自定义 REST API** 默认创建会话（除非 `ephemeral: true`）。
- **OpenAI/Anthropic 端点** 为 **无状态**（不保存会话）。

## 安全与限制

在 `server:` 中配置：

- **API Key**：`CLINVK_API_KEYS` 或 `CLINVK_API_KEYS_GOPASS_PATH`
- **限流**：`rate_limit_enabled` / `rate_limit_rps` / `rate_limit_burst`
- **请求体大小**：`max_request_body_bytes`
- **CORS**：`cors_allowed_origins` / `cors_allow_credentials` / `cors_max_age`
- **工作目录限制**：`allowed_workdir_prefixes` / `blocked_workdir_prefixes`
- **超时**：`request_timeout_secs` / `read_timeout_secs` / `write_timeout_secs` / `idle_timeout_secs`
- **指标**：`metrics_enabled`（暴露 `/metrics`）

详见 [配置参考](../reference/configuration.md)。
