# clinvk compare

对比多个后端的回答。

## 用法

```bash
clinvk compare <prompt> [flags]
```text

## 说明

将同一提示词发送给多个后端进行对比。CLI 的 compare 始终为无状态执行（不持久化会话）。

## 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--backends` | string | | 逗号分隔的后端列表 |
| `--all-backends` | bool | `false` | 使用所有已注册后端（会跳过未安装的 CLI） |
| `--sequential` | bool | `false` | 顺序执行 |
| `--json` | bool | `false` | JSON 输出 |

## 示例

### 指定后端

```bash
clinvk compare --backends claude,codex "explain this code"
```text

### 比较所有后端

```bash
clinvk compare --all-backends "what does this function do"
```text

### 顺序执行

```bash
clinvk compare --all-backends --sequential "review this PR"
```text

### JSON 输出

```bash
clinvk compare --all-backends --json "analyze performance"
```text

## 输出

### 文本输出

```text
Comparing 3 backends: claude, codex, gemini
Prompt: explain this algorithm
================================================================================
[claude] This algorithm implements a binary search...
[codex] The algorithm performs a binary search...
[gemini] This is a classic binary search implementation...

================================================================================
COMPARISON SUMMARY
================================================================================
BACKEND      STATUS     DURATION     MODEL
--------------------------------------------------------------------------------
claude       OK         2.50s        claude-opus-4-5-20251101
codex        OK         3.20s        o3
gemini       OK         2.80s        gemini-2.5-pro
--------------------------------------------------------------------------------
Total time: 3.20s
```text

### JSON 输出

```json
{
  "prompt": "explain this algorithm",
  "backends": ["claude", "codex", "gemini"],
  "results": [
    {
      "backend": "claude",
      "model": "claude-opus-4-5-20251101",
      "output": "This algorithm implements a binary search...",
      "duration_seconds": 2.5,
      "exit_code": 0
    }
  ],
  "total_duration_seconds": 3.2
}
```text

## 错误处理

未安装的后端会被跳过并输出警告。只要有任一后端执行失败，命令将以非 0 退出码结束。

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 所选后端全部成功 |
| 1 | 任一后端失败或无可用后端 |

## 另请参阅

- [parallel](parallel.md) - 不同提示词并行
- [chain](chain.md) - 顺序流水线
