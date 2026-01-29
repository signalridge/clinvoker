# 架构

本文档描述了 clinvk 的内部架构，包括系统设计、请求流程和关键组件。

## 系统架构

```mermaid
flowchart LR
    subgraph clients ["客户端"]
        direction TB
        A1["Claude Code Skills"]
        A2["LangChain/LangGraph"]
        A3["OpenAI SDK"]
        A4["Anthropic SDK"]
        A5["CI/CD"]
    end

    subgraph server ["clinvk 服务"]
        direction TB
        subgraph api ["API 层"]
            B1["/openai/v1/*"]
            B2["/anthropic/v1/*"]
            B3["/api/v1/*"]
        end
        subgraph service ["服务层"]
            C1["Executor"]
            C2["Runner"]
        end
        C3[("后端\n抽象")]
    end

    subgraph backends ["AI CLI 后端"]
        direction TB
        D1["claude"]
        D2["codex"]
        D3["gemini"]
    end

    A1 & A2 & A3 & A4 & A5 --> api
    api --> service
    service --> C3
    C3 --> D1 & D2 & D3

    style clients fill:#e3f2fd,stroke:#1976d2
    style server fill:#fff3e0,stroke:#f57c00
    style backends fill:#f3e5f5,stroke:#7b1fa2
    style C3 fill:#ffecb3,stroke:#ffa000
```

## 层次概览

### HTTP 层

HTTP 层为不同客户端需求提供多个 API 端点：

| 端点 | 格式 | 用例 |
|------|------|------|
| `/openai/v1/*` | OpenAI API 格式 | OpenAI SDK、LangChain |
| `/anthropic/v1/*` | Anthropic API 格式 | Anthropic SDK |
| `/api/v1/*` | 自定义 REST 格式 | 直接集成、Skills |

### 服务层

服务层处理业务逻辑：

- **Executor**：管理任务执行，包括并行和链式模式
- **Runner**：与后端抽象接口交互执行提示
- **Session Manager**：处理会话持久化和检索

### 后端抽象

所有 AI CLI 后端的统一接口：

```go
type Backend interface {
    Name() string
    BuildCommand(req PromptRequest) *exec.Cmd
    ParseResponse(output []byte) (*Response, error)
    SupportsSession() bool
}
```

## 请求流程

### 单个提示请求

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant API as HTTP 处理器
    participant Exec as Executor
    participant Backend as 后端适配器
    participant CLI as 后端 CLI

    Client->>+API: POST /openai/v1/chat/completions
    API->>API: 解析 + 校验请求
    API->>+Exec: PromptRequest
    Exec->>+Backend: 构建命令
    Backend->>+CLI: 执行子进程
    CLI-->>-Backend: 原始输出
    Backend-->>-Exec: 解析后的结果
    Exec-->>-API: PromptResult
    API-->>-Client: OpenAI 兼容响应
```

### 并行执行流程

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Exec as Executor
    participant Claude as claude
    participant Codex as codex
    participant Gemini as gemini

    Client->>+Exec: POST /api/v1/parallel

    par 任务 1
        Exec->>+Claude: 提示 A
        Claude-->>-Exec: 结果 A
    and 任务 2
        Exec->>+Codex: 提示 B
        Codex-->>-Exec: 结果 B
    and 任务 3
        Exec->>+Gemini: 提示 C
        Gemini-->>-Exec: 结果 C
    end

    Exec-->>-Client: 聚合结果
```

### 链式执行流程

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Exec as Executor
    participant Claude as claude
    participant Codex as codex

    Client->>+Exec: POST /api/v1/chain

    Note over Exec,Claude: 步骤 1（分析）
    Exec->>+Claude: 提示 1
    Claude-->>-Exec: 输出 1

    Note over Exec: 用输出 1 替换 {{previous}}

    Note over Exec,Codex: 步骤 2（修复）
    Exec->>+Codex: 提示 2
    Codex-->>-Exec: 输出 2

    Note over Exec: 用输出 2 替换 {{previous}}

    Note over Exec,Claude: 步骤 3（审查）
    Exec->>+Claude: 提示 3
    Claude-->>-Exec: 输出 3

    Exec-->>-Client: 链式结果
```

## 关键组件

### 后端注册表

```mermaid
flowchart TB
    subgraph registry ["后端注册表"]
        direction TB
        subgraph backends ["后端实现"]
            direction LR
            B1["Claude"]
            B2["Codex"]
            B3["Gemini"]
            B4["..."]
        end
        UI[("统一接口")]
    end

    backends --> UI

    style registry fill:#fff8e1,stroke:#ff8f00
    style UI fill:#ffecb3,stroke:#ffa000
```

### 会话管理

会话以 JSON 文件形式存储在 `~/.clinvk/sessions/` 目录下。每个会话都绑定到一个后端（Claude、Codex 或 Gemini）。

```
~/.clinvk/sessions/
├── 4f3a2c1d0e9b8a7c.json
├── 9a8b7c6d5e4f3210.json
└── 4f3a2c1d0e9b8a7c/        # 可选：会话产物
    └── ...
```

### 配置级联

```mermaid
flowchart TB
    A["CLI 参数<br/><small>最高优先级</small>"]
    B["环境变量"]
    C["配置文件<br/><small>~/.clinvk/config.yaml</small>"]
    D["默认值<br/><small>最低优先级</small>"]

    A --> B --> C --> D

    style A fill:#c8e6c9,stroke:#2e7d32
    style B fill:#bbdefb,stroke:#1976d2
    style C fill:#fff9c4,stroke:#f9a825
    style D fill:#ffccbc,stroke:#e64a19
```

## 流式架构

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Server as clinvk
    participant CLI as 后端 CLI

    Client->>+Server: POST /api/v1/prompt<br/>(stream=true)
    Server->>+CLI: 执行并管道输出

    loop 流式传输
        CLI-->>Server: 输出块
        Server-->>Client: SSE: data: {...}
    end

    CLI-->>-Server: 进程退出
    Server-->>-Client: SSE: data: [DONE]
```

## 错误处理

错误通过各层传播，使用适当的 HTTP 状态码：

| 错误类型 | HTTP 状态码 | 描述 |
|----------|-------------|------|
| 无效请求 | 400 | 请求体格式错误 |
| 后端未找到 | 404 | 指定的后端未知 |
| CLI 未安装 | 503 | 后端 CLI 不可用 |
| 执行失败 | 500 | CLI 返回错误 |
| 超时 | 504 | 请求超过超时时间 |

## 下一步

- [设计决策](design-decisions.md) - 了解为什么做出某些选择
- [添加后端](../development/adding-backends.md) - 如何添加新的后端支持
- [REST API 参考](../reference/rest-api.md) - 完整的 API 文档
