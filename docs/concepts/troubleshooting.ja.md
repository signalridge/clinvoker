---
title: トラブルシューティング
description: clinvoker でよくある問題、診断方法、解決策。
---

# トラブルシューティング

このガイドでは、clinvoker の利用中に遭遇しやすい問題と、診断方法および解決策をまとめます。参照しやすいようにカテゴリ別に整理しています。

## 診断方法

個別の問題に入る前に、まずは一般的な診断アプローチを紹介します。

### デバッグモードを有効にする

```bash
# verbose は詳細な実行情報を表示します
clinvk --verbose "your prompt"

# デバッグモードはバックエンドコマンド出力も含めます
CLINVK_DEBUG=1 clinvk "your prompt"
```

### システム状態を確認する

```bash
# バージョン/ビルド情報
clinvk version

# 利用可能なバックエンド
clinvk config show

# 設定ファイルの検証
clinvk config validate
```

### ドライラン

実行せずに、どのコマンドが走るかを確認できます。

```bash
clinvk --dry-run "your prompt"
```

## バックエンドが利用できない

### 症状

- エラー: `Backend 'claude' not found in PATH`
- エラー: `No backends available`
- 終了コード: 126

### 原因

1. バックエンド CLI がインストールされていない
2. バックエンド CLI が PATH にない
3. バックエンド CLI の実行権限が不正
4. バックエンド CLI のバージョンが想定と異なる

### 解決策

#### 1. インストール確認

```bash
# バックエンド CLI がインストールされているか
which claude codex gemini

# バージョン確認
claude --version
codex --version
gemini --version
```

#### 2. PATH へ追加

```bash
# シェルプロファイル（~/.bashrc, ~/.zshrc など）へ追加
export PATH="$PATH:/usr/local/bin"

# プロファイルを再読み込み
source ~/.bashrc  # または ~/.zshrc
```

#### 3. 実行権限を確認

```bash
# 実行権限を確認
ls -la $(which claude)

# 必要なら修正
chmod +x /path/to/claude
```

#### 4. clinvk の検出状況を確認

```bash
# clinvk が検出できるバックエンドを確認
clinvk config show | grep -A2 "available"
```

## セッション関連の問題

### セッション破損

**症状:**

- エラー: `Failed to load session: invalid JSON`
- セッションファイルが空/途中で切れている
- 以前は再開できていたセッションが再開できない

**原因:**

1. セッション書き込み中にプロセスが kill された
2. 書き込み中にディスクが満杯になった
3. ロックなしでの同時変更が起きた

**解決策:**

```bash
# セッションディレクトリを確認
ls -la ~/.clinvk/sessions/

# セッション内容を確認（JSON が妥当であるべき）
cat ~/.clinvk/sessions/<session-id>.json | python -m json.tool

# 破損している場合は削除
clinvk sessions delete <session-id>

# もしくは全削除
rm -rf ~/.clinvk/sessions/*
```

### セッションロックの問題

**症状:**

- エラー: `Failed to acquire store lock: timeout`
- 操作がいつまでも終わらない
- 「Resource temporarily unavailable」エラー

**原因:**

1. 別の clinvk プロセスがロックを保持している
2. 以前のプロセスがクラッシュしてロックを解放しなかった
3. 古いロックファイルが残っている

**解決策:**

```bash
# 動作中の clinvk プロセスを確認
ps aux | grep clinvk

# 必要なら停止
kill -9 <pid>

# 古いロックファイルを削除（注意）
rm -f ~/.clinvk/.store.lock

# ロックファイルの権限を確認
ls -la ~/.clinvk/
```

### セッションが見つからない

**症状:**

- エラー: `Session 'abc123' not found`
- 以前のセッションが再開できない

**解決策:**

```bash
# セッション一覧
clinvk sessions list

# セッションディレクトリ
ls -la ~/.clinvk/sessions/

# 部分 ID で検索
find ~/.clinvk/sessions/ -name "*abc*"
```

## API サーバーの問題

### 起動できない

**症状:**

- エラー: `Address already in use`
- ポートバインド時に `Permission denied`
- サーバーが即終了する

**解決策:**

**ポートが使用中:**

```bash
# ポートを使用しているプロセスを探す
lsof -i :8080
# または
netstat -tlnp | grep 8080

# 別ポートを使う
clinvk serve --port 3000

# もしくは既存プロセスを停止
kill -9 <pid>
```

**権限不足（低いポート）:**

```bash
# 1024 より大きいポートを使う（root 不要）
clinvk serve --port 8080

# sudo で起動（推奨しません）
sudo clinvk serve --port 80
```

### 認証の問題

**症状:**

- エラー: `Unauthorized`（401）
- エラー: `Invalid API key`

**解決策:**

```bash
# API キーが設定されているか確認
clinvk config show | grep api_keys

# リクエストにキーが入っているか確認
curl -H "Authorization: Bearer YOUR_KEY" http://localhost:8080/api/v1/health

# 環境変数を確認
echo $CLINVK_API_KEY

# 認証なしで試す（キーが未設定の場合）
curl http://localhost:8080/health
```

### レート制限の問題

**症状:**

- エラー: `Rate limit exceeded`（429）
- 一定量のリクエスト後に拒否される

**解決策:**

```bash
# レート制限設定を確認
clinvk config show | grep rate_limit

# 設定を調整
clinvk config set server.rate_limit_rps 100

# 並列実行なら並行度を下げる
clinvk parallel --max-parallel 2 --file tasks.json
```

## パフォーマンスの問題

### 応答が遅い

**症状:**

- 想定より時間がかかる
- タイムアウトが発生する

**原因:**

1. バックエンドのレート制限
2. コンテキスト/プロンプトが大きい
3. バックエンド API へのネットワーク遅延
4. システムリソース不足

**解決策:**

```bash
# バックエンド状態を確認
curl http://localhost:8080/api/v1/backends

# リソース監視
top
iostat -x 1

# タイムアウトを増やす
clinvk config set timeout 300

# より高速なモデルを使う（モデル名はバックエンド CLI にそのまま渡されます）
clinvk -b claude -m haiku "prompt"
clinvk -b gemini -m gemini-2.5-flash "prompt"
```

### メモリ使用量が大きい

**症状:**

- clinvk が大量のメモリを使う
- システムが不安定になる

**解決策:**

```bash
# メモリ使用量を確認
ps aux | grep clinvk

# 並列実行の制限
clinvk parallel --max-parallel 2 --file tasks.json

# 古いセッションの削除
clinvk sessions clean --older-than 7d

# ステートレスモード（セッション保存なし）
clinvk --ephemeral "prompt"
```

### ディスク容量の問題

**症状:**

- エラー: `No space left on device`
- セッション操作が失敗する

**解決策:**

```bash
# ディスク容量
df -h

# セッション領域のサイズ
du -sh ~/.clinvk/sessions/

# 古いセッションを削除
clinvk sessions clean --older-than 30d

# 手動で削除
rm -rf ~/.clinvk/sessions/*.json
```

## 設定の問題

### 設定ファイルが読み込まれない

**症状:**

- 設定が反映されない
- デフォルト値のまま動く

**解決策:**

```bash
# 設定ファイルの場所
ls -la ~/.clinvk/config.yaml

# YAML 構文チェック（要 PyYAML）
cat ~/.clinvk/config.yaml | python -c "import yaml,sys; yaml.safe_load(sys.stdin)"

# ファイル権限
chmod 600 ~/.clinvk/config.yaml

# 有効な設定
clinvk config show

# 環境変数の上書きを確認
echo $CLINVK_BACKEND
echo $CLINVK_TIMEOUT
```

### 環境変数が反映されない

**症状:**

- 環境変数が無視される
- CLI フラグは効くが env が効かない

**解決策:**

```bash
# export 済みか確認
export CLINVK_BACKEND=codex

# 現在の環境変数
env | grep CLINVK

# 優先順位: CLI フラグ > 環境変数 > 設定ファイル
```

## プラットフォーム固有の問題

### macOS

**Gatekeeper にブロックされる:**

```bash
# quarantine 属性を削除
xattr -d com.apple.quarantine /path/to/clinvk
```

**Notarization の問題:**

```bash
# ダウンロードしたバイナリが実行できない場合
sudo spctl --master-disable  # 一時的に無効化（注意）
# 初回実行後に戻す
sudo spctl --master-enable
```

### Windows

**PATH の問題:**

```powershell
# PowerShell で PATH を追加
$env:Path += ";C:\\path\\to\\clinvk"

# または恒久的に追加（ユーザー環境変数）
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\\path\\to\\clinvk", "User")
```

**アンチウイルスの誤検知:**

一部のアンチウイルスが clinvk を誤検知する場合があります。必要なら除外設定を追加してください。

### Linux

**Permission denied:**

```bash
chmod +x /path/to/clinvk
```

**SELinux の問題:**

```bash
# SELinux 状態確認
getenforce

# enforcing の場合は denial を確認
ausearch -m avc -ts recent

# 必要ならポリシーモジュール作成
audit2allow -a -M clinvk
semodule -i clinvk.pp
```

## デバッグモードの使い方

### ログを増やす

```bash
# デバッグ環境変数
export CLINVK_DEBUG=1

# verbose
clinvk --verbose "prompt"

# サーバーモード
CLINVK_DEBUG=1 clinvk serve
```

### ログの出力先

**CLI モード:**

- デフォルトでは stderr
- ファイルへリダイレクト: `clinvk "prompt" 2> debug.log`

**サーバーモード:**

- stdout/stderr へ出力
- systemd/journald を使う: `journalctl -u clinvk`
- もしくはリダイレクト: `clinvk serve > server.log 2>&1`

### よくあるログパターン

**バックエンドコマンド:**

```text
[DEBUG] Executing: claude --print --model sonnet "prompt"
[DEBUG] Working directory: /home/user/project
[DEBUG] Exit code: 0
```

**セッション操作:**

```text
[DEBUG] Loading session: abc123
[DEBUG] Acquiring file lock
[DEBUG] Session saved successfully
```

**API リクエスト:**

```text
[DEBUG] POST /api/v1/prompt
[DEBUG] Request body: {...}
[DEBUG] Response: 200 OK
```

## ヘルプを得る

上記を試しても解決しない場合:

1. **ドキュメントを確認**
   - [FAQ](faq.md)
   - [ガイド](../guides/index.md)
   - [リファレンス](../reference/index.md)

2. **既存 Issue を検索**
   - [GitHub Issues](https://github.com/signalridge/clinvoker/issues)
   - エラーメッセージで検索

3. **新しい Issue を作成**
   次を含めてください。
   - clinvk バージョン（`clinvk version`）
   - OS とバージョン
   - バックエンドのバージョン（`claude --version` など）
   - 完全なエラーメッセージ
   - 再現手順
   - デバッグログ（可能なら）

4. **コミュニティサポート**
   - GitHub Discussion を開始
   - 類似トピックがないか確認

## 関連ドキュメント

- [FAQ](faq.md) - よくある質問
- [設計判断](design-decisions.md) - アーキテクチャ説明
- [コントリビューティング](contributing.md) - 開発セットアップ
