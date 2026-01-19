# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

bon3ai is a file explorer TUI (Terminal User Interface) built in Go using the Bubble Tea framework. It features Vim-style keybindings, Git/Jujutsu VCS integration, and file operations.

## Development Commands

```bash
# Build
make build          # Creates ./bon3

# Install to $GOPATH/bin
make install

# Run
./bon3 [path]

# Test
make test                     # Run all tests
go test -v -run TestName      # Run specific test

# Clean
make clean
```

## Architecture

The application follows the Elm architecture (Model-View-Update) via Bubble Tea:

### Core Components

- **main.go**: Entry point, initializes Bubble Tea program with alt screen and mouse support
- **model.go**: Application state (`Model` struct), input modes, and lipgloss styles
- **view.go**: Rendering logic for tree view, preview, status bar, and popups
- **update.go**: Event handling for keyboard, mouse, and window events
- **filetree.go**: `FileTree` and `FileNode` structs for hierarchical tree with lazy loading
- **fileops.go**: File operations (copy, move, delete, rename, create) and clipboard
- **vcs.go**: `VCSRepo` interface and auto-detection (JJ priority over Git)
- **watcher.go**: Real-time file system monitoring using fsnotify

## Key Dependencies

### Go Modules
- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling
- `github.com/charmbracelet/x/ansi`: ANSI string manipulation
- `github.com/fsnotify/fsnotify`: File system notifications
- `github.com/qeesung/image2ascii`: ASCII art image conversion

### External Tools (Optional)
- `chafa`: High-quality image preview in terminal (`brew install chafa`)
  - Uses Kitty graphics protocol
  - Falls back to ASCII art if not installed

## Additional Documentation

### Workflow
- `.claude/WORKFLOW.md`: 開発ワークフローガイド

### Rules
- `.claude/rules/agents.md`: Agent usage guide
- `.claude/rules/architecture.md`: State machine and architecture details
- `.claude/rules/coding-style.md`: Go coding style guide
- `.claude/rules/development.md`: Development guidelines
- `.claude/rules/git-workflow.md`: Git commit and PR workflow
- `.claude/rules/performance.md`: Performance guidelines
- `.claude/rules/security.md`: Security guidelines
- `.claude/rules/testing.md`: Testing guidelines

### Skills & Commands
- `.claude/skills/codex/SKILL.md`: Codex CLI integration (`/codex`)
- `.claude/skills/tdd/SKILL.md`: TDD workflow (`/tdd`)
- `.claude/commands/plan.md`: Feature planning (`/plan`)
- `.claude/commands/tdd.md`: Test-driven development (`/tdd`)
- `.claude/commands/build-fix.md`: Build error fix (`/build-fix`)
- `.claude/commands/refactor-clean.md`: Refactor & cleanup (`/refactor-clean`)
- `.claude/commands/test-coverage.md`: Coverage analysis (`/test-coverage`)
- `.claude/commands/update-docs.md`: Documentation sync (`/update-docs`)
- `.claude/commands/update-codemaps.md`: Code map update (`/update-codemaps`)

### Agents
- `.claude/agents/architect.md`: アーキテクチャ設計・分析
- `.claude/agents/build-fixer.md`: ビルドエラー修正
- `.claude/agents/doc-updater.md`: ドキュメント同期・更新
- `.claude/agents/planner.md`: 実装計画の作成
- `.claude/agents/refactor-cleaner.md`: リファクタリング・整理
- `.claude/agents/tdd-guide.md`: テスト駆動開発ガイド
