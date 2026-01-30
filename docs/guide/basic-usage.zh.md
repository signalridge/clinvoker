# 基础使用

日常使用 `clinvk` 的核心流程。

## 运行 prompt

```bash
clinvk "解释这段代码"
```

未指定后端时，优先级为：

1. 配置中的 `default_backend`
2. 回退到 `claude`

## 选择后端

```bash
clinvk -b claude "评审这个模块"
clinvk -b codex "优化这个函数"
clinvk -b gemini "总结这份文档"
```

## 指定模型

```bash
clinvk -b claude -m claude-opus-4-5-20251101 "深度评审"
clinvk -b codex -m o3 "重构这段代码"
```

## 设置工作目录

```bash
clinvk --workdir /path/to/project "扫描 TODO"
```

省略 `--workdir` 时默认使用当前目录。

## 输出格式

可在命令行或配置中设置（`output.format`）。

### Text

```bash
clinvk --output-format text "快速总结"
```

### JSON（默认）

```bash
clinvk --output-format json "快速总结"
```

### Stream JSON（CLI）

```bash
clinvk --output-format stream-json "流式输出"
```

说明：

- CLI 流式输出为**后端原生格式**。
- 服务器流式输出请参见 [HTTP 服务器](http-server.md)（统一事件）。

## 继续最近会话

```bash
clinvk "写一份迁移计划"
clinvk -c "补充回滚策略"
```

`-c/--continue` 会继续最近可恢复的会话；若没有，会创建新会话。

## Dry run

```bash
clinvk --dry-run "生成发布说明"
```

Dry run 仅打印将要执行的命令，不真正执行。

## 无状态执行

```bash
clinvk --ephemeral "一次性问题"
```

不会创建会话。若后端不支持原生无状态，clinvk 会尽量清理会话。

## 输出附加信息

Text 输出可追加 token 和耗时：

```yaml
output:
  show_tokens: true
  show_timing: true
```

（仅对 text 输出生效。）

## 下一步

- [会话管理](session-management.md)
- [并行执行](parallel-execution.md)
- [后端对比](backend-comparison.md)
