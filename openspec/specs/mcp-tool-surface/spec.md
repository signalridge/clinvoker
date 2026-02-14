# mcp-tool-surface Specification

## Purpose

Define the MCP tool catalog exposed by clinvoker, including stable tool naming, schema contracts, dispatch mapping, and output shape parity with REST responses.

## Requirements

### Requirement: MCP tool surface completeness

The system MUST expose a tool surface covering prompt, parallel, chain, compare, backends, sessions, and health.

#### Scenario: Tool list includes core tools

- **WHEN** a client calls `tools/list`
- **THEN** the response MUST include tools for prompt, parallel, chain, compare, backends, sessions, and health

### Requirement: MCP tool naming

The system MUST use stable tool names for existing capabilities: `clinvk_prompt`, `clinvk_parallel`, `clinvk_chain`, `clinvk_compare`, `clinvk_backends`, `clinvk_sessions`, `clinvk_session_get`, `clinvk_session_delete`, `clinvk_health`.

#### Scenario: Tool names are stable

- **WHEN** a client calls `tools/list`
- **THEN** tool names MUST match the defined names

### Requirement: MCP tool schema includes inputs and outputs

The system MUST include tool input and output schemas for each tool using the REST request and response structs as the source of truth.

#### Scenario: Tool schema includes input/output

- **WHEN** a client calls `tools/list`
- **THEN** each tool MUST include input and output schema definitions derived from REST structs

### Requirement: MCP tool execution mapping

Each tool MUST map to the corresponding existing executor/service method.

#### Scenario: Tool invocation dispatch

- **WHEN** a client calls a tool backed by shared REST request models
- **THEN** the server MUST validate inputs using equivalent rules and error semantics as the REST handlers
- **THEN** the server MUST dispatch to the corresponding executor/service operation

### Requirement: MCP tool output shape

Tool results MUST match the REST response body shape for that tool.

#### Scenario: Prompt output shape

- **WHEN** a client calls `clinvk_prompt`
- **THEN** the tool result content MUST contain a JSON-encoded `PromptResponseBody`

#### Scenario: Parallel output shape

- **WHEN** a client calls `clinvk_parallel`
- **THEN** the tool result content MUST contain a JSON-encoded `ParallelResponseBody`

#### Scenario: Chain output shape

- **WHEN** a client calls `clinvk_chain`
- **THEN** the tool result content MUST contain a JSON-encoded `ChainResponseBody`

#### Scenario: Compare output shape

- **WHEN** a client calls `clinvk_compare`
- **THEN** the tool result content MUST contain a JSON-encoded `CompareResponseBody`

#### Scenario: Backends output shape

- **WHEN** a client calls `clinvk_backends`
- **THEN** the tool result content MUST contain a JSON-encoded `BackendsResponseBody`

#### Scenario: Sessions output shape

- **WHEN** a client calls `clinvk_sessions`
- **THEN** the tool result content MUST contain a JSON-encoded `SessionsResponseBody`

#### Scenario: Session get output shape

- **WHEN** a client calls `clinvk_session_get`
- **THEN** the tool result content MUST contain a JSON-encoded `SessionInfo`

#### Scenario: Session delete output shape

- **WHEN** a client calls `clinvk_session_delete`
- **THEN** the tool result content MUST contain a JSON-encoded `DeleteSessionResponseBody`

#### Scenario: Health output shape

- **WHEN** a client calls `clinvk_health`
- **THEN** the tool result content MUST contain a JSON-encoded `HealthResponseBody`
