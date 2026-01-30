# 开发指南

为希望贡献或扩展 clinvk 的开发者提供的信息。

## 概述

clinvk 使用 Go 编写，采用模块化架构，便于添加新后端和功能。

## 开始

- **[架构](architecture.md)** - 系统架构和设计决策
- **[贡献](contributing.md)** - 如何为项目做贡献
- **[添加后端](adding-backends.md)** - 实现新后端的指南
- **[测试](testing.md)** - 测试指南和实践

## 快速开始

### 使用 Nix（推荐）

```bash
nix develop
just ci
```bash

### 手动设置

```bash
# 要求：Go 1.24+
go mod download
go build ./cmd/clinvk
./clinvk version
```

## 项目结构

```text
clinvoker/
├── cmd/clinvk/           # 入口点
├── internal/
│   ├── app/              # CLI 命令
│   ├── backend/          # 后端实现
│   ├── config/           # 配置
│   ├── errors/           # 错误类型
│   ├── executor/         # 执行逻辑
│   ├── output/           # 输出解析
│   ├── server/           # HTTP 服务器
│   │   ├── handlers/     # API 处理程序
│   │   └── service/      # 业务逻辑
│   ├── session/          # 会话管理
│   └── mock/             # 测试工具
├── docs/                 # 文档
└── testdata/             # 测试数据
```

## 常用任务

```bash
# 构建
just build

# 测试
just test

# 代码检查
just lint

# 运行所有检查
just ci

# 启动开发服务器
just serve
```text

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.24+ |
| CLI 框架 | Cobra |
| HTTP 服务器 | Huma/v2 |
| 配置 | Viper |
| 测试 | 标准库 + testify |
| 代码检查 | golangci-lint |
| 构建 | Just（任务运行器） |
| CI/CD | GitHub Actions |
| 包管理器 | Nix（可选） |

## 关键概念

### 后端抽象

所有后端实现一个通用接口：

```go
type Backend interface {
    Name() string
    IsAvailable() bool
    BuildCommand(prompt string, opts *Options) *exec.Cmd
    ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd
    ParseOutput(rawOutput string) string
}
```

### 配置级联

设置按优先级顺序解析：

1. CLI 参数
2. 环境变量
3. 配置文件
4. 默认值

### 会话管理

会话是 `~/.clinvk/sessions/` 中的 JSON 文件，包含每个对话的元数据。

## 开发资源

- [Go 文档](https://golang.org/doc/)
- [Cobra CLI 框架](https://github.com/spf13/cobra)
- [Huma API 框架](https://huma.rocks/)
- [Just 任务运行器](https://just.systems/)
