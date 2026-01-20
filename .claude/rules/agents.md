# Agent Usage Guide

エージェントの効果的な使い方。

## 利用可能なエージェント

| Agent | 用途 | トリガー |
|-------|------|----------|
| `planner` | 実装計画 | 複雑な機能リクエスト |
| `architect` | 設計判断 | アーキテクチャ変更 |
| `tui-designer` | TUI 設計 | UI コンポーネント、InputMode 追加 |
| `tdd-guide` | TDD支援 | テスト駆動開発 |
| `build-fixer` | ビルド修正 | コンパイルエラー |
| `refactor-cleaner` | リファクタリング | コード整理 |
| `doc-updater` | ドキュメント | コード変更後 |

## 使用タイミング

### 即座にエージェントを使う場面

- **複雑な機能リクエスト** → `planner`
- **アーキテクチャ変更** → `architect`
- **UI コンポーネント設計** → `tui-designer`
- **ビルド失敗** → `build-fixer`
- **コード変更完了** → `doc-updater`

### 組み合わせパターン

```
新機能開発:
  planner + tui-designer + architect + tdd-guide → doc-updater

リファクタリング:
  tui-designer + architect + refactor-cleaner → doc-updater

バグ修正:
  tui-designer + tdd-guide + build-fixer
```

> **Note**: bon3ai は TUI アプリのため、全パターンで `tui-designer` が参加

## 並列実行

独立した操作は並列で実行:

```
# 良い例: 並列実行
- セキュリティチェック
- パフォーマンス分析
- 型チェック

# 悪い例: 順次実行（依存関係がない場合）
```

## bon3ai 固有のガイダンス

### InputMode 追加時
1. `planner` で計画
2. `tui-designer` で OpenTUI パターン参照・UI 設計
3. `architect` でアーキテクチャ影響分析
4. `tdd-guide` でテスト作成
5. 実装
6. `doc-updater` で architecture.md 更新

### 新しいキーバインド追加時
1. `tdd-guide` でテスト作成
2. 実装
3. `doc-updater` で README.md 更新

### 大規模リファクタリング時
1. `tui-designer` で UI パターン参照
2. `architect` で設計レビュー
3. `refactor-cleaner` で段階的実行
4. `doc-updater` で CLAUDE.md 更新
