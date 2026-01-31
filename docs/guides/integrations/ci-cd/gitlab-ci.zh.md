---
title: GitLab CI 集成
description: 将 clinvoker 与 GitLab CI 集成用于自动化工作流。
---

# GitLab CI 集成

将 clinvoker 与 GitLab CI 集成，实现自动化代码审查和文档生成。

## 基本配置

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
    # 运行 clinvoker 审查
  rules:
    - if: $CI_MERGE_REQUEST_IID
```

## 另请参阅

- [CI/CD 集成教程](../../../tutorials/ci-cd-integration.md)
