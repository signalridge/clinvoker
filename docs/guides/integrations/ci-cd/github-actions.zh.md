---
title: GitHub Actions 集成
description: 将 clinvoker 与 GitHub Actions 集成用于 CI/CD 工作流。
---

# GitHub Actions 集成

将 clinvoker 与 GitHub Actions 集成，实现自动化代码审查、文档生成等功能。

## 基本工作流

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

      - name: Install clinvoker
        run: |
          curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Run AI Review
        env:
          CLINVK_SERVER: ${{ secrets.CLINVK_SERVER_URL }}
          CLINVK_API_KEY: ${{ secrets.CLINVK_API_KEY }}
        run: |
          git diff origin/${{ github.base_ref }}...HEAD > pr.diff
          # 使用 clinvoker 运行审查
```

## 高级配置

请参阅 [CI/CD 集成教程](../../../tutorials/ci-cd-integration.md) 获取详细设置说明。

## 相关文档

- [CI/CD 集成教程](../../../tutorials/ci-cd-integration.md)
