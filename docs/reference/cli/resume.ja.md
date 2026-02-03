# clinvk resume

以前のセッションを再開します。

## 用法

```bash
clinvk resume [session-id] [prompt] [flags]
```

## 説明

以前のセッションを再開して会話を継続します。セッションが再開できるのは、バックエンドのセッション ID が記録されていて、かつバックエンドが再開に対応している場合のみです。

## 引数

| 引数 | 説明 |
|----------|-------------|
| `session-id` | セッション ID（または接頭辞）。`--last` や対話式ピッカー利用時は省略可能 |
| `prompt` | 追記プロンプト（省略可能） |

## フラグ

| フラグ | 短縮 | 型 | デフォルト | 説明 |
|------|-------|------|---------|-------------|
| `--last` | | bool | `false` | 直近のセッションを再開（他フラグで絞り込み） |
| `--interactive` | `-i` | bool | `false` | 対話式のセッション選択を表示 |
| `--here` | | bool | `false` | 現在の作業ディレクトリでセッションを絞り込み |
| `--backend` | `-b` | string | | バックエンドでセッションを絞り込み |

## 例

### 直近のセッションを再開

再開可能な直近のセッションを再開します。

```bash
clinvk resume --last
```

### 追記プロンプトを付けて再開

再開して、すぐに追記プロンプトを送信します。

```bash
clinvk resume --last "continue from where we left off"
```

### 対話式ピッカー

対話式ピッカーでセッションを選択します。

```bash
clinvk resume --interactive
```

引数なしで `clinvk resume` を実行し、かつ `--last` も指定しない場合、デフォルトで対話式ピッカーが開きます。

### 現在のディレクトリのセッションを再開

セッションを現在のディレクトリ由来のものに絞り込みます。

```bash
clinvk resume --here
```

### バックエンドで絞り込む

特定バックエンドのセッションのみを対象にします。

```bash
clinvk resume --backend claude
```

### 特定セッションを再開

ID を指定して特定のセッションを再開します。

```bash
clinvk resume abc123
clinvk resume abc123 "now add tests"
```

### フィルターを組み合わせる

複数のフィルターを組み合わせます。

```bash
clinvk resume --here --backend claude --last
```

これは「現在のディレクトリ」かつ「Claude」の直近セッションを再開します。

## 挙動

`resume` コマンドは次の優先順位に従います。

1. `--last` が指定されている場合、フィルターに一致する直近の再開可能セッションを再開
2. それ以外でセッション ID が指定されている場合、そのセッションを再開
3. それ以外は対話式ピッカーを開く（再開可能なセッションがない場合はエラー）

## 出力

セッションを再開して AI の応答を表示します。出力形式はルートコマンドと同じオプションが適用されます。

### 出力例

```text
Resuming session abc123 (claude)

> continue from where we left off

I've reviewed the changes you made to the auth module. Here's what I found...
```

## よくあるエラー

| エラー | 原因 | 対処 |
|-------|-------|----------|
| `session not found` | セッション ID が存在しない | `clinvk sessions list` で有効な ID を確認 |
| `session not resumable` | バックエンドのセッション ID がない | 新しいセッションを開始する |
| `backend not available` | バックエンド CLI が未インストール | バックエンドをインストールする |
| `no resumable sessions` | 再開可能なセッションがない | 新しいセッションを開始する |

## 終了コード

| コード | 説明 |
|------|-------------|
| 0 | 成功 |
| 1 | セッションが見つからない、またはエラー |
| 2 | バックエンド利用不可 |

## 関連コマンド

- [sessions](sessions.md) - セッション一覧と管理
- [prompt](prompt.md) - 新しいプロンプトを実行

## 関連項目

- [セッション管理](../../guides/sessions.md) - セッション管理ガイド
- [設定リファレンス](../configuration.md) - セッション関連の設定
