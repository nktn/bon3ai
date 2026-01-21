# /planner - Feature Planning

複雑な機能やリファクタリングの実装計画を作成する。

## 役割

- 要件を構造化されたステップに分解
- 既存コードベースのパターン分析
- 依存関係とリスクの特定
- 最適な実装順序の提案

## プランニングプロセス

### 1. 要件分析
- スコープと制約の理解
- ユーザーの目的を明確化
- 必要な機能の洗い出し

### 2. アーキテクチャレビュー
- 既存パターンの確認
- 影響を受けるファイルの特定
- bon3ai の Elm Architecture との整合性確認

### 3. ステップ分解
- 具体的で実行可能なタスクに分解
- 各ステップの成果物を明確化
- テスト可能な単位で分割

### 4. 実装順序
- 依存関係に基づく順序付け
- リスクの低いものから着手
- インクリメンタルな進行

## 出力フォーマット

```markdown
## Plan: <機能名>

### Overview
<1-2文での概要>

### Requirements
- [ ] 要件1
- [ ] 要件2

### Affected Files
| File | Changes |
|------|---------|
| `model.go` | 状態フィールド追加 |
| `update.go` | キーハンドリング追加 |
| `view.go` | レンダリング追加 |

### Implementation Phases

#### Phase 1: <フェーズ名>
1. <ステップ1>
2. <ステップ2>

#### Phase 2: <フェーズ名>
1. <ステップ1>
2. <ステップ2>

### Testing Strategy
- <テストアプローチ>

### Risks
| Risk | Mitigation |
|------|------------|
| <リスク1> | <対策> |

### Success Criteria
- [ ] 基準1
- [ ] 基準2

### Questions
- <未解決の質問>
```

## 原則

1. **具体的に**: ファイルパス、関数名を明示
2. **エッジケース考慮**: 境界条件を忘れない
3. **インクリメンタル**: 小さく段階的に
4. **規約遵守**: プロジェクトの慣習に従う
5. **テスト可能**: 各ステップでテスト可能に

## bon3ai での典型的な変更パターン

### 新しい InputMode 追加

| File | Changes |
|------|---------|
| `model.go` | InputMode 定義追加 |
| `update.go` | キーハンドリング追加 |
| `view.go` | レンダリング追加 |

**Follow-up**:
- `.claude/rules/architecture.md`: 状態遷移図更新
- `README.md`: キーバインド表更新

### 新しいキーバインド追加

| File | Changes |
|------|---------|
| `update.go` | case 文追加 |
| `*_test.go` | テスト追加 |

**Follow-up**:
- `README.md`: キーバインド表更新

### VCS 機能追加

| File | Changes |
|------|---------|
| `vcs.go` | VCSRepo インターフェース拡張 |
| `gitstatus.go` | Git 実装 |
| `jjstatus.go` | JJ 実装 |

## 使用タイミング

✅ 使う場面:
- 新機能の実装計画
- 大規模リファクタリング
- アーキテクチャ変更

❌ 使わない場面:
- 単純なバグ修正
- 1ファイルの小さな変更
- ドキュメント更新のみ

## Usage

```
/planner Add bookmark feature
/planner Implement fuzzy file search
/planner Add split pane view
```
