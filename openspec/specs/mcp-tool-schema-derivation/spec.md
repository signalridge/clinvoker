# mcp-tool-schema-derivation Specification

## Purpose

Define how MCP tool schemas are derived from existing REST request/response structs so MCP and REST remain structurally aligned.

## Requirements

### Requirement: MCP tool schema derivation

The system MUST derive MCP tool input schemas from existing REST request structs to ensure consistency.

#### Scenario: Schema derivation

- **WHEN** MCP tool schemas are defined or generated
- **THEN** they MUST match the fields and types of the REST request structs in the server handlers

### Requirement: Schema drift prevention

The system MUST keep MCP tool schemas in sync with REST request structs to prevent drift.

#### Scenario: REST request struct changes

- **WHEN** a REST request struct changes
- **THEN** the corresponding MCP tool schema MUST be updated accordingly

### Requirement: MCP tool output schema derivation

The system MUST derive MCP tool output schemas from REST response body structs.

#### Scenario: Output schema derivation

- **WHEN** MCP tool schemas are defined or generated
- **THEN** each tool MUST include an output schema derived from the corresponding response body struct

### Requirement: Schema annotations

The schema generator MUST use struct tags and doc tags to populate schema metadata.

#### Scenario: Schema metadata

- **WHEN** a field has a `json` tag and `doc` tag
- **THEN** the schema MUST use the `json` name and the `doc` value as the field description

#### Scenario: Optional fields

- **WHEN** a field is marked `omitempty` or is a pointer
- **THEN** the schema MUST treat it as optional

#### Scenario: Non-JSON tags

- **WHEN** a field lacks a `json` tag but has a `query` or `path` tag
- **THEN** the schema MUST use that tag name for the field

#### Scenario: Time fields

- **WHEN** a field is `time.Time`
- **THEN** the schema MUST use a string with `format: date-time`

#### Scenario: Duration fields

- **WHEN** a field is `time.Duration`
- **THEN** the schema MUST use a string representation
