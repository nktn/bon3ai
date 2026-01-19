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

## プロセス

### 1. RED: テストを書く

```bash
# 失敗するテストを作成
go test -v -run TestNewFeature
# --- FAIL
```

### 2. GREEN: 実装する

```bash
# 最小限のコードでテストを通す
go test -v -run TestNewFeature
# --- PASS
```

### 3. REFACTOR: 改善する

```bash
# 全テストが通ることを確認しながらリファクタリング
go test ./...
```

## 出力フォーマット

```markdown
## TDD: <機能名>

### Test Cases
1. <テストケース1>
2. <テストケース2>
3. <エッジケース>

### RED Phase
- File: `xxx_test.go`
- Test: `TestXxx`

### GREEN Phase
- File: `xxx.go`
- Implementation: <簡潔な説明>

### REFACTOR Phase
- <改善点>

### Verification
- [ ] All tests pass
- [ ] No regression
```
