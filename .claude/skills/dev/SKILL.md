# Dev Orchestrator Skill

複数エージェントを並列連携させる開発オーケストレーター。

## トリガー

- `/dev` コマンド
- 「開発して」「実装して」+ 複雑な要求

## 実行ロジック

### Phase 1: 並列分析

Task ツールで複数エージェントを **同時に** 起動:

```
Task 1: planner agent
  prompt: "以下の機能の実装計画を作成: <要求>"

Task 2: architect agent
  prompt: "以下の機能のアーキテクチャ影響を分析: <要求>"

Task 3: tdd-guide agent
  prompt: "以下の機能のテストケースを設計: <要求>"
```

**重要**: 3つの Task を同一メッセージで送信し並列実行。

### Phase 2: 結果統合

各エージェントの結果を統合:

```markdown
## 統合計画

### From planner:
<実装ステップ>

### From architect:
<影響範囲、注意点>

### From tdd-guide:
<テストケース>

### 実行順序:
1. テスト作成
2. 実装
3. リファクタリング
4. ドキュメント
```

### Phase 3: 順次実行

統合計画に基づいて順次実行:

```
1. テストファイル作成 (RED)
   → go test で失敗確認

2. 実装 (GREEN)
   → go test で成功確認

3. リファクタリング (REFACTOR)
   → go test で成功維持確認

4. ドキュメント更新
   → /update-docs 実行
```

## モード別の並列構成

### --mode=feature

```python
parallel([
    Task(agent="planner", prompt="実装計画"),
    Task(agent="architect", prompt="アーキテクチャ分析"),
    Task(agent="tdd-guide", prompt="テスト設計"),
])
```

### --mode=fix

```python
parallel([
    Task(agent="tdd-guide", prompt="回帰テスト設計"),
    Task(agent="build-fixer", prompt="関連エラー確認"),
])
```

### --mode=refactor

```python
parallel([
    Task(agent="architect", prompt="構造分析"),
    Task(agent="refactor-cleaner", prompt="改善ポイント"),
    Bash("go test -cover ./..."),  # カバレッジ
])
```

## 実装パターン

### 並列 Task 実行

```
<thinking>
3つのエージェントを並列で起動する必要がある。
同一メッセージ内で複数の Task ツールを呼び出す。
</thinking>

[Task: planner - 実装計画作成]
[Task: architect - アーキテクチャ分析]
[Task: tdd-guide - テスト設計]
```

### 結果待機と統合

```
<thinking>
3つの Task 結果を受け取った。
結果を統合して次のフェーズに進む。
</thinking>

## 統合結果
...
```

## エラーハンドリング

### いずれかのエージェントが失敗

```
並列実行結果:
- planner: ✅ 成功
- architect: ❌ 失敗 (理由: ...)
- tdd-guide: ✅ 成功

→ architect の分析を手動で補完するか、
   成功した結果のみで続行するか確認
```

### ビルドエラー発生

```
Phase 3 で go build 失敗
→ /build-fix を自動実行
→ 修正後に続行
```

## bon3ai 固有の連携

### InputMode 追加時の並列構成

```
parallel([
    Task("model.go に InputMode 定義を計画"),
    Task("update.go のキーハンドリングを計画"),
    Task("view.go のレンダリングを計画"),
    Task("テストケースを設計"),
])
```

### VCS 機能追加時

```
parallel([
    Task("vcs.go インターフェース拡張を計画"),
    Task("gitstatus.go の Git 実装を計画"),
    Task("jjstatus.go の JJ 実装を計画"),
])
```
