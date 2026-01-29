# Architecture

This document describes the internal architecture of clinvk, including system design, request flow, and key components.

## System Architecture

```mermaid
flowchart LR
    subgraph clients ["Clients"]
        direction TB
        A1["Claude Code Skills"]
        A2["LangChain/LangGraph"]
        A3["OpenAI SDK"]
        A4["Anthropic SDK"]
        A5["CI/CD"]
    end

    subgraph server ["clinvk server"]
        direction TB
        subgraph api ["API layer"]
            B1["/openai/v1/*"]
            B2["/anthropic/v1/*"]
            B3["/api/v1/*"]
        end
        subgraph service ["Service layer"]
            C1["Executor"]
            C2["Runner"]
        end
        C3[("Backend\nabstraction")]
    end

    subgraph backends ["AI CLI backends"]
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

## Layer Overview

### HTTP Layer

The HTTP layer provides multiple API endpoints for different client needs:

| Endpoint | Format | Use Case |
|----------|--------|----------|
| `/openai/v1/*` | OpenAI API format | OpenAI SDK, LangChain |
| `/anthropic/v1/*` | Anthropic API format | Anthropic SDK |
| `/api/v1/*` | Custom REST format | Direct integration, Skills |

### Service Layer

The service layer handles business logic:

- **Executor**: Manages task execution, including parallel and chain modes
- **Runner**: Interfaces with backend abstraction to execute prompts
- **Session Manager**: Handles session persistence and retrieval

### Backend Abstraction

A unified interface for all AI CLI backends:

```go
type Backend interface {
    Name() string
    BuildCommand(req PromptRequest) *exec.Cmd
    ParseResponse(output []byte) (*Response, error)
    SupportsSession() bool
}
```

## Request Flow

### Single Prompt Request

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant API as HTTP handler
    participant Exec as Executor
    participant Backend as Backend adapter
    participant CLI as Backend CLI

    Client->>+API: POST /openai/v1/chat/completions
    API->>API: Parse + validate request
    API->>+Exec: PromptRequest
    Exec->>+Backend: Build command
    Backend->>+CLI: Execute subprocess
    CLI-->>-Backend: Raw output
    Backend-->>-Exec: Parsed result
    Exec-->>-API: PromptResult
    API-->>-Client: OpenAI-compatible response
```

### Parallel Execution Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Exec as Executor
    participant Claude as claude
    participant Codex as codex
    participant Gemini as gemini

    Client->>+Exec: POST /api/v1/parallel

    par task 1
        Exec->>+Claude: prompt A
        Claude-->>-Exec: result A
    and task 2
        Exec->>+Codex: prompt B
        Codex-->>-Exec: result B
    and task 3
        Exec->>+Gemini: prompt C
        Gemini-->>-Exec: result C
    end

    Exec-->>-Client: aggregated results
```

### Chain Execution Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Exec as Executor
    participant Claude as claude
    participant Codex as codex

    Client->>+Exec: POST /api/v1/chain

    Note over Exec,Claude: step 1 (analysis)
    Exec->>+Claude: prompt 1
    Claude-->>-Exec: output 1

    Note over Exec: replace {{previous}} with output 1

    Note over Exec,Codex: step 2 (fix)
    Exec->>+Codex: prompt 2
    Codex-->>-Exec: output 2

    Note over Exec: replace {{previous}} with output 2

    Note over Exec,Claude: step 3 (review)
    Exec->>+Claude: prompt 3
    Claude-->>-Exec: output 3

    Exec-->>-Client: chain results
```

## Key Components

### Backend Registry

```mermaid
flowchart TB
    subgraph registry ["Backend registry"]
        direction TB
        subgraph backends ["Backend implementations"]
            direction LR
            B1["Claude"]
            B2["Codex"]
            B3["Gemini"]
            B4["..."]
        end
        UI[("Unified interface")]
    end

    backends --> UI

    style registry fill:#fff8e1,stroke:#ff8f00
    style UI fill:#ffecb3,stroke:#ffa000
```

### Session Management

Sessions are stored as JSON files under `~/.clinvk/sessions/`. Each session is bound to a single backend (Claude, Codex, or Gemini).

```
~/.clinvk/sessions/
├── 4f3a2c1d0e9b8a7c.json
├── 9a8b7c6d5e4f3210.json
└── 4f3a2c1d0e9b8a7c/        # optional artifacts
    └── ...
```

### Configuration Cascade

```mermaid
flowchart TB
    A["CLI flags<br/><small>Highest priority</small>"]
    B["Environment variables"]
    C["Config file<br/><small>~/.clinvk/config.yaml</small>"]
    D["Default values<br/><small>Lowest priority</small>"]

    A --> B --> C --> D

    style A fill:#c8e6c9,stroke:#2e7d32
    style B fill:#bbdefb,stroke:#1976d2
    style C fill:#fff9c4,stroke:#f9a825
    style D fill:#ffccbc,stroke:#e64a19
```

## Streaming Architecture

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Server as clinvk
    participant CLI as Backend CLI

    Client->>+Server: POST /api/v1/prompt<br/>(stream=true)
    Server->>+CLI: Execute with pipe

    loop streaming
        CLI-->>Server: output chunk
        Server-->>Client: SSE: data: {...}
    end

    CLI-->>-Server: Process exit
    Server-->>-Client: SSE: data: [DONE]
```

## Error Handling

Errors are propagated through the layers with appropriate HTTP status codes:

| Error Type | HTTP Status | Description |
|------------|-------------|-------------|
| Invalid Request | 400 | Malformed request body |
| Backend Not Found | 404 | Unknown backend specified |
| CLI Not Installed | 503 | Backend CLI not available |
| Execution Failed | 500 | CLI returned error |
| Timeout | 504 | Request exceeded timeout |

## Next Steps

- [Design Decisions](design-decisions.md) - Understand why certain choices were made
- [Adding Backends](../development/adding-backends.md) - How to add new backend support
- [REST API Reference](../reference/rest-api.md) - Complete API documentation
