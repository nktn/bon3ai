# TDD Skill

テスト駆動開発ワークフローを実行するスキル。

## トリガー

- `/tdd` コマンド
- 「TDDで実装して」
- 「テストファーストで」

## ワークフロー

### Step 1: テストケース設計

機能要件からテストケースを洗い出す:

- 正常系
- 境界値
- エラーケース
- エッジケース

### Step 2: RED - 失敗するテストを書く

```go
// xxx_test.go
func TestFeatureName(t *testing.T) {
    // Arrange
    input := "test input"

    // Act
    result := FeatureName(input)

    // Assert
    expected := "expected output"
    if result != expected {
        t.Errorf("FeatureName(%q) = %q, want %q", input, result, expected)
    }
}
```

実行して失敗を確認:

```bash
go test -v -run TestFeatureName
```

### Step 3: GREEN - 最小実装

テストを通す最小限のコードを書く:

```go
// xxx.go
func FeatureName(input string) string {
    return "expected output"
}
```

### Step 4: REFACTOR - 改善

テストが通る状態を維持しながら:

- 重複を除去
- 命名を改善
- 構造を整理

```bash
go test ./...  # 全テスト通過を確認
```

### Step 5: 繰り返し

次のテストケースへ進む。

## bon3ai 固有のパターン

### Model テスト

```go
func TestModel_Feature(t *testing.T) {
    m, _ := NewModel(t.TempDir())
    defer m.watcher.Close()

    // テストロジック
}
```

### 状態遷移テスト

```go
func TestModeTransition(t *testing.T) {
    m := &Model{inputMode: ModeNormal}

    m, _ = m.Update(keyMsg("key"))

    if m.inputMode != ExpectedMode {
        t.Errorf("got %v, want %v", m.inputMode, ExpectedMode)
    }
}
```

### ファイルシステムテスト

```go
func TestFileOperation(t *testing.T) {
    tmpDir := t.TempDir()  // 自動クリーンアップ

    // ファイル作成
    os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644)

    // テストロジック
}
```

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
