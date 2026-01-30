# Installation

Install `clinvk` first, then install the backend CLIs you want to use.

## Install clinvk

### Homebrew (macOS/Linux)

```bash
brew install signalridge/tap/clinvk
```

### Scoop (Windows)

```bash
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket
scoop install clinvk
```

### AUR (Arch)

```bash
yay -S clinvk-bin
```

### Nix

```bash
nix run github:signalridge/clinvoker
```

### Docker

```bash
docker run ghcr.io/signalridge/clinvk:latest
```

### Go install

```bash
go install github.com/signalridge/clinvoker/cmd/clinvk@latest
```

## Install backend CLIs

`clinvk` calls external CLIs. Install any of these and ensure they are in `PATH`:

- **Claude Code**: `claude`
- **Codex CLI**: `codex`
- **Gemini CLI**: `gemini`

Follow each CLI’s official installation/auth guide. (Their API keys and login are managed by those CLIs, not by clinvk.)

## Verify

```bash
clinvk version
clinvk "hello"
```

If a backend is missing, clinvk will report it when you use it.

## Optional: config file

Default config path:

```
~/.clinvk/config.yaml
```

See [Configuration](configuration.md) for a starter config.
