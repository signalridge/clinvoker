# clinvk compare

对比多个后端的响应。

## 概要

```bash
clinvk compare [prompt] [flags]
```

## 描述

将相同的提示发送到多个 AI 后端，并排比较它们的响应。

## 标志

| 标志 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `--backends` | string | | 逗号分隔的后端列表 |
| `--all-backends` | bool | `false` | 对比所有启用的后端 |
| `--sequential` | bool | `false` | 一次运行一个 |
| `--json` | bool | `false` | JSON 输出 |

## 示例

### 对比特定后端

```bash
clinvk compare --backends claude,codex "解释这段代码"
```

### 对比所有后端

```bash
clinvk compare --all-backends "这个函数是做什么的"
```

### 顺序执行

```bash
clinvk compare --all-backends --sequential "审查这个 PR"
```

### JSON 输出

```bash
clinvk compare --all-backends --json "分析性能"
```

## 退出码

| 代码 | 描述 |
|------|------|
| 0 | 至少一个后端成功 |
| 1 | 所有后端失败 |

## 另请参阅

- [parallel](parallel.md) - 不同提示，并发
- [chain](chain.md) - 顺序管道
