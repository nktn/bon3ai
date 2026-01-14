# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

bon3ai is a file explorer TUI (Terminal User Interface) built in Go using the Bubble Tea framework. It features Vim-style keybindings, Git/Jujutsu VCS integration, and file operations.

## Plan Mode

- Make the plan extremely concise. Sacrifice grammar for the sake of concision.
- At the end of each plan, give me a list of unresolved questions to answer, if any.

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

## Development Guidelines

### Writing Tests

Always create corresponding tests when implementing new features:

1. Add tests to the `*_test.go` file that corresponds to your implementation
2. Run `make test` to verify all tests pass
3. Write tests especially for the following cases:
   - New functions and methods
   - Changes to existing function behavior
   - Edge cases and error handling

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
- `ModePreview`: File preview (text/binary/image)

### Preview System

`ModePreview` handles three types of file preview:

| Type | Detection | Display | Close Action |
|------|-----------|---------|--------------|
| Text | Not binary, not image | Line-numbered text | None |
| Binary | Null bytes or >30% non-printable | Hex dump (16 bytes/line) | None |
| Image | File extension match | chafa (Kitty) or ASCII art fallback | Kitty graphics delete (chafa only) |

**Supported image formats**: PNG, JPG, JPEG, GIF, BMP, WebP, TIFF, TIF, ICO

**Image preview implementation** (`update.go`):
- `isImageFile()`: Extension-based detection
- `loadImagePreview()`: Tries chafa first, falls back to ASCII art via image2ascii
- `loadASCIIPreview()`: Converts image to colored ASCII art using image2ascii library
- `clearKittyImages()`: Sends `\x1b_Ga=d,d=A\x1b\\` to delete all Kitty graphics

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

- **Sync VCS Refresh**: Git/Jujutsu status runs synchronously for simplicity and correctness
- **Watcher Toggle**: When disabled, watcher is fully closed (zero performance impact)
- **VCS Directory Watching**: Watches `.git` and `.jj` directories to detect commits/status changes

## Documentation

- **docs/state-machine.md**: InputMode の状態遷移図（Mermaid形式）

### 重要: ステートマシン図の更新

InputMode に関する変更を行った場合は、必ず以下を確認・更新すること:

1. `docs/state-machine.md` の状態遷移図が実装と一致しているか確認
2. 新しい状態や遷移を追加した場合は図を更新
3. `README.md` のキーバインド表と整合性があるか確認

## Key Dependencies

### Go Modules
- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling
- `github.com/charmbracelet/x/ansi`: ANSI string manipulation for overlay compositing
- `github.com/fsnotify/fsnotify`: File system notifications for real-time watching
- `github.com/qeesung/image2ascii`: ASCII art image conversion (fallback for image preview)

### External Tools (Optional)
- `chafa`: High-quality image preview in terminal (install: `brew install chafa`)
  - Uses Kitty graphics protocol for best display quality
  - If not installed, falls back to ASCII art preview via image2ascii
