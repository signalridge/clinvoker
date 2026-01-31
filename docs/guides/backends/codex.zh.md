# Codex CLI

OpenAI 的代码专注 CLI 工具，优化用于代码生成和编程任务。

## 概述

Codex CLI 是 OpenAI 的命令行工具，专注于代码生成和编程辅助。它擅长：

- 快速代码生成
- 编写测试和样板
- 代码转换
- 快速编程任务

## 安装

从 [OpenAI](https://github.com/openai/codex-cli) 安装 Codex CLI：

```bash
# 验证安装
which codex
codex --version
```

## 基本用法

```bash
# 使用 clinvk 调用 Codex
clinvk --backend codex "实现一个 REST API 处理器"
clinvk -b codex "为 user.go 生成单元测试"
```

## 模型

| 模型 | 描述 |
|------|------|
| `o3` | 最新最强大的模型 |
| `o3-mini` | 更快、更轻量的模型 |

指定模型：

```bash
clinvk -b codex -m o3-mini "快速代码生成"
```

## 配置

在 `~/.clinvk/config.yaml` 中配置 Codex：

```yaml
backends:
  codex:
    # 默认模型
    model: o3

    # 启用/禁用此后端
    enabled: true

    # 额外 CLI 参数
    extra_flags: []
```

### 环境变量

```bash
export CLINVK_CODEX_MODEL=o3-mini
```

## 会话管理

Codex 使用 `codex exec resume` 进行会话恢复（由 `clinvk` 自动处理）：

```bash
# 使用 clinvk 恢复
clinvk resume --last --backend codex
clinvk resume <session-id>
```

## 统一选项

这些选项适用于 Codex：

| 选项 | 描述 |
|------|------|
| `model` | 使用的模型 |
| `max_tokens` | 最大响应 token 数 |
| `max_turns` | 最大代理轮次 |

## 最佳实践

!!! tip "用于代码生成"
    Codex 优化用于快速生成代码。适合样板和重复任务。

!!! tip "与其他后端结合"
    使用 Codex 生成代码，然后用 Claude 审查 - 利用 chain 命令。

!!! tip "批量相似任务"
    并行运行多个代码生成任务以提高效率。

## 使用场景

### 生成样板

```bash
clinvk -b codex "为 User 模型创建 CRUD API"
```

### 编写测试

```bash
clinvk -b codex "为 auth 模块生成全面的单元测试"
```

### 代码转换

```bash
clinvk -b codex "将这个基于回调的代码转换为 async/await"
```

### 快速实现

```bash
clinvk -b codex "实现一个二分查找函数"
```

## 与 Claude 的对比

| 方面 | Codex | Claude |
|------|-------|--------|
| 速度 | 更快 | 更彻底 |
| 最适合 | 代码生成 | 复杂推理 |
| 上下文 | 好 | 优秀 |
| 安全性重点 | 标准 | 高 |

## 工作流示例

结合使用 Codex 和 Claude：

```json
{
  "steps": [
    {
      "name": "generate",
      "backend": "codex",
      "prompt": "实现用户认证"
    },
    {
      "name": "review",
      "backend": "claude",
      "prompt": "审查这段代码的安全性：{{previous}}"
    }
  ]
}
```

## 下一步

- [Claude Code 指南](claude.md)
- [Gemini CLI 指南](gemini.md)
- [后端对比](../compare.md)
