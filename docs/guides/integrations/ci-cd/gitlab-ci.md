---
title: GitLab CI Integration
description: Integrate clinvoker with GitLab CI for automated workflows.
---

# GitLab CI Integration

Integrate clinvoker with GitLab CI for automated code review and documentation generation.

## Basic Configuration

```yaml
ai-code-review:
  stage: test
  image: alpine/curl
  variables:
    CLINVK_SERVER: $CLINVK_SERVER_URL
    CLINVK_API_KEY: $CLINVK_API_KEY
  script:
    - apk add --no-cache jq git
    - git fetch origin $CI_MERGE_REQUEST_TARGET_BRANCH_NAME
    - git diff origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME...HEAD > pr.diff
    # Run clinvoker review
  rules:
    - if: $CI_MERGE_REQUEST_IID
```text

## See Also

- [CI/CD Integration Tutorial](../../../tutorials/ci-cd-integration.md)
- [Automated Code Review](../../../tutorials/ci-cd-integration.md)
