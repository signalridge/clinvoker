<div align="center">

![header](https://capsule-render.vercel.app/api?type=waving&color=0%3A6366f1,100%3A8b5cf6&height=200&section=header&text=clinvoker&fontSize=48&fontColor=ffffff&fontAlignY=30&desc=Multi-backend%20AI%20CLI%20with%20OpenAI-compatible%20API%20server&descSize=16&descColor=e0e7ff&descAlignY=55&animation=fadeIn)

<p>
  <a href="https://github.com/signalridge/clinvoker/actions/workflows/ci.yaml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/signalridge/clinvoker/ci.yaml?style=for-the-badge&logo=github&label=CI"></a>&nbsp;
  <a href="https://goreportcard.com/report/github.com/signalridge/clinvoker"><img alt="Go Report Card" src="https://img.shields.io/badge/Go_Report-A+-00ADD8?style=for-the-badge&logo=go&logoColor=white"></a>&nbsp;
  <a href="https://github.com/signalridge/clinvoker/releases"><img alt="Release" src="https://img.shields.io/github/v/release/signalridge/clinvoker?style=for-the-badge&logo=github"></a>&nbsp;
  <a href="https://opensource.org/licenses/MIT"><img alt="License" src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge"></a>
</p>

[![Typing SVG](https://readme-typing-svg.demolab.com?font=Fira+Code&weight=600&size=22&pause=1000&color=8B5CF6&center=true&vCenter=true&width=700&lines=One+CLI+for+Claude%2C+Codex%2C+and+Gemini;OpenAI-compatible+HTTP+API+server;Session+management+and+parallel+execution;Cross-platform%3A+Linux%2C+macOS%2C+Windows)](https://git.io/typing-svg)

<p>
  <a href="#-installation"><img alt="Homebrew" src="https://img.shields.io/badge/Homebrew-FBB040?style=flat-square&logo=homebrew&logoColor=black"></a>
  <a href="#-installation"><img alt="Scoop" src="https://img.shields.io/badge/Scoop-00BFFF?style=flat-square&logo=windows&logoColor=white"></a>
  <a href="#-installation"><img alt="AUR" src="https://img.shields.io/badge/AUR-1793D1?style=flat-square&logo=archlinux&logoColor=white"></a>
  <a href="#-installation"><img alt="Nix" src="https://img.shields.io/badge/Nix-5277C3?style=flat-square&logo=nixos&logoColor=white"></a>
  <a href="#-installation"><img alt="Docker" src="https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white"></a>
  <a href="#-installation"><img alt="deb" src="https://img.shields.io/badge/deb-A81D33?style=flat-square&logo=debian&logoColor=white"></a>
  <a href="#-installation"><img alt="rpm" src="https://img.shields.io/badge/rpm-EE0000?style=flat-square&logo=redhat&logoColor=white"></a>
  <a href="#-installation"><img alt="apk" src="https://img.shields.io/badge/apk-0D597F?style=flat-square&logo=alpinelinux&logoColor=white"></a>
  <a href="#-installation"><img alt="Go" src="https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white"></a>
</p>

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

- [🚀 Quick Start](#-quick-start)
- [📦 Installation](#-installation)
- [💡 Usage](#-usage)
- [🌐 HTTP API Server](#-http-api-server)
- [⚙️ Configuration](#️-configuration)
- [📖 Documentation](#-documentation)
- [🤝 Contributing](#-contributing)

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

## 📊 Stats

![Repobeats](https://repobeats.axiom.co/api/embed/640263fe1e0af60e340f1ac257bff68a562c26c6.svg "Repobeats analytics image")

---

## 🙏 Acknowledgments

Inspired by [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI).

---

## 📝 License

MIT License - see [LICENSE](LICENSE).

---

<div align="center">

**[Documentation](https://signalridge.github.io/clinvoker/)** · **[Report Bug](https://github.com/signalridge/clinvoker/issues)** · **[Request Feature](https://github.com/signalridge/clinvoker/issues)**

</div>
