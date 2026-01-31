# clinvk parallel

并行执行多个任务。

## 用法

```bash
clinvk parallel [flags]
```text

## 说明

并发执行多个 AI 任务。任务定义来自 JSON 文件或 stdin。

**注意：** CLI 的 `parallel` 始终为无状态执行（不会持久化会话），并且任务字段 `output_format` 当前会被忽略（内部强制 JSON 以便解析）。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--file` | `-f` | string | | 任务文件（JSON） |
| `--max-parallel` | | int | `3` | 最大并发数（覆盖配置） |
| `--fail-fast` | | bool | `false` | 任一任务失败立即终止 |
| `--json` | | bool | `false` | JSON 输出 |
| `--quiet` | `-q` | bool | `false` | 不输出任务内容，仅输出结果 |

## 任务文件格式

```json
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "task prompt",
      "model": "optional-model",
      "workdir": "/optional/path",
      "approval_mode": "auto",
      "sandbox_mode": "workspace",
      "max_turns": 10
    }
  ],
  "max_parallel": 3,
  "fail_fast": true
}
```text

### 任务字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `backend` | string | 是 | 选择后端 |
| `prompt` | string | 是 | 提示词 |
| `model` | string | 否 | 覆盖模型 |
| `workdir` | string | 否 | 工作目录 |
| `approval_mode` | string | 否 | `default` / `auto` / `none` / `always` |
| `sandbox_mode` | string | 否 | `default` / `read-only` / `workspace` / `full` |
| `output_format` | string | 否 | 目前会被忽略（预留） |
| `max_tokens` | int | 否 | 最大 token（当前未映射到后端参数） |
| `max_turns` | int | 否 | 最大回合数 |
| `system_prompt` | string | 否 | 系统提示词 |
| `extra` | array | 否 | 额外后端参数 |
| `verbose` | bool | 否 | 详细输出 |
| `dry_run` | bool | 否 | 仅模拟执行 |
| `id` | string | 否 | 任务 ID |
| `name` | string | 否 | 任务名称 |
| `tags` | array | 否 | 写入 JSON 输出 / `output_dir` |
| `meta` | object | 否 | 自定义元数据 |

### 顶层字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `tasks` | array | 任务列表 |
| `max_parallel` | int | 最大并发 |
| `fail_fast` | bool | 失败即停 |
| `output_dir` | string | 持久化输出目录（写入 `summary.json` 和单任务 JSON） |

## 示例

### 从文件读取

```bash
clinvk parallel --file tasks.json
```text

### 从 stdin

```bash
cat tasks.json | clinvk parallel
```text

### 限制并发

```bash
clinvk parallel --file tasks.json --max-parallel 2
```text

### Fail-Fast

```bash
clinvk parallel --file tasks.json --fail-fast
```text

### JSON 输出

```bash
clinvk parallel --file tasks.json --json
```text

### 持久化输出

```bash
cat tasks.json | jq '. + {"output_dir": "parallel_runs/run-001"}' | clinvk parallel
```text

## 输出

### 文本输出

```text
Running 3 tasks (max 3 parallel)...

[1] The auth module looks good...
[2] Added logging statements...
[3] Generated 5 test cases...

Results:
--------------------------------------------------------------------------------
#    BACKEND      STATUS   DURATION   TASK
--------------------------------------------------------------------------------
1    claude       OK       2.50s      review the auth module
2    codex        OK       3.20s      add logging to the API
3    gemini       OK       2.80s      generate tests for utils
--------------------------------------------------------------------------------
Total: 3 tasks, 3 completed, 0 failed (3.20s)
```text

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
      "task_id": "task-1",
      "task_name": "Auth Review",
      "backend": "claude",
      "output": "The auth module looks good...",
      "duration_seconds": 2.5,
      "exit_code": 0
    }
  ]
}
```text

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 全部成功 |
| 1 | 有任务失败 |

## 另请参阅

- [chain](chain.md) - 串行执行
- [compare](compare.md) - 后端对比
