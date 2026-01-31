# 集成指南

clinvk 设计为与 AI 开发工作流无缝集成，从 Claude Code Skills 到 LangChain agent 再到 CI/CD 流水线。

## 集成概览

```mermaid
flowchart LR
    subgraph ai ["AI 开发"]
        direction TB
        A1["Claude Code Skills"]
        A2["LangChain/LangGraph"]
        A3["自定义 AI Agent"]
    end

    subgraph sdks ["SDK"]
        direction TB
        B1["OpenAI SDK"]
        B2["Anthropic SDK"]
    end

    subgraph auto ["自动化"]
        direction TB
        C1["CI/CD Pipeline"]
        C2["Shell 脚本"]
    end

    A1 --> D["clinvk 服务"]
    A2 --> D
    A3 --> D
    B1 --> D
    B2 --> D
    C1 --> D
    C2 --> D

    D --> E1["Claude CLI"]
    D --> E2["Codex CLI"]
    D --> E3["Gemini CLI"]

    style ai fill:#e3f2fd,stroke:#1976d2
    style sdks fill:#fff3e0,stroke:#f57c00
    style auto fill:#f3e5f5,stroke:#7b1fa2
    style D fill:#ffecb3,stroke:#ffa000
    style E1 fill:#f3e5f5,stroke:#7b1fa2
    style E2 fill:#e8f5e9,stroke:#388e3c
    style E3 fill:#ffebee,stroke:#c62828
```

## 集成方式

| 方式 | 用例 | API 端点 |
|------|------|----------|
| [Claude Code Skills](claude-code-skills.md) | 用多后端能力扩展 Claude | `/api/v1/*` |
| [LangChain/LangGraph](langchain-langgraph.md) | AI 框架集成 | `/openai/v1/*` |
| [CI/CD](ci-cd/index.md) | 自动化代码审查、文档 | `/api/v1/*` |
| [客户端库](../../reference/api/index.md) | Python、TypeScript、Go 客户端 | 所有端点 |
| [MCP 服务器](mcp-server.md) | Model Context Protocol 集成 | 未来 |

## 快速集成示例

### Claude Code Skill

```bash
# 在你的 skill 脚本中
curl -s http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "gemini", "prompt": "分析这个数据"}'
```

### LangChain

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed"
)
```

### OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)
```

### GitHub Actions

```yaml
- name: AI 代码审查
  run: |
    payload=$(jq -n --arg prompt "审查：\n${{ steps.diff.outputs.changes }}" '{backend:"claude", prompt:$prompt}')
    curl -sS -X POST http://localhost:8080/api/v1/prompt \
      -H "Content-Type: application/json" \
      -d "$payload"
```

## 前提条件

集成之前，确保：

1. **clinvk 已安装**：见 [安装](../../tutorials/getting-started.md)
2. **后端 CLI 可用**：至少有 `claude`、`codex` 或 `gemini` 之一
3. **服务器正在运行**：`clinvk serve --port 8080`

## 选择正确的端点

| 你的情况 | 推荐端点 |
|---------|---------|
| 使用 OpenAI SDK 或 LangChain | `/openai/v1/*` |
| 使用 Anthropic SDK | `/anthropic/v1/*` |
| 构建 Claude Code Skills | `/api/v1/*` |
| 需要并行/链式执行 | `/api/v1/*` |
| 简单 REST 集成 | `/api/v1/*` |

## 下一步

选择你的集成路径：

- **AI Agent 开发**：从 [Claude Code Skills](claude-code-skills.md) 开始
- **框架集成**：见 [LangChain/LangGraph](langchain-langgraph.md)
- **自动化**：查看 [CI/CD 集成](ci-cd/index.md)
- **自定义开发**：查阅 [客户端库](../../reference/api/index.md)
