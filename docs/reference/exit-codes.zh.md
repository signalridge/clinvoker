# 退出码

clinvk 退出码及其含义的完整参考。

## 概述

clinvk 使用退出码来指示命令执行的结果。理解这些代码对于脚本编写和自动化至关重要。

## 退出码参考

| 代码 | 名称 | 描述 | 发生时机 |
|------|------|-------------|----------------|
| 0 | 成功 | 命令成功完成 | 正常完成 |
| 1 | 一般错误 | CLI/校验错误或子命令失败 | 无效输入、执行失败 |
| 2 | 后端不可用 | 请求的后端未安装 | 后端二进制文件未找到 |
| 3 | 配置无效 | 配置文件错误或设置无效 | 错误的配置文件 |
| 4 | 会话错误 | 会话操作失败 | 恢复失败、会话未找到 |
| 5 | API 错误 | HTTP API 请求失败 | 服务器错误、网络问题 |
| 6 | 超时 | 命令执行超时 | 超过超时限制 |
| 7 | 已取消 | 用户取消了操作 | 按下了 Ctrl+C |
| 8+ | 后端退出码 | 从后端 CLI 透传 | 后端特定错误 |

## 详细说明

### 0 - 成功

命令成功完成，没有错误。

```bash
clinvk "hello world"
echo $?  # 输出：0
```

### 1 - 一般错误

执行期间发生一般错误。常见原因包括：

- 无效的命令行参数
- 后端执行失败
- 文件未找到
- 权限被拒绝

```bash
clinvk --invalid-flag "提示词"
echo $?  # 输出：1
```

### 2 - 后端不可用

请求的后端未安装或不在 PATH 中。

```bash
clinvk -b nonexistent "提示词"
echo $?  # 输出：2
```

### 3 - 配置无效

配置文件有错误或包含无效设置。

```bash
clinvk --config /invalid/config.yaml "提示词"
echo $?  # 输出：3
```

### 4 - 会话错误

与会话相关的操作失败。

```bash
clinvk resume nonexistent-session
echo $?  # 输出：4
```

### 5 - API 错误

HTTP API 请求失败（使用 `clinvk serve` 或 API 模式时）。

```bash
# 服务器未运行
clinvk --api-mode "提示词"
echo $?  # 输出：5
```

### 6 - 超时

命令执行超过了配置的超时时间。

```bash
clinvk --timeout 5 "非常长的任务"
echo $?  # 输出：6
```

### 7 - 已取消

用户取消了操作（例如按下了 Ctrl+C）。

```bash
clinvk "长时间运行的任务"
# 按 Ctrl+C
echo $?  # 输出：7
```

### 后端退出码（8+）

运行 `clinvk [prompt]` 或 `clinvk resume` 时，clinvk 执行后端 CLI 并在后端进程以非零退出码结束时透传该退出码。这些代码是后端特定的。

## 命令特定退出码

### prompt / resume

| 代码 | 描述 |
|------|-------------|
| 0 | 成功 |
| 1 | 一般错误 |
| 2+ | 后端退出码（透传） |

### parallel

| 代码 | 描述 |
|------|-------------|
| 0 | 所有任务成功 |
| 1 | 一个或多个任务失败 |
| 2 | 无效的任务文件 |

### compare

| 代码 | 描述 |
|------|-------------|
| 0 | 所有后端成功 |
| 1 | 一个或多个后端失败 |
| 2 | 没有可用的后端 |

### chain

| 代码 | 描述 |
|------|-------------|
| 0 | 所有步骤成功 |
| 1 | 某个步骤失败 |
| 2 | 无效的流水线文件 |

### sessions

| 代码 | 描述 |
|------|-------------|
| 0 | 操作成功 |
| 1 | 操作失败（例如会话未找到） |
| 4 | 会话错误 |

### config

| 代码 | 描述 |
|------|-------------|
| 0 | 操作成功 |
| 1 | 无效的键或值 |
| 3 | 配置错误 |

### serve

| 代码 | 描述 |
|------|-------------|
| 0 | 正常关闭（SIGINT/SIGTERM） |
| 1 | 服务器启动错误 |
| 5 | 操作期间的 API 错误 |

## 脚本示例

### 检查成功

```bash
if clinvk "实现功能"; then
  echo "成功！"
else
  echo "失败！"
fi
```

### 处理特定代码

```bash
clinvk -b codex "提示词"
code=$?

case $code in
  0)
    echo "成功"
    ;;
  1)
    echo "一般错误"
    ;;
  2)
    echo "后端不可用 - 请安装 codex"
    ;;
  4)
    echo "会话错误"
    ;;
  *)
    echo "后端错误：$code"
    ;;
esac
```

### 失败时重试

```bash
max_attempts=3
attempt=1

while [ $attempt -le $max_attempts ]; do
  if clinvk "提示词"; then
    echo "第 $attempt 次尝试成功"
    break
  fi

  if [ $attempt -eq $max_attempts ]; then
    echo "$max_attempts 次尝试后失败"
    exit 1
  fi

  echo "第 $attempt 次尝试失败，5 秒后重试..."
  sleep 5
  attempt=$((attempt + 1))
done
```

### 错误时退出

```bash
#!/bin/bash
set -e  # 任何错误时退出

clinvk "步骤 1"
clinvk "步骤 2"
clinvk "步骤 3"

echo "所有步骤成功完成"
```

### 忽略特定错误

```bash
#!/bin/bash

# 即使失败也继续
clinvk "可选任务" || true

# 这个必须成功
clinvk "关键任务"
```

## CI/CD 集成

### GitHub Actions

```yaml
- name: Run AI task
  run: clinvk "generate tests"
  continue-on-error: true
  id: ai-task

- name: Handle failure
  if: failure() && steps.ai-task.outcome == 'failure'
  run: |
    echo "AI task failed with exit code $?"
    exit 1
```

### GitLab CI

```yaml
ai-task:
  script:
    - clinvk "generate tests" || EXIT_CODE=$?
    - |
      case $EXIT_CODE in
        0) echo "Success" ;;
        2) echo "Backend not installed" ; exit 1 ;;
        *) echo "Error: $EXIT_CODE" ; exit 1 ;;
      esac
```

### Make/Just

```makefile
.PHONY: test lint ai-review

test:
 go test ./...

ai-review:
 clinvk "review the code for issues" || (echo "Review failed" && exit 1)

lint-and-review: lint ai-review
 @echo "All checks passed"
```

## 退出码最佳实践

1. **始终在脚本中检查退出码** 以优雅地处理失败
2. **在 bash 脚本中使用 `set -e`** 以便在错误时立即退出
3. **调试问题时记录退出码**
4. **根据你的需求不同地处理特定代码**
5. **当命令失败不应停止脚本时使用 `|| true`**

## 故障排除

### 意外的退出码

| 症状 | 可能原因 | 解决方案 |
|---------|----------------|----------|
| 总是返回 1 | 后端未配置 | 检查配置和 API Key |
| 返回 2 | 后端未安装 | 安装后端 CLI |
| 返回 4 | 会话过期或无效 | 使用 `clinvk sessions list` 检查会话 |
| 返回 6 | 超时太短 | 增加 `command_timeout_secs` |

### 调试退出码

```bash
# 使用详细输出运行
clinvk -v "提示词"
echo "退出码：$?"

# 直接检查后端
claude "test"
echo "后端退出码：$?"
```

## 另请参阅

- [命令参考](cli/index.md) - 命令文档
- [故障排除](../concepts/troubleshooting.md) - 常见问题和解决方案
