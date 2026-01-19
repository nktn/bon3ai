# /update-codemaps - Update Code Maps

コードベースの構造マップを更新する。

## 使用方法

```
/update-codemaps
```

## コードマップ

### アーキテクチャ図

```
┌─────────────────────────────────────────┐
│                 main.go                  │
│         (Program initialization)         │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│               model.go                   │
│    (State: Model, InputMode, styles)     │
└─────────────────┬───────────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
┌───────▼───────┐   ┌───────▼───────┐
│   update.go   │   │   view.go     │
│  (Events →    │   │  (State →     │
│   State)      │   │   UI)         │
└───────────────┘   └───────────────┘
```

## プロセス

### 1. ファイル構造分析

```bash
# Go ファイル一覧
find . -name "*.go" -not -path "./vendor/*" | head -20

# 各ファイルの行数
wc -l *.go | sort -n

# パッケージ構造
go list ./...
```

### 2. 依存関係分析

```bash
# import 関係
grep -h "^import" *.go | sort | uniq
```

### 3. 更新対象

| ファイル | 内容 |
|----------|------|
| `CLAUDE.md` | Core Components セクション |
| `.claude/rules/architecture.md` | 状態遷移図 |
| `.claude/agents/architect.md` | アーキテクチャ図 |

## 出力フォーマット

```markdown
## Code Map Update Report

### File Structure
| File | Lines | Purpose |
|------|-------|---------|
| main.go | 50 | Entry point |
| model.go | 200 | State management |
| view.go | 450 | Rendering |
| update.go | 600 | Event handling |

### Module Dependencies
- model.go → (none)
- view.go → model.go
- update.go → model.go, view.go

### State Machine
- InputModes: 7
- Transitions: 12

### Updates Made
1. CLAUDE.md: Updated file list
2. architecture.md: Added new state
```

## 更新タイミング

- 新しいファイル追加時
- InputMode 追加時
- 大きな構造変更時
