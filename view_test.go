package main

import "testing"

func TestGetFileIcon(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		// Special filenames
		{"gitignore", ".gitignore", ""},
		{"gitattributes", ".gitattributes", ""},
		{"gitmodules", ".gitmodules", ""},

		// Programming languages
		{"go file", "main.go", ""},
		{"rust file", "main.rs", ""},
		{"python file", "script.py", ""},
		{"javascript file", "app.js", ""},
		{"jsx file", "component.jsx", ""},
		{"typescript file", "app.ts", ""},
		{"tsx file", "component.tsx", ""},

		// Web
		{"html file", "index.html", ""},
		{"css file", "style.css", ""},
		{"scss file", "style.scss", ""},

		// Config
		{"json file", "config.json", ""},
		{"toml file", "config.toml", ""},
		{"yaml file", "config.yaml", ""},
		{"yml file", "config.yml", ""},

		// Documents
		{"markdown file", "README.md", ""},
		{"text file", "notes.txt", ""},
		{"pdf file", "document.pdf", ""},
		{"word file", "document.docx", ""},

		// Media
		{"png image", "image.png", ""},
		{"jpg image", "photo.jpg", ""},
		{"svg image", "icon.svg", ""},
		{"mp3 audio", "song.mp3", ""},
		{"mp4 video", "video.mp4", ""},

		// Archives
		{"zip file", "archive.zip", ""},
		{"tar file", "archive.tar", ""},
		{"gz file", "archive.gz", ""},

		// Other
		{"lock file", "package-lock.json", ""}, // .json takes precedence
		{"shell script", "script.sh", ""},
		{"bash script", "script.bash", ""},
		{"unknown extension", "file.xyz", ""},
		{"no extension", "Makefile", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileIcon(tt.filename)
			if result != tt.expected {
				t.Errorf("getFileIcon(%q) = %q, expected %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetFileIcon_CaseInsensitive(t *testing.T) {
	// Extension matching should be case-insensitive
	tests := []struct {
		filename string
		expected string
	}{
		{"FILE.GO", ""},
		{"Image.PNG", ""},
		{"Style.CSS", ""},
		{"Config.JSON", ""},
	}

	for _, tt := range tests {
		result := getFileIcon(tt.filename)
		if result != tt.expected {
			t.Errorf("getFileIcon(%q) = %q, expected %q", tt.filename, result, tt.expected)
		}
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
