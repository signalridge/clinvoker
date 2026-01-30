# 安装

先安装 `clinvk`，再安装你需要的后端 CLI。

## 安装 clinvk

### Homebrew（macOS/Linux）

```bash
brew install signalridge/tap/clinvk
```

### Scoop（Windows）

```bash
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket
scoop install clinvk
```

### AUR（Arch）

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

### Go 安装

```bash
go install github.com/signalridge/clinvoker/cmd/clinvk@latest
```

## 安装后端 CLI

`clinvk` 依赖外部 CLI，请安装并确保在 `PATH` 中：

- **Claude Code**：`claude`
- **Codex CLI**：`codex`
- **Gemini CLI**：`gemini`

各 CLI 的认证与 API Key 配置由它们自身处理，clinvk 不负责。

## 验证

```bash
clinvk version
clinvk "hello"
```

如果后端缺失，clinvk 会在使用时提示。

## 可选：配置文件

默认配置路径：

```
~/.clinvk/config.yaml
```

参见 [配置](configuration.md)。
