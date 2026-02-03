---
title: マルチバックエンドのコードレビュー
description: 複数 AI バックエンドを並列に使い、アーキテクチャ/性能/セキュリティの観点で包括的なレビューシステムを構築します。
---

# チュートリアル: マルチバックエンドのコードレビュー

Claude / Codex / Gemini を同時に活用し、コードに対して包括的なフィードバックを提供する「本番運用可能な」コードレビューシステムの作り方を説明します。このアプローチは、単一バックエンドでは得られない「観点の網羅」と「相互検証」を提供します。

## なぜマルチバックエンドレビューなのか？

### 単一バックエンドレビューの課題

AI アシスタントにはそれぞれ得意/不得意があります。

| バックエンド | 得意 | 不得意 |
|---------|-----------|------------|
| Claude | アーキテクチャ、推論、安全性 | 慎重になり過ぎることがある |
| Codex | コード生成、性能 | セキュリティの比重が低いことがある |
| Gemini | セキュリティ、広い知識 | 実装の細部を見落とすことがある |

### マルチバックエンドの解決策

複数バックエンドでレビューを並列に走らせることで、次が得られます。

1. **網羅性**: 各バックエンドが得意領域に集中できる
2. **相互検証**: 複数バックエンドが指摘した問題は優先度が高い
3. **多様な視点**: アプローチの違いにより検出できる問題が増える
4. **速いフィードバック**: 並列実行で待ち時間を短縮

### 実際のシナリオ

例えば、API に認証を追加する PR をレビューする場面を想像してください。

- **Claude** が全体アーキテクチャと設計パターンを評価
- **Codex** が性能ボトルネックと実装効率を評価
- **Gemini** が脆弱性と OWASP リスクを評価

3 つを統合することで、「見落としがない」という確信が得られます。

---

## アーキテクチャ概要

```text
                    Code Input
                         |
         +---------------+---------------+
         |               |               |
    [Claude]        [Codex]        [Gemini]
         |               |               |
   Architecture    Performance     Security
      Review         Review         Review
         |               |               |
         +---------------+---------------+
                         |
                 Aggregate Results
                         |
              Actionable Feedback
```

### 仕組み

1. **入力準備**: コードまたは diff をレビュー用プロンプトへ整形
2. **配布**: clinvoker が 3 バックエンドへ同時に送信
3. **処理**: 各バックエンドが専門観点で分析
4. **集約**: 結果を収集し、統一レポートへ整形
5. **出力**: 開発者がカテゴリ別のフィードバックを受け取る

---

## 前提条件

開始前に次を確認してください。

- clinvoker がインストール/設定済み（[Getting Started](getting-started.md)）
- 少なくとも 2 つのバックエンドが利用可能（Claude/Codex/Gemini のいずれか）
- JSON 処理のため `jq` がインストール済み（`brew install jq` または `apt-get install jq`）

セットアップ確認:

```bash
clinvk config show
# Should show available backends

jq --version
# Should show version 1.6 or later
```

---

## ステップ 1: 設定ファイルを作る

各バックエンドがレビューにどう貢献するかを定義するため、`review-config.yaml` を作成します。

```yaml
# Review configuration for multi-backend code review
review:
  name: "Comprehensive Code Review"
  version: "1.0"

  # Each backend has a specialized role
templates:
  architecture:
    backend: claude
    prompt: |
      You are a senior software architect reviewing code for design quality.

      Review the following code for:
      1. SOLID principles adherence
      2. Design patterns usage (appropriate? correctly implemented?)
      3. Code organization and modularity
      4. Maintainability and readability
      5. API design (if applicable)
      6. Error handling strategy

      CODE TO REVIEW:
      ```{{language}}
      {{code}}
      ```

      Provide your findings in this format:
      - **Critical**: Issues that must be fixed
      - **Warning**: Issues that should be addressed
      - **Suggestion**: Improvements to consider
      - **Praise**: What was done well

  performance:
    backend: codex
    prompt: |
      You are a performance engineer reviewing code for efficiency.

      Review the following code for:
      1. Algorithmic complexity (Big O analysis)
      2. Resource usage (memory, CPU, I/O)
      3. Database query efficiency
      4. Caching opportunities
      5. Concurrency issues
      6. Potential bottlenecks

      CODE TO REVIEW:
      ```{{language}}
      {{code}}
      ```

      Provide your findings with specific recommendations for optimization.

  security:
    backend: gemini
    prompt: |
      You are a security engineer reviewing code for vulnerabilities.

      Review the following code for:
      1. Input validation and sanitization
      2. SQL injection risks
      3. XSS and CSRF vulnerabilities
      4. Authentication/authorization flaws
      5. OWASP Top 10 risks
      6. Secrets or credentials in code
      7. Insecure dependencies

      CODE TO REVIEW:
      ```{{language}}
      {{code}}
      ```

      Categorize findings as CRITICAL, HIGH, MEDIUM, or LOW risk.
```

### 設定構造の説明

| セクション | 役割 |
|---------|---------|
| `review` | レビューのメタデータ |
| `templates.architecture` | Claude の役割（設計/パターン） |
| `templates.performance` | Codex の役割（効率/最適化） |
| `templates.security` | Gemini の役割（脆弱性/リスク） |

`{{code}}` と `{{language}}` のプレースホルダは、実行時に実際のコードへ置換されます。

---

## ステップ 2: レビュー用スクリプトを作る

マルチバックエンドレビューをオーケストレーションする `run-review.sh` を作成します。

```bash
#!/bin/bash
# Multi-backend code review script

set -e

# Configuration
CONFIG_FILE="${CONFIG_FILE:-review-config.yaml}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-json}"

# Show usage
usage() {
    echo "Usage: $0 [options] <file-or-directory>"
    echo ""
    echo "Options:"
    echo "  -c, --config    Config file (default: review-config.yaml)"
    echo "  -o, --output    Output format: text, json, markdown (default: json)"
    echo "  -h, --help      Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 src/auth.go"
    echo "  $0 -o markdown src/"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

if [ -z "$TARGET" ]; then
    echo "Error: No target specified"
    usage
fi

# Detect language from file extension
detect_language() {
    local file="$1"
    local ext="${file##*.}"
    case "$ext" in
        go) echo "go" ;;
        py) echo "python" ;;
        js) echo "javascript" ;;
        ts) echo "typescript" ;;
        rs) echo "rust" ;;
        java) echo "java" ;;
        *) echo "text" ;;
    esac
}

# Prepare code input
if [ -f "$TARGET" ]; then
    CODE=$(cat "$TARGET")
    LANGUAGE=$(detect_language "$TARGET")
    FILENAME=$(basename "$TARGET")
elif [ -d "$TARGET" ]; then
    # For directories, create a summary
    CODE=$(find "$TARGET" -type f \\( -name "*.go" -o -name "*.py" -o -name "*.js" -o -name "*.ts" \\) -exec echo "=== {} ===" \\; -exec head -50 {} \\;)
    LANGUAGE="mixed"
    FILENAME=$(basename "$TARGET")
else
    echo "Error: Target not found: $TARGET"
    exit 1
fi

# Escape code for JSON
CODE_JSON=$(echo "$CODE" | jq -Rs '.')

# Create parallel tasks file
echo "Creating review tasks..."
cat > /tmp/review-tasks.json << EOF
{
  "tasks": [
    {
      "name": "architecture-review",
      "backend": "claude",
      "prompt": $(yq e '.templates.architecture.prompt' "$CONFIG_FILE" | sed "s/{{language}}/$LANGUAGE/g" | sed "s/{{code}}/$CODE_JSON/g" | jq -Rs '.'),
      "output_format": "text"
    },
    {
      "name": "performance-review",
      "backend": "codex",
      "prompt": $(yq e '.templates.performance.prompt' "$CONFIG_FILE" | sed "s/{{language}}/$LANGUAGE/g" | sed "s/{{code}}/$CODE_JSON/g" | jq -Rs '.'),
      "output_format": "text"
    },
    {
      "name": "security-review",
      "backend": "gemini",
      "prompt": $(yq e '.templates.security.prompt' "$CONFIG_FILE" | sed "s/{{language}}/$LANGUAGE/g" | sed "s/{{code}}/$CODE_JSON/g" | jq -Rs '.'),
      "output_format": "text"
    }
  ]
}
EOF

# Run parallel review
echo "Running multi-backend review..."
clinvk parallel -f /tmp/review-tasks.json -o json > /tmp/review-results.json

# Format output based on requested format
case "$OUTPUT_FORMAT" in
    text)
        echo ""
        echo "========================================"
        echo "Code Review Report: $FILENAME"
        echo "========================================"
        echo ""
        jq -r '.results[] | \"\\n=== \\(.name | ascii_upcase) ===\\nBackend: \\(.backend)\\nDuration: \\(.duration_ms // \"N/A\")ms\\n\\n\\(.output)\\n\"' /tmp/review-results.json
        ;;
    markdown)
        echo "# Code Review Report"
        echo ""
        echo "**File:** \\`$FILENAME\\`"
        echo "**Language:** $LANGUAGE"
        echo "**Generated:** $(date -u +\"%Y-%m-%d %H:%M:%S UTC\")"
        echo ""
        jq -r '.results[] | \"## \\(.name | split(\"-\") | map(ascii_upcase) | join(\" \"))\\n\\n**Backend:** \\(.backend)\\n**Duration:** \\(.duration_ms // \"N/A\")ms\\n\\n\\(.output)\\n\\n---\\n\"' /tmp/review-results.json
        ;;
    json)
        cat /tmp/review-results.json
        ;;
    *)
        echo "Unknown output format: $OUTPUT_FORMAT"
        exit 1
        ;;
esac

echo ""
echo "Review complete!"
```

スクリプトに実行権限を付けます。

```bash
chmod +x run-review.sh
```

---

## ステップ 3: テスト用のサンプルコードを作る

レビューシステムを試すため、意図的に問題を含んだ `sample-auth.go` を作成します。

```go
package main

import (
    "database/sql"
    "fmt"
    "net/http"
    "time"
)

// User represents a user in the system
type User struct {
    ID       int
    Username string
    Password string // Stored in plain text - security issue
    Email    string
}

// AuthHandler handles authentication requests
type AuthHandler struct {
    db *sql.DB
}

// Login authenticates a user
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    password := r.URL.Query().Get("password")

    // SQL injection vulnerability
    query := fmt.Sprintf("SELECT id, username, password FROM users WHERE username='%s' AND password='%s'", username, password)

    row := h.db.QueryRow(query)

    var user User
    err := row.Scan(&user.ID, &user.Username, &user.Password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Set session cookie without HttpOnly or Secure flags
    http.SetCookie(w, &http.Cookie{
        Name:  "session",
        Value: fmt.Sprintf("user_%d", user.ID),
    })

    fmt.Fprintf(w, "Welcome, %s!", user.Username)
}

// GetUser retrieves a user by ID
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // No authentication check - anyone can access any user
    id := r.URL.Query().Get("id")

    // Another SQL injection
    query := "SELECT * FROM users WHERE id = " + id
    rows, _ := h.db.Query(query)
    defer rows.Close()

    // Inefficient: loading all users into memory
    var users []User
    for rows.Next() {
        var u User
        rows.Scan(&u.ID, &u.Username, &u.Password, &u.Email)
        users = append(users, u)
    }

    // No rate limiting - vulnerable to brute force
    for i := 0; i < len(users); i++ {
        fmt.Fprintf(w, "User: %s\n", users[i].Username)
    }
}

// Global database connection without connection pooling
var globalDB *sql.DB

func init() {
    var err error
    // Hardcoded credentials - security issue
    globalDB, err = sql.Open("postgres", "postgres://admin:secret123@localhost/mydb?sslmode=disable")
    if err != nil {
        panic(err)
    }
}
```

このコードには、アーキテクチャ/性能/セキュリティの各観点で複数の問題が含まれています。

---

## ステップ 4: レビューを実行する

マルチバックエンドレビューを実行します。

```bash
./run-review.sh -o markdown sample-auth.go > review-report.md
cat review-report.md
```

### 期待されるアーキテクチャレビュー（Claude）

Claude は例えば次のような指摘をするはずです。

- **Critical**: グローバル DB 接続のライフサイクル管理が不適切
- **Warning**: AuthHandler に責務が混在（認証 + ユーザー取得）
- **Suggestion**: グローバル変数ではなく DI を使う
- **Praise**: User 構造体定義が明確

### 期待される性能レビュー（Codex）

Codex は例えば次のような指摘をするはずです。

- **Critical**: コネクションプーリング設定がない
- **Warning**: ページングなしで全ユーザーをメモリに読み込む
- **Suggestion**: クエリタイムアウトと context キャンセル
- **Optimization**: 繰り返しクエリに prepared statement

### 期待されるセキュリティレビュー（Gemini）

Gemini は例えば次のような指摘をするはずです。

- **CRITICAL**: Login/GetUser の SQL インジェクション
- **CRITICAL**: パスワードを平文保存
- **HIGH**: DB 認証情報のハードコード
- **HIGH**: Cookie に HttpOnly/Secure がない
- **MEDIUM**: 認証エンドポイントにレート制限がない
- **MEDIUM**: GetUser に認可チェックがない

---

## ステップ 5: 結果を「実行可能な」フィードバックへ統合する

結果を優先度付けしてまとめる `aggregate-review.sh` を作成します。

```bash
#!/bin/bash
# Aggregate and prioritize multi-backend review results

RESULTS_FILE="${1:-/tmp/review-results.json}"

if [ ! -f "$RESULTS_FILE" ]; then
    echo "Error: Results file not found: $RESULTS_FILE"
    exit 1
fi

echo "# Prioritized Code Review Report"
echo ""
echo "Generated: $(date)"
echo ""

# Extract critical issues from all backends
echo "## Critical Issues (Immediate Action Required)"
echo ""

echo "### Security (Gemini)"
jq -r '.results[] | select(.name == "security-review") | .output' "$RESULTS_FILE" | grep -i -E "(critical|sql injection|hardcoded)" || echo "No critical security issues found."
echo ""

echo "### Architecture (Claude)"
jq -r '.results[] | select(.name == "architecture-review") | .output' "$RESULTS_FILE" | grep -i -E "(\\*\\*critical|global variable)" || echo "No critical architecture issues found."
echo ""

echo "### Performance (Codex)"
jq -r '.results[] | select(.name == "performance-review") | .output' "$RESULTS_FILE" | grep -i -E "(\\*\\*critical|memory leak)" || echo "No critical performance issues found."
echo ""

# Summary statistics
echo "## Summary Statistics"
echo ""
echo "| Backend | Duration | Status |"
echo "|---------|----------|--------|"
jq -r '.results[] | \"| \\(.backend) | \\(.duration_ms // \"N/A\")ms | \\(.exit_code | if . == 0 then \"Success\" else \"Failed\" end) |\"' "$RESULTS_FILE"
echo ""

# Full reports
echo "## Detailed Reports"
echo ""
jq -r '.results[] | \"### \\(.name | split(\"-\") | map(ascii_upcase) | join(\" \"))\\n\\n**Backend:** \\(.backend)\\n\\n\\(.output)\\n\"' "$RESULTS_FILE"
```

---

## ステップ 6: 並列実行の内部（理解のため）

### clinvoker が並列実行する仕組み

`clinvk parallel` を実行すると、内部的には概ね次の流れになります。

```text
1. Parse task file
   |
2. Validate all backends are available
   |
3. Create worker pool (default: 3 workers)
   |
4. Submit tasks to worker pool
   |-- Task 1: Claude (architecture) --> Worker 1
   |-- Task 2: Codex (performance) --> Worker 2
   |-- Task 3: Gemini (security)  --> Worker 3
   |
5. Wait for all tasks to complete
   |
6. Aggregate results into JSON response
```

### ワーカープール設定

並行度は設定で制御できます。

```yaml
parallel:
  max_workers: 3        # Number of concurrent tasks
  fail_fast: false      # Continue even if one task fails
  aggregate_output: true # Combine all outputs
```

### 実行上の保証

- **分離**: 各タスクは独立して実行される
- **タイムアウト**: タスクごとに設定可能
- **エラー処理**: 失敗タスクが他をブロックしない
- **順序**: 結果はタスク順を維持し、出力が予測可能

---

## ステップ 7: CI/CD に組み込んで自動レビューする

### GitHub Actions 連携

`.github/workflows/multi-backend-review.yml` を作成します。

```yaml
name: Multi-Backend Code Review

on:
  pull_request:
    paths:
      - "**.go"
      - "**.py"
      - "**.js"
      - "**.ts"

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install clinvoker
        run: |
          curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Get changed files
        id: changed
        run: |
          files=$(git diff --name-only origin/${{ github.base_ref }}...HEAD | grep -E '\\.(go|py|js|ts)$' || true)
          echo "files=$files" >> $GITHUB_OUTPUT

      - name: Run multi-backend review
        if: steps.changed.outputs.files != ''
        run: |
          for file in ${{ steps.changed.outputs.files }}; do
            echo "Reviewing $file..."
            ./run-review.sh -o markdown "$file" >> review-output.md
          done

      - name: Post review comment
        if: steps.changed.outputs.files != ''
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const body = fs.readFileSync('review-output.md', 'utf8');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body.substring(0, 65536) // GitHub comment limit
            });

      - name: Check for critical issues
        if: steps.changed.outputs.files != ''
        run: |
          if grep -i "critical" review-output.md; then
            echo "::error::Critical issues found!"
            exit 1
          fi
```

### GitLab CI 連携

`.gitlab-ci.yml` を作成します。

```yaml
multi-backend-review:
  stage: test
  image: alpine/curl
  variables:
    CLINVK_BACKEND: claude
  before_script:
    - apk add --no-cache jq bash git
    - curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
  script:
    - |
      if [ "$CI_MERGE_REQUEST_IID" ]; then
        git fetch origin $CI_MERGE_REQUEST_TARGET_BRANCH_NAME
        files=$(git diff --name-only origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME...HEAD | grep -E '\\.(go|py|js|ts)$' || true)

        for file in $files; do
          ./run-review.sh "$file"
        done
      fi
  rules:
    - if: $CI_MERGE_REQUEST_IID
```

### Jenkins Pipeline

`Jenkinsfile` を作成します。

```groovy
pipeline {
    agent any

    stages {
        stage('Multi-Backend Review') {
            when {
                changeRequest()
            }
            steps {
                script {
                    def changedFiles = sh(
                        script: \"git diff --name-only origin/${env.CHANGE_TARGET}...HEAD | grep -E '\\\\.(go|py|js|ts)$' || true\",
                        returnStdout: true
                    ).trim()

                    if (changedFiles) {
                        sh 'curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash'

                        changedFiles.split('\\n').each { file ->
                            sh \"./run-review.sh -o markdown '${file}' >> review.md\"
                        }

                        publishHTML([
                            reportDir: '.',
                            reportFiles: 'review.md',
                            reportName: 'AI Code Review'
                        ])
                    }
                }
            }
        }
    }
}
```

---

## ベストプラクティス

### 1. レビュー範囲を適切に管理する

範囲を絞ると結果が良くなることがあります。

```bash
# Review specific functions
clinvk -b claude "Review the Login function in auth.go for architecture issues"

# Review by diff
clinvk -b codex "Review this diff for performance issues: $(git diff HEAD~1)"
```

### 2. 結果の優先順位付け

深刻度と合意（複数バックエンドで同じ指摘）を重み付けします。

| 指摘タイプ | 重み | アクション |
|-------------|--------|--------|
| Critical（2+ バックエンド） | マージブロック | 必ず修正 |
| Critical（1 バックエンド） | 警告 | 修正推奨 |
| Warning（2+ バックエンド） | 警告 | 修正検討 |
| Suggestion | 参考 | 任意 |

### 3. レビューテンプレートはバージョン管理する

```bash
git add review-config.yaml run-review.sh
git commit -m "Add multi-backend review configuration"
```

### 4. コスト最適化

- CI/CD ではセッションオーバーヘッド回避のため `--ephemeral` を検討
- 変更ファイルのみにレビューを限定
- 未変更ファイルの結果をキャッシュ
- まず小さいモデルでスクリーニングし、必要に応じて深掘り

---

## トラブルシューティング

### 問題: レビューに時間がかかる

**解決策**: 並行度を上げる、または高速モデルを使います。

```yaml
parallel:
  max_workers: 5

backends:
  claude:
    model: claude-sonnet-4-20250514  # Faster than Opus
```

### 問題: 結果が不安定（ばらつく）

**解決策**: プロンプトに具体的な出力要求を追加します。

```yaml
prompt: |
  Be specific in your findings. Include:
  - Line numbers where applicable
  - Code snippets demonstrating the issue
  - Specific recommendations for fixes
```

### 問題: バックエンドがタイムアウトする

**解決策**: 設定でタイムアウトを増やします。

```yaml
unified_flags:
  command_timeout_secs: 600  # 10 minutes
```

---

## 次のステップ

- 順次レビューは [チェーン実行](../guides/chains.md)
- 本番デプロイは [CI/CD 連携](ci-cd-integration.md)
- カスタムレビューエージェントは [AI スキルの構築](building-ai-skills.md)
- 内部構造は [アーキテクチャ概要](../concepts/architecture.md)

---

## まとめ

このチュートリアルで構築したマルチバックエンドレビューシステムは次を実現します。

1. Claude によるアーキテクチャ分析
2. Codex による性能最適化の観点
3. Gemini によるセキュリティ監査
4. 並列実行による高効率なレビュー
5. CI/CD 統合による自動フィードバック

このアプローチは、単一の AI アシスタントだけでは得にくい「より丁寧で信頼性の高い」コードレビューを提供します。
