package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetCompletions(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test directories and files
	dirs := []string{"Documents", "Downloads", "Desktop"}
	files := []string{"file1.txt", "file2.txt"}

	for _, dir := range dirs {
		if err := os.Mkdir(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("failed to create test dir: %v", err)
		}
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, file), []byte{}, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	t.Run("empty input returns nothing", func(t *testing.T) {
		candidates, prefix := getCompletions("", "")
		if len(candidates) != 0 {
			t.Errorf("expected no candidates for empty input, got %d", len(candidates))
		}
		if prefix != "" {
			t.Errorf("expected empty prefix for empty input, got %q", prefix)
		}
	})

	t.Run("prefix match returns matching entries", func(t *testing.T) {
		input := filepath.Join(tmpDir, "Do")
		candidates, _ := getCompletions(input, "")

		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates (Documents, Downloads), got %d: %v", len(candidates), candidates)
		}

		// Verify all candidates start with prefix
		for _, c := range candidates {
			if !strings.HasPrefix(c, input) {
				t.Errorf("candidate %q does not start with prefix %q", c, input)
			}
		}
	})

	t.Run("single match returns one candidate", func(t *testing.T) {
		input := filepath.Join(tmpDir, "Des")
		candidates, prefix := getCompletions(input, "")

		if len(candidates) != 1 {
			t.Errorf("expected 1 candidate (Desktop), got %d: %v", len(candidates), candidates)
		}

		expectedPath := filepath.Join(tmpDir, "Desktop") + string(os.PathSeparator)
		if prefix != expectedPath {
			t.Errorf("expected prefix %q, got %q", expectedPath, prefix)
		}
	})

	t.Run("directory trailing slash lists contents", func(t *testing.T) {
		input := tmpDir + string(os.PathSeparator)
		candidates, _ := getCompletions(input, "")

		// Should list all non-hidden entries (3 dirs + 2 files)
		if len(candidates) != 5 {
			t.Errorf("expected 5 candidates, got %d: %v", len(candidates), candidates)
		}
	})

	t.Run("directories have trailing separator", func(t *testing.T) {
		input := filepath.Join(tmpDir, "Doc")
		candidates, _ := getCompletions(input, "")

		if len(candidates) != 1 {
			t.Fatalf("expected 1 candidate, got %d", len(candidates))
		}

		if !strings.HasSuffix(candidates[0], string(os.PathSeparator)) {
			t.Errorf("directory candidate should have trailing separator: %q", candidates[0])
		}
	})

	t.Run("file candidates have no trailing separator", func(t *testing.T) {
		input := filepath.Join(tmpDir, "file1")
		candidates, _ := getCompletions(input, "")

		if len(candidates) != 1 {
			t.Fatalf("expected 1 candidate, got %d", len(candidates))
		}

		if strings.HasSuffix(candidates[0], string(os.PathSeparator)) {
			t.Errorf("file candidate should not have trailing separator: %q", candidates[0])
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		input := filepath.Join(tmpDir, "do")
		candidates, _ := getCompletions(input, "")

		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates (case insensitive), got %d: %v", len(candidates), candidates)
		}
	})

	t.Run("nonexistent path returns nothing", func(t *testing.T) {
		input := filepath.Join(tmpDir, "nonexistent", "path")
		candidates, _ := getCompletions(input, "")

		if len(candidates) != 0 {
			t.Errorf("expected no candidates for nonexistent path, got %d", len(candidates))
		}
	})
}

func TestGetCompletionsHiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hidden and non-hidden files
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte{}, 0644); err != nil {
		t.Fatalf("failed to create hidden file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "visible"), []byte{}, 0644); err != nil {
		t.Fatalf("failed to create visible file: %v", err)
	}

	t.Run("hidden files excluded by default", func(t *testing.T) {
		input := tmpDir + string(os.PathSeparator)
		candidates, _ := getCompletions(input, "")

		if len(candidates) != 1 {
			t.Errorf("expected 1 candidate (visible only), got %d: %v", len(candidates), candidates)
		}
	})

	t.Run("hidden files included when prefix is dot", func(t *testing.T) {
		input := filepath.Join(tmpDir, ".")
		candidates, _ := getCompletions(input, "")

		if len(candidates) != 1 {
			t.Errorf("expected 1 candidate (.hidden), got %d: %v", len(candidates), candidates)
		}
	})
}

func TestGetCompletionsRelativePath(t *testing.T) {
	// Create base directory with subdirectories
	baseDir := t.TempDir()
	subDir := filepath.Join(baseDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create files in subdir
	for _, name := range []string{"alpha.txt", "beta.txt"} {
		if err := os.WriteFile(filepath.Join(subDir, name), []byte{}, 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	t.Run("relative path resolved against baseDir", func(t *testing.T) {
		// "subdir/" should be resolved relative to baseDir
		candidates, _ := getCompletions("subdir/", baseDir)

		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates, got %d: %v", len(candidates), candidates)
		}

		// Verify candidates are absolute paths under subdir
		for _, c := range candidates {
			if !strings.HasPrefix(c, subDir) {
				t.Errorf("candidate %q should be under %q", c, subDir)
			}
		}
	})

	t.Run("relative prefix match with baseDir", func(t *testing.T) {
		candidates, _ := getCompletions("subdir/al", baseDir)

		if len(candidates) != 1 {
			t.Errorf("expected 1 candidate (alpha.txt), got %d: %v", len(candidates), candidates)
		}

		expected := filepath.Join(subDir, "alpha.txt")
		if len(candidates) == 1 && candidates[0] != expected {
			t.Errorf("expected %q, got %q", expected, candidates[0])
		}
	})

	t.Run("absolute path ignores baseDir", func(t *testing.T) {
		// Absolute path should not use baseDir
		input := subDir + string(os.PathSeparator)
		candidates, _ := getCompletions(input, "/some/other/path")

		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates for absolute path, got %d: %v", len(candidates), candidates)
		}
	})

	t.Run("empty baseDir uses relative to cwd", func(t *testing.T) {
		// With empty baseDir, relative paths are relative to cwd
		// Create a file in cwd to test (this is tricky, so we use absolute path test instead)
		candidates, _ := getCompletions("nonexistent_relative_path", "")

		if len(candidates) != 0 {
			t.Errorf("expected no candidates for nonexistent relative path without baseDir, got %d", len(candidates))
		}
	})
}

func TestGetCompletionsTilde(t *testing.T) {
	// Create a temporary directory to use as HOME
	// This avoids flaky tests when real HOME has no visible files
	tmpHome := t.TempDir()

	// Create test files in the fake home directory
	testDirs := []string{"Documents", "Downloads"}
	testFiles := []string{"readme.txt"}

	for _, dir := range testDirs {
		if err := os.Mkdir(filepath.Join(tmpHome, dir), 0755); err != nil {
			t.Fatalf("failed to create test dir: %v", err)
		}
	}
	for _, file := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpHome, file), []byte{}, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Save original HOME and set temporary one
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
	})

	t.Run("tilde alone lists home directory", func(t *testing.T) {
		candidates, _ := getCompletions("~", "")

		// Should return entries from home directory (2 dirs + 1 file = 3)
		if len(candidates) != 3 {
			t.Errorf("expected 3 candidates, got %d: %v", len(candidates), candidates)
		}

		// All candidates should be under home directory
		for _, c := range candidates {
			if !strings.HasPrefix(c, tmpHome) {
				t.Errorf("candidate %q should be under home %q", c, tmpHome)
			}
		}
	})

	t.Run("tilde with slash lists home directory", func(t *testing.T) {
		candidates, _ := getCompletions("~/", "")

		// Should return entries from home directory (2 dirs + 1 file = 3)
		if len(candidates) != 3 {
			t.Errorf("expected 3 candidates, got %d: %v", len(candidates), candidates)
		}

		// All candidates should be under home directory
		for _, c := range candidates {
			if !strings.HasPrefix(c, tmpHome) {
				t.Errorf("candidate %q should be under home %q", c, tmpHome)
			}
		}
	})

	t.Run("tilde with prefix matches in home", func(t *testing.T) {
		candidates, _ := getCompletions("~/Do", "")

		// Should match Documents and Downloads
		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates (Documents, Downloads), got %d: %v", len(candidates), candidates)
		}

		// Candidates should be under home directory
		for _, c := range candidates {
			if !strings.HasPrefix(c, tmpHome) {
				t.Errorf("candidate %q should be under home %q", c, tmpHome)
			}
		}
	})
}

func TestFindCommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name:     "empty list",
			paths:    []string{},
			expected: "",
		},
		{
			name:     "single path",
			paths:    []string{"/usr/local/bin"},
			expected: "/usr/local/bin",
		},
		{
			name:     "common prefix",
			paths:    []string{"/usr/local/bin", "/usr/local/lib"},
			expected: "/usr/local/",
		},
		{
			name:     "no common prefix",
			paths:    []string{"/usr/bin", "/etc/passwd"},
			expected: "/",
		},
		{
			name:     "identical paths",
			paths:    []string{"/home/user", "/home/user", "/home/user"},
			expected: "/home/user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCommonPrefix(tt.paths)
			if result != tt.expected {
				t.Errorf("findCommonPrefix(%v) = %q, want %q", tt.paths, result, tt.expected)
			}
		})
	}
}

func TestCollapseHomePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path under home",
			input:    filepath.Join(home, "Documents"),
			expected: "~/Documents",
		},
		{
			name:     "home directory itself",
			input:    home,
			expected: "~",
		},
		{
			name:     "path not under home",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collapseHomePath(tt.input)
			if result != tt.expected {
				t.Errorf("collapseHomePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
