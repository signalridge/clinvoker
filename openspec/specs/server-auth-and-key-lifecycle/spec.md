# server-auth-and-key-lifecycle Specification

## Purpose

TBD - created by archiving change v05-auth-request-traceability. Update Purpose after archive.

## Requirements

### Requirement: [Feature #6] 鉴权失败响应携带 `request_id`

系统 MUST 在鉴权失败响应中返回 `request_id`，并保持 header/body 一致。

#### Scenario: missing key

- **WHEN** 请求缺少 API key 且服务开启鉴权
- **THEN** 响应 body MUST 包含 `request_id`
- **THEN** 响应头请求标识 MUST 与 body 一致

#### Scenario: invalid key

- **WHEN** 请求 API key 非法
- **THEN** 响应 MUST 返回明确错误语义与 `request_id`

### Requirement: [Feature #6] 错误结构稳定

系统 MUST 为鉴权失败提供稳定可解析的结构化错误字段。

#### Scenario: 稳定字段

- **WHEN** 任一鉴权失败发生
- **THEN** 响应 MUST 包含 `code`、`message`、`request_id`
- **THEN** 字段变更 MUST 通过版本文档声明
