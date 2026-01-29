# 服务器入门

几分钟内启动并运行 clinvk HTTP 服务器。

## 启动服务器

### 基本启动

```bash
clinvk serve
```

服务器默认在 `http://127.0.0.1:8080` 启动。

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

## 测试服务器

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：

```json
{"status": "ok"}
```

### 列出后端

```bash
curl http://localhost:8080/api/v1/backends
```

### 执行提示

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello world"}'
```

## 配置

### 通过配置文件

编辑 `~/.clinvk/config.yaml`：

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300
```

### 通过 CLI 参数

```bash
clinvk serve --host 0.0.0.0 --port 3000
```

CLI 参数覆盖配置文件设置。

## 使用客户端库

### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # clinvk 不需要认证
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "你好！"}]
)
print(response.choices[0].message.content)
```

### Python (Anthropic SDK)

```python
import anthropic

client = anthropic.Client(
    base_url="http://localhost:8080/anthropic/v1",
    api_key="not-needed"
)

message = client.messages.create(
    model="claude",
    max_tokens=1024,
    messages=[{"role": "user", "content": "你好！"}]
)
print(message.content[0].text)
```

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
console.log(data.response);
```

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
```

启用并启动：

```bash
sudo systemctl enable clinvk
sudo systemctl start clinvk
```

### Docker

```bash
docker run -d \
  --name clinvk \
  -p 8080:8080 \
  ghcr.io/signalridge/clinvk serve --host 0.0.0.0
```

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
```

加载服务：

```bash
launchctl load ~/Library/LaunchAgents/com.clinvk.server.plist
```

## 下一步

- [REST API 参考](rest-api.md) - 完整 API 文档
- [OpenAI 兼容](openai-compatible.md) - 使用 OpenAI 客户端
- [Anthropic 兼容](anthropic-compatible.md) - 使用 Anthropic 客户端
