<div align="center">

# clinvoker

**Multi-backend AI CLI with OpenAI-compatible API server**

[![CI](https://img.shields.io/github/actions/workflow/status/signalridge/clinvoker/ci.yaml?logo=github)](https://github.com/signalridge/clinvoker/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/signalridge/clinvoker)](https://github.com/signalridge/clinvoker/releases)
[![Go Report](https://img.shields.io/badge/Go_Report-A+-00ADD8?logo=go&logoColor=white)](https://goreportcard.com/report/github.com/signalridge/clinvoker)
[![License](https://img.shields.io/badge/License-MIT-blue)](https://opensource.org/licenses/MIT)

[Documentation](https://signalridge.github.io/clinvoker/) · [Installation](#-installation) · [Quick Start](#-quick-start)

</div>

---

## ✨ Highlights

- **Multi-Backend** — Seamlessly switch between Claude Code, Codex CLI, and Gemini CLI
- **OpenAI-Compatible API** — Drop-in replacement for OpenAI/Anthropic API endpoints
- **Session Management** — Persist and resume conversations across sessions
- **Parallel Execution** — Run tasks concurrently across multiple backends
- **Cross-Platform** — Native binaries for Linux, macOS, and Windows

---

## 📑 Table of Contents

- [✨ Highlights](#-highlights)
- [📑 Table of Contents](#-table-of-contents)
- [🚀 Quick Start](#-quick-start)
- [📦 Installation](#-installation)
- [💡 Usage](#-usage)
  - [Basic Commands](#basic-commands)
  - [Session Management](#session-management)
- [🌐 HTTP API Server](#-http-api-server)
  - [API Endpoints](#api-endpoints)
- [⚙️ Configuration](#️-configuration)
- [📖 Documentation](#-documentation)
- [🤝 Contributing](#-contributing)
- [📝 License](#-license)

---

## 🚀 Quick Start

```bash
# Install via Homebrew
brew install signalridge/tap/clinvk

# Run with default backend
clinvk "fix the bug in auth.go"

# Start HTTP API server
clinvk serve --port 8080
```

---

## 📦 Installation

| Platform | Method | Command |
|----------|--------|---------|
| macOS/Linux | Homebrew | `brew install signalridge/tap/clinvk` |
| Windows | Scoop | `scoop bucket add signalridge https://github.com/signalridge/scoop-bucket && scoop install clinvk` |
| Arch Linux | AUR | `yay -S clinvk-bin` |
| NixOS | Flake | `nix run github:signalridge/clinvoker` |
| Docker | GHCR | `docker run ghcr.io/signalridge/clinvk:latest` |
| Debian/Ubuntu | deb | Download from [Releases](https://github.com/signalridge/clinvoker/releases) |
| Fedora/RHEL | rpm | Download from [Releases](https://github.com/signalridge/clinvoker/releases) |
| Alpine | apk | Download from [Releases](https://github.com/signalridge/clinvoker/releases) |
| Go | go install | `go install github.com/signalridge/clinvoker/cmd/clinvk@latest` |

See [Installation Guide](https://signalridge.github.io/clinvoker/getting-started/installation/) for detailed instructions.

---

## 💡 Usage

### Basic Commands

```bash
# Run with default backend
clinvk "explain this code"

# Use specific backend
clinvk -b codex "implement user registration"
clinvk -b gemini "review this PR"

# Resume last session
clinvk resume --last "continue where we left off"

# Compare responses across backends
clinvk compare --all-backends "explain this algorithm"
```

### Session Management

```bash
# List all sessions
clinvk sessions list

# Resume a specific session
clinvk resume <session-id>

# Export session history
clinvk sessions export <session-id> -o history.json
```

---

## 🌐 HTTP API Server

Start an OpenAI/Anthropic-compatible API server:

```bash
# Start server on port 8080
clinvk serve --port 8080

# With specific backend
clinvk serve --port 8080 --backend claude
```

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `POST /v1/chat/completions` | OpenAI-compatible chat completions |
| `POST /v1/messages` | Anthropic-compatible messages |
| `GET /v1/models` | List available models |
| `GET /health` | Health check |

---

## ⚙️ Configuration

```bash
# Show current configuration
clinvk config show

# Set default backend
clinvk config set default_backend claude

# Configure API keys
export ANTHROPIC_API_KEY="sk-..."
export OPENAI_API_KEY="sk-..."
export GOOGLE_API_KEY="..."
```

---

## 📖 Documentation

Full documentation: **[signalridge.github.io/clinvoker](https://signalridge.github.io/clinvoker/)**

| Section | Description |
|---------|-------------|
| [Getting Started](https://signalridge.github.io/clinvoker/getting-started/) | Installation and first steps |
| [User Guide](https://signalridge.github.io/clinvoker/user-guide/) | Detailed usage instructions |
| [HTTP API](https://signalridge.github.io/clinvoker/server/) | API server documentation |
| [Reference](https://signalridge.github.io/clinvoker/reference/) | CLI reference and configuration |

---

## 🤝 Contributing

Contributions welcome! See [Contributing Guide](https://signalridge.github.io/clinvoker/development/contributing/).

```bash
# Clone the repo
git clone https://github.com/signalridge/clinvoker.git
cd clinvoker

# Run tests
go test ./...

# Build
go build ./cmd/clinvk
```

---

## 📝 License

MIT License - see [LICENSE](LICENSE).

---

<div align="center">

**[Documentation](https://signalridge.github.io/clinvoker/)** · **[Report Bug](https://github.com/signalridge/clinvoker/issues)** · **[Request Feature](https://github.com/signalridge/clinvoker/issues)**

</div>
