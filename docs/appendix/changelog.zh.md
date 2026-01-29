# 更新日志

clinvk 的所有重要更改都记录在此。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

### 新增
- MkDocs 文档站点，使用 Material 主题
- 双语支持（英文和中文）
- 全面的用户指南和参考文档

### 变更
- 文档重构为 getting-started、user-guide、server、reference、development 和 appendix 部分

## [0.1.0] - 2025-01-27

### 新增

- 首次发布
- 多后端支持（Claude Code、Codex CLI、Gemini CLI）
- 跨后端统一配置选项
- 会话持久化和恢复功能
- 并行任务执行，支持快速失败
- 后端对比功能
- 顺序管道的链式执行
- HTTP API 服务器，支持三种 API 风格：
  - 自定义 REST API (`/api/v1/`)
  - OpenAI 兼容 API (`/openai/v1/`)
  - Anthropic 兼容 API (`/anthropic/v1/`)
- 配置级联（CLI → 环境变量 → 配置文件 → 默认值）
- 跨平台支持（Linux、macOS、Windows）
- 临时（无状态）模式用于一次性查询

### 后端

- Claude Code 后端，支持审批和沙箱模式
- Codex CLI 后端
- Gemini CLI 后端

### 命令

- `clinvk [prompt]` - 执行提示
- `clinvk resume` - 恢复会话
- `clinvk sessions` - 管理会话
- `clinvk config` - 管理配置
- `clinvk parallel` - 并行执行
- `clinvk compare` - 后端对比
- `clinvk chain` - 链式执行
- `clinvk serve` - HTTP 服务器
- `clinvk version` - 版本信息

---

## 版本历史

| 版本 | 日期 | 亮点 |
|------|------|------|
| 0.1.0 | 2025-01-27 | 首次发布 |

## 链接

- [GitHub Releases](https://github.com/signalridge/clinvoker/releases)
- [GitHub Commits](https://github.com/signalridge/clinvoker/commits/main)
