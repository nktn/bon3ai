# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

bon3ai is a file explorer TUI (Terminal User Interface) built in Go using the Bubble Tea framework. It features Vim-style keybindings, Git/Jujutsu VCS integration, and file operations.

## Development Commands

```bash
# Build
make build          # Build to bin/bon3ai
go build -o bin/bon3ai .  # Alternative direct build

# Run
make run            # Build and run
./bin/bon3ai [path]

# Test
make test           # Run all tests with verbose output
go test -v ./...    # Direct test command
go test -v -run TestName  # Run specific test

# Install/Uninstall
make install        # Install to /usr/local/bin
make uninstall      # Remove from /usr/local/bin

# Clean
make clean          # Remove build artifacts
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

## Key Dependencies

- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling
- `github.com/charmbracelet/x/ansi`: ANSI string manipulation for overlay compositing
