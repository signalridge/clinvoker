# 基本用法

学习 clinvk 的日常使用基础。本指南详细介绍命令结构、全局参数、后端选择策略和输出格式。

## 命令结构概述

clinvk 遵循一致的命令结构：

```bash
clinvk [全局参数] [子命令] [子命令参数] [参数]
```bash

### 命令类型

| 类型 | 示例 | 描述 |
|------|------|------|
| **直接提示** | `clinvk "修复 bug"` | 使用默认后端运行提示 |
| **子命令** | `clinvk sessions list` | 执行特定命令 |
| **恢复** | `clinvk resume --last` | 恢复之前的会话 |

## 全局参数详解

全局参数影响 clinvk 的运行方式，与具体命令无关。它们可以在任何子命令之前指定。

### 后端选择 (`--backend`, `-b`)

`--backend` 参数决定哪个 AI 后端处理你的提示。

```bash
# 使用特定后端
clinvk --backend claude "修复 auth.go 中的 bug"
clinvk -b codex "实现用户注册"
clinvk -b gemini "解释这个算法"
```bash

**后端选择策略：**

| 任务类型 | 推荐后端 | 原因 |
|----------|----------|------|
| 复杂推理 | `claude` | 深度上下文理解，安全重点 |
| 代码生成 | `codex` | 针对编程任务优化 |
| 文档编写 | `gemini` | 知识广泛，解释清晰 |
| 安全审查 | `claude` | 分析彻底，风险评估 |
| 快速原型 | `codex` | 代码生成速度快 |

未指定后端时，clinvk 使用配置中的 `default_backend`（默认为 `claude`）。

### 模型选择 (`--model`, `-m`)

覆盖所选后端的默认模型：

```bash
clinvk --model claude-opus-4-5-20251101 "复杂的架构任务"
clinvk -b codex -m o3 "实现功能"
clinvk -b gemini -m gemini-2.5-flash "快速问题"
```text

**何时覆盖模型：**

- 使用大模型（Opus、o3、Pro）处理需要深度推理的复杂任务
- 使用小模型（Sonnet、o3-mini、Flash）处理更快、更简单的任务
- 考虑成本和延迟的权衡

### 工作目录 (`--workdir`, `-w`)

设置 AI 操作的工作目录：

```bash
clinvk --workdir /path/to/project "审查代码库"
clinvk -w ./subproject "修复测试"
```text

**工作目录行为：**

- AI 接收指定目录作为其工作上下文
- 文件操作相对于此目录
- 不同后端处理沙箱的方式不同（参见[后端指南](backends/index.md)）
- 在脚本中使用绝对路径以提高清晰度

**安全考虑：**

```bash
# 良好：明确、范围有限
clinvk -w /home/user/projects/myapp "分析代码"

# 风险：完全系统访问（取决于后端沙箱模式）
clinvk -w / "搜索文件"
```text

### 输出格式 (`--output-format`, `-o`)

控制输出显示方式。有效默认值来自配置中的 `output.format`（内置默认值为 `json`）。

#### 文本格式

带格式的人类可读输出：

```bash
clinvk --output-format text "解释这段代码"
```text

**适用于：** 交互式使用、终端阅读、快速检查

#### JSON 格式

用于程序化处理的结构化输出：

```bash
clinvk --output-format json "解释这段代码"
```text

**输出结构：**

```json
{
  "output": "代码实现了...",
  "backend": "claude",
  "model": "claude-opus-4-5-20251101",
  "duration_seconds": 2.5,
  "exit_code": 0
}
```text

**适用于：** 脚本编写、CI/CD 流水线、存储结果、进一步处理

#### 流式 JSON 格式

```bash
clinvk -o stream-json "解释这段代码"
```text

`stream-json` 直接透传后端的原生流式输出（NDJSON/JSONL）。这提供了 AI 生成内容时的实时更新。

**适用于：** 长时间运行的任务、实时监控、构建交互式工具

**格式对比：**

| 格式 | 人类可读 | 机器可解析 | 流式 | 使用场景 |
|------|----------|------------|------|----------|
| `text` | 是 | 否 | 否 | 交互式使用 |
| `json` | 部分 | 是 | 否 | 脚本编写、存储 |
| `stream-json` | 部分 | 是 | 是 | 实时应用 |

### 继续模式 (`--continue`, `-c`)

无需指定会话 ID 即可继续上一个会话：

```bash
clinvk "实现登录功能"
clinvk -c "现在添加密码验证"
clinvk -c "添加速率限制"
```bash

**继续模式的工作原理：**

1. clinvk 查找最近的可恢复会话
2. 将新提示附加到对话历史
3. AI 拥有先前交互的完整上下文

**会话要求：**

- 上一个会话必须具有后端会话 ID
- 使用 `--ephemeral` 创建的会话无法继续
- 只能继续同一后端的会话

### 临时模式 (`--ephemeral`)

在无状态模式下运行，不创建会话：

```bash
clinvk --ephemeral "2+2 等于多少"
```text

**何时使用临时模式：**

| 场景 | 为什么使用临时模式？ |
|------|---------------------|
| 快速一次性查询 | 不需要历史记录 |
| CI/CD 脚本 | 避免会话累积 |
| 测试/调试 | 每次都有干净状态 |
| 公共/共享系统 | 隐私，不保留数据 |
| 高容量自动化 | 减少存储开销 |

**权衡：**

- **优点：** 无存储、执行更快、隐私
- **缺点：** 无对话历史、无法恢复

### 试运行模式 (`--dry-run`)

预览命令而不执行：

```bash
clinvk --dry-run "实现功能 X"
```text

**输出显示将要运行的确切命令：**

```yaml
Would execute: claude --model claude-opus-4-5-20251101 "实现功能 X"
```text

**使用场景：**

- 在运行昂贵操作前验证配置
- 调试参数解析和后端选择
- 记录预期行为
- 在 CI/CD 中测试而不进行实际 API 调用

### 详细模式 (`--verbose`, `-v`)

启用详细日志记录：

```bash
clinvk --verbose "复杂任务"
```bash

**显示内容：**

- 配置加载详情
- 后端检测信息
- 命令构建步骤
- API 调用和响应（取决于后端）

## 退出代码参考

clinvk 使用标准退出代码进行脚本编写：

| 代码 | 含义 | 何时发生 |
|------|------|----------|
| 0 | 成功 | 命令成功完成 |
| 1 | 一般错误 | CLI 错误、验证失败、后端错误 |
| 后端代码 | 透传 | 后端自身的退出代码（用于 prompt/resume） |

**命令特定的退出代码：**

| 命令 | 退出代码 0 | 退出代码 1 |
|------|------------|------------|
| `prompt` | 成功 | 后端错误 |
| `parallel` | 所有任务成功 | 一个或多个任务失败 |
| `compare` | 所有后端成功 | 一个或多个后端失败 |
| `chain` | 所有步骤成功 | 某步骤失败 |
| `serve` | 干净关闭 | 服务器错误 |

**脚本示例：**

```bash
#!/bin/bash

clinvk "实现功能"
exit_code=$?

case $exit_code in
  0)
    echo "成功 - 功能已实现"
    ;;
  1)
    echo "失败 - 检查日志"
    exit 1
    ;;
  *)
    echo "后端返回代码: $exit_code"
    ;;
esac
```text

## 环境变量

使用环境变量覆盖配置：

```bash
# 设置默认后端
export CLINVK_BACKEND=codex

# 为每个后端设置模型
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3-mini
export CLINVK_GEMINI_MODEL=gemini-2.5-flash

# 使用环境设置运行
clinvk "提示"  # 使用 codex 和 o3-mini
```bash

**优先级顺序**（从高到低）：

1. CLI 参数 (`--backend codex`)
2. 环境变量 (`CLINVK_BACKEND`)
3. 配置文件 (`~/.clinvk/config.yaml`)
4. 内置默认值

## 继续对话

### 快速继续

使用 `--continue`（或 `-c`）继续上一个会话：

```bash
clinvk "实现登录功能"
clinvk -c "现在添加密码验证"
clinvk -c "添加速率限制"
```text

### Resume 命令

使用 `resume` 命令获得更多控制：

```bash
# 恢复上一个会话
clinvk resume --last

# 交互式会话选择器
clinvk resume --interactive

# 恢复并带上特定提示
clinvk resume --last "从上次中断的地方继续"

# 按 ID 恢复
clinvk resume abc123 "添加测试"
```bash

详见 [会话管理](sessions.md)。

## 全局参数汇总

| 参数 | 简写 | 描述 | 默认值 |
|------|------|------|--------|
| `--backend` | `-b` | 使用的 AI 后端 | `claude` |
| `--model` | `-m` | 使用的模型 | (后端默认) |
| `--workdir` | `-w` | 工作目录 | (当前目录) |
| `--output-format` | `-o` | 输出格式 | `json`（可配置） |
| `--continue` | `-c` | 继续上一个会话 | `false` |
| `--dry-run` | | 只显示命令 | `false` |
| `--ephemeral` | | 无状态模式 | `false` |
| `--verbose` | `-v` | 启用详细日志 | `false` |
| `--config` | | 配置文件路径 | `~/.clinvk/config.yaml` |

## 示例

### 快速修复 Bug

```bash
clinvk "utils.go 第 45 行有空指针异常"
```text

### 代码生成

```bash
clinvk -b codex "生成用户 CRUD 操作的 REST API 处理器"
```text

### 代码解释

```bash
clinvk -b gemini "解释 cmd/server/main.go 中的 main 函数是做什么的"
```text

### 带继续的重构

```bash
clinvk "重构数据库模块以使用连接池"
clinvk -c "现在为更改添加单元测试"
clinvk -c "更新文档"
```text

### CI/CD 集成

```bash
# 非交互模式，JSON 输出
clinvk --ephemeral --output-format json \
  --backend codex \
  "为 auth 模块生成测试"
```text

### 多步骤工作流

```bash
#!/bin/bash

# 步骤 1：分析
clinvk -o json "分析代码库架构" > analysis.json

# 步骤 2：基于分析生成
clinvk -c "实现推荐的更改"

# 步骤 3：验证
clinvk -c "运行测试并修复任何失败"
```text

## 常见模式

### 模式 1：文本探索，JSON 自动化

```bash
# 交互式探索 - 使用文本
clinvk -o text "解释这个模块"

# 满意后，切换到 JSON 进行自动化
clinvk -o json --ephemeral "生成实现"
```text

### 模式 2：按任务类型选择后端

```bash
# 架构决策 - Claude
clinvk -b claude "设计 API 结构"

# 实现 - Codex
clinvk -b codex "实现端点"

# 文档 - Gemini
clinvk -b gemini "编写 API 文档"
```text

### 模式 3：执行前试运行

```bash
# 验证将要发生什么
clinvk --dry-run --backend codex "重构整个代码库"

# 如果满意，实际运行
clinvk --backend codex "重构整个代码库"
```text

## 故障排查

### 后端未找到

```bash
# 检查可用后端
clinvk config show | grep available

# 验证 CLI 安装
which claude codex gemini
```text

### 配置未生效

```bash
# 检查有效配置
clinvk config show

# 验证文件存在
ls -la ~/.clinvk/config.yaml
```text

### 会话未恢复

```bash
# 列出可用会话
clinvk sessions list

# 检查会话是否有后端 ID
clinvk sessions show <session-id>
```text

## 下一步

- [会话管理](sessions.md) - 有效地使用会话
- [后端对比](compare.md) - 获取多个视角
- [配置](../reference/configuration.md) - 自定义设置
- [并行执行](parallel.md) - 并发运行多个任务
