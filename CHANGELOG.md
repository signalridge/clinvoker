# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0](https://github.com/signalridge/clinvoker/compare/v0.1.0...v0.2.0) (2026-01-31)


### 🚀 Features

* add HTTP server with OpenAI/Anthropic-compatible APIs and rename to clinvk ([#4](https://github.com/signalridge/clinvoker/issues/4)) ([810d7c8](https://github.com/signalridge/clinvoker/commit/810d7c8aa9e96c3a864ca23e9e9a390b2b13f4ac))
* add HTTP server with OpenAI/Anthropic-compatible APIs and rename to clinvk ([#4](https://github.com/signalridge/clinvoker/issues/4)) ([72744d7](https://github.com/signalridge/clinvoker/commit/72744d7722bb9fbf1795ea91c8a0a340e4ea8843))
* add unified output formats and session resumption support ([#5](https://github.com/signalridge/clinvoker/issues/5)) ([2abe491](https://github.com/signalridge/clinvoker/commit/2abe491b6e47f66b859706a9c5c598208447cab5))
* add unified output formats and session resumption support ([#5](https://github.com/signalridge/clinvoker/issues/5)) ([1719a60](https://github.com/signalridge/clinvoker/commit/1719a602ccfc16a94a58bf59fa5618fdd6f638fe))
* **app:** enable multi-backend workflows and safer sessions ([a204b1c](https://github.com/signalridge/clinvoker/commit/a204b1c571e47ccbf7b80495c6118777b8040aa7))
* **backend:** add unified error handling across all backends ([#6](https://github.com/signalridge/clinvoker/issues/6)) ([9556630](https://github.com/signalridge/clinvoker/commit/9556630afc10a37be357701bd9b501c1e61b43a8))
* **backend:** add unified error handling across all backends ([#6](https://github.com/signalridge/clinvoker/issues/6)) ([8321e07](https://github.com/signalridge/clinvoker/commit/8321e07d9234429f7109f835eaa2dfcd83779dd5))
* **ci:** add package installation verification tests ([#25](https://github.com/signalridge/clinvoker/issues/25)) ([16f760f](https://github.com/signalridge/clinvoker/commit/16f760f170ff002b238ecbe3e1ff4abfa9bf6085))
* comprehensive project improvements with unified error handling and session management ([#8](https://github.com/signalridge/clinvoker/issues/8)) ([65f5d37](https://github.com/signalridge/clinvoker/commit/65f5d37b6fe9c649e551f6c47979f2b19b00ad6e))
* comprehensive project improvements with unified error handling and session management ([#8](https://github.com/signalridge/clinvoker/issues/8)) ([08ec696](https://github.com/signalridge/clinvoker/commit/08ec696face756e36d41dcccaf49ed3c7afccf94))
* initial implementation of clinvoker ([988b617](https://github.com/signalridge/clinvoker/commit/988b6177d9b70b08c2b0950833dc974a0165f0dd))
* **server:** add ephemeral mode for stateless API requests ([#10](https://github.com/signalridge/clinvoker/issues/10)) ([6e2fa57](https://github.com/signalridge/clinvoker/commit/6e2fa5767387d6a344889dd44de8f4e3196a05d0))
* **server:** add ephemeral mode for stateless API requests ([#10](https://github.com/signalridge/clinvoker/issues/10)) ([8c08a90](https://github.com/signalridge/clinvoker/commit/8c08a9064d6bf8cd745d78260b3f842338a8dc87))
* **server:** add security and observability improvements ([#30](https://github.com/signalridge/clinvoker/issues/30)) ([2892dab](https://github.com/signalridge/clinvoker/commit/2892dab4cd9e21dd14826b3ab3f71d82939f011a))
* **server:** security, observability, and reliability improvements ([#32](https://github.com/signalridge/clinvoker/issues/32)) ([70a478c](https://github.com/signalridge/clinvoker/commit/70a478c057ff029accc9dc97ebafbad12e4dacb6))


### 🐛 Bug Fixes

* add .claude/ and site/ to gitignore ([#19](https://github.com/signalridge/clinvoker/issues/19)) ([6135dde](https://github.com/signalridge/clinvoker/commit/6135ddec43b92a77bd5d3ecb4bf24b34b03a441a))
* add pymdownx.emoji extension and replace grid cards with tables ([60f5c08](https://github.com/signalridge/clinvoker/commit/60f5c08b417541830d2df6bee38f1ca37a6d75a2))
* address code review issues and improve server executor ([#11](https://github.com/signalridge/clinvoker/issues/11)) ([e6ee5a2](https://github.com/signalridge/clinvoker/commit/e6ee5a2bacd1864833e6354a84ea6f6c582d64e5))
* address code review issues and improve server executor ([#11](https://github.com/signalridge/clinvoker/issues/11)) ([448c4a4](https://github.com/signalridge/clinvoker/commit/448c4a433a2df693ef94fb5cb7566944d5c1d030))
* **ci:** improve release workflow and add PR title validation ([#46](https://github.com/signalridge/clinvoker/issues/46)) ([246fa04](https://github.com/signalridge/clinvoker/commit/246fa0445ecaa848c3412de4ac9409289cb324d7))
* **ci:** resolve CI pipeline failures ([436c0b3](https://github.com/signalridge/clinvoker/commit/436c0b32b48d0e8da9f6dee4a7947b872edbfa01))
* **ci:** resolve CI pipeline failures ([1a28c40](https://github.com/signalridge/clinvoker/commit/1a28c409f4707673a1b4e8e01e7c63723be35aae))
* **ci:** use ssh-keyscan instead of hardcoded AUR host keys ([#22](https://github.com/signalridge/clinvoker/issues/22)) ([ff4e1c9](https://github.com/signalridge/clinvoker/commit/ff4e1c9e59b004607e4885e25dfe12af598b00b0))
* **docker:** use separate Dockerfile for GoReleaser ([#23](https://github.com/signalridge/clinvoker/issues/23)) ([2a55fbd](https://github.com/signalridge/clinvoker/commit/2a55fbd1d7e16c154e248970e2dcc9012081e21f))
* **docs:** correct broken reference links to API documentation ([#38](https://github.com/signalridge/clinvoker/issues/38)) ([01084b3](https://github.com/signalridge/clinvoker/commit/01084b310e8592cdcd68b0433fb0726090c335d6))
* **docs:** correct broken relative links in integration guides ([#40](https://github.com/signalridge/clinvoker/issues/40)) ([75b2a75](https://github.com/signalridge/clinvoker/commit/75b2a753587c8b6af8b0cd5aa2a339d720a8efda))
* **docs:** correct mermaid code block endings ([#34](https://github.com/signalridge/clinvoker/issues/34)) ([711eb4c](https://github.com/signalridge/clinvoker/commit/711eb4c435f61d8809e546859cf08bbd413de809))
* **docs:** remove language identifiers from code block endings ([#35](https://github.com/signalridge/clinvoker/issues/35)) ([71f0ff9](https://github.com/signalridge/clinvoker/commit/71f0ff94d504258bb4704375b3f376c73af144ce))
* honor output format config defaults and harden tests ([#14](https://github.com/signalridge/clinvoker/issues/14)) ([4296403](https://github.com/signalridge/clinvoker/commit/4296403ff796f6965f42ebf9991a77d988eb39b5))
* honor output format config defaults and harden tests ([#14](https://github.com/signalridge/clinvoker/issues/14)) ([51ca14d](https://github.com/signalridge/clinvoker/commit/51ca14d29f9313795c11e98d31417bd1091a4e8d))
* improve release notes and fix container verification ([#24](https://github.com/signalridge/clinvoker/issues/24)) ([6f6cd77](https://github.com/signalridge/clinvoker/commit/6f6cd77768125abda78473ed552beecaafe1a3e5))


### 🔧 Refactoring

* server architecture and add SSE streaming support ([#12](https://github.com/signalridge/clinvoker/issues/12)) ([2cb82e2](https://github.com/signalridge/clinvoker/commit/2cb82e2f29c7ed2b35897e5da358462513f36f4c))
* server architecture and add SSE streaming support ([#12](https://github.com/signalridge/clinvoker/issues/12)) ([31d8399](https://github.com/signalridge/clinvoker/commit/31d839910c2b36615c29f16809cfa82c200b2ad7))


### 📦 Dependencies

* **actions:** bump the actions group with 7 updates ([#9](https://github.com/signalridge/clinvoker/issues/9)) ([06098a6](https://github.com/signalridge/clinvoker/commit/06098a66104dd502850162b9b849ed963898a482))
* **actions:** bump the actions group with 7 updates ([#9](https://github.com/signalridge/clinvoker/issues/9)) ([01b058a](https://github.com/signalridge/clinvoker/commit/01b058a7928db84b6957173d89f6230124ad656a))

## [Unreleased]

### Added

- Cross-process file locking for session store (Unix: `syscall.Flock`, Windows: `LockFileEx`)
- Request body size limiting middleware with streaming detection and response buffering
- Distributed tracing support with W3C Trace Context propagation and pluggable exporters (logging, OTLP)
- Session index staleness detection based on file modification times
- Stream error handling with user-friendly error events for size limit exceeded
- Auto-cleanup of old sessions on server startup based on retention settings
- Trusted proxy support for rate limiting behind reverse proxies
- Prometheus metrics endpoint (`/metrics`) with request counts, latencies, rate limit hits, and session counts
- CORS configuration options (`cors_allowed_origins`, `cors_allow_credentials`, `cors_max_age`)
- Working directory restrictions (`allowed_workdir_prefixes`, `blocked_workdir_prefixes`)

### Changed

- Improved Windows compatibility with platform-specific path handling
- Enhanced server security with multiple new middleware layers

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
