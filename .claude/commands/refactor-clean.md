# /refactor-clean - Refactor & Clean

不要コードの削除とリファクタリング。

## 使用方法

```
/refactor-clean
/refactor-clean view.go
/refactor-clean --unused
```

## チェック項目

### 不要コード検出

```bash
# 未使用の変数・関数（標準ツール）
go vet ./...

# 未使用の import（要インストール）
goimports -l .

# 静的解析（要インストール）
staticcheck ./...
```

### ツールのインストール

```bash
# goimports
go install golang.org/x/tools/cmd/goimports@latest

# staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
```

**未インストール時の代替**:
- `goimports` → `go fmt` + 手動で import 整理
- `staticcheck` → `go vet` のみで続行（カバレッジ低下を許容）

### コード品質

| チェック | 基準 |
|----------|------|
| 関数の長さ | 50行以下 |
| ファイルの長さ | 800行以下 |
| ネストの深さ | 4レベル以下 |
| 重複コード | なし |

## プロセス

### 1. 分析

- 未使用コードの特定
- 重複コードの検出
- 長い関数の特定

### 2. クリーンアップ

- 未使用 import 削除
- 未使用変数・関数削除
- デッドコード削除

### 3. リファクタリング

- 長い関数の分割
- 重複コードの抽出
- 命名の改善

### 4. 検証

```bash
go build .
go test ./...
```

## 出力フォーマット

```markdown
## Refactor & Clean Report

### Unused Code Removed
- `view.go`: Removed unused `oldFunction()`
- `model.go`: Removed unused import "fmt"

### Refactored
- `update.go`: Split `handleInput()` into 3 functions

### Quality Metrics
| File | Lines | Functions > 50 lines |
|------|-------|---------------------|
| view.go | 450 | 0 |
| update.go | 620 | 1 → 0 |

### Verification
- [x] Build passes
- [x] All tests pass
```

## 原則

- テストが通る状態を維持
- 動作を変えない
- 段階的に進める
