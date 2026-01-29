# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0-alpha] - 2025-01-30

### Added

- Initial alpha release
- Multi-backend support (Claude Code, Codex CLI, Gemini CLI)
- HTTP API server with OpenAI and Anthropic compatible endpoints
- Unified output format support (text, json, stream-json)
- Session persistence and management
- Session resumption with `--continue` flag
- Parallel execution with `parallel` command
- Backend comparison with `compare` command
- Chain execution with `chain` command
- Unified error handling across all backends
- Configuration file support (~/.clinvk/config.yaml)
- Cross-platform support (Linux, macOS, Windows)
- Homebrew, Scoop, AUR, deb, and rpm package distribution
- CONTRIBUTING.md with contribution guidelines
- CHANGELOG.md following Keep a Changelog format
- SECURITY.md with security policy
- docs/ARCHITECTURE.md with system architecture documentation
- GitHub Dependabot configuration for automated dependency updates
- Enhanced CI/CD with Go module caching and SARIF security reports
- Security workflow with Trivy scanning and SBOM generation

### Changed

- Renamed binary from `clinvoker` to `clinvk`
- Improved configuration cascade handling

[Unreleased]: https://github.com/signalridge/clinvoker/compare/v0.1.0-alpha...HEAD
[0.1.0-alpha]: https://github.com/signalridge/clinvoker/releases/tag/v0.1.0-alpha
