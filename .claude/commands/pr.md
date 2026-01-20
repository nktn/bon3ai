# /pr - Pull Request 操作

現在のブランチに関連する Pull Request の操作を行う。

## 使用方法

```
/pr                    # PR 作成
/pr status             # PR 状態確認
/pr comment <内容>     # PR にコメント追加
/pr revert             # PR をリバート
/pr merge              # PR をマージ
```

## サブコマンド

### /pr (作成)

現在のブランチから PR を作成。

```bash
# 内部で実行
git status                      # 未コミット変更確認
git log main..HEAD              # コミット確認
git diff main...HEAD            # 差分確認
gh pr create --title "..." --body "..."
```

**未コミット変更がある場合:**
1. 変更内容を表示（変更ファイル・新規ファイル一覧）
2. コミット対象を確認:
   - 全てコミット
   - 機能別に分割（複数コミット）
   - 個別にファイル選択
3. コミットメッセージ作成
4. コミット後に PR 作成

**既存 PR がある場合:**
- OPEN → PR URL を表示
- MERGED → 新規 PR 作成を提案

**出力:**
```
PR #42 を作成しました
https://github.com/user/repo/pull/42

Next: /codex でコードレビュー
```

### /pr status

現在のブランチの PR 状態を確認。

```bash
# 内部で実行
gh pr view --json number,title,state,reviews,checks
```

**出力:**
```
PR #42: Add fuzzy search feature
State: OPEN
Checks: ✓ All passing
Reviews: 1 approved, 0 changes requested
```

### /pr comment <内容>

現在のブランチの PR にコメントを追加。
主にコードレビュー後の意思決定結果を記録するために使用。

```bash
# 内部で実行
gh pr view --json number -q .number  # PR 番号取得
gh pr comment <番号> --body "<内容>"
```

**使用例:**
```
# レビュー指摘への意思決定
/pr comment 指摘1: 修正します、指摘2: パフォーマンス影響軽微のため見送り

# 作業状況の共有
/pr comment WIP: 明日続きやります
```

### /pr revert

現在のブランチの PR (マージ済み) をリバート。

```bash
# 内部で実行
gh pr view --json number,mergeCommit -q '.mergeCommit.oid'
git revert <マージコミット>
# 新しいリバート PR を作成
```

**出力:**
```
PR #42 をリバートしました
Revert PR: #43
https://github.com/user/repo/pull/43
```

### /pr merge

現在のブランチの PR をマージ。

```bash
# 内部で実行
gh pr merge --squash  # または --merge, --rebase
```

**確認:**
- CI チェックが通っているか
- レビュー承認があるか

## PR 自動検出

すべてのサブコマンドは現在のブランチから PR を自動検出:

```bash
gh pr view --json number -q .number
```

PR が見つからない場合:
- `/pr` (作成) → 新規作成を提案
- その他 → エラーメッセージ表示

## ワークフローでの使い方

```
/update-docs → /pr → /codex → /pr comment (意思決定) → 修正 → /codex
```

<user-request>
$ARGUMENTS
</user-request>
