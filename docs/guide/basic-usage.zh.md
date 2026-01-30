# 基本用法

学习 clinvk 的日常使用基础。

## 运行提示

使用 clinvk 最简单的方式是用默认后端运行提示：

```bash
clinvk "您的提示"
```bash

### 指定后端

使用 `--backend`（或 `-b`）参数选择特定后端：

```bash
clinvk --backend claude "修复 auth.go 中的 bug"
clinvk -b codex "实现用户注册"
clinvk -b gemini "解释这个算法"
```bash

### 指定模型

使用 `--model`（或 `-m`）覆盖默认模型：

```bash
clinvk --model claude-opus-4-5-20251101 "复杂任务"
clinvk -b codex -m o3 "实现功能"
```bash

### 工作目录

设置 AI 操作的工作目录：

```bash
clinvk --workdir /path/to/project "审查代码库"
clinvk -w ./subproject "修复测试"
```

## 输出格式

控制输出的显示方式：

### 文本（默认）

```bash
clinvk "解释这段代码"
```bash

### JSON

```bash
clinvk --output-format json "解释这段代码"
```bash

### 流式 JSON

```bash
clinvk -o stream-json "解释这段代码"
```bash

## 继续对话

### 快速继续

使用 `--continue`（或 `-c`）继续上一个会话：

```bash
clinvk "实现登录功能"
clinvk -c "现在添加密码验证"
clinvk -c "添加速率限制"
```

### Resume 命令

使用 `resume` 命令获得更多控制：

```bash
# 恢复上一个会话
clinvk resume --last

# 交互式会话选择器
clinvk resume --interactive

# 恢复并带上特定提示
clinvk resume --last "从上次中断的地方继续"
```bash

详见 [会话管理](session-management.md)。

## 试运行模式

预览命令而不执行：

```bash
clinvk --dry-run "实现功能 X"
```yaml

输出显示将要运行的确切命令：

```yaml
Would execute: claude --model claude-opus-4-5-20251101 "实现功能 X"
```

## 临时模式

在无状态模式下运行，不创建会话：

```bash
clinvk --ephemeral "2+2 等于多少"
```bash

这对于不需要对话历史的快速一次性查询很有用。

## 全局参数汇总

| 参数 | 简写 | 描述 | 默认值 |
|------|------|------|--------|
| `--backend` | `-b` | 使用的 AI 后端 | `claude` |
| `--model` | `-m` | 使用的模型 | (后端默认) |
| `--workdir` | `-w` | 工作目录 | (当前目录) |
| `--output-format` | `-o` | 输出格式 | `json` |
| `--continue` | `-c` | 继续上一个会话 | `false` |
| `--dry-run` | | 只显示命令 | `false` |
| `--ephemeral` | | 无状态模式 | `false` |
| `--config` | | 配置文件路径 | `~/.clinvk/config.yaml` |

## 示例

### 快速修复 Bug

```bash
clinvk "utils.go 第 45 行有空指针异常"
```bash

### 代码生成

```bash
clinvk -b codex "生成用户 CRUD 操作的 REST API 处理器"
```bash

### 代码解释

```bash
clinvk -b gemini "解释 cmd/server/main.go 中的 main 函数是做什么的"
```

### 重构

```bash
clinvk "重构数据库模块以使用连接池"
clinvk -c "现在为更改添加单元测试"
```

## 下一步

- [会话管理](session-management.md) - 有效地使用会话
- [后端对比](backend-comparison.md) - 获取多个视角
- [配置](../reference/configuration.md) - 自定义设置
