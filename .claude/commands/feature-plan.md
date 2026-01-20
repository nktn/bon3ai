# /feature-plan - Feature Planning

Plan the implementation of a new feature or change.

## Process

1. **Understand the request**
   - Clarify requirements
   - Identify affected components

2. **Analyze codebase**
   - Find relevant files
   - Understand existing patterns
   - Check for similar implementations

3. **Create implementation plan**
   - List files to modify/create
   - Outline changes for each file
   - Identify potential risks

4. **Output format**

```markdown
## Plan: <feature name>

### Affected Files
- `file1.go`: <brief description of changes>
- `file2.go`: <brief description of changes>

### Implementation Steps
1. <step 1>
2. <step 2>
3. <step 3>

### Testing Strategy
- <test approach>

### Risks/Considerations
- <potential issues>

### Questions
- <unresolved questions, if any>
```

## Guidelines

- Keep plans concise
- Focus on "what" and "where", not "how" in detail
- Identify dependencies between steps
- Always include testing strategy
- List unresolved questions at the end

## bon3ai での典型的な変更パターン

### 新しい InputMode 追加

```markdown
### Affected Files
- `model.go`: InputMode 定義追加
- `update.go`: キーハンドリング追加
- `view.go`: レンダリング追加

### Follow-up
- `.claude/rules/architecture.md`: 状態遷移図更新
- `README.md`: キーバインド表更新
```

### 新しいキーバインド追加

```markdown
### Affected Files
- `update.go`: case 文追加
- `*_test.go`: テスト追加

### Follow-up
- `README.md`: キーバインド表更新
```

### VCS 機能追加

```markdown
### Affected Files
- `vcs.go`: VCSRepo インターフェース拡張
- `gitstatus.go`: Git 実装
- `jjstatus.go`: JJ 実装
```

## Usage

```
/feature-plan Add bookmark feature
/feature-plan Implement fuzzy file search
/feature-plan Add split pane view
```
