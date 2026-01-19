# Security Guidelines

セキュリティに関するガイドライン。

## コミット前チェック

### 必須確認項目

- [ ] ハードコードされた認証情報がない
- [ ] ユーザー入力が検証されている
- [ ] ファイルパスがサニタイズされている
- [ ] エラーメッセージに機密情報が含まれていない

### bon3ai 固有のチェック

- [ ] ファイル操作でパストラバーサルを防止
- [ ] シェルコマンド実行時のインジェクション対策
- [ ] 外部コマンド (`chafa`, `git`, `jj`) の入力検証

## ファイル操作

### パストラバーサル防止

```go
// 悪い例
path := filepath.Join(baseDir, userInput)

// 良い例
path := filepath.Join(baseDir, userInput)
if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(baseDir)) {
    return errors.New("invalid path")
}
```

### シンボリックリンク

```go
// リンク先を確認
realPath, err := filepath.EvalSymlinks(path)
if err != nil {
    return err
}
```

## 外部コマンド実行

### コマンドインジェクション防止

```go
// 悪い例
exec.Command("sh", "-c", "git status " + userInput)

// 良い例
exec.Command("git", "status", "--", userInput)
```

### 引数の検証

```go
// パスの検証
if strings.ContainsAny(path, "\x00") {
    return errors.New("invalid path")
}
```

## 機密情報

### 禁止事項

- API キーのハードコード
- パスワードの平文保存
- 認証情報のログ出力

### 環境変数の使用

```go
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    return errors.New("API_KEY not set")
}
```

## 脆弱性発見時

1. **作業を停止**
2. **影響範囲を特定**
3. **修正を実施**
4. **類似の問題を監査**

## gitignore

以下は必ず除外:

```gitignore
# 認証情報
.env
*.pem
*.key
credentials.json

# 機密データ
secrets/
```
