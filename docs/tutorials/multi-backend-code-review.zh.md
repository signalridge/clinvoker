---
title: 多后端代码审查
description: 使用多个 AI 后端并行构建全面的代码审查系统，涵盖架构、性能和安全性分析。
---

# 教程：多后端代码审查

学习如何构建一个生产就绪的代码审查系统，同时利用 Claude、Codex 和 Gemini 对代码提供全面的反馈。这种方法结合了每个 AI 助手的独特优势，比任何单一后端都能提供更彻底的审查。

## 为什么使用多后端审查？

### 单一后端审查的问题

单个 AI 助手有不同的优势和劣势：

| 后端 | 优势 | 劣势 |
|------|------|------|
| Claude | 架构、推理、安全性 | 可能过于谨慎 |
| Codex | 代码生成、性能 | 对安全性关注较少 |
| Gemini | 安全性、广泛知识 | 可能遗漏实现细节 |

### 多后端解决方案

通过在多个后端并行运行审查，您可以获得：

1. **全面覆盖**：每个后端专注于它最擅长的领域
2. **交叉验证**：多个后端发现的问题优先级更高
3. **多元视角**：不同的 AI 方法发现不同的问题
4. **更快的反馈**：并行执行意味着无需等待

### 真实场景

假设您正在审查一个为 API 添加认证的 Pull Request：

- **Claude** 分析整体架构和设计模式
- **Codex** 检查性能瓶颈和实现效率
- **Gemini** 扫描安全漏洞和 OWASP 风险

综合反馈让您确信没有遗漏任何问题。

---

## 架构概述

```text
                    代码输入
                         |
         +---------------+---------------+
         |               |               |
    [Claude]        [Codex]        [Gemini]
         |               |               |
   架构审查         性能审查         安全审查
         |               |               |
         +---------------+---------------+
                         |
                 聚合结果
                         |
              可操作的反馈
```

### 工作原理

1. **输入**：代码或 diff 被准备并模板化到审查提示词中
2. **分发**：clinvoker 同时将代码发送到所有三个后端
3. **处理**：每个后端从其专业角度进行分析
4. **聚合**：结果被收集并格式化为统一报告
5. **输出**：开发人员收到全面、分类的反馈

---

## 前置要求

在开始之前，请确保您具备以下条件：

- clinvoker 已安装和配置（参见[快速开始](getting-started.zh.md)）
- 至少有两个后端可用（Claude、Codex 和/或 Gemini）
- 安装了 `jq` 用于 JSON 处理：`brew install jq` 或 `apt-get install jq`

验证您的设置：

```bash
clinvk config show
# 应显示可用的后端

jq --version
# 应显示 1.6 或更高版本
```

---

## 步骤 1：创建配置文件

创建 `review-config.yaml` 以定义每个后端如何贡献审查：

```yaml
# 多后端代码审查的配置
review:
  name: "全面代码审查"
  version: "1.0"

  # 每个后端都有专门的角色
templates:
  architecture:
    backend: claude
    prompt: |
      您是一位高级软件架构师，正在审查代码的设计质量。

      审查以下代码的：
      1. SOLID 原则遵循情况
      2. 设计模式使用（是否合适？是否正确实现？）
      3. 代码组织和模块化
      4. 可维护性和可读性
      5. API 设计（如适用）
      6. 错误处理策略

      要审查的代码：
      ```{{language}}
      {{code}}
      ```

      请按以下格式提供您的发现：
      - **严重**：必须修复的问题
      - **警告**：应该解决的问题
      - **建议**：考虑的改进
      - **优点**：做得好的地方

  performance:
    backend: codex
    prompt: |
      您是一位性能工程师，正在审查代码的效率。

      审查以下代码的：
      1. 算法复杂度（大 O 分析）
      2. 资源使用（内存、CPU、I/O）
      3. 数据库查询效率
      4. 缓存机会
      5. 并发问题
      6. 潜在瓶颈

      要审查的代码：
      ```{{language}}
      {{code}}
      ```

      请提供具体的优化建议。

  security:
    backend: gemini
    prompt: |
      您是一位安全工程师，正在审查代码的漏洞。

      审查以下代码的：
      1. 输入验证和清理
      2. SQL 注入风险
      3. XSS 和 CSRF 漏洞
      4. 认证/授权缺陷
      5. OWASP Top 10 风险
      6. 代码中的密钥或凭证
      7. 不安全的依赖项

      要审查的代码：
      ```{{language}}
      {{code}}
      ```

      将发现分类为严重、高、中或低风险。
```

### 配置结构说明

| 部分 | 用途 |
|------|------|
| `review` | 关于审查过程的元数据 |
| `templates.architecture` | Claude 的角色 - 设计和模式 |
| `templates.performance` | Codex 的角色 - 效率和优化 |
| `templates.security` | Gemini 的角色 - 漏洞和风险 |

`{{code}}` 和 `{{language}}` 占位符将在运行时被实际代码替换。

---

## 步骤 2：创建审查脚本

创建 `run-review.sh` 来编排多后端审查：

```bash
#!/bin/bash
# 多后端代码审查脚本

set -e

# 配置
CONFIG_FILE="${CONFIG_FILE:-review-config.yaml}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-json}"

# 显示用法
usage() {
    echo "Usage: $0 [options] <file-or-directory>"
    echo ""
    echo "Options:"
    echo "  -c, --config    配置文件 (默认: review-config.yaml)"
    echo "  -o, --output    输出格式: text, json, markdown (默认: json)"
    echo "  -h, --help      显示此帮助"
    echo ""
    echo "Examples:"
    echo "  $0 src/auth.go"
    echo "  $0 -o markdown src/"
    exit 1
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

if [ -z "$TARGET" ]; then
    echo "Error: 未指定目标"
    usage
fi

# 从文件扩展名检测语言
detect_language() {
    local file="$1"
    local ext="${file##*.}"
    case "$ext" in
        go) echo "go" ;;
        py) echo "python" ;;
        js) echo "javascript" ;;
        ts) echo "typescript" ;;
        rs) echo "rust" ;;
        java) echo "java" ;;
        *) echo "text" ;;
    esac
}

# 准备代码输入
if [ -f "$TARGET" ]; then
    CODE=$(cat "$TARGET")
    LANGUAGE=$(detect_language "$TARGET")
    FILENAME=$(basename "$TARGET")
elif [ -d "$TARGET" ]; then
    # 对于目录，创建摘要
    CODE=$(find "$TARGET" -type f \( -name "*.go" -o -name "*.py" -o -name "*.js" -o -name "*.ts" \) -exec echo "=== {} ===" \; -exec head -50 {} \;)
    LANGUAGE="mixed"
    FILENAME=$(basename "$TARGET")
else
    echo "Error: 目标未找到: $TARGET"
    exit 1
fi

# 为 JSON 转义代码
CODE_JSON=$(echo "$CODE" | jq -Rs '.')

# 创建并行任务文件
echo "创建审查任务..."
cat > /tmp/review-tasks.json << EOF
{
  "tasks": [
    {
      "name": "architecture-review",
      "backend": "claude",
      "prompt": $(yq e '.templates.architecture.prompt' "$CONFIG_FILE" | sed "s/{{language}}/$LANGUAGE/g" | sed "s/{{code}}/$CODE_JSON/g" | jq -Rs '.'),
      "output_format": "text"
    },
    {
      "name": "performance-review",
      "backend": "codex",
      "prompt": $(yq e '.templates.performance.prompt' "$CONFIG_FILE" | sed "s/{{language}}/$LANGUAGE/g" | sed "s/{{code}}/$CODE_JSON/g" | jq -Rs '.'),
      "output_format": "text"
    },
    {
      "name": "security-review",
      "backend": "gemini",
      "prompt": $(yq e '.templates.security.prompt' "$CONFIG_FILE" | sed "s/{{language}}/$LANGUAGE/g" | sed "s/{{code}}/$CODE_JSON/g" | jq -Rs '.'),
      "output_format": "text"
    }
  ]
}
EOF

# 运行并行审查
echo "运行多后端审查..."
clinvk parallel -f /tmp/review-tasks.json -o json > /tmp/review-results.json

# 根据请求的格式格式化输出
case "$OUTPUT_FORMAT" in
    text)
        echo ""
        echo "========================================"
        echo "代码审查报告: $FILENAME"
        echo "========================================"
        echo ""
        jq -r '.results[] | "\n=== \(.name | ascii_upcase) ===\n后端: \(.backend)\n耗时: \(.duration_ms // "N/A")ms\n\n\(.output)\n"' /tmp/review-results.json
        ;;
    markdown)
        echo "# 代码审查报告"
        echo ""
        echo "**文件:** \`$FILENAME\`"
        echo "**语言:** $LANGUAGE"
        echo "**生成时间:** $(date -u +"%Y-%m-%d %H:%M:%S UTC")"
        echo ""
        jq -r '.results[] | "## \(.name | split("-") | map(ascii_upcase) | join(" "))\n\n**后端:** \(.backend)\n**耗时:** \(.duration_ms // "N/A")ms\n\n\(.output)\n\n---\n"' /tmp/review-results.json
        ;;
    json)
        cat /tmp/review-results.json
        ;;
    *)
        echo "未知输出格式: $OUTPUT_FORMAT"
        exit 1
        ;;
esac

echo ""
echo "审查完成!"
```

使脚本可执行：

```bash
chmod +x run-review.sh
```

---

## 步骤 3：创建测试用的示例代码

创建包含故意问题的 `sample-auth.go` 来测试审查系统：

```go
package main

import (
    "database/sql"
    "fmt"
    "net/http"
    "time"
)

// User 表示系统中的用户
type User struct {
    ID       int
    Username string
    Password string // 明文存储 - 安全问题
    Email    string
}

// AuthHandler 处理认证请求
type AuthHandler struct {
    db *sql.DB
}

// Login 认证用户
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    password := r.URL.Query().Get("password")

    // SQL 注入漏洞
    query := fmt.Sprintf("SELECT id, username, password FROM users WHERE username='%s' AND password='%s'", username, password)

    row := h.db.QueryRow(query)

    var user User
    err := row.Scan(&user.ID, &user.Username, &user.Password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // 设置 session cookie，没有 HttpOnly 或 Secure 标志
    http.SetCookie(w, &http.Cookie{
        Name:  "session",
        Value: fmt.Sprintf("user_%d", user.ID),
    })

    fmt.Fprintf(w, "Welcome, %s!", user.Username)
}

// GetUser 通过 ID 获取用户
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // 没有认证检查 - 任何人都可以访问任何用户
    id := r.URL.Query().Get("id")

    // 另一个 SQL 注入
    query := "SELECT * FROM users WHERE id = " + id
    rows, _ := h.db.Query(query)
    defer rows.Close()

    // 低效：将所有用户加载到内存中
    var users []User
    for rows.Next() {
        var u User
        rows.Scan(&u.ID, &u.Username, &u.Password, &u.Email)
        users = append(users, u)
    }

    // 没有速率限制 - 容易受到暴力破解攻击
    for i := 0; i < len(users); i++ {
        fmt.Fprintf(w, "User: %s\n", users[i].Username)
    }
}

// 没有连接池的全局数据库连接
var globalDB *sql.DB

func init() {
    var err error
    // 硬编码凭证 - 安全问题
    globalDB, err = sql.Open("postgres", "postgres://admin:secret123@localhost/mydb?sslmode=disable")
    if err != nil {
        panic(err)
    }
}
```

这段代码在架构、性能和安全性维度包含多个问题。

---

## 步骤 4：运行审查

执行多后端审查：

```bash
./run-review.sh -o markdown sample-auth.go > review-report.md
cat review-report.md
```

### 预期的架构审查 (Claude)

Claude 应该识别：

- **严重**：没有适当生命周期管理的全局数据库连接
- **警告**：AuthHandler 中混合的关注点（认证 + 用户检索）
- **建议**：使用依赖注入而不是全局变量
- **优点**：User 的清晰结构定义

### 预期的性能审查 (Codex)

Codex 应该识别：

- **严重**：没有连接池配置
- **警告**：没有分页就将所有用户加载到内存中
- **建议**：添加查询超时和上下文取消
- **优化**：重复查询的预处理语句

### 预期的安全审查 (Gemini)

Gemini 应该识别：

- **严重**：Login 和 GetUser 中的 SQL 注入漏洞
- **严重**：明文密码存储
- **高**：硬编码的数据库凭证
- **高**：cookie 缺少 HttpOnly 和 Secure 标志
- **中**：认证端点没有速率限制
- **中**：GetUser 端点没有认证检查

---

## 步骤 5：将结果组合成可操作的反馈

创建 `aggregate-review.sh` 来智能组合结果：

```bash
#!/bin/bash
# 聚合和优先处理多后端审查结果

RESULTS_FILE="${1:-/tmp/review-results.json}"

if [ ! -f "$RESULTS_FILE" ]; then
    echo "Error: 结果文件未找到: $RESULTS_FILE"
    exit 1
fi

echo "# 优先代码审查报告"
echo ""
echo "生成时间: $(date)"
echo ""

# 从所有后端提取严重问题
echo "## 严重问题 (需要立即处理)"
echo ""

echo "### 安全性 (Gemini)"
jq -r '.results[] | select(.name == "security-review") | .output' "$RESULTS_FILE" | grep -i -E "(critical|sql injection|hardcoded)" || echo "未发现严重安全问题。"
echo ""

echo "### 架构 (Claude)"
jq -r '.results[] | select(.name == "architecture-review") | .output' "$RESULTS_FILE" | grep -i -E "(\*\*critical|global variable)" || echo "未发现严重架构问题。"
echo ""

echo "### 性能 (Codex)"
jq -r '.results[] | select(.name == "performance-review") | .output' "$RESULTS_FILE" | grep -i -E "(\*\*critical|memory leak)" || echo "未发现严重性能问题。"
echo ""

# 摘要统计
echo "## 摘要统计"
echo ""
echo "| 后端 | 耗时 | 状态 |"
echo "|------|------|------|"
jq -r '.results[] | "| \(.backend) | \(.duration_ms // "N/A")ms | \(.exit_code | if . == 0 then "成功" else "失败" end) |"' "$RESULTS_FILE"
echo ""

# 完整报告
echo "## 详细报告"
echo ""
jq -r '.results[] | "### \(.name | split("-") | map(ascii_upcase) | join(" "))\n\n**后端:** \(.backend)\n\n\(.output)\n"' "$RESULTS_FILE"
```

---

## 步骤 6：并行执行内部原理

### clinvoker 如何并行执行

当您运行 `clinvk parallel` 时，内部发生以下过程：

```text
1. 解析任务文件
   |
2. 验证所有后端是否可用
   |
3. 创建工作池 (默认: 3 个工作者)
   |
4. 向工作池提交任务
   |-- 任务 1: Claude (架构) --> 工作者 1
   |-- 任务 2: Codex (性能) --> 工作者 2
   |-- 任务 3: Gemini (安全) --> 工作者 3
   |
5. 等待所有任务完成
   |
6. 将结果聚合为 JSON 响应
```

### 工作池配置

在您的配置中控制并行度：

```yaml
parallel:
  max_workers: 3        # 并发任务数
  fail_fast: false      # 即使一个任务失败也继续
  aggregate_output: true # 组合所有输出
```

### 执行保证

- **隔离**：每个任务独立运行
- **超时**：可配置的任务超时
- **错误处理**：失败的任务不会阻塞其他任务
- **排序**：结果保持任务顺序以确保可预测的输出

---

## 步骤 7：CI/CD 集成实现自动化审查

### GitHub Actions 集成

创建 `.github/workflows/multi-backend-review.yml`：

```yaml
name: 多后端代码审查

on:
  pull_request:
    paths:
      - "**.go"
      - "**.py"
      - "**.js"
      - "**.ts"

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: 安装 clinvoker
        run: |
          curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: 获取变更文件
        id: changed
        run: |
          files=$(git diff --name-only origin/${{ github.base_ref }}...HEAD | grep -E '\.(go|py|js|ts)$' || true)
          echo "files=$files" >> $GITHUB_OUTPUT

      - name: 运行多后端审查
        if: steps.changed.outputs.files != ''
        run: |
          for file in ${{ steps.changed.outputs.files }}; do
            echo "审查 $file..."
            ./run-review.sh -o markdown "$file" >> review-output.md
          done

      - name: 发布审查评论
        if: steps.changed.outputs.files != ''
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const body = fs.readFileSync('review-output.md', 'utf8');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body.substring(0, 65536) // GitHub 评论限制
            });

      - name: 检查严重问题
        if: steps.changed.outputs.files != ''
        run: |
          if grep -i "critical" review-output.md; then
            echo "::error::发现严重问题!"
            exit 1
          fi
```

### GitLab CI 集成

创建 `.gitlab-ci.yml`：

```yaml
multi-backend-review:
  stage: test
  image: alpine/curl
  variables:
    CLINVK_BACKEND: claude
  before_script:
    - apk add --no-cache jq bash git
    - curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash
  script:
    - |
      if [ "$CI_MERGE_REQUEST_IID" ]; then
        git fetch origin $CI_MERGE_REQUEST_TARGET_BRANCH_NAME
        files=$(git diff --name-only origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME...HEAD | grep -E '\.(go|py|js|ts)$' || true)

        for file in $files; do
          ./run-review.sh "$file"
        done
      fi
  rules:
    - if: $CI_MERGE_REQUEST_IID
```

### Jenkins Pipeline

创建 `Jenkinsfile`：

```groovy
pipeline {
    agent any

    stages {
        stage('多后端审查') {
            when {
                changeRequest()
            }
            steps {
                script {
                    def changedFiles = sh(
                        script: "git diff --name-only origin/${env.CHANGE_TARGET}...HEAD | grep -E '\\.(go|py|js|ts)$' || true",
                        returnStdout: true
                    ).trim()

                    if (changedFiles) {
                        sh 'curl -sSL https://raw.githubusercontent.com/signalridge/clinvoker/main/install.sh | bash'

                        changedFiles.split('\n').each { file ->
                            sh "./run-review.sh -o markdown '${file}' >> review.md"
                        }

                        publishHTML([
                            reportDir: '.',
                            reportFiles: 'review.md',
                            reportName: 'AI 代码审查'
                        ])
                    }
                }
            }
        }
    }
}
```

---

## 最佳实践

### 1. 审查范围管理

限制审查范围以获得更好的结果：

```bash
# 审查特定函数
clinvk -b claude "审查 auth.go 中的 Login 函数的架构问题"

# 按 diff 审查
clinvk -b codex "审查此 diff 的性能问题: $(git diff HEAD~1)"
```

### 2. 结果优先级排序

按严重性和共识加权发现：

| 发现类型 | 权重 | 操作 |
|---------|------|------|
| 严重 (2+ 后端) | 阻止合并 | 必须修复 |
| 严重 (1 后端) | 警告 | 应该修复 |
| 警告 (2+ 后端) | 警告 | 考虑修复 |
| 建议 | 信息 | 可选 |

### 3. 审查模板

在版本控制中维护模板：

```bash
git add review-config.yaml run-review.sh
git commit -m "添加多后端审查配置"
```

### 4. 成本优化

- 对 CI/CD 使用 `--ephemeral` 标志以避免会话开销
- 仅限制审查变更的文件
- 缓存未变更文件的审查结果
- 使用较小的模型进行初始筛选

---

## 故障排除

### 问题：审查耗时太长

**解决方案**：增加并行度或使用更快的模型：

```yaml
parallel:
  max_workers: 5

backends:
  claude:
    model: claude-sonnet-4-20250514  # 比 Opus 更快
```

### 问题：结果不一致

**解决方案**：向提示词添加明确说明：

```yaml
prompt: |
  在您的发现中要具体。包括：
  - 适用的行号
  - 演示问题的代码片段
  - 修复的具体建议
```

### 问题：后端超时

**解决方案**：在配置中增加超时：

```yaml
unified_flags:
  command_timeout_secs: 600  # 10 分钟
```

---

## 后续步骤

- 了解[链式执行](../guides/chains.zh.md)以进行顺序审查
- 探索[CI/CD 集成](ci-cd-integration.zh.md)以进行生产部署
- 查看[构建 AI Skills](building-ai-skills.zh.md)以获取自定义审查代理
- 查看[架构概述](../concepts/architecture.zh.md)以了解内部原理

---

## 总结

您已经构建了一个全面的多后端代码审查系统，该系统：

1. 利用 Claude 进行架构分析
2. 使用 Codex 进行性能优化
3. 使用 Gemini 进行安全审计
4. 并行运行所有审查以提高效率
5. 与 CI/CD 集成以实现自动化反馈

这种方法比任何单一 AI 助手单独提供的代码审查都更彻底、更可靠。
