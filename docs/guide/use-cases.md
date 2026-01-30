# Use Cases

Practical, real-world scenarios where clinvk shines. Each use case includes ready-to-run commands and JSON configurations.

---

## Table of Contents

1. [Code Review Workflows](#code-review-workflows)
2. [Development Pipelines](#development-pipelines)
3. [Quality Assurance](#quality-assurance)
4. [Documentation](#documentation)
5. [DevOps & CI/CD](#devops--cicd)
6. [Research & Analysis](#research--analysis)
7. [Learning & Teaching](#learning--teaching)

---

## Code Review Workflows

### 1. Multi-Dimensional Code Review

Get comprehensive feedback from different perspectives simultaneously.

```
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "Review the following code for architecture and design patterns. Focus on: single responsibility, dependency management, API design, and extensibility.",
      "id": "architecture-review"
    },
    {
      "backend": "codex",
      "prompt": "Analyze this code for performance bottlenecks. Look for: algorithmic complexity, memory leaks, unnecessary allocations, and I/O inefficiencies.",
      "id": "performance-review"
    },
    {
      "backend": "gemini",
      "prompt": "Security audit: identify potential vulnerabilities including injection risks, unsafe deserialization, insecure dependencies, and data exposure.",
      "id": "security-review"
    }
  ],
  "max_parallel": 3,
  "fail_fast": false,
  "aggregate_output": true
}
```

```
clinvk parallel -f multi-review.json --output-dir ./reviews
```

**Why it works**: Parallel execution with `fail_fast: false` ensures all reviews complete even if one backend encounters an error.

---

### 2. Risk Assessment for Critical Changes

Before merging high-risk changes, get consensus from all backends.

```
clinvk compare --all-backends --json "Review this database migration script for risks:
$(cat migration.sql)" > risk-assessment.json
```

Then analyze the consensus:

```
# Extract key concerns from each backend
jq -r '.results[] | "\(.backend): \(.error // .content[:100])"' risk-assessment.json
```

**Best for**: Database migrations, API deprecations, infrastructure changes.

---

### 3. Dependency Update Review

Review dependency updates for compatibility and security.

```
{
  "tasks": [
    {
      "backend": "gemini",
      "prompt": "Check changelog of express@5.0.0 for breaking changes that might affect a typical REST API",
      "id": "changelog-review"
    },
    {
      "backend": "codex",
      "prompt": "Review package.json for outdated dependencies and suggest safe update order",
      "id": "dependency-analysis"
    },
    {
      "backend": "claude",
      "prompt": "Generate a migration plan for updating to the latest versions with rollback strategy",
      "id": "migration-plan"
    }
  ]
}
```

---

## Development Pipelines

### 4. Bug Fix Pipeline

End-to-end bug resolution workflow with context passing.

```
{
  "steps": [
    {
      "name": "analyze",
      "backend": "claude",
      "prompt": "Analyze the error logs and code to find the root cause. Explain what's happening and why.",
      "max_turns": 5
    },
    {
      "name": "fix",
      "backend": "codex",
      "prompt": "Based on this analysis, implement the minimal fix: {{previous}}",
      "approval_mode": "auto"
    },
    {
      "name": "test",
      "backend": "gemini",
      "prompt": "Write unit tests to verify this fix and prevent regression: {{previous}}"
    },
    {
      "name": "document",
      "backend": "claude",
      "prompt": "Document the fix in CHANGELOG format: {{previous}}"
    }
  ],
  "stop_on_failure": true,
  "pass_working_dir": true
}
```

```
# With error logs as context
clinvk chain -f bugfix-pipeline.json --input logs.txt
```

**Pipeline flow**: Analyze → Fix → Test → Document

---

### 5. Feature Implementation Chain

Break down complex features into manageable steps.

```
{
  "steps": [
    {
      "name": "design-api",
      "backend": "claude",
      "prompt": "Design the API schema for a user notification system with preferences"
    },
    {
      "name": "implement-backend",
      "backend": "codex",
      "prompt": "Implement the backend service based on this design: {{previous}}",
      "workdir": "./backend"
    },
    {
      "name": "implement-frontend",
      "backend": "gemini",
      "prompt": "Create React components for this notification system: {{previous}}",
      "workdir": "./frontend"
    },
    {
      "name": "integration-tests",
      "backend": "claude",
      "prompt": "Write integration tests for the complete notification flow: {{previous}}"
    }
  ]
}
```

---

### 6. Refactoring Assistant

Safe refactoring with verification at each step.

```
# Step 1: Get refactoring plan
clinvk -b claude "Suggest refactoring plan for extracting authentication logic into middleware" > plan.md

# Step 2: Implement refactoring
clinvk -b codex "Implement this refactoring: $(cat plan.md)" --approval-mode auto

# Step 3: Verify behavior preserved
clinvk -b gemini "Verify all authentication flows still work after refactoring"

# Step 4: Code review the changes
clinvk compare --all-backends "Review the refactored authentication middleware"
```

---

## Quality Assurance

### 7. Security Audit Suite

Comprehensive security analysis using parallel backends.

```
{
  "tasks": [
    {
      "backend": "gemini",
      "prompt": "OWASP Top 10 audit: check for injection, broken auth, sensitive data exposure, XXE, broken access control, security misconfiguration, XSS, insecure deserialization, known vulnerabilities, insufficient logging",
      "id": "owasp-check"
    },
    {
      "backend": "claude",
      "prompt": "Review authentication and authorization flows. Check for: token handling, session management, privilege escalation risks, MFA implementation",
      "id": "auth-audit"
    },
    {
      "backend": "codex",
      "prompt": "Static analysis for hardcoded secrets, unsafe crypto usage, insecure randomness, and dependency vulnerabilities",
      "id": "secrets-scan"
    }
  ],
  "output_dir": "./security-reports"
}
```

---

### 8. Performance Benchmarking

Compare backend performance on the same task.

```
#!/bin/bash
# benchmark.sh - Compare backend performance

PROMPT="Refactor this recursive Fibonacci to iterative implementation"
FILE="fibonacci.py"

for backend in claude codex gemini; do
  echo "Testing $backend..."
  time clinvk -b $backend --output-format json "$PROMPT" < $FILE > result-$backend.json
done

# Compare results
echo "Performance Summary:"
for f in result-*.json; do
  backend=$(echo $f | cut -d'-' -f2 | cut -d'.' -f1)
  duration=$(jq -r '.duration_seconds' $f)
  echo "$backend: ${duration}s"
done
```

---

### 9. Test Generation Pipeline

Generate comprehensive test suites.

```
{
  "steps": [
    {
      "name": "unit-tests",
      "backend": "codex",
      "prompt": "Generate unit tests covering: happy path, edge cases, error conditions. Use Jest with 100% coverage target."
    },
    {
      "name": "integration-tests",
      "backend": "claude",
      "prompt": "Based on these unit tests, create integration tests for API endpoints: {{previous}}"
    },
    {
      "name": "e2e-tests",
      "backend": "gemini",
      "prompt": "Write Playwright E2E tests for the critical user journey: {{previous}}"
    }
  ]
}
```

---

## Documentation

### 10. Code-to-Documentation Pipeline

Automatically generate and sync documentation.

```
{
  "steps": [
    {
      "name": "api-docs",
      "backend": "claude",
      "prompt": "Generate OpenAPI spec from the source code and routes"
    },
    {
      "name": "readme",
      "backend": "gemini",
      "prompt": "Create a comprehensive README based on the OpenAPI spec: {{previous}}"
    },
    {
      "name": "chinese-docs",
      "backend": "claude",
      "prompt": "Translate and adapt the README for Chinese developers: {{previous}}"
    },
    {
      "name": "examples",
      "backend": "codex",
      "prompt": "Generate code examples in Python, JavaScript, and Go based on the API: {{previous}}"
    }
  ],
  "output_dir": "./generated-docs"
}
```

---

### 11. Architecture Decision Records (ADRs)

Generate consistent ADRs from discussions.

```
# Start with architecture discussion
clinvk -b claude "Help me decide between microservices and monolith for this e-commerce platform"

# Generate ADR
cat discussion.txt | clinvk -b claude "Format this as an Architecture Decision Record in Markdown"

# Peer review
clinvk compare --backends claude,codex "Review this ADR for completeness and clarity" < adr.md
```

---

## DevOps & CI/CD

### 12. GitHub Actions Review Bot

Automated PR reviews using clinvk HTTP API.

```
# .github/workflows/ai-review.yml
name: AI Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Get PR diff
        run: git diff origin/main > pr.diff

      - name: Request multi-backend review
        run: |
          curl -sS http://clinvk-server:8080/api/v1/parallel \
            -H 'Content-Type: application/json' \
            -H "X-API-Key: ${{ secrets.CLINVK_API_KEY }}" \
            -d @- << 'EOF'
          {
            "tasks": [
              {"backend": "claude", "prompt": "Review PR for architecture: $(cat pr.diff)"},
              {"backend": "codex", "prompt": "Review PR for performance: $(cat pr.diff)"},
              {"backend": "gemini", "prompt": "Review PR for security: $(cat pr.diff)"}
            ]
          }
          EOF
```

---

### 13. Infrastructure as Code Review

Review Terraform/CloudFormation changes.

```
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "Review Terraform plan for AWS architecture best practices: cost optimization, high availability, security groups",
      "workdir": "./terraform"
    },
    {
      "backend": "gemini",
      "prompt": "Check for security misconfigurations in Terraform: public S3 buckets, overly permissive IAM, unencrypted storage",
      "workdir": "./terraform"
    },
    {
      "backend": "codex",
      "prompt": "Estimate costs and suggest optimizations for this infrastructure",
      "workdir": "./terraform"
    }
  ]
}
```

---

### 14. Deployment Safety Check

Pre-deployment validation chain.

```
{
  "steps": [
    {
      "name": "changelog-check",
      "backend": "claude",
      "prompt": "Verify CHANGELOG.md is updated for version 2.5.0"
    },
    {
      "name": "migration-review",
      "backend": "gemini",
      "prompt": "Review database migrations for backward compatibility and rollback safety"
    },
    {
      "name": "api-compatibility",
      "backend": "codex",
      "prompt": "Check for breaking API changes and document them"
    },
    {
      "name": "deployment-plan",
      "backend": "claude",
      "prompt": "Generate deployment checklist with rollback procedure"
    }
  ],
  "stop_on_failure": true
}
```

---

## Research & Analysis

### 15. Technology Evaluation

Compare technologies with structured analysis.

```
# Evaluate database options
clinvk compare --all-backends --json "Compare PostgreSQL vs MongoDB vs DynamoDB for a high-write analytics workload" > db-comparison.json

# Generate decision matrix
jq -r '.results[] | "\(.backend):\n\(.content)\n---"' db-comparison.json
```

---

### 16. Legacy Code Analysis

Understand and document legacy systems.

```
{
  "steps": [
    {
      "name": "understand",
      "backend": "claude",
      "prompt": "Analyze this legacy COBOL-to-Java conversion. Explain the business logic in modern terms."
    },
    {
      "name": "map-dependencies",
      "backend": "codex",
      "prompt": "Create a dependency graph and identify external system integrations: {{previous}}"
    },
    {
      "name": "modernization-plan",
      "backend": "gemini",
      "prompt": "Suggest modernization approach with risk assessment: {{previous}}"
    },
    {
      "name": "migration-steps",
      "backend": "claude",
      "prompt": "Break down the migration into phases with milestones: {{previous}}"
    }
  ]
}
```

---

### 17. Competitive Analysis

Analyze competitors' technical approaches.

```
# Analyze multiple competitors
for competitor in "competitor-a" "competitor-b" "competitor-c"; do
  curl -s "https://api.github.com/repos/$competitor/tech-stack" | \
    clinvk -b claude "Analyze this tech stack and suggest our competitive advantages" \
    > analysis-$competitor.txt
done

# Synthesize findings
clinvk chain -f synthesis.json
```

---

## Learning & Teaching

### 18. Code Explanation for Onboarding

Generate explanations at different skill levels.

```
{
  "tasks": [
    {
      "backend": "gemini",
      "prompt": "Explain this Python metaclass implementation to a junior developer (1-2 years experience)",
      "id": "junior"
    },
    {
      "backend": "claude",
      "prompt": "Explain the same metaclass code to a mid-level developer, focusing on design decisions and trade-offs",
      "id": "mid-level"
    },
    {
      "backend": "codex",
      "prompt": "Explain the metaclass implementation to a senior engineer, focusing on Python internals and performance implications",
      "id": "senior"
    }
  ]
}
```

---

### 19. Interview Question Generator

Generate contextual interview questions.

```
# Based on your actual codebase
clinvk -b claude "Generate 5 senior backend engineer interview questions based on challenges in this codebase"

# With solutions
clinvk -b codex "Provide model solutions for these interview questions"

# Evaluation rubric
clinvk -b gemini "Create an evaluation rubric for grading responses"
```

---

### 20. Learning Path Generation

Create personalized learning paths.

```
{
  "steps": [
    {
      "name": "skill-assessment",
      "backend": "claude",
      "prompt": "Analyze this codebase and identify the key skills a developer needs to contribute effectively"
    },
    {
      "name": "curriculum",
      "backend": "gemini",
      "prompt": "Create a 12-week learning curriculum for these skills: {{previous}}"
    },
    {
      "name": "resources",
      "backend": "codex",
      "prompt": "Recommend specific books, courses, and documentation for each week: {{previous}}"
    },
    {
      "name": "projects",
      "backend": "claude",
      "prompt": "Suggest hands-on projects that apply these skills to this codebase: {{previous}}"
    }
  ]
}
```

---

## Advanced Patterns

### 21. Multi-Stage Decision Making

Use backends for different decision stages.

```
# Stage 1: Data gathering
clinvk -b gemini "Gather metrics and logs for the last 24 hours" > metrics.txt

# Stage 2: Pattern analysis
clinvk -b claude "Identify anomalies in these metrics: $(cat metrics.txt)" > anomalies.txt

# Stage 3: Root cause
clinvk -b codex "Determine root causes for these anomalies: $(cat anomalies.txt)" > root-causes.txt

# Stage 4: Solution comparison
clinvk compare --all-backends "Evaluate these solution options: $(cat root-causes.txt)"
```

---

### 22. A/B Testing Prompts

Compare how different backends handle the same prompt variations.

```
#!/bin/bash
# prompt-ab-test.sh

BASE_PROMPT="Explain recursion"

for variation in "" "to a 5-year-old" "with code examples" "using math notation"; do
  echo "=== Variation: $variation ==="
  clinvk compare --all-backends "$BASE_PROMPT $variation" | jq -r '.results[] | "\(.backend): \(.content[:200])"'
  echo
done
```

---

### 23. Confidence Voting

Use multiple backends for critical decisions.

```
{
  "tasks": [
    {
      "backend": "claude",
      "prompt": "Should we migrate from REST to GraphQL? Rate confidence 1-10 and explain.",
      "id": "claude-vote"
    },
    {
      "backend": "codex",
      "prompt": "Should we migrate from REST to GraphQL? Rate confidence 1-10 and explain.",
      "id": "codex-vote"
    },
    {
      "backend": "gemini",
      "prompt": "Should we migrate from REST to GraphQL? Rate confidence 1-10 and explain.",
      "id": "gemini-vote"
    }
  ]
}
```

Then analyze consensus programmatically:

```
clinvk parallel -f vote.json --output-format json | \
  jq -r '.results[] | "\(.backend): \(.content)"' | \
  grep -oE '[0-9]+/10' | \
  awk '{sum+=$1; count++} END {print "Average confidence:", sum/count "/10"}'
```

---

## Tips for Production Use

### Configuration Management

Create environment-specific configs:

```
# ~/.clinvk/config.yaml
default_backend: claude

parallel:
  max_workers: 5
  fail_fast: false

server:
  host: "127.0.0.1"
  port: 8080
  rate_limit_enabled: true
  rate_limit_rps: 10
```

### Session Organization

Use tags for better session management:

```
# Create tagged sessions
clinvk --tag "project:api-refactor" --tag "priority:high" "Design new API structure"

# List with filter
clinvk sessions list --tag "project:api-refactor"

# Export project sessions
clinvk sessions export --tag "project:api-refactor" -o api-refactor-sessions.json
```

### Error Handling in Pipelines

Always handle failures in chains:

```
{
  "steps": [...],
  "stop_on_failure": true
}
```

Check exit codes:

```
clinvk chain -f pipeline.json
if [ $? -ne 0 ]; then
  echo "Pipeline failed at step $(jq '.failed_step' result.json)"
  exit 1
fi
```

---

## Next Steps

- [Parallel Execution Details](parallel-execution.md)
- [Chain Execution Details](chain-execution.md)
- [Backend Comparison](backend-comparison.md)
- [HTTP Server](http-server.md)
