# backend-registry-and-adapters Specification

## Purpose

Define current backend abstraction and registry behavior for Claude Code, Codex CLI, and Gemini CLI, including default registration, availability checks, execution guardrails, and stderr handling.

## Requirements

### Requirement: Unified backend interface contract

The system MUST expose a stable backend interface consumed by CLI and server execution paths.

#### Scenario: Backend capability surface

- **WHEN** a backend implementation is registered
- **THEN** it MUST implement name, availability check, command construction, resume command, output parsing, and stderr strategy behaviors

### Requirement: Default backend registration

The default registry MUST register built-in adapters for `claude`, `codex`, and `gemini`.

#### Scenario: Global registry initialization

- **WHEN** the application starts
- **THEN** the global backend registry MUST include `claude`, `codex`, and `gemini`

### Requirement: Backend listing includes availability metadata

Backend listing operations MUST include availability metadata for each listed backend.

#### Scenario: Service backend listing

- **WHEN** the API/service layer lists backends
- **THEN** each listed backend entry MUST include both backend name and availability status

### Requirement: Availability-aware execution behavior

Execution paths MUST apply explicit unavailable-backend behavior by command type.

#### Scenario: Single-backend execution path

- **WHEN** prompt, resume, chain-step, or parallel-task execution selects an unavailable backend
- **THEN** execution MUST fail with a backend-unavailable error

#### Scenario: Multi-backend compare execution path

- **WHEN** compare execution includes unavailable backends
- **THEN** unavailable backends MUST be skipped with warning output
- **AND** compare execution MUST fail if no selected backend is available

### Requirement: Stderr handling strategy

Backends that emit non-user noise on stderr MUST be able to separate stderr handling.

#### Scenario: Separate stderr enabled

- **WHEN** backend adapter marks `SeparateStderr` as true
- **THEN** execution pipeline MUST isolate stderr for filtering and clearer user-facing errors
