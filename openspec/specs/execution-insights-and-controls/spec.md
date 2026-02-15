# execution-insights-and-controls Specification

## Purpose

TBD - created by archiving change v05-execution-insights-controls. Update Purpose after archive.

## Requirements

### Requirement: [Feature #3] chain 步骤级超时

系统 MUST 支持在 chain 步骤中配置 `timeout_secs`。

#### Scenario: 超时失败与停链

- **WHEN** 步骤执行超过 `timeout_secs`
- **THEN** 步骤 MUST 标记为超时失败并记录原因
- **THEN** 当 `stop_on_failure=true` 时 MUST 停止后续步骤

#### Scenario: 全局与步骤超时并存

- **WHEN** 同时配置全局超时和步骤超时
- **THEN** 系统 MUST 采用更短超时

### Requirement: [Feature #2] compare 摘要评分

系统 MUST 提供 compare 摘要评分并公开排序依据。

#### Scenario: 文本摘要

- **WHEN** 执行 compare（非 JSON）
- **THEN** 输出 MUST 包含 latency、status、output_length、score
- **THEN** 输出 MUST 说明排序规则

#### Scenario: JSON 摘要

- **WHEN** 执行 compare --json
- **THEN** 输出 MUST 包含评分输入维度与排名字段

### Requirement: [Feature #9] prompt usage 可选显示

系统 MUST 支持文本模式按需显示 usage/token，且默认保持兼容。

#### Scenario: 显式显示

- **WHEN** 用户开启 usage 显示开关
- **THEN** 输出 MUST 包含 input/output/total token 信息

#### Scenario: usage 缺失

- **WHEN** backend 未返回 usage
- **THEN** 输出 MUST 标注 unknown，而非误导性零值
