# 故障排除

clinvk 常见问题和解决方案。

## 后端问题

### 后端未找到

**错误：** `Backend 'codex' not found in PATH`

**解决方案：**

1. 验证后端 CLI 已安装：
   ```bash
   which claude codex gemini
   ```

2. 将二进制文件位置添加到 PATH：
   ```bash
   export PATH="$PATH:/path/to/backend"
   ```

3. 检查 clinvk 检测：
   ```bash
   clinvk config show | grep available
   ```

### 后端不可用

**错误：** `Backend unavailable` 或退出码 126

**解决方案：**

- 检查后端的 API 是否可访问
- 验证后端 CLI 的 API 凭证已配置
- 尝试直接运行后端 CLI 以诊断

### 模型未找到

**错误：** `Model 'invalid-model' not found`

**解决方案：**

1. 列出后端可用的模型
2. 更新配置：
   ```bash
   clinvk config set backends.claude.model claude-opus-4-5-20251101
   ```

## 配置问题

### 配置文件未加载

**症状：** 设置未应用

**解决方案：**

1. 检查配置文件位置：
   ```bash
   ls -la ~/.clinvk/config.yaml
   ```

2. 验证 YAML 语法：
   ```bash
   cat ~/.clinvk/config.yaml | python -c "import yaml,sys; yaml.safe_load(sys.stdin)"
   ```

3. 查看有效配置：
   ```bash
   clinvk config show
   ```

## 会话问题

### 无法恢复会话

**错误：** `Session 'abc123' not found`

**解决方案：**

1. 列出可用会话：
   ```bash
   clinvk sessions list
   ```

2. 检查会话目录：
   ```bash
   ls ~/.clinvk/sessions/
   ```

3. 会话可能已被清理。创建新会话。

### 会话存储已满

**解决方案：**

清理旧会话：

```bash
clinvk sessions clean --older-than 7d
```

## 服务器问题

### 端口已被占用

**错误：** `Address already in use`

**解决方案：**

1. 找到使用该端口的进程：
   ```bash
   lsof -i :8080
   ```

2. 使用不同端口：
   ```bash
   clinvk serve --port 3000
   ```

## 执行问题

### 命令超时

**错误：** 请求或命令超时

**解决方案：**

1. 增加配置中的超时：
   ```yaml
   server:
     request_timeout_secs: 600
   ```

2. 对于复杂任务，拆分为更小的提示

### 速率限制

**错误：** Rate limit exceeded

**解决方案：**

1. 等待后重试
2. 对 compare 命令使用 `--sequential`
3. 减少并行工作器

## 调试

### 启用详细输出

```bash
clinvk --verbose "提示"
```

### 试运行模式

查看将执行的命令：

```bash
clinvk --dry-run "提示"
```

### 检查版本

验证运行的是预期版本：

```bash
clinvk version
```

## 获取帮助

如果问题持续：

1. 检查 [GitHub Issues](https://github.com/signalridge/clinvoker/issues)
2. 搜索类似问题
3. 开新 issue，包含：
   - clinvk 版本
   - 操作系统和版本
   - 错误消息
   - 复现步骤
