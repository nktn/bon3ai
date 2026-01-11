package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeDroppedPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path",
			input:    "/path/to/file",
			expected: "/path/to/file",
		},
		{
			name:     "path with single quotes",
			input:    "'/path/to/file'",
			expected: "/path/to/file",
		},
		{
			name:     "path with double quotes",
			input:    "\"/path/to/file\"",
			expected: "/path/to/file",
		},
		{
			name:     "path with escaped space",
			input:    "/path/to/my\\ file",
			expected: "/path/to/my file",
		},
		{
			name:     "path with multiple escapes",
			input:    "/path/to/file\\ with\\ spaces",
			expected: "/path/to/file with spaces",
		},
		{
			name:     "path with whitespace",
			input:    "  /path/to/file  ",
			expected: "/path/to/file",
		},
		{
			name:     "path with escaped parentheses",
			input:    "/path/to/file\\(1\\)",
			expected: "/path/to/file(1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeDroppedPath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeDroppedPath(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDroppedPaths(t *testing.T) {
	// Create temporary files for testing
	tmpDir, err := os.MkdirTemp("", "drop_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	file3 := filepath.Join(tmpDir, "file with spaces.txt")

	os.WriteFile(file1, []byte("test"), 0644)
	os.WriteFile(file2, []byte("test"), 0644)
	os.WriteFile(file3, []byte("test"), 0644)

	tests := []struct {
		name          string
		input         string
		expectedCount int
	}{
		{
			name:          "single path",
			input:         file1,
			expectedCount: 1,
		},
		{
			name:          "newline separated paths",
			input:         file1 + "\n" + file2,
			expectedCount: 2,
		},
		{
			name:          "space separated paths",
			input:         file1 + " " + file2,
			expectedCount: 2,
		},
		{
			name:          "quoted path with spaces",
			input:         "\"" + file3 + "\"",
			expectedCount: 1,
		},
		{
			name:          "non-existent path",
			input:         "/non/existent/path",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDroppedPaths(tt.input)
			if len(result) != tt.expectedCount {
				t.Errorf("parseDroppedPaths(%q) returned %d paths, expected %d", tt.input, len(result), tt.expectedCount)
			}
		})
	}
}

func TestParseDroppedPaths_EscapedSpaces(t *testing.T) {
	// Create temporary file with space in name
	tmpDir, err := os.MkdirTemp("", "drop_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fileWithSpace := filepath.Join(tmpDir, "my file.txt")
	os.WriteFile(fileWithSpace, []byte("test"), 0644)

	// Test with escaped space (as would be received from terminal)
	escapedPath := tmpDir + "/my\\ file.txt"
	result := parseDroppedPaths(escapedPath)

	if len(result) != 1 {
		t.Errorf("Expected 1 path, got %d", len(result))
	}

	if len(result) > 0 && result[0] != fileWithSpace {
		t.Errorf("Expected %q, got %q", fileWithSpace, result[0])
	}
}
