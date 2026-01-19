# /tdd - Test-Driven Development

TDD サイクルで機能を実装する。

## 使用方法

```
/tdd <機能の説明>
```

## 例

```
/tdd Add file search feature
/tdd Implement bookmark functionality
/tdd Fix completion not working with special characters
```

## サイクル

```
RED → GREEN → REFACTOR → (繰り返し)
```

1. **RED**: 失敗するテストを書く
2. **GREEN**: 最小限のコードでテストを通す
3. **REFACTOR**: テストが通る状態でコードを改善

## 詳細

詳細なガイドラインは `agents/tdd-guide.md` を参照。

- テストパターン
- bon3ai 固有の例（Model, 状態遷移, ファイルシステム）
- チェックリスト
- アンチパターン
