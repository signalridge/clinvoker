# clinvk config

管理配置。

## 用法

```bash
clinvk config [command]
```

## 子命令

| 命令 | 说明 |
|------|------|
| `show` | 显示当前配置摘要 |
| `set` | 设置配置项 |

---

## clinvk config show

显示当前配置与后端可用性摘要。

```bash
clinvk config show
```

输出示例：

```text
Default Backend: claude

Backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
  codex:
    model: o3
  gemini:
    model: gemini-2.5-pro

Session:
  auto_resume: true
  retention_days: 30

Available backends:
  claude: available
  codex: not installed
  gemini: available
```

---

## clinvk config set

设置配置项。

```bash
clinvk config set <key> <value>
```

示例：

```bash
clinvk config set default_backend gemini
clinvk config set backends.claude.model claude-sonnet-4-20250514
clinvk config set session.retention_days 60
clinvk config set server.port 3000
clinvk config set backends.gemini.enabled false
clinvk config set parallel.max_workers 5
```

> `backends.<name>.enabled` 会被写入配置，但当前 CLI 不会强制禁用该后端（仅保存该标记）。

---

## 配置文件

配置文件位于 `~/.clinvk/config.yaml`。

完整配置请参见 [Configuration Reference](../configuration.zh.md)。

## 配置优先级

1. CLI 参数（最高）
2. 环境变量
3. 配置文件
4. 默认值（最低）

## 另请参阅

- [Configuration Reference](../configuration.zh.md)
- [Environment Variables](../environment.md)
