---
title: Use Cases
description: Real-world scenarios and patterns for using clinvoker in production environments.
---

# Use Cases

This section provides detailed, production-ready use cases that demonstrate how to leverage clinvoker's capabilities in real-world scenarios. Each use case includes:

- **Scenario description** - When and why to use this pattern
- **Architecture** - How the components work together
- **Implementation** - Step-by-step setup and configuration
- **Examples** - Code samples and configuration files
- **Best practices** - Tips for production deployment

## Available Use Cases

<div class="grid cards" markdown>

-   :material-account-group:{ .lg .middle } __AI Team Collaboration__

    ---

    Simulate a multi-AI development team where different models take on
    specialized roles: Claude as architect, Codex as implementer, Gemini as reviewer.

    [:octicons-arrow-right-24: Read more](ai-team-collaboration.md)

-   :material-file-code:{ .lg .middle } __Automated Code Review__

    ---

    CI/CD pipeline that automatically sends pull request diffs to multiple
    backends for comprehensive code reviews covering architecture, performance,
    and security perspectives.

    [:octicons-arrow-right-24: Read more](automated-code-review.md)

-   :material-book-search:{ .lg .middle } __Multi-Model Research__

    ---

    Complex technical research using multiple AI perspectives. Run the same
    research question across different backends and synthesize comprehensive
    reports from diverse viewpoints.

    [:octicons-arrow-right-24: Read more](multi-model-research.md)

-   :material-file-document:{ .lg .middle } __Smart Documentation__

    ---

    Automated documentation generation pipeline. Extract insights from your
    codebase and generate architecture docs, API references, and usage examples
    using specialized backends for each task.

    [:octicons-arrow-right-24: Read more](smart-documentation.md)

-   :material-test-tube:{ .lg .middle } __Test Generation Pipeline__

    ---

    End-to-end automated test generation: design test cases, implement test
    code, and review coverage using a chain of specialized AI backends.

    [:octicons-arrow-right-24: Read more](test-generation-pipeline.md)

-   :material-server:{ .lg .middle } __API Gateway Pattern__

    ---

    Deploy clinvk as a centralized AI gateway for your organization. Route
    requests to different backends based on model parameters, with unified
    authentication, monitoring, and quota management.

    [:octicons-arrow-right-24: Read more](api-gateway-pattern.md)

-   :material-alert:{ .lg .middle } __Incident Response War Room__

    ---

    Triage production incidents with parallel analysis and a synthesis chain.

    [:octicons-arrow-right-24: Read more](incident-response-war-room.md)

</div>

## Use Case Selection Guide

Not sure which use case fits your needs? Here's a quick decision matrix:

| If you need... | Use this pattern |
|----------------|------------------|
| Multiple AIs working together on a single task | [AI Team Collaboration](ai-team-collaboration.md) |
| Automated PR reviews in CI/CD | [Automated Code Review](automated-code-review.md) |
| Comprehensive research across model perspectives | [Multi-Model Research](multi-model-research.md) |
| Auto-generated docs from code | [Smart Documentation](smart-documentation.md) |
| Automated test creation | [Test Generation Pipeline](test-generation-pipeline.md) |
| Centralized AI service for your org | [API Gateway Pattern](api-gateway-pattern.md) |
| Incident triage and rapid mitigation | [Incident Response War Room](incident-response-war-room.md) |

## Common Patterns Across Use Cases

### Parallel Execution Pattern

Many use cases leverage parallel execution for efficiency:

```bash
# Send the same task to multiple backends simultaneously
clinvk parallel -f tasks.json
```

```json title="tasks.json"
{
  "tasks": [
    {"backend": "claude", "prompt": "Analyze architecture..."},
    {"backend": "codex", "prompt": "Analyze performance..."},
    {"backend": "gemini", "prompt": "Analyze security..."}
  ]
}
```

### Chain Execution Pattern

Sequential processing with context passing:

```bash
# Process data through multiple stages
clinvk chain -f pipeline.json
```

```json title="pipeline.json"
{
  "steps": [
    {"name": "extract", "backend": "gemini", "prompt": "Extract data..."},
    {"name": "transform", "backend": "codex", "prompt": "Transform: {{previous}}"},
    {"name": "load", "backend": "claude", "prompt": "Review: {{previous}}"}
  ]
}
```

### Backend Comparison Pattern

Evaluate different approaches:

```bash
# Compare responses from all available backends
clinvk compare --all-backends "How should I implement caching?"
```

## Production Considerations

When deploying these use cases in production:

1. **Rate Limiting** - Configure appropriate rate limits to avoid overwhelming backends
2. **Error Handling** - Implement retry logic and fallback mechanisms
3. **Monitoring** - Use the metrics endpoint for observability
4. **Security** - Enable API key authentication and workdir restrictions
5. **Resource Management** - Set appropriate timeouts and max turns

See the [API Gateway Pattern](api-gateway-pattern.md) for a comprehensive production deployment example.
