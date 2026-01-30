# Changelog

All notable changes to clinvk are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- MkDocs documentation site with Material theme
- Bilingual support (English and Chinese)
- Comprehensive user guide and reference documentation

### Changed

- Documentation restructured into getting-started, user-guide, server, reference, development, and appendix sections

## [0.1.0] - 2025-01-27

### Added

- Initial release
- Multi-backend support (Claude Code, Codex CLI, Gemini CLI)
- Unified configuration options across backends
- Session persistence and resume capability
- Parallel task execution with fail-fast support
- Backend comparison feature
- Chain execution for sequential pipelines
- HTTP API server with three API styles:
  - Custom REST API (`/api/v1/`)
  - OpenAI-compatible API (`/openai/v1/`)
  - Anthropic-compatible API (`/anthropic/v1/`)
- Configuration cascade (CLI → env vars → config file → defaults)
- Cross-platform support (Linux, macOS, Windows)
- Ephemeral (stateless) mode for one-off queries

### Backends

- Claude Code backend with approval and sandbox modes
- Codex CLI backend
- Gemini CLI backend

### Commands

- `clinvk [prompt]` - Execute prompt
- `clinvk resume` - Resume session
- `clinvk sessions` - Manage sessions
- `clinvk config` - Manage configuration
- `clinvk parallel` - Parallel execution
- `clinvk compare` - Backend comparison
- `clinvk chain` - Chain execution
- `clinvk serve` - HTTP server
- `clinvk version` - Version info

---

## Release Notes Format

Each release includes:

- **Added** - New features
- **Changed** - Changes to existing functionality
- **Deprecated** - Features to be removed in future
- **Removed** - Removed features
- **Fixed** - Bug fixes
- **Security** - Security improvements

## Upgrade Guide

### From 0.x to 1.0 (Future)

Upgrade notes will be added here when breaking changes occur.

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| 0.1.0 | 2025-01-27 | Initial release |

## Links

- [GitHub Releases](https://github.com/signalridge/clinvoker/releases)
- [GitHub Commits](https://github.com/signalridge/clinvoker/commits/main)
