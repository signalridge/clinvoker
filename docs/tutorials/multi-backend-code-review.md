---
title: Multi-Backend Code Review
description: Build a comprehensive code review system using multiple AI backends in parallel for architecture, performance, and security analysis.
---

# Tutorial: Multi-Backend Code Review

Learn how to build a production-ready code review system that leverages Claude, Codex, and Gemini simultaneously to provide comprehensive feedback on your code. This approach combines the unique strengths of each AI assistant for more thorough reviews than any single backend could provide.

## Why Multi-Backend Reviews?

### The Problem with Single-Backend Reviews

Individual AI assistants have different strengths and weaknesses:

| Backend | Strengths | Weaknesses |
|---------|-----------|------------|
| Claude | Architecture, reasoning, safety | May be overly cautious |
| Codex | Code generation, performance | Less focus on security |
| Gemini | Security, broad knowledge | May miss implementation details |

### The Multi-Backend Solution

By running reviews in parallel across multiple backends, you get:

1. **Comprehensive Coverage**: Each backend focuses on what it does best
2. **Cross-Validation**: Issues found by multiple backends are higher priority
3. **Diverse Perspectives**: Different AI approaches catch different problems
4. **Faster Feedback**: Parallel execution means no waiting

### Real-World Scenario

Imagine you're reviewing a pull request that adds authentication to your API:

- **Claude** analyzes the overall architecture and design patterns
- **Codex** checks for performance bottlenecks and implementation efficiency
- **Gemini** scans for security vulnerabilities and OWASP risks

The combined feedback gives you confidence that nothing was missed.

---

## Architecture Overview

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
```text

### How It Works

1. **Input**: Code or diff is prepared and templated into review prompts
2. **Distribution**: clinvoker sends the code to all three backends simultaneously
3. **Processing**: Each backend analyzes from its specialized perspective
4. **Aggregation**: Results are collected and formatted into a unified report
5. **Output**: Developers receive comprehensive, categorized feedback

---

## Prerequisites

Before starting, ensure you have:

- clinvoker installed and configured (see [Getting Started](getting-started.md))
- At least two backends available (Claude, Codex, and/or Gemini)
- `jq` installed for JSON processing: `brew install jq` or `apt-get install jq`

Verify your setup:

```bash
clinvk config show
# Should show available backends

jq --version
# Should show version 1.6 or later
```yaml

---

## Step 1: Create the Configuration File

Create `review-config.yaml` to define how each backend contributes to the review:

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
```text

### Configuration Structure Explained

| Section | Purpose |
|---------|---------|
| `review` | Metadata about the review process |
| `templates.architecture` | Claude's role - design and patterns |
| `templates.performance` | Codex's role - efficiency and optimization |
| `templates.security` | Gemini's role - vulnerabilities and risks |

The `{{code}}` and `{{language}}` placeholders will be replaced with actual code at runtime.

---

## Step 2: Create the Review Script

Create `run-review.sh` to orchestrate the multi-backend review:

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
    CODE=$(find "$TARGET" -type f \( -name "*.go" -o -name "*.py" -o -name "*.js" -o -name "*.ts" \) -exec echo "=== {} ===" \; -exec head -50 {} \;)
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
        jq -r '.results[] | "\n=== \(.name | ascii_upcase) ===\nBackend: \(.backend)\nDuration: \(.duration_ms // "N/A")ms\n\n\(.output)\n"' /tmp/review-results.json
        ;;
    markdown)
        echo "# Code Review Report"
        echo ""
        echo "**File:** \`$FILENAME\`"
        echo "**Language:** $LANGUAGE"
        echo "**Generated:** $(date -u +"%Y-%m-%d %H:%M:%S UTC")"
        echo ""
        jq -r '.results[] | "## \(.name | split("-") | map(ascii_upcase) | join(" "))\n\n**Backend:** \(.backend)\n**Duration:** \(.duration_ms // "N/A")ms\n\n\(.output)\n\n---\n"' /tmp/review-results.json
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
```text

Make the script executable:

```bash
chmod +x run-review.sh
```yaml

---

## Step 3: Create Sample Code for Testing

Create `sample-auth.go` with intentional issues to test the review system:

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
```text

This code contains multiple issues across architecture, performance, and security dimensions.

---

## Step 4: Run the Review

Execute the multi-backend review:

```bash
./run-review.sh -o markdown sample-auth.go > review-report.md
cat review-report.md
```text

### Expected Architecture Review (Claude)

Claude should identify:

- **Critical**: Global database connection without proper lifecycle management
- **Warning**: Mixed concerns in AuthHandler (authentication + user retrieval)
- **Suggestion**: Use dependency injection instead of global variables
- **Praise**: Clear struct definition for User

### Expected Performance Review (Codex)

Codex should identify:

- **Critical**: No connection pooling configuration
- **Warning**: Loading all users into memory without pagination
- **Suggestion**: Add query timeouts and context cancellation
- **Optimization**: Prepared statements for repeated queries

### Expected Security Review (Gemini)

Gemini should identify:

- **CRITICAL**: SQL injection vulnerabilities in Login and GetUser
- **CRITICAL**: Plain text password storage
- **HIGH**: Hardcoded database credentials
- **HIGH**: Missing HttpOnly and Secure flags on cookies
- **MEDIUM**: No rate limiting on authentication endpoints
- **MEDIUM**: No authentication check on GetUser endpoint

---

## Step 5: Combine Results into Actionable Feedback

Create `aggregate-review.sh` to intelligently combine results:

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
jq -r '.results[] | select(.name == "architecture-review") | .output' "$RESULTS_FILE" | grep -i -E "(\*\*critical|global variable)" || echo "No critical architecture issues found."
echo ""

echo "### Performance (Codex)"
jq -r '.results[] | select(.name == "performance-review") | .output' "$RESULTS_FILE" | grep -i -E "(\*\*critical|memory leak)" || echo "No critical performance issues found."
echo ""

# Summary statistics
echo "## Summary Statistics"
echo ""
echo "| Backend | Duration | Status |"
echo "|---------|----------|--------|"
jq -r '.results[] | "| \(.backend) | \(.duration_ms // "N/A")ms | \(.exit_code | if . == 0 then "Success" else "Failed" end) |"' "$RESULTS_FILE"
echo ""

# Full reports
echo "## Detailed Reports"
echo ""
jq -r '.results[] | "### \(.name | split("-") | map(ascii_upcase) | join(" "))\n\n**Backend:** \(.backend)\n\n\(.output)\n"' "$RESULTS_FILE"
```yaml

---

## Step 6: Parallel Execution Internals

### How clinvoker Executes in Parallel

When you run `clinvk parallel`, here's what happens internally:

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
```text

### Worker Pool Configuration

Control parallelism in your config:

```yaml
parallel:
  max_workers: 3        # Number of concurrent tasks
  fail_fast: false      # Continue even if one task fails
  aggregate_output: true # Combine all outputs
```text

### Execution Guarantees

- **Isolation**: Each task runs independently
- **Timeout**: Configurable per-task timeout
- **Error Handling**: Failed tasks don't block others
- **Ordering**: Results maintain task order for predictable output

---

## Step 7: CI/CD Integration for Automated Reviews

### GitHub Actions Integration

Create `.github/workflows/multi-backend-review.yml`:

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
          files=$(git diff --name-only origin/${{ github.base_ref }}...HEAD | grep -E '\.(go|py|js|ts)$' || true)
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
```text

### GitLab CI Integration

Create `.gitlab-ci.yml`:

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
        files=$(git diff --name-only origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME...HEAD | grep -E '\.(go|py|js|ts)$' || true)

        for file in $files; do
          ./run-review.sh "$file"
        done
      fi
  rules:
    - if: $CI_MERGE_REQUEST_IID
```text

### Jenkins Pipeline

Create `Jenkinsfile`:

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
                        script: "git diff --name-only origin/${env.CHANGE_TARGET}...HEAD | grep -E '\\.(go|py|js|ts)$' || true",
                        returnStdout: true
                    ).trim()

                    if (changedFiles) {
                        sh 'curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash'

                        changedFiles.split('\n').each { file ->
                            sh "./run-review.sh -o markdown '${file}' >> review.md"
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
```yaml

---

## Best Practices

### 1. Review Scope Management

Limit review scope for better results:

```bash
# Review specific functions
clinvk -b claude "Review the Login function in auth.go for architecture issues"

# Review by diff
clinvk -b codex "Review this diff for performance issues: $(git diff HEAD~1)"
```text

### 2. Result Prioritization

Weight findings by severity and consensus:

| Finding Type | Weight | Action |
|-------------|--------|--------|
| Critical (2+ backends) | Block merge | Must fix |
| Critical (1 backend) | Warning | Should fix |
| Warning (2+ backends) | Warning | Consider fixing |
| Suggestion | Info | Optional |

### 3. Review Templates

Maintain templates in version control:

```bash
git add review-config.yaml run-review.sh
git commit -m "Add multi-backend review configuration"
```text

### 4. Cost Optimization

- Use `--ephemeral` flag for CI/CD to avoid session overhead
- Limit review to changed files only
- Cache review results for unchanged files
- Use smaller models for initial screening

---

## Troubleshooting

### Issue: Reviews Take Too Long

**Solution**: Increase parallelism or use faster models:

```yaml
parallel:
  max_workers: 5

backends:
  claude:
    model: claude-sonnet-4-20250514  # Faster than Opus
```text

### Issue: Inconsistent Results

**Solution**: Add explicit instructions to prompts:

```yaml
prompt: |
  Be specific in your findings. Include:
  - Line numbers where applicable
  - Code snippets demonstrating the issue
  - Specific recommendations for fixes
```text

### Issue: Backend Timeouts

**Solution**: Increase timeout in config:

```yaml
unified_flags:
  command_timeout_secs: 600  # 10 minutes
```text

---

## Next Steps

- Learn about [Chain Execution](../guides/chains.md) for sequential reviews
- Explore [CI/CD Integration](ci-cd-integration.md) for production deployment
- See [Building AI Skills](building-ai-skills.md) for custom review agents
- Review [Architecture Overview](../concepts/architecture.md) for deep internals

---

## Summary

You have built a comprehensive multi-backend code review system that:

1. Leverages Claude for architecture analysis
2. Uses Codex for performance optimization
3. Employs Gemini for security auditing
4. Runs all reviews in parallel for efficiency
5. Integrates with CI/CD for automated feedback

This approach provides more thorough, reliable code reviews than any single AI assistant could deliver alone.
