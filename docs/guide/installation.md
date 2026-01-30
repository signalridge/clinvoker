# Installation

clinvk can be installed using various package managers or from source.

## From Release

Download the latest release for your platform from [GitHub Releases](https://github.com/signalridge/clinvoker/releases).

=== "Linux (amd64)"

    ```bash
    VERSION="<version>" # e.g. 0.1.0-alpha
    ASSET="clinvoker_${VERSION}_linux_amd64.tar.gz"
    curl -LO "https://github.com/signalridge/clinvoker/releases/download/v${VERSION}/${ASSET}"
    tar xzf "${ASSET}"
    sudo mv clinvk /usr/local/bin/

```yaml

=== "macOS (arm64)"

    ```bash
    VERSION="<version>" # e.g. 0.1.0-alpha
    ASSET="clinvoker_${VERSION}_darwin_arm64.tar.gz"
    curl -LO "https://github.com/signalridge/clinvoker/releases/download/v${VERSION}/${ASSET}"
    tar xzf "${ASSET}"
    sudo mv clinvk /usr/local/bin/
```

=== "Windows"

    Download `clinvoker_<version>_windows_amd64.zip` from the releases page and extract to your PATH.

## Package Managers

### Homebrew (macOS/Linux)

```bash
brew install signalridge/tap/clinvk
```text

### Scoop (Windows)

```bash
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket
scoop install clinvk
```

### Nix

```bash
# Run directly
nix run github:signalridge/clinvoker

# Install to profile
nix profile install github:signalridge/clinvoker

# Development shell
nix develop github:signalridge/clinvoker
```yaml

Add to your flake:

```nix
{
  inputs.clinvoker.url = "github:signalridge/clinvoker";

  # Use overlay
  nixpkgs.overlays = [ clinvoker.overlays.default ];
}
```

### Arch Linux (AUR)

```bash
# Using yay
yay -S clinvk-bin

# Or build from source
yay -S clinvk
```text

### Debian/Ubuntu

```bash
# Download the .deb package from releases
sudo dpkg -i clinvk_*.deb
```

### RPM-based (Fedora/RHEL)

```bash
# Download the .rpm package from releases
sudo rpm -i clinvk_*.rpm
```bash

## From Source

### Using Go

Requires Go 1.24 or later:

```bash
go install github.com/signalridge/clinvoker/cmd/clinvk@latest
```

### Manual Build

```bash
git clone https://github.com/signalridge/clinvoker.git
cd clinvoker
go build -o clinvk ./cmd/clinvk
sudo mv clinvk /usr/local/bin/
```bash

## Verify Installation

After installation, verify that clinvk is working:

```bash
clinvk version
```

Expected output:

```yaml
clinvk version v0.x.x
  commit: abc1234
  built:  2025-01-27T00:00:00Z
```

## Backend Detection

clinvk automatically detects available backends in your PATH. Check detected backends:

```bash
clinvk config show
```

!!! tip "Backend Installation"
    If no backends are detected, install at least one AI CLI tool:

    - [Claude Code](https://claude.ai/claude-code)
    - [Codex CLI](https://github.com/openai/codex-cli)
    - [Gemini CLI](https://github.com/google/gemini-cli)

## Next Steps

- [Quick Start](quick-start.md) - Run your first prompt
- [Configuration](../reference/configuration.md) - Customize your setup
