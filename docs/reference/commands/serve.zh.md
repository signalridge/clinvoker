# clinvk serve

启动 HTTP API 服务器。

## 语法

```bash
clinvk serve [--host <host>] [--port <port>]
```

## 参数

| 参数 | 简写 | 默认值 | 说明 |
|------|-------|---------|-------------|
| `--host` | | 配置 / `127.0.0.1` | 绑定地址 |
| `--port` | `-p` | 配置 / `8080` | 监听端口 |

## 说明

- 端点包含 `/api/v1/*`、`/openai/v1/*`、`/anthropic/v1/*`。
- OpenAPI Schema 位于 `/openapi.json`。
- 安全与限制配置位于 `server:`。
