# TDD Guide Agent

テスト駆動開発（TDD）を支援する専門エージェント。

## TDD サイクル

```
┌─────────────────────────────────────┐
│  1. RED: テストを書く（失敗する）    │
└─────────────────┬───────────────────┘
                  ↓
┌─────────────────────────────────────┐
│  2. GREEN: 最小限のコードで通す      │
└─────────────────┬───────────────────┘
                  ↓
┌─────────────────────────────────────┐
│  3. REFACTOR: コードを改善           │
└─────────────────┬───────────────────┘
                  ↓
              (繰り返し)
```

## 実践手順

### 1. RED: 失敗するテストを書く

```go
func TestNewFeature(t *testing.T) {
    result := NewFeature("input")

    if result != "expected" {
        t.Errorf("NewFeature() = %q, want %q", result, "expected")
    }
}
```

```bash
go test -v -run TestNewFeature
# --- FAIL: TestNewFeature
```

### 2. GREEN: テストを通す最小コード

```go
func NewFeature(input string) string {
    return "expected"  // 最小限の実装
}
```

```bash
go test -v -run TestNewFeature
# --- PASS: TestNewFeature
```

### 3. REFACTOR: 改善

- 重複を除去
- 命名を改善
- テストは常にパスさせる

## テストの種類

| 種類 | 対象 | 例 |
|------|------|-----|
| Unit | 単一関数 | `getCompletions()` |
| Integration | 複数コンポーネント | FileTree + VCS |
| 状態遷移 | InputMode 遷移 | Normal → Search → Normal |

## bon3ai での TDD

### Model のテスト
```go
func TestModel_StateTransition(t *testing.T) {
    m, _ := NewModel("/tmp")

    // Normal → Search
    m.inputMode = ModeNormal
    m, _ = m.Update(keyMsg("/"))

    if m.inputMode != ModeSearch {
        t.Errorf("expected ModeSearch, got %v", m.inputMode)
    }
}
```

### 補完機能のテスト
```go
func TestGetCompletions(t *testing.T) {
    tmpDir := t.TempDir()
    os.Mkdir(filepath.Join(tmpDir, "Documents"), 0755)

    candidates, _ := getCompletions(tmpDir+"/Do", "")

    if len(candidates) != 1 {
        t.Errorf("expected 1 candidate, got %d", len(candidates))
    }
}
```

## チェックリスト

### テスト作成時
- [ ] テストが先に書かれている
- [ ] テストが最初は失敗する
- [ ] テスト名が何をテストするか明確
- [ ] エッジケースをカバー

### 実装完了時
- [ ] 全テストがパス
- [ ] 不要なコードがない
- [ ] リファクタリング済み

## アンチパターン

❌ **避けるべき**:
- 実装後にテストを書く
- テスト間の依存関係
- 内部実装の詳細をテスト
- テストのためにコードを変更

✅ **推奨**:
- ユーザーから見える振る舞いをテスト
- 各テストは独立
- `t.TempDir()` で一時ファイル管理
- `t.Helper()` でヘルパー関数をマーク

## 使用タイミング

✅ 使う場面:
- 新機能の実装
- バグ修正（回帰テスト作成）
- リファクタリング前のテスト追加

❌ 使わない場面:
- 緊急のホットフィックス
- 実験的なプロトタイプ
