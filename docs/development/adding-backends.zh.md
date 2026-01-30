# 添加新后端

为 clinvk 实现新 AI 后端的指南。

## 概述

clinvk 的后端系统设计为可扩展的。每个后端都是一个封装外部 CLI 工具的 Go 实现。

## 后端接口

所有后端必须实现 `Backend` 接口：

```go
type Backend interface {
    Name() string
    IsAvailable() bool
    BuildCommand(prompt string, opts *Options) *exec.Cmd
    ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd
    ParseOutput(rawOutput string) string
}
```

## 逐步指南

### 1. 创建后端文件

创建 `internal/backend/mybackend.go`：

```go
package backend

import "os/exec"

type MyBackend struct{}

func NewMyBackend() *MyBackend {
    return &MyBackend{}
}

func (b *MyBackend) Name() string {
    return "mybackend"
}

func (b *MyBackend) IsAvailable() bool {
    _, err := exec.LookPath("myai")
    return err == nil
}
```

### 2. 实现命令构建

```go
func (b *MyBackend) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{}

    if opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    return exec.Command("myai", args...)
}
```

### 3. 注册后端

编辑 `internal/backend/registry.go`：

```go
func init() {
    Register("mybackend", NewMyBackend())
}
```

### 4. 添加配置

编辑 `internal/config/config.go`：

```go
type BackendsConfig struct {
    // ...
    MyBackend BackendConfig `mapstructure:"mybackend"`
}
```

### 5. 编写测试

创建 `internal/backend/mybackend_test.go`。

### 6. 更新文档

添加文档：

- `docs/user-guide/backends/mybackend.md`
- 更新 `docs/user-guide/backends/index.md`
- 添加到 `config.example.yaml`

## 检查清单

- [ ] 实现所有 `Backend` 接口方法
- [ ] 在 `registry.go` 中注册后端
- [ ] 添加配置支持
- [ ] 添加环境变量绑定
- [ ] 编写单元测试
- [ ] 添加用户文档
- [ ] 更新 `config.example.yaml`
- [ ] 使用真实后端 CLI 测试
