---
title: バックエンドシステム
description: clinvoker のバックエンド抽象化レイヤ、レジストリパターン、統一インターフェースの詳細。
---

# バックエンドシステム

このドキュメントでは、clinvoker のバックエンド抽象化レイヤを深掘りします。複数の AI CLI ツールを共通インターフェースで統一する方法、レジストリパターンの実装、スレッドセーフな設計、そして新しいバックエンドを拡張する方法を説明します。

## バックエンドインターフェース設計

`Backend` インターフェース（`internal/backend/backend.go:16-46`）は、clinvoker が複数の AI CLI ツールをシームレスに扱うための中核となる抽象です。

```go
type Backend interface {
    Name() string
    IsAvailable() bool
    BuildCommand(prompt string, opts *Options) *exec.Cmd
    ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd
    BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd
    ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd
    ParseOutput(rawOutput string) string
    ParseJSONResponse(rawOutput string) (*UnifiedResponse, error)
    SeparateStderr() bool
}
```

### インターフェース設計の意図

このインターフェースは、AI とのやり取りのライフサイクルに沿って設計されています。

1. **発見（Discovery）**: `Name()` と `IsAvailable()` による識別と検出
2. **コマンド組み立て**: `BuildCommand*` が実行コマンドを生成
3. **セッション再開**: `ResumeCommand*` が既存会話を継続
4. **出力処理**: `ParseOutput()` と `ParseJSONResponse()` が応答を正規化
5. **stderr 取り扱い**: `SeparateStderr()` が stderr の分離戦略を決定

## レジストリパターン

バックエンドレジストリ（`internal/backend/registry.go`）は、スレッドセーフなレジストリパターンでバックエンドの登録とルックアップを管理します。

### レジストリの構造

```mermaid
flowchart TB
    subgraph Registry["レジストリ（internal/backend/registry.go:11-16）"]
        RWMU[sync.RWMutex]
        BACKENDS[map[string]Backend]
        CACHE[availabilityCache]
        TTL[30 秒 TTL]
    end

    subgraph Operations["レジストリ操作"]
        REGISTER[Register]
        UNREGISTER[Unregister]
        GET[Get]
        LIST[List]
        AVAILABLE[Available]
    end

    RWMU --> BACKENDS
    BACKENDS --> CACHE
    CACHE --> TTL

    REGISTER --> RWMU
    UNREGISTER --> RWMU
    GET --> RWMU
    LIST --> RWMU
    AVAILABLE --> CACHE
```

### スレッドセーフな設計

レジストリは `sync.RWMutex` を使って並行アクセスを制御します。

```go
// Read operations use RLock for concurrent reads
func (r *Registry) Get(name string) (Backend, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // ... lookup logic
}

// Write operations use Lock for exclusive access
func (r *Registry) Register(b Backend) {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ... registration logic
    delete(r.availabilityCache, b.Name()) // Invalidate cache
}
```

この設計により次が可能になります。

- 複数の同時読み取り（例: ヘルスチェック、一覧表示）
- 排他書き込み（例: 登録、登録解除）
- 複数 goroutine からの安全な同時アクセス

### 可用性キャッシュ

レジストリは可用性チェックのために 30 秒 TTL のキャッシュを実装しています。

```go
type cachedAvailability struct {
    available bool
    checkedAt time.Time
}

func (r *Registry) isAvailableCachedLocked(b Backend) bool {
    name := b.Name()
    if cached, ok := r.availabilityCache[name]; ok &&
       time.Since(cached.checkedAt) < r.availabilityCacheTTL {
        return cached.available
    }

    available := b.IsAvailable()
    r.availabilityCache[name] = &cachedAvailability{
        available: available,
        checkedAt: time.Now(),
    }
    return available
}
```

**30 秒 TTL の理由:**

- **性能**: `exec.LookPath()` の頻繁な呼び出しを回避
- **鮮度**: インストール状況の変化を十分早く検出
- **バランス**: 正確性と性能のトレードオフ

## バックエンド実装

### Claude バックエンド

```go
type Claude struct{}

func (c *Claude) Name() string { return "claude" }

func (c *Claude) IsAvailable() bool {
    _, err := exec.LookPath("claude")
    return err == nil
}

func (c *Claude) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--print"}

    if opts != nil {
        if opts.Model != "" {
            args = append(args, "--model", opts.Model)
        }
        // ... additional options
    }

    args = append(args, prompt)
    cmd := exec.Command("claude", args...)

    if opts != nil && opts.WorkDir != "" {
        cmd.Dir = opts.WorkDir
    }
    return cmd
}
```

### Codex バックエンド

```go
type Codex struct{}

func (c *Codex) Name() string { return "codex" }

func (c *Codex) IsAvailable() bool {
    _, err := exec.LookPath("codex")
    return err == nil
}

func (c *Codex) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    return exec.Command("codex", args...)
}
```

### Gemini バックエンド

```go
type Gemini struct{}

func (g *Gemini) Name() string { return "gemini" }

func (g *Gemini) IsAvailable() bool {
    _, err := exec.LookPath("gemini")
    return err == nil
}

func (g *Gemini) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--output-format", "json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    return exec.Command("gemini", args...)
}
```

## 統一オプション（UnifiedOptions）の扱い

`UnifiedOptions` 構造体（`internal/backend/unified.go:174-219`）は、バックエンドに依存しない形で AI CLI コマンドを設定するための手段です。

```go
type UnifiedOptions struct {
    WorkDir       string
    Model         string
    ApprovalMode  ApprovalMode
    SandboxMode   SandboxMode
    OutputFormat  OutputFormat
    AllowedTools  string
    AllowedDirs   []string
    Interactive   bool
    Verbose       bool
    DryRun        bool
    MaxTokens     int
    MaxTurns      int
    SystemPrompt  string
    ExtraFlags    []string
    Ephemeral     bool
}
```

### フラグマッピングのアーキテクチャ

```mermaid
flowchart TB
    subgraph Unified["UnifiedOptions"]
        MODEL[Model]
        APPROVAL[ApprovalMode]
        SANDBOX[SandboxMode]
        OUTPUT[OutputFormat]
    end

    subgraph Mapper["フラグマッパ（internal/backend/unified.go:273-568）"]
        MAP_MODEL[mapModel()]
        MAP_APPROVAL[mapApprovalMode()]
        MAP_SANDBOX[mapSandboxMode()]
        MAP_OUTPUT[mapOutputFormat()]
    end

    subgraph Backends["バックエンド固有フラグ"]
        CLAUDE[Claude フラグ]
        CODEX[Codex フラグ]
        GEMINI[Gemini フラグ]
    end

    MODEL --> MAP_MODEL
    APPROVAL --> MAP_APPROVAL
    SANDBOX --> MAP_SANDBOX
    OUTPUT --> MAP_OUTPUT

    MAP_MODEL --> CLAUDE
    MAP_MODEL --> CODEX
    MAP_MODEL --> GEMINI
    MAP_APPROVAL --> CLAUDE
    MAP_APPROVAL --> CODEX
    MAP_APPROVAL --> GEMINI
    MAP_SANDBOX --> CLAUDE
    MAP_SANDBOX --> CODEX
    MAP_SANDBOX --> GEMINI
    MAP_OUTPUT --> CLAUDE
    MAP_OUTPUT --> CODEX
    MAP_OUTPUT --> GEMINI
```

### モデル名のパススルー

モデル名は変換せず、そのままバックエンド CLI に渡されます。モデルの解決（別名/エイリアス）は各 CLI が担当します。

```bash
# モデル名はバックエンド CLI にそのまま渡されます
clinvk -b claude -m sonnet "task"           # Claude CLI が "sonnet" を解決
clinvk -b gemini -m gemini-2.5-flash "task" # Gemini CLI はモデル名をそのまま利用
clinvk -b codex -m o3 "task"                # Codex CLI はモデル名をそのまま利用
```

### 承認モードのマッピング

承認モードは、バックエンドがユーザー確認をどう扱うかを制御します。

```go
func (m *flagMapper) mapApprovalMode(mode ApprovalMode) []string {
    switch m.backend {
    case "claude":
        switch mode {
        case ApprovalAuto:
            return []string{"--permission-mode", "acceptEdits"}
        case ApprovalNone:
            return []string{"--permission-mode", "dontAsk"}
        case ApprovalAlways:
            return []string{"--permission-mode", "default"}
        }
    case "codex":
        switch mode {
        case ApprovalAuto:
            return []string{"--ask-for-approval", "on-request"}
        case ApprovalNone:
            return []string{"--ask-for-approval", "never"}
        case ApprovalAlways:
            return []string{"--ask-for-approval", "untrusted"}
        }
    // ...
    }
    return nil
}
```

## 出力の解析と正規化

各バックエンドは、ネイティブ出力を統一形式へ解析します。

### JSON レスポンスの解析

```go
func (c *Claude) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
    // First try to parse as error response
    var errResp claudeErrorResponse
    if err := json.Unmarshal([]byte(rawOutput), &errResp); err == nil {
        if errResp.Error != "" {
            return &UnifiedResponse{
                SessionID: errResp.SessionID,
                Error:     errResp.Error,
            }, nil
        }
    }

    var resp claudeJSONResponse
    if err := json.Unmarshal([]byte(rawOutput), &resp); err != nil {
        return nil, err
    }

    return &UnifiedResponse{
        Content:    resp.Result,
        SessionID:  resp.SessionID,
        DurationMs: resp.DurationMs,
        Usage: &TokenUsage{
            InputTokens:  resp.Usage.InputTokens,
            OutputTokens: resp.Usage.OutputTokens,
        },
    }, nil
}
```

### 統一レスポンス構造

```go
type UnifiedResponse struct {
    Content    string
    SessionID  string
    Model      string
    DurationMs int64
    Usage      *TokenUsage
    Error      string
    Raw        map[string]any
}
```

## 新しいバックエンドを追加する

clinvoker に新しい AI CLI バックエンドを追加する手順:

### ステップ 1: 実装ファイルを作成

`internal/backend/newbackend.go` を作成します。

```go
package backend

import "os/exec"

type NewBackend struct{}

func (n *NewBackend) Name() string {
    return "newbackend"
}

func (n *NewBackend) IsAvailable() bool {
    _, err := exec.LookPath("newbackend-cli")
    return err == nil
}

func (n *NewBackend) BuildCommand(prompt string, opts *Options) *exec.Cmd {
    args := []string{"--output", "json"}

    if opts != nil && opts.Model != "" {
        args = append(args, "--model", opts.Model)
    }

    args = append(args, prompt)
    cmd := exec.Command("newbackend-cli", args...)

    if opts != nil && opts.WorkDir != "" {
        cmd.Dir = opts.WorkDir
    }
    return cmd
}

func (n *NewBackend) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
    args := []string{"--resume", sessionID, "--output", "json"}

    if prompt != "" {
        args = append(args, prompt)
    }

    return exec.Command("newbackend-cli", args...)
}

func (n *NewBackend) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
    return n.BuildCommand(prompt, MapFromUnified(n.Name(), opts))
}

func (n *NewBackend) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
    return n.ResumeCommand(sessionID, prompt, MapFromUnified(n.Name(), opts))
}

func (n *NewBackend) ParseOutput(rawOutput string) string {
    return rawOutput
}

func (n *NewBackend) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
    var resp struct {
        Content   string `json:"content"`
        SessionID string `json:"session_id"`
        Usage     struct {
            Input  int `json:"input_tokens"`
            Output int `json:"output_tokens"`
        } `json:"usage"`
    }

    if err := json.Unmarshal([]byte(rawOutput), &resp); err != nil {
        return nil, err
    }

    return &UnifiedResponse{
        Content:   resp.Content,
        SessionID: resp.SessionID,
        Usage: &TokenUsage{
            InputTokens:  resp.Usage.Input,
            OutputTokens: resp.Usage.Output,
        },
    }, nil
}

func (n *NewBackend) SeparateStderr() bool {
    return false
}
```

### ステップ 2: レジストリへ登録

`internal/backend/registry.go` に追加します。

```go
func init() {
    globalRegistry.Register(&Claude{})
    globalRegistry.Register(&Codex{})
    globalRegistry.Register(&Gemini{})
    globalRegistry.Register(&NewBackend{}) // Add this line
}
```

### ステップ 3: 許可フラグを追加

`internal/backend/unified.go:10-27` の allowlist を更新します。

```go
var allowedFlagPatterns = map[string][]string{
    "newbackend": {
        "--model", "--output-format", "--json",
        "--resume", "--sandbox",
    },
    // ...
}
```

## ベストプラクティス

### コマンド組み立て

- 実行前にパスを必ず検証する
- 引数のエスケープを適切に行う（Go の `exec.Command` が多くを担う）
- 対話/バッチの両方をサポートする
- 非対話出力に `--print` 相当があれば利用する

### 出力解析

- 部分的/不正な JSON を適切に扱う
- 制御文字や ANSI コードを除去する
- デバッグ用にエラーメッセージを保持する
- 可能なら構造化エラーを返す

### エラーハンドリング

- バックエンドエラーとシステムエラーを区別する
- 明確で実行可能なエラーメッセージにする
- エラー出力にトラブルシューティングのヒントを含める

## 関連ドキュメント

- [アーキテクチャ概要](architecture.md) - システム全体像
- [セッションシステム](session-system.md) - セッション永続化
- [API 設計](api-design.md) - REST API アーキテクチャ
