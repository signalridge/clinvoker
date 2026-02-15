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
