# 架构

clinvoker 是一个轻量级编排层，封装现有 AI CLI 工具，提供统一访问和强大的组合能力。

## 系统概览

```mermaid
flowchart TB
    subgraph interface ["用户接口"]
        CLI["CLI 命令"]
        HTTP["HTTP 服务器"]
    end

    subgraph core ["核心"]
        Exec["执行器"]
        Session["会话管理器"]
    end

    subgraph backends ["后端"]
        Claude["claude"]
        Codex["codex"]
        Gemini["gemini"]
    end

    CLI --> Exec
    HTTP --> Exec
    Exec --> Claude
    Exec --> Codex
    Exec --> Gemini
    Exec <--> Session
```bash

## 核心原则

### 1. 封装而非替代

clinvk 不替代 AI CLI 工具——而是封装它们：

- **零锁定**：你始终可以直接使用底层 CLI
- **自动更新**：后端更新时，clinvk 立即受益
- **完全兼容**：所有后端功能保持可访问

### 2. 统一接口

尽管不同后端有不同接口，clinvk 提供：

- **一致的命令**：所有后端使用相同语法
- **通用输出格式**：统一的 JSON 结构
- **共享配置**：一个配置文件管理所有后端

### 3. 组合优于复杂

复杂工作流由简单原语构建：

- **并行**：同时运行多个后端
- **链式**：顺序通过后端传递输出
- **对比**：并排获取所有后端的响应

## 组件

| 组件 | 职责 |
|------|------|
| **CLI** | 解析命令，处理用户交互 |
| **HTTP 服务器** | REST API，SDK 兼容端点 |
| **执行器** | 运行后端 CLI，捕获输出 |
| **会话管理器** | 跟踪对话，支持恢复 |
| **配置** | 加载设置，解析优先级 |

## 数据流

### 单个提示

```text
用户 → CLI → 执行器 → 后端 CLI → AI 响应 → 用户
```

### 并行执行

```mermaid
flowchart LR
    User["用户"] --> Exec["执行器"]
    Exec --> B1["后端 1"]
    Exec --> B2["后端 2"]
    Exec --> B3["后端 3"]
    B1 --> Agg["聚合"]
    B2 --> Agg
    B3 --> Agg
    Agg --> User2["用户"]
```text

### 链式执行

链式执行将一个后端的输出传递给下一个后端。每个步骤可以使用不同的后端，通过 `{{previous}}` 占位符传递前一步的结果。

```mermaid
sequenceDiagram
    participant User as 用户
    participant Exec as 执行器
    participant A as 后端 A
    participant B as 后端 B

    User->>Exec: 链式请求
    Exec->>A: 步骤 1 提示
    A-->>Exec: 输出 1
    Exec->>B: 步骤 2 + {{previous}}
    B-->>Exec: 输出 2
    Exec-->>User: 最终结果
```

## 配置级联

设置按优先级顺序解析：

1. **CLI 参数**（最高优先级）
2. **环境变量**
3. **配置文件**（`~/.clinvk/config.yaml`）
4. **默认值**（最低优先级）

## 会话存储

会话以 JSON 文件形式存储：

```text
~/.clinvk/sessions/
├── 4f3a2c1d.json
├── 9a8b7c6d.json
└── ...
```

每个会话绑定到单个后端，可通过 `clinvk resume` 恢复。

## 了解更多

- [设计决策](design-decisions.md) - 了解为什么做出某些选择
- [开发架构](../development/architecture.md) - 完整技术细节
- [添加后端](../development/adding-backends.md) - 如何添加新后端
