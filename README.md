# bon3ai - TUI File Tree Explorer

A TUI file explorer built in Go with Vim keybinds and Git/Jujutsu status display.

> **Note:** The project is called **bon3ai**, but the command is **`bon3`** 

## Features

- **VCS status display** - Git and Jujutsu (jj) support with color-coded file status
- **Real-time file watching** - Auto-refresh on file system changes (toggle with `W`)
- **Vim like navigation** - `h / j / k / l` keys, `g`/`G` for jump
- **Mouse support** - Click, double-click, scroll
- **File operations** - Copy, cut, paste, delete, rename
- **Multi-select** - Mark multiple files with `Space`
- **Quick search** - Incremental search with `/`
- **File preview** - Text, binary (hex), and image preview (PNG, JPG, GIF, etc.)
- **Hidden files toggle** - Show/hide dotfiles with `.`
- **Path copying** - Copy file path to system clipboard
- **File icons** - Icons with Nerd Fonts
- **Drag & Drop** - Drop files to copy into selected folder

## Installation

### From source

```bash
git clone https://github.com/nktn/bon3ai.git
cd bon3ai
make install
```

### Manual build

```bash
make build   # Creates ./bon3
```

## Usage

```bash
bon3              # Current directory
bon3 ~/Documents  # Specific directory
```

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `l` / `Enter` | Expand directory |
| `h` / `Backspace` | Collapse / Go to parent |
| `g` / `G` | Jump to top / bottom |
| `Tab` | Toggle expand/collapse |
| `H` | Collapse all |
| `L` | Expand all |

### File Operations

| Key | Action |
|-----|--------|
| `Space` | Mark/unmark file |
| `Esc` | Clear all marks |
| `y` | Yank |
| `d` | Cut |
| `p` | Paste |
| `D` / `Delete` | Delete |
| `r` | Rename |
| `a` | New file |
| `A` | New directory |
| `o` | Preview file |

### View

| Key | Action |
|-----|--------|
| `.` | Toggle hidden files |
| `R` / `F5` | Reload tree |
| `W` | Toggle file watching |

### Preview Mode

| Key | Action |
|-----|--------|
| `j` / `k` / `↑` / `↓` | Scroll up / down |
| `f` / `Space` / `PgDn` | Page down |
| `b` / `PgUp` | Page up |
| `g` / `G` | Jump to top / bottom |
| `q` / `Esc` / `o` | Close preview |

**Preview types:**
- **Text**: Line-numbered display
- **Binary**: Hex dump view (16 bytes per line)
- **Image**: High-quality display via Kitty graphics protocol (requires `chafa`)

### Other

| Key | Action |
|-----|--------|
| `c` | Copy full path to clipboard |
| `C` | Copy filename to clipboard |
| `/` | Search |
| `n` | Next search match |
| `?` | Show help |
| `q` / `Ctrl+C` | Quit |

## Mouse

| Action | Effect |
|--------|--------|
| Click | Select |
| Double-click | Expand/collapse |
| Scroll | Navigate |
| Drag & Drop | Copy file to selected folder |

## VCS Status Colors

| Color | Status |
|-------|--------|
| Green | New / Untracked |
| Yellow | Modified |
| Red | Deleted |
| Cyan | Renamed |
| Gray | Ignored |
| Magenta | Conflict |

## VCS Status Display

bon3ai automatically detects the version control system and shows status:

- **Git**: Shows branch name in status bar (e.g., ` main`)
- **Jujutsu (jj)**: Shows change ID and bookmark (e.g., ` @hogehoge (main)`)

Priority: If both `.jj` and `.git` exist, Jujutsu is used (common for jj users working with GitHub).

## Requirements

- Go 1.24+
- Terminal with UTF-8 support
- [Nerd Font](https://www.nerdfonts.com/) (recommended for icons)
- Git or Jujutsu (optional, for VCS features)
- [chafa](https://hpjansson.org/chafa/) (optional, for image preview)

## License

MIT
