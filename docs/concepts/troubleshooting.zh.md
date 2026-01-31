---
title: 故障排除
description: clinvoker 的常见问题、诊断方法和解决方案。
---

# 故障排除

本指南涵盖使用 clinvoker 时可能遇到的常见问题，以及诊断方法和解决方案。问题按类别组织以便快速参考。

## 诊断方法

在深入了解具体问题之前，以下是通用的诊断方法：

### 启用调试模式

```bash
# 详细输出显示详细的执行信息
clinvk --verbose "your prompt"

# 调试模式包括后端命令输出
CLINVK_DEBUG=1 clinvk "your prompt"
```

### 检查系统状态

```bash
# 查看版本和构建信息
clinvk version

# 检查可用后端
clinvk config show

# 验证配置文件
clinvk config validate
```

### 试运行模式

查看将要执行的命令而不实际运行：

```bash
clinvk --dry-run "your prompt"
```

## 后端不可用

### 症状

- 错误：`Backend 'claude' not found in PATH`
- 错误：`No backends available`
- 退出码：126

### 原因

1. 后端 CLI 未安装
2. 后端 CLI 不在 PATH 中
3. 后端 CLI 权限不正确
4. 后端 CLI 版本与预期不符

### 解决方案

#### 1. 验证安装

```bash
# 检查后端 CLI 是否已安装
which claude codex gemini

# 检查版本
claude --version
codex --version
gemini --version
```

#### 2. 添加到 PATH

```bash
# 添加到 shell 配置文件（~/.bashrc、~/.zshrc 等）
export PATH="$PATH:/usr/local/bin"

# 重新加载配置文件
source ~/.bashrc  # 或 ~/.zshrc
```

#### 3. 检查权限

```bash
# 验证可执行权限
ls -la $(which claude)

# 如需要则修复
chmod +x /path/to/claude
```

#### 4. 验证 clinvk 检测

```bash
# 检查 clinvk 可以找到哪些后端
clinvk config show | grep -A2 "available"
```

## 会话问题

### 会话损坏

**症状：**
- 错误：`Failed to load session: invalid JSON`
- 会话文件显示为空或被截断
- 无法恢复之前正常工作的会话

**原因：**
1. 会话写入过程中进程被终止
2. 写入时磁盘已满
3. 没有适当锁定的情况下并发修改

**解决方案：**

```bash
# 检查会话文件完整性
ls -la ~/.clinvk/sessions/

# 查看会话内容（JSON 应该有效）
cat ~/.clinvk/sessions/<session-id>.json | python -m json.tool

# 如果损坏，删除该会话
clinvk sessions delete <session-id>

# 或清理所有会话
rm -rf ~/.clinvk/sessions/*
```

### 会话锁定问题

**症状：**
- 错误：`Failed to acquire store lock: timeout`
- 操作无限期挂起
- "Resource temporarily unavailable" 错误

**原因：**
1. 另一个 clinvk 进程持有锁
2. 之前的进程崩溃而未释放锁
3. 陈旧的锁文件

**解决方案：**

```bash
# 检查正在运行的 clinvk 进程
ps aux | grep clinvk

# 如有必要，终止陈旧进程
kill -9 <pid>

# 删除陈旧的锁文件（谨慎使用）
rm -f ~/.clinvk/.store.lock

# 验证锁文件权限
ls -la ~/.clinvk/
```

### 会话未找到

**症状：**
- 错误：`Session 'abc123' not found`
- 无法恢复之前的会话

**解决方案：**

```bash
# 列出所有可用会话
clinvk sessions list

# 检查会话目录
ls -la ~/.clinvk/sessions/

# 通过部分 ID 搜索会话
find ~/.clinvk/sessions/ -name "*abc*"
```

## API 服务器问题

### 服务器启动问题

**症状：**
- 错误：`Address already in use`
- 错误：绑定端口时 `Permission denied`
- 服务器立即退出

**解决方案：**

**端口已被占用：**

```bash
# 查找使用该端口的进程
lsof -i :8080
# 或
netstat -tlnp | grep 8080

# 使用不同端口
clinvk serve --port 3000

# 或终止现有进程
kill -9 <pid>
```

**权限被拒绝（低端口）：**

```bash
# 使用 > 1024 的端口（无需 root）
clinvk serve --port 8080

# 或使用 sudo（不推荐）
sudo clinvk serve --port 80
```

### 认证问题

**症状：**
- 错误：`Unauthorized` (401)
- 错误：`Invalid API key`

**解决方案：**

```bash
# 检查是否配置了 API 密钥
clinvk config show | grep api_keys

# 验证请求中的 API 密钥
curl -H "Authorization: Bearer YOUR_KEY" http://localhost:8080/api/v1/health

# 检查环境变量
echo $CLINVK_API_KEY

# 测试无认证（如果未配置密钥）
curl http://localhost:8080/health
```

### 速率限制问题

**症状：**
- 错误：`Rate limit exceeded` (429)
- 达到一定量后请求被拒绝

**解决方案：**

```bash
# 检查速率限制配置
clinvk config show | grep rate_limit

# 在配置中调整速率限制
clinvk config set server.rate_limit_rps 100

# 对于并行执行，降低并发度
clinvk parallel --max-parallel 2 --file tasks.json
```

## 性能问题

### 响应时间慢

**症状：**
- 命令耗时比预期长
- 发生超时

**原因：**
1. 后端速率限制
2. 上下文/提示词大小过大
3. 到后端 API 的网络延迟
4. 系统资源不足

**解决方案：**

```bash
# 检查后端状态
curl http://localhost:8080/api/v1/backends

# 监控系统资源
top
iostat -x 1

# 增加超时时间
clinvk config set timeout 300

# 使用更快的模型
clinvk -m fast "prompt"
```

### 内存使用过高

**症状：**
- clinvk 进程使用过多内存
- 系统变得无响应

**解决方案：**

```bash
# 检查内存使用
ps aux | grep clinvk

# 限制并行执行
clinvk parallel --max-parallel 2 --file tasks.json

# 清理旧会话
clinvk sessions clean --older-than 7d

# 使用临时模式（无会话存储）
clinvk --ephemeral "prompt"
```

### 磁盘空间问题

**症状：**
- 错误：`No space left on device`
- 会话操作失败

**解决方案：**

```bash
# 检查磁盘空间
df -h

# 检查会话存储大小
du -sh ~/.clinvk/sessions/

# 清理旧会话
clinvk sessions clean --older-than 30d

# 或手动清理
rm -rf ~/.clinvk/sessions/*.json
```

## 配置问题

### 配置文件未加载

**症状：**
- 设置未应用
- 使用默认值

**解决方案：**

```bash
# 检查配置文件位置
ls -la ~/.clinvk/config.yaml

# 验证 YAML 语法
cat ~/.clinvk/config.yaml | python -c "import yaml,sys; yaml.safe_load(sys.stdin)"

# 检查文件权限
chmod 600 ~/.clinvk/config.yaml

# 查看有效配置
clinvk config show

# 检查环境变量覆盖
echo $CLINVK_BACKEND
echo $CLINVK_TIMEOUT
```

### 环境变量未应用

**症状：**
- 环境设置被忽略
- CLI 标志有效但环境变量无效

**解决方案：**

```bash
# 验证变量是否已导出
export CLINVK_BACKEND=codex

# 检查当前 shell 变量
env | grep CLINVK

# 记住优先级：CLI 参数 > 环境变量 > 配置文件
```

## 平台特定问题

### macOS

**Gatekeeper 阻止：**

```bash
# 移除隔离属性
xattr -d com.apple.quarantine /path/to/clinvk

# 或在系统偏好设置 > 安全性与隐私中允许
```

**公证问题：**

```bash
# 如果下载的二进制文件无法运行
sudo spctl --master-disable  # 临时禁用（谨慎使用）
# 首次运行后重新启用
sudo spctl --master-enable
```

### Windows

**PATH 问题：**

```powershell
# 通过 PowerShell 添加到 PATH
$env:Path += ";C:\path\to\clinvk"

# 或通过系统属性永久添加
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\path\to\clinvk", "User")
```

**杀毒软件误报：**

某些杀毒软件可能会标记 clinvk。如有必要，请添加排除项。

### Linux

**权限被拒绝：**

```bash
chmod +x /path/to/clinvk
```

**SELinux 问题：**

```bash
# 检查 SELinux 状态
getenforce

# 如果为 enforcing，检查拒绝
ausearch -m avc -ts recent

# 如需要创建策略模块
audit2allow -a -M clinvk
semodule -i clinvk.pp
```

## 调试模式使用

### 启用综合日志

```bash
# 设置调试环境变量
export CLINVK_DEBUG=1

# 使用详细输出运行
clinvk --verbose "prompt"

# 服务器模式
CLINVK_DEBUG=1 clinvk serve
```

### 日志位置

**CLI 模式：**
- 默认情况下日志输出到 stderr
- 重定向到文件：`clinvk "prompt" 2> debug.log`

**服务器模式：**
- 日志输出到 stdout/stderr
- 使用 systemd/journald 持久化日志：`journalctl -u clinvk`
- 或重定向：`clinvk serve > server.log 2>&1`

### 常见日志模式

**后端命令：**
```text
[DEBUG] Executing: claude --print --model sonnet "prompt"
[DEBUG] Working directory: /home/user/project
[DEBUG] Exit code: 0
```

**会话操作：**
```text
[DEBUG] Loading session: abc123
[DEBUG] Acquiring file lock
[DEBUG] Session saved successfully
```

**API 请求：**
```text
[DEBUG] POST /api/v1/prompt
[DEBUG] Request body: {...}
[DEBUG] Response: 200 OK
```

## 获取帮助

如果尝试上述解决方案后问题仍然存在：

1. **检查文档**
   - [FAQ](faq.zh.md)
   - [指南](../guides/index.zh.md)
   - [参考](../reference/index.zh.md)

2. **搜索 Issues**
   - [GitHub Issues](https://github.com/signalridge/clinvoker/issues)
   - 使用错误消息进行搜索

3. **开新 Issue**
   包括：
   - clinvk 版本 (`clinvk version`)
   - 操作系统和版本
   - 后端版本 (`claude --version` 等)
   - 完整的错误消息
   - 复现步骤
   - 调试日志（如果可能）

4. **社区支持**
   - 发起 GitHub Discussion
   - 查看现有讨论中类似的问题

## 相关文档

- [FAQ](faq.zh.md) - 常见问题
- [设计决策](design-decisions.zh.md) - 架构说明
- [贡献指南](contributing.zh.md) - 开发设置
