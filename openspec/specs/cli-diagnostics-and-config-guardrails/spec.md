# cli-diagnostics-and-config-guardrails Specification

## Purpose

TBD - created by archiving change v05-cli-diagnostics-guardrails. Update Purpose after archive.

## Requirements

### Requirement: [Feature #1] `clinvk doctor` 聚合自检

系统 MUST 提供 `clinvk doctor` 命令，对运行环境和关键配置做聚合检查。

#### Scenario: 基础自检

- **WHEN** 用户执行 `clinvk doctor`
- **THEN** 系统 MUST 检查 backend 可用性、配置合法性、会话存储可访问性
- **THEN** 每个检查项 MUST 输出 `pass/warn/fail` 状态和建议

#### Scenario: JSON 输出契约

- **WHEN** 用户执行 `clinvk doctor --json`
- **THEN** 输出 MUST 包含稳定字段与 `summary`
- **THEN** 输出 MUST 包含 `schema_version` 字段

### Requirement: [Feature #8] `clinvk config lint` 静态校验

系统 MUST 提供 `config lint` 命令并复用统一配置校验器。

#### Scenario: 默认路径校验

- **WHEN** 用户执行 `clinvk config lint`
- **THEN** 系统 MUST 按现有配置加载优先级校验默认配置
- **THEN** 系统 MUST 返回全部错误而非首错即停

#### Scenario: 指定文件与错误语义

- **WHEN** 用户执行 `clinvk config lint --config <path>`
- **THEN** 系统 MUST 校验指定文件
- **THEN** 文件不可读或不存在时 MUST 返回非零退出码与明确错误信息

### Requirement: [Feature #7] `serve` 启动安全状态摘要

系统 MUST 在服务启动时输出关键安全状态并标注高风险组合。

#### Scenario: 启动摘要输出

- **WHEN** `clinvk serve` 启动完成
- **THEN** 日志 MUST 包含 auth、rate limit、metrics、trusted proxies、bind host、CORS 状态

#### Scenario: 高风险 warning

- **WHEN** 出现高风险组合（例如外网绑定且无 auth）
- **THEN** 系统 MUST 输出显式 warning 与修复建议
