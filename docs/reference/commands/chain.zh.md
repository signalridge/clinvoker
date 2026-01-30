# clinvk chain

按顺序执行一组 prompt。

## 语法

```bash
clinvk chain --file chain.json [--json]
cat chain.json | clinvk chain
```

## 参数

| 参数 | 简写 | 默认值 | 说明 |
|------|-------|---------|-------------|
| `--file` | `-f` | | 链式定义文件（JSON） |
| `--input` | | | `--file` 的废弃别名 |
| `--json` | | false | 输出 JSON 汇总 |

## 链式格式

```json
{
  "steps": [
    {"backend": "claude", "prompt": "分析"},
    {"backend": "codex", "prompt": "修复：{{previous}}"}
  ],
  "stop_on_failure": true,
  "pass_working_dir": false
}
```

## 说明

- CLI chain **始终无状态**（不保存会话）。
- 仅支持 `{{previous}}` 占位符。
- 当前 CLI 行为为遇错即停。

## 退出码

- `0` 全部成功
- `1` 存在失败
