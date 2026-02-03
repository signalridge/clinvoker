---
title: チュートリアル
description: clinvoker を学び、よくあるワークフローを実装するためのステップバイステップのチュートリアル集。
---

# チュートリアル

このチュートリアル集では、clinvoker を学び、一般的なワークフローを実装するための実践的な手順を段階的に解説します。各チュートリアルは 15〜30 分程度で完了するように設計されています。

## 利用できるチュートリアル

<div class="grid cards" markdown>

-   **マルチバックエンドのコードレビュー**

    ---

    Claude / Codex / Gemini を並列に使って、Pull Request に対して包括的なフィードバックを提供する
    コードレビューシステムを構築します。

    [チュートリアルを開始 &rarr;](multi-backend-code-review.md)

-   **AI スキルの構築**

    ---

    clinvoker を利用して他の AI バックエンドを呼び出す Claude Code Skills を作成し、
    特定タスク向けに Claude の機能を拡張します。

    [チュートリアルを開始 &rarr;](building-ai-skills.md)

-   **CI/CD 統合**

    ---

    clinvoker を CI/CD パイプラインへ組み込み、コードレビュー、ドキュメント生成、テストを自動化します。

    [チュートリアルを開始 &rarr;](ci-cd-integration.md)

-   **LangChain 統合**

    ---

    clinvoker を LangChain / LangGraph と接続し、複数 AI バックエンドを使う複雑なエージェントワークフローを構築します。

    [チュートリアルを開始 &rarr;](langchain-integration.md)

</div>

## チュートリアルの難易度

| チュートリアル | レベル | 所要時間 | 前提 |
|----------|-------|------|---------------|
| マルチバックエンドのコードレビュー | 初級 | 20 分 | 基本的な CLI の知識 |
| AI スキルの構築 | 中級 | 30 分 | Claude Code、JSON |
| CI/CD 統合 | 中級 | 25 分 | GitHub Actions または GitLab CI |
| LangChain 統合 | 上級 | 30 分 | Python、LangChain の基礎 |

## 始める前に

すべてのチュートリアルは、次の条件を満たしていることを前提としています。

1. **clinvk をインストール済み** - [はじめに](../tutorials/getting-started.md) を参照
2. **少なくとも 1 つのバックエンド** - Claude Code / Codex CLI / Gemini CLI のいずれか
3. [基本的な使い方](../guides/basic-usage.md) にある **コア概念の基礎理解**

## チュートリアルの構成

各チュートリアルは次の構成です。

1. **概要** - 何を作り、何を学ぶか
2. **前提条件** - 開始前に必要なもの
3. **手順** - 詳細なガイダンス
4. **検証** - 正常に動作していることの確認方法
5. **次のステップ** - ここから先の進め方

## 困ったときは

詰まったら、次を確認してください。

- [FAQ](../concepts/faq.md) を確認する
- [トラブルシューティング](../concepts/troubleshooting.md) を確認する
- [GitHub](https://github.com/signalridge/clinvoker/issues) で Issue を作成する
