---
title: GitHub Actions 連携
description: GitHub Actions と clinvoker を連携し、CI/CD ワークフローを構築します。
---

# GitHub Actions 連携

GitHub Actions と clinvoker を連携し、コードレビューやドキュメント生成などを自動化します。

## 基本ワークフロー

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
          # Run review using clinvoker
```

## 高度な設定

詳しい設定手順は [CI/CD 統合チュートリアル](../../../tutorials/ci-cd-integration.md) を参照してください。

## 関連ドキュメント

- [自動コードレビュー](../../../tutorials/ci-cd-integration.md)
- [CI/CD 統合チュートリアル](../../../tutorials/ci-cd-integration.md)
