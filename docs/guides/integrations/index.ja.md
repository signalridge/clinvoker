# 連携ガイド

clinvk は、Claude Code Skills、LangChain エージェント、CI/CD パイプラインなど、AI 開発ワークフローへスムーズに統合できるように設計されています。

## 連携の概要

clinvk は HTTP API を通じて、さまざまなツールやフレームワークと連携できます。

**連携方法:**

| カテゴリ | ツール | API エンドポイント |
|----------|-------|--------------|
| AI 開発 | Claude Code Skills, LangChain/LangGraph, Custom Agents | `/api/v1/*` または `/openai/v1/*` |
| SDK | OpenAI SDK, Anthropic SDK | `/openai/v1/*` または `/anthropic/v1/*` |
| 自動化 | CI/CD パイプライン、シェルスクリプト | `/api/v1/*` |

**データフロー:**

```bash
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  AI Development │     │      SDKs       │     │   Automation    │
│  ─────────────  │     │  ────────────   │     │  ────────────   │
│  Claude Skills  │     │   OpenAI SDK    │     │  CI/CD Pipelines│
│  LangChain      │────▶│   Anthropic SDK │────▶│  Shell Scripts  │
│  Custom Agents  │     └─────────────────┘     └─────────────────┘
└─────────────────┘              │                       │
                                 └───────────┬───────────┘
                                             ▼
                                   ┌─────────────────┐
                                   │  clinvk server  │
                                   └────────┬────────┘
                                            │
                       ┌────────────────────┼────────────────────┐
                       ▼                    ▼                    ▼
               ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
               │  Claude CLI  │   │  Codex CLI   │   │  Gemini CLI  │
               └──────────────┘   └──────────────┘   └──────────────┘
```

## 連携方法

| 方法 | ユースケース | API エンドポイント |
|--------|----------|--------------|
| [Claude Code Skills](claude-code-skills.md) | Claude をマルチバックエンド対応で拡張 | `/api/v1/*` |
| [LangChain/LangGraph](langchain-langgraph.md) | AI フレームワーク連携 | `/openai/v1/*` |
| [CI/CD](ci-cd/index.md) | 自動コードレビュー、ドキュメント生成 | `/api/v1/*` |
| [クライアントライブラリ](../../reference/api/index.md) | Python / TypeScript / Go クライアント | すべてのエンドポイント |
| [MCP Server](mcp-server.md) | Model Context Protocol 連携 | 将来対応 |

## すぐに試せる連携例

### Claude Code Skill

```bash
# In your skill script
curl -s http://localhost:8080/api/v1/prompt \
  -H "Content-Type: application/json" \
  -d '{"backend": "gemini", "prompt": "Analyze this data"}'
```

### LangChain

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/openai/v1",
    model="claude",
    api_key="not-needed"
)
```

### OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/openai/v1",
    api_key="not-needed"
)
```

### GitHub Actions

```yaml
- name: AI Code Review
  run: |
    payload=$(jq -n --arg prompt "Review:\n${{ steps.diff.outputs.changes }}" '{backend:"claude", prompt:$prompt}')
    curl -sS -X POST http://localhost:8080/api/v1/prompt \
      -H "Content-Type: application/json" \
      -d "$payload"
```

## 前提条件

連携する前に、次を確認してください。

1. **clinvk がインストール済み**: [はじめに](../../tutorials/getting-started.md) を参照
2. **バックエンド CLI が利用可能**: `claude`, `codex`, `gemini` のいずれか 1 つ以上
3. **サーバーが起動している**: `clinvk serve --port 8080`

## 適切なエンドポイントの選び方

| 状況 | 推奨エンドポイント |
|---------------|---------------------|
| OpenAI SDK / LangChain を使う | `/openai/v1/*` |
| Anthropic SDK を使う | `/anthropic/v1/*` |
| Claude Code Skills を作る | `/api/v1/*` |
| parallel/chain が必要 | `/api/v1/*` |
| シンプルな REST 連携 | `/api/v1/*` |

## 次のステップ

目的に合う連携パスを選んでください。

- **AI エージェント開発**: [Claude Code Skills](claude-code-skills.md) から始める
- **フレームワーク連携**: [LangChain/LangGraph](langchain-langgraph.md) を参照
- **自動化**: [CI/CD](ci-cd/index.md) を参照
- **カスタム開発**: [クライアントライブラリ](../../reference/api/index.md) を参照
