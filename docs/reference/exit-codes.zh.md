# 退出码

clinvk 退出码及其含义参考。

## 退出码汇总

| 代码 | 名称 | 描述 |
|------|------|------|
| 0 | 成功 | 命令成功完成 |
| 1 | 错误 | CLI/校验错误或子命令失败 |
|（后端）| 后端退出码 | 对 `clinvk [prompt]` 与 `clinvk resume`，会透传后端 CLI 的非零退出码 |

## 详细说明

### 0 - 成功

命令成功完成。

```bash
clinvk "hello world"
echo $?  # 0
```bash

### 1 - 错误

执行期间发生一般错误：

- 后端执行失败
- 无效输入

### 后端退出码（prompt/resume）

当运行 `clinvk [prompt]` 或 `clinvk resume` 时，clinvk 会执行后端 CLI，并在后端进程以非零退出码结束时透传该退出码。

## 命令特定退出码

### parallel

| 代码 | 描述 |
|------|------|
| 0 | 所有任务成功 |
| 1 | 一个或多个任务失败 |

### compare

| 代码 | 描述 |
|------|------|
| 0 | 所有后端成功 |
| 1 | 一个或多个后端失败 |

### chain

| 代码 | 描述 |
|------|------|
| 0 | 所有步骤成功 |
| 1 | 某个步骤失败 |

### serve

| 代码 | 描述 |
|------|------|
| 0 | 正常关闭 (SIGINT/SIGTERM) |
| 1 | 服务器错误 |

## 脚本示例

### 检查成功

```bash
if clinvk "实现功能"; then
  echo "成功"
else
  echo "失败"
fi
```

### 处理特定代码

```bash
clinvk -b codex "提示"
code=$?

case $code in
  0)
    echo "成功"
    ;;
  *)
    echo "错误：$code"
    ;;
esac
```

## 另请参阅

- [命令参考](commands/index.md)
- [故障排除](../development/troubleshooting.md)
