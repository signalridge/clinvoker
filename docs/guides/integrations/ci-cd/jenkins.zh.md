---
title: Jenkins 集成
description: 将 clinvoker 与 Jenkins 集成用于 CI/CD 工作流。
---

# Jenkins 集成

将 clinvoker 与 Jenkins 集成，实现自动化代码审查和文档生成。

## Pipeline 示例

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
                    # 安装 clinvoker
                    curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash

                    # 运行代码审查
                    git diff origin/${CHANGE_TARGET}...HEAD > pr.diff
                    # 使用 clinvoker 进行审查
                '''
            }
        }
    }
}
```bash

## 另请参阅

- [CI/CD 集成教程](../../../tutorials/ci-cd-integration.md)
```text
</thinking>

现在更新 mkdocs.yml 以匹配新的严格 Diátaxis 结构：

<function=Edit>
<parameter=file_path>/Users/yixianlu/ghq/github.com/signalridge/clinvoker/mkdocs.yml
