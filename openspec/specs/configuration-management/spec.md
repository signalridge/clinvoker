# configuration-management Specification

## Purpose

Define current configuration behavior for CLI and server, including source loading, runtime overrides, supported config commands, and validation APIs.

## Requirements

### Requirement: Configuration loading and runtime overrides

The system MUST load configuration from defaults, config file, and environment, and apply CLI runtime overrides for supported flags.

#### Scenario: Default backend resolution

- **WHEN** `default_backend` is set via config file or `CLINVK_BACKEND` and no `--backend` flag is provided
- **THEN** prompt execution MUST use the effective configured backend

#### Scenario: CLI flag override

- **WHEN** `--backend`, `--model`, `--output-format`, or `--dry-run` flags are explicitly provided
- **THEN** command runtime behavior MUST use the flag values over loaded configuration

### Requirement: Config CLI management commands

The CLI MUST support configuration introspection and mutation commands.

#### Scenario: Show config

- **WHEN** user runs `clinvk config show`
- **THEN** command MUST print effective configuration

#### Scenario: Set config key

- **WHEN** user runs `clinvk config set <key> <value>`
- **THEN** command MUST write the key/value using nested key notation semantics supported by the config store
- **AND** the updated config MUST be persisted to disk

### Requirement: Custom config file path support

The system MUST support alternative config file paths.

#### Scenario: Custom config path

- **WHEN** user runs command with `--config <path>`
- **THEN** command MUST load and apply configuration from specified path

### Requirement: Explicit configuration validation API

The configuration package MUST provide validation helpers that return actionable field-level errors when invoked.

#### Scenario: Invalid server/rate-limit fields

- **WHEN** validation runs against configuration containing invalid server/rate-limit settings
- **THEN** validation MUST return explicit field-level errors describing invalid values
