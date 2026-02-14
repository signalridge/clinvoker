# 退出码

基于当前实现行为的 `clinvk` 退出码参考。

## 概述

`clinvk` 使用进程退出码来表示执行结果，便于脚本与自动化处理。

## 可保证的退出码契约

| 代码 | 名称 | 描述 |
|------|------|------|
| 0 | 成功 | 命令成功完成 |
| 1 | 一般错误 | 校验 / 配置 / 运行时错误，或聚合子命令失败 |
| 6 | 超时 | 命令超过 `unified_flags.command_timeout_secs` 且超时错误冒泡到顶层 |
| 2+ | 后端退出码 | 在 `prompt` / `resume` 路径透传后端 CLI 退出码 |

## 各命令行为

### prompt / resume

| 代码 | 含义 |
|------|------|
| 0 | 成功 |
| 1 | 在后端退出码透传前发生的一般错误 |
| 6 | 超时（`command_timeout_secs`） |
| 2+ | 透传后端 CLI 退出码 |

### compare / chain / parallel

| 代码 | 含义 |
|------|------|
| 0 | 所有单元执行成功 |
| 1 | 一个或多个单元失败（包含单元内超时） |

### sessions / config / serve / 其他工具命令

| 代码 | 含义 |
|------|------|
| 0 | 命令成功 |
| 1 | 命令失败 |

## 超时语义

当 `unified_flags.command_timeout_secs` 配置为正数时：

- 后端进程超时会被识别为超时错误。
- 顶层超时错误会规范化为退出码 `6`。
- 对于聚合命令（`compare` / `chain` / `parallel`），单元超时按单元失败处理，最终聚合为退出码 `1`。

示例：

```bash
# config.yaml
unified_flags:
  command_timeout_secs: 5

clinvk "长时间任务"
echo $?  # 在顶层 prompt/resume 路径超时时为 6
```

## 脚本示例

### 按退出码分支

```bash
clinvk "任务"
code=$?

case $code in
  0)
    echo "成功"
    ;;
  6)
    echo "超时"
    ;;
  *)
    echo "失败，退出码: $code"
    ;;
esac
```

### Shell 快速失败

```bash
set -e

clinvk "步骤 1"
clinvk "步骤 2"
```

## 说明

- 请不要依赖本文件未列出的历史退出码映射。
- 如果 CI 依赖特定退出码行为，请以本文件契约为准。

## 另请参见

- [命令参考](cli/index.zh.md)
- [配置参考](configuration.zh.md)
