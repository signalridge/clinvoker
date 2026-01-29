# clinvk serve

启动 HTTP API 服务器。

## 概要

```
clinvk serve [flags]
```

## 描述

启动一个通过 REST API 暴露 clinvk 功能的 HTTP 服务器。服务器提供三种 API 风格：

- 自定义 REST API (`/api/v1/`)
- OpenAI 兼容 API (`/openai/v1/`)
- Anthropic 兼容 API (`/anthropic/v1/`)

## 标志

| 标志 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--host` | | string | `127.0.0.1` | 绑定的主机 |
| `--port` | `-p` | int | `8080` | 监听端口 |

## 示例

### 使用默认设置启动

```bash
clinvk serve
# 服务器运行在 http://127.0.0.1:8080
```

### 自定义端口

```bash
clinvk serve --port 3000
```

### 绑定到所有接口

```bash
clinvk serve --host 0.0.0.0 --port 8080
```

!!! warning "安全"
    绑定到 `0.0.0.0` 会将服务器暴露到网络。没有内置认证。

## 端点

### 自定义 REST API

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | `/api/v1/prompt` | 执行提示 |
| POST | `/api/v1/parallel` | 并行执行 |
| POST | `/api/v1/chain` | 链式执行 |
| POST | `/api/v1/compare` | 后端对比 |
| GET | `/api/v1/backends` | 列出后端 |
| GET | `/api/v1/sessions` | 列出会话 |

### OpenAI 兼容

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/openai/v1/models` | 列出模型 |
| POST | `/openai/v1/chat/completions` | 聊天补全 |

### Anthropic 兼容

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | `/anthropic/v1/messages` | 创建消息 |

### 元数据

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/openapi.json` | OpenAPI 规范 |

## 退出码

| 代码 | 描述 |
|------|------|
| 0 | 正常关闭 |
| 1 | 服务器错误 |

## 另请参阅

- [REST API 参考](../../server/rest-api.md)
- [OpenAI 兼容](../../server/openai-compatible.md)
- [Anthropic 兼容](../../server/anthropic-compatible.md)
