# 配置

用配置设定默认值，减少命令行参数。

## 路径

默认配置路径：

```
~/.clinvk/config.yaml
```

可用 `--config` 指定自定义路径。

## 优先级

1. CLI 参数
2. 环境变量
3. 配置文件
4. 内置默认值

## 最小示例

```yaml
default_backend: claude

output:
  format: text

session:
  auto_resume: true
  retention_days: 30
```

## 常用设置

### 默认后端

```yaml
default_backend: codex
```

### 输出格式

```yaml
output:
  format: json   # text | json | stream-json
```

### 命令超时

```yaml
unified_flags:
  command_timeout_secs: 600
```

### 会话保留

```yaml
session:
  retention_days: 14
  store_token_usage: true
```

### HTTP 服务器默认值

```yaml
server:
  host: "127.0.0.1"
  port: 8080
```

## CLI 写入配置

```bash
clinvk config set default_backend codex
clinvk config set output.format text
clinvk config set session.auto_resume true
```

`config set` 会写入 `~/.clinvk/config.yaml`，支持点号路径。

## 环境变量

常用环境变量：

- `CLINVK_BACKEND`
- `CLINVK_CLAUDE_MODEL`
- `CLINVK_CODEX_MODEL`
- `CLINVK_GEMINI_MODEL`
- `CLINVK_API_KEYS`（服务器鉴权）
- `CLINVK_API_KEYS_GOPASS_PATH`

完整列表见：[环境变量](../reference/environment.md)。

## 详细参考

参见 [配置参考](../reference/configuration.md)。
