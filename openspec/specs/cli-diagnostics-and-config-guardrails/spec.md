# cli-diagnostics-and-config-guardrails Specification

## Purpose

TBD - created by archiving change v05-cli-diagnostics-guardrails. Update Purpose after archive.

## Requirements

### Requirement: [Feature #1] `clinvk doctor` aggregated self-check

The system MUST provide a `clinvk doctor` command to run aggregated checks on runtime environment and key configuration.

#### Scenario: Basic self-check

- **WHEN** the user runs `clinvk doctor`
- **THEN** the system MUST check backend availability, configuration validity, and session store accessibility
- **THEN** each check item MUST output `pass/warn/fail` status and remediation guidance

#### Scenario: JSON output contract

- **WHEN** the user runs `clinvk doctor --json`
- **THEN** the output MUST include stable fields and `summary`
- **THEN** the output MUST include a `schema_version` field

### Requirement: [Feature #8] `clinvk config lint` static validation

The system MUST provide a `config lint` command and reuse the unified configuration validator.

#### Scenario: Default-path validation

- **WHEN** the user runs `clinvk config lint`
- **THEN** the system MUST validate default configuration using existing config-loading precedence
- **THEN** the system MUST return all validation errors instead of stopping at the first error

#### Scenario: Target file and error semantics

- **WHEN** the user runs `clinvk config lint --config <path>`
- **THEN** the system MUST validate the specified file
- **THEN** if the file is unreadable or missing, the command MUST return a non-zero exit code with a clear error message

### Requirement: [Feature #7] `serve` startup security status summary

The system MUST output key security status at service startup and flag high-risk combinations.

#### Scenario: Startup summary output

- **WHEN** `clinvk serve` startup completes
- **THEN** logs MUST include auth, rate limit, metrics, trusted proxies, bind host, and CORS status

#### Scenario: High-risk warning

- **WHEN** a high-risk combination appears (for example, external bind with no auth)
- **THEN** the system MUST output an explicit warning and remediation guidance
