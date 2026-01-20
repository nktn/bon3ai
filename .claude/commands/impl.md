# /impl - Implementation

計画とテストに基づいて実装を行う。

## 使用場面

- `/feature-plan` で計画作成後
- `/tdd` でテスト作成後
- 実装の方向性が決まっている時

## ワークフロー

### Step 1: コンテキスト収集

1. **計画確認**: 直前の `/feature-plan` 出力があれば参照
2. **テスト確認**: 関連する `*_test.go` ファイルを読み込み
3. **既存コード確認**: 変更対象ファイルの現状を把握

### Step 2: 実装方針決定

$ARGUMENTS から実装対象を特定:

```
/impl Add bookmark save function
      ├── 対象ファイル: fileops.go or 新規 bookmark.go
      ├── 関連テスト: bookmark_test.go
      └── 依存: Model struct, filetree.go
```

### Step 3: ルール確認

実装前に以下を確認:

| ルール | チェック項目 |
|--------|-------------|
| `coding-style.md` | 命名規則、関数サイズ、ファイルサイズ |
| `security.md` | パストラバーサル、入力検証 |
| `performance.md` | メモリ割り当て、ループ最適化 |

### Step 4: 実装実行

1. **テストが通るコードを書く** (TDD の場合)
2. **既存パターンに従う** (codebase の一貫性)
3. **最小限の変更** (over-engineering 禁止)

### Step 5: 検証

```bash
make test           # テスト実行
go vet ./...        # 静的解析
go fmt ./...        # フォーマット
```

### Step 6: サマリー

```
=== Implementation Complete ===

Files modified:
- fileops.go (+45 lines)
- model.go (+3 lines)

Tests passing: ✓

Next: /build-fix (if needed) → /codex
```

## 重要ルール

1. **テストを先に読む** - 何を実装すべきか明確にする
2. **既存コードを先に読む** - パターンを理解してから書く
3. **小さく実装** - 一度に大きな変更をしない
4. **over-engineering 禁止** - 必要なものだけ実装

## TUI 実装時

InputMode や UI コンポーネント実装時は `/opentui` でパターン参照:

```bash
/opentui keyboard handling  # キーバインド設計参考
/impl Add ModeFilter        # 実装
```

<user-request>
$ARGUMENTS
</user-request>
