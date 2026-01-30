# 快速开始

几分钟内上手使用 clinvk。

## 您的第一个提示

使用默认后端（Claude Code）运行一个简单的提示：

```bash
clinvk "解释这个项目是做什么的"
```

## 指定后端

使用 `--backend` 或 `-b` 参数选择特定后端：

```bash
# 使用 Claude Code
clinvk --backend claude "修复 auth.go 中的 bug"

# 使用 Codex CLI
clinvk -b codex "实现用户注册"

# 使用 Gemini CLI
clinvk -b gemini "生成单元测试"
```

## 继续会话

恢复上一个会话以继续对话：

```bash
# 使用后续提示继续
clinvk --continue "现在添加错误处理"

# 或使用 resume 命令
clinvk resume --last "为更改添加测试"
```

## 对比后端

获取多个后端对同一提示的响应：

```bash
# 对比所有启用的后端
clinvk compare --all-backends "这段代码是做什么的"

# 对比特定后端
clinvk compare --backends claude,codex "解释这个算法"
```

## 并行运行任务

并发执行多个任务。创建 `tasks.json` 文件：

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "审查 auth 模块"},
    {"backend": "codex", "prompt": "为 API 添加日志"},
    {"backend": "gemini", "prompt": "为 utils 生成测试"}
  ]
}
```

运行任务：

```bash
clinvk parallel --file tasks.json
```

## 链式执行后端

顺序通过多个后端传递输出。创建 `pipeline.json`：

```json
{
  "steps": [
    {"name": "review", "backend": "claude", "prompt": "审查这段代码的 bug"},
    {"name": "security", "backend": "gemini", "prompt": "检查安全问题：{{previous}}"},
    {"name": "summary", "backend": "codex", "prompt": "总结发现：{{previous}}"}
  ]
}
```

运行链：

```bash
clinvk chain --file pipeline.json
```

## 启动 HTTP 服务器

将 clinvk 作为 HTTP API 服务器运行：

```bash
# 在默认端口启动 (8080)
clinvk serve

# 自定义端口
clinvk serve --port 3000
```

然后发送 API 请求：

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "claude", "prompt": "hello world"}'
```

## 常用选项

| 选项 | 简写 | 描述 |
|------|------|------|
| `--backend` | `-b` | 使用的后端 (claude, codex, gemini) |
| `--model` | `-m` | 使用的模型 |
| `--workdir` | `-w` | 工作目录 |
| `--output-format` | `-o` | 输出格式 (text, json, stream-json) |
| `--continue` | `-c` | 继续上一个会话 |
| `--dry-run` | | 只显示命令不执行 |

## 下一步

- [基本用法](basic-usage.md) - 详细使用指南
- [会话管理](session-management.md) - 使用会话
- [配置](../reference/configuration.md) - 自定义设置
