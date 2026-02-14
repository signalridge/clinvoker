# mcp-prompt-streaming Specification

## Purpose

Define prompt streaming behavior for MCP tools, including event-to-notification translation, explicit stream gating, completion semantics, and error propagation.

## Requirements

### Requirement: MCP prompt streaming event translation

The system MUST translate supported `StreamPrompt` `UnifiedEvent` outputs into MCP streaming notifications for prompt tool calls.

#### Scenario: Translate streaming events

- **WHEN** `StreamPrompt` emits message, tool_use, tool_result, thinking, or error events
- **THEN** the MCP server MUST emit corresponding streaming notifications for the client

#### Scenario: Terminal events are completion-only

- **WHEN** `StreamPrompt` emits terminal events such as `done` or token-usage summaries
- **THEN** the MCP server MUST NOT emit extra streaming notifications for those terminal events
- **AND** completion MUST be represented by the final JSON-RPC response

### Requirement: MCP prompt streaming gate

Streaming MUST be explicitly requested via `output_format=stream-json`.

#### Scenario: Streaming not requested

- **WHEN** a client calls `tools/call` for `clinvk_prompt` without `output_format=stream-json`
- **THEN** the server MUST return a single JSON-RPC response and MUST NOT emit notifications

### Requirement: MCP prompt streaming completion

The system MUST send a terminal completion response after streaming notifications end.

#### Scenario: Completion response

- **WHEN** a streaming prompt call completes
- **THEN** the server MUST send a final completion response consistent with MCP expectations

### Requirement: MCP prompt streaming error consistency

Streaming errors MUST be surfaced consistently in both notifications and the final JSON-RPC response.

#### Scenario: Streaming error

- **WHEN** a streaming prompt call encounters an error
- **THEN** the server MUST emit an error notification and return a JSON-RPC error in the final response
