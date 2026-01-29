# 开发指南

为希望贡献或扩展 clinvk 的开发者提供的信息。

## 概述

clinvk 使用 Go 编写，采用模块化架构，便于添加新后端和功能。

## 开始

<div class="grid cards" markdown>

-   :material-cog:{ .lg .middle } **[架构](architecture.md)**

    ---

    系统架构和设计决策

-   :material-account-group:{ .lg .middle } **[贡献](contributing.md)**

    ---

    如何为项目做贡献

-   :material-plus-box:{ .lg .middle } **[添加后端](adding-backends.md)**

    ---

    实现新后端的指南

-   :material-test-tube:{ .lg .middle } **[测试](testing.md)**

    ---

    测试指南和实践

</div>

## 快速开始

### 使用 Nix（推荐）

```bash
nix develop
just ci
```

### 手动设置

```bash
# 要求：Go 1.24+
go mod download
go build ./cmd/clinvk
./clinvk version
```

## 项目结构

```
clinvoker/
├── cmd/clinvk/           # 入口点
├── internal/
│   ├── app/              # CLI 命令
│   ├── backend/          # 后端实现
│   ├── config/           # 配置
│   ├── executor/         # 执行逻辑
│   ├── server/           # HTTP 服务器
│   └── session/          # 会话管理
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
```

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
