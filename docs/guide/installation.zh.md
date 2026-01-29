# 安装

clinvk 可以通过多种包管理器安装，也可以从源码编译。

## 从发布版安装

从 [GitHub Releases](https://github.com/signalridge/clinvoker/releases) 下载适合您平台的最新版本。

=== "Linux (amd64)"

    ```bash
    curl -LO https://github.com/signalridge/clinvoker/releases/latest/download/clinvk_linux_amd64.tar.gz
    tar xzf clinvk_linux_amd64.tar.gz
    sudo mv clinvk /usr/local/bin/
    ```

=== "macOS (arm64)"

    ```bash
    curl -LO https://github.com/signalridge/clinvoker/releases/latest/download/clinvk_darwin_arm64.tar.gz
    tar xzf clinvk_darwin_arm64.tar.gz
    sudo mv clinvk /usr/local/bin/
    ```

=== "Windows"

    从发布页面下载 `clinvk_windows_amd64.zip` 并解压到您的 PATH 中。

## 包管理器

### Homebrew (macOS/Linux)

```bash
brew install signalridge/tap/clinvk
```

### Scoop (Windows)

```bash
scoop bucket add signalridge https://github.com/signalridge/scoop-bucket
scoop install clinvk
```

### Nix

```bash
# 直接运行
nix run github:signalridge/clinvoker

# 安装到 profile
nix profile install github:signalridge/clinvoker

# 开发环境
nix develop github:signalridge/clinvoker
```

添加到您的 flake：

```nix
{
  inputs.clinvoker.url = "github:signalridge/clinvoker";

  # 使用 overlay
  nixpkgs.overlays = [ clinvoker.overlays.default ];
}
```

### Arch Linux (AUR)

```bash
# 使用 yay
yay -S clinvk-bin

# 或从源码构建
yay -S clinvk
```

### Debian/Ubuntu

```bash
# 从发布页面下载 .deb 包
sudo dpkg -i clinvk_*.deb
```

### RPM 系 (Fedora/RHEL)

```bash
# 从发布页面下载 .rpm 包
sudo rpm -i clinvk_*.rpm
```

## 从源码安装

### 使用 Go

需要 Go 1.24 或更高版本：

```bash
go install github.com/signalridge/clinvoker/cmd/clinvk@latest
```

### 手动构建

```bash
git clone https://github.com/signalridge/clinvoker.git
cd clinvoker
go build -o clinvk ./cmd/clinvk
sudo mv clinvk /usr/local/bin/
```

## 验证安装

安装后，验证 clinvk 是否正常工作：

```bash
clinvk version
```

预期输出：

```
clinvk version v0.x.x
  commit: abc1234
  built:  2025-01-27T00:00:00Z
```

## 后端检测

clinvk 会自动检测 PATH 中可用的后端。检查已检测的后端：

```bash
clinvk config show
```

!!! tip "后端安装"
    如果没有检测到后端，请至少安装一个 AI CLI 工具：

    - [Claude Code](https://claude.ai/claude-code)
    - [Codex CLI](https://github.com/openai/codex-cli)
    - [Gemini CLI](https://github.com/google/gemini-cli)

## 下一步

- [快速开始](quick-start.md) - 运行您的第一个提示
- [配置](../reference/configuration.md) - 自定义设置
