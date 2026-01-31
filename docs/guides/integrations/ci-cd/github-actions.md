---
title: GitHub Actions Integration
description: Integrate clinvoker with GitHub Actions for CI/CD workflows.
---

# GitHub Actions Integration

Integrate clinvoker with GitHub Actions for automated code review, documentation generation, and more.

## Basic Workflow

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
```text

## Advanced Configuration

See the [CI/CD Integration Tutorial](../../../tutorials/ci-cd-integration.md) for detailed setup instructions.

## Related Documentation

- [Automated Code Review](../../../tutorials/ci-cd-integration.md)
- [CI/CD Integration Tutorial](../../../tutorials/ci-cd-integration.md)
