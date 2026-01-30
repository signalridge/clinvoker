# 环境变量

clinvk 支持的环境变量。

## 变量列表

| 变量 | 说明 | 默认值 |
|----------|-------------|---------|
| `CLINVK_BACKEND` | 默认后端 | `claude` |
| `CLINVK_CLAUDE_MODEL` | Claude 模型 |（后端默认）|
| `CLINVK_CODEX_MODEL` | Codex 模型 |（后端默认）|
| `CLINVK_GEMINI_MODEL` | Gemini 模型 |（后端默认）|
| `CLINVK_API_KEYS` | 服务器 API Key（逗号分隔） |（未设置）|
| `CLINVK_API_KEYS_GOPASS_PATH` | gopass 路径 |（未设置）|

## 优先级

1. CLI 参数
2. 环境变量
3. 配置文件
4. 默认值

## 示例

```bash
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3

clinvk "优化这个函数"
```

### 服务器鉴权

```bash
export CLINVK_API_KEYS="key1,key2"
clinvk serve
```

设置 `CLINVK_API_KEYS` 后，所有 API 请求都必须携带 key。
