---
title: Building AI Skills Tutorial
description: Create Claude Code Skills that leverage clinvoker to call other AI backends.
---

# Tutorial: Building AI Skills

Create Claude Code Skills that leverage clinvoker to call other AI backends for specialized tasks.

## What You'll Build

A Claude Code Skill that:

1. Analyzes code using Claude
2. Generates fixes using Codex
3. Reviews security using Gemini
4. Returns a comprehensive solution

## Prerequisites

- Claude Code installed
- clinvk installed
- At least two backends configured

## Step 1: Create the Skill

Create the skill directory and file:

```bash
mkdir -p .claude/skills
cat > .claude/skills/ai-team.json << 'EOF'
{
  "name": "ai-team",
  "description": "Collaborate with AI team members (Claude, Codex, Gemini) to solve complex problems",
  "prompt": "You are coordinating an AI development team with three specialists:\n\n1. **Claude** (System Architect) - Best for design, architecture, and code review\n2. **Codex** (Implementer) - Best for writing and refactoring code\n3. **Gemini** (Security & Docs) - Best for security analysis and documentation\n\nWhen the user asks you to implement a feature or solve a problem:\n\n1. First, use Claude to create an architecture/design\n2. Then, use Codex to implement the solution\n3. Finally, use Gemini to review for security issues\n4. Synthesize all feedback into a final recommendation\n\nUse the clinvk command to call different backends:\n- clinvk -b claude \"<prompt>\"\n- clinvk -b codex \"<prompt>\"\n- clinvk -b gemini \"<prompt>\"\n\nAlways explain which team member you're consulting and why.",
  "tools": ["clinvk"]
}
EOF
```

## Step 2: Test the Skill

Start Claude Code with the skill:

```bash
claude
```

In Claude Code, use the skill:

```text
/ai-team I need to implement a secure user authentication system in Go
```

Claude should:

1. Use Claude backend for architecture
2. Use Codex for implementation
3. Use Gemini for security review
4. Present a comprehensive solution

## Step 3: Create a Specialized Skill

Let's create a more focused skill for code review:

```bash
cat > .claude/skills/multi-review.json << 'EOF'
{
  "name": "multi-review",
  "description": "Get code reviews from multiple AI perspectives",
  "prompt": "You are a code review coordinator. When the user shares code:\n\n1. Create parallel review tasks for:\n   - Claude: Architecture and design patterns\n   - Codex: Implementation and performance\n   - Gemini: Security and best practices\n\n2. Use clinvk parallel execution:\n   ```\n   echo '{\"tasks\":[{\"backend\":\"claude\",\"prompt\":\"Review...\"},...]}' | clinvk parallel -f -\n   ```\n\n3. Synthesize the feedback into a structured review report\n\nFormat your response as:\n- Summary (key findings)\n- Detailed Feedback (by category)\n- Recommendations (prioritized)",
  "tools": ["clinvk", "jq"]
}
EOF
```

## Step 4: Advanced Skill with Session Management

Create a skill that maintains context across sessions:

```bash
cat > .claude/skills/long-term-project.json << 'EOF'
{
  "name": "long-term-project",
  "description": "Manage long-term projects with persistent sessions",
  "prompt": "You help manage long-term coding projects using clinvoker's session persistence.\n\nKey capabilities:\n1. Resume previous sessions: clinvk resume --last\n2. List active sessions: clinvk sessions list\n3. Continue work across multiple Claude Code sessions\n\nWhen starting work:\n1. Check for existing sessions: clinvk sessions list\n2. If found, ask user whether to resume\n3. If resuming: clinvk resume --last\n4. If new: proceed with new session\n\nBest practices:\n- Tag sessions appropriately\n- Summarize progress at end of each session\n- Use specific session IDs for important work",
  "tools": ["clinvk"]
}
EOF
```

## Step 5: Create a Reusable Skill Template

```bash
cat > .claude/skills/backend-router.json << 'EOF'
{
  "name": "backend-router",
  "description": "Route tasks to the most appropriate AI backend",
  "prompt": "You route user requests to the best AI backend:\n\n**Routing Rules:**\n- Architecture, design, complex reasoning → Claude\n- Code generation, implementation, refactoring → Codex\n- Security, research, documentation → Gemini\n\n**Automatic Routing:**\nYou can detect the task type and route automatically, or ask the user for preference.\n\n**Usage:**\n- Default: clinvk \"<prompt>\"\n- Specific: clinvk -b <backend> \"<prompt>\"\n\nAlways explain which backend you chose and why.",
  "tools": ["clinvk"]
}
EOF
```

## Verification

Test each skill:

```bash
# Test ai-team skill
claude
/ai-team Design a microservices architecture for an e-commerce platform

# Test multi-review skill
/multi-review
```text
paste your code here
```

## Best Practices for AI Skills

### 1. Clear Role Definition

Define what each backend does best:

```json
{
  "prompt": "Claude is best for: architecture, reasoning, review\nCodex is best for: implementation, code generation\nGemini is best for: security, research, trends"
}
```

### 2. Error Handling

Include fallback instructions:

```json
{
  "prompt": "If a backend is unavailable, try another:\n- If Claude fails, use Gemini for analysis\n- If Codex fails, use Claude for implementation"
}
```

### 3. Output Formatting

Request structured output:

```json
{
  "prompt": "Format results as:\n## Backend: <name>\n### Strengths\n...\n### Output\n..."
}
```

## Next Steps

- Learn about [Session Management](../how-to/session-management.md)
- See [AI Team Collaboration](../use-cases/ai-team-collaboration.md) for complex workflows
- Explore [LangChain Integration](langchain-integration.md) for programmatic workflows
