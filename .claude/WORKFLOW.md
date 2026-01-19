# Development Workflow

Claude Code を活用した bon3ai 開発ワークフロー。

## 基本サイクル

```
┌─────────────────────────────────────────────────────────────┐
│                     開発ワークフロー                          │
└─────────────────────────────────────────────────────────────┘

  ユーザー要求
       │
       ▼
  ┌─────────┐
  │  /plan  │ ─── 複雑な機能は計画から
  └────┬────┘
       │
       ▼
  ┌─────────┐
  │  /tdd   │ ─── テストファースト
  └────┬────┘
       │
       ▼
  ┌─────────┐
  │  実装   │ ─── rules に従ってコーディング
  └────┬────┘
       │
       ▼
  ┌──────────────┐
  │ /build-fix   │ ─── ビルドエラーがあれば修正
  └──────┬───────┘
       │
       ▼
  ┌─────────┐
  │ /codex  │ ─── コードレビュー
  └────┬────┘
       │
       ▼
  ┌──────────────┐
  │ /update-docs │ ─── ドキュメント同期
  └──────┬───────┘
       │
       ▼
    PR 作成
```

---

## シナリオ別ワークフロー

### 1. 新機能追加（例: ブックマーク機能）

```bash
# Step 1: 計画
/plan Add bookmark feature for frequently accessed directories

# Step 2: TDD で実装
/tdd Implement bookmark save/load

# Step 3: ビルド確認
/build-fix

# Step 4: レビュー
/codex コードレビューして

# Step 5: ドキュメント更新
/update-docs

# Step 6: PR 作成
gh pr create
```

**使用される Rules:**
- `coding-style.md` → コード規約
- `testing.md` → テスト作成
- `git-workflow.md` → コミットメッセージ

**使用される Agents:**
- `planner` → 計画作成
- `tdd-guide` → TDD 支援
- `doc-updater` → ドキュメント同期

---

### 2. バグ修正

```bash
# Step 1: 問題の再現テスト作成
/tdd Write regression test for the bug

# Step 2: 修正
# (rules/coding-style.md に従う)

# Step 3: ビルド確認
/build-fix

# Step 4: レビュー
/codex バグ修正をレビューして

# Step 5: PR 作成
```

**使用される Rules:**
- `testing.md` → 回帰テスト
- `security.md` → セキュリティ確認

---

### 3. リファクタリング

```bash
# Step 1: 現状分析
/test-coverage

# Step 2: リファクタリング計画
# (agents/refactor-cleaner.md のパターン参照)

# Step 3: 段階的にリファクタリング
/refactor-clean update.go

# Step 4: テスト確認
make test

# Step 5: レビュー
/codex リファクタリングをレビューして

# Step 6: ドキュメント更新（構造変更があれば）
/update-codemaps
```

**使用される Rules:**
- `coding-style.md` → 品質基準
- `performance.md` → パフォーマンス考慮

---

### 4. 新しい InputMode 追加

```bash
# Step 1: 計画
/plan Add ModeFilter for filtering file list

# Step 2: TDD
/tdd Implement filter mode state transitions

# Step 3: 実装
# model.go: InputMode 追加
# update.go: キーハンドリング
# view.go: レンダリング

# Step 4: ビルド確認
/build-fix

# Step 5: レビュー
/codex

# Step 6: ドキュメント更新（必須）
/update-docs  # README.md キーバインド表
# .claude/rules/architecture.md 状態遷移図を手動更新
```

**使用される Agents:**
- `planner` → 設計
- `architect` → アーキテクチャ確認
- `doc-updater` → ドキュメント同期

---

### 5. パフォーマンス改善

```bash
# Step 1: プロファイリング
go test -cpuprofile=cpu.out -bench .
go tool pprof cpu.out

# Step 2: 改善計画
/plan Optimize file tree loading performance

# Step 3: 実装
# (rules/performance.md のパターン参照)

# Step 4: ベンチマーク比較
go test -bench . -benchmem

# Step 5: レビュー
/codex パフォーマンス改善をレビューして
```

---

## クイックリファレンス

### よく使うコマンド

| 状況 | コマンド |
|------|----------|
| 何から始めるか分からない | `/plan` |
| 新機能を実装したい | `/tdd` |
| ビルドが通らない | `/build-fix` |
| コードをレビューしたい | `/codex` |
| テストカバレッジを見たい | `/test-coverage` |
| ドキュメントを更新したい | `/update-docs` |
| コードを整理したい | `/refactor-clean` |

### Rules の適用タイミング

| Rule | いつ参照するか |
|------|----------------|
| `coding-style.md` | コード書く時は常に |
| `testing.md` | テスト書く時 |
| `git-workflow.md` | コミット/PR 時 |
| `security.md` | ファイル操作、外部コマンド時 |
| `performance.md` | パフォーマンス懸念時 |
| `agents.md` | どのエージェント使うか迷った時 |

### Agents の使い分け

| Agent | 使う場面 |
|-------|----------|
| `planner` | 複雑な機能、何から始めるか |
| `architect` | 設計判断、大きな構造変更 |
| `tdd-guide` | テストの書き方 |
| `build-fixer` | ビルドエラー |
| `refactor-cleaner` | コード整理 |
| `doc-updater` | ドキュメント更新 |
