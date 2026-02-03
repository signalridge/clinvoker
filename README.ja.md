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
  <a href="#-インストール"><img alt="Homebrew" src="https://img.shields.io/badge/Homebrew-FBB040?style=flat-square&logo=homebrew&logoColor=black"></a>
  <a href="#-インストール"><img alt="Scoop" src="https://img.shields.io/badge/Scoop-00BFFF?style=flat-square&logo=windows&logoColor=white"></a>
  <a href="#-インストール"><img alt="AUR" src="https://img.shields.io/badge/AUR-1793D1?style=flat-square&logo=archlinux&logoColor=white"></a>
  <a href="#-インストール"><img alt="Nix" src="https://img.shields.io/badge/Nix-5277C3?style=flat-square&logo=nixos&logoColor=white"></a>
  <a href="#-インストール"><img alt="Docker" src="https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white"></a>
  <a href="#-インストール"><img alt="deb" src="https://img.shields.io/badge/deb-A81D33?style=flat-square&logo=debian&logoColor=white"></a>
  <a href="#-インストール"><img alt="rpm" src="https://img.shields.io/badge/rpm-EE0000?style=flat-square&logo=redhat&logoColor=white"></a>
  <a href="#-インストール"><img alt="apk" src="https://img.shields.io/badge/apk-0D597F?style=flat-square&logo=alpinelinux&logoColor=white"></a>
  <a href="#-インストール"><img alt="Go" src="https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white"></a>
</p>

**[English](README.md)** | [简体中文](README.zh.md) | 日本語 | [Documentation](https://clinvoker.dev)

</div>

---

## ✨ ハイライト

- **マルチバックエンド** — Claude Code / Codex CLI / Gemini CLI をシームレスに切り替え
- **OpenAI 互換 API** — OpenAI/Anthropic API エンドポイントのドロップイン置換
- **セッション管理** — プロセス間ロックにより会話を永続化し、再開可能
- **並列実行** — 複数バックエンドでタスクを同時実行
- **セキュリティ** — レート制限、リクエストサイズ制限、信頼済みプロキシのサポート
- **可観測性** — 分散トレーシング、Prometheus メトリクス、構造化ログ
- **クロスプラットフォーム** — Linux / macOS / Windows のネイティブバイナリ

---

## 📑 目次

- [✨ ハイライト](#-ハイライト)
- [📑 目次](#-目次)
- [🚀 クイックスタート](#-クイックスタート)
- [📦 インストール](#-インストール)
- [💡 使い方](#-使い方)
  - [基本コマンド](#基本コマンド)
  - [セッション管理](#セッション管理)
- [🌐 HTTP API サーバー](#-http-api-サーバー)
  - [API エンドポイント](#api-エンドポイント)
- [⚙️ 設定](#️-設定)
- [📖 ドキュメント](#-ドキュメント)
- [🤝 貢献](#-貢献)
- [📊 統計](#-統計)
- [🙏 謝辞](#-謝辞)
- [📝 ライセンス](#-ライセンス)

---

## 🚀 クイックスタート

```bash
# Homebrew でインストール
brew install signalridge/tap/clinvk

# 既定のバックエンドで実行
clinvk "auth.go のバグを修正して"

# HTTP API サーバーを起動
clinvk serve --port 8080
```

---

## 📦 インストール

| プラットフォーム | 方法 | コマンド |
|----------|--------|---------|
| macOS/Linux | Homebrew | `brew install signalridge/tap/clinvk` |
| Windows | Scoop | `scoop bucket add signalridge https://github.com/signalridge/scoop-bucket && scoop install clinvk` |
| Arch Linux | AUR | `yay -S clinvk-bin` |
| NixOS | Flake | `nix run github:signalridge/clinvoker` |
| Docker | GHCR | `docker run ghcr.io/signalridge/clinvk:latest` |
| Debian/Ubuntu | deb | [Releases](https://github.com/signalridge/clinvoker/releases) からダウンロード |
| Fedora/RHEL | rpm | [Releases](https://github.com/signalridge/clinvoker/releases) からダウンロード |
| Alpine | apk | [Releases](https://github.com/signalridge/clinvoker/releases) からダウンロード |
| Go | go install | `go install github.com/signalridge/clinvoker/cmd/clinvk@latest` |

詳しい手順は [インストールガイド](https://signalridge.github.io/clinvoker/getting-started/installation/) を参照してください。

---

## 💡 使い方

### 基本コマンド

```bash
# 既定のバックエンドで実行
clinvk "このコードを説明して"

# バックエンドを指定する
clinvk -b codex "ユーザー登録を実装して"
clinvk -b gemini "この PR をレビューして"

# 直近のセッションを再開する
clinvk resume --last "前回の続きからお願い"

# バックエンド間で回答を比較する
clinvk compare --all-backends "このアルゴリズムを説明して"
```

### セッション管理

```bash
# セッション一覧
clinvk sessions list

# セッション詳細を表示
clinvk sessions show <session-id>

# 特定のセッションを再開
clinvk resume <session-id>

# 古いセッションをクリーンアップ
clinvk sessions clean --older-than 30d
```

---

## 🌐 HTTP API サーバー

OpenAI/Anthropic 互換の API サーバーを起動します。

```bash
# 8080 で起動
clinvk serve --port 8080

# 全インターフェースへバインド
clinvk serve --host 0.0.0.0 --port 8080
```

### API エンドポイント

| エンドポイント | 説明 |
|----------|-------------|
| `POST /openai/v1/chat/completions` | OpenAI 互換のチャット補完 |
| `POST /anthropic/v1/messages` | Anthropic 互換のメッセージ |
| `GET /openai/v1/models` | 利用可能なモデル一覧 |
| `POST /api/v1/prompt` | プロンプト用のカスタム REST API |
| `GET /health` | ヘルスチェック |

---

## ⚙️ 設定

```bash
# 現在の設定を表示
clinvk config show

# 既定のバックエンドを設定
clinvk config set default_backend claude

# API キーを設定
export ANTHROPIC_API_KEY="sk-..."
export OPENAI_API_KEY="sk-..."
export GOOGLE_API_KEY="..."
```

---

## 📖 ドキュメント

ドキュメント全体: **[signalridge.github.io/clinvoker](https://signalridge.github.io/clinvoker/)**

| セクション | 説明 |
|---------|-------------|
| [はじめに](https://signalridge.github.io/clinvoker/tutorials/getting-started/) | インストールと最初のステップ |
| [ガイド](https://signalridge.github.io/clinvoker/guides/) | 詳細な利用手順 |
| [HTTP API](https://signalridge.github.io/clinvoker/guides/http-server/) | API サーバーのドキュメント |
| [リファレンス](https://signalridge.github.io/clinvoker/reference/) | CLI リファレンスと設定 |

---

## 🤝 貢献

コントリビュート歓迎です。詳しくは [Contributing ガイド](https://signalridge.github.io/clinvoker/concepts/contributing/) を参照してください。

```bash
# リポジトリを clone
git clone https://github.com/signalridge/clinvoker.git
cd clinvoker

# テスト
go test ./...

# ビルド
go build ./cmd/clinvk
```

---

## 📊 統計

![Alt](https://repobeats.axiom.co/api/embed/b841d080442a754e7f11d8514e3e82db6ae1b120.svg "Repobeats analytics image")

---

## 🙏 謝辞

このプロジェクトは、次の素晴らしいプロジェクトに着想を得ています。

- **[AgentAPI](https://github.com/coder/agentapi)** — コーディングエージェントを HTTP API で制御する先駆け。clinvoker はバックエンド横断の比較、並列実行、セッション永続化を追加します。
- **[CCG-Workflow](https://github.com/fengshao1227/ccg-workflow)** — タスクルーティングによる Claude + Codex + Gemini の協調を実証。clinvoker は compare/parallel/chain コマンドを内蔵し、単体での運用を可能にします。
- **[CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)** — CLI ツール向けの OpenAI/Anthropic 互換 API を確立。clinvoker はこれを CLI ラッパー、セッション管理、マルチバックエンドのオーケストレーションと統合します。
- **[MyClaude](https://github.com/cexll/myclaude)** — マルチバックエンド実行のための codeagent-wrapper を提供。clinvoker は回答比較、並列実行、永続セッションを追加で提供します。

---

## 📝 ライセンス

MIT License - [LICENSE](LICENSE) を参照してください。

---

<div align="center">

**[Documentation](https://signalridge.github.io/clinvoker/)** · **[Report Bug](https://github.com/signalridge/clinvoker/issues)** · **[Request Feature](https://github.com/signalridge/clinvoker/issues)**

</div>
