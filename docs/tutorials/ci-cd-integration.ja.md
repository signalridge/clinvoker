---
title: CI/CD 連携
description: 自動コードレビュー、テスト、ドキュメント生成のために clinvoker を CI/CD パイプラインへ組み込みます。
---

# チュートリアル: CI/CD 連携

自動コードレビュー、ドキュメント生成、AI 支援テストのために clinvoker を CI/CD パイプラインへ統合する方法を説明します。このチュートリアルでは GitHub Actions / GitLab CI / Jenkins の設定例を扱います。

## CI/CD に clinvoker を統合する理由

### AI を使った CI/CD のメリット

| メリット | 説明 |
|---------|-------------|
| 自動コードレビュー | 人がレビューする前に問題を検出 |
| 品質の一貫性 | すべての PR に同じ基準を適用 |
| ドキュメント | 変更コードのドキュメントを自動生成 |
| セキュリティスキャン | 脆弱性を早期に発見 |
| コスト削減 | 手動レビュー時間を削減 |

### 連携アーキテクチャ

```mermaid
flowchart LR
    PR["Pull Request"] -->|トリガ| ACTION["CI/CD パイプライン"]
    ACTION -->|API 呼び出し| CLINVK["clinvoker サーバー"]

    CLINVK -->|アーキテクチャ| CLAUDE["Claude バックエンド"]
    CLINVK -->|性能| CODEX["Codex バックエンド"]
    CLINVK -->|セキュリティ| GEMINI["Gemini バックエンド"]

    CLAUDE --> RESULT["レビュー結果"]
    CODEX --> RESULT
    GEMINI --> RESULT
```

---

## 前提条件

CI/CD に統合する前に、次を準備してください。

- ローカル検証のために clinvoker がインストール済み
- CI/CD プラットフォーム（GitHub Actions / GitLab CI / Jenkins）へのアクセス
- AI バックエンド用の API キー設定
- YAML とシェルスクリプトの基本理解

### clinvoker サーバーのセットアップ

CI/CD では永続的な clinvoker サーバーを用意すると運用しやすくなります。

```yaml
# docker-compose.ci.yml
services:
  clinvk:
    image: signalridge/clinvoker:latest
    command: serve --port 8080
    environment:
      - CLINVK_API_KEYS=${CLINVK_API_KEYS}
      - CLINVK_BACKEND=claude
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/root/.clinvk/config.yaml
```

サーバーをデプロイします。

```bash
export CLINVK_API_KEYS="your-api-key-1,your-api-key-2"
docker-compose -f docker-compose.ci.yml up -d
```

---

## GitHub Actions 連携

### ワークフロー例（完全版）

`.github/workflows/ai-code-review.yml` を作成します。

```yaml
name: AI Code Review

on:
  pull_request:
    types: [opened, synchronize]
    paths:
      - "**.go"
      - "**.py"
      - "**.js"
      - "**.ts"
      - "**.rs"

env:
  CLINVK_SERVER: ${{ secrets.CLINVK_SERVER_URL }}
  CLINVK_API_KEY: ${{ secrets.CLINVK_API_KEY }}

jobs:
  # Job 1: Architecture Review with Claude
  architecture-review:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install clinvoker CLI
        run: |
          curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Get PR diff
        id: diff
        run: |
          git diff origin/${{ github.base_ref }}...HEAD > pr.diff
          echo "size=$(wc -c < pr.diff)" >> $GITHUB_OUTPUT

      - name: Skip if diff too large
        if: steps.diff.outputs.size > 50000
        run: |
          echo "Diff too large for AI review (>50KB)"
          exit 0

      - name: Run architecture review
        id: review
        run: |
          DIFF=$(cat pr.diff | jq -Rs '.')

          curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
            -H "Authorization: Bearer ${CLINVK_API_KEY}" \
            -H "Content-Type: application/json" \
            -d "{
              \"backend\": \"claude\",
              \"prompt\": \"Review this PR diff for architecture and design patterns. Focus on: 1) SOLID principles, 2) Design patterns, 3) Code organization. DIFF: ${DIFF}\",
              \"output_format\": \"json\"
            }" > architecture-review.json

          echo "result=$(cat architecture-review.json | jq -c .)" >> $GITHUB_OUTPUT

      - name: Upload review artifact
        uses: actions/upload-artifact@v4
        with:
          name: architecture-review
          path: architecture-review.json

  # Job 2: Performance Review with Codex
  performance-review:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install clinvoker CLI
        run: |
          curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Get PR diff
        run: git diff origin/${{ github.base_ref }}...HEAD > pr.diff

      - name: Run performance review
        run: |
          DIFF=$(cat pr.diff | jq -Rs '.')

          curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
            -H "Authorization: Bearer ${CLINVK_API_KEY}" \
            -H "Content-Type: application/json" \
            -d "{
              \"backend\": \"codex\",
              \"prompt\": \"Review this PR diff for performance implications. Focus on: 1) Algorithmic complexity, 2) Resource usage, 3) Optimization opportunities. DIFF: ${DIFF}\",
              \"output_format\": \"json\"
            }" > performance-review.json

      - name: Upload review artifact
        uses: actions/upload-artifact@v4
        with:
          name: performance-review
          path: performance-review.json

  # Job 3: Security Review with Gemini
  security-review:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install clinvoker CLI
        run: |
          curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Get PR diff
        run: git diff origin/${{ github.base_ref }}...HEAD > pr.diff

      - name: Run security review
        run: |
          DIFF=$(cat pr.diff | jq -Rs '.')

          curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
            -H "Authorization: Bearer ${CLINVK_API_KEY}" \
            -H "Content-Type: application/json" \
            -d "{
              \"backend\": \"gemini\",
              \"prompt\": \"Security audit of this PR diff. Focus on: 1) SQL injection, 2) XSS vulnerabilities, 3) Authentication issues, 4) OWASP risks. DIFF: ${DIFF}\",
              \"output_format\": \"json\"
            }" > security-review.json

      - name: Upload review artifact
        uses: actions/upload-artifact@v4
        with:
          name: security-review
          path: security-review.json

  # Job 4: Aggregate and Post Results
  post-review:
    needs: [architecture-review, performance-review, security-review]
    runs-on: ubuntu-latest
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4

      - name: Aggregate reviews
        id: aggregate
        run: |
          # Create combined report
          cat > review-report.md << 'EOF'
          ## AI Code Review Results

          ### Summary
          This PR has been reviewed by three AI backends for comprehensive feedback.

          EOF

          # Add architecture review
          echo "#### Architecture Review (Claude)" >> review-report.md
          jq -r '.output' architecture-review/architecture-review.json >> review-report.md
          echo "" >> review-report.md

          # Add performance review
          echo "#### Performance Review (Codex)" >> review-report.md
          jq -r '.output' performance-review/performance-review.json >> review-report.md
          echo "" >> review-report.md

          # Add security review
          echo "#### Security Review (Gemini)" >> review-report.md
          jq -r '.output' security-review/security-review.json >> review-report.md

          # Check for critical issues
          CRITICAL=$(cat architecture-review/architecture-review.json performance-review/performance-review.json security-review/security-review.json | grep -i "critical" || true)
          if [ -n "$CRITICAL" ]; then
            echo "has_critical=true" >> $GITHUB_OUTPUT
          else
            echo "has_critical=false" >> $GITHUB_OUTPUT
          fi

      - name: Post review comment
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const body = fs.readFileSync('review-report.md', 'utf8');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body.substring(0, 65536)
            });

      - name: Fail on critical issues
        if: steps.aggregate.outputs.has_critical == 'true'
        run: |
          echo "::error::Critical issues found in AI review"
          exit 1
```

### PR コメントの自動投稿

ワークフローは PR にコメントを自動投稿できます。

![PR Comment Example](https://example.com/pr-comment.png)

### 失敗条件と終了コード

clinvoker は CI/CD で扱いやすい標準的な終了コードを使用します。

| 終了コード | 意味 | CI/CD での扱い |
|-----------|---------|--------------|
| 0 | 成功 | 継続 |
| 1 | 一般エラー | ビルド失敗 |
| 2 | バックエンドエラー | リトライ or 失敗 |
| 3 | タイムアウト | 長めのタイムアウトでリトライ |
| 4 | 入力不正 | 早期に失敗 |

ワークフロー側で失敗条件を組む例:

```yaml
- name: Check for critical issues
  run: |
    # Check if any review found critical issues
    if grep -qi "critical\\|severe\\|blocker" review-*.json; then
      echo "::error::Critical issues found"
      exit 1
    fi
```

---

## GitLab CI 連携

### 設定例（完全版）

`.gitlab-ci.yml` を作成します。

```yaml
stages:
  - review
  - report

variables:
  CLINVK_SERVER: $CLINVK_SERVER_URL
  CLINVK_API_KEY: $CLINVK_API_KEY

# Template for review jobs
.ai_review:
  stage: review
  image: alpine/curl
  before_script:
    - apk add --no-cache jq git bash
    - curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
    - export PATH="$PATH:$HOME/.local/bin"
  rules:
    - if: $CI_MERGE_REQUEST_IID

# Architecture Review
architecture-review:
  extends: .ai_review
  script:
    - git fetch origin $CI_MERGE_REQUEST_TARGET_BRANCH_NAME
    - git diff origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME...HEAD > mr.diff
    - |
      DIFF=$(cat mr.diff | jq -Rs '.')
      curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
        -H "Authorization: Bearer ${CLINVK_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "{
          \"backend\": \"claude\",
          \"prompt\": \"Review architecture: ${DIFF}\"
        }" > architecture-review.json
    - cat architecture-review.json | jq -r '.output'
  artifacts:
    paths:
      - architecture-review.json
    expire_in: 1 week

# Performance Review
performance-review:
  extends: .ai_review
  script:
    - git fetch origin $CI_MERGE_REQUEST_TARGET_BRANCH_NAME
    - git diff origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME...HEAD > mr.diff
    - |
      DIFF=$(cat mr.diff | jq -Rs '.')
      curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
        -H "Authorization: Bearer ${CLINVK_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "{
          \"backend\": \"codex\",
          \"prompt\": \"Review performance: ${DIFF}\"
        }" > performance-review.json
  artifacts:
    paths:
      - performance-review.json
    expire_in: 1 week

# Security Review
security-review:
  extends: .ai_review
  script:
    - git fetch origin $CI_MERGE_REQUEST_TARGET_BRANCH_NAME
    - git diff origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME...HEAD > mr.diff
    - |
      DIFF=$(cat mr.diff | jq -Rs '.')
      curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
        -H "Authorization: Bearer ${CLINVK_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "{
          \"backend\": \"gemini\",
          \"prompt\": \"Security audit: ${DIFF}\"
        }" > security-review.json
  artifacts:
    paths:
      - security-review.json
    expire_in: 1 week

# Aggregate and Post Results
post-review:
  stage: report
  image: alpine/jq
  needs: [architecture-review, performance-review, security-review]
  script:
    - |
      echo "## AI Code Review Results" > review-report.md
      echo "" >> review-report.md
      echo "### Architecture (Claude)" >> review-report.md
      jq -r '.output' architecture-review.json >> review-report.md
      echo "" >> review-report.md
      echo "### Performance (Codex)" >> review-report.md
      jq -r '.output' performance-review.json >> review-report.md
      echo "" >> review-report.md
      echo "### Security (Gemini)" >> review-report.md
      jq -r '.output' security-review.json >> review-report.md
    - cat review-report.md
  artifacts:
    paths:
      - review-report.md
    expire_in: 1 week
  rules:
    - if: $CI_MERGE_REQUEST_IID
```

---

## Jenkins パイプライン

### Pipeline 例（完全版）

`Jenkinsfile` を作成します。

```groovy
pipeline {
    agent any

    environment {
        CLINVK_SERVER = credentials('clinvk-server-url')
        CLINVK_API_KEY = credentials('clinvk-api-key')
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
                sh 'git fetch origin'
            }
        }

        stage('Install clinvoker') {
            steps {
                sh '''
                    curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
                    export PATH="$PATH:$HOME/.local/bin"
                    clinvk version
                '''
            }
        }

        stage('Get Diff') {
            when {
                changeRequest()
            }
            steps {
                sh '''
                    git diff origin/${CHANGE_TARGET}...HEAD > pr.diff || true
                    if [ -s pr.diff ]; then
                        echo "Diff size: $(wc -c < pr.diff) bytes"
                    else
                        echo "No changes to review"
                        touch pr.diff
                    fi
                '''
            }
        }

        stage('Parallel AI Review') {
            when {
                changeRequest()
            }
            parallel {
                stage('Architecture Review') {
                    steps {
                        sh '''
                            export PATH="$PATH:$HOME/.local/bin"
                            DIFF=$(cat pr.diff | jq -Rs '.')
                            curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
                                -H "Authorization: Bearer ${CLINVK_API_KEY}" \
                                -H "Content-Type: application/json" \
                                -d "{
                                    \"backend\": \"claude\",
                                    \"prompt\": \"Architecture review: ${DIFF}\"
                                }" > architecture-review.json
                        '''
                    }
                }
                stage('Performance Review') {
                    steps {
                        sh '''
                            export PATH="$PATH:$HOME/.local/bin"
                            DIFF=$(cat pr.diff | jq -Rs '.')
                            curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
                                -H "Authorization: Bearer ${CLINVK_API_KEY}" \
                                -H "Content-Type: application/json" \
                                -d "{
                                    \"backend\": \"codex\",
                                    \"prompt\": \"Performance review: ${DIFF}\"
                                }" > performance-review.json
                        '''
                    }
                }
                stage('Security Review') {
                    steps {
                        sh '''
                            export PATH="$PATH:$HOME/.local/bin"
                            DIFF=$(cat pr.diff | jq -Rs '.')
                            curl -X POST "${CLINVK_SERVER}/api/v1/prompt" \
                                -H "Authorization: Bearer ${CLINVK_API_KEY}" \
                                -H "Content-Type: application/json" \
                                -d "{
                                    \"backend\": \"gemini\",
                                    \"prompt\": \"Security audit: ${DIFF}\"
                                }" > security-review.json
                        '''
                    }
                }
            }
        }

        stage('Aggregate Results') {
            when {
                changeRequest()
            }
            steps {
                sh '''
                    echo "# AI Code Review Report" > review-report.md
                    echo "" >> review-report.md

                    if [ -f architecture-review.json ]; then
                        echo "## Architecture Review" >> review-report.md
                        jq -r '.output' architecture-review.json >> review-report.md
                        echo "" >> review-report.md
                    fi

                    if [ -f performance-review.json ]; then
                        echo "## Performance Review" >> review-report.md
                        jq -r '.output' performance-review.json >> review-report.md
                        echo "" >> review-report.md
                    fi

                    if [ -f security-review.json ]; then
                        echo "## Security Review" >> review-report.md
                        jq -r '.output' security-review.json >> review-report.md
                    fi

                    cat review-report.md
                '''
            }
        }

        stage('Check Critical Issues') {
            when {
                changeRequest()
            }
            steps {
                script {
                    def hasCritical = sh(
                        script: 'grep -qi \"critical\\|severe\" *-review.json && echo \"true\" || echo \"false\"',
                        returnStdout: true
                    ).trim()

                    if (hasCritical == \"true\") {
                        error(\"Critical issues found in AI review\")
                    }
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: '*-review.json,review-report.md', allowEmptyArchive: true
        }
    }
}
```

---

## レート制限の考慮

### レート制限を理解する

AI バックエンドのレート制限は CI/CD の挙動に影響します。

| バックエンド | レート制限 | バースト |
|---------|-----------|-------|
| Claude | 40 requests/min | 60 |
| Codex | 60 requests/min | 100 |
| Gemini | 60 requests/min | 100 |

### レート制限への対策

#### 1. リクエストのバッチ化

複数ファイルを 1 リクエストにまとめます。

```yaml
- name: Batch review
  run: |
    # Combine multiple small files
    cat file1.go file2.go file3.go > combined.go
    clinvk -b claude "Review these files: $(cat combined.go)"
```

#### 2. 指数バックオフ

バックオフ付きでリトライします。

```yaml
- name: Review with retry
  run: |
    for i in 1 2 3; do
      clinvk -b claude "Review this" && break
      sleep $((2 ** i))
    done
```

#### 3. clinvoker サーバー側のレート制限

clinvoker サーバー側でレート制限を有効化できます。

```yaml
server:
  rate_limit_enabled: true
  rate_limit_rps: 10
  rate_limit_burst: 20
```

---

## Secrets 管理のベストプラクティス

### 1. Secrets をハードコードしない

```yaml
# BAD - Never do this
env:
  CLINVK_API_KEY: "sk-12345..."

# GOOD - Use secrets
env:
  CLINVK_API_KEY: ${{ secrets.CLINVK_API_KEY }}
```

### 2. 環境別の Secrets を使う

```yaml
# Production
- name: Production Review
  if: github.ref == 'refs/heads/main'
  env:
    CLINVK_API_KEY: ${{ secrets.CLINVK_API_KEY_PROD }}

# Development
- name: Development Review
  if: github.ref != 'refs/heads/main'
  env:
    CLINVK_API_KEY: ${{ secrets.CLINVK_API_KEY_DEV }}
```

### 3. 定期的にキーをローテーションする

```bash
# Set up key rotation reminder
# Rotate every 90 days
# Update in GitHub/GitLab/Jenkins secrets
```

### 4. キーのスコープを絞る

CI/CD 専用の API キーを作成します。

| キー | 用途 | 権限 |
|-----|---------|-------------|
| ci-readonly | コードレビュー | 読み取り専用 |
| ci-full | フル自動化 | 読み書き |

---

## コスト最適化の戦略

### 1. diff ベースのレビューのみ

```yaml
- name: Check if review needed
  run: |
    # Only review changed files
    CHANGED=$(git diff --name-only origin/main...HEAD | grep -E '\\.(go|py|js)$')
    if [ -z "$CHANGED" ]; then
      echo "No code changes to review"
      exit 0
    fi
```

### 2. サイズ上限

```yaml
- name: Skip large diffs
  run: |
    SIZE=$(git diff origin/main...HEAD | wc -c)
    if [ $SIZE -gt 50000 ]; then
      echo "Diff too large for AI review"
      exit 0
    fi
```

### 3. バックエンドを選択的に使う

```yaml
# Use cheaper backends for simple tasks
- name: Quick check
  run: |
    clinvk -b codex "Quick review"  # Faster, cheaper

- name: Deep review
  run: |
    clinvk -b claude "Deep analysis"  # More expensive
```

### 4. 結果のキャッシュ

```yaml
- uses: actions/cache@v4
  with:
    path: .ai-reviews
    key: ai-reviews-${{ hashFiles('**/*.go') }}
```

---

## トラブルシューティング

### 問題: バックエンドがタイムアウトする

**解決策**: タイムアウトを増やし、必要ならリトライします。

```yaml
- name: Review with timeout
  run: |
    clinvk -b claude --timeout 300 "Review this"
```

### 問題: レート制限超過

**解決策**: リクエスト間にディレイを入れます。

```yaml
- name: Rate-limited review
  run: |
    for file in *.go; do
      clinvk -b claude "Review $file"
      sleep 2  # Rate limit buffer
    done
```

### 問題: Secrets が使えない

**解決策**: Secrets 設定を確認します。

```bash
# Verify secret is set
echo "Secret length: ${#CLINVK_API_KEY}"
```

---

## 次のステップ

- 高度なパターンは [マルチバックエンドのコードレビュー](multi-backend-code-review.md)
- デプロイは [HTTP サーバー](../guides/http-server.md)
- カスタム自動化は [AI スキルの構築](building-ai-skills.md)
- オプション一覧は [設定リファレンス](../reference/configuration.md)

---

## まとめ

このチュートリアルで学んだこと:

1. CI/CD 連携のための clinvoker サーバーセットアップ
2. GitHub Actions での自動コードレビュー設定
3. GitLab CI でのマルチバックエンドレビュー設定
4. Jenkins パイプラインへの AI 統合
5. レート制限とコスト最適化の導入
6. 安全な secrets 管理
7. よくある CI/CD 統合問題の対処

CI/CD に clinvoker を統合することで、コード品質チェックを自動化し、問題を早期に検出し、手動レビューを減らしつつ高い品質基準を維持できます。
