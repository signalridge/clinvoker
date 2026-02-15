# session-governance-and-discovery Specification

## Purpose

TBD - created by archiving change v05-session-list-clean. Update Purpose after archive.

## Requirements

### Requirement: [Feature #4] `sessions list --json` 稳定输出

系统 MUST 提供稳定且可脚本化的会话列表 JSON 输出。

#### Scenario: 基础字段

- **WHEN** 用户执行 `clinvk sessions list --json`
- **THEN** 每个会话项 MUST 包含 `id`、`backend`、`status`、`last_used`、`model`、`tags`
- **THEN** 输出 MUST 包含过滤上下文字段

#### Scenario: 顺序与分页

- **WHEN** 多次执行相同查询
- **THEN** 结果排序 MUST 稳定（默认按 `last_used` 倒序）
- **THEN** 输出 MUST 包含分页语义字段

### Requirement: [Feature #5] `sessions clean --dry-run` 无副作用预览

系统 MUST 提供 dry-run 模式并保证无删除副作用。

#### Scenario: 预览结果

- **WHEN** 用户执行 `sessions clean --older-than <value> --dry-run`
- **THEN** 系统 MUST 输出候选会话数量与样例标识
- **THEN** 系统 MUST NOT 删除任何会话数据或索引记录

#### Scenario: 差异可解释

- **WHEN** dry-run 后执行真实清理
- **THEN** 若结果不一致，系统 SHOULD 给出可解释原因（例如并发写入）
