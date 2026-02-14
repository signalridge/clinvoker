# http-api-and-compatibility Specification

## Purpose

Define baseline HTTP server behavior across native REST endpoints and OpenAI/Anthropic compatibility routes.

## Requirements

### Requirement: Native REST surface

The server MUST expose native REST endpoints for prompt, parallel, chain, compare, backends, sessions, and health.

#### Scenario: Native prompt execution

- **WHEN** client POSTs `/api/v1/prompt`
- **THEN** server MUST execute prompt and return structured response

### Requirement: OpenAI-compatible routes

The server MUST provide OpenAI-compatible endpoints for model list and chat completions.

#### Scenario: OpenAI chat completion

- **WHEN** client POSTs `/openai/v1/chat/completions`
- **THEN** server MUST return OpenAI-compatible response shape

### Requirement: Anthropic-compatible routes

The server MUST provide Anthropic-compatible message endpoint.

#### Scenario: Anthropic messages

- **WHEN** client POSTs `/anthropic/v1/messages`
- **THEN** server MUST return Anthropic-compatible response shape

### Requirement: API discoverability

The server MUST expose API schema and health endpoints.

#### Scenario: OpenAPI schema

- **WHEN** client requests `/openapi.json`
- **THEN** server MUST return machine-readable API schema

#### Scenario: Health check

- **WHEN** client requests `/health`
- **THEN** server MUST return service health status
