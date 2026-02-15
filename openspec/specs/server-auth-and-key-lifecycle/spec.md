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

### Requirement: [Feature #11] Key metadata contract

The system MUST provide key metadata without exposing plaintext secret material.

#### Scenario: Metadata query

- **WHEN** an operator queries key metadata
- **THEN** the response MUST include lifecycle fields (status/timestamps/source) and sanitized key identifier
- **THEN** the response MUST NOT expose plaintext secret values

### Requirement: [Feature #11] Rollable rotation workflow

The system MUST support key rotation with a transition window and rollback capability.

#### Scenario: Rotation transition window

- **WHEN** rotation is initiated
- **THEN** old and new keys MUST both be accepted during the configured transition window
- **THEN** old keys MUST be revocable after transition completes

#### Scenario: Rollback in transition window

- **WHEN** rotation issues are detected during transition window
- **THEN** the system MUST allow rollback to the previous active key set
- **THEN** rollback action MUST be audited with reason

### Requirement: [Feature #11] Lifecycle auditability

The system MUST emit structured audit events for key lifecycle operations.

#### Scenario: Lifecycle operation audit

- **WHEN** create, rotate, revoke, or rollback occurs
- **THEN** audit events MUST include operator, operation type, target key identifier, timestamp, and result

### Requirement: [Feature #11] Legacy source compatibility

The system MUST preserve auth source compatibility while introducing lifecycle controls.

#### Scenario: Existing env/gopass source usage

- **WHEN** keys are sourced from `CLINVK_API_KEYS` or `api_keys_gopass_path`
- **THEN** auth behavior MUST remain compatible with existing loading semantics
