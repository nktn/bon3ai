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

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
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
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains(%v, %q) = %v, expected %v", tt.slice, tt.item, result, tt.expected)
			}
		})
	}
}
