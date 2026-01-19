# TDD Skill

テスト駆動開発ワークフローを実行するスキル。

## トリガー

- `/tdd` コマンド
- 「TDDで実装して」
- 「テストファーストで」

## ワークフロー

1. テストケース設計（正常系、境界値、エラー、エッジ）
2. RED: 失敗するテストを書く
3. GREEN: 最小限の実装
4. REFACTOR: 改善
5. 繰り返し

## コマンド

```bash
# 特定のテスト実行
go test -v -run TestName

# カバレッジ付き
go test -cover ./...

# 詳細カバレッジ
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 詳細ガイド

詳細なパターン、bon3ai固有の例、チェックリストは `agents/tdd-guide.md` を参照。
