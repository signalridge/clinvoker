# HTTP 服务器

clinvk 内置 HTTP API 服务器，通过 REST 端点暴露所有功能。

## 概述

`clinvk serve` 命令启动一个 HTTP 服务器，提供：

- **自定义 REST API** - 完全访问所有 clinvk 功能
- **OpenAI 兼容 API** - 可直接替换 OpenAI 客户端
- **Anthropic 兼容 API** - 可直接替换 Anthropic 客户端

## 启动服务器

### 基本启动

```bash
clinvk serve
```text

服务器默认在 `http://127.0.0.1:8080` 启动。

### 自定义端口

```bash
clinvk serve --port 3000
```text

### 绑定到所有接口

```bash
clinvk serve --host 0.0.0.0 --port 8080
```text

!!! warning "安全"
    绑定到 `0.0.0.0` 会将服务器暴露到网络。建议启用 API Key 并限制 CORS/工作目录。

## 认证

可通过 `CLINVK_API_KEYS` 或 `server.api_keys_gopass_path` 启用 API Key。请求需带 `Authorization: Bearer <key>` 或 `X-Api-Key: <key>`。

## 测试服务器

### 健康检查

```bash
curl http://localhost:8080/health
```text

响应包含 `status`、`version`、`uptime`、后端可用性与会话存储状态。

### 列出后端

```bash
curl http://localhost:8080/api/v1/backends
```text

### 执行提示

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello world"}'
```text

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
| GET | `/docs` | API 文档 UI |
| GET | `/metrics` | Prometheus 指标（需开启） |

## 使用客户端库

### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # 开启 API Key 时才需要
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "你好！"}]
)
print(response.choices[0].message.content)
```text

### Python (Anthropic SDK)

```python
import anthropic

client = anthropic.Client(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"  # 开启 API Key 时才需要
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "你好！"}]
)
print(message.content[0].text)
```text

### JavaScript/TypeScript

```typescript
const response = await fetch('http://localhost:8080/api/v1/prompt', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    backend: 'claude',
    prompt: 'hello world'
  })
});

const data = await response.json();
console.log(data.output);
```text

### cURL

```bash
# 简单提示
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "解释 async/await"}'

# 并行执行
curl -X POST http://localhost:8080/api/v1/parallel \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {"backend": "claude", "prompt": "任务 1"},
      {"backend": "codex", "prompt": "任务 2"}
    ]
  }'
```bash

## 配置

### 通过配置文件

编辑 `~/.clinvk/config.yaml`：

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

  # 可选：gopass 中的 API Key 路径
  api_keys_gopass_path: ""
```text

### 通过 CLI 参数

```bash
clinvk serve --host 0.0.0.0 --port 3000
```bash

CLI 参数覆盖配置文件设置。

## 作为服务运行

### systemd (Linux)

创建 `/etc/systemd/system/clinvk.service`：

```ini
[Unit]
Description=clinvk API Server
After=network.target

[Service]
Type=simple
User=youruser
ExecStart=/usr/local/bin/clinvk serve --port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```text

启用并启动：

```bash
sudo systemctl enable clinvk
sudo systemctl start clinvk
```text

### Docker

```bash
docker run -d \
  --name clinvk \
  -p 8080:8080 \
  ghcr.io/signalridge/clinvk serve --host 0.0.0.0
```bash

### launchd (macOS)

创建 `~/Library/LaunchAgents/com.clinvk.server.plist`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.clinvk.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/clinvk</string>
        <string>serve</string>
        <string>--port</string>
        <string>8080</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```text

加载服务：

```bash
launchctl load ~/Library/LaunchAgents/com.clinvk.server.plist
```text

## 安全说明

!!! warning "仅限本地使用"
    默认情况下，服务器绑定到 `127.0.0.1`（仅限本地）。如果绑定到 `0.0.0.0` 以公开暴露，请注意**没有认证机制**。

!!! tip "生产环境使用"
    对于生产部署，将服务器放在反向代理（nginx、Caddy）后面处理：

    - TLS 终结
    - 认证
    - 速率限制
    - 请求日志

## OpenAPI 规范

服务器在 `/openapi.json` 提供 OpenAPI 规范：

```bash
# 下载规范
curl http://localhost:8080/openapi.json > openapi.json

# 在 Swagger UI 查看或导入 API 工具
```text

## 下一步

- [REST API 参考](../reference/api/rest-api.md) - 完整 API 文档
- [OpenAI 兼容](../reference/api/openai-compatible.md) - 使用 OpenAI 客户端
- [Anthropic 兼容](../reference/api/anthropic-compatible.md) - 使用 Anthropic 客户端
