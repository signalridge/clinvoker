# l4-test-coverage Specification

## Purpose

Define the current multi-layer test coverage baseline, spanning Go unit tests, shell E2E suites, fuzz targets, and targeted race checks.

## Requirements

### Requirement: L4 test coverage matrix

The system MUST provide test coverage across Go unit tests, Shell E2E tests, fuzz tests, and race tests for all feasible, deterministic code paths.

#### Scenario: Test matrix is implemented

- **WHEN** the test suite is executed
- **THEN** Go unit tests, Shell E2E tests, fuzz tests, and race tests MUST be runnable and documented

### Requirement: Shell E2E and MCP stdio coverage

The system MUST include end-to-end shell tests covering CLI command execution paths and MCP transport behavior.

#### Scenario: Shell E2E command coverage

- **WHEN** shell E2E suites run
- **THEN** they MUST exercise prompt, chain, parallel, compare, and mcp command paths with valid and invalid inputs

#### Scenario: MCP stdio tests

- **WHEN** MCP stdio tests run
- **THEN** they MUST validate JSON-RPC request/response handling and error semantics over stdin/stdout

### Requirement: Error and edge case coverage

The system MUST include tests that cover error paths and edge cases for handlers, service logic, and cleanup utilities.

#### Scenario: Error branch coverage

- **WHEN** tests execute error scenarios
- **THEN** REST handler validation errors, MCP invalid params, and cleanup error paths MUST be exercised

### Requirement: Deterministic fuzz and race coverage

The system MUST include bounded fuzz tests and targeted race tests to validate input parsing and concurrency behavior.

#### Scenario: Fuzz execution

- **WHEN** fuzz tests run with time bounds
- **THEN** they MUST cover JSON parsing and input validation surfaces without flakiness

#### Scenario: Race execution

- **WHEN** race tests run on targeted packages
- **THEN** they MUST complete without data race reports
