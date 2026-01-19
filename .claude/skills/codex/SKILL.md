---
name: codex
description: >
  OpenAI Codex CLIを使用したコードレビュー、分析、コードベースへの質問を実行する。
  使用場面: (1) コードレビュー依頼時、(2) コードベース全体の分析、(3) 実装に関する質問、
  (4) バグの調査、(5) リファクタリング提案、(6) 解消が難しい問題の調査。
  トリガー: "codex", "コードレビュー", "レビューして", "分析して", "/codex"
---

# Codex

Codex CLIを使用してコードレビュー・分析を実行するスキル。

## 実行コマンド

codex exec --full-auto --sandbox read-only --cd <project_directory> "<request>"

## パラメータ

| パラメータ | 説明 |
|-----------|------|
| `--full-auto` | 完全自動モードで実行 |
| `--sandbox read-only` | 読み取り専用サンドボックス（安全な分析用） |
| `--cd <dir>` | 対象プロジェクトのディレクトリ |
| `"<request>"` | 依頼内容（日本語可） |

## 使用例

### コードレビュー
codex exec --full-auto --sandbox read-only --cd /path/to/project "このプロジェクトのコードをレビューして、改善点を指摘してください"

### バグ調査
codex exec --full-auto --sandbox read-only --cd /path/to/project "認証処理でエラーが発生する原因を調査してください"

## 実行手順

1. ユーザーから依頼内容を受け取る
2. 対象プロジェクトのディレクトリを特定する
3. 上記コマンド形式でCodexを実行
4. 結果をユーザーに報告
5. **PRが存在する場合**: レビュー結果をPRコメントに追記

## PRコメント追記

レビュー完了後、オープンなPRがある場合は自動的にコメントを追記する。

### コメント形式

```markdown
## Codex Review

### Findings
- **高**: <重大な問題>
- **中**: <中程度の問題>
- **低**: <軽微な問題>

### Suggestions
<改善提案>

---
🤖 Reviewed by Codex CLI
```

### 実行コマンド

```bash
# 現在のブランチのPR番号を取得
gh pr view --json number -q .number

# PRにコメント追記
gh pr comment <PR番号> --body "<レビュー結果>"
```

### 条件

- PRがオープン状態であること
- レビューで指摘事項がある場合のみコメント追記
- 指摘なしの場合は「問題なし」の簡潔なコメント
