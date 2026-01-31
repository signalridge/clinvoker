---
title: 贡献指南
description: clinvoker 的开发设置、编码标准和贡献流程。
---

# 贡献指南

感谢您有兴趣为 clinvoker 做贡献！本指南涵盖开发环境设置、项目结构、编码标准、测试要求和贡献流程。

## 开发环境设置

### 前置要求

- Go 1.24 或更高版本
- Git
- （可选）Nix 用于可复现环境
- （可选）golangci-lint 用于代码检查
- （可选）pre-commit 用于 git 钩子

### Fork 和克隆

```bash
# 在 GitHub 上 fork 仓库，然后：
git clone https://github.com/YOUR_USERNAME/clinvoker.git
cd clinvoker
git remote add upstream https://github.com/signalridge/clinvoker.git
```

### 使用 Nix（推荐）

```bash
nix develop
```

这会在可复现环境中提供所有必需的工具。

### 手动设置

```bash
go mod download
go build ./cmd/clinvk
./clinvk version
```

### 预提交钩子

```bash
pre-commit install
pre-commit install --hook-type commit-msg
```

## 项目结构

```text
clinvoker/
├── cmd/
│   └── clinvk/           # 主应用入口点
│       └── main.go
├── internal/
│   ├── app/              # CLI 命令实现
│   ├── backend/          # 后端抽象层
│   ├── config/           # 配置管理
│   ├── executor/         # 命令执行
│   ├── output/           # 输出格式化
│   ├── server/           # HTTP API 服务器
│   ├── session/          # 会话管理
│   ├── auth/             # API 密钥管理
│   ├── metrics/          # Prometheus 指标
│   └── resilience/       # 熔断器
├── docs/                 # 文档
├── scripts/              # 构建和实用脚本
└── test/                 # 集成测试
```

### 包职责

| 包 | 用途 | 关键文件 |
|---------|---------|-----------|
| `app/` | 使用 Cobra 的 CLI 命令 | `app.go`, `cmd_*.go` |
| `backend/` | 后端抽象 | `backend.go`, `registry.go`, `claude.go`, `codex.go`, `gemini.go` |
| `config/` | 基于 Viper 的配置 | `config.go`, `validate.go` |
| `executor/` | 子进程执行 | `executor.go`, `signal.go` |
| `server/` | HTTP 服务器 | `server.go`, `routes.go`, `handlers/`, `middleware/` |
| `session/` | 会话持久化 | `session.go`, `store.go`, `filelock.go` |

## 编码标准

### Go 指南

遵循 [Effective Go](https://golang.org/doc/effective_go.html) 和 [Google Go 风格指南](https://google.github.io/styleguide/go/)：

1. **格式化**：使用 `gofmt` 或 `goimports`
2. **代码检查**：提交前运行 `golangci-lint run`
3. **命名**：使用描述性、惯用的名称
4. **注释**：为所有导出的类型和函数添加文档
5. **错误处理**：用上下文包装错误，避免裸返回

### 代码风格示例

```go
// 良好：清晰的函数名、适当的文档、错误处理
// ExecuteCommand 运行给定的命令并返回输出。
func ExecuteCommand(ctx context.Context, cfg *Config, cmd *exec.Cmd) (*Result, error) {
    if cfg == nil {
        return nil, fmt.Errorf("config is required")
    }

    // 实现
    result, err := runWithTimeout(ctx, cmd, cfg.Timeout)
    if err != nil {
        return nil, fmt.Errorf("failed to execute command: %w", err)
    }

    return result, nil
}

// 不良：不清晰的名称、缺少文档、错误的错误处理
func exec(cfg *Config, c *exec.Cmd) (*Result, error) {
    res, _ := runWithTimeout(context.Background(), c, cfg.Timeout)
    return res, nil
}
```

### 错误处理

使用项目的错误包进行一致的错误处理：

```go
import apperrors "github.com/signalridge/clinvoker/internal/errors"

// 创建带上下文的错误
return apperrors.BackendError("claude", err)

// 检查错误类型
if apperrors.IsCode(err, apperrors.ErrCodeBackendUnavailable) {
    // 处理特定错误
}
```

### 测试标准

所有代码都必须有测试。遵循以下指南：

1. **文件命名**：与源文件一起的 `*_test.go`
2. **表驱动测试**：用于多个测试用例
3. **并行测试**：对独立测试使用 `t.Parallel()`
4. **覆盖率**：新代码目标 >80% 覆盖率
5. **模拟**：使用接口实现可测试性

```go
func TestExecuteCommand(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        cfg     *Config
        cmd     *exec.Cmd
        wantErr bool
    }{
        {
            name:    "valid command",
            cfg:     &Config{Timeout: 30 * time.Second},
            cmd:     exec.Command("echo", "hello"),
            wantErr: false,
        },
        {
            name:    "nil config",
            cfg:     nil,
            cmd:     exec.Command("echo", "hello"),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            ctx := context.Background()
            _, err := ExecuteCommand(ctx, tt.cfg, tt.cmd)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 测试要求

### 运行测试

```bash
# 所有测试
go test ./...

# 带竞态检测
go test -race ./...

# 带覆盖率
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt -o coverage.html

# 仅短测试
go test -short ./...

# 特定包
go test ./internal/backend/...
```

### 集成测试

集成测试需要安装实际的后端 CLI：

```bash
# 运行集成测试
CLINVK_TEST_INTEGRATION=1 go test ./test/...

# 使用特定后端运行
CLINVK_TEST_BACKEND=claude go test ./test/...
```

### 基准测试

```bash
# 运行基准测试
go test -bench=. ./...

# 带内存分析运行
go test -bench=. -benchmem ./...
```

## 文档要求

### 代码注释

为所有导出的类型和函数添加文档：

```go
// Backend 表示一个 AI CLI 后端。
type Backend interface {
    // Name 返回后端标识符。
    Name() string

    // IsAvailable 检查后端 CLI 是否已安装。
    IsAvailable() bool
}

// NewStore 在给定目录创建新的会话存储。
// 如果目录不存在则创建它。
func NewStore(dir string) (*Store, error) {
    // ...
}
```

### 用户文档

为用户可见的更改更新文档：

1. **概念**：如果设计更改则更新架构文档
2. **指南**：为新功能添加/更新操作指南
3. **参考**：为更改更新 API/CLI 参考
4. **变更日志**：添加条目到 CHANGELOG.md

### 文档风格

- 使用清晰、简洁的语言
- 包含代码示例
- 为复杂概念添加图表（Mermaid）
- 保持中文和英文版本同步

## PR 流程

### 分支命名

使用约定的分支名称：

| 前缀 | 用途 | 示例 |
|--------|---------|---------|
| `feat/` | 新功能 | `feat/add-gemini-backend` |
| `fix/` | Bug 修复 | `fix/session-locking` |
| `docs/` | 文档 | `docs/api-examples` |
| `refactor/` | 代码重构 | `refactor/executor` |
| `test/` | 测试添加 | `test/backend-coverage` |
| `chore/` | 维护 | `chore/update-deps` |

### 提交消息约定

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```text
type(scope): description

[可选正文]

[可选页脚]
```

类型：`feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

示例：

```text
feat(backend): add support for new AI provider

添加对新 XYZ AI CLI 工具的支持。这包括：
- 后端实现
- 标志映射
- 文档更新

fix(session): handle concurrent access correctly

修复当多个进程尝试同时更新同一会话时
会话存储中的竞态条件。

docs(readme): update installation instructions

test(backend): add unit tests for registry

chore(deps): update golang.org/x packages
```

### Pull Request 模板

```markdown
## 摘要

更改的简要描述。

## 更改

- 更改 1
- 更改 2

## 测试计划

如何测试更改：
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 执行了手动测试

## 相关 Issues

Fixes #123
Closes #456
```

### 代码审查清单

提交 PR 前：

- [ ] 所有测试在本地通过
- [ ] 代码遵循风格指南
- [ ] 文档已更新
- [ ] 提交消息遵循约定
- [ ] 没有包含不必要的文件
- [ ] CHANGELOG.md 已更新（如适用）

### 审查流程

1. PR 需要至少一个批准
2. CI 必须在合并前通过
3. 及时解决审查意见
4. 保持 PR 专注且大小适中（首选 <500 行）
5. 如果有冲突则变基到 main

## 发布流程

### 版本控制

clinvoker 遵循 [语义化版本](https://semver.org/)：

- `MAJOR`：破坏性更改
- `MINOR`：新功能，向后兼容
- `PATCH`：错误修复，向后兼容

### 创建发布

1. 更新 `internal/version/version.go` 中的版本
2. 更新 CHANGELOG.md
3. 创建 git 标签：`git tag v1.2.3`
4. 推送标签：`git push origin v1.2.3`
5. GitHub Actions 构建并发布

## 问题？

- 为 bug 或功能开 issue
- 为问题开讨论
- 创建新 issue 前先检查现有的

## 行为准则

参与即表示您同意：

1. 尊重和包容
2. 优雅地接受建设性批评
3. 专注于对社区最有利的事情
4. 对他人表示同理心

## 相关文档

- [架构概述](architecture.zh.md) - 系统架构
- [设计决策](design-decisions.zh.md) - 架构决策记录
- [故障排除](troubleshooting.zh.md) - 常见问题
