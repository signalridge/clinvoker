# mcp-http-sse-server Specification

## Purpose

Define MCP HTTP transport behavior with SSE support, including route shape, middleware reuse, stream gating, notification semantics, and MCP-only server surface in HTTP mode.

## Requirements

### Requirement: MCP HTTP server lifecycle

The system MUST provide an MCP server over HTTP with SSE support that accepts JSON-RPC requests and streams notifications.

#### Scenario: Startup and initialization

- **WHEN** the MCP server is started in HTTP mode
- **THEN** it MUST accept JSON-RPC requests over HTTP and stream notifications over SSE

### Requirement: MCP HTTP method support

The system MUST support MCP methods `initialize`, `tools/list`, `tools/call`, and `ping` (if required by clients) over HTTP transport.

#### Scenario: Listing tools over HTTP

- **WHEN** a client calls `tools/list` over HTTP
- **THEN** the server MUST return the list of available tools

### Requirement: MCP HTTP notification semantics

Notification requests MUST NOT receive JSON-RPC responses.

#### Scenario: Notification request

- **WHEN** a client sends a request with no `id` (notification)
- **THEN** the server MUST NOT return a JSON-RPC response (HTTP 204 is acceptable)

### Requirement: MCP HTTP tool surface

The system MUST expose tools mapped to existing capabilities: `clinvk_prompt`, `clinvk_parallel`, `clinvk_chain`, `clinvk_compare`, `clinvk_backends`, `clinvk_sessions`, `clinvk_session_get`, `clinvk_session_delete`, `clinvk_health`.

#### Scenario: Tool list contains required tools

- **WHEN** a client calls `tools/list`
- **THEN** the response MUST include all required tool names

### Requirement: MCP HTTP streaming

For streaming prompt calls, the system MUST translate `StreamPrompt` events into SSE notifications.

#### Scenario: Streaming prompt call over HTTP

- **WHEN** a client calls `tools/call` for `clinvk_prompt` with streaming enabled
- **THEN** the server MUST emit SSE notifications and a final completion response

### Requirement: MCP HTTP execution reuse

The system MUST reuse existing executor/service logic and MUST NOT duplicate execution implementation.

#### Scenario: Executing a tool call

- **WHEN** the server handles any tool call
- **THEN** it MUST dispatch to the existing executor/service layer

### Requirement: MCP HTTP route and middleware

The MCP HTTP route MUST be mounted on the existing server/router so that auth, rate limits, CORS, and timeouts apply.

#### Scenario: Auth middleware applied

- **WHEN** a request is made to the MCP HTTP endpoint
- **THEN** existing auth and rate-limiting middleware MUST be enforced

### Requirement: MCP HTTP endpoint path

The MCP HTTP endpoint MUST be configurable and default to `/mcp`.

#### Scenario: Default path is /mcp

- **WHEN** the MCP HTTP server starts without an explicit path override
- **THEN** it MUST serve MCP requests at `/mcp`

### Requirement: MCP HTTP server surface

The MCP HTTP server MUST expose only the MCP endpoint and optional `/health` when running in MCP mode.

#### Scenario: Only MCP routes are exposed

- **WHEN** the MCP HTTP server starts in MCP mode
- **THEN** it MUST expose `/mcp` (configurable) and MUST NOT expose REST/OpenAI/Anthropic routes

#### Scenario: Optional health endpoint

- **WHEN** `mcp.expose_health` is enabled
- **THEN** the server MUST expose `/health`
- **WHEN** `mcp.expose_health` is disabled
- **THEN** `/health` MUST NOT be exposed

### Requirement: MCP HTTP middleware reuse

The MCP HTTP server MUST reuse the shared router/middleware stack (request size limits, auth, rate limits, CORS, timeouts).

#### Scenario: Middleware parity

- **WHEN** the MCP HTTP server handles a request
- **THEN** request size limits, API key auth, rate limits, CORS, and timeouts MUST be enforced

### Requirement: MCP HTTP streaming gate

Streaming over HTTP MUST be explicitly requested via `output_format=stream-json` and require SSE Accept headers.

#### Scenario: Streaming requires SSE Accept

- **WHEN** a client calls `tools/call` with `output_format=stream-json`
- **AND** the `Accept` header does not include `text/event-stream`
- **THEN** the server MUST return a JSON-RPC error indicating streaming is not supported for this request

#### Scenario: Multi-value Accept header

- **WHEN** a client sends `Accept: text/event-stream, application/json`
- **THEN** the server MUST treat the request as SSE-capable

#### Scenario: Streaming emits SSE notifications

- **WHEN** a client calls `tools/call` for `clinvk_prompt` with `output_format=stream-json`
- **AND** the `Accept` header includes `text/event-stream`
- **THEN** the server MUST emit SSE notifications and a final completion response

### Requirement: MCP HTTP host and port configuration

The MCP HTTP server MUST support host/port configuration and default to 127.0.0.1:8081.

#### Scenario: Default host and port

- **WHEN** the MCP HTTP server starts without host/port overrides
- **THEN** it MUST bind to `127.0.0.1:8081`
