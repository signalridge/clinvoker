# clinvk serve

启动 HTTP API 服务。

## 用法

```bash
clinvk serve [flags]
```

## 说明

启动 HTTP 服务器并暴露三类 API：

- 自定义 REST API（`/api/v1/`）
- OpenAI 兼容 API（`/openai/v1/`）
- Anthropic 兼容 API（`/anthropic/v1/`）

## 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--host` | | string | `127.0.0.1` | 绑定地址（可由配置覆盖） |
| `--port` | `-p` | int | `8080` | 监听端口（可由配置覆盖） |

## 认证

可选 API Key 认证（未配置时默认放行）：

- 环境变量 `CLINVK_API_KEYS`（逗号分隔）
- 配置 `server.api_keys_gopass_path`（gopass）

请求需带：

- `X-Api-Key: <key>`
- 或 `Authorization: Bearer <key>`

## 端点

### 自定义 REST API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/prompt` | 执行单次提示 |
| POST | `/api/v1/parallel` | 并行执行 |
| POST | `/api/v1/chain` | 串行执行 |
| POST | `/api/v1/compare` | 后端对比 |
| GET | `/api/v1/backends` | 列出后端 |
| GET | `/api/v1/sessions` | 列出会话 |
| GET | `/api/v1/sessions/{id}` | 查看会话 |
| DELETE | `/api/v1/sessions/{id}` | 删除会话 |

### OpenAI 兼容

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/openai/v1/models` | 列出模型 |
| POST | `/openai/v1/chat/completions` | 对话补全 |

### Anthropic 兼容

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/anthropic/v1/messages` | 创建消息 |

### Meta

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/openapi.json` | OpenAPI 说明 |
| GET | `/docs` | API 文档 UI |
| GET | `/metrics` | Prometheus 指标（需开启） |

## 配置示例

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
  read_timeout_secs: 30
  write_timeout_secs: 300
  idle_timeout_secs: 120
  api_keys_gopass_path: "myproject/server/api-keys"
  rate_limit_enabled: false
  metrics_enabled: false
```

## 启动输出

```text
clinvk API server starting on http://127.0.0.1:8080

Available endpoints:
  Custom API:     /api/v1/prompt, /api/v1/parallel, /api/v1/chain, /api/v1/compare
  OpenAI:         /openai/v1/models, /openai/v1/chat/completions
  Anthropic:      /anthropic/v1/messages
  Docs:           /openapi.json
  Health:         /health

Press Ctrl+C to stop
```

## 另请参阅

- [REST API](../api/rest-api.md)
- [OpenAI Compatible](../api/openai-compatible.md)
- [Anthropic Compatible](../api/anthropic-compatible.md)
