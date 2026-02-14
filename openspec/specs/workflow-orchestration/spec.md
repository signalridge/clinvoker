# workflow-orchestration Specification

## Purpose

Define baseline orchestration behaviors for `parallel`, `chain`, and `compare` commands.

## Requirements

### Requirement: Parallel task execution

The system MUST execute multiple task prompts concurrently with configurable max parallelism.

#### Scenario: File-driven parallel run

- **WHEN** user runs `clinvk parallel --file <tasks.json>`
- **THEN** tasks MUST run with configured/default max parallel workers

#### Scenario: Fail-fast behavior

- **WHEN** fail-fast mode is enabled and one task fails
- **THEN** the system MUST cancel remaining pending tasks and avoid executing additional backend work

### Requirement: Chain sequential pipeline

The system MUST execute chain steps sequentially and support `{{previous}}` templating.

#### Scenario: Previous-output templating

- **WHEN** a step prompt contains `{{previous}}`
- **THEN** system MUST substitute prior step output before execution

### Requirement: Compare same-prompt multi-backend evaluation

The system MUST execute the same prompt against multiple backends and return comparable outputs.

#### Scenario: Compare all backends

- **WHEN** user runs compare with all backends mode
- **THEN** each available backend MUST be invoked and included in final comparison result

### Requirement: Orchestration CLI runs are ephemeral

CLI `parallel`, `chain`, and `compare` runs MUST be ephemeral by default.

#### Scenario: No session persistence

- **WHEN** orchestration commands complete
- **THEN** command MUST NOT persist conversation session state unless explicitly supported in a future capability
