# 快速开始

几分钟内完成上手。

## 1) 运行第一条 prompt

```bash
clinvk "解释这个仓库是做什么的"
```

如果未设置默认后端，clinvk 会回退到 `claude`。

## 2) 显式选择后端

```bash
clinvk -b codex "优化这个函数"
clinvk -b gemini "总结这份文档"
```

## 3) 体验会话

```bash
clinvk "写一份迁移计划"
clinvk -c "补充回滚策略"
```

`-c/--continue` 会继续最近可恢复的会话。

## 4) 并行任务

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "评审架构"},
    {"backend": "codex", "prompt": "评审性能"},
    {"backend": "gemini", "prompt": "评审安全"}
  ]
}
```

```bash
clinvk parallel -f tasks.json
```

## 5) 启动 API 服务器

```bash
clinvk serve --port 8080
```

用 curl 调用：

```bash
curl -sS http://localhost:8080/api/v1/prompt \
  -H 'Content-Type: application/json' \
  -d '{"backend":"claude","prompt":"Hello"}'
```

## 下一步

- [配置](configuration.md) — 设定默认项
- [基础使用](basic-usage.md) — 核心 CLI 工作流
- [HTTP 服务器](http-server.md) — OpenAI/Anthropic 兼容
