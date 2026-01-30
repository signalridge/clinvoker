# 链式执行

按顺序执行多个步骤，并把上一步输出传给下一步。

## 基本用法

```bash
clinvk chain --file chain.json
```

或从 stdin 读取：

```bash
cat chain.json | clinvk chain
```

## 链式格式

```json
{
  "steps": [
    {"name": "analyze", "backend": "claude", "prompt": "分析这段代码"},
    {"name": "fix", "backend": "codex", "prompt": "根据以下分析修复：{{previous}}"},
    {"name": "verify", "backend": "gemini", "prompt": "复核修改：{{previous}}"}
  ]
}
```

### 占位符

- `{{previous}}` → 上一步的**文本输出**

## Step 字段（CLI）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 否 | 步骤名，用于输出展示 |
| `backend` | string | 是 | `claude` / `codex` / `gemini` |
| `prompt` | string | 是 | 支持 `{{previous}}` |
| `model` | string | 否 | 模型覆盖 |
| `workdir` | string | 否 | 工作目录 |
| `approval_mode` | string | 否 | 尽力映射 |
| `sandbox_mode` | string | 否 | 尽力映射 |
| `max_turns` | int | 否 | 仅 Claude 支持 |

## 顶层选项

```json
{
  "stop_on_failure": true,
  "pass_working_dir": false
}
```

说明：

- CLI 的 chain **当前总是遇错即停**，即使 `stop_on_failure` 设置为 `false`。
- `pass_working_dir` 为 true 时，下一步会继承上一步的 `workdir`。

## 输出

```bash
clinvk chain --file chain.json --json
```

会输出结构化 JSON 总结。

## 仅无状态

chain 始终无状态，不会持久化会话。
