# /update-docs - Documentation Update

ドキュメントをコードと同期する。

## 使用方法

```
/update-docs
/update-docs readme
/update-docs keybindings
```

## 対象ドキュメント

| ファイル | 同期元 |
|----------|--------|
| `README.md` | キーバインド → `update.go` |
| `CLAUDE.md` | アーキテクチャ → `*.go` ファイル構成 |
| `.claude/rules/architecture.md` | 状態遷移 → `model.go`, `update.go` |

## プロセス

### 1. 差分検出

```bash
# 最近変更されたファイル
git diff --name-only HEAD~5

# キーハンドリング変更
git diff HEAD~5 -- update.go | grep -E "case \"|key ==\""
```

### 2. 同期チェック

- README.md のキーバインド表 vs update.go の実装
- architecture.md の InputMode vs model.go の定義
- CLAUDE.md のファイル一覧 vs 実際のファイル

### 3. 更新実行

必要な箇所を更新:
- キーバインド追加/削除
- 状態遷移図の更新
- ファイル一覧の更新

### 4. 検証

- [ ] キーバインド表が update.go と一致
- [ ] 状態遷移図が model.go と一致
- [ ] ファイルパスが実在する

## 出力フォーマット

```markdown
## Documentation Update Report

### Files Checked
- README.md
- CLAUDE.md
- .claude/rules/architecture.md

### Updates Made
1. README.md: Added `gn` keybinding
2. architecture.md: Updated state diagram

### Verification
- [x] All keybindings match code
- [x] All file paths exist
```

## よくある更新パターン

### 新しいキーバインド追加時

1. `update.go` でキーハンドリング実装
2. `README.md` のキーバインド表に追加
3. 必要なら `architecture.md` も更新

### 新しい InputMode 追加時

1. `model.go` で InputMode 定義
2. `architecture.md` の状態遷移図を更新
3. `README.md` に操作方法を追加
