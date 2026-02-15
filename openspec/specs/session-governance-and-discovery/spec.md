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

### Requirement: [Feature #12] Idempotent tag management

The system MUST support session tag add/remove operations with idempotent behavior.

#### Scenario: Repeated tag mutation

- **WHEN** the same tag is added or removed repeatedly on the same session
- **THEN** the operation MUST be idempotent (no duplicate writes, no spurious errors)

#### Scenario: Invalid tag input

- **WHEN** a tag violates validation constraints
- **THEN** the system MUST reject the operation with explicit validation error details

### Requirement: [Feature #12] Tag-based filtering

The system MUST support filtering sessions by tags.

#### Scenario: Filter by tag

- **WHEN** users filter sessions by tag criteria
- **THEN** the system MUST return only sessions that satisfy the filter

### Requirement: [Feature #12] Keyword search across core fields

The system MUST support keyword search on `id`, `title`, and `initial_prompt`.

#### Scenario: Keyword hit ordering

- **WHEN** keyword search returns multiple results
- **THEN** results MUST be sorted by `last_used desc` with deterministic tie-break rules

#### Scenario: Keyword miss

- **WHEN** no session matches the keyword
- **THEN** the system MUST return an empty collection instead of an error

### Requirement: [Feature #12] Batch operation traceability

The system MUST provide traceable output for batch tag/search-related operations.

#### Scenario: Batch result summary

- **WHEN** a batch tag update or bulk retrieval operation completes
- **THEN** output MUST include affected session count and key identifiers/samples
