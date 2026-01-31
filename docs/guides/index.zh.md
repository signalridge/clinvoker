---
title: 指南
description: 有效使用 clinvoker 的分步指南。
---

# 指南

欢迎使用 clinvoker 指南。这些指南提供了常见任务和工作流的分步说明。

## 快速开始

刚接触 clinvoker？从这里开始：

- [基本用法](basic-usage.md) - 学习使用 clinvoker 的基础知识
- [配置](configuration.md) - 自定义 clinvoker 设置
- [会话管理](sessions.md) - 使用持久会话

## 执行模式

了解执行 AI 任务的不同方式：

- [并行执行](parallel.md) - 同时运行多个提示
- [链式执行](chains.md) - 创建顺序工作流
- [HTTP 服务器](http-server.md) - 将 clinvoker 作为 API 服务器运行

## 后端

探索支持的 AI 后端：

- [后端概述](backends/index.md) - 比较所有可用后端
- [Claude Code](backends/claude.md) - Anthropic 的 Claude Code 集成
- [Codex CLI](backends/codex.md) - OpenAI 的 Codex CLI 集成
- [Gemini CLI](backends/gemini.md) - Google 的 Gemini CLI 集成
- [后端对比](compare.md) - 后端详细对比

## 集成

将 clinvoker 与现有工具连接：

- [集成概述](integrations/index.md) - 所有集成选项
- [Claude Code Skills](integrations/claude-code-skills.md) - 创建可复用的 skills
- [LangChain/LangGraph](integrations/langchain-langgraph.md) - 构建 LLM 应用
- [OpenAI SDK](integrations/openai-sdk.md) - 使用 OpenAI 兼容客户端
- [Anthropic SDK](integrations/anthropic-sdk.md) - 使用 Anthropic 兼容客户端
- [MCP Server](integrations/mcp-server.md) - Model Context Protocol 集成
- [CI/CD 平台](integrations/ci-cd/index.md) - 使用 CI/CD 自动化

## 选择指南

不确定从哪里开始？根据您的目标选择：

| 目标 | 推荐指南 |
|------|----------|
| 运行第一个提示 | [基本用法](basic-usage.md) |
| 对比 AI 响应 | [后端对比](compare.md) |
| 自动化代码审查 | [CI/CD 集成](integrations/ci-cd/index.md) |
| 构建 AI 工作流 | [链式执行](chains.md) |
| 部署为服务 | [HTTP 服务器](http-server.md) |
| 与我的应用集成 | [集成概述](integrations/index.md) |
