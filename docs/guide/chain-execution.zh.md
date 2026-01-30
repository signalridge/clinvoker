# 链式执行

通过多个后端顺序传递提示，每个步骤基于前一个输出构建。

## 概述

`chain` 命令按顺序执行一系列步骤，将一个步骤的输出传递给下一个。这使得复杂的多阶段工作流成为可能，不同后端可以发挥各自的优势。

## 基本用法

### 创建管道文件

创建 `pipeline.json` 文件：

```json
{
  "steps": [
    {
      "name": "initial-review",
      "backend": "claude",
      "prompt": "审查这段代码的 bug"
    },
    {
      "name": "security-check",
      "backend": "gemini",
      "prompt": "检查安全问题：{{previous}}"
    },
    {
      "name": "final-summary",
      "backend": "codex",
      "prompt": "总结发现：{{previous}}"
    }
  ]
}
```

### 运行链

```bash
clinvk chain --file pipeline.json
```

## 模板变量

使用这些占位符在提示中引用之前的输出：

| 变量 | 描述 |
|------|------|
| `{{previous}}` | 上一步的输出文本 |

### 带变量的示例

```json
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "分析这个代码库结构"
    },
    {
      "name": "recommend",
      "backend": "gemini",
      "prompt": "基于这个分析：{{previous}}\n\n推荐改进方案"
    },
    {
      "name": "implement",
      "backend": "codex",
      "prompt": "实现这些推荐：{{previous}}"
    }
  ]
}
```

## 步骤选项

每个步骤可以指定各种选项：

```json
{
  "steps": [
    {
      "name": "step-name",
      "backend": "claude",
      "prompt": "任务描述",
      "model": "claude-opus-4-5-20251101",
      "workdir": "/path/to/project",
      "approval_mode": "auto",
      "sandbox_mode": "workspace",
      "max_turns": 10
    }
  ]
}
```

### 步骤字段

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `name` | string | 否 | 步骤标识符 |
| `backend` | string | 是 | 使用的后端 |
| `prompt` | string | 是 | 提示（支持 `{{previous}}`） |
| `model` | string | 否 | 模型覆盖 |
| `workdir` | string | 否 | 工作目录 |
| `approval_mode` | string | 否 | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | 否 | `default`, `read-only`, `workspace`, `full` |
| `max_turns` | int | 否 | 最大代理轮次 |

## 链式选项

链式执行的顶层字段：

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `stop_on_failure` | bool | `true` | 失败即停止 |
| `pass_working_dir` | bool | `false` | 在步骤间传递工作目录 |

!!! note "仅临时模式"
    chain 只能以临时模式运行 - 不保存会话，也不支持 `{{session}}`。

## 输出选项

### JSON 输出

获取结构化结果以便程序使用：

```bash
clinvk chain --file pipeline.json --json
```

输出：

```json
{
  "total_steps": 2,
  "completed_steps": 2,
  "failed_step": 0,
  "total_duration_seconds": 3.5,
  "results": [
    {
      "step": 1,
      "name": "initial-review",
      "backend": "claude",
      "output": "发现了几个问题...",
      "duration_seconds": 2.0,
      "exit_code": 0
    },
    {
      "step": 2,
      "name": "security-check",
      "backend": "gemini",
      "output": "没有关键漏洞...",
      "duration_seconds": 1.5,
      "exit_code": 0
    }
  ]
}
```

## 使用场景

### 代码审查管道

```json
{
  "steps": [
    {
      "name": "functionality-review",
      "backend": "claude",
      "prompt": "审查这段代码的正确性和逻辑错误"
    },
    {
      "name": "security-review",
      "backend": "gemini",
      "prompt": "审查代码的安全漏洞。之前的分析：{{previous}}"
    },
    {
      "name": "performance-review",
      "backend": "codex",
      "prompt": "审查代码的性能问题。之前的发现：{{previous}}"
    },
    {
      "name": "summary",
      "backend": "claude",
      "prompt": "从所有审查中创建摘要报告：{{previous}}"
    }
  ]
}
```

## 错误处理

如果步骤失败，链会停止并报告错误：

```text
Step 1 (analyze): Completed (2.1s)
Step 2 (implement): Failed - Backend error: rate limit exceeded

Chain failed at step 2
```

## 提示

!!! tip "使用描述性步骤名称"
    好的步骤名称使输出更容易理解和调试。

!!! tip "从简单开始"
    从 2-3 个步骤开始，根据需要添加更多。复杂的链更难调试。

!!! tip "注意上下文长度"
    使用 `{{previous}}` 时，注意早期步骤的输出会增加提示长度。

!!! tip "使用不同后端"
    利用每个后端的优势 - Claude 用于推理，Codex 用于代码生成，Gemini 用于广泛知识。

## 下一步

- [并行执行](parallel-execution.md) - 并发运行独立任务
- [后端对比](backend-comparison.md) - 并排比较响应
