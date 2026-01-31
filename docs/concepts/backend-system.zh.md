---
title: 后端系统
description: 深入解析 clinvoker 的后端抽象层、注册表模式和统一接口。
---

# 后端系统

本文档全面深入介绍 clinvoker 的后端抽象层，解释不同的 AI CLI 工具如何统一在通用接口下、注册表模式的实现、线程安全设计，以及如何扩展系统以支持新后端。

## 后端接口设计

`Backend` 接口（`internal/backend/backend.go:16-46`）是使 clinvoker 能够与多个 AI CLI 工具无缝协作的核心抽象：

```go
type Backend interface {
    Name() string
    IsAvailable() bool
    BuildCommand(prompt string, opts *Options) *exec.Cmd
    ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd
    BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd
    ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd
    ParseOutput(rawOutput string) string
    ParseJSONResponse(rawOutput string) (*UnifiedResponse, error)
    SeparateStderr() bool
}
```

### 接口设计原理

该接口围绕 AI 交互的生命周期设计：

1. **发现**：`Name()` 和 `IsAvailable()` 用于后端识别和检测
2. **命令构建**：`BuildCommand*` 方法创建可执行命令
3. **会话恢复**：`ResumeCommand*` 方法继续现有对话
4. **输出处理**：`ParseOutput()` 和 `ParseJSONResponse()` 规范化响应
5. **错误处理**：`SeparateStderr()` 确定 stderr 处理策略

## 注册表模式

后端注册表（`internal/backend/registry.go`）使用线程安全的注册表模式管理后端注册和查找。

### 注册表结构

```mermaid
flowchart TB
    subgraph 注册表["注册表 (internal/backend/registry.go:11-16)"]
        RWMU[sync.RWMutex]
        BACKENDS[map[string]Backend]
        CACHE[availabilityCache]
        TTL[30秒 TTL]
    end

    subgraph 操作["注册表操作"]
        REGISTER[Register]
        UNREGISTER[Unregister]
        GET[Get]
        LIST[List]
        AVAILABLE[Available]
    end

    RWMU --> BACKENDS
    BACKENDS --> CACHE
    CACHE --> TTL

    REGISTER --> RWMU
    UNREGISTER --> RWMU
    GET --> RWMU
    LIST --> RWMU
    AVAILABLE --> CACHE
```

### 线程安全设计

注册表使用 `sync.RWMutex` 进行并发访问：

```go
// 读操作使用 RLock 进行并发读取
func (r *Registry) Get(name string) (Backend, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // ... 查找逻辑
}

// 写操作使用 Lock 进行独占访问
func (r *Registry) Register(b Backend) {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ... 注册逻辑
    delete(r.availabilityCache, b.Name()) // 使缓存失效
}
```

这种设计允许：
- 多个并发读取器（例如健康检查、列表）
- 独占写入器（例如注册、注销）
- 来自多个 goroutine 的安全并发访问

### 可用性缓存

注册表实现了 30 秒 TTL 缓存用于可用性检查：

```go
type cachedAvailability struct {
    available bool
    checkedAt time.Time
}

func (r *Registry) isAvailableCachedLocked(b Backend) bool {
    name := b.Name()
    if cached, ok := r.availabilityCache[name]; ok &&
       time.Since(cached.checkedAt) < r.availabilityCacheTTL {
        return cached.available
    }

    available := b.IsAvailable()
    r.availabilityCache[name] = &cachedAvailability{
        available: available,
        checkedAt: time.Now(),
    }
    return available
}
```

**30 秒 TTL 的原理**：
- **性能**：避免频繁的 `exec.LookPath()` 调用
- **新鲜度**：30 秒足够短以检测安装变化
- **平衡**：准确性和性能之间的权衡

## 后端实现

### Claude 后端

```go
type Claude struct{}

func (c *Claude) Name() string { return "claude" }

func (c *Claude) IsAvailable() bool {
    _, err := exec.LookPath("claude")
    return err == nil
}

func (c *Claude) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--print"}

    if opts != nil {
        if opts.Model != "" {
            args = append(args, "--model", opts.Model)
        }
        // ... 附加选项
    }

    args = append(args, prompt)
    cmd := exec.Command("claude", args...)

    if opts != nil && opts.WorkDir != "" {
        cmd.Dir = opts.WorkDir
    }
    return cmd
}
```

### Codex 后端

```go
type Codex struct{}

func (c *Codex) Name() string { return "codex" }

func (c *Codex) IsAvailable() bool {
    _, err := exec.LookPath("codex")
    return err == nil
}

func (c *Codex) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    return exec.Command("codex", args...)
}
```

### Gemini 后端

```go
type Gemini struct{}

func (g *Gemini) Name() string { return "gemini" }

func (g *Gemini) IsAvailable() bool {
    _, err := exec.LookPath("gemini")
    return err == nil
}

func (g *Gemini) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--output-format", "json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    return exec.Command("gemini", args...)
}
```

## 统一选项处理

`UnifiedOptions` 结构体（`internal/backend/unified.go:174-219`）提供了一种后端无关的方式来配置 AI CLI 命令：

```go
type UnifiedOptions struct {
    WorkDir       string
    Model         string
    ApprovalMode  ApprovalMode
    SandboxMode   SandboxMode
    OutputFormat  OutputFormat
    AllowedTools  string
    AllowedDirs   []string
    Interactive   bool
    Verbose       bool
    DryRun        bool
    MaxTokens     int
    MaxTurns      int
    SystemPrompt  string
    ExtraFlags    []string
    Ephemeral     bool
}
```

### 标志映射架构

```mermaid
flowchart TB
    subgraph 统一["UnifiedOptions"]
        MODEL[Model]
        APPROVAL[ApprovalMode]
        SANDBOX[SandboxMode]
        OUTPUT[OutputFormat]
    end

    subgraph 映射器["标志映射器 (internal/backend/unified.go:273-568)"]
        MAP_MODEL[mapModel()]
        MAP_APPROVAL[mapApprovalMode()]
        MAP_SANDBOX[mapSandboxMode()]
        MAP_OUTPUT[mapOutputFormat()]
    end

    subgraph 后端["后端特定标志"]
        CLAUDE[Claude 标志]
        CODEX[Codex 标志]
        GEMINI[Gemini 标志]
    end

    MODEL --> MAP_MODEL
    APPROVAL --> MAP_APPROVAL
    SANDBOX --> MAP_SANDBOX
    OUTPUT --> MAP_OUTPUT

    MAP_MODEL --> CLAUDE
    MAP_MODEL --> CODEX
    MAP_MODEL --> GEMINI
    MAP_APPROVAL --> CLAUDE
    MAP_APPROVAL --> CODEX
    MAP_APPROVAL --> GEMINI
    MAP_SANDBOX --> CLAUDE
    MAP_SANDBOX --> CODEX
    MAP_SANDBOX --> GEMINI
    MAP_OUTPUT --> CLAUDE
    MAP_OUTPUT --> CODEX
    MAP_OUTPUT --> GEMINI
```

### 模型名称映射

统一模型别名映射到后端特定名称：

| 统一别名 | Claude | Codex | Gemini |
|--------------|--------|-------|--------|
| `fast` | `haiku` | `gpt-4.1-mini` | `gemini-2.5-flash` |
| `balanced` | `sonnet` | `gpt-5.2` | `gemini-2.5-pro` |
| `best` | `opus` | `gpt-5-codex` | `gemini-2.5-pro` |

### 审批模式映射

审批模式控制后端如何请求用户确认：

```go
func (m *flagMapper) mapApprovalMode(mode ApprovalMode) []string {
    switch m.backend {
    case "claude":
        switch mode {
        case ApprovalAuto:
            return []string{"--permission-mode", "acceptEdits"}
        case ApprovalNone:
            return []string{"--permission-mode", "dontAsk"}
        case ApprovalAlways:
            return []string{"--permission-mode", "default"}
        }
    case "codex":
        switch mode {
        case ApprovalAuto:
            return []string{"--ask-for-approval", "on-request"}
        case ApprovalNone:
            return []string{"--ask-for-approval", "never"}
        case ApprovalAlways:
            return []string{"--ask-for-approval", "untrusted"}
        }
    // ...
    }
    return nil
}
```

## 输出解析和规范化

每个后端将其原生输出解析为统一格式：

### JSON 响应解析

```go
func (c *Claude) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
    // 首先尝试解析为错误响应
    var errResp claudeErrorResponse
    if err := json.Unmarshal([]byte(rawOutput), &errResp); err == nil {
        if errResp.Error != "" {
            return &UnifiedResponse{
                SessionID: errResp.SessionID,
                Error:     errResp.Error,
            }, nil
        }
    }

    var resp claudeJSONResponse
    if err := json.Unmarshal([]byte(rawOutput), &resp); err != nil {
        return nil, err
    }

    return &UnifiedResponse{
        Content:    resp.Result,
        SessionID:  resp.SessionID,
        DurationMs: resp.DurationMs,
        Usage: &TokenUsage{
            InputTokens:  resp.Usage.InputTokens,
            OutputTokens: resp.Usage.OutputTokens,
        },
    }, nil
}
```

### 统一响应结构

```go
type UnifiedResponse struct {
    Content    string
    SessionID  string
    Model      string
    DurationMs int64
    Usage      *TokenUsage
    Error      string
    Raw        map[string]any
}
```

## 添加新后端

要向 clinvoker 添加新的 AI CLI 后端：

### 步骤 1：创建实现文件

创建 `internal/backend/newbackend.go`：

```go
package backend

import "os/exec"

type NewBackend struct{}

func (n *NewBackend) Name() string {
    return "newbackend"
}

func (n *NewBackend) IsAvailable() bool {
    _, err := exec.LookPath("newbackend-cli")
    return err == nil
}

func (n *NewBackend) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--output", "json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    cmd := exec.Command("newbackend-cli", args...)

    if opts != nil && opts.WorkDir != "" {
        cmd.Dir = opts.WorkDir
    }
    return cmd
}

func (n *NewBackend) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
    args := []string{"--resume", sessionID, "--output", "json"}

    if prompt != "" {
        args = append(args, prompt)
    }

    return exec.Command("newbackend-cli", args...)
}

func (n *NewBackend) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
    return n.BuildCommand(prompt, MapFromUnified(n.Name(), opts))
}

func (n *NewBackend) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
    return n.ResumeCommand(sessionID, prompt, MapFromUnified(n.Name(), opts))
}

func (n *NewBackend) ParseOutput(rawOutput string) string {
    return rawOutput
}

func (n *NewBackend) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
    var resp struct {
        Content   string `json:"content"`
        SessionID string `json:"session_id"`
        Usage     struct {
            Input  int `json:"input_tokens"`
            Output int `json:"output_tokens"`
        } `json:"usage"`
    }

    if err := json.Unmarshal([]byte(rawOutput), &resp); err != nil {
        return nil, err
    }

    return &UnifiedResponse{
        Content:   resp.Content,
        SessionID: resp.SessionID,
        Usage: &TokenUsage{
            InputTokens:  resp.Usage.Input,
            OutputTokens: resp.Usage.Output,
        },
    }, nil
}

func (n *NewBackend) SeparateStderr() bool {
    return false
}
```

### 步骤 2：在注册表中注册

添加到 `internal/backend/registry.go`：

```go
func init() {
    globalRegistry.Register(&Claude{})
    globalRegistry.Register(&Codex{})
    globalRegistry.Register(&Gemini{})
    globalRegistry.Register(&NewBackend{}) // 添加此行
}
```

### 步骤 3：添加统一选项映射

更新 `internal/backend/unified.go` 以添加标志映射：

```go
func (m *flagMapper) mapModel(model string) string {
    switch m.backend {
    case "newbackend":
        switch model {
        case "fast":
            return "newbackend-fast"
        case "balanced":
            return "newbackend-balanced"
        case "best":
            return "newbackend-pro"
        default:
            return model
        }
    // ...
    }
}
```

### 步骤 4：添加允许的标志

更新 `internal/backend/unified.go:10-27` 中的允许列表：

```go
var allowedFlagPatterns = map[string][]string{
    "newbackend": {
        "--model", "--output", "--verbose",
        "--resume", "--sandbox",
    },
    // ...
}
```

## 最佳实践

### 命令构建

- 执行前始终验证路径
- 正确转义参数（Go 的 `exec.Command` 会处理此问题）
- 支持交互式和批处理模式
- 使用 `--print` 或等效选项进行非交互式输出

### 输出解析

- 优雅地处理部分/无效的 JSON
- 去除控制字符和 ANSI 代码
- 保留错误消息用于调试
- 尽可能返回结构化错误

### 错误处理

- 区分后端错误和系统错误
- 提供清晰、可操作的错误消息
- 在错误输出中包含故障排除提示

## 相关文档

- [架构概述](architecture.zh.md) - 高级系统架构
- [会话系统](session-system.zh.md) - 会话持久化机制
- [API 设计](api-design.zh.md) - REST API 架构
