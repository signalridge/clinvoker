# HTTP API 服务器

clinvk 内置 HTTP API 服务器，通过 REST 端点暴露所有功能。

## 概述

`clinvk serve` 命令启动一个 HTTP 服务器，提供：

- **自定义 REST API** - 完全访问所有 clinvk 功能
- **OpenAI 兼容 API** - 可直接替换 OpenAI 客户端
- **Anthropic 兼容 API** - 可直接替换 Anthropic 客户端

## 快速开始

```bash
# 使用默认设置启动 (127.0.0.1:8080)
clinvk serve

# 自定义端口
clinvk serve --port 3000

# 绑定到所有接口
clinvk serve --host 0.0.0.0 --port 8080
```

## API 风格

<div class="grid cards" markdown>

-   :material-api:{ .lg .middle } **[REST API](rest-api.md)**

    ---

    用于所有 clinvk 操作的全功能自定义 API

-   :material-openai:{ .lg .middle } **[OpenAI 兼容](openai-compatible.md)**

    ---

    使用现有的 OpenAI 客户端库

-   :material-robot-outline:{ .lg .middle } **[Anthropic 兼容](anthropic-compatible.md)**

    ---

    使用现有的 Anthropic 客户端库

</div>

## 端点概览

### 自定义 REST API (`/api/v1/`)

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | `/api/v1/prompt` | 执行单个提示 |
| POST | `/api/v1/parallel` | 执行多个提示 |
| POST | `/api/v1/chain` | 执行提示链 |
| POST | `/api/v1/compare` | 对比后端响应 |
| GET | `/api/v1/backends` | 列出后端 |
| GET | `/api/v1/sessions` | 列出会话 |
| GET | `/api/v1/sessions/{id}` | 获取会话 |
| DELETE | `/api/v1/sessions/{id}` | 删除会话 |

### OpenAI 兼容 (`/openai/v1/`)

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/openai/v1/models` | 列出模型 |
| POST | `/openai/v1/chat/completions` | 聊天补全 |

### Anthropic 兼容 (`/anthropic/v1/`)

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | `/anthropic/v1/messages` | 创建消息 |

### 元数据端点

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/openapi.json` | OpenAPI 规范 |

## 配置

`~/.clinvk/config.yaml` 中的服务器设置：

```yaml
server:
  # 绑定地址
  host: "127.0.0.1"

  # 端口号
  port: 8080

  # 请求处理超时（秒）
  request_timeout_secs: 300

  # 读取请求超时（秒）
  read_timeout_secs: 30

  # 写入响应超时（秒）
  write_timeout_secs: 300

  # 空闲连接超时（秒）
  idle_timeout_secs: 120
```

## 安全说明

!!! warning "仅限本地使用"
    默认情况下，服务器绑定到 `127.0.0.1`（仅限本地）。如果绑定到 `0.0.0.0` 以公开暴露，请注意**没有认证机制**。

!!! tip "生产环境使用"
    对于生产部署，将服务器放在反向代理（nginx、Caddy）后面处理：

    - TLS 终结
    - 认证
    - 速率限制
    - 请求日志

## 示例请求

### 执行提示

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "解释这段代码"}'
```

### OpenAI 风格聊天

```bash
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "messages": [{"role": "user", "content": "你好！"}]
  }'
```

### 健康检查

```bash
curl http://localhost:8080/health
```

## 下一步

- [入门](getting-started.md) - 逐步服务器设置
- [REST API 参考](rest-api.md) - 完整 API 文档
- [OpenAI 兼容](openai-compatible.md) - 使用 OpenAI 客户端
