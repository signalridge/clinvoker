# 环境变量

这里列出当前 `clinvk` 代码实际支持的环境变量。

## 适用范围

本页只包含已在代码中生效的变量。
未列出的变量不应依赖。

## 支持的环境变量

### 核心运行配置

| 变量 | 说明 | 示例 |
|---|---|---|
| `CLINVK_BACKEND` | 默认后端 | `claude`, `codex`, `gemini` |
| `CLINVK_CLAUDE_MODEL` | Claude 默认模型 | `claude-opus-4-5-20251101` |
| `CLINVK_CODEX_MODEL` | Codex 默认模型 | `o3`, `o3-mini` |
| `CLINVK_GEMINI_MODEL` | Gemini 默认模型 | `gemini-2.5-pro` |

### MCP 服务器

| 变量 | 说明 | 示例 |
|---|---|---|
| `CLINVK_MCP_TRANSPORT` | MCP 传输类型 | `stdio`, `http` |
| `CLINVK_MCP_HOST` | MCP HTTP 绑定地址 | `127.0.0.1`, `0.0.0.0` |
| `CLINVK_MCP_PORT` | MCP HTTP 端口 | `8081` |
| `CLINVK_MCP_HTTP_PATH` | MCP HTTP 路径 | `/mcp` |
| `CLINVK_MCP_EXPOSE_HEALTH` | 是否在 MCP 模式暴露 `/health` | `true`, `false` |

### HTTP API 认证

| 变量 | 说明 | 示例 |
|---|---|---|
| `CLINVK_API_KEYS` | `serve`/`mcp` 认证 API Key（逗号分隔） | `key1,key2,key3` |
| `CLINVK_API_KEYS_GOPASS_PATH` | 从 gopass 读取 API Key 的路径 | `myproject/clinvk/api-keys` |

### 后端提供商 API Key

| 变量 | 后端 | 说明 |
|---|---|---|
| `ANTHROPIC_API_KEY` | Claude | 提供商 API Key |
| `OPENAI_API_KEY` | Codex | 提供商 API Key |
| `GOOGLE_API_KEY` | Gemini | 提供商 API Key |

## 常见误区（当前不支持）

以下变量在当前 `clinvk` 版本中不是正式运行时配置：

- `CLINVK_TIMEOUT`
- `CLINVK_DEBUG`
- `CLINVK_SERVER_PORT`
- `CLINVK_HOME`
- `CLINVK_CONFIG`

建议替代方式：

- 超时配置：使用配置文件中的 `unified_flags.command_timeout_secs`
- 调试排查：使用 `--dry-run`、`--output-format json` 和 stderr 日志
- 指定配置文件：使用 `--config /path/to/config.yaml`

## 示例

### 后端与模型

```bash
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3-mini
clinvk "review this patch"
```

### 使用环境变量启动 MCP

```bash
export CLINVK_MCP_TRANSPORT=http
export CLINVK_MCP_HOST=0.0.0.0
export CLINVK_MCP_PORT=8081
export CLINVK_MCP_HTTP_PATH=/mcp
export CLINVK_MCP_EXPOSE_HEALTH=true

clinvk mcp
```

### HTTP/MCP API Key 认证

```bash
export CLINVK_API_KEYS="prod-key-1,prod-key-2"
clinvk serve --port 8080
```

## 优先级

对受支持的键，优先级如下：

1. CLI 参数
2. 环境变量
3. 配置文件
4. 默认值

## 排查建议

```bash
# 查看已导出的 CLINVK 变量
env | grep '^CLINVK_'

# 确认提供商 API Key 是否设置
echo "${OPENAI_API_KEY:+set}"

# 查看实际将执行的命令
clinvk --dry-run "check behavior"
```

## 相关文档

- [配置参考](configuration.md)
- [CLI config 命令](cli/config.md)
- [CLI mcp 命令](cli/mcp.md)
