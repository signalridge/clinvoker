---
title: GitLab CI 連携
description: GitLab CI と clinvoker を連携し、ワークフローを自動化します。
---

# GitLab CI 連携

GitLab CI と clinvoker を連携し、コードレビューやドキュメント生成を自動化します。

## 基本設定

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
```

## 関連項目

- [CI/CD 統合チュートリアル](../../../tutorials/ci-cd-integration.md)
- [自動コードレビュー](../../../tutorials/ci-cd-integration.md)
