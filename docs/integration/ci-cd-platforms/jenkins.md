---
title: Jenkins Integration
description: Integrate clinvoker with Jenkins for CI/CD workflows.
---

# Jenkins Integration

Integrate clinvoker with Jenkins for automated code review and documentation generation.

## Pipeline Example

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

## See Also

- [CI/CD Integration Tutorial](../../tutorials/ci-cd-integration.md)
- [Automated Code Review](../../use-cases/automated-code-review.md)
