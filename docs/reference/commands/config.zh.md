# clinvk config

管理配置。

## 概要

```bash
clinvk config [command]
```

## 子命令

| 命令 | 描述 |
|------|------|
| `show` | 显示当前配置 |
| `set` | 设置配置值 |

---

## clinvk config show

显示解析了所有来源的当前配置。

### 用法

```bash
clinvk config show
```

---

## clinvk config set

设置配置值。

### 用法

```bash
clinvk config set <key> <value>
```

### 示例

```bash
# 设置默认后端
clinvk config set default_backend gemini

# 设置后端特定模型
clinvk config set backends.claude.model claude-sonnet-4-20250514

# 设置会话保留期限
clinvk config set session.retention_days 60

# 设置服务器端口
clinvk config set server.port 3000
```

### 键路径格式

使用点表示法访问嵌套值：

| 键路径 | 描述 |
|--------|------|
| `default_backend` | 默认后端 |
| `backends.<name>.model` | 后端模型 |
| `backends.<name>.enabled` | 启用/禁用后端 |
| `session.retention_days` | 会话保留期限 |
| `server.port` | 服务器端口 |

---

## 配置文件

配置存储在 `~/.clinvk/config.yaml`。

详见 [配置参考](../configuration.md) 了解所有选项的完整文档。

## 另请参阅

- [配置参考](../configuration.md)
- [环境变量](../environment.md)
