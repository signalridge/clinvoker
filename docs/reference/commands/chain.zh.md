# clinvk chain

执行顺序提示管道。

## 概要

```bash
clinvk chain [flags]
```

## 描述

按顺序执行一系列提示，将每个步骤的输出传递给下一个。这使得多阶段工作流成为可能，不同后端可以发挥各自的优势。

## 标志

| 标志 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--file` | `-f` | string | | 管道文件 (JSON) |
| `--json` | | bool | `false` | JSON 输出 |

## 管道文件格式

```json
{
  "steps": [
    {
      "name": "步骤名称",
      "backend": "claude",
      "prompt": "第一个提示"
    },
    {
      "name": "第二步",
      "backend": "gemini",
      "prompt": "处理这个：{{previous}}"
    }
  ]
}
```bash

### 步骤字段

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `name` | string | 否 | 步骤标识符 |
| `backend` | string | 是 | 使用的后端 |
| `prompt` | string | 是 | 提示内容 |
| `model` | string | 否 | 模型覆盖 |
| `workdir` | string | 否 | 工作目录 |
| `approval_mode` | string | 否 | `default`, `auto`, `none`, `always` |
| `sandbox_mode` | string | 否 | `default`, `read-only`, `workspace`, `full` |
| `max_turns` | int | 否 | 最大 agentic 回合数 |

### 顶层字段

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `steps` | array | | 步骤列表（必需） |
| `stop_on_failure` | bool | `true` | 失败即停止 |
| `pass_working_dir` | bool | `false` | 在步骤间传递工作目录 |

### 模板变量

| 变量 | 描述 |
|------|------|
| `{{previous}}` | 上一步的输出文本 |

!!! note "仅临时模式"
    chain 只能以临时模式运行 - 不保存会话，也不支持 `{{session}}`。

## 示例

### 基本链

```bash
clinvk chain --file pipeline.json
```

### JSON 输出

```bash
clinvk chain --file pipeline.json --json
```

## 退出码

| 代码 | 描述 |
|------|------|
| 0 | 所有步骤成功 |
| 1 | 某个步骤失败 |

## 另请参阅

- [parallel](parallel.md) - 并发执行
- [compare](compare.md) - 后端对比
