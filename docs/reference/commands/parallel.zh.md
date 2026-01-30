# clinvk parallel

从 JSON 文件或 stdin 并行执行多个任务。

## 语法

```bash
clinvk parallel --file tasks.json [--max-parallel N] [--fail-fast] [--json] [--quiet]
cat tasks.json | clinvk parallel
```

## 参数

| 参数 | 简写 | 默认值 | 说明 |
|------|-------|---------|-------------|
| `--file` | `-f` | | 任务文件（JSON） |
| `--max-parallel` | | 3 | 最大并发数 |
| `--fail-fast` | | false | 首错即停 |
| `--json` | | false | 输出 JSON 汇总 |
| `--quiet` | `-q` | false | 不输出任务内容 |

## 输入格式

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "评审 auth 模块"}
  ],
  "max_parallel": 3,
  "fail_fast": false,
  "output_dir": "./parallel-results"
}
```

### Task 字段（CLI）

- `backend`（必填）
- `prompt`（必填）
- `workdir`, `model`, `approval_mode`, `sandbox_mode`, `max_turns`, `max_tokens`, `system_prompt`, `verbose`, `dry_run`, `extra`
- `id`, `name`, `tags`, `meta`

说明：

- 并行执行 **始终无状态**（不保存会话）。
- CLI 模式下 `output_format` 当前被忽略。
- `parallel.aggregate_output=false` 会隐藏汇总表格。

## 输出文件

若设置 `output_dir`，会写入 `summary.json` 与每个任务的输出文件。

## 退出码

- `0` 全部成功
- `1` 存在失败
