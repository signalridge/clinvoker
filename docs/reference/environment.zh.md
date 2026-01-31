# 环境变量

clinvk 支持的所有环境变量完整参考。

## 概述

环境变量提供了一种无需修改配置文件即可配置 clinvk 的便捷方式。它们特别适用于：

- CI/CD 流水线
- Docker 容器
- 临时覆盖
- 使用 direnv 的按项目设置

## 变量参考

### 核心变量

| 变量 | 必需 | 描述 | 示例 |
|----------|----------|-------------|---------|
| `CLINVK_BACKEND` | 否 | 默认使用的后端 | `claude`、`codex`、`gemini` |
| `CLINVK_CLAUDE_MODEL` | 否 | Claude 后端的默认模型 | `claude-opus-4-5-20251101` |
| `CLINVK_CODEX_MODEL` | 否 | Codex 后端的默认模型 | `o3`、`o3-mini` |
| `CLINVK_GEMINI_MODEL` | 否 | Gemini 后端的默认模型 | `gemini-2.5-pro` |

### 服务器变量

| 变量 | 必需 | 描述 | 示例 |
|----------|----------|-------------|---------|
| `CLINVK_API_KEYS` | 否 | HTTP 服务器认证的 API Key（逗号分隔） | `key1,key2,key3` |
| `CLINVK_API_KEYS_GOPASS_PATH` | 否 | 用于获取 API Key 的 gopass 路径 | `myproject/api-keys` |
| `CLINVK_CONFIG` | 否 | 自定义配置文件路径 | `/etc/clinvk/config.yaml` |
| `CLINVK_HOME` | 否 | clinvk 数据目录（会话、配置） | `~/.clinvk` |

### 后端 API Key

以下变量直接传递给相应的后端 CLI：

| 变量 | 后端 | 描述 |
|----------|---------|-------------|
| `ANTHROPIC_API_KEY` | Claude | Anthropic API Key |
| `OPENAI_API_KEY` | Codex | OpenAI API Key |
| `GOOGLE_API_KEY` | Gemini | Google API Key |

## 使用示例

### 设置默认后端

```bash
export CLINVK_BACKEND=codex
clinvk "实现功能"  # 使用 codex
```text

### 为每个后端设置模型

```bash
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3-mini

clinvk -b claude "复杂任务"  # 使用 claude-sonnet
clinvk -b codex "快速任务"   # 使用 o3-mini
```text

### 临时覆盖

为单个命令设置变量：

```bash
CLINVK_BACKEND=gemini clinvk "解释这个"
```text

### HTTP 服务器的 API Key

```bash
export CLINVK_API_KEYS="prod-key-1,prod-key-2,dev-key-1"
clinvk serve
```text

客户端必须包含以下头之一：

```bash
# 选项 1：X-Api-Key 头
curl -H "X-Api-Key: prod-key-1" http://localhost:8080/api/v1/prompt \
  -d '{"backend":"claude","prompt":"hello"}'

# 选项 2：Authorization 头
curl -H "Authorization: Bearer prod-key-1" http://localhost:8080/api/v1/prompt \
  -d '{"backend":"claude","prompt":"hello"}'
```text

## 优先级

环境变量在配置层次结构中具有中等优先级：

1. **CLI 参数**（最高优先级）
2. **环境变量**
3. **配置文件**
4. **默认值**（最低优先级）

演示优先级的示例：

```bash
export CLINVK_BACKEND=codex
clinvk -b claude "提示词"  # 使用 claude（CLI 参数优先）
```text

## Shell 配置

### Bash

添加到 `~/.bashrc` 或 `~/.bash_profile`：

```bash
# clinvk 配置
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
export CLINVK_CODEX_MODEL=o3
```text

### Zsh

添加到 `~/.zshrc`：

```zsh
# clinvk 配置
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
export CLINVK_CODEX_MODEL=o3
```text

### Fish

添加到 `~/.config/fish/config.fish`：

```fish
# clinvk 配置
set -gx CLINVK_BACKEND claude
set -gx CLINVK_CLAUDE_MODEL claude-opus-4-5-20251101
set -gx CLINVK_CODEX_MODEL o3
```text

## 按目录配置

使用 [direnv](https://direnv.net/) 进行项目特定设置：

```bash
# 项目根目录的 .envrc
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3
```text

当你进入该目录时，direnv 会自动加载这些变量。

## CI/CD 使用

### GitHub Actions

```yaml
name: AI Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    env:
      CLINVK_BACKEND: codex
      CLINVK_CODEX_MODEL: o3
      OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
    steps:
      - uses: actions/checkout@v4
      - name: Review code
        run: clinvk "review this PR for security issues"
```text

### GitLab CI

```yaml
ai-review:
  image: alpine/clinvk
  variables:
    CLINVK_BACKEND: claude
    CLINVK_CLAUDE_MODEL: claude-sonnet-4-20250514
    ANTHROPIC_API_KEY: $ANTHROPIC_API_KEY
  script:
    - clinvk "review the changes"
```text

### Jenkins

```groovy
pipeline {
    agent any
    environment {
        CLINVK_BACKEND = 'gemini'
        CLINVK_GEMINI_MODEL = 'gemini-2.5-pro'
        GOOGLE_API_KEY = credentials('google-api-key')
    }
    stages {
        stage('AI Analysis') {
            steps {
                sh 'clinvk "analyze code quality"'
            }
        }
    }
}
```text

## Docker 使用

### Dockerfile

```dockerfile
FROM alpine:latest
RUN apk add --no-cache clinvk

ENV CLINVK_BACKEND=claude
ENV CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101

ENTRYPOINT ["clinvk"]
```text

### Docker Run

```bash
docker run -e CLINVK_BACKEND=codex -e OPENAI_API_KEY=$OPENAI_API_KEY clinvk "提示词"
```text

### Docker Compose

```yaml
version: '3'
services:
  ai-task:
    image: clinvk
    environment:
      - CLINVK_BACKEND=claude
      - CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    command: clinvk "analyze codebase"
```text

## Kubernetes 使用

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: clinvk-config
data:
  CLINVK_BACKEND: "claude"
  CLINVK_CLAUDE_MODEL: "claude-opus-4-5-20251101"
---
apiVersion: v1
kind: Secret
metadata:
  name: clinvk-secrets
type: Opaque
stringData:
  ANTHROPIC_API_KEY: "your-api-key"
---
apiVersion: v1
kind: Pod
metadata:
  name: clinvk-job
spec:
  containers:
    - name: clinvk
      image: clinvk:latest
      envFrom:
        - configMapRef:
            name: clinvk-config
        - secretRef:
            name: clinvk-secrets
```bash

## 故障排除

### 变量未生效

1. 确认变量已导出：`export VAR=value` 而不是仅 `VAR=value`
2. 检查变量名是否有拼写错误
3. 确保 CLI 参数没有覆盖你的变量
4. 检查配置文件是否设置了相同的值

### 调试环境

```bash
# 显示所有 CLINVK 变量
env | grep CLINVK

# 显示特定变量
echo $CLINVK_BACKEND

# 使用调试输出运行
CLINVK_DEBUG=1 clinvk "提示词"
```text

## 另请参阅

- [配置参考](configuration.md) - 配置文件选项
- [config 命令](cli/config.md) - 通过 CLI 管理配置
