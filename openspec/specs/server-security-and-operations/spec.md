# server-security-and-operations Specification

## Purpose

Define baseline operational safety controls for server auth, rate limiting, request constraints, and observability endpoints.

## Requirements

### Requirement: Optional API key authentication

Server auth MUST be enforceable via configured API keys and remain optional when no keys are configured.

#### Scenario: Auth enabled

- **WHEN** API keys are configured
- **THEN** protected requests without valid key MUST return unauthorized response

#### Scenario: Auth disabled by configuration state

- **WHEN** no API keys are configured
- **THEN** requests MAY pass without API key (compatibility mode)

### Requirement: API key source compatibility

The server MUST load keys from environment and gopass-backed configuration.

#### Scenario: Environment key source

- **WHEN** `CLINVK_API_KEYS` is set
- **THEN** server MUST load comma-separated keys from environment

#### Scenario: Gopass key source

- **WHEN** `api_keys_gopass_path` is configured
- **THEN** server MUST load keys from gopass path

### Requirement: Rate limiting middleware

The server MUST support configurable per-IP rate limiting.

#### Scenario: Rate limit exceeded

- **WHEN** request rate exceeds configured threshold
- **THEN** server MUST return rate-limit response without invoking handler logic

### Requirement: Metrics endpoint control

The server MUST expose Prometheus metrics endpoint when enabled.

#### Scenario: Metrics enabled

- **WHEN** metrics setting is enabled
- **THEN** `/metrics` MUST provide Prometheus-compatible metrics output
