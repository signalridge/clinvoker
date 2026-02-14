# session-management Specification

## Purpose

Define baseline session lifecycle behavior for persisted conversations, metadata indexing, and CLI session administration.

## Requirements

### Requirement: Session persistence and retrieval

The system MUST persist session records and allow retrieval by session ID.

#### Scenario: Show session

- **WHEN** user runs `clinvk sessions show <id>`
- **THEN** system MUST return session details or a clear not-found error

### Requirement: Session listing and filtering

The system MUST provide session listing with filtering controls.

#### Scenario: List sessions

- **WHEN** user runs `clinvk sessions list`
- **THEN** output MUST include session identifiers, backend, status, and recency metadata

### Requirement: Session deletion and cleanup

The system MUST support deleting individual sessions and bulk cleanup by age/retention.

#### Scenario: Delete by id

- **WHEN** user runs `clinvk sessions delete <id>`
- **THEN** target session MUST be removed from storage and index

#### Scenario: Clean old sessions

- **WHEN** user runs `clinvk sessions clean --older-than <days>`
- **THEN** only sessions older than threshold MUST be removed

### Requirement: Cross-process storage safety

Session store operations MUST remain safe under concurrent process access.

#### Scenario: Concurrent writes

- **WHEN** multiple processes write sessions concurrently
- **THEN** file locking and atomic write logic MUST prevent index/session corruption
