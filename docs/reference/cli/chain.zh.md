# clinvk chain

顺序执行多步提示词流水线。

## 用法

```bash
clinvk chain [flags]
```text

## 说明

按顺序执行多个步骤，并用 `{{previous}}` 将上一步输出传递给下一步。

**注意：** CLI 的 `chain` 始终为无状态执行（不持久化会话）。`{{session}}`、`pass_session_id`、`persist_sessions` 不支持并会报错。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--file` | `-f` | string | | 流水线文件（JSON） |
| `--json` | | bool | `false` | JSON 输出 |

## 流水线文件格式

```json
{
  "steps": [
    {
      "name": "step-name",
      "backend": "claude",
      "prompt": "First prompt",
      "model": "optional-model"
    },
    {
      "name": "second-step",
      "backend": "gemini",
      "prompt": "Process this: {{previous}}"
    }
  ]
}
```text

### 步骤字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 否 | 步骤名称 |
| `backend` | string | 是 | 后端 |
| `prompt` | string | 是 | 提示词 |
| `model` | string | 否 | 模型覆盖 |
| `workdir` | string | 否 | 工作目录 |
| `approval_mode` | string | 否 | `default` / `auto` / `none` / `always` |
| `sandbox_mode` | string | 否 | `default` / `read-only` / `workspace` / `full` |
| `max_turns` | int | 否 | 最大回合数 |

### 顶层字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `steps` | array | | 步骤列表（必填） |
| `stop_on_failure` | bool | `true` | **CLI 始终失败即停**（字段会被接受，但 `false` 会被忽略） |
| `pass_working_dir` | bool | `false` | 传递工作目录 |

### 模板变量

| 变量 | 说明 |
|------|------|
| `{{previous}}` | 上一步输出 |

## 输出

### 文本输出

```text
Executing chain with 3 steps
================================================================================

[1/3] analyze (claude)
--------------------------------------------------------------------------------
Analysis result text...

[2/3] recommend (gemini)
--------------------------------------------------------------------------------
Recommendations text...

[3/3] implement (codex)
--------------------------------------------------------------------------------
Implementation text...

================================================================================
CHAIN EXECUTION SUMMARY
================================================================================
STEP   BACKEND      STATUS   DURATION   NAME
--------------------------------------------------------------------------------
1      claude       OK       2.10s      analyze
2      gemini       OK       1.80s      recommend
3      codex        OK       3.20s      implement
--------------------------------------------------------------------------------
Total: 3/3 steps completed (7.10s)
```text

### JSON 输出

```json
{
  "total_steps": 3,
  "completed_steps": 3,
  "failed_step": 0,
  "total_duration_seconds": 7.1,
  "results": [
    {
      "step": 1,
      "name": "analyze",
      "backend": "claude",
      "output": "Analysis result...",
      "duration_seconds": 2.1,
      "exit_code": 0
    }
  ]
}
```text

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 全部成功 |
| 1 | 有步骤失败 |

## 另请参阅

- [parallel](parallel.md) - 并行执行
- [compare](compare.md) - 后端对比
