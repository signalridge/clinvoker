# 環境変数

clinvk がサポートする環境変数の完全リファレンスです。

## 概要

環境変数を使うと、設定ファイルを編集せずに clinvk を手軽に設定できます。特に次の用途で便利です。

- CI/CD パイプライン
- Docker コンテナ
- 一時的な上書き
- direnv を使ったプロジェクト単位の設定

## 変数一覧

### コア変数

| 変数 | 必須 | 説明 | 例 |
|----------|----------|-------------|---------|
| `CLINVK_BACKEND` | いいえ | 既定で使用するバックエンド | `claude`, `codex`, `gemini` |
| `CLINVK_CLAUDE_MODEL` | いいえ | Claude バックエンドの既定モデル | `claude-opus-4-5-20251101` |
| `CLINVK_CODEX_MODEL` | いいえ | Codex バックエンドの既定モデル | `o3`, `o3-mini` |
| `CLINVK_GEMINI_MODEL` | いいえ | Gemini バックエンドの既定モデル | `gemini-2.5-pro` |

### サーバー関連

| 変数 | 必須 | 説明 | 例 |
|----------|----------|-------------|---------|
| `CLINVK_API_KEYS` | いいえ | HTTP サーバー認証用の API キー（カンマ区切り） | `key1,key2,key3` |
| `CLINVK_API_KEYS_GOPASS_PATH` | いいえ | API キー取得のための gopass パス | `myproject/api-keys` |
| `CLINVK_CONFIG` | いいえ | カスタム設定ファイルのパス | `/etc/clinvk/config.yaml` |
| `CLINVK_HOME` | いいえ | clinvk のデータディレクトリ（セッション/設定） | `~/.clinvk` |

### バックエンドの API キー

次の環境変数は、それぞれのバックエンド CLI にそのまま渡されます。

| 変数 | バックエンド | 説明 |
|----------|---------|-------------|
| `ANTHROPIC_API_KEY` | Claude | Anthropic API キー |
| `OPENAI_API_KEY` | Codex | OpenAI API キー |
| `GOOGLE_API_KEY` | Gemini | Google API キー |

## 利用例

### 既定バックエンドを設定する

```bash
export CLINVK_BACKEND=codex
clinvk "implement feature"  # Uses codex
```

### バックエンドごとにモデルを設定する

```bash
export CLINVK_CLAUDE_MODEL=claude-sonnet-4-20250514
export CLINVK_CODEX_MODEL=o3-mini

clinvk -b claude "complex task"  # Uses claude-sonnet
clinvk -b codex "quick task"     # Uses o3-mini
```

### 一時的に上書きする

1 回のコマンドだけに環境変数を適用します。

```bash
CLINVK_BACKEND=gemini clinvk "explain this"
```

### HTTP サーバー用の API キー

```bash
export CLINVK_API_KEYS="prod-key-1,prod-key-2,dev-key-1"
clinvk serve
```

クライアントは次のいずれかのヘッダーを含める必要があります。

```bash
# Option 1: X-Api-Key header
curl -H "X-Api-Key: prod-key-1" http://localhost:8080/api/v1/prompt \
  -d '{"backend":"claude","prompt":"hello"}'

# Option 2: Authorization header
curl -H "Authorization: Bearer prod-key-1" http://localhost:8080/api/v1/prompt \
  -d '{"backend":"claude","prompt":"hello"}'
```

## 優先順位

環境変数は設定階層の中で中程度の優先順位です。

1. **CLI フラグ**（最優先）
2. **環境変数**
3. **設定ファイル**
4. **デフォルト値**（最下位）

優先順位の例:

```bash
export CLINVK_BACKEND=codex
clinvk -b claude "prompt"  # Uses claude (CLI flag wins)
```

## シェル設定

### Bash

`~/.bashrc` または `~/.bash_profile` に追加します。

```bash
# clinvk configuration
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
export CLINVK_CODEX_MODEL=o3
```

### Zsh

`~/.zshrc` に追加します。

```zsh
# clinvk configuration
export CLINVK_BACKEND=claude
export CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
export CLINVK_CODEX_MODEL=o3
```

### Fish

`~/.config/fish/config.fish` に追加します。

```fish
# clinvk configuration
set -gx CLINVK_BACKEND claude
set -gx CLINVK_CLAUDE_MODEL claude-opus-4-5-20251101
set -gx CLINVK_CODEX_MODEL o3
```

## ディレクトリ単位の設定

[direnv](https://direnv.net/) を使うと、プロジェクト固有の設定ができます。

```bash
# .envrc in your project root
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3
```

ディレクトリに入ると、direnv が自動的にこれらの変数を読み込みます。

## CI/CD での利用

### GitHub Actions

```yaml
name: AI Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    env:
      CLINVK_BACKEND: codex
      CLINVK_CODEX_MODEL: o3
      OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
    steps:
      - uses: actions/checkout@v4
      - name: Review code
        run: clinvk "review this PR for security issues"
```

### GitLab CI

```yaml
ai-review:
  image: alpine/clinvk
  variables:
    CLINVK_BACKEND: claude
    CLINVK_CLAUDE_MODEL: claude-sonnet-4-20250514
    ANTHROPIC_API_KEY: $ANTHROPIC_API_KEY
  script:
    - clinvk "review the changes"
```

### Jenkins

```groovy
pipeline {
    agent any
    environment {
        CLINVK_BACKEND = 'gemini'
        CLINVK_GEMINI_MODEL = 'gemini-2.5-pro'
        GOOGLE_API_KEY = credentials('google-api-key')
    }
    stages {
        stage('AI Analysis') {
            steps {
                sh 'clinvk "analyze code quality"'
            }
        }
    }
}
```

## Docker での利用

### Dockerfile

```dockerfile
FROM alpine:latest
RUN apk add --no-cache clinvk

ENV CLINVK_BACKEND=claude
ENV CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101

ENTRYPOINT ["clinvk"]
```

### docker run

```bash
docker run -e CLINVK_BACKEND=codex -e OPENAI_API_KEY=$OPENAI_API_KEY clinvk "prompt"
```

### Docker Compose

```yaml
version: '3'
services:
  ai-task:
    image: clinvk
    environment:
      - CLINVK_BACKEND=claude
      - CLINVK_CLAUDE_MODEL=claude-opus-4-5-20251101
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    command: clinvk "analyze codebase"
```

## Kubernetes での利用

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: clinvk-config
data:
  CLINVK_BACKEND: "claude"
  CLINVK_CLAUDE_MODEL: "claude-opus-4-5-20251101"
---
apiVersion: v1
kind: Secret
metadata:
  name: clinvk-secrets
type: Opaque
stringData:
  ANTHROPIC_API_KEY: "your-api-key"
---
apiVersion: v1
kind: Pod
metadata:
  name: clinvk-job
spec:
  containers:
    - name: clinvk
      image: clinvk:latest
      envFrom:
        - configMapRef:
            name: clinvk-config
        - secretRef:
            name: clinvk-secrets
```

## トラブルシューティング

### 変数が反映されない

1. 変数が export されているか確認する（`export VAR=value` であり、`VAR=value` だけではない）
2. 変数名のタイプミスがないか確認する
3. CLI フラグが環境変数を上書きしていないか確認する
4. 設定ファイルが同じ値を設定していないか確認する

### 環境のデバッグ

```bash
# Show all CLINVK variables
env | grep CLINVK

# Show specific variable
echo $CLINVK_BACKEND

# Run with debug output
CLINVK_DEBUG=1 clinvk "prompt"
```

## 関連項目

- [設定リファレンス](configuration.md) - 設定ファイルの項目
- [`config` コマンド](cli/config.md) - CLI で設定を管理する
