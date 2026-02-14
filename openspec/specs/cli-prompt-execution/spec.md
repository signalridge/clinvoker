# cli-prompt-execution Specification

## Purpose

Define baseline CLI prompt execution behavior, including default command semantics, output formats, and session continuation.

## Requirements

### Requirement: Root command executes prompt

The root CLI command MUST execute prompt requests without requiring a subcommand.

#### Scenario: Default prompt invocation

- **WHEN** user runs `clinvk "<prompt>"`
- **THEN** the command MUST execute prompt using configured/default backend

### Requirement: Prompt output formats

Prompt execution MUST support `text`, `json`, and `stream-json` output formats.

#### Scenario: JSON output

- **WHEN** output format is `json`
- **THEN** response MUST include structured fields for backend, content, timing, and error state

#### Scenario: Stream JSON output

- **WHEN** output format is `stream-json`
- **THEN** command MUST emit NDJSON unified events until completion

### Requirement: Session continuation support

Prompt execution MUST support continuing previous sessions through CLI continuation semantics.

#### Scenario: Continue recent session

- **WHEN** user invokes continuation mode (`-c` / resume path)
- **THEN** system MUST resume from the most recent resumable session context

### Requirement: Ephemeral and dry-run modes

Prompt execution MUST support ephemeral and dry-run behavior.

#### Scenario: Ephemeral mode

- **WHEN** ephemeral mode is enabled
- **THEN** command MUST execute without persisting session state

#### Scenario: Dry run

- **WHEN** dry-run mode is enabled
- **THEN** command MUST print the command plan and MUST NOT execute backend command
