# 命令参考

clinvk CLI 的完整命令说明。

## 语法

```bash
clinvk [flags] [prompt]
clinvk [command] [flags]
```

## 命令列表

| 命令 | 说明 |
|---------|-------------|
| [`[prompt]`](prompt.md) | 执行 prompt（根命令） |
| [`resume`](resume.md) | 恢复会话 |
| [`sessions`](sessions.md) | 管理会话 |
| [`config`](config.md) | 管理配置 |
| [`parallel`](parallel.md) | 并行执行 |
| [`compare`](compare.md) | 多后端对比 |
| [`chain`](chain.md) | 链式执行 |
| [`serve`](serve.md) | 启动 HTTP API 服务 |
| `version` | 显示版本信息 |
| `help` | 查看帮助 |

## 持久化参数

以下参数适用于所有命令：

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|-------|------|---------|-------------|
| `--backend` | `-b` | string | 配置 / `claude` | 选择后端 |
| `--model` | `-m` | string | | 模型覆盖 |
| `--workdir` | `-w` | string | 当前目录 | 工作目录 |
| `--output-format` | `-o` | string | 配置 / `json` | `text` / `json` / `stream-json` |
| `--config` | | string | `~/.clinvk/config.yaml` | 配置文件路径 |
| `--dry-run` | | bool | `false` | 仅打印命令，不执行 |
| `--ephemeral` | | bool | `false` | 无状态执行（不保存会话） |

说明：

- 未显式传参时，会读取配置默认值。
- `--output-format` 会自动转为小写。

## 仅根命令参数

| 参数 | 简写 | 说明 |
|------|-------|-------------|
| `--continue` | `-c` | 继续最近可恢复会话 |
