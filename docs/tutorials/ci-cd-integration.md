---
title: CI/CD Integration Tutorial
description: Integrate clinvoker into your CI/CD pipelines for automated workflows.
---

# Tutorial: CI/CD Integration

Integrate clinvoker into your CI/CD pipelines for automated code review, documentation generation, and testing.

## What You'll Build

A GitHub Actions workflow that:

1. Triggers on pull requests
2. Runs clinvoker for automated code review
3. Posts review comments back to the PR
4. Fails the build on critical issues

## Prerequisites

- GitHub repository
- clinvoker server deployed (or use local installation)
- GitHub Actions enabled

## Step 1: Set Up clinvoker Server

For CI/CD, you need a persistent clinvoker server:

```yaml title="docker-compose.ci.yml"
services:
  clinvk:
    image: signalridge/clinvoker:latest
    command: serve --port 8080
    environment:
      - CLINVK_API_KEYS=${CLINVK_API_KEYS}
    ports:
      - "8080:8080"
```

Add server settings via config file (mounted at `~/.clinvk/config.yaml`):

```yaml
server:
  rate_limit_enabled: true
  rate_limit_rps: 10
  api_keys_gopass_path: "ci/api-keys"
```

Deploy to your infrastructure or use a cloud instance.

## Step 2: Create Review Tasks

Create `.github/code-review-tasks.json`:

```json
{
  "tasks": [
    {
      "name": "architecture-review",
      "backend": "claude",
      "prompt": "Review this code change for architecture and design patterns.\n\nDIFF:\n{{DIFF}}",
      "output_format": "json"
    },
    {
      "name": "performance-review",
      "backend": "codex",
      "prompt": "Review this code change for performance implications.\n\nDIFF:\n{{DIFF}}",
      "output_format": "json"
    },
    {
      "name": "security-review",
      "backend": "gemini",
      "prompt": "Review this code change for security issues.\n\nDIFF:\n{{DIFF}}",
      "output_format": "json"
    }
  ]
}
```

## Step 3: Create GitHub Action

Create `.github/workflows/ai-review.yml`:

```yaml
name: AI Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  ai-review:
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
        run: |
          git diff origin/${{ github.base_ref }}...HEAD > pr.diff
          echo "Diff size: $(wc -c < pr.diff) bytes"

      - name: Prepare review tasks
        run: |
          DIFF=$(cat pr.diff | jq -Rs .)
          jq --arg diff "$DIFF" '.tasks[].prompt |= sub("\\{\\{DIFF\\}\\}"; $diff)' \
            .github/code-review-tasks.json > review-request.json

      - name: Run AI review
        id: review
        env:
          CLINVK_SERVER: ${{ secrets.CLINVK_SERVER_URL }}
          CLINVK_API_KEY: ${{ secrets.CLINVK_API_KEY }}
        run: |
          curl -X POST "${CLINVK_SERVER}/api/v1/parallel" \
            -H "Authorization: Bearer ${CLINVK_API_KEY}" \
            -H "Content-Type: application/json" \
            -d @review-request.json > review-results.json

          echo "results=$(cat review-results.json)" >> $GITHUB_OUTPUT

      - name: Post review comment
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const results = JSON.parse(fs.readFileSync('review-results.json', 'utf8'));

            let body = '## AI Code Review Results\n\n';

            for (const result of results.results) {
              const icon = result.exit_code === 0 ? '✅' : '⚠️';
              body += `### ${icon} ${result.name}\n\n`;
              body += `${result.output}\n\n`;
              body += '---\n\n';
            }

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });
```

## Step 4: Add Secrets

In your GitHub repository settings, add:

- `CLINVK_SERVER_URL`: Your clinvoker server URL (e.g., `https://clinvk.company.com`)
- `CLINVK_API_KEY`: API key for authentication

## Step 5: Test the Integration

1. Create a new branch
2. Make a code change
3. Open a pull request
4. The AI review should run automatically
5. Check the PR comments for review results

## Advanced: Conditional Reviews

Only run reviews on specific file types:

```yaml
on:
  pull_request:
    paths:
      - "**.go"
      - "**.py"
      - "**.js"
      - "**.ts"
```

Skip reviews for trivial changes:

```yaml
- name: Check diff size
  run: |
    if [ $(wc -c < pr.diff) -gt 50000 ]; then
      echo "Diff too large, skipping AI review"
      exit 0
    fi
```

## GitLab CI Integration

For GitLab, create `.gitlab-ci.yml`:

```yaml
ai-review:
  stage: test
  image: alpine/curl
  variables:
    CLINVK_SERVER: $CLINVK_SERVER_URL
    CLINVK_API_KEY: $CLINVK_API_KEY
  script:
    - apk add --no-cache jq git
    - git fetch origin $CI_MERGE_REQUEST_TARGET_BRANCH_NAME
    - git diff origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME...HEAD > pr.diff
    - |
      DIFF=$(cat pr.diff | jq -Rs .)
      jq --arg diff "$DIFF" '.tasks[].prompt |= sub("\\{\\{DIFF\\}\\}"; $code)' \
        ci/code-review-tasks.json > review-request.json
    - |
      curl -X POST "${CLINVK_SERVER}/api/v1/parallel" \
        -H "Authorization: Bearer ${CLINVK_API_KEY}" \
        -H "Content-Type: application/json" \
        -d @review-request.json > review-results.json
    - cat review-results.json | jq -r '.results[].output'
  rules:
    - if: $CI_MERGE_REQUEST_IID
```

## Best Practices

### 1. Rate Limiting

Configure appropriate rate limits in your clinvoker server:

```yaml
server:
  rate_limit_enabled: true
  rate_limit_rps: 5
  rate_limit_burst: 10
```

### 2. Diff Size Limits

Prevent oversized diffs from overwhelming the system:

```bash
MAX_SIZE=50000  # 50KB
if [ $(wc -c < pr.diff) -gt $MAX_SIZE ]; then
  echo "Diff too large for AI review"
  exit 0
fi
```

### 3. Caching

Cache review results for unchanged files:

```yaml
- uses: actions/cache@v4
  with:
    path: .ai-reviews
    key: ai-reviews-${{ hashFiles('**/*.go') }}
```

## Verification

Test your CI/CD integration:

1. Push a commit with a deliberate issue
2. Verify the AI review catches it
3. Check that the PR comment is posted
4. Fix the issue and verify the review updates

## Next Steps

- See [Automated Code Review](../use-cases/automated-code-review.md) for production patterns
- Learn about [API Gateway](../use-cases/api-gateway-pattern.md) for centralized deployment
- Explore [Test Generation](../use-cases/test-generation-pipeline.md) for CI/CD testing
