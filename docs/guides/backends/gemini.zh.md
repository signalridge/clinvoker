# Gemini CLI

Google 的 Gemini AI 助手，具有广泛的知识和多模态能力。

## 概述

Gemini CLI 是 Google 的 Gemini AI 模型命令行接口。它擅长：

- 广泛的知识和一般问题
- 文档和解释
- 多模态任务（如果支持）
- 研究和信息收集

## 安装

从 [Google](https://github.com/google/gemini-cli) 安装 Gemini CLI：

```bash
# 验证安装
which gemini
gemini --version
```

## 基本用法

```bash
# 使用 clinvk 调用 Gemini
clinvk --backend gemini "解释这个算法是如何工作的"
clinvk -b gemini "为这个 API 编写文档"
```

## 模型

| 模型 | 描述 |
|------|------|
| `gemini-2.5-pro` | 最新最强大的模型 |
| `gemini-2.5-flash` | 更快，优化速度 |

指定模型：

```bash
clinvk -b gemini -m gemini-2.5-flash "快速解释"
```

## 配置

在 `~/.clinvk/config.yaml` 中配置 Gemini：

```yaml
backends:
  gemini:
    # 默认模型
    model: gemini-2.5-pro

    # 启用/禁用此后端
    enabled: true

    # 额外 CLI 参数
    extra_flags: []
```

### 环境变量

```bash
export CLINVK_GEMINI_MODEL=gemini-2.5-flash
```

## 会话管理

Gemini 使用 `--resume` 进行会话恢复：

```bash
# 使用 clinvk 恢复
clinvk resume --last --backend gemini
clinvk resume <session-id>
```

## 沙箱模式

Gemini 支持沙箱模式以进行受控执行：

```yaml
backends:
  gemini:
    extra_flags:
      - "--sandbox"
```

## 统一选项

这些选项适用于 Gemini：

| 选项 | 描述 |
|------|------|
| `model` | 使用的模型 |
| `max_tokens` | 最大响应 token 数 |
| `max_turns` | 最大代理轮次 |

## 最佳实践

!!! tip "用于解释"
    Gemini 的广泛知识使其非常适合解释概念和提供上下文。

!!! tip "利用于文档"
    使用 Gemini 编写或改进文档，凭借其清晰的解释。

!!! tip "研究任务"
    Gemini 非常适合收集信息和面向研究的查询。

## 使用场景

### 文档

```bash
clinvk -b gemini "为这个模块编写全面的文档"
```

### 解释

```bash
clinvk -b gemini "解释这个微服务的架构"
```

### 研究

```bash
clinvk -b gemini "实现速率限制的最佳实践是什么"
```

### 代码审查

```bash
clinvk -b gemini "审查这段代码并解释潜在问题"
```

## 与其他后端的对比

| 方面 | Gemini | Claude | Codex |
|------|--------|--------|-------|
| 知识广度 | 优秀 | 好 | 好 |
| 代码生成 | 好 | 优秀 | 优秀 |
| 解释 | 优秀 | 优秀 | 好 |
| 速度 | 快 | 中等 | 快 |

## 工作流示例

使用 Gemini 进行研究和文档：

```json
{
  "steps": [
    {
      "name": "research",
      "backend": "gemini",
      "prompt": "研究 Go 中认证的最佳实践"
    },
    {
      "name": "implement",
      "backend": "claude",
      "prompt": "基于以下内容实现认证：{{previous}}"
    },
    {
      "name": "document",
      "backend": "gemini",
      "prompt": "为以下内容编写文档：{{previous}}"
    }
  ]
}
```

## 下一步

- [Claude Code 指南](claude.md)
- [Codex CLI 指南](codex.md)
- [后端对比](../compare.md)
