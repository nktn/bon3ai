# /dev - Orchestrated Development

複数のエージェントを並列で連携させて開発を進める。

## 使用方法

```
/dev <機能の説明>
/dev --mode=feature <説明>    # デフォルト (tui-designer 含む)
/dev --mode=fix <説明>
/dev --mode=refactor <説明>
```

## トリガー

- `/dev` コマンド
- 「開発して」「実装して」+ 複雑な要求

## モード

### feature (デフォルト): 新機能開発

```
Phase 1: 並列分析
├─ [planner]      → 実装計画作成
├─ [tui-designer] → OpenTUI パターン参照・UI 設計
├─ [architect]    → アーキテクチャ影響分析
└─ [tdd-guide]    → テストケース設計

↓ 結果を統合

Phase 2: 順次実行
├─ テスト作成 (/tdd)
├─ 実装 (/impl)
├─ ビルド確認 (/build-fix)
└─ ドキュメント更新 (/update-docs)

Phase 3: レビュー
├─ /pr
├─ コードレビュー (/codex)
└─ /pr comment (意思決定) → 修正 → /codex
```

### fix: バグ修正

```
Phase 1: 並列分析
├─ [tui-designer]   → UI パターン参照
├─ [tdd-guide]      → 回帰テスト設計
└─ [build-fixer]    → 関連エラー確認

↓ 結果を統合

Phase 2: 順次実行
├─ 回帰テスト作成 (/tdd)
├─ 修正実装 (/impl)
├─ ビルド確認 (/build-fix)
└─ 全テスト実行 (make test)

Phase 3: レビュー
├─ /pr
├─ コードレビュー (/codex)
└─ /pr comment (意思決定) → 修正 → /codex
```

### refactor: リファクタリング

```
Phase 1: 並列分析
├─ [tui-designer]      → UI パターン参照
├─ [architect]         → 構造分析
├─ [refactor-cleaner]  → 改善ポイント特定
└─ coverage 分析 (/test-coverage)

↓ 結果を統合

Phase 2: 順次実行
├─ 段階的リファクタリング (/refactor-clean)
├─ テスト確認 (make test)
└─ ドキュメント更新 (/update-codemaps)

Phase 3: レビュー
├─ /pr
├─ コードレビュー (/codex)
└─ /pr comment (意思決定) → 修正 → /codex
```

## 実行ロジック

### Phase 1: 並列分析

Task ツールで複数エージェントを **同時に** 起動:

```
Task 1: planner agent
  prompt: "以下の機能の実装計画を作成: <要求>"

Task 2: architect agent
  prompt: "以下の機能のアーキテクチャ影響を分析: <要求>"

Task 3: tdd-guide agent
  prompt: "以下の機能のテストケースを設計: <要求>"
```

**重要**: 3つの Task を同一メッセージで送信し並列実行。

### Phase 2: 結果統合

各エージェントの結果を統合:

```markdown
## 統合計画

### From planner:
<実装ステップ>

### From architect:
<影響範囲、注意点>

### From tdd-guide:
<テストケース>

### 実行順序:
1. テスト作成
2. 実装
3. リファクタリング
4. ドキュメント
```

### Phase 3: 順次実行

統合計画に基づいて順次実行:

```
1. テストファイル作成 (RED)
   → go test で失敗確認

2. 実装 (GREEN)
   → go test で成功確認

3. リファクタリング (REFACTOR)
   → go test で成功維持確認

4. ドキュメント更新
   → /update-docs 実行
```

## 実装パターン

### 並列 Task 実行

```
<thinking>
3つのエージェントを並列で起動する必要がある。
同一メッセージ内で複数の Task ツールを呼び出す。
</thinking>

[Task: planner - 実装計画作成]
[Task: architect - アーキテクチャ分析]
[Task: tdd-guide - テスト設計]
```

### 結果待機と統合

```
<thinking>
3つの Task 結果を受け取った。
結果を統合して次のフェーズに進む。
</thinking>

## 統合結果
...
```

## エラーハンドリング

### いずれかのエージェントが失敗

```
並列実行結果:
- planner: ✅ 成功
- architect: ❌ 失敗 (理由: ...)
- tdd-guide: ✅ 成功

→ architect の分析を手動で補完するか、
   成功した結果のみで続行するか確認
```

### ビルドエラー発生

```
Phase 3 で go build 失敗
→ /build-fix を自動実行
→ 修正後に続行
```

## bon3ai 固有の連携

### InputMode 追加時の並列構成

```
parallel([
    Task("model.go に InputMode 定義を計画"),
    Task("update.go のキーハンドリングを計画"),
    Task("view.go のレンダリングを計画"),
    Task("テストケースを設計"),
])
```

### VCS 機能追加時

```
parallel([
    Task("vcs.go インターフェース拡張を計画"),
    Task("gitstatus.go の Git 実装を計画"),
    Task("jjstatus.go の JJ 実装を計画"),
])
```

## 実行例

### 新機能開発

```
User: /dev Add fuzzy file search with fzf-like interface

Claude: 並列でエージェントを起動します...

[planner] 実装計画:
- model.go: ModeSearch 拡張、fuzzy match state
- update.go: 検索キーハンドリング
- view.go: 検索結果ハイライト表示
- search.go: fuzzy matching アルゴリズム

[architect] アーキテクチャ分析:
- 既存の ModeSearch を拡張可能
- FileTree との統合が必要
- パフォーマンス考慮: 大規模ディレクトリ対応

[tdd-guide] テストケース:
1. 完全一致検索
2. 部分一致検索
3. 大文字小文字無視
4. 空の検索結果
5. 特殊文字のエスケープ

統合計画を作成しました。実装を開始しますか？
```

## 自動連携のルール

### 並列実行可能な組み合わせ

| Agent 1 | Agent 2 | Agent 3 |
|---------|---------|---------|
| planner | architect | tdd-guide |
| build-fixer | tdd-guide | - |
| refactor-cleaner | architect | coverage |

### 順次実行が必要な場合

- 前のステップの出力が次の入力になる場合
- ファイル変更が競合する可能性がある場合

## 出力フォーマット

```markdown
## /dev Report: <機能名>

### Phase 1: 並列分析 (3 agents)
| Agent | Status | Summary |
|-------|--------|---------|
| planner | ✅ | 5 steps identified |
| architect | ✅ | No breaking changes |
| tdd-guide | ✅ | 8 test cases |

### Phase 2: 順次実行
- [x] テスト作成 (/tdd)
- [x] 実装 (/impl)
- [x] ビルド確認 (/build-fix)
- [x] ドキュメント更新 (/update-docs)

### Phase 3: レビュー
- [x] PR 作成
- [ ] コードレビュー (/codex)
- [ ] 修正対応

### Summary
- Tests: 8 passed
- Files changed: 4
- PR: #123
```

## 中断と再開

```
/dev --resume   # 前回の続きから
/dev --status   # 現在の進捗確認
/dev --abort    # 中断
```
