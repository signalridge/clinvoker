# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- CONTRIBUTING.md with contribution guidelines
- CHANGELOG.md following Keep a Changelog format
- SECURITY.md with security policy
- docs/ARCHITECTURE.md with system architecture documentation
- GitHub Dependabot configuration for automated dependency updates
- Enhanced CI/CD with Go module caching and SARIF security reports
- Security workflow with Trivy scanning and SBOM generation

### Changed

- Improved CI workflow with better caching and security scanning

## [0.2.0] - 2025-01-XX

### Added

- HTTP API server with OpenAI and Anthropic compatible endpoints
- Unified output format support (text, json, stream-json)
- Session resumption with `--continue` flag
- Unified error handling across all backends

### Changed

- Renamed binary from `clinvoker` to `clinvk`
- Improved configuration cascade handling

## [0.1.0] - 2025-01-XX

### Added

- Initial release
- Multi-backend support (Claude Code, Codex CLI, Gemini CLI)
- Session persistence and management
- Parallel execution with `parallel` command
- Backend comparison with `compare` command
- Chain execution with `chain` command
- Configuration file support (~/.clinvk/config.yaml)
- Cross-platform support (Linux, macOS, Windows)
- Homebrew, Scoop, AUR, deb, and rpm package distribution

[Unreleased]: https://github.com/signalridge/clinvoker/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/signalridge/clinvoker/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/signalridge/clinvoker/releases/tag/v0.1.0
