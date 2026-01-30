# 贡献指南

感谢您有兴趣为 clinvk 做贡献！

## 行为准则

参与项目即表示您同意维护一个尊重和包容的环境。

## 开始

### 前置要求

- Go 1.24 或更高版本
- Git
- （可选）Nix 用于可复现环境
- （可选）golangci-lint 用于代码检查

### Fork 和克隆

```bash
# 在 GitHub 上 fork 仓库，然后：
git clone https://github.com/YOUR_USERNAME/clinvoker.git
cd clinvoker
git remote add upstream https://github.com/signalridge/clinvoker.git
```

## 开发设置

### 使用 Nix（推荐）

```bash
nix develop
```

### 手动设置

```bash
go mod download
go build ./cmd/clinvk
./clinvk version
```

## 进行更改

### 分支命名

使用约定的分支名称：

| 前缀 | 用途 |
|------|------|
| `feat/` | 新功能 |
| `fix/` | Bug 修复 |
| `docs/` | 文档 |
| `refactor/` | 代码重构 |
| `test/` | 测试添加 |
| `chore/` | 维护 |

### 提交消息

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```text
type(scope): description

[可选正文]

[可选页脚]

```

类型：`feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

## 测试

### 运行测试

```bash
# 所有测试
go test ./...

# 带竞态检测
go test -race ./...

# 带覆盖率
go test -coverprofile=coverage.txt ./...
```

### 编写测试

- 将测试放在 `*_test.go` 文件中
- 使用表驱动测试
- 测试成功和错误路径
- 对独立测试使用 `t.Parallel()`

## 代码风格

### Go 指南

- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 提交前运行 `golangci-lint run`
- 保持函数专注且大小适中
- 使用有意义的名称
- 为导出的函数添加文档

## 提交更改

### Pull Request 流程

1. 确保本地所有测试通过
2. 如需要更新文档
3. 创建带有清晰描述的 PR
4. 引用相关 issues

### PR 模板

```markdown
## 摘要

更改的简要描述。

## 更改

- 更改 1
- 更改 2

## 测试计划

如何测试更改。

## 相关 Issues

Fixes #123
```

## 问题？

- 为 bug 或功能开 issue
- 为问题开讨论
- 创建新 issue 前先检查现有的
