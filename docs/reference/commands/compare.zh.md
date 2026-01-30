# clinvk compare

在多个后端上对比同一条 prompt。

## 语法

```bash
clinvk compare <prompt> --backends claude,gemini
clinvk compare <prompt> --all-backends
```

## 参数

| 参数 | 默认值 | 说明 |
|------|---------|-------------|
| `--backends` | | 以逗号分隔的后端列表 |
| `--all-backends` | false | 使用全部后端 |
| `--json` | false | 输出 JSON |
| `--sequential` | false | 串行执行 |

## 说明

- compare **始终无状态**。
- 未安装的后端会提示并跳过。
- 必须使用 `--backends` 或 `--all-backends`。

## 退出码

- `0` 全部成功
- `1` 有后端失败
