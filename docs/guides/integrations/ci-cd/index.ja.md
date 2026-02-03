# CI/CD 統合

このガイドでは、clinvk を CI/CD パイプラインへ組み込み、コードレビューやドキュメント生成などの AI タスクを自動化する方法を説明します。

## 概要

clinvk は CI/CD ワークフローに統合することで、次のような処理を自動化できます。

- Pull Request を自動レビューする
- 変更されたコードのドキュメントを生成する
- セキュリティ分析を実行する
- 複数観点のコード監査を実行する

```mermaid
flowchart LR
    Trigger["SCM event<br/>(push / pull_request)"] --> CI["CI job"]
    CI --> Server["clinvk server<br/>(HTTP API)"]
    Server --> Claude["Claude CLI"]
    Server --> Codex["Codex CLI"]
    Server --> Gemini["Gemini CLI"]
    Server --> CI
    CI --> Output["PR comment / artifacts"]

    style Trigger fill:#e3f2fd,stroke:#1976d2
    style CI fill:#fff3e0,stroke:#f57c00
    style Server fill:#ffecb3,stroke:#ffa000
    style Claude fill:#f3e5f5,stroke:#7b1fa2
    style Codex fill:#e8f5e9,stroke:#388e3c
    style Gemini fill:#ffebee,stroke:#c62828
    style Output fill:#e3f2fd,stroke:#1976d2
```

## 前提条件 - 認証

!!! warning "重要: バックエンドの認証が必要"
    clinvk は認証を代行しません。CI 環境で、各バックエンド CLI ツール側の認証を済ませておく必要があります。

### 必要なシークレット

| バックエンド | 環境変数 | 入手方法 |
|---------|---------------------|------------|
| Claude | `ANTHROPIC_API_KEY` | [Anthropic Console](https://console.anthropic.com/) |
| Codex | `OPENAI_API_KEY` | [OpenAI Platform](https://platform.openai.com/) |
| Gemini | `GOOGLE_API_KEY` | [Google AI Studio](https://makersuite.google.com/) |

### GitHub Actions での設定

```yaml
env:
  ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
  GOOGLE_API_KEY: ${{ secrets.GOOGLE_API_KEY }}
```

### GitLab CI での設定

```yaml
variables:
  ANTHROPIC_API_KEY: $ANTHROPIC_API_KEY  # Set in CI/CD Settings
  OPENAI_API_KEY: $OPENAI_API_KEY
```

## GitHub Actions

### 基本のコードレビュー

```yaml
name: AI Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest

    services:
      clinvk:
        image: ghcr.io/signalridge/clinvk:latest
        ports:
          - 8080:8080
        env:
          CLAUDE_CLI_PATH: /usr/local/bin/claude

    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Get changed files
        id: changed
        run: |
          FILES=$(git diff --name-only ${{ github.event.pull_request.base.sha }} ${{ github.sha }} | tr '\n' ' ')
          echo "files=$FILES" >> $GITHUB_OUTPUT

      - name: Wait for clinvk
        run: |
          for i in {1..30}; do
            curl -s http://localhost:8080/health && break
            sleep 1
          done

      - name: AI Review
        id: review
        run: |
          DIFF=$(git diff ${{ github.event.pull_request.base.sha }} ${{ github.sha }})

          RESULT=$(curl -s -X POST http://localhost:8080/api/v1/prompt \
            -H "Content-Type: application/json" \
            -d "{
              \"backend\": \"claude\",
              \"prompt\": \"Review these code changes for bugs, security issues, and improvements:\\n\\n$DIFF\",
              \"ephemeral\": true
            }")

          echo "review<<EOF" >> $GITHUB_OUTPUT
          echo "$RESULT" | jq -r '.output' >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Post Review Comment
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `## AI Code Review\n\n${{ steps.review.outputs.review }}`
            })
```

### マルチモデルレビュー

```yaml
name: Multi-Model Code Review

on:
  pull_request:

jobs:
  multi-review:
    runs-on: ubuntu-latest

    services:
      clinvk:
        image: ghcr.io/signalridge/clinvk:latest
        ports:
          - 8080:8080

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Get Diff
        id: diff
        run: |
          DIFF=$(git diff ${{ github.event.pull_request.base.sha }} ${{ github.sha }} | head -c 10000)
          echo "diff<<EOF" >> $GITHUB_OUTPUT
          echo "$DIFF" >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Parallel Review
        id: parallel
        run: |
          RESULT=$(curl -s -X POST http://localhost:8080/api/v1/parallel \
            -H "Content-Type: application/json" \
            -d '{
              "tasks": [
                {
                  "backend": "claude",
                  "prompt": "Review architecture and design patterns:\n${{ steps.diff.outputs.diff }}"
                },
                {
                  "backend": "codex",
                  "prompt": "Review for performance issues:\n${{ steps.diff.outputs.diff }}"
                },
                {
                  "backend": "gemini",
                  "prompt": "Review for security vulnerabilities:\n${{ steps.diff.outputs.diff }}"
                }
              ]
            }')

          echo "$RESULT" > review-results.json

      - name: Format and Post Results
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const results = JSON.parse(fs.readFileSync('review-results.json', 'utf8'));

            const body = `## Multi-Model Code Review

            ### Architecture Review (Claude)
            ${results.results[0].output || results.results[0].error}

            ### Performance Review (Codex)
            ${results.results[1].output || results.results[1].error}

            ### Security Review (Gemini)
            ${results.results[2].output || results.results[2].error}
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });
```

### ドキュメント生成

```yaml
name: Generate Documentation

on:
  push:
    branches: [main]
    paths:
      - 'src/**/*.ts'
      - 'src/**/*.py'

jobs:
  docs:
    runs-on: ubuntu-latest

    services:
      clinvk:
        image: ghcr.io/signalridge/clinvk:latest
        ports:
          - 8080:8080

    steps:
      - uses: actions/checkout@v4

      - name: Generate API Docs
        run: |
          for file in src/**/*.ts; do
            CODE=$(cat "$file")

            DOC=$(curl -s -X POST http://localhost:8080/api/v1/chain \
              -H "Content-Type: application/json" \
              -d "{
                \"steps\": [
                  {
                    \"name\": \"analyze\",
                    \"backend\": \"claude\",
                    \"prompt\": \"Analyze this TypeScript file structure:\\n$CODE\"
                  },
                  {
                    \"name\": \"document\",
                    \"backend\": \"codex\",
                    \"prompt\": \"Generate API documentation in Markdown:\\n{{previous}}\"
                  }
                ]
              }" | jq -r '.results[-1].output')

            echo "$DOC" > "docs/api/$(basename $file .ts).md"
          done

      - name: Commit Documentation
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add docs/
          git diff --staged --quiet || git commit -m "docs: auto-generate API documentation"
          git push
```

## GitLab CI

### 基本連携

```yaml
# .gitlab-ci.yml
stages:
  - review

variables:
  CLINVK_URL: http://clinvk:8080

services:
  - name: ghcr.io/signalridge/clinvk:latest
    alias: clinvk

ai-review:
  stage: review
  script:
    - |
      DIFF=$(git diff $CI_MERGE_REQUEST_DIFF_BASE_SHA HEAD)

      RESULT=$(curl -s -X POST $CLINVK_URL/api/v1/prompt \
        -H "Content-Type: application/json" \
        -d "{
          \"backend\": \"claude\",
          \"prompt\": \"Review this code:\\n$DIFF\"
        }")

      echo "$RESULT" | jq -r '.output' > review.txt

    - cat review.txt
  artifacts:
    paths:
      - review.txt
  only:
    - merge_requests
```

## Jenkins

### Jenkinsfile

```groovy
pipeline {
    agent any

    environment {
        CLINVK_URL = 'http://localhost:8080'
    }

    stages {
        stage('Start clinvk') {
            steps {
                sh 'docker run -d --name clinvk -p 8080:8080 ghcr.io/signalridge/clinvk:latest'
                sh 'sleep 5'  // Wait for startup
            }
        }

        stage('AI Review') {
            steps {
                script {
                    def diff = sh(script: 'git diff HEAD~1', returnStdout: true).trim()

                    def response = sh(script: """
                        curl -s -X POST ${CLINVK_URL}/api/v1/prompt \\
                          -H "Content-Type: application/json" \\
                          -d '{"backend": "claude", "prompt": "Review:\\n${diff}"}'
                    """, returnStdout: true)

                    echo "Review Result: ${response}"
                }
            }
        }
    }

    post {
        always {
            sh 'docker stop clinvk || true'
            sh 'docker rm clinvk || true'
        }
    }
}
```

## セルフホストランナーのセットアップ

本番用途では、clinvk を永続サービスとして稼働させることを推奨します。

```bash
# /etc/systemd/system/clinvk.service
[Unit]
Description=clinvk AI CLI Server
After=network.target

[Service]
Type=simple
User=ci
ExecStart=/usr/local/bin/clinvk serve --port 8080 --host 0.0.0.0
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl enable clinvk
sudo systemctl start clinvk
```

## ベストプラクティス

### 1. エフェメラル（ステートレス）セッションを使う

```bash
curl -X POST http://localhost:8080/api/v1/prompt \
  -d '{"backend": "claude", "prompt": "...", "ephemeral": true}'
```

### 2. タイムアウトを設定する

```yaml
- name: AI Review
  timeout-minutes: 10
  run: |
    curl --max-time 300 -X POST http://localhost:8080/api/v1/prompt ...
```

### 3. 大きな diff を扱う

```bash
# Truncate large diffs
DIFF=$(git diff HEAD~1 | head -c 50000)
```

### 4. clinvk イメージをキャッシュする

```yaml
services:
  clinvk:
    image: ghcr.io/signalridge/clinvk:latest
    # Add image pull policy for caching
```

### 5. エラーハンドリング

```bash
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/prompt ...)
HTTP_CODE=$(echo "$RESPONSE" | tail -1)

if [ "$HTTP_CODE" != "200" ]; then
  echo "AI review failed with code $HTTP_CODE"
  exit 1
fi
```

## セキュリティ上の注意

1. **ネットワーク分離**: clinvk をプライベートネットワークで実行する
2. **プロンプトに機密情報を含めない**: API キーやパスワードをプロンプトに入れない
3. **レート制限**: API 呼び出しにレート制限を導入する
4. **監査ログ**: コンプライアンスのために AI のやり取りを記録する

## 次のステップ

- [クライアントライブラリ](../openai-sdk.md) - 言語別 SDK
- [REST API リファレンス](../../../reference/api/rest.md) - API ドキュメント一式
- [トラブルシューティング](../../../concepts/troubleshooting.md) - よくある問題
