# clinvk config

管理配置。

## 语法

```bash
clinvk config [command]
```

## 子命令

### show

```bash
clinvk config show
```

输出配置摘要与后端可用性。

### set

```bash
clinvk config set <key> <value>
```

使用点号路径写入 `~/.clinvk/config.yaml`。

示例：

```bash
clinvk config set default_backend codex
clinvk config set output.format text
clinvk config set session.auto_resume true
```

说明：

- `set` 会把值写成字符串；复杂结构建议直接编辑 YAML。
- 详细配置见 [配置参考](../configuration.md)。
