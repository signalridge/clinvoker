# 环境变量

clinvk 支持的所有环境变量参考。

## 变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `CLINVK_BACKEND` | 默认后端 | `claude` |
| `CLINVK_CLAUDE_MODEL` | Claude 模型 | `claude-opus-4-5-20251101` |
| `CLINVK_CODEX_MODEL` | Codex 模型 | `o3` |
| `CLINVK_GEMINI_MODEL` | Gemini 模型 | `gemini-2.5-pro` |

## 使用示例

### 设置默认后端

```bash
export CLINVK_BACKEND=codex
clinvk "实现功能"  # 使用 codex
```

### 设置每个后端的模型

```bash
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3-mini

clinvk -b claude "复杂任务"  # 使用 claude-sonnet-4-20250514
clinvk -b codex "快速任务"   # 使用 o3-mini
```

### 临时覆盖

```bash
CLINVK_BACKEND=gemini clinvk "解释这个"
```

## 优先级

环境变量具有中等优先级：

1. **CLI 参数**（最高）
2. **环境变量**
3. **配置文件**
4. **默认值**（最低）

示例：

```bash
export CLINVK_BACKEND=codex
clinvk -b claude "提示"  # 使用 claude（CLI 参数优先）
```

## Shell 配置

### Bash

添加到 `~/.bashrc` 或 `~/.bash_profile`：

```bash
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
```

### Zsh

添加到 `~/.zshrc`：

```zsh
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
```

### Fish

添加到 `~/.config/fish/config.fish`：

```fish
set -gx CLINVK_BACKEND claude
set -gx CLINVK_CLAUDE_MODEL claude-opus-4-5-20251101
```

## 项目级配置

使用 direnv 进行项目特定设置：

```bash
# .envrc
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3
```

## 另请参阅

- [配置参考](configuration.md)
- [config 命令](commands/config.md)
