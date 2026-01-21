package main

import "testing"

func TestGetFileIconByExt(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		// Special filenames
		{"gitignore", ".gitignore", icons.Git},
		{"gitattributes", ".gitattributes", icons.Git},
		{"gitmodules", ".gitmodules", icons.Git},

		// Programming languages
		{"go file", "main.go", icons.Go},
		{"rust file", "main.rs", icons.Rust},
		{"python file", "script.py", icons.Python},
		{"javascript file", "app.js", icons.JavaScript},
		{"jsx file", "component.jsx", icons.React},
		{"typescript file", "app.ts", icons.TypeScript},
		{"tsx file", "component.tsx", icons.React},

		// Web
		{"html file", "index.html", icons.HTML},
		{"css file", "style.css", icons.CSS},
		{"scss file", "style.scss", icons.CSS},

		// Config
		{"json file", "config.json", icons.JSON},
		{"toml file", "config.toml", icons.JSON},
		{"yaml file", "config.yaml", icons.JSON},
		{"yml file", "config.yml", icons.JSON},

		// Documents
		{"markdown file", "README.md", icons.Markdown},
		{"text file", "notes.txt", icons.FileText},
		{"pdf file", "document.pdf", icons.PDF},
		{"word file", "document.docx", icons.Word},

		// Media
		{"png image", "image.png", icons.Image},
		{"jpg image", "photo.jpg", icons.Image},
		{"svg image", "icon.svg", icons.Image},
		{"mp3 audio", "song.mp3", icons.Audio},
		{"mp4 video", "video.mp4", icons.Video},

		// Archives
		{"zip file", "archive.zip", icons.Archive},
		{"tar file", "archive.tar", icons.Archive},
		{"gz file", "archive.gz", icons.Archive},

		// Other
		{"lock file", "package-lock.json", icons.JSON}, // .json takes precedence
		{"shell script", "script.sh", icons.Terminal},
		{"bash script", "script.bash", icons.Terminal},
		{"unknown extension", "file.xyz", icons.File},
		{"no extension", "Makefile", icons.File},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileIconByExt(tt.filename)
			if result != tt.expected {
				t.Errorf("getFileIconByExt(%q) = %q, expected %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetFileIconByExt_CaseInsensitive(t *testing.T) {
	// Extension matching should be case-insensitive
	tests := []struct {
		filename string
		expected string
	}{
		{"FILE.GO", icons.Go},
		{"Image.PNG", icons.Image},
		{"Style.CSS", icons.CSS},
		{"Config.JSON", icons.JSON},
	}

	for _, tt := range tests {
		result := getFileIconByExt(tt.filename)
		if result != tt.expected {
			t.Errorf("getFileIconByExt(%q) = %q, expected %q", tt.filename, result, tt.expected)
		}
	}
}

func TestGetFileIconByExt_Dotfiles(t *testing.T) {
	// Dotfiles like .bashrc should not be treated as having an extension
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"dotfile bashrc", ".bashrc", icons.File},
		{"dotfile zshrc", ".zshrc", icons.File},
		{"dotfile profile", ".profile", icons.File},
		{"dotfile vimrc", ".vimrc", icons.File},
		{"dotfile with extension", ".config.json", icons.JSON},
		{"dotfile with extension yaml", ".eslintrc.yml", icons.JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileIconByExt(tt.filename)
			if result != tt.expected {
				t.Errorf("getFileIconByExt(%q) = %q, expected %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestSetIconMode(t *testing.T) {
	// Save original mode
	originalMode := GetIconMode()
	defer SetIconMode(originalMode)

	// Test switching to ASCII mode
	SetIconMode(IconModeASCII)
	if GetIconMode() != IconModeASCII {
		t.Error("SetIconMode(IconModeASCII) did not set mode correctly")
	}
	if icons.FolderClosed != "+" {
		t.Errorf("ASCII mode FolderClosed = %q, expected %q", icons.FolderClosed, "+")
	}
	if icons.FolderOpen != "-" {
		t.Errorf("ASCII mode FolderOpen = %q, expected %q", icons.FolderOpen, "-")
	}

	// Test switching back to Nerd Font mode
	SetIconMode(IconModeNerdFont)
	if GetIconMode() != IconModeNerdFont {
		t.Error("SetIconMode(IconModeNerdFont) did not set mode correctly")
	}
	if icons.FolderClosed != "\uf07b" {
		t.Errorf("NerdFont mode FolderClosed = %q, expected %q", icons.FolderClosed, "\uf07b")
	}
}

func TestASCIIIconsConsistency(t *testing.T) {
	// Save original mode
	originalMode := GetIconMode()
	defer SetIconMode(originalMode)

	// Switch to ASCII mode
	SetIconMode(IconModeASCII)

	// Verify getFileIconByExt uses ASCII icons
	goIcon := getFileIconByExt("main.go")
	if goIcon != " " {
		t.Errorf("ASCII mode Go icon = %q, expected %q", goIcon, " ")
	}

	folderIcon := icons.FolderClosed
	if folderIcon != "+" {
		t.Errorf("ASCII mode FolderClosed = %q, expected %q", folderIcon, "+")
	}
}

func TestClipboardContains(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		item     string
		expected bool
	}{
		{"item exists", []string{"a", "b", "c"}, "b", true},
		{"item not exists", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"single item match", []string{"a"}, "a", true},
		{"single item no match", []string{"a"}, "b", false},
		{"path match", []string{"/path/to/file"}, "/path/to/file", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clipboard := Clipboard{}
			clipboard.Copy(tt.paths)
			result := clipboard.Contains(tt.item)
			if result != tt.expected {
				t.Errorf("Clipboard.Contains(%q) = %v, expected %v", tt.item, result, tt.expected)
			}
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Bytes
		{"zero bytes", 0, "0B"},
		{"single byte", 1, "1B"},
		{"small bytes", 500, "500B"},
		{"max bytes", 1023, "1023B"},

		// Kilobytes
		{"exact KB", 1024, "1.0KB"},
		{"KB with decimal", 1536, "1.5KB"},
		{"large KB", 102400, "100.0KB"},
		{"max KB", 1048575, "1024.0KB"},

		// Megabytes
		{"exact MB", 1048576, "1.0MB"},
		{"MB with decimal", 1572864, "1.5MB"},
		{"large MB", 104857600, "100.0MB"},

		// Gigabytes
		{"exact GB", 1073741824, "1.0GB"},
		{"GB with decimal", 1610612736, "1.5GB"},
		{"large GB", 10737418240, "10.0GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatFileSize(%d) = %q, expected %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		expected []string
	}{
		// No wrapping needed
		{"short text", "hello", 10, []string{"hello"}},
		{"exact width", "hello", 5, []string{"hello"}},
		{"empty text", "", 10, []string{""}},

		// Basic wrapping
		{"simple wrap", "hello world", 6, []string{"hello ", "world"}},
		{"long text", "abcdefghij", 3, []string{"abc", "def", "ghi", "j"}},

		// Edge cases
		{"zero width", "hello", 0, []string{"hello"}},
		{"negative width", "hello", -5, []string{"hello"}},
		{"single char width", "abc", 1, []string{"a", "b", "c"}},

		// Unicode/CJK handling
		{"japanese chars", "あいう", 4, []string{"あい", "う"}},
		{"mixed text", "aあbいc", 4, []string{"aあb", "いc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.maxWidth)
			if len(result) != len(tt.expected) {
				t.Errorf("wrapText(%q, %d) returned %d lines, expected %d lines",
					tt.text, tt.maxWidth, len(result), len(tt.expected))
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("wrapText(%q, %d)[%d] = %q, expected %q",
						tt.text, tt.maxWidth, i, line, tt.expected[i])
				}
			}
		})
	}
}
