package main

import (
	"os"
	"strings"
)

// IconSet contains icons for a specific display mode
type IconSet struct {
	FolderClosed string
	FolderOpen   string
	File         string
	FileText     string
	Ghost        string
	// File type icons
	Go         string
	Rust       string
	Python     string
	JavaScript string
	TypeScript string
	React      string
	HTML       string
	CSS        string
	JSON       string
	Markdown   string
	Lock       string
	Image      string
	Audio      string
	Video      string
	Archive    string
	PDF        string
	Word       string
	Terminal   string
	Git        string
}

// Nerd Font icons (requires Nerd Font)
var nerdFontIcons = IconSet{
	FolderClosed: "\uf07b", // nf-fa-folder
	FolderOpen:   "\uf07c", // nf-fa-folder_open
	File:         "\uf15b", // nf-fa-file
	FileText:     "\uf15c", // nf-fa-file_text
	Ghost:        "\uf4a4", // nf-md-ghost
	Go:           "\ue627", // nf-seti-go
	Rust:         "\ue7a8", // nf-dev-rust
	Python:       "\ue73c", // nf-dev-python
	JavaScript:   "\ue781", // nf-dev-javascript
	TypeScript:   "\ue628", // nf-seti-typescript
	React:        "\ue7ba", // nf-dev-react
	HTML:         "\ue736", // nf-dev-html5
	CSS:          "\ue749", // nf-dev-css3
	JSON:         "\ue60b", // nf-seti-json
	Markdown:     "\ue73e", // nf-dev-markdown
	Lock:         "\uf023", // nf-fa-lock
	Image:        "\uf1c5", // nf-fa-file_image
	Audio:        "\uf1c7", // nf-fa-file_audio
	Video:        "\uf1c8", // nf-fa-file_video
	Archive:      "\uf1c6", // nf-fa-file_archive
	PDF:          "\uf1c1", // nf-fa-file_pdf
	Word:         "\uf1c2", // nf-fa-file_word
	Terminal:     "\ue795", // nf-dev-terminal
	Git:          "\ue702", // nf-dev-git
}

// ASCII fallback icons (works on any terminal)
var asciiIcons = IconSet{
	FolderClosed: "+",
	FolderOpen:   "-",
	File:         " ",
	FileText:     " ",
	Ghost:        "x",
	Go:           " ",
	Rust:         " ",
	Python:       " ",
	JavaScript:   " ",
	TypeScript:   " ",
	React:        " ",
	HTML:         " ",
	CSS:          " ",
	JSON:         " ",
	Markdown:     " ",
	Lock:         " ",
	Image:        " ",
	Audio:        " ",
	Video:        " ",
	Archive:      " ",
	PDF:          " ",
	Word:         " ",
	Terminal:     " ",
	Git:          " ",
}

// IconMode represents the icon display mode
type IconMode int

const (
	// IconModeNerdFont uses Nerd Font icons (default)
	IconModeNerdFont IconMode = iota
	// IconModeASCII uses ASCII fallback icons
	IconModeASCII
)

// icons is the current icon set in use
var icons = nerdFontIcons

// currentIconMode tracks the current icon mode for testing
var currentIconMode = IconModeNerdFont

func init() {
	if os.Getenv("BON3_ASCII_ICONS") != "" {
		SetIconMode(IconModeASCII)
	}
}

// SetIconMode changes the icon mode (useful for testing)
func SetIconMode(mode IconMode) {
	currentIconMode = mode
	switch mode {
	case IconModeASCII:
		icons = asciiIcons
	default:
		icons = nerdFontIcons
	}
}

// GetIconMode returns the current icon mode
func GetIconMode() IconMode {
	return currentIconMode
}

// getFileIconByExt returns the icon for a file based on its extension
func getFileIconByExt(name string) string {
	// Check for special filenames first
	lowerName := strings.ToLower(name)
	switch lowerName {
	case ".gitignore", ".gitattributes", ".gitmodules":
		return icons.Git
	}

	// Check extension
	// Note: dotfiles like ".bashrc" have idx=0, so we skip them (no extension)
	// Files like ".config.json" have idx>0 for the last dot, so they have an extension
	if idx := strings.LastIndex(name, "."); idx > 0 {
		ext := strings.ToLower(name[idx+1:])
		switch ext {
		case "go":
			return icons.Go
		case "rs":
			return icons.Rust
		case "py":
			return icons.Python
		case "js":
			return icons.JavaScript
		case "jsx", "tsx":
			return icons.React
		case "ts":
			return icons.TypeScript
		case "html":
			return icons.HTML
		case "css", "scss", "sass":
			return icons.CSS
		case "json", "toml", "yaml", "yml":
			return icons.JSON
		case "md":
			return icons.Markdown
		case "txt":
			return icons.FileText
		case "lock":
			return icons.Lock
		case "png", "jpg", "jpeg", "gif", "svg", "ico", "webp", "bmp":
			return icons.Image
		case "mp3", "wav", "flac", "ogg", "m4a":
			return icons.Audio
		case "mp4", "mkv", "avi", "mov", "webm":
			return icons.Video
		case "zip", "tar", "gz", "rar", "7z":
			return icons.Archive
		case "pdf":
			return icons.PDF
		case "doc", "docx":
			return icons.Word
		case "sh", "bash", "zsh", "fish":
			return icons.Terminal
		}
	}

	return icons.File
}
