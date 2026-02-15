# execution-insights-and-controls Specification

## Purpose

TBD - created by archiving change v05-execution-insights-controls. Update Purpose after archive.

## Requirements

### Requirement: [Feature #3] chain step-level timeout

The system MUST support `timeout_secs` configuration at the chain step level.

#### Scenario: Timeout failure and chain stop

- **WHEN** a step runs longer than `timeout_secs`
- **THEN** that step MUST be marked as timed-out failure and record the reason
- **THEN** when `stop_on_failure=true`, the system MUST stop subsequent steps

#### Scenario: Coexisting global and step timeout

- **WHEN** both global timeout and step timeout are configured
- **THEN** the system MUST use the shorter effective timeout

### Requirement: [Feature #2] compare summary scoring

The system MUST provide compare summary scoring and publish ranking criteria.

#### Scenario: Text summary

- **WHEN** compare runs in non-JSON mode
- **THEN** output MUST include latency, status, output_length, and score
- **THEN** output MUST describe sorting rules

#### Scenario: JSON summary

- **WHEN** compare runs with `--json`
- **THEN** output MUST include score input dimensions and ranking fields

### Requirement: [Feature #9] optional prompt usage display

The system MUST support on-demand usage/token display in text mode while preserving default compatibility.

#### Scenario: Explicit display

- **WHEN** the user enables the usage display flag
- **THEN** output MUST include input/output/total token information

#### Scenario: Missing usage

- **WHEN** backend does not return usage
- **THEN** output MUST show `unknown` instead of misleading zero values

### Requirement: [Feature #10] Hierarchical retry policy

The system MUST support retry policy configuration at global, backend, and command-type levels.

#### Scenario: Policy precedence resolution

- **WHEN** all three levels define retry policy fields
- **THEN** the effective policy MUST be resolved using precedence `command-type > backend > global`

### Requirement: [Feature #10] Retry eligibility and backoff

The system MUST retry only eligible transient failures and apply configured backoff behavior.

#### Scenario: Retryable transient failure

- **WHEN** execution fails with a retryable error and retry budget remains
- **THEN** the system MUST schedule the next attempt using configured backoff and jitter

#### Scenario: Non-retryable failure

- **WHEN** execution fails with a non-retryable error
- **THEN** the system MUST terminate immediately without additional retry attempts

### Requirement: [Feature #10] Non-idempotent safety default

The system MUST protect non-idempotent commands from implicit retries.

#### Scenario: Non-idempotent default behavior

- **WHEN** a command is non-idempotent and no explicit override is set
- **THEN** automatic retries MUST remain disabled

#### Scenario: Explicit non-idempotent override

- **WHEN** non-idempotent retry is explicitly enabled
- **THEN** the system MUST emit a clear risk warning in output

### Requirement: [Feature #10] Retry and timeout combined budgets

The system MUST enforce both retry attempt budget and timeout budget.

#### Scenario: Timeout budget exhausted first

- **WHEN** remaining timeout budget is exhausted before another retry attempt
- **THEN** the system MUST terminate with timeout-exhausted reason

#### Scenario: Retry budget exhausted first

- **WHEN** max attempts are exhausted while timeout budget remains
- **THEN** the system MUST terminate with retry-budget-exhausted reason

### Requirement: [Feature #10] Retry observability contract

The system MUST expose stable retry observability fields in command output.

#### Scenario: Successful convergence after retries

- **WHEN** execution succeeds after one or more retry attempts
- **THEN** output MUST include total attempts used and a success termination reason

#### Scenario: Terminal failure

- **WHEN** execution terminates after retries
- **THEN** output MUST include attempts used, termination reason, and last error category
