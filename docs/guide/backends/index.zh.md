# 后端说明

clinvk 通过外部 CLI 工作。请安装需要的后端，并按需配置默认值。

## 支持的后端

| 后端 | CLI 命令 | 说明 |
|---------|-------------|-------|
| Claude Code | `claude` | 会话与 system prompt 支持最完整 |
| Codex CLI | `codex` | 使用 `codex exec --json` 执行 |
| Gemini CLI | `gemini` | 支持 `--output-format` 与会话清理 |

## 按后端配置

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all
  codex:
    model: o3
  gemini:
    model: gemini-2.5-pro
```

查看：

- [Claude Code](claude.md)
- [Codex CLI](codex.md)
- [Gemini CLI](gemini.md)
