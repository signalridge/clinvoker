# Architecture

clinvoker is a lightweight orchestration layer that wraps existing AI CLI tools, providing unified access and powerful composition capabilities.

## System Overview

```mermaid
flowchart TB
    subgraph interface ["User Interface"]
        CLI["CLI Commands"]
        HTTP["HTTP Server"]
    end

    subgraph core ["Core"]
        Exec["Executor"]
        Session["Session Manager"]
    end

    subgraph backends ["Backends"]
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
```

## Key Principles

### 1. Wrapper, Not Replacement

clinvk doesn't replace AI CLI tools—it wraps them:

- **Zero Lock-in**: You can always use the underlying CLIs directly
- **Automatic Updates**: When backends update, clinvk benefits immediately
- **Full Compatibility**: All backend features remain accessible

### 2. Unified Interface

Despite different backends having different interfaces, clinvk provides:

- **Consistent Commands**: Same syntax for all backends
- **Common Output Format**: Unified JSON structure
- **Shared Configuration**: One config file for all backends

### 3. Composition Over Complexity

Complex workflows are built from simple primitives:

- **Parallel**: Run multiple backends simultaneously
- **Chain**: Pipeline output through backends sequentially
- **Compare**: Get responses from all backends side-by-side

## Components

| Component | Responsibility |
|-----------|----------------|
| **CLI** | Parse commands, handle user interaction |
| **HTTP Server** | REST API, SDK-compatible endpoints |
| **Executor** | Run backend CLIs, capture output |
| **Session Manager** | Track conversations, enable resume |
| **Config** | Load settings, resolve priorities |

## Data Flow

### Single Prompt

```
User → CLI → Executor → Backend CLI → AI Response → User
```

### Parallel Execution

```mermaid
flowchart LR
    User --> Exec["Executor"]
    Exec --> B1["Backend 1"]
    Exec --> B2["Backend 2"]
    Exec --> B3["Backend 3"]
    B1 --> Agg["Aggregate"]
    B2 --> Agg
    B3 --> Agg
    Agg --> User2["User"]
```

### Chain Execution

Chain execution pipelines output from one backend to the next. Each step can use a different backend, with `{{previous}}` placeholder passing the prior result.

```mermaid
sequenceDiagram
    participant User
    participant Exec as Executor
    participant A as Backend A
    participant B as Backend B

    User->>Exec: chain request
    Exec->>A: step 1 prompt
    A-->>Exec: output 1
    Exec->>B: step 2 + {{previous}}
    B-->>Exec: output 2
    Exec-->>User: final result
```

## Configuration Cascade

Settings are resolved in priority order:

1. **CLI flags** (highest priority)
2. **Environment variables**
3. **Config file** (`~/.clinvk/config.yaml`)
4. **Default values** (lowest priority)

## Session Storage

Sessions are stored as JSON files:

```
~/.clinvk/sessions/
├── 4f3a2c1d.json
├── 9a8b7c6d.json
└── ...
```

Each session is bound to a single backend and can be resumed with `clinvk resume`.

## Learn More

- [Design Decisions](design-decisions.md) - Why certain choices were made
- [Development Architecture](../development/architecture.md) - Full technical details
- [Adding Backends](../development/adding-backends.md) - How to add new backends
