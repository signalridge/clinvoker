# 配置指南

学习如何为你的工作流配置 clinvk。本指南涵盖常见场景和最佳实践。

## 快速设置

### 1. 查看当前配置

```bash
clinvk config show
```text

这会显示所有设置，包括系统上可用的后端。

### 2. 设置默认后端

```bash
# 使用 Claude 作为默认
clinvk config set default_backend claude

# 或者使用 Gemini
clinvk config set default_backend gemini
```bash

### 3. 完成

基本设置就这些。clinvk 开箱即用，具有合理的默认值。

## 配置文件

clinvk 将配置存储在 `~/.clinvk/config.yaml`。你可以直接编辑它或使用 `clinvk config set`。

### 最小配置

```yaml
# ~/.clinvk/config.yaml
default_backend: claude
```text

### 推荐配置

```yaml
# ~/.clinvk/config.yaml
default_backend: claude

# 在输出中显示执行时间
output:
  show_timing: true

# 保留会话 60 天
session:
  retention_days: 60
  auto_resume: true
```text

## 常见场景

### 场景 1：使用多个后端

如果你针对不同任务使用不同的 AI 模型：

```yaml
default_backend: claude

backends:
  claude:
    model: claude-opus-4-5-20251101    # 用于复杂推理
  codex:
    model: o3                           # 用于代码生成
  gemini:
    model: gemini-2.5-pro              # 用于通用任务
```text

**使用方法：**

```bash
# 使用默认后端（Claude）
clinvk "分析这个架构"

# 为特定任务指定后端
clinvk -b codex "生成单元测试"
clinvk -b gemini "总结这个文档"
```text

### 场景 2：自动化的自动批准模式

对于 CI/CD 或脚本化工作流，不需要交互式提示：

```yaml
unified_flags:
  approval_mode: auto    # 自动批准所有操作
output:
  format: json           # 机器可读输出
```text

!!! warning "安全说明"
    只在受信任的环境中使用 `auto` 批准模式。AI 可以执行文件操作和命令。

### 场景 3：只读分析

用于代码审查或分析，AI 不应修改文件：

```yaml
unified_flags:
  sandbox_mode: read-only    # 不允许文件修改
  # approval_mode 控制“是否询问/如何询问”，并不等同于“允许/拒绝”。
  # 若你更重视安全性，建议使用 `always` 并配合 `sandbox_mode` 限制访问范围。
  approval_mode: always
```text

### 场景 4：团队共享配置

为了团队成员之间保持一致的设置，创建项目级配置：

```bash
# 在项目根目录
cat > .clinvk.yaml << 'EOF'
default_backend: claude
unified_flags:
  sandbox_mode: workspace    # 只访问项目文件
backends:
  claude:
    system_prompt: "你正在开发 MyApp 项目。请遵循我们的编码规范。"
EOF

# 使用项目配置
clinvk --config .clinvk.yaml "审查 auth 模块"
```bash

### 场景 5：用于集成的 HTTP 服务器

将 clinvk 用作 API 后端：

```yaml
server:
  host: "127.0.0.1"          # 仅本地访问（安全）
  port: 8080
  request_timeout_secs: 300  # 长任务 5 分钟超时

# 局域网访问（谨慎使用）
# server:
#   host: "0.0.0.0"          # 所有接口
#   port: 8080
```text

### 场景 6：并行任务优化

用于批处理或多视角审查：

```yaml
parallel:
  max_workers: 5       # 最多同时运行 5 个任务
  fail_fast: false     # 即使部分任务失败也继续
  aggregate_output: true
```text

## 后端特定设置

### Claude Code

```yaml
backends:
  claude:
    model: claude-opus-4-5-20251101
    allowed_tools: all              # 或者: read,write,edit
    system_prompt: "回答要简洁。"
    extra_flags:
      - "--add-dir"
      - "./docs"                    # 将文档包含在上下文中
```text

**可用模型：**

| 模型 | 最适合 |
|------|--------|
| `claude-opus-4-5-20251101` | 复杂推理、架构设计 |
| `claude-sonnet-4-20250514` | 速度和能力平衡 |

### Codex CLI

```yaml
backends:
  codex:
    model: o3
    extra_flags:
      - "--quiet"                   # 减少输出详细程度
```text

### Gemini CLI

```yaml
backends:
  gemini:
    model: gemini-2.5-pro
    extra_flags:
      - "--sandbox"
```text

## 环境变量

使用环境变量覆盖任何配置：

```bash
# 覆盖默认后端
export CLINVK_BACKEND=gemini

# 覆盖模型
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3
export CLINVK_GEMINI_MODEL=gemini-2.5-pro
```bash

**优先级顺序**（从高到低）：

1. CLI 参数（`--backend codex`）
2. 环境变量（例如 `CLINVK_BACKEND`）
3. 配置文件（`~/.clinvk/config.yaml`）
4. 内置默认值

## 最佳实践

### 1. 从默认值开始

clinvk 开箱即用效果很好。只自定义你需要的部分。

### 2. 使用项目级配置

将项目特定设置保存在仓库中的 `.clinvk.yaml`：

```bash
clinvk --config .clinvk.yaml "你的提示"
```text

### 3. 保护你的服务器

如果暴露 HTTP 服务器：

```yaml
server:
  host: "127.0.0.1"    # 没有反向代理时不要使用 0.0.0.0
```text

### 4. 设置适当的超时

对于长时间运行的任务：

```yaml
server:
  request_timeout_secs: 600    # 10 分钟
```text

### 5. 审查时使用只读模式

当你只想分析而不想修改时：

```yaml
unified_flags:
  sandbox_mode: read-only
```text

## 故障排查

### 配置未生效

```bash
# 检查有效配置
clinvk config show

# 验证配置文件位置
ls -la ~/.clinvk/config.yaml
```text

### 后端不可用

```bash
# 检查检测到哪些后端
clinvk config show | grep available

# 验证 CLI 在 PATH 中
which claude codex gemini
```text

### 重置为默认值

```bash
# 删除配置文件
rm ~/.clinvk/config.yaml

# 验证默认值
clinvk config show
```text

## 配置模板

### 开发工作站

```yaml
default_backend: claude

unified_flags:
  sandbox_mode: workspace

output:
  show_timing: true
  color: true

session:
  auto_resume: true
  retention_days: 30
```text

### CI/CD 流水线

```yaml
default_backend: claude

unified_flags:
  approval_mode: auto
output:
  format: json

parallel:
  max_workers: 3
  fail_fast: true
```text

### API 服务器

```yaml
default_backend: claude

server:
  host: "127.0.0.1"
  port: 8080
  request_timeout_secs: 300

output:
  format: json
```text

## 下一步

- [配置参考](../reference/configuration.md) - 完整选项参考
- [环境变量](../reference/environment.md) - 所有环境变量
- [config 命令](../reference/cli/config.md) - CLI 配置命令
