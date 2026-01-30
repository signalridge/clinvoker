# 典型用例

下面是贴近真实场景的用例，所有示例都对应现有命令和 API。

## 1) 多模型并行评审（Parallel）

同时获得架构、安全、性能多个视角。

```json
{
  "tasks": [
    {"backend": "claude", "prompt": "从架构与设计角度评审这个改动。"},
    {"backend": "codex", "prompt": "从性能角度评审潜在瓶颈与风险。"},
    {"backend": "gemini", "prompt": "从安全角度评审潜在漏洞与风险。"}
  ],
  "max_parallel": 3,
  "fail_fast": false
}
```

```bash
clinvk parallel -f tasks.json
```

为什么有效：`parallel` 始终为无状态执行，并汇总输出。

## 2) 定位 → 修复 → 复核 → 总结（Chain）

通过链式流程把上下文传递给下一步。

```json
{
  "steps": [
    {"name": "analyze", "backend": "claude", "prompt": "定位 bug 并解释根因。"},
    {"name": "fix", "backend": "codex", "prompt": "根据以下分析修复：{{previous}}"},
    {"name": "verify", "backend": "gemini", "prompt": "复核修复是否引入回归：{{previous}}"},
    {"name": "summary", "backend": "claude", "prompt": "总结修复点与原因：{{previous}}"}
  ]
}
```

```bash
clinvk chain -f chain.json
```

为什么有效：chain 仅通过 `{{previous}}` 传递输出，并且保持无状态。

## 3) 高风险改动前的对比验证

对关键改动先比较多个后端的结论。

```bash
clinvk compare --all-backends "这个迁移安全吗？请列出具体风险。"
```

如需机器读取，添加 `--json`。

## 4) CI 审查机器人（HTTP Server）

暴露 API，让 CI 直接调用：

```bash
clinvk serve --host 0.0.0.0 --port 8080
```

在 CI 中调用：

```bash
curl -sS http://localhost:8080/api/v1/parallel \
  -H 'Content-Type: application/json' \
  -d '{"tasks":[{"backend":"claude","prompt":"Review diff"},{"backend":"codex","prompt":"Suggest tests"}]}'
```

为什么有效：自定义 REST API 同时支持 `prompt` / `parallel` / `chain` / `compare`。

## 5) 复用 SDK，不改业务代码

只替换 base URL，即可复用现有 SDK：

```python
from openai import OpenAI

client = OpenAI(base_url="http://localhost:8080/openai/v1", api_key="not-needed")
resp = client.chat.completions.create(
    model="claude-opus-4-5-20251101",
    messages=[{"role": "user", "content": "解释这个函数。"}]
)
print(resp.choices[0].message.content)
```

为什么有效：OpenAI 兼容端点是无状态的，`model` 字段用于路由。

## 6) Claude Code Skill 中转站

在 Claude Code Skills 中把任务分发给其它后端：

```bash
clinvk -b gemini --ephemeral "总结这个数据集"
clinvk -b codex --ephemeral "为这个补丁生成测试"
```

为什么有效：`clinvk` 用统一参数封装多种 CLI。

---

下一步：查看 [并行执行](parallel-execution.md) 或 [HTTP 服务器](http-server.md)
