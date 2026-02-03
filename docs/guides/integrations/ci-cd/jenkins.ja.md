---
title: Jenkins 連携
description: Jenkins と clinvoker を連携し、CI/CD ワークフローを実現します。
---

# Jenkins 連携

Jenkins と clinvoker を連携し、コードレビューやドキュメント生成を自動化します。

## パイプライン例

```groovy
pipeline {
    agent any

    environment {
        CLINVK_SERVER = credentials('clinvk-server-url')
        CLINVK_API_KEY = credentials('clinvk-api-key')
    }

    stages {
        stage('AI Code Review') {
            when {
                changeRequest()
            }
            steps {
                sh '''
                    # Install clinvoker if needed
                    curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash

                    # Run code review
                    git diff origin/${CHANGE_TARGET}...HEAD > pr.diff
                    # Use clinvoker for review
                '''
            }
        }
    }
}
```

## 関連項目

- [CI/CD 統合チュートリアル](../../../tutorials/ci-cd-integration.md)
- [自動コードレビュー](../../../tutorials/ci-cd-integration.md)
