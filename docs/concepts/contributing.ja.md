---
title: コントリビューティングガイド
description: clinvoker の開発環境セットアップ、コーディング規約、貢献プロセス。
---

# コントリビューティングガイド

clinvoker への貢献に興味を持っていただきありがとうございます。このガイドでは、開発環境セットアップ、プロジェクト構造、コーディング規約、テスト要件、貢献プロセスを説明します。

## 開発環境セットアップ

### 前提

- Go 1.24 以降
- Git
- （任意）再現性のある環境のための Nix
- （任意）Lint 用の golangci-lint
- （任意）Git フック用の pre-commit

### Fork と Clone

```bash
# GitHub 上でリポジトリを fork してから:
git clone https://github.com/YOUR_USERNAME/clinvoker.git
cd clinvoker
git remote add upstream https://github.com/signalridge/clinvoker.git
```

### Nix を使う（推奨）

```bash
nix develop
```

必要なツール一式を再現可能な環境で提供します。

### 手動セットアップ

```bash
go mod download
go build ./cmd/clinvk
./clinvk version
```

### pre-commit フック

```bash
pre-commit install
pre-commit install --hook-type commit-msg
```

## プロジェクト構造

```text
clinvoker/
├── cmd/
│   └── clinvk/           # メインアプリのエントリポイント
│       └── main.go
├── internal/
│   ├── app/              # CLI コマンド実装
│   ├── backend/          # バックエンド抽象化レイヤ
│   ├── config/           # 設定管理
│   ├── executor/         # コマンド実行
│   ├── output/           # 出力整形
│   ├── server/           # HTTP API サーバー
│   ├── session/          # セッション管理
│   ├── auth/             # API キー管理
│   ├── metrics/          # Prometheus メトリクス
│   └── resilience/       # サーキットブレーカ
├── docs/                 # ドキュメント
├── scripts/              # ビルド/ユーティリティスクリプト
└── test/                 # 統合テスト
```

### パッケージの責務

| パッケージ | 目的 | 主なファイル |
|---------|---------|-----------|
| `app/` | Cobra を使った CLI コマンド | `app.go`, `cmd_*.go` |
| `backend/` | バックエンド抽象化 | `backend.go`, `registry.go`, `claude.go`, `codex.go`, `gemini.go` |
| `config/` | Viper ベースの設定 | `config.go`, `validate.go` |
| `executor/` | サブプロセス実行 | `executor.go`, `signal.go` |
| `server/` | HTTP サーバー | `server.go`, `routes.go`, `handlers/`, `middleware/` |
| `session/` | セッション永続化 | `session.go`, `store.go`, `filelock.go` |

## コーディング規約

### Go のガイドライン

[Effective Go](https://golang.org/doc/effective_go.html) と [Google Go Style Guide](https://google.github.io/styleguide/go/) に従ってください。

1. **フォーマット**: `gofmt` または `goimports` を使用
2. **Lint**: コミット前に `golangci-lint run` を実行
3. **命名**: 説明的で Go らしい名前にする
4. **コメント**: export する型/関数には必ずドキュメントコメントを書く
5. **エラーハンドリング**: コンテキストを付けて wrap、naked return を避ける

### スタイル例

```go
// Good: Clear function name, proper documentation, error handling
// ExecuteCommand runs the given command and returns the output.
func ExecuteCommand(ctx context.Context, cfg *Config, cmd *exec.Cmd) (*Result, error) {
    if cfg == nil {
        return nil, fmt.Errorf("config is required")
    }

    // Implementation
    result, err := runWithTimeout(ctx, cmd, cfg.Timeout)
    if err != nil {
        return nil, fmt.Errorf("failed to execute command: %w", err)
    }

    return result, nil
}

// Bad: Unclear name, missing documentation, poor error handling
func exec(cfg *Config, c *exec.Cmd) (*Result, error) {
    res, _ := runWithTimeout(context.Background(), c, cfg.Timeout)
    return res, nil
}
```

### エラーハンドリング

プロジェクトの errors パッケージを使って、一貫したエラー処理を行います。

```go
import apperrors "github.com/signalridge/clinvoker/internal/errors"

// Create error with context
return apperrors.BackendError("claude", err)

// Check error type
if apperrors.IsCode(err, apperrors.ErrCodeBackendUnavailable) {
    // Handle specific error
}
```

### テスト規約

新しいコードにはテストが必要です。次の方針に従ってください。

1. **ファイル名**: ソースと同階層に `*_test.go`
2. **テーブル駆動テスト**: 複数ケースに有効
3. **並列テスト**: 独立ケースは `t.Parallel()` を使う
4. **カバレッジ**: 新規コードは 80% 以上を目安
5. **モック**: テスト容易性のためにインターフェースを活用

```go
func TestExecuteCommand(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        cfg     *Config
        cmd     *exec.Cmd
        wantErr bool
    }{
        {
            name:    "valid command",
            cfg:     &Config{Timeout: 30 * time.Second},
            cmd:     exec.Command("echo", "hello"),
            wantErr: false,
        },
        {
            name:    "nil config",
            cfg:     nil,
            cmd:     exec.Command("echo", "hello"),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            ctx := context.Background()
            _, err := ExecuteCommand(ctx, tt.cfg, tt.cmd)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## テスト要件

### テストの実行

```bash
# すべてのテスト
go test ./...

# race 検出
go test -race ./...

# カバレッジ
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt -o coverage.html

# short テストのみ
go test -short ./...

# 特定パッケージ
go test ./internal/backend/...
```

### 統合テスト

統合テストは実際のバックエンド CLI のインストールが必要です。

```bash
# 統合テストを実行
CLINVK_TEST_INTEGRATION=1 go test ./test/...

# 特定バックエンドで実行
CLINVK_TEST_BACKEND=claude go test ./test/...
```

### ベンチマーク

```bash
# ベンチマーク
go test -bench=. ./...

# メモリプロファイル付き
go test -bench=. -benchmem ./...
```

## ドキュメント要件

### コードコメント

export する型/関数は必ずドキュメントコメントを書いてください。

```go
// Backend represents an AI CLI backend.
type Backend interface {
    // Name returns the backend identifier.
    Name() string

    // IsAvailable checks if the backend CLI is installed.
    IsAvailable() bool
}

// NewStore creates a new session store at the given directory.
// The directory is created if it doesn't exist.
func NewStore(dir string) (*Store, error) {
    // ...
}
```

### ユーザー向けドキュメント

ユーザーに影響する変更を加えた場合、ドキュメントも更新してください。

1. **Concepts**: 設計変更がある場合はアーキテクチャドキュメントを更新
2. **Guides**: 新機能の How-to ガイドを追加/更新
3. **Reference**: API/CLI リファレンスを更新
4. **Changelog**: `CHANGELOG.md` にエントリ追加

### ドキュメントスタイル

- 明確で簡潔な表現にする
- コード例を含める
- 複雑な概念は図（Mermaid）を使う
- 英語/中国語/日本語の内容が大きく乖離しないようにする

## PR プロセス

### ブランチ命名

慣例的なプレフィックスを使ってください。

| Prefix | 用途 | 例 |
|--------|---------|---------|
| `feat/` | 新機能 | `feat/add-gemini-backend` |
| `fix/` | バグ修正 | `fix/session-locking` |
| `docs/` | ドキュメント | `docs/api-examples` |
| `refactor/` | リファクタ | `refactor/executor` |
| `test/` | テスト追加 | `test/backend-coverage` |
| `chore/` | 保守 | `chore/update-deps` |

### コミットメッセージ規約

[Conventional Commits](https://www.conventionalcommits.org/) に従ってください。

```text
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

例:

```text
feat(backend): add support for new AI provider

Add support for the new XYZ AI CLI tool. This includes:
- Backend implementation
- Flag mapping
- Documentation updates

fix(session): handle concurrent access correctly

Fix race condition in session store when multiple processes
attempt to update the same session simultaneously.

docs(readme): update installation instructions

test(backend): add unit tests for registry

chore(deps): update golang.org/x packages
```

### Pull Request テンプレート

```markdown
## Summary

Brief description of changes.

## Changes

- Change 1
- Change 2

## Test Plan

How changes were tested:
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing performed

## Related Issues

Fixes #123
Closes #456
```

### コードレビュー チェックリスト

PR を出す前に確認してください。

- [ ] ローカルで全テストが通る
- [ ] スタイルガイドに従っている
- [ ] ドキュメントを更新した
- [ ] コミットメッセージが規約に従っている
- [ ] 不要なファイルを含めていない
- [ ] `CHANGELOG.md` を更新した（必要な場合）

### レビュープロセス

1. PR には少なくとも 1 人の承認が必要です
2. マージ前に CI が通る必要があります
3. レビューコメントへ迅速に対応してください
4. PR は焦点を絞り、適切なサイズにしてください（目安: 500 行未満）
5. コンフリクトがあれば main へリベースしてください

## リリースプロセス

### バージョニング

clinvoker は [Semantic Versioning](https://semver.org/) に従います。

- `MAJOR`: 破壊的変更
- `MINOR`: 後方互換な新機能
- `PATCH`: 後方互換なバグ修正

### release-please による自動リリース

自動バージョニングと CHANGELOG 生成に [release-please](https://github.com/googleapis/release-please) を使います。

#### 仕組み

```text
feat: add new feature  →  push to main  →  Release PR created automatically
                                                  ↓
                                         merge Release PR
                                                  ↓
                                         tag created (v1.1.0)
                                                  ↓
                                         GoReleaser builds binaries
```

1. [Conventional Commits](#コミットメッセージ規約) で **コミット** する
2. **main へ push** - release-please が Release PR を作成/更新
3. **Release PR をマージ** - tag と GitHub Release が作られる
4. **GoReleaser が起動** - 各プラットフォーム向けバイナリをビルド

#### Conventional Commits とバージョン番号

| コミット種別 | バージョン増分 | CHANGELOG 区分 |
|-------------|--------------|-------------------|
| `feat:` | Minor（0.1.0 → 0.2.0） | Features |
| `fix:` | Patch（0.1.0 → 0.1.1） | Bug Fixes |
| `feat!:` または `BREAKING CHANGE:` | Major（0.1.0 → 1.0.0） | Breaking Changes |
| `perf:` | Patch | Performance |
| `refactor:` | Patch | Refactoring |
| `deps:` | Patch | Dependencies |
| `security:` | Patch | Security |
| `docs:`, `chore:`, `test:`, `ci:` | リリースなし | Hidden |

#### ワークフロー例

```bash
# 1. 機能ブランチ作成
git checkout -b feat/new-feature

# 2. 変更して conventional 形式でコミット
git commit -m "feat(backend): add support for new AI provider"

# 3. PR を作成して main へマージ
gh pr create && gh pr merge

# 4. release-please が Release PR を自動作成
#    Title: "chore: release v0.2.0"
#    Contains: Updated CHANGELOG.md and version bumps

# 5. Release PR をレビューしてマージ
#    → Tag v0.2.0 created
#    → GoReleaser builds and publishes

# 6. 完了。バイナリは次で利用可能:
#    - GitHub Releases
#    - Homebrew: brew install signalridge/tap/clinvk
#    - go install: go install github.com/signalridge/clinvoker/cmd/clinvk@latest
```

#### 手動リリース（緊急時のみ）

自動化が使えない緊急時の手順:

```bash
# 1. CHANGELOG.md を手動更新
# 2. tag を作成して push
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
# 3. GoReleaser ワークフローが自動で起動
```

## 質問がありますか？

- バグ/機能要望は issue を作成
- 質問は discussion を開始
- 新規作成前に既存 issue を確認

## 行動規範

参加することで、次に同意したものとみなされます。

1. 敬意を持ち、包摂的である
2. 建設的な批判を受け入れる
3. コミュニティにとって最善を優先する
4. 他者へ共感を持つ

## 関連ドキュメント

- [アーキテクチャ概要](architecture.md) - システムアーキテクチャ
- [設計判断](design-decisions.md) - アーキテクチャ判断（ADR）
- [トラブルシューティング](troubleshooting.md) - よくある問題
