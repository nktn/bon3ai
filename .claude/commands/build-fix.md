# /build-fix - Build Error Fix

ビルドエラーを順番に修正する。

## 使用方法

```
/build-fix
```

## プロセス

### 1. エラー収集

```bash
go build . 2>&1
```

### 2. エラー分類

| 優先度 | エラータイプ |
|--------|-------------|
| High | 型エラー、未定義 |
| Medium | 未使用 import/変数 |
| Low | 警告 |

### 3. 順次修正

各エラーに対して:
1. エラー箇所のコンテキスト表示
2. 原因を特定
3. 最小限の修正を適用
4. 再ビルドで確認

### 4. 停止条件

- 修正で新たなエラーが発生
- 同じエラーが3回続く
- 全エラー解消

## 出力フォーマット

```markdown
## Build Fix Report

### Errors Found
1. `file.go:10`: undefined: FunctionName
2. `file.go:20`: cannot use X as Y

### Fixes Applied
1. ✅ Added import for FunctionName
2. ✅ Fixed type conversion

### Status
- Resolved: 2
- Remaining: 0
- New issues: 0
```

## 原則

- **1つずつ修正**: 一度に複数の修正をしない
- **最小限の変更**: リファクタリングしない
- **検証**: 各修正後に再ビルド
