# /test-coverage - Test Coverage Analysis

テストカバレッジの分析と改善。

## 使用方法

```
/test-coverage
/test-coverage view.go
/test-coverage --html
```

## コマンド

### カバレッジ取得

```bash
# 基本
go test -cover ./...

# 詳細レポート
go test -coverprofile=coverage.out ./...

# HTML レポート
go tool cover -html=coverage.out -o coverage.html

# 関数別カバレッジ
go tool cover -func=coverage.out
```

## プロセス

### 1. 現状分析

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### 2. 低カバレッジ箇所の特定

- カバーされていない関数
- 分岐の未テスト部分
- エラーハンドリングの未テスト

### 3. テスト追加

優先順位:
1. パブリック関数
2. エラーパス
3. エッジケース

### 4. 再測定

```bash
go test -cover ./...
```

## 出力フォーマット

```markdown
## Test Coverage Report

### Summary
- Total: 75.2%
- Files: 12
- Functions: 48

### By File
| File | Coverage | Uncovered Lines |
|------|----------|-----------------|
| model.go | 82% | 45-50, 120-125 |
| view.go | 68% | 200-230 |
| update.go | 71% | 300-320 |

### Uncovered Functions
- `view.go`: `renderPreview()` - 0%
- `fileops.go`: `moveFile()` - 45%

### Recommendations
1. Add tests for `renderPreview()`
2. Add error case tests for `moveFile()`

### Improvement Plan
- [ ] Test A: +5% coverage
- [ ] Test B: +3% coverage
```

## 目標

- 新規コード: 80%以上
- クリティカルパス: 90%以上
- 全体: 改善傾向を維持
