# session-governance-and-discovery Specification

## Purpose

TBD - created by archiving change v05-session-list-clean. Update Purpose after archive.

## Requirements

### Requirement: [Feature #4] stable output for `sessions list --json`

The system MUST provide stable and scriptable JSON output for session listing.

#### Scenario: Core fields

- **WHEN** the user runs `clinvk sessions list --json`
- **THEN** each session item MUST include `id`, `backend`, `status`, `last_used`, `model`, and `tags`
- **THEN** output MUST include filtering context fields

#### Scenario: Ordering and pagination

- **WHEN** the same query is run repeatedly
- **THEN** result ordering MUST be stable (default order is `last_used` descending)
- **THEN** output MUST include pagination semantic fields

### Requirement: [Feature #5] side-effect-free preview for `sessions clean --dry-run`

The system MUST provide dry-run mode and guarantee no deletion side effects.

#### Scenario: Preview result

- **WHEN** the user runs `sessions clean --older-than <value> --dry-run`
- **THEN** the system MUST output candidate session count and sample identifiers
- **THEN** the system MUST NOT delete any session data or index records

#### Scenario: Explainable differences

- **WHEN** a real clean is executed after dry-run
- **THEN** if results differ, the system SHOULD provide an explainable reason (for example, concurrent writes)
