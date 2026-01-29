# 后端

clinvk 支持多个 AI CLI 后端，每个都有独特的优势和特点。

## 支持的后端

<div class="grid cards" markdown>

-   :material-robot:{ .lg .middle } **[Claude Code](claude.md)**

    ---

    Anthropic 的 AI 编程助手，具有深度推理能力

-   :material-code-tags:{ .lg .middle } **[Codex CLI](codex.md)**

    ---

    OpenAI 的代码专注 CLI 工具，优化用于代码生成

-   :material-google:{ .lg .middle } **[Gemini CLI](gemini.md)**

    ---

    Google 的 Gemini AI，具有广泛的知识和能力

</div>

## 后端对比

| 特性 | Claude Code | Codex CLI | Gemini CLI |
|------|-------------|-----------|------------|
| 二进制文件 | `claude` | `codex` | `gemini` |
| 默认模型 | claude-opus-4-5-20251101 | o3 | gemini-2.5-pro |
| 会话恢复 | `--resume` | `--session` | `-s` |
| 优势 | 复杂推理、安全性 | 代码生成 | 广泛知识 |

## 后端检测

clinvk 通过检查 PATH 中的二进制文件自动检测可用的后端：

```bash
clinvk config show
```

输出显示哪些后端可用：

```yaml
backends:
  claude:
    enabled: true
    available: true  # 'claude' 在 PATH 中找到
  codex:
    enabled: true
    available: false  # 'codex' 未找到
  gemini:
    enabled: true
    available: true  # 'gemini' 在 PATH 中找到
```

## 选择后端

### 通过 CLI

```bash
clinvk --backend claude "提示"
clinvk -b codex "提示"
clinvk -b gemini "提示"
```

### 通过配置

在 `~/.clinvk/config.yaml` 中设置默认后端：

```yaml
default_backend: claude
```

### 通过环境变量

```bash
export CLINVK_BACKEND=codex
clinvk "提示"  # 使用 codex
```

## 后端特定选项

每个后端支持统一选项，以及自己的特定参数：

### 统一选项

这些适用于所有后端：

| 选项 | 描述 |
|------|------|
| `model` | 使用的模型 |
| `approval_mode` | 审批行为 |
| `sandbox_mode` | 文件访问权限 |
| `max_turns` | 最大代理轮次 |
| `max_tokens` | 最大响应 token 数 |

### 后端特定参数

通过配置中的 `extra_flags` 传递额外参数：

```yaml
backends:
  claude:
    extra_flags: ["--add-dir", "./docs"]
  codex:
    extra_flags: ["--quiet"]
  gemini:
    extra_flags: ["--sandbox"]
```

## 选择后端

### 使用 Claude Code 当：

- 处理复杂的多步骤任务
- 需要彻底的代码审查和分析
- 安全性和准确性至关重要

### 使用 Codex CLI 当：

- 生成样板代码
- 编写测试
- 快速代码转换

### 使用 Gemini CLI 当：

- 需要广泛的知识上下文
- 处理文档
- 一般性解释

## 提示

!!! tip "尝试多个后端"
    使用 `clinvk compare --all-backends` 查看不同后端如何处理同一问题。

!!! tip "匹配后端与任务"
    不同后端擅长不同任务。实验找到最适合您工作流程的选择。

!!! tip "配置默认值"
    在配置文件中设置后端特定的模型和选项，获得个性化体验。

## 下一步

- [Claude Code 指南](claude.md)
- [Codex CLI 指南](codex.md)
- [Gemini CLI 指南](gemini.md)
