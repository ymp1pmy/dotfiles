---
name: split-commit
description: >-
  git diff を分析して変更を論理的なコミットに分割する。
  複数の無関係な変更（ステージ済み・未ステージ）を個別にコミットしたい場合に使用。
  サブエージェントで差分解析し、contextual-commit でメッセージを生成する。
---

# Split Commit

`git diff` で変更（ステージ済み・未ステージ）を分析し、**ファイル単位ではなくハンク単位**で論理的なアトミックコミットに分割する。

## 使用タイミング

- 作業ディレクトリに複数の無関係な変更がある（ステージ済み・未ステージ問わず）
- ユーザーが「commit」「コミットして」と言い、変更が保留中
- **`git add -A` で全変更をステージしたが、アトミックコミットに分割したい**
- 1セッションで機能追加、バグ修正、リファクタ、設定変更が混在
- 1ファイルに複数の無関係な変更が含まれている

## 手順

### 1. ステージング状態の確認と差分取得

まず変更がどこにあるか確認:

```bash
# ステージ済み変更を確認
git diff --cached --stat

# 未ステージ変更を確認
git diff --stat
```

**変更が既にステージ済みの場合（例: `git add -A` 後）:**

```bash
# ステージ済み差分を取得
git diff --cached
```

**変更が未ステージの場合:**

```bash
# 未ステージ差分を取得
git diff
```

**両方存在する場合:**
両方を組み合わせて完全な分析を行う。

### 2. 再グループ化のためのアンステージ（必要に応じて）

変更が既にステージ済みで分割が必要な場合:

```bash
git reset HEAD  # 全てアンステージ、変更は作業ディレクトリに保持
```

これにより次のステップでハンク単位のステージングが可能になる。

### 3. 差分の分析（ハンク単位）

差分出力全体をレビュー。各ハンク（`@@` マーク）は個別の変更単位を表す。

### 4. 意味的な意図でハンクをグループ化

**Explore サブエージェント**を起動して差分分析:

```
サブエージェント: Agent ツールの subagent_type: "Explore" を使用（読み取り専用・高速）
タスク: git diff 出力を分析し、ハンク（ファイルではない）を意味的な作業単位でグループ化。

各ハンクについて以下を特定:
- どの論理的変更に属するか（機能、バグ修正、リファクタ、設定、ドキュメント等）
- 変更の意図

以下のJSON構造を返す:
{
  "groups": [
    {
      "name": "auth-session-fix",
      "type": "fix",
      "intent": "セッション有効期限処理の修正",
      "hunks": [
        {"file": "src/auth/session.ts", "hunk_header": "@@ -45,7 +45,9 @@"},
        {"file": "src/auth/login.ts", "hunk_header": "@@ -12,3 +12,5 @@"}
      ]
    },
    ...
  ]
}
```

### 5. 各グループのステージングとコミット（自動）

各グループに対して**確認を求めずに**:

1. **パッチ適用でハンクをステージ:**

   特定のハンクをパッチファイルに抽出して適用:

   ```bash
   git diff src/file.ts | head -n 30 > /tmp/partial.patch
   git apply --cached /tmp/partial.patch
   ```

   または printf でパイプして git add -p を使用:

   ```bash
   printf 'y\nn\ny\n' | git add -p
   ```

2. **`.claude/skills/contextual-commit` スキルを呼び出し**:
   - ステージ済み差分をコンテキストとして渡す
   - contextual-commit が gitmoji、件名、アクション行を処理

3. 生成されたメッセージで**コミット**

4. 残りのグループに対して**繰り返し**

### 6. 検証

```bash
git log --oneline -n <コミット数>
git diff --stat  # 全てコミット済みなら空のはず
```

## エージェント選択

| 優先度 | subagent_type | ユースケース |
| ------ | ------------- | ------------ |
| 1 | `Explore` | 高速分析、読み取り専用の探索 |
| 2 | `general-purpose` | 複雑な分析、複数ファイルにまたがる調査 |

## contextual-commit との連携

各コミットは必ず `.claude/skills/contextual-commit` を使用:

- gitmoji ショートコード接頭辞（`:sparkles:`、`:bug:` 等）
- Conventional Commit 形式の件名
- 本文にアクション行（intent、decision、rejected、constraint、learned）

**コミットメッセージを直接書かないこと。** 常に contextual-commit に委譲する。

## ハンクステージング技法

### 方法A: パイプ入力（`git add -p`）

ハンクへの y/n 応答をパイプ:

```bash
printf 'y\nn\ny\n' | git add -p
```

### 方法B: パッチ適用

1. 特定のハンクをパッチファイルに抽出
2. `git apply --cached` で適用

```bash
git diff src/file.ts | head -n 30 > /tmp/partial.patch
git apply --cached /tmp/partial.patch
```

## フロー例

### 例A: ステージ済み変更（`git add -A` 実行後）

```
ユーザー: git add -A した、コミットして

1. ステージング状態を確認:
   - git diff --cached --stat → 59ファイル変更
   - git diff --stat → なし（全てステージ済み）

2. 再グループ化のためアンステージ:
   - git reset HEAD

3. git diff → 59ファイルにハンク

4. Explore サブエージェントが意図別にグループ化（例）:
   - グループA: "fix-auth" → 認証関連ハンク
   - グループB: "update-config" → 設定ファイルハンク
   - グループC: "refactor-utils" → ユーティリティ関数ハンク
   - グループD: "add-feature-x" → 新機能ハンク

5. 各グループに対して（自動、確認なし）:
   - git add -p または git apply --cached で特定ハンクをステージ
   - contextual-commit → メッセージ生成
   - git commit

6. 結果: 複数のアトミックコミット、各々が単一の意図
```

### 例B: 未ステージ変更

```
ユーザー: コミットして

1. ステージング状態を確認:
   - git diff --cached --stat → なし（ステージなし）
   - git diff --stat → 4ファイル変更

2. リセット不要（既に未ステージ）

3. git diff → 2ファイルに4ハンク:
   - src/auth/session.ts: 2ハンク（セッション修正 + 無関係なログ出力）
   - config/app.yml: 2ハンク（タイムアウト設定 + 新機能フラグ）

4. Explore サブエージェントが意図別にグループ化:
   - グループA: "session-fix" → session.ts ハンク1
   - グループB: "add-logging" → session.ts ハンク2
   - グループC: "config-timeout" → app.yml ハンク1
   - グループD: "feature-flag" → app.yml ハンク2

5. 各グループに対して（自動、確認なし）:
   - 特定ハンクをステージ
   - contextual-commit → メッセージ生成
   - git commit

6. 結果: 4つのアトミックコミット、各々が単一の意図
```

## 出力規約

- 作成したコミット数を報告
- 各コミットの件名をリスト
- 全変更がコミット済みであることを確認（または残りの未コミットハンクをリスト）
