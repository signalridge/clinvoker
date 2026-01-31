---
title: Multi-Backend Code Review Tutorial
description: Build a code review system using multiple AI backends in parallel.
---

# Tutorial: Multi-Backend Code Review

Build a complete code review system that uses Claude, Codex, and Gemini in parallel to provide comprehensive feedback on your code.

## What You'll Build

A code review pipeline that:

1. Takes code as input
2. Sends it to three backends simultaneously
3. Collects architecture, performance, and security feedback
4. Aggregates results into a structured report

## Prerequisites

- clinvk installed and configured
- At least two backends available (Claude, Codex, or Gemini)
- A code file to review (we'll use an example)

## Step 1: Set Up the Review Tasks

Create a file named `review-tasks.json`:

```json
{
  "tasks": [
    {
      "name": "architecture-review",
      "backend": "claude",
      "prompt": "Review this code for architecture and design patterns.\n\nFocus on:\n1. SOLID principles\n2. Design patterns\n3. Code organization\n4. Maintainability\n\nCODE:\n{{CODE}}"
    },
    {
      "name": "performance-review",
      "backend": "codex",
      "prompt": "Review this code for performance implications.\n\nFocus on:\n1. Algorithmic complexity\n2. Resource usage\n3. Potential bottlenecks\n4. Optimization opportunities\n\nCODE:\n{{CODE}}"
    },
    {
      "name": "security-review",
      "backend": "gemini",
      "prompt": "Review this code for security issues.\n\nFocus on:\n1. Input validation\n2. Injection vulnerabilities\n3. Authentication/authorization\n4. OWASP risks\n\nCODE:\n{{CODE}}"
    }
  ]
}
```

## Step 2: Create Sample Code

Create a file named `sample.go`:

```go
package main

import (
    "database/sql"
    "fmt"
    "net/http"
)

func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    db, _ := sql.Open("postgres", "connection-string")
    query := "SELECT * FROM users WHERE id = " + id
    rows, _ := db.Query(query)
    defer rows.Close()
    fmt.Fprintf(w, "User found")
}
```

## Step 3: Execute the Review

Run the parallel review:

```bash
# Substitute code into tasks
CODE=$(cat sample.go | jq -Rs .)
jq --arg code "$CODE" '.tasks[].prompt |= sub("\\{\\{CODE\\}\\}"; $code)' review-tasks.json > review-request.json

# Execute parallel review
clinvk parallel -f review-request.json -o json > review-results.json

# View results
jq -r '.results[] | "\n=== \(.name | ascii_upcase) ===\n\(.output)"' review-results.json
```

## Step 4: Format the Report

Create a script to format the output:

```bash
#!/bin/bash
# format-review.sh

echo "# Code Review Report"
echo ""
echo "Generated: $(date)"
echo ""

jq -r '
  .results[] |
  "## \(.name | split("-") | map(ascii_upcase) | join(" "))\n" +
  "**Backend:** \(.backend)\n" +
  "**Duration:** \(.duration_ms // "N/A")ms\n\n" +
  "\(.output)\n" +
  "---\n"
' review-results.json
```

```bash
chmod +x format-review.sh
./format-review.sh > review-report.md
cat review-report.md
```

## Step 5: Automate for Git Hooks

Create a pre-commit hook:

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running AI code review..."

# Get staged Go files
STAGED=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')

if [ -z "$STAGED" ]; then
    exit 0
fi

# Run review on staged files
for file in $STAGED; do
    echo "Reviewing $file..."

    CODE=$(git show :"$file" | jq -Rs .)
    jq --arg code "$CODE" '.tasks[].prompt |= sub("\\{\\{CODE\\}\\}"; $code)' review-tasks.json > /tmp/review-request.json

    clinvk parallel -f /tmp/review-request.json -o json > /tmp/review-results.json

    # Check for critical issues
    if jq -e '.results[] | select(.output | contains("CRITICAL"))' /tmp/review-results.json > /dev/null; then
        echo "Critical issues found in $file!"
        jq -r '.results[].output' /tmp/review-results.json
        exit 1
    fi
done

exit 0
```

```bash
chmod +x .git/hooks/pre-commit
```

## Verification

Test your setup:

1. Make a change to `sample.go`
2. Stage it: `git add sample.go`
3. Try to commit: `git commit -m "test"`
4. The review should run automatically

## Next Steps

- Learn about [Chain Execution](../how-to/chain-execution.md) for sequential reviews
- See the [Automated Code Review](../use-cases/automated-code-review.md) use case for CI/CD integration
- Explore [AI Team Collaboration](../use-cases/ai-team-collaboration.md) for more complex workflows
