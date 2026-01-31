# 并行执行

并发运行多个 AI 任务以节省时间并提高生产力。

## 概述

`parallel` 命令在一个或多个后端上同时执行多个任务。这适用于：

- 更快地运行独立任务
- 同时获取多个视角
- 批量处理

## 基本用法

### 创建任务文件

创建 `tasks.json` 文件：

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "审查 auth 模块"
    },
    {
      "backend": "codex",
      "prompt": "为 API 添加日志"
    },
    {
      "backend": "gemini",
      "prompt": "为 utils 生成测试"
    }
  ]
}
```

### 运行任务

```bash
clinvk parallel --file tasks.json
```

### 从标准输入

您也可以通过管道传递任务定义：

```bash
cat tasks.json | clinvk parallel
```

## 任务选项

每个任务可以指定各种选项：

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "审查代码",
      "model": "claude-opus-4-5-20251101",
      "workdir": "/path/to/project",
      "approval_mode": "auto",
      "sandbox_mode": "workspace",
      "output_format": "json",
      "max_tokens": 4096,
      "max_turns": 10,
      "system_prompt": "你是一个代码审查员。"
    }
  ]
}
```

### 任务字段

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `backend` | string | 是 | 使用的后端 (claude, codex, gemini) |
| `prompt` | string | 是 | 要执行的提示 |
| `model` | string | 否 | 模型覆盖 |
| `workdir` | string | 否 | 工作目录 |
| `approval_mode` | string | 否 | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | 否 | `default`, `read-only`, `workspace`, `full` |
| `output_format` | string | 否 | 可填写但 CLI parallel 会忽略（预留） |
| `max_tokens` | int | 否 | 最大响应 token 数 |
| `max_turns` | int | 否 | 最大代理轮次 |
| `system_prompt` | string | 否 | 自定义系统提示 |

## 执行选项

### 限制并行工作器

控制同时运行的任务数：

```bash
# 最多同时运行 2 个任务
clinvk parallel --file tasks.json --max-parallel 2
```

### 快速失败模式

在第一个失败时停止所有任务：

```bash
clinvk parallel --file tasks.json --fail-fast
```

### JSON 输出

获取结构化输出以便程序处理：

```bash
clinvk parallel --file tasks.json --json
```

### 静默模式

抑制任务输出，只显示摘要：

```bash
clinvk parallel --file tasks.json --quiet
```

## 顶级选项

您可以在文件级别指定应用于所有任务的选项：

```json
{
  "tasks": [...],
  "max_parallel": 3,
  "fail_fast": true
}
```

CLI 参数会覆盖文件级别的设置。

## 输出格式

### 文本输出

显示任务完成时的进度和结果：

```text
Running 3 tasks (max 3 parallel)...

[1] auth 模块看起来不错...
[2] 添加了日志语句...
[3] 生成了 5 个测试用例...

Results
============================================================

BACKEND      STATUS   DURATION   TASK

------------------------------------------------------------

1    claude       OK       2.50s      审查 auth 模块
2    codex        OK       3.20s      为 API 添加日志
3    gemini       OK       2.80s      为 utils 生成测试
------------------------------------------------------------

Total: 3 tasks, 3 completed, 0 failed (3.20s)

```

### JSON 输出

```json
{
  "total_tasks": 3,
  "completed": 3,
  "failed": 0,
  "total_duration_seconds": 3.2,
  "results": [
    {
      "index": 0,
      "backend": "claude",
      "output": "auth 模块看起来不错...",
      "duration_seconds": 2.5,
      "exit_code": 0
    }
  ]
}
```

## 配置

`~/.clinvk/config.yaml` 中的默认并行执行设置：

```yaml
parallel:
  # 最大并发任务数
  max_workers: 3

  # 第一个失败时停止
  fail_fast: false

  # 合并所有任务的输出
  aggregate_output: true
```

## 提示

!!! tip "为依赖使用快速失败"
    如果后续任务依赖于前面任务的成功，使用 `--fail-fast` 在失败时立即停止。

!!! tip "平衡并行度"
    运行太多并行任务可能会触及速率限制或资源约束。从 2-3 个工作器开始。

!!! tip "脚本使用 JSON"
    自动化工作流时，使用 `--json` 输出以便可靠解析。

## 下一步

- [链式执行](chain-execution.md) - 顺序管道
- [后端对比](backend-comparison.md) - 比较响应
