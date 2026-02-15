# server-auth-and-key-lifecycle Specification

## Purpose

TBD - created by archiving change v05-auth-request-traceability. Update Purpose after archive.

## Requirements

### Requirement: [Feature #6] auth failure response includes `request_id`

The system MUST return `request_id` in auth failure responses, and keep header/body values consistent.

#### Scenario: missing key

- **WHEN** the request is missing API key and auth is enabled
- **THEN** response body MUST include `request_id`
- **THEN** the request identifier in response header MUST match the body value

#### Scenario: invalid key

- **WHEN** the request API key is invalid
- **THEN** the response MUST return clear error semantics and `request_id`

### Requirement: [Feature #6] stable error structure

The system MUST provide stable, machine-parseable structured error fields for auth failures.

#### Scenario: stable fields

- **WHEN** any auth failure occurs
- **THEN** response MUST include `code`, `message`, and `request_id`
- **THEN** field changes MUST be declared through versioned documentation
