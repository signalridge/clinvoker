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
  <a href="#-安装"><img alt="Homebrew" src="https://img.shields.io/badge/Homebrew-FBB040?style=flat-square&logo=homebrew&logoColor=black"></a>
  <a href="#-安装"><img alt="Scoop" src="https://img.shields.io/badge/Scoop-00BFFF?style=flat-square&logo=windows&logoColor=white"></a>
  <a href="#-安装"><img alt="AUR" src="https://img.shields.io/badge/AUR-1793D1?style=flat-square&logo=archlinux&logoColor=white"></a>
  <a href="#-安装"><img alt="Nix" src="https://img.shields.io/badge/Nix-5277C3?style=flat-square&logo=nixos&logoColor=white"></a>
  <a href="#-安装"><img alt="Docker" src="https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white"></a>
  <a href="#-安装"><img alt="deb" src="https://img.shields.io/badge/deb-A81D33?style=flat-square&logo=debian&logoColor=white"></a>
  <a href="#-安装"><img alt="rpm" src="https://img.shields.io/badge/rpm-EE0000?style=flat-square&logo=redhat&logoColor=white"></a>
  <a href="#-安装"><img alt="apk" src="https://img.shields.io/badge/apk-0D597F?style=flat-square&logo=alpinelinux&logoColor=white"></a>
  <a href="#-安装"><img alt="Go" src="https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white"></a>
</p>

**[English](README.md)** | 简体中文

</div>

---

## ✨ 亮点

- **多后端支持** — 在 Claude Code、Codex CLI 和 Gemini CLI 之间无缝切换
- **OpenAI 兼容 API** — 可直接替代 OpenAI/Anthropic API 端点
- **会话管理** — 跨进程文件锁定，持久化并恢复对话
- **并行执行** — 跨多个后端并发运行任务
- **安全性** — 速率限制、请求大小限制、可信代理支持
- **可观测性** — 分布式追踪、Prometheus 指标、结构化日志
- **跨平台** — 支持 Linux、macOS 和 Windows 原生二进制

---

## 📑 目录

- [✨ 亮点](#-亮点)
- [📑 目录](#-目录)
- [🚀 快速开始](#-快速开始)
- [📦 安装](#-安装)
- [💡 使用](#-使用)
  - [基本命令](#基本命令)
  - [会话管理](#会话管理)
- [🌐 HTTP API 服务器](#-http-api-服务器)
  - [API 端点](#api-端点)
- [⚙️ 配置](#️-配置)
- [📖 文档](#-文档)
- [🤝 贡献](#-贡献)
- [📊 统计](#-统计)
- [🙏 致谢](#-致谢)
- [📝 许可证](#-许可证)

---

## 🚀 快速开始

```bash
# 通过 Homebrew 安装
brew install signalridge/tap/clinvk

# 使用默认后端运行
clinvk "修复 auth.go 中的 bug"

# 启动 HTTP API 服务器
clinvk serve --port 8080
```

---

## 📦 安装

| 平台 | 方式 | 命令 |
|------|------|------|
| macOS/Linux | Homebrew | `brew install signalridge/tap/clinvk` |
| Windows | Scoop | `scoop bucket add signalridge https://github.com/signalridge/scoop-bucket && scoop install clinvk` |
| Arch Linux | AUR | `yay -S clinvk-bin` |
| NixOS | Flake | `nix run github:signalridge/clinvoker` |
| Docker | GHCR | `docker run ghcr.io/signalridge/clinvk:latest` |
| Debian/Ubuntu | deb | 从 [Releases](https://github.com/signalridge/clinvoker/releases) 下载 |
| Fedora/RHEL | rpm | 从 [Releases](https://github.com/signalridge/clinvoker/releases) 下载 |
| Alpine | apk | 从 [Releases](https://github.com/signalridge/clinvoker/releases) 下载 |
| Go | go install | `go install github.com/signalridge/clinvoker/cmd/clinvk@latest` |

详细说明请参阅 [安装指南](https://signalridge.github.io/clinvoker/tutorials/getting-started/)。

---

## 💡 使用

### 基本命令

```bash
# 使用默认后端运行
clinvk "解释这段代码"

# 使用指定后端
clinvk -b codex "实现用户注册"
clinvk -b gemini "审查这个 PR"

# 恢复最近会话
clinvk resume --last "从上次继续"

# 比较多个后端的响应
clinvk compare --all-backends "解释这个算法"
```

### 会话管理

```bash
# 列出所有会话
clinvk sessions list

# 查看会话详情
clinvk sessions show <session-id>

# 恢复指定会话
clinvk resume <session-id>

# 清理旧会话
clinvk sessions clean --older-than 30d
```

---

## 🌐 HTTP API 服务器

启动 OpenAI/Anthropic 兼容的 API 服务器：

```bash
# 在 8080 端口启动服务器
clinvk serve --port 8080

# 绑定到所有网络接口
clinvk serve --host 0.0.0.0 --port 8080
```

### API 端点

| 端点 | 描述 |
|------|------|
| `POST /openai/v1/chat/completions` | OpenAI 兼容的聊天补全 |
| `POST /anthropic/v1/messages` | Anthropic 兼容的消息 |
| `GET /openai/v1/models` | 列出可用模型 |
| `POST /api/v1/prompt` | 自定义 REST API |
| `GET /health` | 健康检查 |

---

## ⚙️ 配置

```bash
# 显示当前配置
clinvk config show

# 设置默认后端
clinvk config set default_backend claude

# 配置 API 密钥
export ANTHROPIC_API_KEY="sk-..."
export OPENAI_API_KEY="sk-..."
export GOOGLE_API_KEY="..."
```

---

## 📖 文档

完整文档：**[signalridge.github.io/clinvoker](https://signalridge.github.io/clinvoker/)**

| 章节 | 描述 |
|------|------|
| [快速开始](https://signalridge.github.io/clinvoker/tutorials/getting-started/) | 安装和入门 |
| [使用指南](https://signalridge.github.io/clinvoker/guides/) | 详细使用说明 |
| [HTTP API](https://signalridge.github.io/clinvoker/guides/http-server/) | API 服务器文档 |
| [参考](https://signalridge.github.io/clinvoker/reference/) | CLI 参考和配置 |

---

## 🤝 贡献

欢迎贡献！请参阅 [贡献指南](https://signalridge.github.io/clinvoker/concepts/contributing/)。

```bash
# 克隆仓库
git clone https://github.com/signalridge/clinvoker.git
cd clinvoker

# 运行测试
go test ./...

# 构建
go build ./cmd/clinvk
```

---

## 📊 统计

![Alt](https://repobeats.axiom.co/api/embed/b841d080442a754e7f11d8514e3e82db6ae1b120.svg "Repobeats analytics image")

---

## 🙏 致谢

本项目受到以下优秀项目的启发：

- **[AgentAPI](https://github.com/coder/agentapi)** — 开创了对编程代理的 HTTP API 控制。clinvoker 在此基础上增加了跨后端比较、并行执行和会话持久化。
- **[CCG-Workflow](https://github.com/fengshao1227/ccg-workflow)** — 展示了 Claude + Codex + Gemini 协作及任务路由。clinvoker 实现了独立运行，内置 compare/parallel/chain 命令。
- **[CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)** — 为 CLI 工具建立了 OpenAI/Anthropic 兼容 API。clinvoker 将其与 CLI 包装器、会话管理和多后端编排相结合。
- **[MyClaude](https://github.com/cexll/myclaude)** — 创建了用于多后端执行的 codeagent-wrapper。clinvoker 扩展了响应比较、并行运行和持久会话功能。

---

## 📝 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

<div align="center">

**[文档](https://signalridge.github.io/clinvoker/)** · **[报告 Bug](https://github.com/signalridge/clinvoker/issues)** · **[功能请求](https://github.com/signalridge/clinvoker/issues)**

</div>
