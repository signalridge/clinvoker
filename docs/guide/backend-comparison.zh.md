# 后端对比

并排比较多个 AI 后端的响应，获取同一问题的不同视角。

## 概述

`compare` 命令同时向多个后端发送相同的提示，并将它们的响应一起显示。这适用于：

- 获取多样化的视角
- 评估后端的优势
- 做出明智的决策
- 学习不同 AI 模型如何处理问题

## 基本用法

### 对比所有后端

对所有启用的后端运行：

```bash
clinvk compare --all-backends "解释这个算法"
```bash

### 对比特定后端

选择要对比的后端：

```bash
clinvk compare --backends claude,codex "这段代码是做什么的"
clinvk compare --backends claude,gemini "审查这个 PR"
```

## 执行模式

### 并行（默认）

同时运行所有后端：

```bash
clinvk compare --all-backends "解释这段代码"
```bash

### 顺序

一次运行一个后端：

```bash
clinvk compare --all-backends --sequential "审查这个实现"
```

顺序模式适用于：

- 避免速率限制
- 系统资源受限时
- 希望逐个查看响应时

## 输出格式

### 文本输出（默认）

显示每个后端的响应，带有清晰的分隔：

```yaml
Prompt: 解释这个算法

━━━ claude ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: claude-opus-4-5-20251101
Duration: 2.5s

这个算法实现了二分查找...

━━━ codex ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: o3
Duration: 3.2s

该算法执行二分查找...

━━━ gemini ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model: gemini-2.5-pro
Duration: 2.8s

这是一个经典的二分查找实现...
```

### JSON 输出

获取结构化数据以便程序处理：

```bash
clinvk compare --all-backends --json "解释这段代码"
```yaml

输出：

```json
{
  "prompt": "解释这段代码",
  "backends": ["claude", "codex", "gemini"],
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "output": "这个算法实现了二分查找...",
      "duration_seconds": 2.5,
      "exit_code": 0
    },
    {
      "backend": "codex",
      "model": "o3",
      "output": "该算法执行二分查找...",
      "duration_seconds": 3.2,
      "exit_code": 0
    },
    {
      "backend": "gemini",
      "model": "gemini-2.5-pro",
      "output": "这是一个经典的二分查找实现...",
      "duration_seconds": 2.8,
      "exit_code": 0
    }
  ],
  "total_duration_seconds": 3.2
}
```

## 使用场景

### 代码审查

获取多个视角的代码质量评估：

```bash
clinvk compare --all-backends "审查这段代码的 bug 和改进建议"
```bash

### 架构决策

比较设计选择的建议：

```bash
clinvk compare --backends claude,gemini "这里实现缓存的最佳方式是什么？"
```

### 学习

查看不同 AI 模型如何解释概念：

```bash
clinvk compare --all-backends "解释 JavaScript 中的 async/await 工作原理"
```bash

### 验证

交叉检查重要决策：

```bash
clinvk compare --all-backends "这个实现安全吗？"
```

## 失败处理

如果后端失败，对比会继续处理剩余的后端：

```yaml
Comparing 3 backends: claude, codex, gemini
Prompt: explain this code
============================================================
[claude] 响应内容...
[codex] Error: Backend unavailable
[gemini] 响应内容...

============================================================
COMPARISON SUMMARY
============================================================
BACKEND      STATUS     DURATION     SESSION    MODEL
------------------------------------------------------------
claude       OK         2.50s        abc123     claude-opus-4-5-20251101
codex        FAILED     0.50s        -          o3
             Error: Backend unavailable
gemini       OK         2.80s        def456     gemini-2.5-pro
------------------------------------------------------------
Total time: 2.80s
```

## 命令选项

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `--backends` | 逗号分隔的后端列表 | - |
| `--all-backends` | 对比所有启用的后端 | `false` |
| `--sequential` | 一次运行一个 | `false` |
| `--json` | JSON 输出 | `false` |

## 提示

!!! tip "用于重要决策"
    对于关键代码更改或架构决策，比较多个后端可以发现单个可能遗漏的问题。

!!! tip "注意差异"
    关注后端一致的地方（高置信度）和分歧的地方（值得调查）。

!!! tip "考虑响应时间"
    JSON 输出包含持续时间，可用于基准测试后端性能。

## 下一步

- [并行执行](parallel-execution.md) - 并发运行独立任务
- [链式执行](chain-execution.md) - 顺序多后端管道
- [后端指南](backends/index.md) - 了解每个后端
