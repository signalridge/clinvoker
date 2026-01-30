# 并行执行

在多个后端上并行执行任务。

## 基本用法

```bash
clinvk parallel --file tasks.json
```

或从 stdin 读取：

```bash
cat tasks.json | clinvk parallel
```

## 任务文件格式

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "评审架构"},
    {"backend": "codex", "prompt": "评审性能"},
    {"backend": "gemini", "prompt": "评审安全"}
  ],
  "max_parallel": 3,
  "fail_fast": false,
  "output_dir": "./parallel-results"
}
```

### Task 字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `backend` | string | 是 | `claude` / `codex` / `gemini` |
| `prompt` | string | 是 | 任务内容 |
| `workdir` | string | 否 | 工作目录 |
| `model` | string | 否 | 模型覆盖 |
| `approval_mode` | string | 否 | `default` / `auto` / `none` / `always`（尽力映射） |
| `sandbox_mode` | string | 否 | `default` / `read-only` / `workspace` / `full`（尽力映射） |
| `max_turns` | int | 否 | 仅 Claude 支持 |
| `max_tokens` | int | 否 | 当前未映射到后端 CLI |
| `system_prompt` | string | 否 | 仅 Claude 支持 |
| `verbose` | bool | 否 | 后端特定行为 |
| `dry_run` | bool | 否 | 单任务 dry run |
| `extra` | array | 否 | 后端额外参数（直接透传；API 会校验） |
| `id` | string | 否 | 任务 ID |
| `name` | string | 否 | 任务名 |
| `tags` | array | 否 | 自定义标签 |
| `meta` | object | 否 | 自定义元数据 |

说明：

- 并行任务 **始终无状态**（不保存会话）。
- CLI 模式下 `output_format` 目前被忽略；请使用 `--json` 或配置输出。

## 执行控制

```bash
# 限制并发
clinvk parallel -f tasks.json --max-parallel 2

# 首错即停
clinvk parallel -f tasks.json --fail-fast

# JSON 输出
clinvk parallel -f tasks.json --json

# 仅输出汇总
clinvk parallel -f tasks.json --quiet
```

### 汇总输出（配置）

```yaml
parallel:
  aggregate_output: true
```

当 `aggregate_output` 为 `false` 时，只打印任务输出，不显示汇总表格。

并发优先级（CLI）：

1. `--max-parallel` 参数
2. 文件中的 `max_parallel`
3. 默认值 = 3

说明：HTTP API 在未提供 `max_parallel` 时会使用配置 `parallel.max_workers`。

## 输出目录

设置 `output_dir` 后，会写入：

- `summary.json`
- 每个任务一个输出文件（自动清理文件名）

```json
{ "output_dir": "./parallel-results" }
```

## 示例：多视角评审

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "评审架构"},
    {"backend": "codex", "prompt": "评审性能"},
    {"backend": "gemini", "prompt": "评审安全"}
  ]
}
```
