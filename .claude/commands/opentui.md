# /opentui - OpenTUI Development

OpenTUI フレームワークを使用した TUI 開発支援。

## 使用場面

- TUI コンポーネントの実装
- レイアウト・キーボード操作の実装
- アニメーションの追加
- テストの作成
- トラブルシューティング

## ワークフロー

### Step 1: タスク分析

$ARGUMENTS から以下を判定:
- **フレームワーク**: Core (imperative) / React / Solid
- **タスク種別**: 新規プロジェクト / コンポーネント / レイアウト / キーボード / デバッグ / テスト

### Step 2: 参照ファイル読み込み

`.claude/skills/opentui/references/` から関連ファイルを読み込み:

| タスク | 参照ファイル |
|--------|-------------|
| 新規プロジェクト | `<framework>/README.md` + `configuration.md` |
| コンポーネント | `<framework>/api.md` + `components/<category>.md` |
| レイアウト | `layout/README.md` + `layout/patterns.md` |
| キーボード | `keyboard/README.md` |
| アニメーション | `animation/README.md` |
| デバッグ | `<framework>/gotchas.md` + `testing/README.md` |
| テスト | `testing/README.md` |

### Step 3: 実行

OpenTUI のパターンと API を適用してタスクを完了。

### Step 4: サマリー

```
=== OpenTUI Task Complete ===

Framework: <core | react | solid>
Files referenced: <参照したファイル>

<実行内容のサマリー>
```

## 重要ルール

1. **`create-tui` で新規プロジェクト作成** - オプションは引数の前に指定
2. **`process.exit()` を直接呼ばない** - `renderer.destroy()` を使用
3. **React/Solid でのテキストスタイル** - props ではなくネストタグを使用

<user-request>
$ARGUMENTS
</user-request>
