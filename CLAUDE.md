# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

bon3ai is a file explorer TUI (Terminal User Interface) built in Go using the Bubble Tea framework. It features Vim-style keybindings, Git/Jujutsu VCS integration, and file operations.

## Development Commands

```bash
# Install
go install github.com/nktn/bon3ai@latest

# Build
go build

# Run
./bon3ai [path]

# Test
go test -v ./...              # Run all tests
go test -v -run TestName      # Run specific test
```

## Architecture

The application follows the Elm architecture (Model-View-Update) via Bubble Tea:

### Core Components

- **main.go**: Entry point, initializes Bubble Tea program with alt screen and mouse support
- **model.go**: Application state (`Model` struct), input modes, and lipgloss styles
- **view.go**: Rendering logic for tree view, preview, status bar, and popups
- **update.go**: Event handling for keyboard, mouse, and window events

### Input Modes

The `InputMode` enum controls application behavior:
- `ModeNormal`: Default navigation/file operations
- `ModeSearch`, `ModeRename`, `ModeNewFile`, `ModeNewDir`: Text input modes
- `ModeConfirmDelete`: Deletion confirmation dialog
- `ModePreview`: File preview with hex view for binary files

### File System

- **filetree.go**: `FileTree` and `FileNode` structs manage the hierarchical tree structure with lazy loading of children. Uses a flattened `Nodes` slice for display indexing.
- **fileops.go**: File operations (copy, move, delete, rename, create) and internal `Clipboard` for yank/cut/paste

### VCS Integration

The VCS system uses an interface pattern for abstraction:

- **vcs.go**: `VCSRepo` interface and auto-detection logic (JJ takes priority over Git)
- **gitstatus.go**: `GitRepo` implementation using `git status --porcelain`
- **jjstatus.go**: `JJRepo` implementation using `jj status` and `jj log`

Both implementations propagate file status to parent directories for visual indication.

### Drag & Drop

- **drop.go**: Handles terminal paste events as file drops, with path normalization for quoted/escaped paths

### File Watching

- **watcher.go**: Real-time file system monitoring using fsnotify
  - Debounces events (200ms) to handle rapid file changes
  - Ignores Chmod events (Spotlight, antivirus, etc.)
  - Lazy watching: only watches expanded directories
  - Toggle with `W` key (complete resource cleanup when disabled)

### Performance Optimizations

- **Async VCS Refresh**: Git/Jujutsu status runs in background goroutine to avoid UI blocking
- **VCS Throttling**: VCS refresh limited to every 5 seconds (tree refresh remains immediate)
- **Watcher Toggle**: When disabled, watcher is fully closed (zero performance impact)

## Key Dependencies

- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling
- `github.com/charmbracelet/x/ansi`: ANSI string manipulation for overlay compositing
- `github.com/fsnotify/fsnotify`: File system notifications for real-time watching
