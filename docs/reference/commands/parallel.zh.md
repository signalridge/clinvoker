# clinvk parallel

并行执行多个任务。

## 概要

```bash
clinvk parallel [flags]
```

## 描述

并发运行多个 AI 任务。任务定义在 JSON 文件中或通过 stdin 管道传入。

## 标志

| 标志 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--file` | `-f` | string | | 任务文件 (JSON) |
| `--max-parallel` | | int | 3 | 最大并发任务数 |
| `--fail-fast` | | bool | `false` | 第一个失败时停止 |
| `--json` | | bool | `false` | JSON 输出 |
| `--quiet` | `-q` | bool | `false` | 抑制任务输出 |

## 任务文件格式

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "任务提示",
      "model": "可选模型",
      "workdir": "/可选/路径"
    }
  ],
  "max_parallel": 3,
  "fail_fast": true
}
```bash

### 任务字段

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `backend` | string | 是 | 使用的后端 |
| `prompt` | string | 是 | 提示内容 |
| `model` | string | 否 | 模型覆盖 |
| `workdir` | string | 否 | 工作目录 |
| `approval_mode` | string | 否 | 审批模式 |
| `sandbox_mode` | string | 否 | 沙箱模式 |
| `output_format` | string | 否 | `text`, `json`, `stream-json` |
| `max_tokens` | int | 否 | 最大响应 token 数 |
| `max_turns` | int | 否 | 最大代理轮次 |
| `system_prompt` | string | 否 | 系统提示 |
| `extra` | array | 否 | 额外后端参数 |
| `verbose` | bool | 否 | 启用详细输出 |
| `dry_run` | bool | 否 | 仅模拟执行 |
| `id` | string | 否 | 任务标识 |
| `name` | string | 否 | 任务显示名 |
| `tags` | array | 否 | 写入 JSON 输出 / `output_dir` 产物 |
| `meta` | object | 否 | 写入 JSON 输出 / `output_dir` 产物 |

### 顶层字段

| 字段 | 类型 | 描述 |
|------|------|------|
| `tasks` | array | 任务列表 |
| `max_parallel` | int | 最大并发数 |
| `fail_fast` | bool | 失败即停止 |
| `output_dir` | string | 可选输出目录，写入 `summary.json` 和每个任务的 JSON |

## 示例

### 从文件

```bash
clinvk parallel --file tasks.json
```

### 从标准输入

```bash
cat tasks.json | clinvk parallel
```bash

### 限制工作器

```bash
clinvk parallel --file tasks.json --max-parallel 2
```

### 快速失败模式

```bash
clinvk parallel --file tasks.json --fail-fast
```bash

### JSON 输出

```bash
clinvk parallel --file tasks.json --json
```

### 持久化输出

```bash
cat tasks.json | jq '. + {"output_dir": "parallel_runs/run-001"}' | clinvk parallel
```

写入内容：

- `summary.json`（汇总结果）
- 每个任务的 JSON 文件（包含 task + result）

## 退出码

| 代码 | 描述 |
|------|------|
| 0 | 所有任务成功 |
| 1 | 一个或多个任务失败 |

## 另请参阅

- [chain](chain.md) - 顺序执行
- [compare](compare.md) - 后端对比
