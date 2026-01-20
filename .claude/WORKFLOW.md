# Development Workflow

Claude Code を活用した bon3ai 開発ワークフロー。

## ワークフロー選択

| パターン | コマンド | 特徴 |
|---------|---------|------|
| **自動オーケストレーション** | `/dev` | エージェント並列実行、全ステップ自動連携 |
| **手動** | `/plan` → `/tdd` → `/impl` → ... | ステップごとに確認・調整可能 |

### 自動オーケストレーション (`/dev`) - 推奨

```bash
/dev Add <機能の説明>
```

複数エージェントが並列で動き、計画→テスト→実装→レビュー→ドキュメントまで自動連携。

**向いているケース:**
- 明確な機能追加
- 標準的なバグ修正
- 定型的なリファクタリング

### 手動パターン

```bash
/plan → /tdd → /impl → /build-fix → /update-docs → /pr → /codex → (修正 → /codex)
```

各ステップで結果を確認しながら進める。

**向いているケース:**
- 複雑な設計判断が必要
- 途中で方針を変える可能性がある
- 学習目的で各ステップを理解したい

---

## 手動パターン詳細

### 基本フロー

```
┌─────────────────────────────────────────────────────────────┐
│                     手動開発ワークフロー                       │
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
  │  /impl  │ ─── 計画とテストに基づいて実装
  └────┬────┘
       │
       ▼
  ┌──────────────┐
  │ /build-fix   │ ─── ビルドエラーがあれば修正
  └──────┬───────┘
       │
       ▼
  ┌──────────────┐
  │ /update-docs │ ─── ドキュメント同期
  └──────┬───────┘
       │
       ▼
  ┌─────────┐
  │   /pr   │ ─── PR 作成
  └────┬────┘
       │
       ▼
  ┌─────────┐
  │ /codex  │ ─── PR コードレビュー
  └────┬────┘
       │
       ▼
   修正あり？ ──No──→ 完了
       │
      Yes
       │
       ▼
  ┌─────────────────────────────────┐
  │ /pr comment 意思決定結果          │
  │ (例: 指摘1修正、指摘2見送り理由)   │
  └────────────┬────────────────────┘
       │
       ▼
  ┌─────────────────┐
  │  修正実施        │
  │  (必要に応じて    │
  │  /update-docs)  │
  └────────┬────────┘
       │
       └──→ /codex (再レビュー)
```

---

## シナリオ別ワークフロー

### 1. 新機能追加（例: ブックマーク機能）

```bash
# Step 1: 計画
/plan Add bookmark feature for frequently accessed directories

# Step 2: テスト作成
/tdd Implement bookmark save/load

# Step 3: 実装
/impl Add bookmark save/load functions

# Step 4: ビルド確認
/build-fix

# Step 5: ドキュメント更新
/update-docs

# Step 6: /pr
/pr

# Step 7: コードレビュー
/codex

# Step 8: 修正があれば意思決定をコメント
/pr comment 指摘1修正、指摘2は○○の理由で見送り

# Step 9: 修正実施（必要に応じて /update-docs も再実行）

# Step 10: 再レビュー
/codex
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

# Step 2: 修正実装
/impl Fix the bug

# Step 3: ビルド確認
/build-fix

# Step 4: /pr
/pr

# Step 5: コードレビュー
/codex

# Step 6: 修正があれば /pr comment (意思決定) → 修正 → /codex
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

# Step 5: ドキュメント更新（構造変更があれば）
/update-codemaps

# Step 6: /pr
/pr

# Step 7: コードレビュー
/codex

# Step 8: 修正があれば /pr comment (意思決定) → 修正 → /codex
```

**使用される Rules:**
- `coding-style.md` → 品質基準
- `performance.md` → パフォーマンス考慮

---

### 4. 新しい InputMode 追加

```bash
# Step 0: TUI パターン参照（推奨）
/opentui input component patterns
/opentui keyboard handling

# Step 1: 計画
/plan Add ModeFilter for filtering file list

# Step 2: TDD
/tdd Implement filter mode state transitions

# Step 3: 実装
/impl Add ModeFilter
# model.go: InputMode 追加
# update.go: キーハンドリング
# view.go: レンダリング

# Step 4: ビルド確認
/build-fix

# Step 5: ドキュメント更新（必須）
/update-docs  # README.md キーバインド表
# .claude/rules/architecture.md 状態遷移図を手動更新

# Step 6: /pr
/pr

# Step 7: コードレビュー
/codex

# Step 8: 修正があれば /pr comment (意思決定) → 修正 → /codex
```

**使用される Skills:**
- `opentui` → TUI パターン参照

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
/impl Optimize file tree loading
# (rules/performance.md のパターン参照)

# Step 4: ベンチマーク比較
go test -bench . -benchmem

# Step 5: /pr
/pr

# Step 6: コードレビュー
/codex

# Step 7: 修正があれば /pr comment (意思決定) → 修正 → /codex
```

---

## 自動オーケストレーション詳細 (`/dev`)

### 全体フロー

```
/dev Add bookmark feature

┌─────────────────────────────────────────────────────────────┐
│ Phase 1: 並列分析                                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌────────┐ ┌────────────┐ ┌─────────┐ ┌──────────┐        │
│  │planner │ │tui-designer│ │architect│ │tdd-guide │ ← 並列  │
│  └───┬────┘ └─────┬──────┘ └────┬────┘ └────┬─────┘        │
│      │            │             │           │               │
│      └────────────┴──────┬──────┴───────────┘               │
│                          ▼                                  │
│                   ┌───────────┐                             │
│                   │ 結果統合   │                             │
│                   └─────┬─────┘                             │
└─────────────────────────┼───────────────────────────────────┘
                          │
┌─────────────────────────┼───────────────────────────────────┐
│ Phase 2: 順次実行        │                                   │
├─────────────────────────┼───────────────────────────────────┤
│                         ▼                                   │
│     /tdd → /impl → /build-fix → /update-docs               │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┼───────────────────────────────────┐
│ Phase 3: レビュー        │                                   │
├─────────────────────────┼───────────────────────────────────┤
│                         ▼                                   │
│     /pr → /codex → 修正対応 (→ /codex)                  │
└─────────────────────────────────────────────────────────────┘
```

### モード別の並列構成

| モード | Phase 1 (並列) | Phase 2 (順次) |
|--------|----------------|----------------|
| `feature` | planner + **tui-designer** + architect + tdd-guide | /tdd → /impl → /build-fix → /update-docs |
| `fix` | **tui-designer** + tdd-guide + build-fixer | /tdd → /impl → /build-fix |
| `refactor` | **tui-designer** + architect + refactor-cleaner + coverage | /refactor-clean → make test → /update-codemaps |

> **Note**: bon3ai は TUI アプリのため、全モードで `tui-designer` が OpenTUI パターンを参照

### 使用例

```bash
# 新機能（デフォルト・tui-designer が常に参加）
/dev Add split pane view

# バグ修正
/dev --mode=fix Fix completion not working with symlinks

# リファクタリング
/dev --mode=refactor Clean up update.go
```

---

## TUI パターン参照 (/opentui)

TUI コンポーネントやインタラクション設計で迷ったら、OpenTUI のパターンを参照。

### 参照タイミング

| 場面 | 参照先 |
|------|--------|
| 入力コンポーネント設計 | `components/inputs.md` |
| レイアウト・配置 | `layout/README.md`, `layout/patterns.md` |
| キーボード操作設計 | `keyboard/README.md` |
| アニメーション追加 | `animation/README.md` |
| テスト設計 | `testing/README.md` |

### 使用例

```bash
# 新しい入力モード設計時
/opentui input component patterns

# レイアウト参考
/opentui flexbox layout patterns

# キーボードナビゲーション設計
/opentui keyboard handling best practices
```

### Bubble Tea への適用

OpenTUI のパターンを Go/Bubble Tea に翻訳する際の対応:

| OpenTUI | Bubble Tea |
|---------|------------|
| React hooks | Model fields + Update |
| JSX components | View functions |
| Flexbox | lipgloss.JoinVertical/Horizontal |
| Event handlers | tea.Msg + Update |
| Focus management | InputMode state |

---

## クイックリファレンス

### よく使うコマンド

| 状況 | コマンド |
|------|----------|
| 何から始めるか分からない | `/plan` |
| テストを先に書きたい | `/tdd` |
| 計画・テスト後に実装 | `/impl` |
| ビルドが通らない | `/build-fix` |
| ドキュメントを更新したい | `/update-docs` |
| PR を作成したい | `/pr` |
| PR の状態を確認 | `/pr status` |
| PR にコメント | `/pr comment <内容>` |
| PR をリバート | `/pr revert` |
| PR をレビューしたい | `/codex` |
| テストカバレッジを見たい | `/test-coverage` |
| コードを整理したい | `/refactor-clean` |
| TUI パターンを参照 | `/opentui` |

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
