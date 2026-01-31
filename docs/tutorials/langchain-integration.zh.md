---
title: LangChain 集成
description: 将 clinvoker 连接到 LangChain 和 LangGraph，构建具有多个 AI 后端的复杂代理工作流。
---

# 教程：LangChain 集成

学习如何将 clinvoker 与 LangChain 和 LangGraph 集成，构建能够利用多个后端的复杂 AI 应用程序。

## 为什么要与 LangChain 集成？

### 可组合性的力量

LangChain 提供了一个通过可组合组件构建 LLM 应用程序的框架。通过集成 clinvoker，您可以获得：

1. **统一后端访问**：使用 LangChain 熟悉的接口，同时路由到 Claude、Codex 或 Gemini
2. **链式组合**：在顺序或并行链中组合多个 AI 后端
3. **代理工作流**：构建能够为每个任务选择最佳后端的自主代理
4. **生态系统兼容性**：访问 LangChain 丰富的工具和集成生态系统

### 集成架构

```mermaid
flowchart TB
    subgraph APP["您的应用程序"]
        LC["LangChain / LangGraph"]
        CHAINS["链式"]
        AGENTS["代理"]
        TOOLS["工具调用"]
        CHAT["ChatOpenAI"]

        CHAINS --> AGENTS --> TOOLS --> CHAT
    end

    CHAT -->|HTTP/REST| CLINVK["clinvoker Server<br/>/openai/v1/chat/completions"]

    CLINVK -->|路由| CLAUDE["Claude CLI"]
    CLINVK -->|路由| CODEX["Codex CLI"]
    CLINVK -->|路由| GEMINI["Gemini CLI"]

    style APP fill:#e3f2fd,stroke:#1976d2
    style CLINVK fill:#ffecb3,stroke:#ffa000
    style CLAUDE fill:#f3e5f5,stroke:#7b1fa2
    style CODEX fill:#e8f5e9,stroke:#388e3c
    style GEMINI fill:#ffebee,stroke:#c62828
```yaml

---

## 前置要求

在与 LangChain 集成之前：

- 安装 Python 3.9 或更高版本
- clinvoker 服务器正在运行（本地或远程）
- 对 LangChain 概念有基本了解
- 为您的 AI 后端配置 API 密钥

### 安装依赖

```bash
pip install langchain langchain-openai langgraph
```text

### 验证 clinvoker 服务器

确保您的 clinvoker 服务器可访问：

```bash
# 测试服务器连接
curl http://localhost:8080/health

# 预期响应
{"status":"ok"}
```yaml

---

## 了解 OpenAI 兼容端点

### 为什么选择 OpenAI 兼容性？

clinvoker 提供 OpenAI 兼容的 API 端点（`/openai/v1`），因为：

1. **行业标准**：OpenAI 的 API 是最广泛支持的接口
2. **LangChain 支持**：LangChain 对 OpenAI 兼容 API 有一流支持
3. **工具生态系统**：大多数 AI 工具支持 OpenAI API 格式
4. **易于迁移**：现有的 OpenAI 代码只需最少更改即可工作

### 端点结构

| clinvoker 端点 | OpenAI 等效端点 | 用途 |
|---------------|----------------|------|
| `/openai/v1/chat/completions` | `/v1/chat/completions` | 聊天完成 |
| `/openai/v1/models` | `/v1/models` | 列出可用模型 |

### 模型映射

在 clinvoker 中，模型名称映射到后端：

| 模型名称 | 后端 | 最适合 |
|---------|------|--------|
| `claude` | Claude Code | 架构、推理 |
| `codex` | Codex CLI | 代码生成 |
| `gemini` | Gemini CLI | 安全、研究 |

---

## 步骤 1：基本 LangChain 集成

### 使用 clinvoker 配置 ChatOpenAI

创建 `basic_integration.py`：

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, SystemMessage

# 配置 LangChain 使用 clinvoker
llm = ChatOpenAI(
    model_name="claude",  # 路由到 Claude 后端
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",  # clinvoker 处理认证
    temperature=0.7,
    max_tokens=2000,
)

# 简单调用
messages = [
    SystemMessage(content="您是一个有帮助的编程助手。"),
    HumanMessage(content="解释微服务架构的好处。")
]

response = llm.invoke(messages)
print(response.content)
```text

### 工作原理

1. LangChain 发送请求到 `openai_api_base`
2. clinvoker 将 OpenAI 格式转换为后端特定格式
3. 指定的后端（Claude、Codex 或 Gemini）处理请求
4. clinvoker 以 OpenAI 格式返回响应
5. LangChain 像往常一样接收和处理响应

---

## 步骤 2：多后端链

### 并行链执行

创建 `parallel_chain.py`：

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
from langchain_core.runnables import RunnableParallel

# 为不同后端定义 LLM
claude_llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    temperature=0.7,
)

codex_llm = ChatOpenAI(
    model_name="codex",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    temperature=0.5,
)

gemini_llm = ChatOpenAI(
    model_name="gemini",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    temperature=0.7,
)

# 创建并行审查链
review_chain = RunnableParallel(
    architecture=lambda x: claude_llm.invoke([
        HumanMessage(content=f"审查架构：\n{x['code']}")
    ]),
    implementation=lambda x: codex_llm.invoke([
        HumanMessage(content=f"审查实现：\n{x['code']}")
    ]),
    security=lambda x: gemini_llm.invoke([
        HumanMessage(content=f"安全审计：\n{x['code']}")
    ]),
)

# 要审查的示例代码
code = """
def authenticate(user, password):
    query = f"SELECT * FROM users WHERE user='{user}' AND pass='{password}'"
    return db.execute(query)
"""

# 执行并行审查
results = review_chain.invoke({"code": code})

print("=== 架构审查 (Claude) ===")
print(results["architecture"].content)
print("\n=== 实现审查 (Codex) ===")
print(results["implementation"].content)
print("\n=== 安全审查 (Gemini) ===")
print(results["security"].content)
```text

### 顺序链执行

创建 `sequential_chain.py`：

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
from langchain_core.runnables import RunnableSequence

# 定义 LLM
claude = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

codex = ChatOpenAI(
    model_name="codex",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

# 创建顺序链：设计 -> 实现 -> 审查
def design_step(inputs):
    """Claude 设计架构"""
    response = claude.invoke([
        HumanMessage(content=f"为以下需求设计解决方案：{inputs['requirement']}")
    ])
    return {"design": response.content, "requirement": inputs["requirement"]}

def implement_step(inputs):
    """Codex 基于设计实现"""
    response = codex.invoke([
        HumanMessage(content=f"实现以下设计：\n{inputs['design']}")
    ])
    return {"implementation": response.content, "design": inputs["design"]}

def review_step(inputs):
    """Claude 审查实现"""
    response = claude.invoke([
        HumanMessage(content=f"审查以下实现：\n{inputs['implementation']}")
    ])
    return {
        "design": inputs["design"],
        "implementation": inputs["implementation"],
        "review": response.content
    }

# 构建链
chain = RunnableSequence(
    design_step,
    implement_step,
    review_step
)

# 执行
result = chain.invoke({"requirement": "创建一个用户认证系统"})

print("=== 设计 ===")
print(result["design"])
print("\n=== 实现 ===")
print(result["implementation"])
print("\n=== 审查 ===")
print(result["review"])
```yaml

---

## 步骤 3：LangGraph 代理工作流

### 构建多代理系统

创建 `langgraph_agent.py`：

```python
from typing import TypedDict, Annotated
from langgraph.graph import StateGraph, END
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage

# 状态定义
class AgentState(TypedDict):
    code: str
    architecture_review: str
    implementation: str
    security_review: str
    final_output: str

# 初始化 LLM
claude = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

codex = ChatOpenAI(
    model_name="codex",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

gemini = ChatOpenAI(
    model_name="gemini",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

# 节点函数
def architect_review(state: AgentState):
    """Claude 审查架构"""
    prompt = f"""审查此代码架构并提出改进建议：

{state['code']}

提供具体的设计模式和结构改进建议。"""

    response = claude.invoke([HumanMessage(content=prompt)])
    return {"architecture_review": response.content}

def implement(state: AgentState):
    """Codex 实现改进"""
    prompt = f"""基于以下架构审查，实现改进版本：

架构审查：
{state['architecture_review']}

原始代码：
{state['code']}

提供完整的改进实现。"""

    response = codex.invoke([HumanMessage(content=prompt)])
    return {"implementation": response.content}

def security_check(state: AgentState):
    """Gemini 检查安全"""
    prompt = f"""对此代码执行安全审计：

{state['implementation']}

识别任何安全漏洞并建议修复。"""

    response = gemini.invoke([HumanMessage(content=prompt)])
    return {"security_review": response.content}

def finalize(state: AgentState):
    """Claude 综合最终输出"""
    prompt = f"""基于以下信息综合最终解决方案：

架构审查：
{state['architecture_review']}

实现：
{state['implementation']}

安全审查：
{state['security_review']}

提供一个完整的、生产就绪的解决方案，整合所有反馈。"""

    response = claude.invoke([HumanMessage(content=prompt)])
    return {"final_output": response.content}

# 构建图
workflow = StateGraph(AgentState)

# 添加节点
workflow.add_node("architect", architect_review)
workflow.add_node("implement", implement)
workflow.add_node("security", security_check)
workflow.add_node("finalize", finalize)

# 定义边
workflow.set_entry_point("architect")
workflow.add_edge("architect", "implement")
workflow.add_edge("implement", "security")
workflow.add_edge("security", "finalize")
workflow.add_edge("finalize", END)

# 编译
app = workflow.compile()

# 使用示例代码执行
result = app.invoke({
    "code": """
def process_payment(card_number, amount):
    # 处理支付
    db.execute(f"INSERT INTO payments VALUES ('{card_number}', {amount})")
    return True
"""
})

print("=== 最终解决方案 ===")
print(result["final_output"])
```text

### 条件路由

创建 `conditional_routing.py`：

```python
from typing import TypedDict
from langgraph.graph import StateGraph, END
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage

class RouterState(TypedDict):
    task: str
    task_type: str
    result: str

# 初始化 LLM
claude = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

codex = ChatOpenAI(
    model_name="codex",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

gemini = ChatOpenAI(
    model_name="gemini",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

def classify_task(state: RouterState):
    """Claude 分类任务类型"""
    prompt = f"""将此任务分类为以下之一：architecture, implementation, security, research

任务：{state['task']}

仅回复分类。"""

    response = claude.invoke([HumanMessage(content=prompt)])
    task_type = response.content.strip().lower()

    # 规范化分类
    if "architect" in task_type:
        task_type = "architecture"
    elif "implement" in task_type or "code" in task_type:
        task_type = "implementation"
    elif "security" in task_type:
        task_type = "security"
    else:
        task_type = "research"

    return {"task_type": task_type}

def route_to_backend(state: RouterState):
    """基于任务类型路由到适当后端"""
    task = state["task"]
    task_type = state["task_type"]

    if task_type == "architecture":
        response = claude.invoke([HumanMessage(content=task)])
    elif task_type == "implementation":
        response = codex.invoke([HumanMessage(content=task)])
    elif task_type == "security":
        response = gemini.invoke([HumanMessage(content=task)])
    else:
        # 默认使用 Gemini 进行研究
        response = gemini.invoke([HumanMessage(content=task)])

    return {"result": response.content}

# 构建图
workflow = StateGraph(RouterState)
workflow.add_node("classify", classify_task)
workflow.add_node("execute", route_to_backend)

workflow.set_entry_point("classify")
workflow.add_edge("classify", "execute")
workflow.add_edge("execute", END)

app = workflow.compile()

# 用不同任务测试
tasks = [
    "为电商平台设计微服务架构",
    "用 Python 实现快速排序算法",
    "安全审计：检查 SQL 注入漏洞",
]

for task in tasks:
    result = app.invoke({"task": task})
    print(f"\n任务：{task}")
    print(f"类型：{result['task_type']}")
    print(f"结果：{result['result'][:200]}...")
```yaml

---

## 步骤 4：流式响应

### 实时流式

创建 `streaming_example.py`：

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
import sys

# 配置 LLM 使用流式
llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    streaming=True,
)

# 流式响应
messages = [HumanMessage(content="撰写一份关于 Python async/await 的综合指南")]

print("流式响应：")
for chunk in llm.stream(messages):
    # chunk.content 包含文本增量
    print(chunk.content, end="", flush=True)
    sys.stdout.flush()

print()  # 最终换行
```text

### 使用回调的流式

创建 `streaming_callbacks.py`：

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
from langchain.callbacks.streaming_stdout import StreamingStdOutCallbackHandler

# 配置流式回调
llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    streaming=True,
    callbacks=[StreamingStdOutCallbackHandler()],
)

# 这将自动流式传输到 stdout
messages = [HumanMessage(content="解释 SOLID 原则")]
response = llm.invoke(messages)
```yaml

---

## 步骤 5：错误处理模式

### 带指数退避的重试

创建 `error_handling.py`：

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
import time
import random

def invoke_with_retry(llm, messages, max_retries=3):
    """使用重试逻辑调用 LLM"""
    for attempt in range(max_retries):
        try:
            return llm.invoke(messages)
        except Exception as e:
            if attempt == max_retries - 1:
                raise

            # 指数退避带抖动
            wait_time = (2 ** attempt) + random.uniform(0, 1)
            print(f"尝试 {attempt + 1} 失败：{e}。{wait_time:.2f}秒后重试...")
            time.sleep(wait_time)

# 用法
llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

try:
    response = invoke_with_retry(
        llm,
        [HumanMessage(content="生成复杂分析")]
    )
    print(response.content)
except Exception as e:
    print(f"重试后失败：{e}")
```text

### 回退链

创建 `fallback_chain.py`：

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage

def fallback_invoke(prompt, backends=["claude", "codex", "gemini"]):
    """按顺序尝试后端直到一个成功"""
    for backend in backends:
        try:
            llm = ChatOpenAI(
                model_name=backend,
                openai_api_base="http://localhost:8080/openai/v1",
                openai_api_key="not-needed",
            )
            response = llm.invoke([HumanMessage(content=prompt)])
            print(f"后端 {backend} 成功")
            return response
        except Exception as e:
            print(f"后端 {backend} 失败：{e}")
            continue

    raise Exception("所有后端都失败")

# 用法
response = fallback_invoke("解释量子计算")
print(response.content)
```yaml

---

## 步骤 6：自定义回调处理器

### 跟踪 Token 使用

创建 `custom_callbacks.py`：

```python
from langchain_core.callbacks import BaseCallbackHandler
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
import time

class UsageCallbackHandler(BaseCallbackHandler):
    """自定义回调以跟踪使用统计"""

    def __init__(self):
        self.start_time = None
        self.token_usage = {"prompt": 0, "completion": 0}
        self.backend = None

    def on_llm_start(self, serialized, prompts, **kwargs):
        self.start_time = time.time()
        # 从模型名称提取后端
        if serialized and "kwargs" in serialized:
            self.backend = serialized["kwargs"].get("model_name", "unknown")
        print(f"开始向 {self.backend} 发送请求...")

    def on_llm_end(self, response, **kwargs):
        duration = time.time() - self.start_time
        print(f"\n请求在 {duration:.2f}秒内完成")

        # 如果可用则提取 token 使用
        if hasattr(response, 'llm_output') and response.llm_output:
            token_usage = response.llm_output.get('token_usage', {})
            print(f"Token 使用：{token_usage}")

    def on_llm_error(self, error, **kwargs):
        print(f"发生错误：{error}")

# 用法
handler = UsageCallbackHandler()

llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    callbacks=[handler],
)

response = llm.invoke([HumanMessage(content="解释机器学习")])
print(f"\n响应：{response.content[:200]}...")
```yaml

---

## 最佳实践

### 1. 连接池

重用 LLM 实例以获得更好性能：

```python
# 好的：重用 LLM 实例
claude_llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
)

# 多次使用相同实例
for prompt in prompts:
    response = claude_llm.invoke([HumanMessage(content=prompt)])
```text

### 2. 超时配置

为您的用例设置适当的超时：

```python
llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    request_timeout=60,  # 60 秒
)
```text

### 3. 模型选择策略

基于任务特征选择后端：

```python
def get_llm_for_task(task_type: str):
    """基于任务类型获取适当的 LLM"""
    config = {
        "architecture": ("claude", 0.7),
        "implementation": ("codex", 0.5),
        "security": ("gemini", 0.7),
        "research": ("gemini", 0.8),
    }

    model, temp = config.get(task_type, ("claude", 0.7))

    return ChatOpenAI(
        model_name=model,
        openai_api_base="http://localhost:8080/openai/v1",
        openai_api_key="not-needed",
        temperature=temp,
    )
```yaml

---

## 故障排除

### 连接错误

如果您遇到连接错误：

```python
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# 对于自签名证书
llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    # 如果使用 HTTPS 和自签名证书
    # http_client=httpx.Client(verify=False),
)
```text

### 模型未找到

如果您遇到 "model not found" 错误：

```python
# 验证可用模型
import requests

response = requests.get("http://localhost:8080/openai/v1/models")
print(response.json())

# 使用响应中的确切模型名称
```text

### 超时问题

对于长时间运行的任务：

```python
llm = ChatOpenAI(
    model_name="claude",
    openai_api_base="http://localhost:8080/openai/v1",
    openai_api_key="not-needed",
    request_timeout=300,  # 5 分钟
    max_retries=3,
)
```text

---

## 后续步骤

- 了解[构建 AI Skills](building-ai-skills.zh.md)以进行 Claude Code 集成
- 探索[多后端代码审查](multi-backend-code-review.zh.md)以获取审查自动化
- 查看[CI/CD 集成](ci-cd-integration.zh.md)以进行生产部署
- 查看[架构概述](../concepts/architecture.zh.md)以了解内部原理

---

## 总结

您已经学会如何：

1. 配置 LangChain ChatOpenAI 使用 clinvoker 的 OpenAI 兼容端点
2. 构建多后端链进行并行和顺序执行
3. 创建具有多个 AI 后端的 LangGraph 代理工作流
4. 为实时应用程序实现流式响应
5. 使用重试逻辑和回退链处理错误
6. 创建自定义回调处理器进行监控和日志记录

通过将 clinvoker 与 LangChain 集成，您可以利用 LangChain 生态系统的全部功能，同时为每个任务路由到最合适的 AI 后端。
