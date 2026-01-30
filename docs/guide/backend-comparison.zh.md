# 后端对比

在多个后端上运行同一条 prompt 并比较结果。

## 基本用法

```bash
clinvk compare --backends claude,gemini "解释这个算法"
```

使用所有后端：

```bash
clinvk compare --all-backends "评审这个 PR"
```

若某个后端 CLI 未安装，clinvk 会提示并跳过。

## 串行与并行

```bash
clinvk compare --all-backends --sequential "检查安全风险"
```

默认并行执行。

## JSON 输出

```bash
clinvk compare --all-backends --json "总结这个补丁"
```

## 说明

- compare 始终 **无状态**（不保存会话）。
- 任一后端失败会返回非零退出码。
