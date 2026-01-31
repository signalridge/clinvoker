---
title: 架构概述
description: clinvoker 的高级系统架构、组件交互和数据流。
---

# 架构概述

本文档全面介绍 clinvoker 的系统架构，解释组件如何交互、数据如何在系统中流动，以及使系统关键能力成为可能的设计模式。

## 高级架构

clinvoker 遵循分层架构模式，具有清晰的关注点分离：

```mermaid
flowchart TB
    subgraph 客户端层["客户端层"]
        CLI[CLI 工具]
        SDK[OpenAI SDK]
        ANTH[Anthropic SDK]
        LANG[LangChain]
        CURL[curl/HTTP]
    end

    subgraph API层["API 层"]
        REST[REST API /api/v1]
        OPENAI[OpenAI 兼容 /openai/v1]
        ANTHAPI[Anthropic 兼容 /anthropic/v1]
    end

    subgraph 核心服务["核心服务"]
        ROUTER[请求路由器]
        RATE[速率限制器]
        AUTH[认证中间件]
        EXEC[执行器]
        SESSION[会话管理器]
        CONFIG[配置管理器]
    end

    subgraph 后端层["后端层"]
        CL[Claude Code]
        CO[Codex CLI]
        GM[Gemini CLI]
    end

    subgraph 存储层["存储层"]
        DISC[磁盘会话]
        CONFIG_FILE[配置文件]
    end

    CLI --> REST
    SDK --> OPENAI
    ANTH --> ANTHAPI
    LANG --> OPENAI
    CURL --> REST

    REST --> ROUTER
    OPENAI --> ROUTER
    ANTHAPI --> ROUTER

    ROUTER --> RATE
    RATE --> AUTH
    AUTH --> EXEC

    EXEC --> SESSION
    EXEC --> CONFIG

    EXEC --> CL
    EXEC --> CO
    EXEC --> GM

    SESSION --> DISC
    CONFIG --> CONFIG_FILE
```

## CLI 层深入解析

### 入口点 (`cmd/clinvk/main.go`)

入口点有意保持简洁，遵循 Go 最佳实践：

```go
package main

import (
    "os"
    "github.com/signalridge/clinvoker/internal/app"
)

func main() {
    if err := app.Execute(); err != nil {
        os.Exit(1)
    }
}
```

这种设计：
- 保持 main 包简洁专注
- 便于测试应用逻辑
- 允许导入 app 包进行集成测试

### Cobra 框架使用 (`internal/app/app.go`)

CLI 使用 Cobra 框架进行命令管理：

```mermaid
flowchart LR
    subgraph Cobra命令结构
        ROOT[rootCmd
        持久化标志]
        VERSION[versionCmd]
        RESUME[resumeCmd]
        SESSIONS[sessionsCmd]
        CONFIG[configCmd]
        PARALLEL[parallelCmd]
        COMPARE[compareCmd]
        CHAIN[chainCmd]
    end

    ROOT --> VERSION
    ROOT --> RESUME
    ROOT --> SESSIONS
    ROOT --> CONFIG
    ROOT --> PARALLEL
    ROOT --> COMPARE
    ROOT --> CHAIN
```

使用的关键 Cobra 特性：

1. **持久化标志**：对所有子命令可用
   - `--config`：配置文件路径
   - `--backend`：后端选择
   - `--model`：模型选择
   - `--workdir`：工作目录
   - `--dry-run`：模拟模式
   - `--output-format`：输出格式
   - `--ephemeral`：无状态模式

2. **本地标志**：特定于单个命令
   - `--continue`：继续上次会话（仅根命令）

3. **预运行初始化**：`initConfig()` 在命令执行前加载配置

### 命令执行流程

```mermaid
sequenceDiagram
    autonumber
    participant 用户
    participant CLI as CLI 层
    participant Config as 配置管理器
    participant Backend as 后端注册表
    participant Session as 会话存储
    participant Executor as 执行器
    participant AI as AI 后端

    用户->>CLI: clinvk "prompt"
    CLI->>Config: 加载配置
    Config-->>CLI: 后端设置、默认值

    CLI->>Backend: Get(backendName)
    Backend-->>CLI: 后端实例

    alt 临时模式
        CLI->>CLI: 跳过会话创建
    else 普通模式
        CLI->>Session: CreateWithOptions()
        Session-->>CLI: 会话实例
    end

    CLI->>Backend: BuildCommandUnified(prompt, opts)
    Backend-->>CLI: *exec.Cmd

    CLI->>Executor: ExecuteCommand(config, cmd)
    Executor->>AI: 执行子进程
    AI-->>Executor: 输出
    Executor-->>CLI: 结果

    alt 普通模式
        CLI->>Session: Save(session)
    end

    CLI-->>用户: 格式化输出
```

## 核心组件交互

### 后端注册表模式

后端注册表使用线程安全的注册表模式管理 AI CLI 后端：

```mermaid
flowchart TB
    subgraph 注册表["后端注册表 (internal/backend/registry.go)"]
        RWMUTEX[sync.RWMutex]
        BACKENDS[map[string]Backend]
        CACHE[availabilityCache
        30秒 TTL]
    end

    subgraph 后端["已注册后端"]
        CL[Claude 后端]
        CO[Codex 后端]
        GM[Gemini 后端]
    end

    RWMUTEX --> BACKENDS
    BACKENDS --> CL
    BACKENDS --> CO
    BACKENDS --> GM
    BACKENDS --> CACHE
```

注册表提供：
- **线程安全访问**：`sync.RWMutex` 用于并发读/写
- **可用性缓存**：30 秒 TTL 避免频繁的 PATH 查找
- **动态注册**：后端可以在运行时注册/注销

### 会话管理器架构

```mermaid
flowchart TB
    subgraph 会话管理器["会话管理器 (internal/session/)"]
        STORE[存储]
        INDEX[内存索引
        map[string]*SessionMeta]
        FILELOCK[FileLock
        跨进程同步]
        RWLOCK[sync.RWMutex
        进程内同步]
    end

    subgraph 存储["文件系统存储"]
        DIR[~/.clinvk/sessions/]
        SESSION_FILES[*.json 文件]
        INDEX_FILE[index.json]
    end

    STORE --> INDEX
    STORE --> FILELOCK
    STORE --> RWLOCK
    FILELOCK --> DIR
    RWLOCK --> DIR
    DIR --> SESSION_FILES
    DIR --> INDEX_FILE
```

会话管理器使用双重锁定策略：
1. **进程内**：`sync.RWMutex` 用于 goroutine 安全
2. **跨进程**：文件锁用于 CLI/服务器共存

### HTTP 服务器请求处理

```mermaid
flowchart LR
    subgraph 中间件栈["中间件栈 (internal/server/server.go:58-131)"]
        REQID[RequestID]
        REALIP[RealIP]
        RECOVER[Recoverer]
        LOGGER[RequestLogger]
        SIZE[RequestSize 限制]
        RATE[速率限制器]
        AUTH[API 密钥认证]
        TIMEOUT[超时]
        CORS[CORS 处理器]
    end

    subgraph 处理器["API 处理器"]
        CUSTOM[自定义处理器
        /api/v1/*]
        OPENAI[OpenAI 处理器
        /openai/v1/*]
        ANTH[Anthropic 处理器
        /anthropic/v1/*]
    end

    REQID --> REALIP
    REALIP --> RECOVER
    RECOVER --> LOGGER
    LOGGER --> SIZE
    SIZE --> RATE
    RATE --> AUTH
    AUTH --> TIMEOUT
    TIMEOUT --> CORS
    CORS --> CUSTOM
    CORS --> OPENAI
    CORS --> ANTH
```

中间件执行顺序至关重要：
1. **RequestID**：分配唯一请求 ID 用于跟踪
2. **RealIP**：提取真实客户端 IP（代理后面）
3. **Recoverer**：从 panic 中恢复
4. **RequestLogger**：记录请求详情
5. **RequestSize**：强制执行请求体大小限制
6. **RateLimiter**：应用速率限制
7. **API Key Auth**：认证请求
8. **Timeout**：强制执行请求超时
9. **CORS**：处理跨域请求

## 数据流

### CLI 提示流程

```mermaid
sequenceDiagram
    autonumber
    participant 用户
    participant CLI
    participant Config as 配置
    participant Backend as 后端
    participant Session as 会话
    participant AI

    用户->>CLI: clinvk "prompt"
    CLI->>Config: 加载配置
    Config-->>CLI: 后端设置
    CLI->>Session: 创建/恢复会话
    CLI->>Backend: 构建命令
    Backend-->>CLI: exec.Cmd
    CLI->>AI: 执行命令
    AI-->>CLI: 输出
    CLI->>Session: 保存会话
    CLI-->>用户: 格式化输出
```

### HTTP API 流程

```mermaid
sequenceDiagram
    autonumber
    participant 客户端
    participant API
    participant Auth as 认证
    participant Rate as 速率限制
    participant Executor as 执行器
    participant Backend as 后端

    客户端->>API: POST /api/v1/prompt
    API->>Auth: 验证 API 密钥
    Auth-->>API: OK
    API->>Rate: 检查限制
    Rate-->>API: 允许
    API->>Executor: 执行提示
    Executor->>Backend: 运行命令
    Backend-->>Executor: 结果
    Executor-->>API: 响应
    API-->>客户端: JSON 响应
```

### 流式响应流程

```mermaid
sequenceDiagram
    autonumber
    participant 客户端
    participant API
    participant Streamer as 流处理器
    participant Backend as 后端
    participant AI

    客户端->>API: POST /openai/v1/chat/completions
    Note over 客户端,API: stream: true
    API->>Streamer: 创建 SSE 流
    Streamer->>Backend: 带流式执行
    Backend->>AI: 启动子进程

    loop 每个数据块
        AI-->>Streamer: 输出数据块
        Streamer-->>客户端: data: {...}
    end

    AI-->>Streamer: 完成
    Streamer-->>客户端: data: [DONE]
```

## 后端抽象架构

### 统一后端接口

所有后端都实现 `Backend` 接口（`internal/backend/backend.go:16-46`）：

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

### 后端实现结构

```mermaid
flowchart TB
    subgraph 后端接口["后端接口"]
        INTERFACE[Backend 接口
        internal/backend/backend.go]
    end

    subgraph 实现["后端实现"]
        CLAUDE[Claude
        internal/backend/claude.go]
        CODEX[Codex
        internal/backend/codex.go]
        GEMINI[Gemini
        internal/backend/gemini.go]
    end

    subgraph 统一层["统一选项层"]
        UNIFIED[UnifiedOptions
        internal/backend/unified.go]
        MAPPER[标志映射器
        MapToOptions()]
    end

    INTERFACE --> CLAUDE
    INTERFACE --> CODEX
    INTERFACE --> GEMINI
    UNIFIED --> MAPPER
    MAPPER --> CLAUDE
    MAPPER --> CODEX
    MAPPER --> GEMINI
```

## 并发模式

### 注册表并发

后端注册表使用读写互斥锁模式：

```go
// 读操作（并发安全）
func (r *Registry) Get(name string) (Backend, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // ... 查找后端
}

// 写操作（独占）
func (r *Registry) Register(b Backend) {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ... 注册后端
}
```

### 会话存储并发

会话存储结合多种同步机制：

```mermaid
flowchart TB
    subgraph 进程内["进程内"]
        RW[sync.RWMutex]
        INDEX[内存索引]
    end

    subgraph 跨进程["跨进程"]
        FLOCK[FileLock
        flock 系统调用]
        FILES[会话文件]
    end

    RW --> INDEX
    FLOCK --> FILES
```

读流程：
1. 获取读锁（`RLock`）
2. 检查内存索引
3. 如需要加载会话文件
4. 释放读锁

写流程：
1. 获取跨进程文件锁
2. 获取写锁（`Lock`）
3. 原子写入会话
4. 更新索引
5. 释放锁

## 扩展点

### 添加新后端

添加新的 AI CLI 后端：

1. **创建实现文件**：`internal/backend/newbackend.go`
2. **实现 Backend 接口**：所有必需的方法
3. **在注册表中注册**：添加到 `registry.go` 的 `init()`
4. **添加统一选项映射**：更新 `unified.go` 标志映射器

### 添加新 CLI 命令

1. **创建命令文件**：`internal/app/cmd_newcommand.go`
2. **定义 Cobra 命令**：使用现有命令作为模板
3. **添加到根命令**：在 `app.go` 的 `init()` 函数中

### 添加新 API 端点

1. **创建处理器**：在适当的处理器文件中（`custom.go`、`openai.go` 或 `anthropic.go`）
2. **使用 Huma 注册**：使用 `huma.Register()` 和操作配置
3. **如需要添加中间件**：在 `server.go` 中更新中间件栈

## 可扩展性考虑

### 水平扩展

服务器组件可以水平扩展：

- **无状态**：请求之间不共享内存状态
- **会话存储**：会话在共享文件系统或数据库上
- **配置**：启动时加载，运行时不变

### 后端池化（未来）

对于高吞吐量场景，后端可以池化：

```mermaid
flowchart LR
    EXEC[执行器]
    subgraph Claude池["Claude 池"]
        C1[实例 1]
        C2[实例 2]
        C3[实例 N]
    end
    EXEC --> C1
    EXEC --> C2
    EXEC --> C3
```

## 安全架构

### 认证

- 入口点的 API 密钥验证
- 多个密钥来源（环境变量、gopass、请求头）
- 每个密钥的速率限制

### 授权

- 通过配置的后端级权限
- 工作目录限制
- 沙盒模式支持

### 隔离

- 会话隔离（无跨会话数据泄漏）
- 工作目录限制
- 子进程隔离

## 监控和可观测性

### 指标

- 请求计数和延迟（Prometheus）
- 后端可用性
- Token 使用量
- 错误率

### 日志

- 结构化 JSON 日志
- 请求/响应跟踪
- 后端命令日志（可选）

### 健康检查

- 负载均衡器的 `/health` 端点
- 后端可用性检查
- 会话存储健康

## 相关文档

- [后端系统](backend-system.zh.md) - 后端实现详情
- [会话系统](session-system.zh.md) - 会话持久化深入解析
- [API 设计](api-design.zh.md) - API 架构
- [设计决策](design-decisions.zh.md) - 架构决策记录
