---
title: ガイド
description: clinvoker を効果的に使うためのステップバイステップ・ガイド。
---

# ガイド

clinvoker のガイドへようこそ。ここでは、よくあるタスクやワークフローを進めるための手順を段階的に説明します。

## はじめに

clinvoker が初めての方は、まずここから始めてください。

- [基本的な使い方](basic-usage.md) - clinvoker 利用の基礎を学ぶ
- [設定](configuration.md) - 用途に合わせて clinvoker をカスタマイズする
- [セッション管理](sessions.md) - 永続セッションを扱う

## 実行パターン

AI タスクを実行するためのさまざまな方法を学びます。

- [並列実行](parallel.md) - 複数のプロンプトを同時に実行する
- [チェーン実行](chains.md) - 逐次ワークフローを作成する
- [HTTP サーバー](http-server.md) - clinvoker を API サーバーとして動かす

## バックエンド

対応している AI バックエンドを確認します。

- [バックエンド概要](backends/index.md) - 利用可能なバックエンドを比較する
- [Claude Code](backends/claude.md) - Anthropic の Claude Code 連携
- [Codex CLI](backends/codex.md) - OpenAI の Codex CLI 連携
- [Gemini CLI](backends/gemini.md) - Google の Gemini CLI 連携
- [バックエンド比較](compare.md) - バックエンドの詳細比較

## 連携

既存ツールと clinvoker をつなぎます。

- [連携の概要](integrations/index.md) - 連携オプション一覧
- [Claude Code Skills](integrations/claude-code-skills.md) - 再利用可能なスキルを作成する
- [LangChain/LangGraph](integrations/langchain-langgraph.md) - LLM アプリケーションを構築する
- [OpenAI SDK](integrations/openai-sdk.md) - OpenAI 互換クライアントを利用する
- [Anthropic SDK](integrations/anthropic-sdk.md) - Anthropic 互換クライアントを利用する
- [MCP Server](integrations/mcp-server.md) - Model Context Protocol 連携
- [CI/CD プラットフォーム](integrations/ci-cd/index.md) - CI/CD で自動化する

## ガイドの選び方

どこから始めるか迷う場合は、目的から選んでください。

| 目的 | おすすめ |
|------|----------|
| 最初のプロンプトを実行したい | [基本的な使い方](basic-usage.md) |
| AI の回答を比較したい | [バックエンド比較](compare.md) |
| コードレビューを自動化したい | [CI/CD 統合](integrations/ci-cd/index.md) |
| AI ワークフローを組みたい | [チェーン実行](chains.md) |
| サービスとしてデプロイしたい | [HTTP サーバー](http-server.md) |
| アプリに組み込みたい | [連携の概要](integrations/index.md) |
