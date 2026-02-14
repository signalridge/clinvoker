# mcp-stdio-server Specification

## Purpose

Define MCP stdio transport behavior, including JSON-RPC lifecycle, tool exposure, stream notification gating, notification semantics, and CLI entrypoint expectations.

## Requirements

### Requirement: MCP stdio server lifecycle

The system MUST provide an MCP server over stdio that reads JSON-RPC requests from stdin and writes responses to stdout.

#### Scenario: Startup and initialization

- **WHEN** the MCP server is started in stdio mode
- **THEN** it MUST accept JSON-RPC requests over stdin and respond over stdout

### Requirement: MCP stdio method support

The system MUST support MCP methods `initialize`, `tools/list`, `tools/call`, and `ping` (if required by clients).

#### Scenario: Listing tools

- **WHEN** a client calls `tools/list`
- **THEN** the server MUST return the list of available tools

### Requirement: MCP stdio tool surface

The system MUST expose tools mapped to existing capabilities: `clinvk_prompt`, `clinvk_parallel`, `clinvk_chain`, `clinvk_compare`, `clinvk_backends`, `clinvk_sessions`, `clinvk_session_get`, `clinvk_session_delete`, `clinvk_health`.

#### Scenario: Tool list contains required tools

- **WHEN** a client calls `tools/list`
- **THEN** the response MUST include all required tool names

### Requirement: MCP stdio prompt streaming

For streaming prompt calls, the system MUST translate `StreamPrompt` events into MCP notifications.

#### Scenario: Streaming prompt call

- **WHEN** a client calls `tools/call` for `clinvk_prompt` with `output_format=stream-json`
- **THEN** the server MUST emit streaming notifications and a final completion response

### Requirement: MCP stdio execution reuse

The system MUST reuse existing executor/service logic and MUST NOT duplicate execution implementation.

#### Scenario: Executing a tool call

- **WHEN** the server handles any tool call
- **THEN** it MUST dispatch to the existing executor/service layer

### Requirement: MCP stdio CLI entrypoint

The system MUST provide a CLI entrypoint to start the stdio MCP server (e.g., `clinvk mcp --transport stdio`).

#### Scenario: CLI starts MCP stdio server

- **WHEN** the CLI is invoked with the stdio transport
- **THEN** the MCP server MUST start and accept requests

### Requirement: MCP stdio schema alignment

Tool input schemas MUST align with existing REST request struct fields.

#### Scenario: Schema alignment

- **WHEN** tool input schemas are generated or defined
- **THEN** they MUST match the REST request structs used by current handlers

### Requirement: MCP stdio error handling

Errors MUST be surfaced as JSON-RPC errors with consistent codes and messages.

#### Scenario: Invalid request

- **WHEN** the server receives an invalid request
- **THEN** it MUST return a JSON-RPC error response with a consistent error code and message

### Requirement: MCP stdio streaming gate

Streaming notifications MUST be emitted only when `output_format=stream-json` is requested.

#### Scenario: Non-streaming prompt call

- **WHEN** a client calls `tools/call` for `clinvk_prompt` without `output_format=stream-json`
- **THEN** the server MUST return a single JSON-RPC response and MUST NOT emit notifications

### Requirement: MCP stdio notification semantics

Notification requests MUST NOT receive JSON-RPC responses.

#### Scenario: Notification request

- **WHEN** a client sends a request with no `id` (notification)
- **THEN** the server MUST NOT return a JSON-RPC response
