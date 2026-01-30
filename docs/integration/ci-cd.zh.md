# CI/CD

在 CI 中使用 clinvk 做评审、总结或生成测试。

## 直接调用 CLI

```yaml
- name: AI 评审（并行）
  run: |
    cat > tasks.json <<'JSON'
    {"tasks":[
      {"backend":"claude","prompt":"评审架构"},
      {"backend":"codex","prompt":"评审性能"}
    ]}
    JSON
    clinvk parallel -f tasks.json
```

任一任务失败将返回非零退出码。

## 在 CI 中调用 HTTP Server

先启动服务：

```bash
clinvk serve --host 0.0.0.0 --port 8080
```

然后调用 API：

```bash
curl -sS http://localhost:8080/api/v1/compare \
  -H 'Content-Type: application/json' \
  -d '{"backends":["claude","gemini"],"prompt":"Review diff"}'
```

## 建议

- CI 中建议使用 `--ephemeral` 或 API `ephemeral: true`。
- 若开启服务器鉴权，请把 API Key 放入 CI Secret。
