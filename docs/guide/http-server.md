# HTTP Server

clinvk includes a built-in HTTP API server that exposes all functionality via REST endpoints.

## Overview

The `clinvk serve` command starts an HTTP server that provides:

- **Custom REST API** - Full access to all clinvk features
- **OpenAI Compatible API** - Drop-in replacement for OpenAI clients
- **Anthropic Compatible API** - Drop-in replacement for Anthropic clients

## Starting the Server

### Basic Start

```bash
clinvk serve
```

The server starts on `http://127.0.0.1:8080` by default.

### Custom Port

```bash
clinvk serve --port 3000
```

### Bind to All Interfaces

```bash
clinvk serve --host 0.0.0.0 --port 8080
```

!!! warning "Security"
    Binding to `0.0.0.0` exposes the server to the network. There is no built-in authentication.

## Testing the Server

### Health Check

```bash
curl http://localhost:8080/health
```

Response:

```json
{"status": "ok"}
```

### List Backends

```bash
curl http://localhost:8080/api/v1/backends
```

### Execute a Prompt

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello world"}'
```

## Endpoints Overview

### Custom REST API (`/api/v1/`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/prompt` | Execute single prompt |
| POST | `/api/v1/parallel` | Execute multiple prompts |
| POST | `/api/v1/chain` | Execute prompt chain |
| POST | `/api/v1/compare` | Compare backend responses |
| GET | `/api/v1/backends` | List backends |
| GET | `/api/v1/sessions` | List sessions |
| GET | `/api/v1/sessions/{id}` | Get session |
| DELETE | `/api/v1/sessions/{id}` | Delete session |

### OpenAI Compatible (`/openai/v1/`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/openai/v1/models` | List models |
| POST | `/openai/v1/chat/completions` | Chat completion |

### Anthropic Compatible (`/anthropic/v1/`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/anthropic/v1/messages` | Create message |

### Meta Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/openapi.json` | OpenAPI spec |

## Using with Client Libraries

### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"  # clinvk doesn't require auth
)

response = client.chat.completions.create(
    model="claude",
    messages=[{"role": "user", "content": "Hello!"}]
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
    messages=[{"role": "user", "content": "Hello!"}]
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

### cURL

```bash
# Simple prompt
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "explain async/await"}'

# Parallel execution
curl -X POST http://localhost:8080/api/v1/parallel \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {"backend": "claude", "prompt": "task 1"},
      {"backend": "codex", "prompt": "task 2"}
    ]
  }'
```

## Configuration

### Via Config File

Edit `~/.clinvk/config.yaml`:

```yaml
server:
  # Bind address
  host: "127.0.0.1"

  # Port number
  port: 8080

  # Request processing timeout (seconds)
  request_timeout_secs: 300

  # Read request timeout (seconds)
  read_timeout_secs: 30

  # Write response timeout (seconds)
  write_timeout_secs: 300

  # Idle connection timeout (seconds)
  idle_timeout_secs: 120
```

### Via CLI Flags

```bash
clinvk serve --host 0.0.0.0 --port 3000
```

CLI flags override config file settings.

## Running as a Service

### systemd (Linux)

Create `/etc/systemd/system/clinvk.service`:

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

Enable and start:

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

Create `~/Library/LaunchAgents/com.clinvk.server.plist`:

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

Load the service:

```bash
launchctl load ~/Library/LaunchAgents/com.clinvk.server.plist
```

## Security Notes

!!! warning "Local Use Only"
    By default, the server binds to `127.0.0.1` (localhost only). If you bind to `0.0.0.0` to expose it publicly, be aware that **there is no authentication**.

!!! tip "Production Use"
    For production deployments, place the server behind a reverse proxy (nginx, Caddy) that handles:

    - TLS termination
    - Authentication
    - Rate limiting
    - Request logging

## OpenAPI Specification

The server provides an OpenAPI specification at `/openapi.json`:

```bash
# Download spec
curl http://localhost:8080/openapi.json > openapi.json

# View in Swagger UI or import into API tools
```

## Next Steps

- [REST API Reference](../reference/rest-api.md) - Full API documentation
- [OpenAI Compatible](../reference/openai-compatible.md) - Use with OpenAI clients
- [Anthropic Compatible](../reference/anthropic-compatible.md) - Use with Anthropic clients
