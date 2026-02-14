# 環境変数

現在の `clinvk` が実際に読み取る環境変数のリファレンスです。

## 対象範囲

このページには、コード上で有効な変数のみを記載します。
記載のない変数は動作保証しません。

## サポートされる環境変数

### コア実行設定

| 変数 | 説明 | 例 |
|---|---|---|
| `CLINVK_BACKEND` | 既定のバックエンド | `claude`, `codex`, `gemini` |
| `CLINVK_CLAUDE_MODEL` | Claude の既定モデル | `claude-opus-4-5-20251101` |
| `CLINVK_CODEX_MODEL` | Codex の既定モデル | `o3`, `o3-mini` |
| `CLINVK_GEMINI_MODEL` | Gemini の既定モデル | `gemini-2.5-pro` |

### MCP サーバー

| 変数 | 説明 | 例 |
|---|---|---|
| `CLINVK_MCP_TRANSPORT` | MCP トランスポート種別 | `stdio`, `http` |
| `CLINVK_MCP_HOST` | MCP HTTP のバインド先ホスト | `127.0.0.1`, `0.0.0.0` |
| `CLINVK_MCP_PORT` | MCP HTTP のポート | `8081` |
| `CLINVK_MCP_HTTP_PATH` | MCP エンドポイントのパス | `/mcp` |
| `CLINVK_MCP_EXPOSE_HEALTH` | MCP モードで `/health` を公開するか | `true`, `false` |

### HTTP API 認証

| 変数 | 説明 | 例 |
|---|---|---|
| `CLINVK_API_KEYS` | `serve`/`mcp` 用 API キー（カンマ区切り） | `key1,key2,key3` |
| `CLINVK_API_KEYS_GOPASS_PATH` | API キーを取得する gopass パス | `myproject/clinvk/api-keys` |

### バックエンド提供元の API キー

| 変数 | バックエンド | 説明 |
|---|---|---|
| `ANTHROPIC_API_KEY` | Claude | 提供元 API キー |
| `OPENAI_API_KEY` | Codex | 提供元 API キー |
| `GOOGLE_API_KEY` | Gemini | 提供元 API キー |

## よくある誤解（非サポート）

次の変数は、現行 `clinvk` では正式な実行時設定としてサポートされません。

- `CLINVK_TIMEOUT`
- `CLINVK_DEBUG`
- `CLINVK_SERVER_PORT`
- `CLINVK_HOME`
- `CLINVK_CONFIG`

代替手段:

- タイムアウト: 設定ファイルの `unified_flags.command_timeout_secs`
- デバッグ確認: `--dry-run` / `--output-format json` / stderr ログ
- 設定ファイル切替: `--config /path/to/config.yaml`

## 例

### バックエンドとモデル

```bash
export CLINVK_BACKEND=codex
export CLINVK_CODEX_MODEL=o3-mini
clinvk "review this patch"
```

### 環境変数で MCP を起動

```bash
export CLINVK_MCP_TRANSPORT=http
export CLINVK_MCP_HOST=0.0.0.0
export CLINVK_MCP_PORT=8081
export CLINVK_MCP_HTTP_PATH=/mcp
export CLINVK_MCP_EXPOSE_HEALTH=true

clinvk mcp
```

### HTTP/MCP の API キー認証

```bash
export CLINVK_API_KEYS="prod-key-1,prod-key-2"
clinvk serve --port 8080
```

## 優先順位

サポート対象のキーについて、優先順位は次の通りです。

1. CLI フラグ
2. 環境変数
3. 設定ファイル
4. デフォルト値

## トラブルシューティング

```bash
# export 済みの CLINVK 変数を確認
env | grep '^CLINVK_'

# 提供元 API キーの有無を確認
echo "${OPENAI_API_KEY:+set}"

# 実行コマンドを確認
clinvk --dry-run "check behavior"
```

## 関連ドキュメント

- [設定リファレンス](configuration.md)
- [CLI config コマンド](cli/config.md)
- [CLI mcp コマンド](cli/mcp.md)
